package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	configDir  = ".config"
	configFile = "scat/config.json"
)

// Config represents the overall structure of the application's configuration file.
type Config struct {
	CurrentProfile string             `json:"current_profile"`
	Profiles       map[string]Profile `json:"profiles"`
}

// Profile defines the settings for a specific destination endpoint.
type Profile struct {
	Provider string `json:"provider,omitempty"` // "generic" or "slack"
	Endpoint string `json:"endpoint,omitempty"` // Used by "generic" provider
	Channel  string `json:"channel,omitempty"`  // Used by "slack" provider
	Token    string `json:"token,omitempty"`
	Username string `json:"username,omitempty"`
	Limits   Limits `json:"limits,omitempty"`
}

// Limits defines the size limits for inputs.
type Limits struct {
	MaxFileSizeBytes int64 `json:"max_file_size_bytes,omitempty"`
	MaxStdinSizeBytes int64 `json:"max_stdin_size_bytes,omitempty"`
}

// newDefaultLimits returns a Limits struct with default values.
func newDefaultLimits() Limits {
	return Limits{
		MaxFileSizeBytes: 1024 * 1024 * 1024, // 1 GB
		MaxStdinSizeBytes: 10 * 1024 * 1024,  // 10 MB
	}
}

// Load reads the configuration file from the user's config directory.
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If the config file does not exist, return a default configuration.
			return &Config{
				CurrentProfile: "default",
				Profiles: map[string]Profile{
					"default": {
						Provider: "mock",
						Channel:  "#mock-channel",
						Limits:   newDefaultLimits(),
					},
				},
			}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// For backward compatibility, populate limits if they are not set.
	for name, profile := range cfg.Profiles {
		if profile.Provider == "" {
			profile.Provider = "generic" // Or mock, depending on desired behavior for old configs
		}
		if profile.Limits.MaxFileSizeBytes == 0 && profile.Limits.MaxStdinSizeBytes == 0 {
			profile.Limits = newDefaultLimits()
		}
		cfg.Profiles[name] = profile
	}

	return &cfg, nil
}

// Save writes the current configuration to the user's config directory.
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// GetConfigPath returns the absolute path to the configuration file.
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDir, configFile), nil
}
