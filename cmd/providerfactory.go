package cmd

import (
	"fmt"

	"github.com/magifd2/scat/internal/appcontext"
	"github.com/magifd2/scat/internal/config"
	"github.com/magifd2/scat/internal/provider"
	"github.com/magifd2/scat/internal/provider/mock"
	"github.com/magifd2/scat/internal/provider/slack"
)

// GetProvider returns a provider instance based on the provider name in the profile.
func GetProvider(ctx appcontext.Context, p config.Profile) (provider.Interface, error) {
	switch p.Provider {
	case "slack":
		return slack.NewProvider(p, ctx)
	case "mock":
		return mock.NewProvider(p, ctx)
	default:
		return nil, fmt.Errorf("unknown provider: '%s'", p.Provider)
	}
}