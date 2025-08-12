package cmd

import (
	"fmt"
	"os"

	"github.com/magifd2/scat/internal/appcontext"
	"github.com/magifd2/scat/internal/config"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [profile_name]",
	Short: "Remove a profile",
	Long:  `Removes a specified profile from the configuration.`, 
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appCtx := cmd.Context().Value(appcontext.CtxKey).(appcontext.Context)
		configPath, err := config.GetConfigPath(appCtx.ConfigPath)
		if err != nil {
			return fmt.Errorf("failed to get config path: %w", err)
		}

		profileName := args[0]

		if profileName == "default" {
			return fmt.Errorf("the 'default' profile cannot be removed")
		}

		cfg, err := config.Load(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("configuration file not found. Please run 'scat config init' to create a default configuration")
			}
			return fmt.Errorf("loading config: %w", err)
		}

		if _, ok := cfg.Profiles[profileName]; !ok {
			return fmt.Errorf("profile '%s' not found", profileName)
		}

		if cfg.CurrentProfile == profileName {
			return fmt.Errorf("cannot remove the currently active profile. Please switch to another profile first")
		}

		delete(cfg.Profiles, profileName)

		if err := cfg.Save(configPath); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Profile '%s' removed.\n", profileName)
		return nil
	},
}


func init() {
	profileCmd.AddCommand(removeCmd)
}