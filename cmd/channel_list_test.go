
package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	_ "github.com/magifd2/scat/internal/provider/testprovider"
)

func TestChannelList_Default(t *testing.T) {
	configPath, cleanup := setupTest(t)
	defer cleanup()

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newChannelCmd())

	// Execute the command
	stdout, stderr, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "channel", "list")
	if err != nil {
		t.Fatalf("testExecuteCommandAndCapture returned an error: %v\nStderr: %s", err, stderr)
	}

	// Check the output
	expectedOutput := "- #test-channel-1\n  - #test-channel-2"
	if !strings.Contains(stdout, expectedOutput) {
		t.Errorf("Expected stdout to contain '%s', got: '%s'", expectedOutput, stdout)
	}

	// Check stderr for the profile header
	expectedStderr := "Channels for profile: test"
	if !strings.Contains(stderr, expectedStderr) {
		t.Errorf("Expected stderr to contain '%s', got: '%s'", expectedStderr, stderr)
	}
}

func TestChannelList_JSON(t *testing.T) {
	configPath, cleanup := setupTest(t)
	defer cleanup()

	rootCmd := newRootCmd()
	rootCmd.AddCommand(newChannelCmd())

	// Execute the command
	stdout, _, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "channel", "list", "--json")
	if err != nil {
		t.Fatalf("testExecuteCommandAndCapture returned an error: %v", err)
	}

	// Check the JSON output
	var result map[string][]string
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to unmarshal json output: %v", err)
	}

	if _, ok := result["test"]; !ok {
		t.Fatal("Expected 'test' profile in json output")
	}

	if len(result["test"]) != 2 {
		t.Errorf("Expected 2 channels for test profile, got %d", len(result["test"]))
	}

	if result["test"][0] != "#test-channel-1" {
		t.Errorf("Unexpected channel name: got %s, want #test-channel-1", result["test"][0])
	}
}
