package provider

// This file defines common, provider-agnostic data structures for API responses
// used by the LogExporter interface.

// GetConversationHistoryOptions defines the parameters for a GetConversationHistory call.
type GetConversationHistoryOptions struct {
	ChannelID string // Required
	Latest    string // Optional
	Oldest    string // Optional
	Limit     int    // Optional
	Cursor    string // Optional
}

// ConversationHistoryResponse represents the response from a conversation history API call.
type ConversationHistoryResponse struct {
	Messages         []Message
	HasMore          bool
	NextCursor       string
	ResponseMetadata ResponseMetadata
}

// UserInfoResponse represents the response from a user info API call.
type UserInfoResponse struct {
	User User
}

// Message represents a single message in a channel.
type Message struct {
	Type      string
	Timestamp string
	UserID    string
	Text      string
	Files     []File
}

// File represents a file attached to a message.
type File struct {
	ID                 string
	Created            int
	Timestamp          int
	Name               string
	Title              string
	Mimetype           string
	Filetype           string
	PrettyType         string
	User               string
	Editable           bool
	Size               int
	Mode               string
	IsExternal         bool
	ExternalType       string
	IsPublic           bool
	PublicURLShared    bool
	DisplayAsBot       bool
	Username           string
	URLPrivate         string `json:"url_private"`
	URLPrivateDownload string `json:"url_private_download"`
	Permalink          string
	PermalinkPublic    string `json:"permalink_public"`
}

// User represents a user.
type User struct {
	ID       string
	TeamID   string
	Name     string
	RealName string
}

// ResponseMetadata contains metadata about the response, like cursors for pagination.
type ResponseMetadata struct {
	NextCursor string `json:"next_cursor"`
}