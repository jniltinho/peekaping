package cmdmigrate

import (
	"database/sql"
	"fmt"
	"strings"

	"peekaping/cmd/bun/migrations"
	"peekaping/internal/config"

	"github.com/spf13/cobra"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/migrate"

	_ "github.com/go-sql-driver/mysql"
)

// NewMigrateCommand returns the Cobra command tree for database migrations.
func NewMigrateCommand(cfg *config.DBConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration commands",
	}

	cmd.AddCommand(
		newInitCmd(cfg),
		newUpCmd(cfg),
		newRollbackCmd(cfg),
		newLockCmd(cfg),
		newUnlockCmd(cfg),
		newCreateGoCmd(cfg),
		newCreateSQLCmd(cfg),
		newCreateTxSQLCmd(cfg),
		newStatusCmd(cfg),
		newMarkAppliedCmd(cfg),
	)

	return cmd
}

func newInitCmd(cfg *config.DBConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create migration tables",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithDB(cfg, func(db *bun.DB) error {
				migrator := migrate.NewMigrator(db, migrations.Migrations)
				return migrator.Init(cmd.Context())
			})
		},
	}
}

func newUpCmd(cfg *config.DBConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Run pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithDB(cfg, func(db *bun.DB) error {
				migrator := migrate.NewMigrator(db, migrations.Migrations)
				if err := migrator.Lock(cmd.Context()); err != nil {
					return err
				}
				defer migrator.Unlock(cmd.Context()) //nolint:errcheck

				group, err := migrator.Migrate(cmd.Context())
				if err != nil {
					migrator.Rollback(cmd.Context()) //nolint:errcheck
					return err
				}
				if group.IsZero() {
					fmt.Println("there are no new migrations to run (database is up to date)")
					return nil
				}
				fmt.Printf("migrated to %s\n", group)
				return nil
			})
		},
	}
}

func newRollbackCmd(cfg *config.DBConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "rollback",
		Short: "Rollback the last migration group",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithDB(cfg, func(db *bun.DB) error {
				migrator := migrate.NewMigrator(db, migrations.Migrations)
				if err := migrator.Lock(cmd.Context()); err != nil {
					return err
				}
				defer migrator.Unlock(cmd.Context()) //nolint:errcheck

				group, err := migrator.Rollback(cmd.Context())
				if err != nil {
					return err
				}
				if group.IsZero() {
					fmt.Println("there are no groups to roll back")
					return nil
				}
				fmt.Printf("rolled back %s\n", group)
				return nil
			})
		},
	}
}

func newLockCmd(cfg *config.DBConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "lock",
		Short: "Lock migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithDB(cfg, func(db *bun.DB) error {
				migrator := migrate.NewMigrator(db, migrations.Migrations)
				return migrator.Lock(cmd.Context())
			})
		},
	}
}

func newUnlockCmd(cfg *config.DBConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "unlock",
		Short: "Unlock migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithDB(cfg, func(db *bun.DB) error {
				migrator := migrate.NewMigrator(db, migrations.Migrations)
				return migrator.Unlock(cmd.Context())
			})
		},
	}
}

func newCreateGoCmd(cfg *config.DBConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "create-go <name>",
		Short: "Create a Go migration file",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithDB(cfg, func(db *bun.DB) error {
				migrator := migrate.NewMigrator(db, migrations.Migrations)
				name := strings.Join(args, "_")
				mf, err := migrator.CreateGoMigration(cmd.Context(), name)
				if err != nil {
					return err
				}
				fmt.Printf("created migration %s (%s)\n", mf.Name, mf.Path)
				return nil
			})
		},
	}
}

func newCreateSQLCmd(cfg *config.DBConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "create-sql <name>",
		Short: "Create up and down SQL migrations",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithDB(cfg, func(db *bun.DB) error {
				migrator := migrate.NewMigrator(db, migrations.Migrations)
				name := strings.Join(args, "_")
				files, err := migrator.CreateSQLMigrations(cmd.Context(), name)
				if err != nil {
					return err
				}
				for _, mf := range files {
					fmt.Printf("created migration %s (%s)\n", mf.Name, mf.Path)
				}
				return nil
			})
		},
	}
}

