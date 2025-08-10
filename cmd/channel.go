package cmd

import (
	"github.com/spf13/cobra"
)

var channelCmd = &cobra.Command{
	Use:   "channel",
	Short: "Manage and view channel information",
	Long:  `The channel command and its subcommands help you interact with channel-related features, such as listing available channels for a provider.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(channelCmd)
}
