package cmd

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/magifd2/scat/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var setCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a value in the current profile",
	Long: `Set a configuration value for the currently active profile.
For the 'token' key, run 'scat profile set token' and you will be prompted to enter the value securely.`, 
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("Error loading config: %w", err)
		}

		profile, ok := cfg.Profiles[cfg.CurrentProfile]
		if !ok {
			return fmt.Errorf("active profile '%s' not found", cfg.CurrentProfile)
		}

		var value string
		if key == "token" {
			if len(args) != 1 {
				return fmt.Errorf("'set token' does not accept a value argument. Run it alone to be prompted.")
			}
			fmt.Print("Enter new Token (will not be displayed): ")
		
tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return fmt.Errorf("failed to read token: %w", err)
			}
			fmt.Println()
			value = string(tokenBytes)
		} else {
			if len(args) != 2 {
				return fmt.Errorf("key '%s' requires a value", key)
			}
			value = args[1]
		}

		switch key {
		case "provider":
			profile.Provider = value
		case "endpoint":
			profile.Endpoint = value
		case "channel":
			profile.Channel = value
		case "token":
			profile.Token = value
		case "username":
			profile.Username = value
		default:
			availableKeys := []string{"provider", "endpoint", "channel", "token", "username"}
			return fmt.Errorf("unknown configuration key '%s'.\nAvailable keys: %s", key, strings.Join(availableKeys, ", "))
		}

		cfg.Profiles[cfg.CurrentProfile] = profile
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Printf("Set %s in profile %s\n", key, cfg.CurrentProfile)
		return nil
	},
}

func init() {
	profileCmd.AddCommand(setCmd)
}