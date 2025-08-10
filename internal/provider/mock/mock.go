package mock

import (
	"fmt"

	"github.com/magifd2/scat/internal/config"
	"github.com/magifd2/scat/internal/provider"
)

// Provider implements the provider.Interface for mocking.
type Provider struct {
	Profile config.Profile
	NoOp    bool
}

// NewProvider creates a new mock Provider.
func NewProvider(p config.Profile, noop bool) (provider.Interface, error) {
	return &Provider{Profile: p, NoOp: noop}, nil
}

// Capabilities returns the features supported by the mock provider.
func (p *Provider) Capabilities() provider.Capabilities {
	return provider.Capabilities{
		CanListChannels: false,
		CanPostFile:     true, // Mock can "handle" file posts
		CanUseThreads:   false,
		CanUseIconEmoji: false,
	}
}

// PostMessage prints a mock message.
func (p *Provider) PostMessage(text, overrideUsername, iconEmoji string, thread bool, threadTS string) (string, error) {
	fmt.Println("--- [MOCK] PostMessage called ---")
	fmt.Printf("Text: %s\n", text)
	return "", nil
}

// PostFile prints a mock message.
func (p *Provider) PostFile(filePath, filename, filetype, comment, overrideUsername, iconEmoji string, thread bool, threadTS string) (string, error) {
	fmt.Println("--- [MOCK] PostFile called ---")
	fmt.Printf("File: %s\n", filePath)
	return "", nil
}

// ListChannels returns an error as it's not supported.
func (p *Provider) ListChannels() ([]string, error) {
	return nil, fmt.Errorf("ListChannels is not supported by the mock provider")
}