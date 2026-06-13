package main

import (
	"fmt"
	"os"
	"path/filepath"

	cmdapi "peekaping/cmd/api"
	cmdmigrate "peekaping/cmd/bun"
	cmdengine "peekaping/cmd/engine"
	"peekaping/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var envFile string

func main() {
	rootCmd := &cobra.Command{
		Use:   "monitoring",
		Short: "Monitoring uptime platform",
		Long:  "Monitoring — uptime monitoring platform. Use subcommands to start the API server, engine, or run database migrations.",
	}

	rootCmd.PersistentFlags().StringVar(&envFile, "env-file", "", "Path to .env configuration file (default: auto-discover)")
	viper.BindPFlag("env_file", rootCmd.PersistentFlags().Lookup("env-file")) //nolint:errcheck

	rootCmd.AddCommand(newAPICmd())
	rootCmd.AddCommand(newEngineCmd())
	rootCmd.AddCommand(newMigrateCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// envFileDir returns the directory to pass to LoadConfig so it finds the .env file.
// An empty string lets LoadConfig auto-discover (CWD, executable dir).
func envFileDir() string {
	if envFile == "" {
		return ""
	}
	return filepath.Dir(envFile)
}

func newAPICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "api",
		Short: "Start the REST API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdapi.LoadAndValidate(envFileDir())
			if err != nil {
				return fmt.Errorf("config: %w", err)
			}
			return cmdapi.Run(cfg)
		},
	}
}

func newEngineCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "engine",
		Short: "Start the engine (scheduler/worker/ingester)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdengine.LoadAndValidate(envFileDir())
			if err != nil {
				return fmt.Errorf("config: %w", err)
			}
			return cmdengine.Run(cfg)
		},
	}
}

func newMigrateCmd() *cobra.Command {
	// Allocate a shared DBConfig pointer; PersistentPreRunE populates it
	// before any subcommand RunE executes.
	dbCfg := &config.DBConfig{}

	migrateCmd := cmdmigrate.NewMigrateCommand(dbCfg)

	migrateCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig[config.DBConfig](envFileDir())
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}
		if err := config.ValidateDatabaseCustomRules(&cfg); err != nil {
			return fmt.Errorf("database config: %w", err)
		}
		*dbCfg = cfg
		return nil
	}

	return migrateCmd
}
