package slack

// This file defines the Go structs that directly map to the JSON responses
// from the Slack API. These are internal to the slack provider.

// apiResponse is a generic response for checking `ok` status and errors.
type apiResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// messagePayload is the structure for sending a message.
type messagePayload struct {
	Channel   string `json:"channel"`
	Text      string `json:"text"`
	Username  string `json:"username,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
}

// conversationsListResponse corresponds to the JSON from conversations.list API
type conversationsListResponse struct {
	Ok               bool      `json:"ok"`
	Error            string    `json:"error,omitempty"`
	Channels         []channel `json:"channels"`
	ResponseMetadata metadata  `json:"response_metadata"`
}

type channel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// conversationsHistoryResponse corresponds to the JSON from conversations.history API
type conversationsHistoryResponse struct {
	Ok               bool      `json:"ok"`
	Error            string    `json:"error,omitempty"`
	Messages         []message `json:"messages"`
	HasMore          bool      `json:"has_more"`
	ResponseMetadata metadata  `json:"response_metadata"`
}

// message represents a single message object from the Slack API.
type message struct {
	Type      string `json:"type"`
	Timestamp string `json:"ts"`
	UserID    string `json:"user"`
	Text      string `json:"text"`
	Files     []file `json:"files,omitempty"`
}

// file represents a file object from the Slack API.
type file struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Mimetype           string `json:"mimetype"`
	URLPrivateDownload string `json:"url_private_download"`
}

// userInfoResponse corresponds to the JSON from users.info API
type userInfoResponse struct {
	Ok    bool   `json:"ok"`
	User  user   `json:"user"`
	Error string `json:"error,omitempty"`
}

// user represents a user object from the Slack API.
type user struct {
	ID       string `json:"id"`
	TeamID   string `json:"team_id"`
	Name     string `json:"name"`
	RealName string `json:"real_name"`
}

// metadata contains pagination information.
type metadata struct {
	NextCursor string `json:"next_cursor"`
}

// getUploadURLExternalResponse corresponds to the JSON from files.getUploadURLExternal API
type getUploadURLExternalResponse struct {
	Ok        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
	UploadURL string `json:"upload_url"`
	FileID    string `json:"file_id"`
}

// fileInfo is used in the payload for completing a file upload.
type fileInfo struct {
	ID string `json:"id"`
}

// completeUploadExternalPayload is the payload for files.completeUploadExternal API
type completeUploadExternalPayload struct {
	Files          []fileInfo `json:"files"`
	ChannelID      string     `json:"channel_id,omitempty"`
	InitialComment string     `json:"initial_comment,omitempty"`
}