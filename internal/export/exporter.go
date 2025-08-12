package export

import (
	"fmt"
	"path/filepath"
	"sync"

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
	// This is a placeholder for the main export logic.
	// It will involve:
	// 1. Paginating through GetConversationHistory.
	// 2. For each message, resolving the user name using resolveUserName.
	// 3. If opts.IncludeFiles, downloading files for each message.
	// 4. Assembling the final ExportedLog struct.

	return nil, fmt.Errorf("ExportLog not yet implemented")
}

// resolveUserName gets a user's name from the cache or fetches it from the provider.
func (e *Exporter) resolveUserName(userID string) (string, error) {
	e.userCacheMux.Lock()
	defer e.userCacheMux.Unlock()

	if name, ok := e.userCache[userID]; ok {
		return name, nil
	}

	userInfo, err := e.prov.GetUserInfo(userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user info for %s: %w", userID, err)
	}

	name := userInfo.User.RealName
	if name == "" {
		name = userInfo.User.Name
	}

	e.userCache[userID] = name
	return name, nil
}

// handleAttachedFiles downloads files attached to a message.
func (e *Exporter) handleAttachedFiles(files []provider.File, outputDir string) ([]ExportedFile, error) {
	var exportedFiles []ExportedFile
	for _, f := range files {
		localPath := filepath.Join(outputDir, f.ID+"_"+f.Name)

		// Download logic will be added here.

		exportedFiles = append(exportedFiles, ExportedFile{
			ID:        f.ID,
			Name:      f.Name,
			Mimetype:  f.Mimetype,
			LocalPath: localPath, // Will be empty if download fails or is skipped
		})
	}
	return exportedFiles, nil
}
