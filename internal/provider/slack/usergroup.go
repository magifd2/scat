package slack

import (
	"fmt"
	"os"
	"strings"
)

// populateUserGroupCache fetches all user groups and populates the userGroupIDCache.
func (p *Provider) populateUserGroupCache() error {
	userGroups, err := p.getUserGroups()
	if err != nil {
		return fmt.Errorf("failed to get user groups for cache: %w", err)
	}

	p.userGroupIDCache = make(map[string]string)
	for _, ug := range userGroups {
		p.userGroupIDCache[ug.Handle] = ug.ID
		if p.Context.Debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Caching user group: Handle=%s, ID=%s\n", ug.Handle, ug.ID)
		}
	}

	if p.Context.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] User group cache populated with %d groups.\n", len(p.userGroupIDCache))
	}
	return nil
}

// ResolveUserGroupID finds a user group ID for a given user group handle.
// It checks the cache first, and repopulates it if the user group is not found.
func (p *Provider) ResolveUserGroupID(handle string) (string, error) {
	cleanHandle := strings.TrimPrefix(handle, "@")

	id, ok := p.userGroupIDCache[cleanHandle]
	if ok {
		return id, nil
	}

	// If not found, refresh the cache and try again.
	if p.Context.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] User group '%s' not found in cache, repopulating...\n", cleanHandle)
	}
	if err := p.populateUserGroupCache(); err != nil {
		return "", fmt.Errorf("failed to repopulate user group cache: %w", err)
	}

	id, ok = p.userGroupIDCache[cleanHandle]
	if !ok {
		return "", fmt.Errorf("user group '%s' not found", handle)
	}

	return id, nil
}
