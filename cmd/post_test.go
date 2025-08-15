package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPost_FromArgument(t *testing.T) {
	configPath, cleanup := setupTest(t)
	defer cleanup()

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newPostCmd())

	// Execute the command
	message := "hello from test"
	_, stderr, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "post", message)
	if err != nil {
		t.Fatalf("testExecuteCommandAndCapture returned an error: %v\nStderr: %s", err, stderr)
	}

	// Check if the test provider's PostMessage was called with the correct options
	expectedLog := fmt.Sprintf("PostMessage called with opts: {Text:%s OverrideUsername: IconEmoji: Blocks:}", message)
	if !strings.Contains(stderr, expectedLog) {
		t.Errorf("Expected stderr to contain '%s', got: '%s'", expectedLog, stderr)
	}
}

func TestPost_FromFile(t *testing.T) {
	configPath, cleanup := setupTest(t)
	defer cleanup()

	// Create a temporary file with content
	message := "hello from file"
	file, err := os.CreateTemp(t.TempDir(), "test-message-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	if _, err := file.WriteString(message); err != nil {
		t.Fatal(err)
	}
	file.Close()

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newPostCmd())

	// Execute the command
	_, stderr, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "post", "--from-file", file.Name())
	if err != nil {
		t.Fatalf("testExecuteCommandAndCapture returned an error: %v\nStderr: %s", err, stderr)
	}

	// Check if the test provider's PostMessage was called with the correct options
	expectedLog := fmt.Sprintf("PostMessage called with opts: {Text:%s OverrideUsername: IconEmoji: Blocks:}", message)
	if !strings.Contains(stderr, expectedLog) {
		t.Errorf("Expected stderr to contain '%s', got: '%s'", expectedLog, stderr)
	}
}

func TestPost_FromStdin(t *testing.T) {
	configPath, cleanup := setupTest(t)
	defer cleanup()

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newPostCmd())

	// Simulate stdin
	message := "hello from stdin"
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		defer w.Close()
		_, _ = w.WriteString(message) // Modified: Ignore error return
	}()

	// Execute the command
	_, stderr, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "post")
	if err != nil {
		t.Fatalf("testExecuteCommandAndCapture returned an error: %v\nStderr: %s", err, stderr)
	}

	// Check if the test provider's PostMessage was called with the correct options
	expectedLog := fmt.Sprintf("PostMessage called with opts: {Text:%s OverrideUsername: IconEmoji: Blocks:}", message)
	if !strings.Contains(stderr, expectedLog) {
		t.Errorf("Expected stderr to contain '%s', got: '%s'", expectedLog, stderr)
	}
}

func TestPost_WithOptions(t *testing.T) {
	configPath, cleanup := setupTest(t)
	defer cleanup()

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newPostCmd())

	// Execute the command with options
	message := "hello with options"
	username := "test-user"
	iconEmoji := ":tada:"
	channel := "#override-channel"
	_, stderr, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "post",
		"--username", username,
		"--iconemoji", iconEmoji,
		"--channel", channel,
		message,
	)
	if err != nil {
		t.Fatalf("testExecuteCommandAndCapture returned an error: %v\nStderr: %s", err, stderr)
	}

	// Check if the test provider's PostMessage was called with the correct options
	expectedLog := fmt.Sprintf("PostMessage called with opts: {Text:%s OverrideUsername:%s IconEmoji:%s Blocks:}", message, username, iconEmoji)
	if !strings.Contains(stderr, expectedLog) {
		t.Errorf("Expected stderr to contain '%s', got: '%s'", expectedLog, stderr)
	}
}

func TestPost_NoMessage(t *testing.T) {
	configPath, cleanup := setupTest(t)
	defer cleanup()

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newPostCmd())

	// Execute the command without a message
	_, _, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "post")
	if err == nil {
		t.Fatal("Expected an error for no message, but got nil")
	}

	// Check the error message
	expectedError := "no message content provided via argument, --from-file, or stdin"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error message to contain '%s', got: %v", expectedError, err)
	}
}

func TestPost_Stream(t *testing.T) {
	configPath, cleanup := setupTest(t)
	defer cleanup()

	// Mock the ticker
	oldCreateTicker := CreateTicker
	tickerChan := make(chan time.Time, 1) // Use a buffered channel
	CreateTicker = func(d time.Duration) *time.Ticker {
		// Return a ticker that uses our manual channel
		// Do not call time.NewTicker() to avoid its internal goroutine.
		return &time.Ticker{C: tickerChan}
	}
	// Restore the original function after the test
	defer func() { CreateTicker = oldCreateTicker }()

	// Simulate stdin
	message := "line 1\nline 2"
	r, w, _ := os.Pipe()
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Command execution and stderr capture
	var stderr string
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		rootCmd := newRootCmd()
		rootCmd.AddCommand(newPostCmd())
		_, stderr, _ = testExecuteCommandAndCapture(rootCmd, "--config", configPath, "post", "--stream")
	}()

	// Write to stdin in a separate goroutine and close the writer when done.
	go func() {
		_, _ = w.WriteString(message) // Modified: Ignore error return
		_ = w.Close()                 // Modified: Ignore error return
	}()

	// Wait a moment for the command to start up and read the input.
	time.Sleep(100 * time.Millisecond)

	// Trigger the ticker, allowing the buffered content to be posted.
	tickerChan <- time.Now()

	// Wait for the command execution goroutine to finish.
	wg.Wait()

	// Check if the test provider's PostMessage was called with the correct, combined text.
	expectedLog := "Text:line 1\nline 2"
	if !strings.Contains(stderr, expectedLog) {
		t.Errorf("Expected stderr to contain '%s', got: '%s'", expectedLog, stderr)
	}
}

