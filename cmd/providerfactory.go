package cmd

import (
	"fmt"

	"github.com/magifd2/scat/internal/config"
	"github.com/magifd2/scat/internal/provider"
	"github.com/magifd2/scat/internal/provider/mock"
	"github.com/magifd2/scat/internal/provider/slack"
	"github.com/spf13/cobra"
)

// GetProvider returns a provider instance based on the provider name in the profile.
func GetProvider(cmd *cobra.Command, p config.Profile, noop bool) (provider.Interface, error) {
	debug, _ := cmd.PersistentFlags().GetBool("debug")

	switch p.Provider {
	case "slack":
		return slack.NewProvider(p, noop, debug)
	case "mock":
		return mock.NewProvider(p, noop, debug)
	default:
		return nil, fmt.Errorf("unknown provider: '%s'", p.Provider)
	}
}