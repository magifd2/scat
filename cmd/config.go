package cmd

import (
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `Provides commands to manage the application configuration, such as initializing a new config file.`,
}

func init() {
	rootCmd.AddCommand(configCmd)
}
