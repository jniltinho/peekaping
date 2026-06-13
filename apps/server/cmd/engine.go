package cmd

import (
	"fmt"

	cmdengine "peekaping/cmd/engine"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var engineCmd = &cobra.Command{
	Use:   "engine",
	Short: "Start the engine (scheduler/worker/ingester)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := cmdengine.LoadAndValidate(envFileDir(), viper.New())
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}
		return cmdengine.Run(cfg)
	},
}

func init() {
	rootCmd.AddCommand(engineCmd)
}
