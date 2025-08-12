package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
		startTimeStr, _ := cmd.Flags().GetString("start-time")
		endTimeStr, _ := cmd.Flags().GetString("end-time")
		includeFiles, _ := cmd.Flags().GetBool("include-files")
		outputDir, _ := cmd.Flags().GetString("output-dir")
		outputFormat, _ := cmd.Flags().GetString("output-format")

		// Parse timestamps
		startTime, err := parseTime(startTimeStr)
		if err != nil {
			return fmt.Errorf("invalid start time: %w", err)
		}
		endTime, err := parseTime(endTimeStr)
		if err != nil {
			return fmt.Errorf("invalid end time: %w", err)
		}

		if !appCtx.Silent {
			fmt.Fprintf(os.Stderr, "> Exporting messages from %s to %s (UTC: %s to %s)\n",
				displayTime(startTime), displayTime(endTime),
				startTime.UTC().Format(time.RFC3339), endTime.UTC().Format(time.RFC3339))
		}

		if outputDir == "" {
			outputDir = fmt.Sprintf("./scat-export-%s", time.Now().UTC().Format("20060102T150405Z"))
		}

		// Create exporter and run
		exporter := export.NewExporter(prov.LogExporter())
		opts := export.Options{
			ChannelID:    channel,
			StartTime:    startTimeStr,
			EndTime:      endTimeStr,
			IncludeFiles: includeFiles,
			OutputDir:    outputDir,
			OutputFormat: outputFormat,
		}

		exportedLog, err := exporter.ExportLog(opts)
		if err != nil {
			return fmt.Errorf("failed to export log: %w", err)
		}

		// Save the log to a file
		logFileName := fmt.Sprintf("export-%s-%s.json", exportedLog.ChannelName, time.Now().UTC().Format("20060102T150405Z"))
		logFilePath := filepath.Join(outputDir, logFileName)
		logFile, err := os.Create(logFilePath)
		if err != nil {
			return fmt.Errorf("failed to create log file: %w", err)
		}
		defer logFile.Close()

		encoder := json.NewEncoder(logFile)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(exportedLog); err != nil {
			return fmt.Errorf("failed to write log file: %w", err)
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

// parseTime parses a string into a time.Time object.
// It accepts RFC3339 format or a local time format.
func parseTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, nil
	}
	// Try parsing with timezone first
	t, err := time.Parse(time.RFC3339, timeStr)
	if err == nil {
		return t, nil
	}
	// If that fails, try parsing as local time
	return time.Parse("2006-01-02T15:04:05", timeStr)
}

// displayTime formats a time for display, showing timezone info.
func displayTime(t time.Time) string {
	if t.IsZero() {
		return "(not set)"
	}
	return t.Format(time.RFC3339)
}