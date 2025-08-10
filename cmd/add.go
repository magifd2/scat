package cmd

import (
	"fmt"
	"syscall"

	"github.com/magifd2/scat/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var addCmd = &cobra.Command{
	Use:   "add [profile_name]",
	Short: "Add a new profile",
	Long:  `Adds a new profile. You will be prompted to enter the authentication token securely.`, 
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

		provider, _ := cmd.Flags().GetString("provider")
		endpoint, _ := cmd.Flags().GetString("endpoint")
		channel, _ := cmd.Flags().GetString("channel")
		username, _ := cmd.Flags().GetString("username")

		newProfile := config.Profile{
			Provider: provider,
			Endpoint: endpoint,
			Channel:  channel,
			Username: username,
		}

		// Prompt for token securely
		fmt.Print("Enter Token (will not be displayed): ")
	
tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read token: %w", err)
		}
		fmt.Println()
		newProfile.Token = string(tokenBytes)

		cfg.Profiles[profileName] = newProfile

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("Error saving config: %w", err)
		}

		fmt.Printf("Profile '%s' added.\n", profileName)
		return nil
	},
}

func init() {
	profileCmd.AddCommand(addCmd)

	addCmd.Flags().String("provider", "generic", "Provider type: 'generic' or 'slack'")
	addCmd.Flags().String("endpoint", "", "API endpoint URL (for generic provider)")
	addCmd.Flags().String("channel", "", "Channel name (for slack provider)")
	addCmd.Flags().String("username", "", "Default username for posts")
}