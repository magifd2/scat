package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// setupTest an initial setup for tests
func setupTest(t *testing.T) (string, func()) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	configContent := `{
		"current_profile": "test",
		"profiles": {
			"test": {
				"provider": "test",
				"channel": "#test-channel"
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(configContent), 0666); err != nil {
		t.Fatal(err)
	}

	return configPath, func() {
		// Cleanup if necessary
	}
}

// testExecuteCommandAndCapture is a helper function for testing cobra commands.
// It executes a cobra command and captures its stdout and stderr.
func testExecuteCommandAndCapture(root *cobra.Command, args ...string) (stdout string, stderr string, err error) {
	// Redirect stdout and stderr to buffers
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// Execute the command
	root.SetArgs(args)
	err = root.Execute()

	// Restore stdout/stderr and read the captured output
	_ = wOut.Close()
	_ = wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var bufOut, bufErr bytes.Buffer
	_, _ = io.Copy(&bufOut, rOut)
	_, _ = io.Copy(&bufErr, rErr)

	stdout = bufOut.String()
	stderr = bufErr.String()
	return
}