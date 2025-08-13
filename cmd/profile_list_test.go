package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProfileList_Success(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	// Correctly define the multi-line string literal
	configContent := `{
		"current_profile": "default",
		"profiles": {
			"default": {
				"provider": "mock"
			},
			"profile1": {
				"provider": "slack"
			},
			"profile2": {
				"provider": "mock"
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(configContent), 0666); err != nil {
		t.Fatal(err)
	}

	// Create a clean command tree for the test
	rootCmd := newRootCmd()
	rootCmd.AddCommand(newProfileCmd())

	// Execute the command
	stdout, stderr, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "profile", "list")
	if err != nil {
		t.Fatalf("testExecuteCommandAndCapture returned an error: %v\nStderr: %s", err, stderr)
	}

	// Check the output
	expectedProfiles := []string{"* default", "  profile1", "  profile2"}
	for _, p := range expectedProfiles {
		if !strings.Contains(stdout, p) {
			t.Errorf("Output does not contain expected profile line \"%s\". Got:\n%s", p, stdout)
		}
	}

	// Stderr should be empty for this command
	if stderr != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr)
	}
}

func TestProfileList_NoConfig(t *testing.T) {
	// Point to a non-existent config file
	configPath := filepath.Join(t.TempDir(), "non-existent.json")

	// Create a clean command tree for the test
	rootCmd := newRootCmd()
	rootCmd.AddCommand(newProfileCmd())

	// Execute the command
	stdout, stderr, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "profile", "list")
	if err == nil {
		t.Fatal("Expected an error for non-existent config, but got nil")
	}

	// Check the error message
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected error message to contain 'not found', got: %v", err)
	}

	// Stderr should be empty for this error type
	if stderr != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr)
	}

	if stdout != "" {
		t.Errorf("Expected no stdout, got: %s", stdout)
	}
}