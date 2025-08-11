package cmd

import (
	"context"

	"github.com/magifd2/scat/internal/appcontext"
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "scat",
	Version: version,
	Short:   "A general-purpose tool for posting messages from the command line.",
	Long: `scat is a versatile command-line interface for sending content from files or stdin to a configured HTTP endpoint.

It is inspired by slackcat but generalized to work with any compatible webhook or API endpoint.

Features:
- Post content from files or stdin.
- Stream stdin continuously.
- Manage multiple destination endpoints through profiles.`,
	SilenceUsage: true, // Suppress usage message on error
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		debug, _ := cmd.Flags().GetBool("debug")
		noOp, _ := cmd.Flags().GetBool("noop")
		silent, _ := cmd.Flags().GetBool("silent")
		appCtx := appcontext.NewContext(debug, noOp, silent)
		cmd.SetContext(context.WithValue(cmd.Context(), appcontext.CtxKey, appCtx))
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().Bool("noop", false, "Dry run, do not actually post or upload")
	rootCmd.PersistentFlags().Bool("silent", false, "Suppress informational messages")
}