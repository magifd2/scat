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
	getUploadURLExternalURL   = "https://slack.com/api/files.getUploadURLExternal"
	completeUploadExternalURL = "https://slack.com/api/files.completeUploadExternal"
	conversationsListURL      = "https://slack.com/api/conversations.list"
	conversationsJoinURL      = "https://slack.com/api/conversations.join"
	conversationsHistoryURL   = "https://slack.com/api/conversations.history"
	conversationsRepliesURL   = "https://slack.com/api/conversations.replies"
	conversationsOpenURL      = "https://slack.com/api/conversations.open"
	usersListURL              = "https://slack.com/api/users.list"
	usersInfoURL              = "https://slack.com/api/users.info"
)

// conversationsOpenResponse defines the structure for the conversations.open API response.
type conversationsOpenResponse struct {
	Ok      bool   `json:"ok"`
	Error   string `json:"error"`
	Channel struct {
		ID string `json:"id"`
	} `json:"channel"`
}

// usersListResponse defines the structure for the users.list API response.
type usersListResponse struct {
	Ok      bool   `json:"ok"`
	Error   string `json:"error"`
	Members []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		IsBot   bool   `json:"is_bot"`
		Deleted bool   `json:"deleted"`
		Profile struct {
			DisplayName string `json:"display_name"`
		} `json:"profile"`
	} `json:"members"`
	ResponseMetadata struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
}

// openDMChannel opens a direct message channel with a user and returns the channel ID.
func (p *Provider) openDMChannel(userID string) (string, error) {
	payload := map[string]string{"users": userID}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal conversations.open payload: %w", err)
	}

	respBody, err := p.sendRequest("POST", conversationsOpenURL, bytes.NewBuffer(jsonPayload), "application/json; charset=utf-8")
	if err != nil {
		return "", err
	}

	var openResp conversationsOpenResponse
	if err := json.Unmarshal(respBody, &openResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal conversations.open response: %w", err)
	}

	if !openResp.Ok {
		return "", fmt.Errorf("slack API error on conversations.open: %s", openResp.Error)
	}
	if openResp.Channel.ID == "" {
		return "", fmt.Errorf("conversations.open did not return a channel ID")
	}

	return openResp.Channel.ID, nil
}

// getUsers fetches all non-bot, non-deleted users from the workspace.
func (p *Provider) getUsers() ([]struct{ ID, Name string }, error) {
	if p.Context.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Fetching users by calling users.list...\n")
	}
	var allUsers []struct{ ID, Name string }
	cursor := ""

	for {
		url := fmt.Sprintf("%s?cursor=%s&limit=200", usersListURL, cursor)
		body, err := p.sendRequest("GET", url, nil, "")
		if err != nil {
			return nil, err
		}

		var listResp usersListResponse
		if err := json.Unmarshal(body, &listResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal users.list response: %w", err)
		}

		if !listResp.Ok {
			return nil, fmt.Errorf("slack API error on users.list: %s", listResp.Error)
		}

		for _, user := range listResp.Members {
			if !user.IsBot && !user.Deleted {
				name := user.Profile.DisplayName
				if name == "" {
					name = user.Name
				}
				allUsers = append(allUsers, struct{ ID, Name string }{ID: user.ID, Name: name})
			}
		}

		cursor = listResp.ResponseMetadata.NextCursor
		if cursor == "" {
			break
		}
	}
	return allUsers, nil
}

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

// populateUserCache fetches all users and populates the userIDCache.
func (p *Provider) populateUserCache() error {
	users, err := p.getUsers()
	if err != nil {
		return fmt.Errorf("failed to get users for cache: %w", err)
	}

	p.userIDCache = make(map[string]string)
	for _, user := range users {
		p.userIDCache[user.Name] = user.ID
		if p.Context.Debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Caching user: Name=%s, ID=%s\n", user.Name, user.ID)
		}
	}

	if p.Context.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] User cache populated with %d users.\n", len(p.userIDCache))
	}
	return nil
}

// ResolveUserID finds a user ID for a given user name.
// It checks the cache first, and repopulates it if the user is not found.
func (p *Provider) ResolveUserID(userName string) (string, error) {
	cleanUserName := strings.TrimPrefix(userName, "@")

	id, ok := p.userIDCache[cleanUserName]
	if ok {
		return id, nil
	}

	// If not found, refresh the cache and try again.
	if p.Context.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] User '%s' not found in cache, repopulating...\n", cleanUserName)
	}
	if err := p.populateUserCache(); err != nil {
		return "", fmt.Errorf("failed to repopulate user cache: %w", err)
	}

	id, ok = p.userIDCache[cleanUserName]
	if !ok {
		return "", fmt.Errorf("user '%s' not found", userName)
	}

	return id, nil
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

func (p *Provider) getConversationReplies(channelID, ts, cursor string) (*conversationsHistoryResponse, error) {
	params := url.Values{}
	params.Add("channel", channelID)
	params.Add("ts", ts)
	if cursor != "" {
		params.Add("cursor", cursor)
	}
	params.Add("limit", "200")

	respBody, err := p.sendRequest("GET", conversationsRepliesURL+"?"+params.Encode(), nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to call conversations.replies: %w", err)
	}

	var slackResp conversationsHistoryResponse
	if err := json.Unmarshal(respBody, &slackResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conversations.replies response: %w", err)
	}
	return &slackResp, nil
}

