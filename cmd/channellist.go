package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/magifd2/scat/internal/client"
	"github.com/magifd2/scat/internal/config"
	"github.com/spf13/cobra"
)

var channelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available channels for Slack providers",
	Long:  `Iterates through all configured profiles and lists the available channels for each profile where the provider is set to "slack".`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		jsonOutput, _ := cmd.Flags().GetBool("json")
		results := make(map[string][]string)

		for profileName, profile := range cfg.Profiles {
			if profile.Provider == "slack" {
				apiClient := client.NewClient(profile, false) // noop is false for listing
				channels, err := apiClient.ListSlackChannels()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not list channels for profile '%s': %v\n", profileName, err)
					continue
				}
				if jsonOutput {
					results[profileName] = channels
				} else {
					fmt.Printf("Channels for profile: %s\n", profileName)
					for _, ch := range channels {
						fmt.Printf("  - %s\n", ch)
					}
				}
			}
		}

		if jsonOutput {
			jsonBytes, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal results to json: %w", err)
			}
			fmt.Println(string(jsonBytes))
		}

		return nil
	},
}

func init() {
	channelCmd.AddCommand(channelListCmd)
	channelListCmd.Flags().Bool("json", false, "Output the list of channels in JSON format")
}