func newCreateTxSQLCmd(cfg *config.DBConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "create-tx-sql <name>",
		Short: "Create up and down transactional SQL migrations",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithDB(cfg, func(db *bun.DB) error {
				migrator := migrate.NewMigrator(db, migrations.Migrations)
				name := strings.Join(args, "_")
				files, err := migrator.CreateTxSQLMigrations(cmd.Context(), name)
				if err != nil {
					return err
				}
				for _, mf := range files {
					fmt.Printf("created transaction migration %s (%s)\n", mf.Name, mf.Path)
				}
				return nil
			})
		},
	}
}

func newStatusCmd(cfg *config.DBConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Print migrations status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithDB(cfg, func(db *bun.DB) error {
				migrator := migrate.NewMigrator(db, migrations.Migrations)
				ms, err := migrator.MigrationsWithStatus(cmd.Context())
				if err != nil {
					return err
				}
				fmt.Printf("migrations: %s\n", ms)
				fmt.Printf("unapplied migrations: %s\n", ms.Unapplied())
				fmt.Printf("last migration group: %s\n", ms.LastGroup())
				return nil
			})
		},
	}
}

func newMarkAppliedCmd(cfg *config.DBConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "mark-applied",
		Short: "Mark migrations as applied without running them",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithDB(cfg, func(db *bun.DB) error {
				migrator := migrate.NewMigrator(db, migrations.Migrations)
				group, err := migrator.Migrate(cmd.Context(), migrate.WithNopMigration())
				if err != nil {
					return err
				}
				if group.IsZero() {
					fmt.Println("there are no new migrations to mark as applied")
					return nil
				}
				fmt.Printf("marked as applied %s\n", group)
				return nil
			})
		},
	}
}

func runWithDB(cfg *config.DBConfig, fn func(*bun.DB) error) error {
	db, err := connectToDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithEnabled(false),
		bundebug.FromEnv(),
	))

	return fn(db)
}

func connectToDatabase(cfg *config.DBConfig) (*bun.DB, error) {
	var sqldb *sql.DB
	var db *bun.DB
	var err error

	switch cfg.DBType {
	case "postgres", "postgresql":
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=public",
			cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
		sqldb = sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
		db = bun.NewDB(sqldb, pgdialect.New())

	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=UTC&multiStatements=true",
			cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
		sqldb, err = sql.Open("mysql", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open MySQL connection: %w", err)
		}
		db = bun.NewDB(sqldb, mysqldialect.New())

	case "sqlite":
		dbPath := cfg.DBName
		if dbPath == "" {
			dbPath = "./data.db"
		}
		sqldb, err = sql.Open(sqliteshim.ShimName, fmt.Sprintf("file:%s?cache=shared&mode=rwc", dbPath))
		if err != nil {
			return nil, fmt.Errorf("failed to open SQLite connection: %w", err)
		}
		sqldb.SetMaxOpenConns(1)
		sqldb.SetMaxIdleConns(1)
		sqldb.SetConnMaxLifetime(0)
		db = bun.NewDB(sqldb, sqlitedialect.New())
		db.Exec("PRAGMA busy_timeout=5000")
		db.Exec("PRAGMA journal_mode=WAL")
		db.Exec("PRAGMA synchronous=NORMAL")
		db.Exec("PRAGMA foreign_keys=ON")
		db.Exec("PRAGMA temp_store=MEMORY")
		db.Exec("PRAGMA wal_autocheckpoint=1000")
		db.Exec("PRAGMA cache_size=-64000")
		db.Exec("PRAGMA auto_vacuum=INCREMENTAL")
		fmt.Printf("Connected to SQLite database: %s (WAL mode enabled)\n", dbPath)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.DBType)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
