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
		CanExportLogs:   true, // Mock supports exporting for testing purposes
	}
}

// LogExporter returns the log exporter implementation for the mock provider.
func (p *Provider) LogExporter() provider.LogExporter {
	return p
}

// PostMessage prints a mock message.
func (p *Provider) PostMessage(text, overrideUsername, iconEmoji string) error {
	if !p.Context.Silent {
		fmt.Fprintln(os.Stderr, "--- [MOCK] PostMessage called ---")
		fmt.Fprintf(os.Stderr, "Text: %s\n", text)
	}
	if p.Context.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Mock PostMessage: Text=\"%s\", Username=\"%s\", IconEmoji=\"%s\"\n", text, overrideUsername, iconEmoji)
	}
	return nil
}

// PostFile prints a mock message.
func (p *Provider) PostFile(filePath, filename, filetype, comment, overrideUsername, iconEmoji string) error {
	if !p.Context.Silent {
		fmt.Fprintln(os.Stderr, "--- [MOCK] PostFile called ---")
		fmt.Fprintf(os.Stderr, "File: %s\n", filePath)
	}
	if p.Context.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Mock PostFile: FilePath=\"%s\", Filename=\"%s\"\n", filePath, filename)
	}
	return nil
}

// ListChannels returns an error as it's not supported.
func (p *Provider) ListChannels() ([]string, error) {
	if !p.Context.Silent {
		fmt.Fprintln(os.Stderr, "--- [MOCK] ListChannels called ---")
	}
	if p.Context.Debug {
		fmt.Fprintln(os.Stderr, "[DEBUG] Mock ListChannels called")
	}
	return nil, fmt.Errorf("ListChannels is not supported by the mock provider")
}

// --- LogExporter Methods ---

func (p *Provider) GetConversationHistory(opts provider.GetConversationHistoryOptions) (*provider.ConversationHistoryResponse, error) {
	if !p.Context.Silent {
		fmt.Fprintf(os.Stderr, "--- [MOCK] GetConversationHistory called for channel %s ---\\n", opts.ChannelID)
	}
	// Return a dummy response for testing
	resp := &provider.ConversationHistoryResponse{
		Messages: []provider.Message{
			{Type: "message", Timestamp: "1672531200.000000", UserID: "U012AB3CDE", Text: "Hello from mock"},
		},
		HasMore:    false,
		NextCursor: "",
	}
	return resp, nil
}

func (p *Provider) GetUserInfo(userID string) (*provider.UserInfoResponse, error) {
	if !p.Context.Silent {
		fmt.Fprintln(os.Stderr, "--- [MOCK] GetUserInfo called ---")
	}
	resp := &provider.UserInfoResponse{
		User: provider.User{
			ID:       userID,
			Name:     fmt.Sprintf("mockuser_%s", userID),
			RealName: fmt.Sprintf("Mock User %s", userID),
		},
	}
	return resp, nil
}

func (p *Provider) DownloadFile(fileURL string) ([]byte, error) {
	if !p.Context.Silent {
		fmt.Fprintln(os.Stderr, "--- [MOCK] DownloadFile called ---")
	}
	return []byte("mock file content for " + fileURL), nil
}