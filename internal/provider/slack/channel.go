package slack

import (
	"fmt"
	"os"
	"strings"
)

// ResolveChannelID ensures a channel ID is returned for a given name.
// It first checks the local cache. If the name is not found, it refreshes
// the cache from the API and checks again.
func (p *Provider) ResolveChannelID(name string) (string, error) {
	// First, try to get the ID from the existing cache.
	id, err := p.getCachedChannelID(name)
	if err == nil {
		return id, nil // Found in cache
	}

	// If not found, refresh the cache from the API.
	if p.Context.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Channel \"%s\" not in cache. Refreshing...\n", name)
	}
	if refreshErr := p.populateChannelCache(); refreshErr != nil {
		return "", fmt.Errorf("failed to refresh channel list: %w", refreshErr)
	}

	// Try checking the cache again after refreshing.
	id, err = p.getCachedChannelID(name)
	if err == nil {
		return id, nil // Found after refresh
	}

	// If it's still not found, the channel likely doesn't exist.
	return "", fmt.Errorf("channel \"%s\" not found after refreshing cache", name)
}

// getCachedChannelID is a helper that only checks the local cache.
func (p *Provider) getCachedChannelID(name string) (string, error) {
	name = strings.TrimPrefix(name, "#")

	if p.channelIDCache != nil {
		if id, ok := p.channelIDCache[name]; ok {
			return id, nil
		}
	}

	// Also consider the case where the name is already a valid ID.
	if strings.HasPrefix(name, "C") || strings.HasPrefix(name, "G") || strings.HasPrefix(name, "D") {
		return name, nil
	}

	return "", fmt.Errorf("not found in cache")
}

func (p *Provider) ListChannels() ([]string, error) {
	// Ensure the cache is populated before listing.
	if p.channelIDCache == nil {
		if err := p.populateChannelCache(); err != nil {
			return nil, err
		}
	}
	var channelNames []string
	for name := range p.channelIDCache {
		channelNames = append(channelNames, "#"+name)
	}
	return channelNames, nil
}