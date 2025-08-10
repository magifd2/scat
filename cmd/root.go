package cmd

import (
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "scat",
	Version: version,
	Short: "A general-purpose tool for posting messages from the command line.",
	Long: `scat is a versatile command-line interface for sending content from files or stdin to a configured HTTP endpoint.

It is inspired by slackcat but generalized to work with any compatible webhook or API endpoint.

Features:
- Post content from files or stdin.
- Stream stdin continuously.
- Manage multiple destination endpoints through profiles.`, 
	SilenceUsage: true, // Suppress usage message on error
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
}
