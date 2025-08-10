package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/magifd2/scat/internal/config"
)

const (
	slackPostMessageURL    = "https://slack.com/api/chat.postMessage"
	slackFileUploadURL     = "https://slack.com/api/files.upload"
	slackConversationsListURL = "https://slack.com/api/conversations.list"
)

// Client is responsible for sending messages to a configured endpoint.
type Client struct {
	Profile config.Profile
	NoOp    bool // Dry-run flag
}

// NewClient creates a new Client with the given profile.
func NewClient(p config.Profile, noop bool) *Client {
	return &Client{Profile: p, NoOp: noop}
}

// --- Generic Provider Structs ---
type GenericMessagePayload struct {
	Text     string `json:"text"`
	Username string `json:"username,omitempty"`
}

// --- Slack Provider Structs ---
type SlackMessagePayload struct {
	Channel   string `json:"channel"`
	Text      string `json:"text"`
	Username  string `json:"username,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
	ThreadTS  string `json:"thread_ts,omitempty"`
}

type SlackResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
	TS    string `json:"ts,omitempty" // For chat.postMessage`
	File  struct {
		Shares struct {
			Public map[string][]struct {
				TS string `json:"ts"`
			} `json:"public"`
		} `json:"shares"`
	} `json:"file,omitempty" // For files.upload`
}

type SlackConversationsListResponse struct {
	Ok       bool `json:"ok"`
	Error    string `json:"error,omitempty"`
	Channels []struct {
		Name string `json:"name"`
	} `json:"channels"`
	ResponseMetadata struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
}

// PostMessage sends a simple text message to the endpoint.
func (c *Client) PostMessage(text, overrideUsername, iconEmoji string, thread bool, threadTS string) (string, error) {
	switch c.Profile.Provider {
	case "slack":
		return c.postMessageToSlack(text, overrideUsername, iconEmoji, thread, threadTS)
	case "generic":
		return "", c.postMessageToGeneric(text, overrideUsername)
	default:
		return "", fmt.Errorf("unknown provider: %s", c.Profile.Provider)
	}
}

// PostFile uploads a file using multipart/form-data.
func (c *Client) PostFile(filePath, filename, filetype, comment, overrideUsername, iconEmoji string, thread bool, threadTS string) (string, error) {
	switch c.Profile.Provider {
	case "slack":
		return c.postFileToSlack(filePath, filename, filetype, comment, overrideUsername, thread, threadTS)
	case "generic":
		return "", c.postFileToGeneric(filePath, filename, filetype, comment, overrideUsername)
	default:
		return "", fmt.Errorf("unknown provider: %s", c.Profile.Provider)
	}
}

// --- Generic Provider Methods ---

func (c *Client) postMessageToGeneric(text, overrideUsername string) error {
	username := c.Profile.Username
	if overrideUsername != "" {
		username = overrideUsername
	}
	payload := GenericMessagePayload{
		Text:     text,
		Username: username,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	_, err = c.sendRequest(c.Profile.Endpoint, bytes.NewBuffer(jsonPayload), "application/json")
	return err
}

func (c *Client) postFileToGeneric(filePath, filename, filetype, comment, overrideUsername string) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err = io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file to buffer: %w", err)
	}

	username := c.Profile.Username
	if overrideUsername != "" {
		username = overrideUsername
	}
	if username != "" {
		_ = writer.WriteField("username", username)
	}
	if comment != "" {
		_ = writer.WriteField("comment", comment)
	}
	if filetype != "" {
		_ = writer.WriteField("filetype", filetype)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	_, err = c.sendRequest(c.Profile.Endpoint, body, writer.FormDataContentType())
	return err
}

// --- Slack Provider Methods ---

func (c *Client) postMessageToSlack(text, overrideUsername, iconEmoji string, thread bool, threadTS string) (string, error) {
	username := c.Profile.Username
	if overrideUsername != "" {
		username = overrideUsername
	}
	payload := SlackMessagePayload{
		Channel:   c.Profile.Channel,
		Text:      text,
		Username:  username,
		IconEmoji: iconEmoji,
	}
	if thread && threadTS != "" {
		payload.ThreadTS = threadTS
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal slack payload: %w", err)
	}
	return c.sendRequest(slackPostMessageURL, bytes.NewBuffer(jsonPayload), "application/json; charset=utf-8")
}

func (c *Client) postFileToSlack(filePath, filename, filetype, comment, overrideUsername string, thread bool, threadTS string) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err = io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy file to buffer: %w", err)
	}

	_ = writer.WriteField("channels", c.Profile.Channel)
	if comment != "" {
		_ = writer.WriteField("initial_comment", comment)
	}
	if filename != "" {
		_ = writer.WriteField("filename", filename)
	}
	if filetype != "" {
		_ = writer.WriteField("filetype", filetype)
	}
	if thread && threadTS != "" {
		_ = writer.WriteField("thread_ts", threadTS)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	return c.sendRequest(slackFileUploadURL, body, writer.FormDataContentType())
}

func (c *Client) ListSlackChannels() ([]string, error) {
	var allChannels []string
	cursor := ""

	for {
		url := fmt.Sprintf("%s?cursor=%s&types=public_channel,private_channel&limit=200", slackConversationsListURL, cursor)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request for conversations.list: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.Profile.Token)

		httpClient := &http.Client{}
		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to send request to conversations.list: %w", err)
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body from conversations.list: %w", err)
		}

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("conversations.list request failed with status code %d: %s", resp.StatusCode, string(bodyBytes))
		}

		var listResp SlackConversationsListResponse
		if err := json.Unmarshal(bodyBytes, &listResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conversations.list response: %w", err)
		}

		if !listResp.Ok {
			return nil, fmt.Errorf("slack API error on conversations.list: %s", listResp.Error)
		}

		for _, ch := range listResp.Channels {
			allChannels = append(allChannels, "#"+ch.Name)
		}

		cursor = listResp.ResponseMetadata.NextCursor
		if cursor == "" {
			break
		}
	}
	return allChannels, nil
}

// --- Shared Methods ---

func (c *Client) sendRequest(url string, body io.Reader, contentType string) (string, error) {
	if c.NoOp {
		fmt.Printf("--- NOOP: Dry run ---\n")
		fmt.Printf("Provider: %s\n", c.Profile.Provider)
		fmt.Printf("URL: %s\n", url)
		fmt.Printf("Content-Type: %s\n", contentType)
		fmt.Printf("---------------------\n")
		return "", nil
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	if c.Profile.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Profile.Token)
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var slackResp SlackResponse
	if err := json.Unmarshal(bodyBytes, &slackResp); err != nil {
		// If it's not a valid slack response, it might be a generic one, so don't error out.
		return "", nil
	}

	if !slackResp.Ok {
		return "", fmt.Errorf("slack API error: %s", slackResp.Error)
	}

	// Return timestamp for threading
	if slackResp.TS != "" {
		return slackResp.TS, nil
	}
	if slackResp.File.Shares.Public != nil {
		for _, shares := range slackResp.File.Shares.Public {
			if len(shares) > 0 {
				return shares[0].TS, nil
			}
		}
	}

	return "", nil
}