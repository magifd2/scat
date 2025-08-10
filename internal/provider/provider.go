package provider

// Capabilities defines what features a provider supports.
type Capabilities struct {
	CanListChannels bool // Whether the provider can list channels.
	CanPostFile     bool // Whether the provider can post files.
	CanUseThreads   bool // Whether the provider supports threaded messages.
	CanUseIconEmoji bool // Whether the provider supports custom icon emojis.
}

// Interface defines the methods that a provider must implement.
type Interface interface {
	// Capabilities returns a struct indicating supported features.
	Capabilities() Capabilities

	// PostMessage sends a text-based message.
	PostMessage(text, overrideUsername, iconEmoji string, thread bool, threadTS string) (string, error)

	// PostFile sends a file.
	PostFile(filePath, filename, filetype, comment, overrideUsername, iconEmoji string, thread bool, threadTS string) (string, error)

	// ListChannels lists available channels for the provider.
	ListChannels() ([]string, error)
}