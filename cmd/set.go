package cmd

import (
	"fmt"
	"strings"

	"github.com/magifd2/scat/internal/config"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a value in the current profile",
	Long:  `Set a configuration value for the currently active profile.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("Error loading config: %w", err)
		}

		profile, ok := cfg.Profiles[cfg.CurrentProfile]
		if !ok {
			return fmt.Errorf("active profile '%s' not found", cfg.CurrentProfile)
		}

		switch key {
		case "endpoint":
			profile.Endpoint = value
		case "token":
			profile.Token = value
		case "username":
			profile.Username = value
		default:
			availableKeys := []string{"endpoint", "token", "username"}
			return fmt.Errorf("unknown configuration key '%s'.\nAvailable keys: %s", key, strings.Join(availableKeys, ", "))
		}

		cfg.Profiles[cfg.CurrentProfile] = profile
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Printf("Set %s = %s in profile %s\n", key, value, cfg.CurrentProfile)
		return nil
	},
}

func init() {
	profileCmd.AddCommand(setCmd)
}
