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

// Provider implements the provider.Interface for Slack.
type Provider struct {
	Profile        config.Profile
	Context        appcontext.Context // Use appcontext.Context
	channelIDCache map[string]string // Cache for channel name to ID mapping
}

// NewProvider creates a new Slack Provider.
func NewProvider(p config.Profile, ctx appcontext.Context) (provider.Interface, error) {
	prov := &Provider{Profile: p, Context: ctx}
	// Best-effort attempt to populate the channel cache on initialization.
	// If it fails (e.g., due to missing permissions), we don't treat it as a fatal error.
	// The cache will be populated on-demand later if needed.
	if err := prov.populateChannelCache(); err != nil {
		if ctx.Debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Failed to populate channel cache on init: %v\n", err)
		}
	}
	return prov, nil
}

// Capabilities returns the features supported by the Slack provider.
func (p *Provider) Capabilities() provider.Capabilities {
	return provider.Capabilities{
		CanListChannels: true,
		CanPostFile:     true,
		CanUseIconEmoji: true,
	}
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
		// Check if the error is 'not_in_channel' and retry if so.
		if strings.Contains(err.Error(), "slack API error: not_in_channel") {
			if !p.Context.Silent {
				fmt.Fprintf(os.Stderr, "Bot not in channel '%s'. Attempting to join...\n", p.Profile.Channel)
			}
			if joinErr := p.joinChannel(channelID); joinErr != nil {
				return fmt.Errorf("failed to join channel '%s': %w", p.Profile.Channel, joinErr)
			}
			if !p.Context.Silent {
				fmt.Fprintf(os.Stderr, "Successfully joined channel '%s'. Retrying file upload completion...\n", p.Profile.Channel)
			}
			// Retry completing the upload after joining.
			_, retryErr := p.sendRequest("POST", completeUploadExternalURL, bytes.NewBuffer(completePayloadBytes), "application/json; charset=utf-8")
			if retryErr != nil {
				return fmt.Errorf("step 3 (completeUploadExternal) failed on retry: %w", retryErr)
			}
			return nil // Success on retry
		}
		return fmt.Errorf("step 3 (completeUploadExternal) failed: %w", err)
	}

	return nil
}

func (p *Provider) ListChannels() ([]string, error) {
	// Ensure the cache is populated before listing.
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