
package testprovider

import (
	"fmt"
	"os"

	"github.com/magifd2/scat/internal/appcontext"
	"github.com/magifd2/scat/internal/config"
	"github.com/magifd2/scat/internal/export"
	"github.com/magifd2/scat/internal/provider"
)

// Provider implements the provider.Interface for testing purposes.
type Provider struct {
	Profile config.Profile
	Context appcontext.Context
}

// NewProvider creates a new test Provider.
func NewProvider(p config.Profile, ctx appcontext.Context) (provider.Interface, error) {
	fmt.Fprintf(os.Stderr, "[TESTPROVIDER] NewProvider called with profile: %s, context: %+v\n", p.Provider, ctx)
	return &Provider{Profile: p, Context: ctx},
		nil
}

// Capabilities returns the features supported by the test provider.
func (p *Provider) Capabilities() provider.Capabilities {
	fmt.Fprintf(os.Stderr, "[TESTPROVIDER] Capabilities called\n")
	// Default capabilities for a test provider
	return provider.Capabilities{
		CanListChannels: true,
		CanPostFile:     true,
		CanUseIconEmoji: true,
		CanExportLogs:   true,
	}
}

// PostMessage logs the message options to stderr.
func (p *Provider) PostMessage(opts provider.PostMessageOptions) error {
	fmt.Fprintf(os.Stderr, "[TESTPROVIDER] PostMessage called with opts: %+v\n", opts)
	return nil
}

// PostFile logs the file options to stderr.
func (p *Provider) PostFile(opts provider.PostFileOptions) error {
	fmt.Fprintf(os.Stderr, "[TESTPROVIDER] PostFile called with opts: %+v\n", opts)
	return nil
}

// ListChannels logs the call and returns dummy data.
func (p *Provider) ListChannels() ([]string, error) {
	fmt.Fprintf(os.Stderr, "[TESTPROVIDER] ListChannels called\n")
	return []string{"#test-channel-1", "#test-channel-2"}, nil
}

// ExportLog logs the export options and returns dummy data.
func (p *Provider) ExportLog(opts export.Options) (*export.ExportedLog, error) {
	fmt.Fprintf(os.Stderr, "[TESTPROVIDER] ExportLog called with opts: %+v\n", opts)
	return &export.ExportedLog{
		ChannelName: opts.ChannelName,
		Messages: []export.ExportedMessage{
			{Text: "Test message from ExportLog"},
		},
	},
	nil
}
