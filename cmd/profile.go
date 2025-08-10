package cmd

import (
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage configuration profiles",
	Long:  `The profile command and its subcommands help you manage different configurations for various destinations.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(profileCmd)
}
