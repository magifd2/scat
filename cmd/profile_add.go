package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/magifd2/scat/internal/appcontext"
	"github.com/magifd2/scat/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// newProfileAddCmd creates the command for adding a new profile.
func newProfileAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [profile_name]",
		Short: "Add a new profile",
		Long:  `Adds a new profile. You will be prompted to enter the authentication token securely.`, 
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appCtx := cmd.Context().Value(appcontext.CtxKey).(appcontext.Context)
			configPath, err := config.GetConfigPath(appCtx.ConfigPath)
			if err != nil {
				return fmt.Errorf("failed to get config path: %w", err)
			}

			profileName := args[0]

			cfg, err := config.Load(configPath)
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("configuration file not found. Please run 'scat config init' to create a default configuration before adding a profile")
				}
				return fmt.Errorf("Error loading config: %w", err)
			}

			if _, ok := cfg.Profiles[profileName]; ok {
				return fmt.Errorf("Error: Profile '%s' already exists", profileName)
			}

			provider, _ := cmd.Flags().GetString("provider")
			channel, _ := cmd.Flags().GetString("channel")
			username, _ := cmd.Flags().GetString("username")
			maxFile, _ := cmd.Flags().GetInt64("limits-max-file-size-bytes")
			maxStdin, _ := cmd.Flags().GetInt64("limits-max-stdin-size-bytes")

			newProfile := config.Profile{
				Provider: provider,
				Channel:  channel,
				Username: username,
				Limits: config.Limits{
					MaxFileSizeBytes: maxFile,
					MaxStdinSizeBytes: maxStdin,
				},
			}

			// Prompt for token securely
			fmt.Fprint(os.Stderr, "Enter Token (will not be displayed): ")
			tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return fmt.Errorf("failed to read token: %w", err)
			}
			fmt.Fprintln(os.Stderr)
			newProfile.Token = string(tokenBytes)

			cfg.Profiles[profileName] = newProfile

			if err := cfg.Save(configPath); err != nil {
				return fmt.Errorf("Error saving config: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Profile '%s' added.\n", profileName)
			return nil
		},
	}

	cmd.Flags().String("provider", "slack", "Provider type: 'slack' or 'mock'")
	cmd.Flags().String("channel", "", "Channel name or ID (for slack provider)")
	cmd.Flags().String("username", "", "Default username for posts")
	cmd.Flags().Int64("limits-max-file-size-bytes", 1024*1024*1024, "Max file size for uploads in bytes (1GB)")
	cmd.Flags().Int64("limits-max-stdin-size-bytes", 10*1024*1024, "Max size for stdin in bytes (10MB)")

	return cmd
}