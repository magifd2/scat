package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/magifd2/scat/internal/config"
)

func TestProfileUse_Success(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	configContent := `{
		"current_profile": "default",
		"profiles": {
			"default": {
				"provider": "mock"
			},
			"test_profile": {
				"provider": "slack"
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
	stdout, stderr, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "profile", "use", "test_profile")
	if err != nil {
		t.Fatalf("testExecuteCommandAndCapture returned an error: %v\nStderr: %s", err, stderr)
	}

	// Check stderr output
	if !strings.Contains(stderr, "Switched to profile: test_profile") {
		t.Errorf("Expected stderr to contain 'Switched to profile: test_profile', got: %s", stderr)
	}

	// Verify the config file was updated
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config after use command: %v", err)
	}
	if cfg.CurrentProfile != "test_profile" {
		t.Errorf("Expected current profile to be 'test_profile', got: %s", cfg.CurrentProfile)
	}

	if stdout != "" {
		t.Errorf("Expected no stdout, got: %s", stdout)
	}
}

func TestProfileUse_ProfileNotFound(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	configContent := `{
		"current_profile": "default",
		"profiles": {
			"default": {
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

	// Execute the command with a non-existent profile
	stdout, stderr, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "profile", "use", "non_existent_profile")
	if err == nil {
		t.Fatal("Expected an error for non-existent profile, but got nil")
	}

	// Check the error message
	if !strings.Contains(err.Error(), "profile 'non_existent_profile' not found") {
		t.Errorf("Expected error message to contain 'profile not found', got: %v", err)
	}

	if stdout != "" {
		t.Errorf("Expected no stdout, got: %s", stdout)
	}

	// Stderr should be empty for this error type
	if stderr != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr)
	}
}

func TestProfileUse_NoConfig(t *testing.T) {
	// Point to a non-existent config file
	configPath := filepath.Join(t.TempDir(), "non-existent.json")

	// Create a clean command tree for the test
	rootCmd := newRootCmd()
	rootCmd.AddCommand(newProfileCmd())

	// Execute the command
	stdout, stderr, err := testExecuteCommandAndCapture(rootCmd, "--config", configPath, "profile", "use", "any_profile")
	if err == nil {
		t.Fatal("Expected an error for non-existent config, but got nil")
	}

	// Check the error message
	if !strings.Contains(err.Error(), "configuration file not found") {
		t.Errorf("Expected error message to contain 'configuration file not found', got: %v", err)
	}

	if stdout != "" {
		t.Errorf("Expected no stdout, got: %s", stdout)
	}

	// Stderr should be empty for this error type
	if stderr != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr)
	}
}