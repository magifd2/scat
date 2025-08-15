package testprovider

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/magifd2/scat/internal/appcontext"
	"github.com/magifd2/scat/internal/config"
	"github.com/magifd2/scat/internal/export"
	"github.com/magifd2/scat/internal/provider"
)

var PostMessageSignal chan struct{}

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
		CanPostBlocks:   true, // New: Test provider supports posting Block Kit messages
	}
}

// PostMessage logs the message options to stderr.
func (p *Provider) PostMessage(opts provider.PostMessageOptions) error {
	if opts.Text == `{"test_command": "signal_done"}` {
		if PostMessageSignal != nil {
			PostMessageSignal <- struct{}{}
		}
		return nil
	}
	fmt.Fprintf(os.Stderr, "[TESTPROVIDER] PostMessage called with opts: {Text:%s OverrideUsername:%s IconEmoji:%s Blocks:%s}\n", opts.Text, opts.OverrideUsername, opts.IconEmoji, string(opts.Blocks))
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

// ExportLog logs the export options and returns dummy data that reflects the options.
func (p *Provider) ExportLog(opts export.Options) (*export.ExportedLog, error) {
	fmt.Fprintf(os.Stderr, "[TESTPROVIDER] ExportLog called with opts: %+v\n", opts)

	// Create a dummy message
	message := export.ExportedMessage{
		Text:      "Test message from ExportLog",
		UserName:  "testuser",
		Timestamp: "1672531200.000000", // 2023-01-01 00:00:00 UTC
	}

	// If file export is requested, add dummy file info
	if opts.IncludeFiles {
		message.Files = []export.ExportedFile{
			{
				ID:        "F12345678",
				Name:      "test-file.txt",
				Mimetype:  "text/plain",
				LocalPath: filepath.Join(opts.OutputDir, "test-file.txt"),
			},
		}
	}

	return &export.ExportedLog{
		ChannelName:     opts.ChannelName,
		ExportTimestamp: time.Now().UTC().Format(time.RFC3339),
		Messages:        []export.ExportedMessage{message},
	}, nil
}
