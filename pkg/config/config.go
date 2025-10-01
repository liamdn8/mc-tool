package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// MCConfig represents the MinIO client configuration
type MCConfig struct {
	Version string                 `json:"version"`
	Aliases map[string]AliasConfig `json:"aliases"`
}

// AliasConfig represents a single alias configuration
type AliasConfig struct {
	URL       string `json:"url"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	API       string `json:"api"`
	Path      string `json:"path"`
	Insecure  bool   `json:"insecure,omitempty"`
}

// LoadMCConfig loads the MinIO client configuration from the default location
func LoadMCConfig() (*MCConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	configPath := filepath.Join(homeDir, ".mc", "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read MC config file: %v", err)
	}

	var config MCConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse MC config: %v", err)
	}

	return &config, nil
}