package cmd

import (
	"github.com/spf13/cobra"
)

// newConfigCmd creates the command for configuration management.
func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  `Provides commands to manage the application configuration, such as initializing a new config file.`,
	}

	// Add subcommands
	cmd.AddCommand(newConfigInitCmd()) // Assuming newConfigInitCmd() will be created

	return cmd
}