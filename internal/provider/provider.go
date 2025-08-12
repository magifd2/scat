package provider

import "github.com/magifd2/scat/internal/export"

// Capabilities defines what features a provider supports.
type Capabilities struct {
	CanListChannels bool // Whether the provider can list channels.
	CanPostFile     bool // Whether the provider can post files.
	CanUseIconEmoji bool // Whether the provider supports custom icon emojis.
	CanExportLogs   bool // Whether the provider can export channel logs.
}

// Interface defines the methods that a provider must implement.
type Interface interface {
	// Capabilities returns a struct indicating supported features.
	Capabilities() Capabilities

	// PostMessage sends a text-based message.
	PostMessage(text, overrideUsername, iconEmoji string) error

	// PostFile sends a file.
	PostFile(filePath, filename, filetype, comment, overrideUsername, iconEmoji string) error

	// ListChannels lists available channels for the provider.
	ListChannels() ([]string, error)

	// ExportLog performs the entire export operation.
	// This should only be called if Capabilities().CanExportLogs is true.
	ExportLog(opts export.Options) (*export.ExportedLog, error)
}
