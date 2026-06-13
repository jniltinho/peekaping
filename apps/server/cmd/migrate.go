package cmd

import (
	"fmt"

	cmdmigrate "peekaping/cmd/migrate"
	"peekaping/internal/config"

	"github.com/spf13/cobra"
)

func init() {
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

	rootCmd.AddCommand(migrateCmd)
}
