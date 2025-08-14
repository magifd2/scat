package mock

import (
	"fmt"
	"os"
	"time"

	"github.com/magifd2/scat/internal/appcontext"
	"github.com/magifd2/scat/internal/config"
	"github.com/magifd2/scat/internal/export"
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
		CanExportLogs:   true, // Mock supports exporting for testing purposes
	}
}

// PostMessage prints a mock message.
func (p *Provider) PostMessage(opts provider.PostMessageOptions) error {
	if !p.Context.Silent {
		fmt.Fprintln(os.Stderr, "--- [MOCK] PostMessage called ---")
		fmt.Fprintf(os.Stderr, "Text: %s\n", opts.Text)
	}
	if p.Context.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Mock PostMessage: Text=\"%s\", Username=\"%s\", IconEmoji=\"%s\"\n", opts.Text, opts.OverrideUsername, opts.IconEmoji)
	}
	return nil
}

// PostFile prints a mock message.
func (p *Provider) PostFile(opts provider.PostFileOptions) error {
	if !p.Context.Silent {
		fmt.Fprintln(os.Stderr, "--- [MOCK] PostFile called ---")
		fmt.Fprintf(os.Stderr, "File: %s\n", opts.FilePath)
	}
	if p.Context.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Mock PostFile: FilePath=\"%s\", Filename=\"%s\"\n", opts.FilePath, opts.Filename)
	}
	return nil
}

// ListChannels returns an error as it's not supported.
func (p *Provider) ListChannels() ([]string, error) {
	return nil, fmt.Errorf("ListChannels is not supported by the mock provider")
}

// ExportLog returns a dummy log for testing.
func (p *Provider) ExportLog(opts export.Options) (*export.ExportedLog, error) {
	if !p.Context.Silent {
		fmt.Fprintf(os.Stderr, "--- [MOCK] ExportLog called for channel %s ---\n", opts.ChannelName)
	}
	return &export.ExportedLog{
		ExportTimestamp: time.Now().UTC().Format(time.RFC3339),
		ChannelName:     opts.ChannelName,
		Messages: []export.ExportedMessage{
			{
				UserID:        "U012AB3CDE",
				UserName:      "Mock User",
				Timestamp:     time.Now().UTC().Format(time.RFC3339),
				TimestampUnix: fmt.Sprintf("%d.000000", time.Now().Unix()),
				Text:          "Hello from mock exporter!",
			},
		},
	}, nil
}
