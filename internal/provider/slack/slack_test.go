
package slack

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
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
	// For the actual file upload, the URL is already the mock server's URL, so we don't rewrite it.
	if strings.HasPrefix(req.URL.String(), t.serverURL.String()) {
		return http.DefaultTransport.RoundTrip(req)
	}
	req.URL.Scheme = t.serverURL.Scheme
	req.URL.Host = t.serverURL.Host
	return http.DefaultTransport.RoundTrip(req)
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
	mux := http.NewServeMux()
	mux.HandleFunc("/api/chat.postMessage", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok": true, "ts": "12345.67890"}`))
	})
	mux.HandleFunc("/api/conversations.list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok": true, "channels": [{"id": "C01", "name": "general"}]}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	p := newTestProvider(server, "general")

	opts := provider.PostMessageOptions{
		Text: "hello world",
	}

	if err := p.PostMessage(opts); err != nil {
		t.Errorf("PostMessage() returned an unexpected error: %v", err)
	}
}

func TestPostFile_Success(t *testing.T) {
	// Create a dummy file to upload
	tempDir := t.TempDir()
	filePath := tempDir + "/test.txt"
	if err := os.WriteFile(filePath, []byte("hello file"), 0666); err != nil {
		t.Fatal(err)
	}

	// The mock server needs its own URL to correctly form the upload_url
	var server *httptest.Server
	mux := http.NewServeMux()
	mux.HandleFunc("/api/files.getUploadURLExternal", func(w http.ResponseWriter, r *http.Request) {
		// This handler needs access to the server's URL, which is possible via this closure.
		uploadURL := server.URL + "/upload-here"
		resp := fmt.Sprintf(`{"ok": true, "upload_url": "%s", "file_id": "F01"}`, uploadURL)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(resp))
	})
	mux.HandleFunc("/upload-here", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/files.completeUploadExternal", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok": true, "files": []}`))
	})
	mux.HandleFunc("/api/conversations.list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok": true, "channels": [{"id": "C01", "name": "general"}]}`))
	})

	server = httptest.NewServer(mux)
	defer server.Close()

	p := newTestProvider(server, "general")

	opts := provider.PostFileOptions{
		FilePath: filePath,
		Filename: "test.txt",
		Comment:  "a test file",
	}

	if err := p.PostFile(opts); err != nil {
		t.Errorf("PostFile() returned an unexpected error: %v", err)
	}
}
