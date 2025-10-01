package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadMCConfig(t *testing.T) {
	// Create a temporary directory for test
	tempDir, err := os.MkdirTemp("", "mc-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create .mc directory
	mcDir := filepath.Join(tempDir, ".mc")
	err = os.MkdirAll(mcDir, 0755)
	require.NoError(t, err)

	// Test config data
	testConfig := MCConfig{
		Version: "10",
		Aliases: map[string]AliasConfig{
			"minio1": {
				URL:       "https://minio1.example.com",
				AccessKey: "testkey1",
				SecretKey: "testsecret1",
				API:       "s3v4",
				Path:      "auto",
				Insecure:  false,
			},
			"minio2": {
				URL:       "http://localhost:9000",
				AccessKey: "testkey2",
				SecretKey: "testsecret2",
				API:       "s3v4",
				Path:      "auto",
				Insecure:  true,
			},
		},
	}

	// Write test config
	configPath := filepath.Join(mcDir, "config.json")
	configData, err := json.MarshalIndent(testConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(configPath, configData, 0644)
	require.NoError(t, err)

	// Temporarily change HOME to our test directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Test loading the config
	loadedConfig, err := LoadMCConfig()
	require.NoError(t, err)
	assert.Equal(t, testConfig.Version, loadedConfig.Version)
	assert.Len(t, loadedConfig.Aliases, 2)
	assert.Equal(t, testConfig.Aliases["minio1"], loadedConfig.Aliases["minio1"])
	assert.Equal(t, testConfig.Aliases["minio2"], loadedConfig.Aliases["minio2"])
}

func TestLoadMCConfigError(t *testing.T) {
	// Test with non-existent config file
	tempDir, err := os.MkdirTemp("", "mc-config-error-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	_, err = LoadMCConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read MC config file")
}

func TestMCConfigStructure(t *testing.T) {
	config := MCConfig{
		Version: "10",
		Aliases: make(map[string]AliasConfig),
	}

	alias := AliasConfig{
		URL:       "https://example.com",
		AccessKey: "key",
		SecretKey: "secret",
		API:       "s3v4",
		Path:      "auto",
		Insecure:  false,
	}

	config.Aliases["test"] = alias

	assert.Equal(t, "10", config.Version)
	assert.Len(t, config.Aliases, 1)
	assert.Equal(t, alias, config.Aliases["test"])
}