
package cmd

import (
	"context"

	"github.com/magifd2/scat/internal/appcontext"
	"github.com/spf13/cobra"
)

var version = "dev"

// newRootCmd creates the root command for scat.
func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "scat",
		Version: version,
		Short:   "A general-purpose tool for posting messages from the command line.",
		Long: `scat is a versatile command-line interface for sending content from files or stdin to a configured HTTP endpoint.

It is inspired by slackcat but generalized to work with any compatible webhook or API endpoint.

Features:
- Post content from files or stdin.
- Stream stdin continuously.
- Manage multiple destination endpoints through profiles.`,
		SilenceUsage:  true, // Suppress usage message on error
		SilenceErrors: true, // Suppress cobra's own error reporting
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			debug, _ := cmd.Flags().GetBool("debug")
			noOp, _ := cmd.Flags().GetBool("noop")
			silent, _ := cmd.Flags().GetBool("silent")
			configPath, _ := cmd.Flags().GetString("config")
			appCtx := appcontext.NewContext(debug, noOp, silent, configPath)
			cmd.SetContext(context.WithValue(cmd.Context(), appcontext.CtxKey, appCtx))
			return nil
		},
	}

	// Persistent flags
	cmd.PersistentFlags().Bool("debug", false, "Enable debug logging")
	cmd.PersistentFlags().Bool("noop", false, "Dry run, do not actually post or upload")
	cmd.PersistentFlags().Bool("silent", false, "Suppress informational messages")
	cmd.PersistentFlags().String("config", "", "Path to an alternative config file")

	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	rootCmd := newRootCmd()

	// Add child commands
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newProfileCmd())
	rootCmd.AddCommand(newPostCmd())
	rootCmd.AddCommand(newUploadCmd())
	rootCmd.AddCommand(newExportCmd())
	rootCmd.AddCommand(newChannelCmd())

	return rootCmd.Execute()
}
