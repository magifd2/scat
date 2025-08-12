package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/magifd2/scat/internal/appcontext"
	"github.com/magifd2/scat/internal/config"
	"github.com/magifd2/scat/internal/export"
	"github.com/spf13/cobra"
)

var exportLogCmd = &cobra.Command{
	Use:   "log",
	Short: "Export a channel log",
	Long:  `Exports a channel log from a supported provider, saving messages and optionally files to a local directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		appCtx := cmd.Context().Value(appcontext.CtxKey).(appcontext.Context)

		// Load config
		cfg, err := config.Load()
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("configuration file not found. Please run 'scat config init' to create a default configuration")
			}
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Determine profile
		profileName, _ := cmd.Flags().GetString("profile")
		if profileName == "" {
			profileName = cfg.CurrentProfile
		}
		profile, ok := cfg.Profiles[profileName]
		if !ok {
			return fmt.Errorf("profile '%s' not found", profileName)
		}

		// Get provider
		prov, err := GetProvider(appCtx, profile)
		if err != nil {
			return err
		}

		// Check capability
		if !prov.Capabilities().CanExportLogs {
			return fmt.Errorf("the provider for profile '%s' does not support exporting logs", profileName)
		}

		// Get flags
		channel, _ := cmd.Flags().GetString("channel")
		startTime, _ := cmd.Flags().GetString("start-time")
		endTime, _ := cmd.Flags().GetString("end-time")
		includeFiles, _ := cmd.Flags().GetBool("include-files")
		outputDir, _ := cmd.Flags().GetString("output-dir")
		outputFormat, _ := cmd.Flags().GetString("output-format")

		// TODO: Add timestamp parsing and status message printing here

		if outputDir == "" {
			outputDir = fmt.Sprintf("./scat-export-%s", time.Now().UTC().Format("20060102T150405Z"))
		}

		// Create exporter and run
		exporter := export.NewExporter(prov.LogExporter())
		opts := export.Options{
			ChannelID:    channel, // Note: This is the channel NAME, it will be resolved to ID inside the exporter
			StartTime:    startTime,
			EndTime:      endTime,
			IncludeFiles: includeFiles,
			OutputDir:    outputDir,
			OutputFormat: outputFormat,
		}

		_, err = exporter.ExportLog(opts)
		if err != nil {
			return fmt.Errorf("failed to export log: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Log export completed successfully. Files saved in %s\n", outputDir)

		return nil
	},
}

func init() {
	exportCmd.AddCommand(exportLogCmd)

	exportLogCmd.Flags().StringP("profile", "p", "", "Profile to use for this export")
	exportLogCmd.Flags().StringP("channel", "c", "", "Channel to export from (required)")
	exportLogCmd.MarkFlagRequired("channel")

	exportLogCmd.Flags().String("output-format", "json", "Output format (json or text)")
	exportLogCmd.Flags().String("start-time", "", "Start of time range (RFC3339 format, e.g., 2023-01-01T15:04:05Z)")
	exportLogCmd.Flags().String("end-time", "", "End of time range (RFC3339 format)")
	exportLogCmd.Flags().Bool("include-files", false, "Download attached files")
	exportLogCmd.Flags().String("output-dir", "", "Directory to save exported files (defaults to ./scat-export-<UTC-timestamp>/)")
}
