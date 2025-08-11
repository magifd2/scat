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

	"github.com/magifd2/scat/internal/appcontext"
	"github.com/magifd2/scat/internal/config"
	"github.com/magifd2/scat/internal/provider"
)

const (
	postMessageURL         = "https://slack.com/api/chat.postMessage"
	getUploadURLExternalURL  = "https://slack.com/api/files.getUploadURLExternal"
	completeUploadExternalURL = "https://slack.com/api/files.completeUploadExternal"
	conversationsListURL   = "https://slack.com/api/conversations.list"
	conversationsJoinURL   = "https://slack.com/api/conversations.join"
)

// Provider implements the provider.Interface for Slack.
type Provider struct {
	Profile        config.Profile
	Context        appcontext.Context // Use appcontext.Context
	channelIDCache map[string]string // Cache for channel name to ID mapping
}

// NewProvider creates a new Slack Provider.
func NewProvider(p config.Profile, ctx appcontext.Context) (provider.Interface, error) {
	return &Provider{Profile: p, Context: ctx}, nil
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
	if p.Context.Debug {
		fmt.Fprintln(os.Stderr, "[DEBUG] PostMessage called with Debug mode ON.")
	}

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

	// Attempt to post message
	_, err = p.sendRequest("POST", postMessageURL, bytes.NewBuffer(jsonPayload), "application/json; charset=utf-8")
	if err != nil {
		// Check if the error is 'not_in_channel'
		if strings.Contains(err.Error(), "slack API error: not_in_channel") {
			if !p.Context.Silent {
				fmt.Fprintf(os.Stderr, "Bot not in channel '%s'. Attempting to join...\n", p.Profile.Channel)
			}
			if joinErr := p.joinChannel(channelID); joinErr != nil {
				return fmt.Errorf("failed to join channel '%s': %w", p.Profile.Channel, joinErr)
			}
			if !p.Context.Silent {
				fmt.Fprintf(os.Stderr, "Successfully joined channel '%s'. Retrying post...\n", p.Profile.Channel)
			}
			// Retry post after joining
			_, retryErr := p.sendRequest("POST", postMessageURL, bytes.NewBuffer(jsonPayload), "application/json; charset=utf-8")
			return retryErr
		}
		return err // Return original error if not 'not_in_channel'
	}

	return nil
}

func (p *Provider) PostFile(filePath, filename, filetype, comment, overrideUsername, iconEmoji string) error {
	if p.Context.NoOp {
		fmt.Printf("---\n")
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

	respBody, err := p.sendRequest("GET", getUploadURLExternalURL+"?"+getURLParams.Encode(), nil, "")
	if err != nil {
		return fmt.Errorf("step 1 (getUploadURLExternal) failed: %w", err)
	}

	var getURLResp getUploadURLExternalResponse
	if err := json.Unmarshal(respBody, &getURLResp); err != nil {
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

func (p *Provider) joinChannel(channelID string) error {
	joinPayload := map[string]string{"channel": channelID}
	jsonPayload, err := json.Marshal(joinPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal join payload: %w", err)
	}

	respBody, err := p.sendRequest("POST", conversationsJoinURL, bytes.NewBuffer(jsonPayload), "application/json; charset=utf-8")
	if err != nil {
		return err
	}

	var joinResp apiResponse
	if err := json.Unmarshal(respBody, &joinResp); err != nil {
		return fmt.Errorf("failed to unmarshal join response: %w", err)
	}

	if !joinResp.Ok {
		return fmt.Errorf("slack API error joining channel: %s", joinResp.Error)
	}

	return nil
}

func (p *Provider) sendRequest(method, url string, body io.Reader, contentType string) ([]byte, error) {
	if p.Context.NoOp {
		fmt.Fprintf(os.Stderr, "[DEBUG] Request: %s %s\n", method, url)
		if body != nil {
			// Read body for logging, then reset for actual request
			var buf bytes.Buffer
			t_body := io.TeeReader(body, &buf)
			requestBytes, _ := io.ReadAll(t_body)
			fmt.Fprintf(os.Stderr, "[DEBUG] Request Body: %s\n", string(requestBytes))
			body = &buf // Reset body for actual request
		}
	}

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

	if p.Context.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Response Status: %s\n", resp.Status)
		fmt.Fprintf(os.Stderr, "[DEBUG] Response Body: %s\n", string(bodyBytes))
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