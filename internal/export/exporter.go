package export

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/magifd2/scat/internal/provider"
)

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

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory %s: %w", opts.OutputDir, err)
	}

	for {
	historyOpts := provider.GetConversationHistoryOptions{
			ChannelID: opts.ChannelID,
			Latest:    opts.EndTime,
			Oldest:    opts.StartTime,
			Cursor:    cursor,
			Limit:     200, // Sensible limit per page
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

			var files []ExportedFile
			if opts.IncludeFiles && len(msg.Files) > 0 {
				files, err = e.handleAttachedFiles(msg.Files, opts.OutputDir)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not download files for message %s: %v\n", msg.Timestamp, err)
				}
			}

			exportedMsg := ExportedMessage{
				UserID:    msg.UserID,
				UserName:  userName,
				Timestamp: msg.Timestamp,
				Text:      msg.Text,
				Files:     files,
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
		ChannelName:     opts.ChannelID, // This is a name, should be resolved to name if it's an ID
		Messages:        exportedMessages,
	}

	return log, nil
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

// handleAttachedFiles downloads files attached to a message.
func (e *Exporter) handleAttachedFiles(files []provider.File, outputDir string) ([]ExportedFile, error) {
	var exportedFiles []ExportedFile
	for _, f := range files {
		// Sanitize filename to prevent path traversal issues
		safeFilename := filepath.Base(f.Name)
		localPath := filepath.Join(outputDir, f.ID+"_"+safeFilename)

		fileData, err := e.prov.DownloadFile(f.URLPrivateDownload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to download file %s (%s): %v\n", f.Name, f.URLPrivateDownload, err)
			continue // Skip this file
		}

		if err := os.WriteFile(localPath, fileData, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save file %s to %s: %v\n", f.Name, localPath, err)
			continue // Skip this file
		}

		exportedFiles = append(exportedFiles, ExportedFile{
			ID:        f.ID,
			Name:      f.Name,
			Mimetype:  f.Mimetype,
			LocalPath: localPath,
		})
	}
	return exportedFiles, nil
}
