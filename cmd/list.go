package cmd

import (
	"fmt"
	"os"

	"github.com/magifd2/scat/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available profiles",
	Long:  `Lists all saved profiles and indicates which one is currently active.`, 
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Available profiles:")
		for name, p := range cfg.Profiles {
			activeMarker := " "
			if name == cfg.CurrentProfile {
				activeMarker = "*"
			}
			fmt.Printf("  %s %s (provider: %s", activeMarker, name, p.Provider)
			if p.Provider == "slack" {
				fmt.Printf(", channel: %s)\n", p.Channel)
			} else {
				fmt.Printf(", endpoint: %s)\n", p.Endpoint)
			}
		}
	},
}

func init() {
	profileCmd.AddCommand(listCmd)
}