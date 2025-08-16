package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	expectedLog := fmt.Sprintf("PostFile called with opts: {TargetChannel: TargetUserID: FilePath:%s Filename:%s Filetype: Comment: OverrideUsername: IconEmoji:}", filePath, filePath)
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
		_, _ = w.WriteString(content) // Modified: Ignore error return
	}()

	// Execute the command
	_, stderr, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "upload", "--file", "-")
	if err != nil {
		t.Fatalf("testExecuteCommandAndCapture returned an error: %v\nStderr: %s", err, stderr)
	}

	// Check if the test provider's PostFile was called with the correct options
	// Note: The exact temp file path is unknown, so we check for the known parts.
	if !strings.Contains(stderr, "PostFile called with opts: {TargetChannel: TargetUserID: FilePath:") {
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
	channel := "#override-channel"
	_, stderr, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "upload", "--file", filePath, "--comment", comment, "--filename", filename, "--filetype", filetype, "--channel", channel)
	if err != nil {
		t.Fatalf("testExecuteCommandAndCapture returned an error: %v\nStderr: %s", err, stderr)
	}

	// Check if the test provider's PostFile was called with the correct options
	expectedLog := fmt.Sprintf("PostFile called with opts: {TargetChannel:%s TargetUserID:%s FilePath:%s Filename:%s Filetype:%s Comment:%s OverrideUsername: IconEmoji:}", channel, "", filePath, filename, filetype, comment)
	if !strings.Contains(stderr, expectedLog) {
		t.Errorf("Expected stderr to contain '%s', got: '%s'", expectedLog, stderr)
	}
}

func TestUpload_ToUser(t *testing.T) {
	configPath, cleanup := setupTest(t)
	defer cleanup()

	// Create a dummy file to upload
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "upload-user.txt")
	if err := os.WriteFile(filePath, []byte("hello user"), 0666); err != nil {
		t.Fatal(err)
	}

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newUploadCmd())

	// Execute the command
	user := "U123ABCDE"
	_, stderr, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "upload", "--file", filePath, "--user", user)
	if err != nil {
		t.Fatalf("testExecuteCommandAndCapture returned an error: %v\nStderr: %s", err, stderr)
	}

	// Check if the test provider's PostFile was called with the correct options
	expectedLog := fmt.Sprintf("PostFile called with opts: {TargetChannel: TargetUserID:%s FilePath:%s Filename:%s Filetype: Comment: OverrideUsername: IconEmoji:}", user, filePath, filePath)
	if !strings.Contains(stderr, expectedLog) {
		t.Errorf("Expected stderr to contain '%s', got: '%s'", expectedLog, stderr)
	}
}

func TestUpload_UserAndChannelError(t *testing.T) {
	configPath, cleanup := setupTest(t)
	defer cleanup()

	// Create a dummy file to upload
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "upload-error.txt")
	if err := os.WriteFile(filePath, []byte("hello error"), 0666); err != nil {
		t.Fatal(err)
	}

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newUploadCmd())

	// Execute the command with both --user and --channel
	_, _, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "upload", "--file", filePath, "--user", "U123ABCDE", "--channel", "#test")
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}

	expectedError := "cannot use --user and --channel flags simultaneously"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got: '%s'", expectedError, err.Error())
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
