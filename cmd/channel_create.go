package cmd

import (
	"fmt"
	"os"

	"github.com/magifd2/scat/internal/appcontext"
	"github.com/magifd2/scat/internal/config"
	"github.com/spf13/cobra"
)

func newChannelCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [channel-name]",
		Short: "Create a new channel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			channelName := args[0]

			appCtx := cmd.Context().Value(appcontext.CtxKey).(appcontext.Context)
			configPath, err := config.GetConfigPath(appCtx.ConfigPath)
			if err != nil {
				return fmt.Errorf("failed to get config path: %w", err)
			}

			cfg, err := config.Load(configPath)
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("configuration file not found. Please run 'scat config init' to create a default configuration")
				}
				return fmt.Errorf("failed to load config: %w", err)
			}

			profileName, _ := cmd.Flags().GetString("profile")
			if profileName == "" {
				profileName = cfg.CurrentProfile
			}
			profile, ok := cfg.Profiles[profileName]
			if !ok {
				return fmt.Errorf("profile '%s' not found", profileName)
			}

			p, err := GetProvider(appCtx, profile)
			if err != nil {
				return fmt.Errorf("failed to get provider: %w", err)
			}

			if !p.Capabilities().CanCreateChannel {
				return fmt.Errorf("the provider for profile '%s' does not support creating channels", profileName)
			}

			channelID, err := p.CreateChannel(channelName)
			if err != nil {
				return fmt.Errorf("failed to create channel: %w", err)
			}

			if !appCtx.Silent {
				fmt.Printf("info: Successfully created channel with ID: %s\n", channelID)
			}

			return nil
		},
	}

	cmd.Flags().StringP("profile", "p", "", "Profile to use for this command")
	return cmd
}