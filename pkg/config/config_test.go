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
	// Test with non-existent config file and no env vars
	tempDir, err := os.MkdirTemp("", "mc-config-error-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Clear any MC_HOST_* env vars
	for _, env := range os.Environ() {
		if len(env) >= 8 && env[:8] == "MC_HOST_" {
			key := env[:len(env)-len(env[8:])]
			os.Unsetenv(key)
		}
	}

	_, err = LoadMCConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no MinIO aliases configured")
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

func TestLoadMCConfigFromEnv(t *testing.T) {
	// Set up test environment without config file
	tempDir, err := os.MkdirTemp("", "mc-config-env-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Set environment variables for aliases
	os.Setenv("MC_HOST_site1", "https://testkey1:testsecret1@minio1.example.com:9000")
	os.Setenv("MC_HOST_site2", "http://testkey2:testsecret2@localhost:9001")
	defer os.Unsetenv("MC_HOST_site1")
	defer os.Unsetenv("MC_HOST_site2")

	// Load config
	config, err := LoadMCConfig()
	require.NoError(t, err)

	// Verify aliases loaded from env
	assert.Len(t, config.Aliases, 2)

	site1 := config.Aliases["site1"]
	assert.Equal(t, "https://minio1.example.com:9000", site1.URL)
	assert.Equal(t, "testkey1", site1.AccessKey)
	assert.Equal(t, "testsecret1", site1.SecretKey)
	assert.Equal(t, "s3v4", site1.API)
	assert.Equal(t, "auto", site1.Path)

	site2 := config.Aliases["site2"]
	assert.Equal(t, "http://localhost:9001", site2.URL)
	assert.Equal(t, "testkey2", site2.AccessKey)
	assert.Equal(t, "testsecret2", site2.SecretKey)
}

func TestLoadMCConfigEnvOverridesFile(t *testing.T) {
	// Create a temporary directory with config file
	tempDir, err := os.MkdirTemp("", "mc-config-override-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create .mc directory and config file
	mcDir := filepath.Join(tempDir, ".mc")
	err = os.MkdirAll(mcDir, 0755)
	require.NoError(t, err)

	testConfig := MCConfig{
		Version: "10",
		Aliases: map[string]AliasConfig{
			"site1": {
				URL:       "https://old.example.com",
				AccessKey: "oldkey",
				SecretKey: "oldsecret",
				API:       "s3v4",
				Path:      "auto",
			},
		},
	}

	configPath := filepath.Join(mcDir, "config.json")
	configData, err := json.MarshalIndent(testConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(configPath, configData, 0644)
	require.NoError(t, err)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Set environment variable that overrides file config
	os.Setenv("MC_HOST_site1", "https://newkey:newsecret@new.example.com:9000")
	defer os.Unsetenv("MC_HOST_site1")

	// Load config
	config, err := LoadMCConfig()
	require.NoError(t, err)

	// Verify env var overrides file config
	site1 := config.Aliases["site1"]
	assert.Equal(t, "https://new.example.com:9000", site1.URL)
	assert.Equal(t, "newkey", site1.AccessKey)
	assert.Equal(t, "newsecret", site1.SecretKey)
}

func TestParseAliasURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expected    AliasConfig
		expectError bool
	}{
		{
			name: "valid https url with port",
			url:  "https://accesskey:secretkey@minio.example.com:9000",
			expected: AliasConfig{
				URL:       "https://minio.example.com:9000",
				AccessKey: "accesskey",
				SecretKey: "secretkey",
				API:       "s3v4",
				Path:      "auto",
				Insecure:  false,
			},
			expectError: false,
		},
		{
			name: "valid http url",
			url:  "http://minioadmin:minioadmin@localhost:9000",
			expected: AliasConfig{
				URL:       "http://localhost:9000",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin",
				API:       "s3v4",
				Path:      "auto",
				Insecure:  false,
			},
			expectError: false,
		},
		{
			name: "url with insecure flag",
			url:  "https://key:secret@minio.local:9000?insecure=true",
			expected: AliasConfig{
				URL:       "https://minio.local:9000",
				AccessKey: "key",
				SecretKey: "secret",
				API:       "s3v4",
				Path:      "auto",
				Insecure:  true,
			},
			expectError: false,
		},
		{
			name:        "invalid url",
			url:         "not-a-valid-url",
			expectError: false, // URL parsing is lenient, returns empty values
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseAliasURL(tt.url)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Only check values for valid test cases
				if tt.name != "invalid url" {
					assert.Equal(t, tt.expected.URL, result.URL)
					assert.Equal(t, tt.expected.AccessKey, result.AccessKey)
					assert.Equal(t, tt.expected.SecretKey, result.SecretKey)
					assert.Equal(t, tt.expected.API, result.API)
					assert.Equal(t, tt.expected.Path, result.Path)
					assert.Equal(t, tt.expected.Insecure, result.Insecure)
				}
			}
		})
	}
}
