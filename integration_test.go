package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHelper provides utilities for testing
type TestHelper struct {
	TempDir    string
	ConfigPath string
}

// NewTestHelper creates a new test helper with temporary directory
func NewTestHelper(t *testing.T) *TestHelper {
	tempDir := t.TempDir()
	mcDir := filepath.Join(tempDir, ".mc")
	err := os.MkdirAll(mcDir, 0755)
	require.NoError(t, err)

	return &TestHelper{
		TempDir:    tempDir,
		ConfigPath: filepath.Join(mcDir, "config.json"),
	}
}

// CreateTestConfig creates a test configuration file
func (th *TestHelper) CreateTestConfig(t *testing.T, config MCConfig) {
	configData, err := json.MarshalIndent(config, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(th.ConfigPath, configData, 0644)
	require.NoError(t, err)
}

// SetAsHome sets the temporary directory as HOME environment variable
func (th *TestHelper) SetAsHome(t *testing.T) func() {
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", th.TempDir)

	return func() {
		os.Setenv("HOME", originalHome)
	}
}

// GetTestConfig returns a standard test configuration
func GetTestConfig() MCConfig {
	return MCConfig{
		Version: "10",
		Aliases: map[string]AliasConfig{
			"local": {
				URL:       "http://localhost:9000",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin",
				API:       "s3v4",
				Path:      "auto",
				Insecure:  false,
			},
			"prod": {
				URL:       "https://minio-prod.example.com",
				AccessKey: "prod-access-key",
				SecretKey: "prod-secret-key",
				API:       "s3v4",
				Path:      "auto",
				Insecure:  false,
			},
			"staging": {
				URL:       "https://minio-staging.example.com",
				AccessKey: "staging-access-key",
				SecretKey: "staging-secret-key",
				API:       "s3v4",
				Path:      "auto",
				Insecure:  true,
			},
			"insecure-local": {
				URL:       "https://localhost:9000",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin",
				API:       "s3v4",
				Path:      "auto",
				Insecure:  true,
			},
		},
	}
}

func TestCreateMinIOClientConfiguration(t *testing.T) {
	helper := NewTestHelper(t)
	cleanup := helper.SetAsHome(t)
	defer cleanup()

	config := GetTestConfig()
	helper.CreateTestConfig(t, config)

	tests := []struct {
		name             string
		alias            string
		insecureFlag     bool
		expectedInsecure bool
		expectError      bool
		expectedEndpoint string
		expectedSecure   bool
	}{
		{
			name:             "local HTTP without insecure flag",
			alias:            "local",
			insecureFlag:     false,
			expectedInsecure: false,
			expectError:      false,
			expectedEndpoint: "localhost:9000",
			expectedSecure:   false,
		},
		{
			name:             "prod HTTPS without insecure flag",
			alias:            "prod",
			insecureFlag:     false,
			expectedInsecure: false,
			expectError:      false,
			expectedEndpoint: "minio-prod.example.com",
			expectedSecure:   true,
		},
		{
			name:             "staging HTTPS with config insecure",
			alias:            "staging",
			insecureFlag:     false,
			expectedInsecure: true,
			expectError:      false,
			expectedEndpoint: "minio-staging.example.com",
			expectedSecure:   true,
		},
		{
			name:             "prod with command line insecure override",
			alias:            "prod",
			insecureFlag:     true,
			expectedInsecure: true,
			expectError:      false,
			expectedEndpoint: "minio-prod.example.com",
			expectedSecure:   true,
		},
		{
			name:             "insecure local HTTPS",
			alias:            "insecure-local",
			insecureFlag:     false,
			expectedInsecure: true,
			expectError:      false,
			expectedEndpoint: "localhost:9000",
			expectedSecure:   true,
		},
		{
			name:         "non-existent alias",
			alias:        "nonexistent",
			insecureFlag: false,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global insecure flag
			insecure = tt.insecureFlag

			loadedConfig, err := loadMCConfig()
			require.NoError(t, err)

			if tt.expectError {
				_, err := createMinIOClient(loadedConfig, tt.alias)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not found in MC configuration")
				return
			}

			// Test that we can create the client (though it won't connect to real servers)
			client, err := createMinIOClient(loadedConfig, tt.alias)

			// We expect this to succeed in creating the client object
			// Even though the servers don't exist, the client creation should work
			if err != nil {
				// Only check for configuration errors, not connection errors
				assert.NotContains(t, err.Error(), "not found in MC configuration")
			} else {
				assert.NotNil(t, client)
			}

			// Verify the configuration was parsed correctly
			aliasConfig := loadedConfig.Aliases[tt.alias]
			assert.Equal(t, tt.expectedInsecure, insecure || aliasConfig.Insecure)
		})
	}
}

