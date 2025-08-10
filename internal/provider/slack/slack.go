package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/magifd2/scat/internal/config"
	"github.com/magifd2/scat/internal/provider"
)

const (
	postMessageURL         = "https://slack.com/api/chat.postMessage"
	getUploadURLExternalURL  = "https://slack.com/api/files.getUploadURLExternal"
	completeUploadExternalURL = "https://slack.com/api/files.completeUploadExternal"
	conversationsListURL   = "https://slack.com/api/conversations.list"
)

// Provider implements the provider.Interface for Slack.
type Provider struct {
	Profile        config.Profile
	NoOp           bool
	channelIDCache map[string]string // Cache for channel name to ID mapping
}

// NewProvider creates a new Slack Provider.
func NewProvider(p config.Profile, noop bool) (provider.Interface, error) {
	return &Provider{Profile: p, NoOp: noop}, nil
}

// Capabilities returns the features supported by the Slack provider.
func (p *Provider) Capabilities() provider.Capabilities {
	return provider.Capabilities{
		CanListChannels: true,
		CanPostFile:     true,
		CanUseIconEmoji: true,
	}
}

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

// --- Provider Methods ---

func (p *Provider) PostMessage(text, overrideUsername, iconEmoji string) error {
	channelID, err := p.getChannelID(p.Profile.Channel)
	if err != nil {
		return err
	}

	username := p.Profile.Username
	if overrideUsername != "" {
		username = overrideUsername
	}
	payload := messagePayload{
		Channel:   channelID,
		Text:      text,
		Username:  username,
		IconEmoji: iconEmoji,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}
	_, err = p.sendRequest("POST", postMessageURL, bytes.NewBuffer(jsonPayload), "application/json; charset=utf-8")
	return err
}

func (p *Provider) PostFile(filePath, filename, filetype, comment, overrideUsername, iconEmoji string) error {
	if p.NoOp {
		fmt.Printf("--- NOOP: Dry run ---\n")
		fmt.Printf("Provider: slack\n")
		fmt.Printf("Action: Upload file %s\n", filePath)
		fmt.Printf("---------------------\n")
		return nil
	}

	// Step 1: Get Upload URL
	fi, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	getURLParams := url.Values{}
	getURLParams.Add("filename", filename)
	getURLParams.Add("length", fmt.Sprintf("%d", fi.Size()))

	getURLRespBytes, err := p.sendRequest("GET", getUploadURLExternalURL+"?"+getURLParams.Encode(), nil, "")
	if err != nil {
		return fmt.Errorf("step 1 (getUploadURLExternal) failed: %w", err)
	}

	var getURLResp getUploadURLExternalResponse
	if err := json.Unmarshal(getURLRespBytes, &getURLResp); err != nil {
		return fmt.Errorf("failed to unmarshal getUploadURLExternal response: %w", err)
	}
	if !getURLResp.Ok {
		return fmt.Errorf("slack API error on getUploadURLExternal: %s", getURLResp.Error)
	}

	// Step 2: Upload file to the provided URL
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for upload: %w", err)
	}
	defer file.Close()

	uploadReq, err := http.NewRequest("POST", getURLResp.UploadURL, file)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}
	uploadReq.Header.Set("Content-Type", "application/octet-stream")

	httpClient := &http.Client{}
	uploadResp, err := httpClient.Do(uploadReq)
	if err != nil {
		return fmt.Errorf("step 2 (upload to url) failed: %w", err)
	}
	defer uploadResp.Body.Close()
	if uploadResp.StatusCode != 200 {
		body, _ := io.ReadAll(uploadResp.Body)
		return fmt.Errorf("upload to url failed with status %d: %s", uploadResp.StatusCode, string(body)) 
	}

	// Step 3: Complete the upload
	channelID, err := p.getChannelID(p.Profile.Channel)
	if err != nil {
		return err
	}

	completePayload := completeUploadExternalPayload{
		Files:          []fileInfo{{ID: getURLResp.FileID}},
		ChannelID:      channelID,
		InitialComment: comment,
	}
	completePayloadBytes, err := json.Marshal(completePayload)
	if err != nil {
		return fmt.Errorf("failed to marshal completeUploadExternal payload: %w", err)
	}

	_, err = p.sendRequest("POST", completeUploadExternalURL, bytes.NewBuffer(completePayloadBytes), "application/json; charset=utf-8")
	if err != nil {
		return fmt.Errorf("step 3 (completeUploadExternal) failed: %w", err)
	}

	return nil
}

func (p *Provider) ListChannels() ([]string, error) {
	if p.channelIDCache == nil {
		if err := p.populateChannelCache(); err != nil {
			return nil, err
		}
	}
	var channelNames []string
	for name := range p.channelIDCache {
		channelNames = append(channelNames, "#"+name)
	}
	return channelNames, nil
}

// --- Helper Methods ---

func (p *Provider) getChannelID(name string) (string, error) {
	if p.channelIDCache == nil {
		if err := p.populateChannelCache(); err != nil {
			return "", err
		}
	}
	// Slack channel names can be with or without a leading #
	name = strings.TrimPrefix(name, "#")
	if id, ok := p.channelIDCache[name]; ok {
		return id, nil
	}
	return "", fmt.Errorf("channel '%s' not found", name)
}

func (p *Provider) populateChannelCache() error {
	p.channelIDCache = make(map[string]string)
	cursor := ""

	for {
		url := fmt.Sprintf("%s?cursor=%s&types=public_channel,private_channel&limit=200", conversationsListURL, cursor)
		body, err := p.sendRequest("GET", url, nil, "")
		if err != nil {
			return err
		}

		var listResp apiResponse
		if err := json.Unmarshal(body, &listResp); err != nil {
			return fmt.Errorf("failed to unmarshal conversations.list response: %w", err)
		}

		if !listResp.Ok {
			return fmt.Errorf("slack API error on conversations.list: %s", listResp.Error)
		}

		for _, ch := range listResp.Channels {
			p.channelIDCache[ch.Name] = ch.ID
		}

		cursor = listResp.ResponseMetadata.NextCursor
		if cursor == "" {
			break
		}
	}
	return nil
}

func (p *Provider) sendRequest(method, url string, body io.Reader, contentType string) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Authorization", "Bearer "+p.Profile.Token)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Check for `ok: false` in the response body itself.
	var baseResp apiResponse
	if err := json.Unmarshal(bodyBytes, &baseResp); err == nil {
		if !baseResp.Ok {
			return nil, fmt.Errorf("slack API error: %s", baseResp.Error)
		}
	}

	return bodyBytes, nil
}
