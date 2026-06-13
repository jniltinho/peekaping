package cmd

import (
	"fmt"

	"peekaping/version"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Peekaping\n")
		fmt.Printf("  Version:    %s\n", version.Version)
		fmt.Printf("  Build Time: %s\n", version.BuildTime)
		fmt.Printf("  Git Commit: %s\n", version.GitCommit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
