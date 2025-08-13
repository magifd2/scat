
package slack

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/magifd2/scat/internal/appcontext"
	"github.com/magifd2/scat/internal/config"
	"github.com/magifd2/scat/internal/provider"
)

// mockServerTransport is a custom http.RoundTripper that rewrites request URLs to the mock server.
type mockServerTransport struct {
	serverURL *url.URL
}

// RoundTrip rewrites the request URL's Scheme and Host to point to the mock server
// before delegating the request to the default transport.
func (t *mockServerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = t.serverURL.Scheme
	req.URL.Host = t.serverURL.Host
	return http.DefaultTransport.RoundTrip(req)
}

// mockSlackAPI creates a generic mock Slack API server that handles multiple endpoints.
func mockSlackAPI(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	// Handles chat.postMessage API calls
	mux.HandleFunc("/api/chat.postMessage", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		json.Unmarshal(body, &payload)

		if payload["text"] == "fail-me" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"ok": false, "error": "invalid_auth"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok": true, "ts": "12345.67890"}`))
	})

	// Handles conversations.list API calls
	mux.HandleFunc("/api/conversations.list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok": true, "channels": [{"id": "C01", "name": "general"}]}`))
	})

	return httptest.NewServer(mux)
}

// newTestProvider creates a provider configured to use the mock server.
func newTestProvider(server *httptest.Server, channelName string) *Provider {
	profile := config.Profile{Token: "test-token", Channel: channelName}
	ctx := appcontext.NewContext(false, false, false, "")
	serverURL, _ := url.Parse(server.URL)

	// Create a custom http client with our transport that rewrites URLs.
	client := &http.Client{
		Transport: &mockServerTransport{serverURL: serverURL},
	}

	p := &Provider{
		Profile:        profile,
		Context:        ctx,
		httpClient:     client,
		channelIDCache: make(map[string]string),
	}
	return p
}

func TestPostMessage_Success(t *testing.T) {
	server := mockSlackAPI(t)
	defer server.Close()

	p := newTestProvider(server, "general")

	opts := provider.PostMessageOptions{
		Text: "hello world",
	}

	if err := p.PostMessage(opts); err != nil {
		t.Errorf("PostMessage() returned an unexpected error: %v", err)
	}
}

func TestPostMessage_APIFailure(t *testing.T) {
	server := mockSlackAPI(t)
	defer server.Close()

	p := newTestProvider(server, "general")

	opts := provider.PostMessageOptions{
		Text: "fail-me", // This specific text triggers the failure case
	}

	err := p.PostMessage(opts)
	if err == nil {
		t.Fatal("PostMessage() did not return an error as expected")
	}
	if !strings.Contains(err.Error(), "invalid_auth") {
		t.Errorf("Expected error to contain 'invalid_auth', got: %v", err)
	}
}

func TestPostMessage_ChannelResolutionFailure(t *testing.T) {
	server := mockSlackAPI(t)
	defer server.Close()

	p := newTestProvider(server, "non-existent-channel")

	opts := provider.PostMessageOptions{
		Text: "hello world",
	}

	err := p.PostMessage(opts)
	if err == nil {
		t.Fatal("PostMessage() with non-existent channel did not return an error")
	}
	// The actual error message comes from channel.go, let's match it.
	expectedError := "channel \"non-existent-channel\" not found"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain \"%s\", got: %v", expectedError, err)
	}
}
