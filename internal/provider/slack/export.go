package slack

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/magifd2/scat/internal/export"
	"github.com/magifd2/scat/internal/util"
)

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
			var userID string
			var postType string
			userName := ""
			if msg.SubType == "bot_message" {
				if msg.Username != "" {
					userName = msg.Username
				} else if msg.BotID != "" {
					// For now, just use BotID as the name.
					// A more sophisticated solution might involve calling bots.info API.
					userName = fmt.Sprintf("bot:%s", msg.BotID)
				}
				userID = msg.BotID // Set UserID to BotID for bot messages
				postType = "bot"
			} else {
				var err error
				userName, err = p.resolveUserName(msg.UserID, userCache, &userCacheMux)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not resolve user %s: %v\n", msg.UserID, err)
				}
				userID = msg.UserID // Keep original UserID for non-bot messages
				postType = "user"
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
				UserID:        userID, // Use the new userID variable
				UserName:      userName,
				PostType:      postType, // Populate PostType
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
