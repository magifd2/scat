package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/magifd2/scat/internal/config"
)

// Client is responsible for sending messages to a configured endpoint.
type Client struct {
	Profile config.Profile
}

// NewClient creates a new Client with the given profile.
func NewClient(p config.Profile) *Client {
	return &Client{Profile: p}
}

// PostMessage sends a simple text message to the endpoint.
type MessagePayload struct {
	Text     string `json:"text"`
	Username string `json:"username,omitempty"`
}

func (c *Client) PostMessage(text string) error {
	payload := MessagePayload{
		Text:     text,
		Username: c.Profile.Username,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.Profile.Endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
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
		return fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	return nil
}
