package cmd

import (
	"fmt"
	"os"

	"github.com/magifd2/scat/internal/config"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use [profile_name]",
	Short: "Set the active profile",
	Long:  `Set the active profile for scat.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		cfg, err := config.Load()
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("configuration file not found. Please run 'scat config init' to create a default configuration")
			}
			return fmt.Errorf("loading config: %w", err)
		}

		if _, ok := cfg.Profiles[profileName]; !ok {
			return fmt.Errorf("profile '%s' not found", profileName)
		}

		cfg.CurrentProfile = profileName
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Switched to profile: %s\n", profileName)
		return nil
	},
}

func init() {
	profileCmd.AddCommand(useCmd)
}
