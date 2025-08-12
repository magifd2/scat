package slack

import "github.com/magifd2/scat/internal/provider"

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

// conversationsHistoryResponse corresponds to the JSON response from conversations.history API
type conversationsHistoryResponse struct {
	Ok               bool                `json:"ok"`
	Messages         []provider.Message  `json:"messages"`
	HasMore          bool                `json:"has_more"`
	ResponseMetadata provider.ResponseMetadata `json:"response_metadata"`
	Error            string              `json:"error,omitempty"`
}

// userInfoResponse corresponds to the JSON response from users.info API
type userInfoResponse struct {
	Ok    bool           `json:"ok"`
	User  provider.User  `json:"user"`
	Error string         `json:"error,omitempty"`
}