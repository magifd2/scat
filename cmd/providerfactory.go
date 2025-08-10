package cmd

import (
	"fmt"

	"github.com/magifd2/scat/internal/config"
	"github.com/magifd2/scat/internal/provider"
	"github.com/magifd2/scat/internal/provider/mock"
	"github.com/magifd2/scat/internal/provider/slack"
)

// GetProvider returns a provider instance based on the provider name in the profile.
func GetProvider(p config.Profile, noop bool) (provider.Interface, error) {
	switch p.Provider {
	case "slack":
		return slack.NewProvider(p, noop)
	case "mock":
		return mock.NewProvider(p, noop)
	case "generic":
		// The user decided to remove the generic provider, but we can keep the case
		// and point it to mock to avoid breaking old configs.
		return mock.NewProvider(p, noop)
	default:
		return nil, fmt.Errorf("unknown provider: %s", p.Provider)
	}
}
