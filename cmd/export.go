package cmd

import (
	"github.com/spf13/cobra"
)

// newExportCmd creates the command for exporting data.
func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export data from a provider",
		Long:  `The export command and its subcommands allow you to export data, such as channel logs, from a supported provider.`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(newExportLogCmd()) // from export_log.go

	return cmd
}