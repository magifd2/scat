package slack

import (
	"encoding/json"
	"regexp"
	"sync"
)

var mentionRegex = regexp.MustCompile(`<@(U[A-Z0-9]+)>`)

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
