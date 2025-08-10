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

// Client is responsible for sending messages to a configured endpoint.
type Client struct {
	Profile config.Profile
	NoOp    bool // Dry-run flag
}

// NewClient creates a new Client with the given profile.
func NewClient(p config.Profile, noop bool) *Client {
	return &Client{Profile: p, NoOp: noop}
}

// MessagePayload defines the JSON structure for a simple text message.
type MessagePayload struct {
	Text     string `json:"text"`
	Username string `json:"username,omitempty"`
}

// PostMessage sends a simple text message to the endpoint.
func (c *Client) PostMessage(text, overrideUsername string) error {
	username := c.Profile.Username
	if overrideUsername != "" {
		username = overrideUsername
	}

	payload := MessagePayload{
		Text:     text,
		Username: username,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if c.NoOp {
		fmt.Printf("---\n")
		fmt.Printf("Profile: %s\n", c.Profile.Endpoint) // Simplified profile name for now
		fmt.Printf("Payload: %s\n", string(jsonPayload))
		fmt.Printf("-----\n")
		return nil
	}

	return c.sendRequest(bytes.NewBuffer(jsonPayload), "application/json")
}

// PostFile uploads a file using multipart/form-data.
func (c *Client) PostFile(filePath, filename, filetype, comment, overrideUsername string) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file part
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

	// Add other fields
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

	if c.NoOp {
		fmt.Printf("---\n")
		fmt.Printf("Profile: %s\n", c.Profile.Endpoint) // Simplified profile name for now
		fmt.Printf("File: %s\n", filePath)
		fmt.Printf("Filename: %s\n", filename)
		fmt.Printf("Comment: %s\n", comment)
		fmt.Printf("Filetype: %s\n", filetype)
		fmt.Printf("-----\n")
		return nil
	}

	return c.sendRequest(body, writer.FormDataContentType())
}

// sendRequest is a helper to send the actual HTTP request.
func (c *Client) sendRequest(body io.Reader, contentType string) error {
	req, err := http.NewRequest("POST", c.Profile.Endpoint, body)
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
