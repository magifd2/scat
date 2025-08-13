package cmd

import (
	"fmt"
	"os"

	"github.com/magifd2/scat/internal/appcontext"
	"github.com/magifd2/scat/internal/config"
	"github.com/spf13/cobra"
)

func newConfigInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new configuration file",
		Long:  `Creates a new default configuration file at the default location (~/.config/scat/config.json). If a configuration file already exists, this command will do nothing.`, 
		RunE: func(cmd *cobra.Command, args []string) error {
			appCtx := cmd.Context().Value(appcontext.CtxKey).(appcontext.Context)
			configPath, err := config.GetConfigPath(appCtx.ConfigPath)
			if err != nil {
				return fmt.Errorf("failed to get config path: %w", err)
			}

			// Guardrail: Check if the config file already exists.
			_, err = os.Stat(configPath)
			if err == nil {
				fmt.Printf("Configuration file already exists at: %s\n", configPath)
				return nil // Not an error, just informational.
			}
			if !os.IsNotExist(err) {
				// For other errors (e.g., permission denied), return the error.
				return fmt.Errorf("failed to check for existing config file: %w", err)
			}

			// File does not exist, so create a new one.
			cfg := config.NewDefaultConfig()
			if err := cfg.Save(configPath); err != nil {
				return fmt.Errorf("failed to save new configuration file: %w", err)
			}

			fmt.Printf("Successfully created a new configuration file at: %s\n", configPath)
			return nil
		},
	}
	return cmd
}