func TestIntegrationURLParsing(t *testing.T) {
	testCases := []struct {
		description  string
		url          string
		expectAlias  string
		expectBucket string
		expectPath   string
		expectError  bool
	}{
		{
			description:  "Production bucket",
			url:          "prod/critical-data",
			expectAlias:  "prod",
			expectBucket: "critical-data",
			expectPath:   "",
			expectError:  false,
		},
		{
			description:  "Staging with deep path",
			url:          "staging/backup/2024/01/15/data.tar.gz",
			expectAlias:  "staging",
			expectBucket: "backup",
			expectPath:   "2024/01/15/data.tar.gz",
			expectError:  false,
		},
		{
			description:  "Local development",
			url:          "local/test-bucket/uploads/images/photo.jpg",
			expectAlias:  "local",
			expectBucket: "test-bucket",
			expectPath:   "uploads/images/photo.jpg",
			expectError:  false,
		},
		{
			description:  "Edge case with many separators",
			url:          "alias/bucket/a/very/deep/path/structure/file.ext",
			expectAlias:  "alias",
			expectBucket: "bucket",
			expectPath:   "a/very/deep/path/structure/file.ext",
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			alias, bucket, path, err := parseURL(tc.url)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectAlias, alias)
				assert.Equal(t, tc.expectBucket, bucket)
				assert.Equal(t, tc.expectPath, path)
			}
		})
	}
}

func TestFullConfigurationWorkflow(t *testing.T) {
	helper := NewTestHelper(t)
	cleanup := helper.SetAsHome(t)
	defer cleanup()

	// Create a comprehensive test configuration
	config := MCConfig{
		Version: "10",
		Aliases: map[string]AliasConfig{
			"primary": {
				URL:       "https://primary.minio.example.com",
				AccessKey: "primary-access",
				SecretKey: "primary-secret",
				API:       "s3v4",
				Path:      "auto",
				Insecure:  false,
			},
			"replica": {
				URL:       "https://replica.minio.example.com",
				AccessKey: "replica-access",
				SecretKey: "replica-secret",
				API:       "s3v4",
				Path:      "auto",
				Insecure:  true,
			},
		},
	}

	helper.CreateTestConfig(t, config)

	// Test loading the configuration
	loadedConfig, err := loadMCConfig()
	require.NoError(t, err)

	assert.Equal(t, "10", loadedConfig.Version)
	assert.Len(t, loadedConfig.Aliases, 2)

	// Test primary alias
	primaryAlias := loadedConfig.Aliases["primary"]
	assert.Equal(t, "https://primary.minio.example.com", primaryAlias.URL)
	assert.Equal(t, "primary-access", primaryAlias.AccessKey)
	assert.Equal(t, "primary-secret", primaryAlias.SecretKey)
	assert.False(t, primaryAlias.Insecure)

	// Test replica alias
	replicaAlias := loadedConfig.Aliases["replica"]
	assert.Equal(t, "https://replica.minio.example.com", replicaAlias.URL)
	assert.Equal(t, "replica-access", replicaAlias.AccessKey)
	assert.Equal(t, "replica-secret", replicaAlias.SecretKey)
	assert.True(t, replicaAlias.Insecure)

	// Test URL parsing for typical use case
	sourceAlias, sourceBucket, sourcePath, err := parseURL("primary/production-data/current")
	assert.NoError(t, err)
	assert.Equal(t, "primary", sourceAlias)
	assert.Equal(t, "production-data", sourceBucket)
	assert.Equal(t, "current", sourcePath)

	targetAlias, targetBucket, targetPath, err := parseURL("replica/production-data/current")
	assert.NoError(t, err)
	assert.Equal(t, "replica", targetAlias)
	assert.Equal(t, "production-data", targetBucket)
	assert.Equal(t, "current", targetPath)

	// Verify the aliases exist in configuration
	_, exists := loadedConfig.Aliases[sourceAlias]
	assert.True(t, exists)

	_, exists = loadedConfig.Aliases[targetAlias]
	assert.True(t, exists)
}

func TestErrorHandlingIntegration(t *testing.T) {
	helper := NewTestHelper(t)
	cleanup := helper.SetAsHome(t)
	defer cleanup()

	// Test with missing config file
	_, err := loadMCConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read MC config file")

	// Create invalid JSON config
	invalidJSON := `{"version": "10", "aliases": {`
	err = os.WriteFile(helper.ConfigPath, []byte(invalidJSON), 0644)
	require.NoError(t, err)

	_, err = loadMCConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse MC config")

	// Test with valid config but missing alias
	config := GetTestConfig()
	helper.CreateTestConfig(t, config)

	loadedConfig, err := loadMCConfig()
	require.NoError(t, err)

	_, err = createMinIOClient(loadedConfig, "nonexistent-alias")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found in MC configuration")
}
