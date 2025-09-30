package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseURL(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		expectAlias  string
		expectBucket string
		expectPath   string
		expectError  bool
	}{
		{
			name:         "simple bucket",
			url:          "alias1/bucket1",
			expectAlias:  "alias1",
			expectBucket: "bucket1",
			expectPath:   "",
			expectError:  false,
		},
		{
			name:         "bucket with path",
			url:          "alias1/bucket1/folder/subfolder",
			expectAlias:  "alias1",
			expectBucket: "bucket1",
			expectPath:   "folder/subfolder",
			expectError:  false,
		},
		{
			name:        "invalid format - only alias",
			url:         "alias1",
			expectError: true,
		},
		{
			name:        "invalid format - empty",
			url:         "",
			expectError: true,
		},
		{
			name:         "bucket with single slash",
			url:          "prod/data",
			expectAlias:  "prod",
			expectBucket: "data",
			expectPath:   "",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alias, bucket, path, err := parseURL(tt.url)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectAlias, alias)
				assert.Equal(t, tt.expectBucket, bucket)
				assert.Equal(t, tt.expectPath, path)
			}
		})
	}
}

func TestLoadMCConfig(t *testing.T) {
	// Create a temporary directory for test config
	tempDir := t.TempDir()
	mcDir := filepath.Join(tempDir, ".mc")
	err := os.MkdirAll(mcDir, 0755)
	require.NoError(t, err)

	// Create test config
	testConfig := MCConfig{
		Version: "10",
		Aliases: map[string]AliasConfig{
			"local": {
				URL:       "http://localhost:9000",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin",
				API:       "s3v4",
				Path:      "auto",
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
		},
	}

	configPath := filepath.Join(mcDir, "config.json")
	configData, err := json.Marshal(testConfig)
	require.NoError(t, err)

	err = os.WriteFile(configPath, configData, 0644)
	require.NoError(t, err)

	// Temporarily change home directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Test loading config
	config, err := loadMCConfig()
	require.NoError(t, err)
	assert.Equal(t, "10", config.Version)
	assert.Len(t, config.Aliases, 3)

	// Test local alias
	localAlias := config.Aliases["local"]
	assert.Equal(t, "http://localhost:9000", localAlias.URL)
	assert.Equal(t, "minioadmin", localAlias.AccessKey)
	assert.False(t, localAlias.Insecure)

	// Test staging alias with insecure flag
	stagingAlias := config.Aliases["staging"]
	assert.Equal(t, "https://minio-staging.example.com", stagingAlias.URL)
	assert.True(t, stagingAlias.Insecure)
}

func TestLoadMCConfigError(t *testing.T) {
	// Set home to non-existent directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", "/non/existent/directory")
	defer os.Setenv("HOME", originalHome)

	_, err := loadMCConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read MC config file")
}

func TestCompareCurrentVersions(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		sourceObj      *ObjectInfo
		targetObj      *ObjectInfo
		expectedStatus string
		expectedDiffs  []string
	}{
		{
			name: "identical objects",
			key:  "test/file.txt",
			sourceObj: &ObjectInfo{
				Key:          "test/file.txt",
				ETag:         "abc123",
				Size:         1024,
				LastModified: time.Now(),
				VersionID:    "v1",
				IsLatest:     true,
			},
			targetObj: &ObjectInfo{
				Key:          "test/file.txt",
				ETag:         "abc123",
				Size:         1024,
				LastModified: time.Now(),
				VersionID:    "v1",
				IsLatest:     true,
			},
			expectedStatus: "identical",
			expectedDiffs:  []string{},
		},
		{
			name: "different etag",
			key:  "test/file.txt",
			sourceObj: &ObjectInfo{
				Key:          "test/file.txt",
				ETag:         "abc123",
				Size:         1024,
				LastModified: time.Now(),
				VersionID:    "v1",
				IsLatest:     true,
			},
			targetObj: &ObjectInfo{
				Key:          "test/file.txt",
				ETag:         "def456",
				Size:         1024,
				LastModified: time.Now(),
				VersionID:    "v1",
				IsLatest:     true,
			},
			expectedStatus: "different",
			expectedDiffs:  []string{"ETag differs"},
		},
		{
			name: "different size",
			key:  "test/file.txt",
			sourceObj: &ObjectInfo{
				Key:          "test/file.txt",
				ETag:         "abc123",
				Size:         1024,
				LastModified: time.Now(),
				VersionID:    "v1",
				IsLatest:     true,
			},
			targetObj: &ObjectInfo{
				Key:          "test/file.txt",
				ETag:         "abc123",
				Size:         2048,
				LastModified: time.Now(),
				VersionID:    "v1",
				IsLatest:     true,
			},
			expectedStatus: "different",
			expectedDiffs:  []string{"Size differs"},
		},
		{
			name: "different etag and size",
			key:  "test/file.txt",
			sourceObj: &ObjectInfo{
				Key:          "test/file.txt",
				ETag:         "abc123",
				Size:         1024,
				LastModified: time.Now(),
				VersionID:    "v1",
				IsLatest:     true,
			},
			targetObj: &ObjectInfo{
				Key:          "test/file.txt",
				ETag:         "def456",
				Size:         2048,
				LastModified: time.Now(),
				VersionID:    "v1",
				IsLatest:     true,
			},
			expectedStatus: "different",
			expectedDiffs:  []string{"ETag differs", "Size differs"},
		},
		{
			name:      "missing source",
			key:       "test/file.txt",
			sourceObj: nil,
			targetObj: &ObjectInfo{
				Key:          "test/file.txt",
				ETag:         "abc123",
				Size:         1024,
				LastModified: time.Now(),
				VersionID:    "v1",
				IsLatest:     true,
			},
			expectedStatus: "missing_source",
			expectedDiffs:  []string{},
		},
		{
			name: "missing target",
			key:  "test/file.txt",
			sourceObj: &ObjectInfo{
				Key:          "test/file.txt",
				ETag:         "abc123",
				Size:         1024,
				LastModified: time.Now(),
				VersionID:    "v1",
				IsLatest:     true,
			},
			targetObj:      nil,
			expectedStatus: "missing_target",
			expectedDiffs:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareCurrentVersions(tt.key, tt.sourceObj, tt.targetObj)

			assert.Equal(t, tt.key, result.Key)
			assert.Equal(t, tt.expectedStatus, result.Status)
			if len(tt.expectedDiffs) == 0 {
				assert.Empty(t, result.Differences)
			} else {
				assert.Equal(t, tt.expectedDiffs, result.Differences)
			}
			assert.Equal(t, tt.sourceObj, result.SourceInfo)
			assert.Equal(t, tt.targetObj, result.TargetInfo)
		})
	}
}

