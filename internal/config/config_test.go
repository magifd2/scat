package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()

	if cfg.CurrentProfile != "default" {
		t.Errorf("expected current profile to be 'default', got '%s'", cfg.CurrentProfile)
	}
	if len(cfg.Profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(cfg.Profiles))
	}
	if _, ok := cfg.Profiles["default"]; !ok {
		t.Fatal("expected 'default' profile to exist")
	}
	defaultProfile := cfg.Profiles["default"]
	if defaultProfile.Provider != "mock" {
		t.Errorf("expected default provider to be 'mock', got '%s'", defaultProfile.Provider)
	}
}

func TestSaveLoad(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "config.json")

	// 1. Create a config, save it
	cfg1 := NewDefaultConfig()
	err := cfg1.Save(testPath)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// 2. Load it back
	cfg2, err := Load(testPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// 3. Compare
	if !reflect.DeepEqual(cfg1, cfg2) {
		t.Errorf("loaded config does not match saved config")
		t.Logf("Saved: %+v", cfg1)
		t.Logf("Loaded: %+v", cfg2)
	}
}

func TestLoad_FileNotExists(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "non_existent_config.json")

	_, err := Load(testPath)
	if err == nil {
		t.Fatal("expected an error when loading non-existent file, but got nil")
	}
}

func TestLoad_InvalidJson(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "invalid.json")

	err := os.WriteFile(testPath, []byte("{ not a valid json }"), 0600)
	if err != nil {
		t.Fatalf("failed to write invalid json file: %v", err)
	}

	_, err = Load(testPath)
	if err == nil {
		t.Fatal("expected an error when loading invalid json, but got nil")
	}
}

func TestProfileManagement_DirectManipulation(t *testing.T) {
	cfg := NewDefaultConfig()

	// Add a new profile by direct map manipulation
	newProfile := Profile{
		Provider: "slack",
		Channel:  "#general",
		Limits:   NewDefaultLimits(),
	}
	cfg.Profiles["my-slack"] = newProfile

	if len(cfg.Profiles) != 2 {
		t.Errorf("expected 2 profiles after adding, got %d", len(cfg.Profiles))
	}
	if !reflect.DeepEqual(cfg.Profiles["my-slack"], newProfile) {
		t.Error("added profile does not match original")
	}

	// Set current profile by direct field manipulation
	cfg.CurrentProfile = "my-slack"
	if cfg.CurrentProfile != "my-slack" {
		t.Errorf("expected current profile to be 'my-slack', got '%s'", cfg.CurrentProfile)
	}

	// Remove profile by using delete
	delete(cfg.Profiles, "my-slack")
	if len(cfg.Profiles) != 1 {
		t.Errorf("expected 1 profile after removing, got %d", len(cfg.Profiles))
	}
	if _, ok := cfg.Profiles["my-slack"]; ok {
		t.Error("removed profile 'my-slack' still exists")
	}
}