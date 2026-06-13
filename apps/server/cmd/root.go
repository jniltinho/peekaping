package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var envFile string

var rootCmd = &cobra.Command{
	Use:   "monitoring",
	Short: "Monitoring — uptime monitoring platform",
	Long:  "Monitoring — uptime monitoring platform. Use subcommands to start the API server, engine, or run database migrations.",
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&envFile, "env-file", "", "Path to .env configuration file (default: auto-discover)")
	viper.BindPFlag("env_file", rootCmd.PersistentFlags().Lookup("env-file")) //nolint:errcheck
}

// envFileDir returns the directory to pass to LoadConfig so it finds the .env file.
// An empty string lets LoadConfig auto-discover (CWD, executable dir).
func envFileDir() string {
	if envFile == "" {
		return ""
	}
	return filepath.Dir(envFile)
}