func TestPost_BlockKitFormat_FromArgument(t *testing.T) {
	configPath, cleanup := setupTest(t)
	defer cleanup()

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newPostCmd())

	blockKitJSON := `[{"type": "section", "text": {"type": "mrkdwn", "text": "Hello, Block Kit!"}}]`

	_, stderr, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "post", "--format", "blocks", blockKitJSON)
	if err != nil {
		t.Fatalf("testExecuteCommandAndCapture returned an error: %v\nStderr: %s", err, stderr)
	}

	expectedLog := fmt.Sprintf("PostMessage called with opts: {Text: OverrideUsername: IconEmoji: Blocks:%s}", blockKitJSON)
	if !strings.Contains(stderr, expectedLog) {
		t.Errorf("Expected stderr to contain '%s', got: '%s'", expectedLog, stderr)
	}
}

func TestPost_BlockKitFormat_FromFile(t *testing.T) {
	configPath, cleanup := setupTest(t)
	defer cleanup()

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newPostCmd())

	blockKitJSON := `[{"type": "section", "text": {"type": "mrkdwn", "text": "Hello from file!"}}]`
	file, err := os.CreateTemp(t.TempDir(), "block-kit-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	if _, err := file.WriteString(blockKitJSON); err != nil {
		t.Fatal(err)
	}
	file.Close()

	_, stderr, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "post", "--format", "blocks", "--from-file", file.Name())
	if err != nil {
		t.Fatalf("testExecuteCommandAndCapture returned an error: %v\nStderr: %s", err, stderr)
	}

	expectedLog := fmt.Sprintf("PostMessage called with opts: {Text: OverrideUsername: IconEmoji: Blocks:%s}", blockKitJSON)
	if !strings.Contains(stderr, expectedLog) {
		t.Errorf("Expected stderr to contain '%s', got: '%s'", expectedLog, stderr)
	}
}

func TestPost_BlockKitFormat_FromStdin(t *testing.T) {
	configPath, cleanup := setupTest(t)
	defer cleanup()

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newPostCmd())

	blockKitJSON := `[{"type": "section", "text": {"type": "mrkdwn", "text": "Hello from stdin!"}}]`
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		defer w.Close()
		_, _ = w.WriteString(blockKitJSON)
	}()

	_, stderr, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "post", "--format", "blocks")
	if err != nil {
		t.Fatalf("testExecuteCommandAndCapture returned an error: %v\nStderr: %s", err, stderr)
	}

	expectedLog := fmt.Sprintf("PostMessage called with opts: {Text: OverrideUsername: IconEmoji: Blocks:%s}", blockKitJSON)
	if !strings.Contains(stderr, expectedLog) {
		t.Errorf("Expected stderr to contain '%s', got: '%s'", expectedLog, stderr)
	}
}

func TestPost_BlockKitFormat_InvalidJson(t *testing.T) {
	configPath, cleanup := setupTest(t)
	defer cleanup()

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newPostCmd())

	invalidJSON := `{"type": "section", "text": {"type": "mrkdwn", "text": "Hello, Block Kit!"}` // Missing closing brace

	_, _, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "post", "--format", "blocks", invalidJSON)
	if err == nil {
		t.Fatal("Expected an error for invalid JSON, but got nil")
	}

	if !strings.Contains(err.Error(), "failed to parse block kit JSON") {
		t.Errorf("Expected error message to contain 'failed to parse block kit JSON', got: %v", err)
	}
}

func TestPost_BlockKitFormatAndStreamError(t *testing.T) {
	configPath, cleanup := setupTest(t)
	defer cleanup()

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newPostCmd())

	_, _, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "post", "--format", "blocks", "--stream")
	if err == nil {
		t.Fatal("Expected an error for --format blocks and --stream, but got nil")
	}

	if !strings.Contains(err.Error(), "cannot use --stream with --format blocks") {
		t.Errorf("Expected error message to contain 'cannot use --stream with --format blocks', got: %v", err)
	}
}

func TestPost_InvalidFormatFlag(t *testing.T) {
	configPath, cleanup := setupTest(t)
	defer cleanup()

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newPostCmd())

	_, _, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "post", "--format", "invalid-format", "test message")
	if err == nil {
		t.Fatal("Expected an error for invalid --format flag, but got nil")
	}

	if !strings.Contains(err.Error(), "invalid value for --format") {
		t.Errorf("Expected error message to contain 'invalid value for --format', got: %v", err)
	}
}
