package cmd

import (
	"github.com/spf13/cobra"
)

// newProfileCmd creates the command for profile management.
func newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage configuration profiles",
		Long:  `The profile command and its subcommands help you manage different configurations for various destinations.`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(newProfileListCmd())   // from list.go
	cmd.AddCommand(newProfileUseCmd())     // from use.go
	cmd.AddCommand(newProfileAddCmd())     // from add.go
	cmd.AddCommand(newProfileRemoveCmd()) // from remove.go
	cmd.AddCommand(newProfileSetCmd())     // from set.go

	return cmd
}