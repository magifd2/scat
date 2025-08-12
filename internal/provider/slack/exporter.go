package slack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/magifd2/scat/internal/provider"
)

// LogExporter returns the log exporter implementation for Slack.
func (p *Provider) LogExporter() provider.LogExporter {
	return p // The Provider itself implements the LogExporter interface
}

// --- LogExporter Methods ---

func (p *Provider) GetConversationHistory(opts provider.GetConversationHistoryOptions) (*provider.ConversationHistoryResponse, error) {
	params := url.Values{}
	params.Add("channel", opts.ChannelID)
	if opts.Latest != "" {
		params.Add("latest", opts.Latest)
	}
	if opts.Oldest != "" {
		params.Add("oldest", opts.Oldest)
	}
	if opts.Limit > 0 {
		params.Add("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Cursor != "" {
		params.Add("cursor", opts.Cursor)
	}

	respBody, err := p.sendRequest("GET", conversationsHistoryURL+"?"+params.Encode(), nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to call conversations.history: %w", err)
	}

	var slackResp conversationsHistoryResponse
	if err := json.Unmarshal(respBody, &slackResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conversations.history response: %w", err)
	}

	// Convert Slack-specific response to the provider-agnostic type
	providerResp := &provider.ConversationHistoryResponse{
		HasMore:          slackResp.HasMore,
		NextCursor:       slackResp.ResponseMetadata.NextCursor,
		ResponseMetadata: slackResp.ResponseMetadata,
		Messages:         slackResp.Messages,
	}

	return providerResp, nil
}

func (p *Provider) GetUserInfo(userID string) (*provider.UserInfoResponse, error) {
	params := url.Values{}
	params.Add("user", userID)

	respBody, err := p.sendRequest("GET", usersInfoURL+"?"+params.Encode(), nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to call users.info: %w", err)
	}

	var slackResp userInfoResponse
	if err := json.Unmarshal(respBody, &slackResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal users.info response: %w", err)
	}

	return &provider.UserInfoResponse{User: slackResp.User}, nil
}

func (p *Provider) DownloadFile(fileURL string) ([]byte, error) {
	// Note: This uses sendRequest which automatically adds the auth header.
	return p.sendRequest("GET", fileURL, nil, "")
}
