package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// executeCommand is a helper function to execute a cobra command and capture its output.
func executeCommand(root *cobra.Command, args ...string) (string, error) {
	// Redirect stdout to a buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the command
	root.SetArgs(args)
	err := root.Execute()

	// Restore stdout and read the captured output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String(), err
}

func TestProfileList_Success(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	configContent := `{`
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

	// We need to re-initialize rootCmd for each test to have a clean state
	rootCmd, _ := newRootCmd()
	rootCmd.AddCommand(profileCmd)
	profileCmd.AddCommand(profileListCmd)

	// Execute the command
	output, err := executeCommand(rootCmd, "--config", configPath, "profile", "list")
	if err != nil {
		t.Fatalf("executeCommand returned an error: %v", err)
	}

	// Check the output
	expectedProfiles := []string{"* default", "  profile1", "  profile2"}
	for _, p := range expectedProfiles {
		if !strings.Contains(output, p) {
			t.Errorf("Output does not contain expected profile line \"%s\". Got:\n%s", p, output)
		}
	}
}

func TestProfileList_NoConfig(t *testing.T) {
	// Point to a non-existent config file
	configPath := filepath.Join(t.TempDir(), "non-existent.json")

	rootCmd, _ := newRootCmd()
	rootCmd.AddCommand(profileCmd)
	profileCmd.AddCommand(profileListCmd)

	// Execute the command
	_, err := executeCommand(rootCmd, "--config", configPath, "profile", "list")
	if err == nil {
		t.Fatal("Expected an error for non-existent config, but got nil")
	}

	// Check the error message
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected error message to contain 'not found', got: %v", err)
	}
}
