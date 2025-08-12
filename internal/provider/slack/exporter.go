package slack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/magifd2/scat/internal/export"
	"github.com/magifd2/scat/internal/util"
)

var mentionRegex = regexp.MustCompile(`<@(U[A-Z0-9]+)>`)

// ExportLog performs the entire export operation for Slack.
func (p *Provider) ExportLog(opts export.Options) (*export.ExportedLog, error) {
	var exportedMessages []export.ExportedMessage
	cursor := ""
	userCache := make(map[string]string)
	var userCacheMux sync.Mutex

	// Resolve channel name to ID
	channelID, err := p.ResolveChannelID(opts.ChannelName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve channel ID for \"%s\": %w", opts.ChannelName, err)
	}

	for {
		resp, err := p.getConversationHistory(channelID, opts, cursor)
		if err != nil {
			return nil, err
		}

		for _, msg := range resp.Messages {
			userName, err := p.resolveUserName(msg.UserID, userCache, &userCacheMux)
			if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not resolve user %s: %v\n", msg.UserID, err)
			}

			files, err := p.handleAttachedFiles(msg.Files, opts.OutputDir, opts.IncludeFiles)
			if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not process files for message %s: %v\n", msg.Timestamp, err)
			}

			resolvedText, err := p.resolveMentions(msg.Text, userCache, &userCacheMux)
			if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not resolve mentions in message %s: %v\n", msg.Timestamp, err)
					resolvedText = msg.Text
			}

			rfc3339Time, err := util.ToRFC3339(msg.Timestamp)
			if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not parse timestamp %s: %v\n", msg.Timestamp, err)
					rfc3339Time = ""
			}

			exportedMsg := export.ExportedMessage{
				UserID:        msg.UserID,
				UserName:      userName,
				Timestamp:     rfc3339Time,
				TimestampUnix: msg.Timestamp,
				Text:          resolvedText,
				Files:         files,
			}
			exportedMessages = append(exportedMessages, exportedMsg)
		}

		if !resp.HasMore {
			break
		}
		cursor = resp.ResponseMetadata.NextCursor
	}

	// Reverse messages
	for i, j := 0, len(exportedMessages)-1; i < j; i, j = i+1, j-1 {
		exportedMessages[i], exportedMessages[j] = exportedMessages[j], exportedMessages[i]
	}

	return &export.ExportedLog{
		ExportTimestamp: time.Now().UTC().Format(time.RFC3339),
		ChannelName:     opts.ChannelName,
		Messages:        exportedMessages,
	}, nil
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

func (p *Provider) resolveUserName(userID string, cache map[string]string, mu *sync.Mutex) (string, error) {
	if userID == "" {
		return "", nil
	}
	mu.Lock()
	name, ok := cache[userID]
	mu.Unlock()
	if ok {
		return name, nil
	}

	respBody, err := p.sendRequest("GET", usersInfoURL+"?user="+userID, nil, "")
	if err != nil {
		return "", err
	}
	var userInfoResp userInfoResponse
	if err := json.Unmarshal(respBody, &userInfoResp); err != nil {
		return "", err
	}

	name = userInfoResp.User.RealName
	if name == "" {
		name = userInfoResp.User.Name
	}

	mu.Lock()
	cache[userID] = name
	mu.Unlock()
	return name, nil
}

func (p *Provider) resolveMentions(text string, cache map[string]string, mu *sync.Mutex) (string, error) {
	var firstErr error
	resolvedText := mentionRegex.ReplaceAllStringFunc(text, func(match string) string {
		userID := mentionRegex.FindStringSubmatch(match)[1]
		userName, err := p.resolveUserName(userID, cache, mu)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			return match
		}
		return "@" + userName
	})
	return resolvedText, firstErr
}

func (p *Provider) handleAttachedFiles(files []file, outputDir string, download bool) ([]export.ExportedFile, error) {
	var exportedFiles []export.ExportedFile
	for _, f := range files {
		exportedFile := export.ExportedFile{
			ID:       f.ID,
			Name:     f.Name,
			Mimetype: f.Mimetype,
		}
		if download {
			safeFilename := filepath.Base(f.Name)
			localPath := filepath.Join(outputDir, f.ID+"_"+safeFilename)
			fileData, err := p.sendRequest("GET", f.URLPrivateDownload, nil, "")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not download file %s: %v\n", f.Name, err)
				continue
			}
			if err := os.WriteFile(localPath, fileData, 0600); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not save file %s: %v\n", f.Name, err)
				continue
			}
			exportedFile.LocalPath = localPath
		}
		exportedFiles = append(exportedFiles, exportedFile)
	}
	return exportedFiles, nil
}