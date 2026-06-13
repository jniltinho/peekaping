package cmd

import (
	"fmt"

	cmdapi "peekaping/cmd/api"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the REST API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		v := viper.New()
		// --port overrides SERVER_PORT only when explicitly passed by the user.
		v.BindPFlag("SERVER_PORT", cmd.Flags().Lookup("port")) //nolint:errcheck
		cfg, err := cmdapi.LoadAndValidate(envFileDir(), v)
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}
		return cmdapi.Run(cfg)
	},
}

func init() {
	apiCmd.Flags().StringP("port", "p", "8383", "Port the API server listens on (overrides SERVER_PORT)")
	rootCmd.AddCommand(apiCmd)
}