func TestCompareVersions(t *testing.T) {
	key := "test/file.txt"

	sourceObjs := []*ObjectInfo{
		{
			Key:          "test/file.txt",
			ETag:         "abc123",
			Size:         1024,
			LastModified: time.Now(),
			VersionID:    "v1",
			IsLatest:     false,
		},
		{
			Key:          "test/file.txt",
			ETag:         "def456",
			Size:         2048,
			LastModified: time.Now(),
			VersionID:    "v2",
			IsLatest:     true,
		},
	}

	targetObjs := []*ObjectInfo{
		{
			Key:          "test/file.txt",
			ETag:         "abc123",
			Size:         1024,
			LastModified: time.Now(),
			VersionID:    "v1",
			IsLatest:     false,
		},
		{
			Key:          "test/file.txt",
			ETag:         "xyz789",
			Size:         2048,
			LastModified: time.Now(),
			VersionID:    "v2",
			IsLatest:     true,
		},
		{
			Key:          "test/file.txt",
			ETag:         "new999",
			Size:         4096,
			LastModified: time.Now(),
			VersionID:    "v3",
			IsLatest:     false,
		},
	}

	results := compareVersions(key, sourceObjs, targetObjs)

	// Should have 3 results (v1, v2, v3)
	assert.Len(t, results, 3)

	// Find results by version
	resultMap := make(map[string]ComparisonResult)
	for _, result := range results {
		if result.Key == "test/file.txt (version: v1)" {
			resultMap["v1"] = result
		} else if result.Key == "test/file.txt (version: v2)" {
			resultMap["v2"] = result
		} else if result.Key == "test/file.txt (version: v3)" {
			resultMap["v3"] = result
		}
	}

	// v1 should be identical
	assert.Equal(t, "identical", resultMap["v1"].Status)

	// v2 should be different (different ETag)
	assert.Equal(t, "different", resultMap["v2"].Status)
	assert.Contains(t, resultMap["v2"].Differences, "ETag differs")

	// v3 should be missing in source
	assert.Equal(t, "missing_source", resultMap["v3"].Status)
}

func TestInsecureFlagLogic(t *testing.T) {
	tests := []struct {
		name             string
		aliasInsecure    bool
		cmdLineInsecure  bool
		expectedInsecure bool
	}{
		{
			name:             "both false",
			aliasInsecure:    false,
			cmdLineInsecure:  false,
			expectedInsecure: false,
		},
		{
			name:             "config true, cmd false",
			aliasInsecure:    true,
			cmdLineInsecure:  false,
			expectedInsecure: true,
		},
		{
			name:             "config false, cmd true",
			aliasInsecure:    false,
			cmdLineInsecure:  true,
			expectedInsecure: true,
		},
		{
			name:             "both true",
			aliasInsecure:    true,
			cmdLineInsecure:  true,
			expectedInsecure: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the logic from createMinIOClient
			insecure = tt.cmdLineInsecure
			aliasConfig := AliasConfig{
				Insecure: tt.aliasInsecure,
			}

			skipVerify := insecure || aliasConfig.Insecure
			assert.Equal(t, tt.expectedInsecure, skipVerify)
		})
	}
}

func TestObjectInfoStructure(t *testing.T) {
	obj := ObjectInfo{
		Key:          "test/file.txt",
		ETag:         "abc123",
		Size:         1024,
		LastModified: time.Now(),
		VersionID:    "v1",
		IsLatest:     true,
	}

	assert.Equal(t, "test/file.txt", obj.Key)
	assert.Equal(t, "abc123", obj.ETag)
	assert.Equal(t, int64(1024), obj.Size)
	assert.Equal(t, "v1", obj.VersionID)
	assert.True(t, obj.IsLatest)
}

func TestComparisonResultStructure(t *testing.T) {
	sourceObj := &ObjectInfo{
		Key:          "test/file.txt",
		ETag:         "abc123",
		Size:         1024,
		LastModified: time.Now(),
		VersionID:    "v1",
		IsLatest:     true,
	}

	result := ComparisonResult{
		Key:         "test/file.txt",
		Status:      "different",
		SourceInfo:  sourceObj,
		TargetInfo:  nil,
		Differences: []string{"ETag differs", "Size differs"},
	}

	assert.Equal(t, "test/file.txt", result.Key)
	assert.Equal(t, "different", result.Status)
	assert.Equal(t, sourceObj, result.SourceInfo)
	assert.Nil(t, result.TargetInfo)
	assert.Len(t, result.Differences, 2)
	assert.Contains(t, result.Differences, "ETag differs")
	assert.Contains(t, result.Differences, "Size differs")
}
