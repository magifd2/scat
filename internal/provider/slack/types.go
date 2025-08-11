package slack

// --- API Payload and Response Structs ---

type messagePayload struct {
	Channel   string `json:"channel"` // This must be a Channel ID
	Text      string `json:"text"`
	Username  string `json:"username,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
}

type apiResponse struct {
	Ok               bool   `json:"ok"`
	Error            string `json:"error,omitempty"`
	Channels         []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"channels"`
	ResponseMetadata struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
}

type getUploadURLExternalResponse struct {
	Ok        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
	UploadURL string `json:"upload_url"`
	FileID    string `json:"file_id"`
}

type fileInfo struct {
	ID string `json:"id"`
}

type completeUploadExternalPayload struct {
	Files          []fileInfo `json:"files"`
	ChannelID      string     `json:"channel_id,omitempty"`
	InitialComment string     `json:"initial_comment,omitempty"`
}