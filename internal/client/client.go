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
	slackPostMessageURL = "https://slack.com/api/chat.postMessage"
	slackFileUploadURL  = "https://slack.com/api/files.upload"
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
	Channel  string `json:"channel"`
	Text     string `json:"text"`
	Username string `json:"username,omitempty"`
}

// PostMessage sends a simple text message to the endpoint.
func (c *Client) PostMessage(text, overrideUsername string) error {
	switch c.Profile.Provider {
	case "slack":
		return c.postMessageToSlack(text, overrideUsername)
	case "generic":
		return c.postMessageToGeneric(text, overrideUsername)
	default:
		return fmt.Errorf("unknown provider: %s", c.Profile.Provider)
	}
}

// PostFile uploads a file using multipart/form-data.
func (c *Client) PostFile(filePath, filename, filetype, comment, overrideUsername string) error {
	switch c.Profile.Provider {
	case "slack":
		return c.postFileToSlack(filePath, filename, filetype, comment, overrideUsername)
	case "generic":
		return c.postFileToGeneric(filePath, filename, filetype, comment, overrideUsername)
	default:
		return fmt.Errorf("unknown provider: %s", c.Profile.Provider)
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
	return c.sendRequest(c.Profile.Endpoint, bytes.NewBuffer(jsonPayload), "application/json")
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

	return c.sendRequest(c.Profile.Endpoint, body, writer.FormDataContentType())
}

// --- Slack Provider Methods ---

func (c *Client) postMessageToSlack(text, overrideUsername string) error {
	username := c.Profile.Username
	if overrideUsername != "" {
		username = overrideUsername
	}
	payload := SlackMessagePayload{
		Channel:  c.Profile.Channel,
		Text:     text,
		Username: username,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}
	return c.sendRequest(slackPostMessageURL, bytes.NewBuffer(jsonPayload), "application/json; charset=utf-8")
}

func (c *Client) postFileToSlack(filePath, filename, filetype, comment, overrideUsername string) error {
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

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	return c.sendRequest(slackFileUploadURL, body, writer.FormDataContentType())
}

// --- Shared Methods ---

func (c *Client) sendRequest(url string, body io.Reader, contentType string) error {
	if c.NoOp {
		fmt.Printf("--- NOOP: Dry run ---\n")
		fmt.Printf("Provider: %s\n", c.Profile.Provider)
		fmt.Printf("URL: %s\n", url)
		fmt.Printf("Content-Type: %s\n", contentType)
		fmt.Printf("---------------------\n")
		return nil
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	if c.Profile.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Profile.Token)
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}