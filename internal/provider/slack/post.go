package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func (p *Provider) PostMessage(text, overrideUsername, iconEmoji string) error {
	if p.Context.Debug {
		fmt.Fprintln(os.Stderr, "[DEBUG] PostMessage called with Debug mode ON.")
	}

	channelID, err := p.ResolveChannelID(p.Profile.Channel)
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
				fmt.Fprintf(os.Stderr, "Bot not in channel \"%s\". Attempting to join...\n", p.Profile.Channel)
			}
			if joinErr := p.joinChannel(channelID); joinErr != nil {
				return fmt.Errorf("failed to join channel \"%s\": %w", p.Profile.Channel, joinErr)
			}
			if !p.Context.Silent {
				fmt.Fprintf(os.Stderr, "Successfully joined channel \"%s\". Retrying post...\n", p.Profile.Channel)
			}
			// Retry post after joining
			_, retryErr := p.sendRequest("POST", postMessageURL, bytes.NewBuffer(jsonPayload), "application/json; charset=utf-8")
			return retryErr
		}
		return err // Return original error if not 'not_in_channel'
	}

	return nil
}