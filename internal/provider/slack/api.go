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

	"github.com/magifd2/scat/internal/export"
)

const (
	postMessageURL            = "https://slack.com/api/chat.postMessage"
	getUploadURLExternalURL     = "https://slack.com/api/files.getUploadURLExternal"
	completeUploadExternalURL    = "https://slack.com/api/files.completeUploadExternal"
	conversationsListURL      = "https://slack.com/api/conversations.list"
	conversationsJoinURL      = "https://slack.com/api/conversations.join"
	conversationsHistoryURL   = "https://slack.com/api/conversations.history"
	usersInfoURL              = "https://slack.com/api/users.info"
)

func (p *Provider) getConversationHistory(channelID string, opts export.Options, cursor string) (*conversationsHistoryResponse, error) {
	params := url.Values{}
	params.Add("channel", channelID)
	if opts.EndTime != "" {
		params.Add("latest", opts.EndTime)
	}
	if opts.StartTime != "" {
		params.Add("oldest", opts.StartTime)
	}
	if cursor != "" {
		params.Add("cursor", cursor)
	}
	params.Add("limit", "200")

	respBody, err := p.sendRequest("GET", conversationsHistoryURL+"?"+params.Encode(), nil, "")
	if err != nil && strings.Contains(err.Error(), "not_in_channel") {
		if !p.Context.Silent {
			fmt.Fprintf(os.Stderr, "Bot not in channel '%s'. Attempting to join...\n", opts.ChannelName)
		}
		if joinErr := p.joinChannel(channelID); joinErr != nil {
			return nil, fmt.Errorf("failed to auto-join channel '%s': %w", opts.ChannelName, joinErr)
		}
		if !p.Context.Silent {
			fmt.Fprintf(os.Stderr, "Successfully joined channel '%s'. Retrying...\n", opts.ChannelName)
		}
		respBody, err = p.sendRequest("GET", conversationsHistoryURL+"?"+params.Encode(), nil, "")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to call conversations.history: %w", err)
	}

	var slackResp conversationsHistoryResponse
	if err := json.Unmarshal(respBody, &slackResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conversations.history response: %w", err)
	}
	return &slackResp, nil
}

func (p *Provider) populateChannelCache() error {
	if p.Context.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Populating channel cache by calling conversations.list...\n")
	}
	p.channelIDCache = make(map[string]string)
	cursor := ""

	for {
		url := fmt.Sprintf("%s?cursor=%s&types=public_channel,private_channel&limit=200", conversationsListURL, cursor)
		body, err := p.sendRequest("GET", url, nil, "")
		if err != nil {
			return err
		}

		var listResp conversationsListResponse
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

	// Use the httpClient from the Provider struct
	resp, err := p.httpClient.Do(req)
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
