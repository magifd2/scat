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
		CanPostBlocks:   true,
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

	// Note: No complex logic for channel/user resolution in test provider.
	// We just log the raw options to verify that the command layer is sending them correctly.
	fmt.Fprintf(os.Stderr, "[TESTPROVIDER] PostMessage called with opts: {TargetChannel:%s TargetUserID:%s Text:%s OverrideUsername:%s IconEmoji:%s Blocks:%s}\n", opts.TargetChannel, opts.TargetUserID, opts.Text, opts.OverrideUsername, opts.IconEmoji, string(opts.Blocks))
	return nil
}

// PostFile logs the file options to stderr.
func (p *Provider) PostFile(opts provider.PostFileOptions) error {
	// Create a temporary struct for logging that includes all relevant fields.
	logOpts := struct {
		TargetChannel    string
		TargetUserID     string
		FilePath         string
		Filename         string
		Filetype         string
		Comment          string
		OverrideUsername string
		IconEmoji        string
	}{
		TargetChannel:    opts.TargetChannel,
		TargetUserID:     opts.TargetUserID,
		FilePath:         opts.FilePath,
		Filename:         opts.Filename,
		Filetype:         opts.Filetype,
		Comment:          opts.Comment,
		OverrideUsername: opts.OverrideUsername,
		IconEmoji:        opts.IconEmoji,
	}

	fmt.Fprintf(os.Stderr, "[TESTPROVIDER] PostFile called with opts: %+v\n", logOpts)
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
