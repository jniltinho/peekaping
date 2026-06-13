package infra

import (
	"fmt"

	"peekaping/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// NewGormDB opens a GORM connection used exclusively for schema migrations
// (AutoMigrate). Application queries use the bun.DB provided by ProvideSQLDB.
// Supported backends: postgres, mysql, mariadb, sqlite.
func NewGormDB(cfg *config.DBConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.DBType {
	case "postgres", "postgresql":
		dsn := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName,
		)
		dialector = postgres.Open(dsn)

	case "mysql", "mariadb":
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?parseTime=True&loc=UTC&charset=utf8mb4",
			cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName,
		)
		dialector = mysql.Open(dsn)

	case "sqlite":
		dbPath := cfg.DBName
		if dbPath == "" {
			dbPath = "./data.db"
		}
		connStr := fmt.Sprintf(
			"file:%s?cache=shared&mode=rwc&_journal_mode=WAL&_foreign_keys=ON&_busy_timeout=5000",
			dbPath,
		)
		dialector = sqlite.Open(connStr)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.DBType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger:                                   logger.Default.LogMode(logger.Warn),
		DisableForeignKeyConstraintWhenMigrating: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open GORM connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
