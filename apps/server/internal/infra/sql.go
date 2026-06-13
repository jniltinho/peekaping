// Package infra provides the infrastructure layer: database connections,
// distributed event bus, and task queue providers consumed by the DI container.
package infra

import (
	"database/sql"
	"fmt"
	"peekaping/internal/config"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
	"go.uber.org/dig"
	"go.uber.org/zap"

	_ "github.com/go-sql-driver/mysql"
)

// ProvideSQLDB opens and verifies a database connection based on cfg.DBType,
// returning a bun.DB ready for use. Supported backends: postgres, mysql,
// mariadb, sqlite. For SQLite, WAL mode and serialised access are configured
// automatically to prevent lock contention under concurrent processes.
func ProvideSQLDB(
	cfg *config.Config,
	logger *zap.SugaredLogger,
) (*bun.DB, error) {
	var sqldb *sql.DB
	var db *bun.DB
	var err error

	switch cfg.DBType {
	case "postgres", "postgresql":
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
		sqldb = sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
		db = bun.NewDB(sqldb, pgdialect.New())
		logger.Infof("Connecting to PostgreSQL database: %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)

	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=UTC",
			cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
		sqldb, err = sql.Open("mysql", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open MySQL connection: %w", err)
		}
		db = bun.NewDB(sqldb, mysqldialect.New())
		logger.Infof("Connecting to MySQL database: %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)

	case "mariadb":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=UTC",
			cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
		sqldb, err = sql.Open("mysql", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open MariaDB connection: %w", err)
		}
		db = bun.NewDB(sqldb, mysqldialect.New())
		logger.Infof("Connecting to MySQL-compatible database (type: mariadb): %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)

	case "sqlite":
		dbPath := cfg.DBName
		if dbPath == "" {
			dbPath = "./data.db"
		}

		sqldb, err = sql.Open(sqliteshim.ShimName, fmt.Sprintf("file:%s?cache=shared&mode=rwc", dbPath))
		if err != nil {
			return nil, fmt.Errorf("failed to open SQLite connection: %w", err)
		}

		// SQLite requires serialised access; multiple open connections cause lock contention.
		sqldb.SetMaxOpenConns(1)
		sqldb.SetMaxIdleConns(1)
		sqldb.SetConnMaxLifetime(0)

		db = bun.NewDB(sqldb, sqlitedialect.New())

		// busy_timeout must be set first so that subsequent PRAGMAs can wait when the DB is locked.
		if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
			logger.Warnf("Failed to set busy_timeout (non-fatal): %v", err)
		}

		var currentJournalMode string
		if err := db.QueryRow("PRAGMA journal_mode").Scan(&currentJournalMode); err != nil {
			logger.Warnf("Failed to check journal_mode (non-fatal): %v", err)
		}

		if currentJournalMode != "wal" {
			if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
				logger.Warnf("Failed to set journal_mode to WAL (non-fatal, current mode: %s): %v", currentJournalMode, err)
			} else {
				logger.Infof("SQLite journal mode set to WAL")
			}
		} else {
			logger.Infof("SQLite already in WAL mode")
		}

		pragmas := []string{
			"PRAGMA synchronous=NORMAL",
			"PRAGMA foreign_keys=ON",
			"PRAGMA temp_store=MEMORY",
			"PRAGMA wal_autocheckpoint=1000",
			"PRAGMA cache_size=-64000",
			"PRAGMA auto_vacuum=INCREMENTAL",
		}
		for _, p := range pragmas {
			if _, err := db.Exec(p); err != nil {
				logger.Warnf("Failed to set %s (non-fatal): %v", p, err)
			}
		}

		var integrityResult string
		if err := db.QueryRow("PRAGMA integrity_check").Scan(&integrityResult); err != nil {
			logger.Warnf("Failed to run integrity check (non-fatal): %v", err)
		} else if integrityResult != "ok" {
			logger.Errorf("Database integrity check FAILED: %s - database may be corrupted!", integrityResult)
		} else {
			logger.Info("Database integrity check passed")
		}

		logger.Infof("Connecting to SQLite database: %s (WAL mode enabled)", dbPath)

	default:
		return nil, fmt.Errorf("unsupported database type: %s. Supported types: postgres, mysql, mariadb, sqlite", cfg.DBType)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.AddQueryHook(bundebug.NewQueryHook(bundebug.FromEnv()))

	logger.Info("Successfully connected to SQL database")
	return db, nil
}

// GracefulSQLiteShutdown checkpoints the WAL file, runs an incremental vacuum,
// and closes the connection. It is a no-op for non-SQLite backends.
func GracefulSQLiteShutdown(db *bun.DB, dbType string, logger *zap.SugaredLogger) error {
	if dbType != "sqlite" {
		return nil
	}

	logger.Info("Performing graceful SQLite shutdown...")

	if _, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		logger.Warnf("Failed to checkpoint WAL (non-fatal): %v", err)
	} else {
		logger.Info("WAL checkpoint completed successfully")
	}

	if _, err := db.Exec("PRAGMA incremental_vacuum"); err != nil {
		logger.Warnf("Failed to perform incremental vacuum (non-fatal): %v", err)
	}

	var integrityResult string
	if err := db.QueryRow("PRAGMA integrity_check").Scan(&integrityResult); err != nil {
		logger.Warnf("Failed to run shutdown integrity check (non-fatal): %v", err)
	} else if integrityResult != "ok" {
		logger.Errorf("Shutdown integrity check FAILED: %s", integrityResult)
	} else {
		logger.Info("Shutdown integrity check passed")
	}

	if err := db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	logger.Info("SQLite database closed gracefully")
	return nil
}

// GracefulDatabaseShutdown resolves the *bun.DB from the DI container and
// delegates to GracefulSQLiteShutdown for all supported database types.
func GracefulDatabaseShutdown(container *dig.Container, cfg *config.Config, logger *zap.SugaredLogger) error {
	switch cfg.DBType {
	case "postgres", "postgresql", "mysql", "mariadb", "sqlite":
		return container.Invoke(func(db *bun.DB) {
			if err := GracefulSQLiteShutdown(db, cfg.DBType, logger); err != nil {
				logger.Errorw("Failed to gracefully shutdown database", "error", err)
			}
		})
	default:
		logger.Warnf("Unknown database type for shutdown: %s", cfg.DBType)
		return nil
	}
}
