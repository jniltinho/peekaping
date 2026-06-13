package cmdmigrate

import (
	"fmt"

	"peekaping/internal/config"
	"peekaping/internal/db"
	"peekaping/internal/infra"

	"github.com/spf13/cobra"
)

// NewMigrateCommand returns the Cobra command tree for database migrations.
func NewMigrateCommand(cfg *config.DBConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration commands",
	}

	cmd.AddCommand(
		newUpCmd(cfg),
		newStatusCmd(cfg),
	)

	return cmd
}

func newUpCmd(cfg *config.DBConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Run AutoMigrate to create or update the database schema",
		RunE: func(cmd *cobra.Command, args []string) error {
			gormDB, err := infra.NewGormDB(cfg)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			sqlDB, _ := gormDB.DB()
			defer sqlDB.Close()

			fmt.Println("Running GORM AutoMigrate...")
			if err := gormDB.AutoMigrate(db.AllModels()...); err != nil {
				return fmt.Errorf("AutoMigrate failed: %w", err)
			}
			fmt.Println("Database schema is up to date.")
			return nil
		},
	}
}

func newStatusCmd(cfg *config.DBConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check database connectivity",
		RunE: func(cmd *cobra.Command, args []string) error {
			gormDB, err := infra.NewGormDB(cfg)
			if err != nil {
				return fmt.Errorf("database unreachable: %w", err)
			}
			sqlDB, _ := gormDB.DB()
			defer sqlDB.Close()
			fmt.Printf("Database connection OK (%s)\n", cfg.DBType)
			return nil
		},
	}
}
