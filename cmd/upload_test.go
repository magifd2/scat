package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/magifd2/scat/internal/provider/testprovider"
)

func TestUpload_FromFile(t *testing.T) {
	configPath, cleanup := setupTest(t)
	defer cleanup()

	// Create a dummy file to upload
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "upload-test.txt")
	if err := os.WriteFile(filePath, []byte("hello upload"), 0666); err != nil {
		t.Fatal(err)
	}

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newUploadCmd())

	// Execute the command
	_, stderr, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "upload", "--file", filePath)
	if err != nil {
		t.Fatalf("testExecuteCommandAndCapture returned an error: %v\nStderr: %s", err, stderr)
	}

	// Check if the test provider's PostFile was called with the correct options
	expectedLog := fmt.Sprintf("PostFile called with opts: {FilePath:%s Filename:%s Filetype: Comment: OverrideUsername: IconEmoji:}", filePath, filePath)
	if !strings.Contains(stderr, expectedLog) {
		t.Errorf("Expected stderr to contain '%s', got: '%s'", expectedLog, stderr)
	}
}

func TestUpload_FromStdin(t *testing.T) {
	configPath, cleanup := setupTest(t)
	defer cleanup()

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newUploadCmd())

	// Simulate stdin
	content := "hello from stdin"
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		defer w.Close()
		w.WriteString(content)
	}()

	// Execute the command
	_, stderr, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "upload", "--file", "-")
	if err != nil {
		t.Fatalf("testExecuteCommandAndCapture returned an error: %v\nStderr: %s", err, stderr)
	}

	// Check if the test provider's PostFile was called with the correct options
	// Note: The exact temp file path is unknown, so we check for the known parts.
	if !strings.Contains(stderr, "PostFile called with opts: {FilePath:") {
		t.Errorf("Expected stderr to contain PostFile marker, got: %s", stderr)
	}
	if !strings.Contains(stderr, "Filename:stdin-upload Filetype: Comment: OverrideUsername: IconEmoji:}") {
		t.Errorf("Expected stderr to contain correct filename, got: %s", stderr)
	}
}

func TestUpload_WithOptions(t *testing.T) {
	configPath, cleanup := setupTest(t)
	defer cleanup()

	// Create a dummy file to upload
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "upload-options.txt")
	if err := os.WriteFile(filePath, []byte("hello options"), 0666); err != nil {
		t.Fatal(err)
	}

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newUploadCmd())

	// Execute the command
	comment := "this is a comment"
	filename := "new-name.txt"
	filetype := "text"
	_, stderr, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "upload", "--file", filePath, "--comment", comment, "--filename", filename, "--filetype", filetype)
	if err != nil {
		t.Fatalf("testExecuteCommandAndCapture returned an error: %v\nStderr: %s", err, stderr)
	}

	// Check if the test provider's PostFile was called with the correct options
	expectedLog := fmt.Sprintf("PostFile called with opts: {FilePath:%s Filename:%s Filetype:%s Comment:%s OverrideUsername: IconEmoji:}", filePath, filename, filetype, comment)
	if !strings.Contains(stderr, expectedLog) {
		t.Errorf("Expected stderr to contain '%s', got: '%s'", expectedLog, stderr)
	}
}

func TestUpload_NoFile(t *testing.T) {
	configPath, cleanup := setupTest(t)
	defer cleanup()

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newUploadCmd())

	// Execute the command without the required --file flag
	_, _, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "upload")
	if err == nil {
		t.Fatal("Expected an error for missing --file flag, but got nil")
	}

	// Check the error message
	expectedError := "required flag(s) \"file\" not set"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error message to contain '%s', got: '%v'", expectedError, err)
	}
}
