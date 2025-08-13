package cmd

import (
	"fmt"

	"github.com/magifd2/scat/internal/appcontext"
	"github.com/magifd2/scat/internal/config"
	"github.com/magifd2/scat/internal/provider"
	"github.com/magifd2/scat/internal/provider/mock"
	"github.com/magifd2/scat/internal/provider/slack"
	"github.com/magifd2/scat/internal/provider/testprovider"
)

// providerFactory defines the function signature for creating a new provider.Interface.
type providerFactory func(p config.Profile, ctx appcontext.Context) (provider.Interface, error)

// providerRegistry holds the mapping from a provider name to its factory function.
var providerRegistry = map[string]providerFactory{
	"slack": slack.NewProvider,
	"mock":  mock.NewProvider,
	"test":  testprovider.NewProvider,
}

// GetProvider retrieves a provider instance based on the provider name in the profile.
func GetProvider(ctx appcontext.Context, p config.Profile) (provider.Interface, error) {
	factory, ok := providerRegistry[p.Provider]
	if !ok {
		return nil, fmt.Errorf("unknown provider: '%s'", p.Provider)
	}
	return factory(p, ctx)
}
