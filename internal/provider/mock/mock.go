package mock

import (
	"fmt"
	"os"

	"github.com/magifd2/scat/internal/appcontext"
	"github.com/magifd2/scat/internal/config"
	"github.com/magifd2/scat/internal/provider"
)

// Provider implements the provider.Interface for mocking.
type Provider struct {
	Profile config.Profile
	Context appcontext.Context // Use appcontext.Context
}

// NewProvider creates a new mock Provider.
func NewProvider(p config.Profile, ctx appcontext.Context) (provider.Interface, error) {
	return &Provider{Profile: p, Context: ctx}, nil
}

// Capabilities returns the features supported by the mock provider.
func (p *Provider) Capabilities() provider.Capabilities {
	return provider.Capabilities{
		CanListChannels: false,
		CanPostFile:     true, // Mock can "handle" file posts
		CanUseIconEmoji: false,
	}
}

// PostMessage prints a mock message.
func (p *Provider) PostMessage(text, overrideUsername, iconEmoji string) error {
	if p.Context.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Mock PostMessage: Text=\"%s\", Username=\"%s\", IconEmoji=\"%s\"\n", text, overrideUsername, iconEmoji)
	}
	fmt.Println("---" + " [MOCK] PostMessage called ---")
	fmt.Printf("Text: %s\n", text)
	return nil
}

// PostFile prints a mock message.
func (p *Provider) PostFile(filePath, filename, filetype, comment, overrideUsername, iconEmoji string) error {
	if p.Context.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Mock PostFile: FilePath=\"%s\", Filename=\"%s\"\n", filePath, filename)
	}
	fmt.Println("---" + " [MOCK] PostFile called ---")
	fmt.Printf("File: %s\n", filePath)
	return nil
}

// ListChannels returns an error as it's not supported.
func (p *Provider) ListChannels() ([]string, error) {
	if p.Context.Debug {
		fmt.Fprintln(os.Stderr, "[DEBUG] Mock ListChannels called")
	}
	return nil, fmt.Errorf("ListChannels is not supported by the mock provider")
}
