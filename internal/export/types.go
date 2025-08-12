package export

// ExportedLog is the top-level structure for the exported log file.
type ExportedLog struct {
	ExportTimestamp string            `json:"export_timestamp"`
	ChannelName     string            `json:"channel_name"`
	Messages        []ExportedMessage `json:"messages"`
}

// ExportedMessage represents a single message in the exported log.
type ExportedMessage struct {
	UserID        string         `json:"user_id"`
	UserName      string         `json:"user_name,omitempty"`
	Timestamp     string         `json:"timestamp"`
	TimestampUnix string         `json:"timestamp_unix"`
	Text          string         `json:"text"`
	Files         []ExportedFile `json:"files,omitempty"`
}

// ExportedFile represents a file attached to a message in the exported log.
type ExportedFile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Mimetype  string `json:"mimetype"`
	LocalPath string `json:"local_path,omitempty"` // Path to the downloaded file
}

// Options defines the parameters for an export operation.
type Options struct {
	ChannelName  string
	StartTime    string
	EndTime      string
	IncludeFiles bool
	OutputDir    string
}
