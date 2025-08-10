package cmd

import (
	"fmt"

	"github.com/magifd2/scat/internal/config"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [profile_name]",
	Short: "Add a new profile",
	Long:  `Adds a new profile. If no specific parameters are provided, it copies settings from the default profile.`, 
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("Error loading config: %w", err)
		}

		if _, ok := cfg.Profiles[profileName]; ok {
			return fmt.Errorf("Error: Profile '%s' already exists", profileName)
		}

		newProfile := config.Profile{}
		if !cmd.Flags().Changed("endpoint") && !cmd.Flags().Changed("token") && !cmd.Flags().Changed("username") {
			defaultProfile, ok := cfg.Profiles["default"]
			if !ok {
				return fmt.Errorf("Error: Default profile not found. Cannot create new profile without parameters.")
			}
			newProfile = defaultProfile
		} else {
			endpoint, _ := cmd.Flags().GetString("endpoint")
			token, _ := cmd.Flags().GetString("token")
			username, _ := cmd.Flags().GetString("username")

			newProfile.Endpoint = endpoint
			newProfile.Token = token
			newProfile.Username = username
		}

		cfg.Profiles[profileName] = newProfile

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("Error saving config: %w", err)
		}

		fmt.Printf("Profile '%s' added.\n", profileName)
		fmt.Printf("To switch to the new profile, run: scat profile use %s\n", profileName)
		return nil
	},
}

func init() {
	profileCmd.AddCommand(addCmd)

	addCmd.Flags().String("endpoint", "", "API endpoint URL")
	addCmd.Flags().String("token", "", "Authentication token")
	addCmd.Flags().String("username", "", "Default username for posts")
}
