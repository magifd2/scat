package export

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/magifd2/scat/internal/provider"
)

var mentionRegex = regexp.MustCompile(`<@(U[A-Z0-9]+)>`)

// Exporter handles the logic of exporting channel logs.
type Exporter struct {
	prov         provider.LogExporter
	userCache    map[string]string
	userCacheMux sync.Mutex
}

// NewExporter creates a new Exporter.
func NewExporter(prov provider.LogExporter) *Exporter {
	return &Exporter{
		prov:      prov,
		userCache: make(map[string]string),
	}
}

// ExportLog performs the export operation based on the given options.
func (e *Exporter) ExportLog(opts Options) (*ExportedLog, error) {
	var exportedMessages []ExportedMessage
	cursor := ""

	for {
		historyOpts := provider.GetConversationHistoryOptions{
			ChannelName: opts.ChannelName,
			Latest:      opts.EndTime,
			Oldest:      opts.StartTime,
			Cursor:      cursor,
			Limit:       200, // Sensible limit per page
		}

		resp, err := e.prov.GetConversationHistory(historyOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to get conversation history: %w", err)
		}

		for _, msg := range resp.Messages {
			userName, err := e.resolveUserName(msg.UserID)
			if err != nil {
				// Log error but continue, maybe the user is deactivated
				fmt.Fprintf(os.Stderr, "Warning: could not resolve user %s: %v\n", msg.UserID, err)
			}

			files, err := e.handleAttachedFiles(msg.Files, opts.OutputDir, opts.IncludeFiles)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not process files for message %s: %v\n", msg.Timestamp, err)
			}

			resolvedText, err := e.resolveMentions(msg.Text)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not resolve mentions in message %s: %v\n", msg.Timestamp, err)
				resolvedText = msg.Text // Use original text on failure
			}

			rfc3339Time, err := toRFC3339(msg.Timestamp)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not parse timestamp %s: %v\n", msg.Timestamp, err)
				rfc3339Time = "" // Set to empty on failure
			}

			exportedMsg := ExportedMessage{
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
		cursor = resp.NextCursor
	}

	// Reverse messages to have them in chronological order
	for i, j := 0, len(exportedMessages)-1; i < j; i, j = i+1, j-1 {
		exportedMessages[i], exportedMessages[j] = exportedMessages[j], exportedMessages[i]
	}

	log := &ExportedLog{
		ExportTimestamp: time.Now().UTC().Format(time.RFC3339),
		ChannelName:     opts.ChannelName,
		Messages:        exportedMessages,
	}

	return log, nil
}

// toRFC3339 converts a string unix timestamp (e.g., "1234567890.123456") to RFC3339 format.
func toRFC3339(unixTs string) (string, error) {
	if unixTs == "" {
		return "", nil
	}
	floatTs, err := strconv.ParseFloat(unixTs, 64)
	if err != nil {
		return "", err
	}
	sec := int64(floatTs)
	nsec := int64((floatTs - float64(sec)) * 1e9)
	t := time.Unix(sec, nsec)
	return t.UTC().Format(time.RFC3339), nil
}

// resolveMentions replaces all user ID mentions in a text with their user names.
func (e *Exporter) resolveMentions(text string) (string, error) {
	var firstErr error
	resolvedText := mentionRegex.ReplaceAllStringFunc(text, func(match string) string {
		userID := mentionRegex.FindStringSubmatch(match)[1]
		userName, err := e.resolveUserName(userID)
		if err != nil {
			// Keep track of the first error, but return the original match to not lose data.
			if firstErr == nil {
				firstErr = fmt.Errorf("failed to resolve mention for %s: %w", userID, err)
			}
			return match
		}
		return "@" + userName
	})
	return resolvedText, firstErr
}

// resolveUserName gets a user's name from the cache or fetches it from the provider.
func (e *Exporter) resolveUserName(userID string) (string, error) {
	if userID == "" {
		return "", nil // Skip if user ID is empty (e.g., for bot messages)
	}
	e.userCacheMux.Lock()
	name, ok := e.userCache[userID]
	e.userCacheMux.Unlock()

	if ok {
		return name, nil
	}

	userInfo, err := e.prov.GetUserInfo(userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user info for %s: %w", userID, err)
	}

	name = userInfo.User.RealName
	if name == "" {
		name = userInfo.User.Name
	}

	e.userCacheMux.Lock()
	e.userCache[userID] = name
	e.userCacheMux.Unlock()

	return name, nil
}

// handleAttachedFiles processes file attachments, downloading them if requested.
func (e *Exporter) handleAttachedFiles(files []provider.File, outputDir string, download bool) ([]ExportedFile, error) {
	var exportedFiles []ExportedFile
	for _, f := range files {
		exportedFile := ExportedFile{
			ID:       f.ID,
			Name:     f.Name,
			Mimetype: f.Mimetype,
		}

		if download {
			// Sanitize filename to prevent path traversal issues
			safeFilename := filepath.Base(f.Name)
			localPath := filepath.Join(outputDir, f.ID+"_"+safeFilename)

			fileData, err := e.prov.DownloadFile(f.URLPrivateDownload)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to download file %s (%s): %v\n", f.Name, f.URLPrivateDownload, err)
				continue // Skip this file
			}

			if err := os.WriteFile(localPath, fileData, 0600); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save file %s to %s: %v\n", f.Name, localPath, err)
				continue // Skip this file
			}
			exportedFile.LocalPath = localPath
		}

		exportedFiles = append(exportedFiles, exportedFile)
	}
	return exportedFiles, nil
}
