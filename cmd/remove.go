package cmd

import (
	"fmt"

	"github.com/magifd2/scat/internal/config"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [profile_name]",
	Short: "Remove a profile",
	Long:  `Removes a specified profile from the configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]

		if profileName == "default" {
			return fmt.Errorf("the 'default' profile cannot be removed")
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if _, ok := cfg.Profiles[profileName]; !ok {
			return fmt.Errorf("profile '%s' not found", profileName)
		}

		if cfg.CurrentProfile == profileName {
			return fmt.Errorf("cannot remove the currently active profile. Please switch to another profile first")
		}

		delete(cfg.Profiles, profileName)

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Printf("Profile '%s' removed.\n", profileName)
		return nil
	},
}

func init() {
	profileCmd.AddCommand(removeCmd)
}
