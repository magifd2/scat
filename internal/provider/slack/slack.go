package slack

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
	"github.com/magifd2/scat/internal/provider"
)

const (
	postMessageURL       = "https://slack.com/api/chat.postMessage"
	fileUploadURL        = "https://slack.com/api/files.upload"
	conversationsListURL = "https://slack.com/api/conversations.list"
)

// Provider implements the provider.Interface for Slack.
type Provider struct {
	Profile config.Profile
	NoOp    bool
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
		CanUseThreads:   true,
		CanUseIconEmoji: true,
	}
}

type messagePayload struct {
	Channel   string `json:"channel"`
	Text      string `json:"text"`
	Username  string `json:"username,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
	ThreadTS  string `json:"thread_ts,omitempty"`
}

type apiResponse struct {
	Ok               bool   `json:"ok"`
	Error            string `json:"error,omitempty"`
	TS               string `json:"ts,omitempty"`
	File             struct {
		Shares struct {
			Public map[string][]struct {
				TS string `json:"ts"`
			} `json:"public"`
		} `json:"shares"`
	} `json:"file,omitempty"`
	Channels         []struct {
		Name string `json:"name"`
	} `json:"channels"`
	ResponseMetadata struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
}

func (p *Provider) PostMessage(text, overrideUsername, iconEmoji string, thread bool, threadTS string) (string, error) {
	username := p.Profile.Username
	if overrideUsername != "" {
		username = overrideUsername
	}
	payload := messagePayload{
		Channel:   p.Profile.Channel,
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
	return p.sendRequest(postMessageURL, bytes.NewBuffer(jsonPayload), "application/json; charset=utf-8")
}

func (p *Provider) PostFile(filePath, filename, filetype, comment, overrideUsername, iconEmoji string, thread bool, threadTS string) (string, error) {
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

	_ = writer.WriteField("channels", p.Profile.Channel)
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

	return p.sendRequest(fileUploadURL, body, writer.FormDataContentType())
}

func (p *Provider) ListChannels() ([]string, error) {
	var allChannels []string
	cursor := ""

	for {
		url := fmt.Sprintf("%s?cursor=%s&types=public_channel,private_channel&limit=200", conversationsListURL, cursor)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request for conversations.list: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+p.Profile.Token)

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

		var listResp apiResponse
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

func (p *Provider) sendRequest(url string, body io.Reader, contentType string) (string, error) {
	if p.NoOp {
		fmt.Printf("---\nProvider: slack\nURL: %s\n-----\n", url)
		return "", nil
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+p.Profile.Token)

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

	var slackResp apiResponse
	if err := json.Unmarshal(bodyBytes, &slackResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal slack response: %w", err)
	}

	if !slackResp.Ok {
		return "", fmt.Errorf("slack API error: %s", slackResp.Error)
	}

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