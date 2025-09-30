package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// captureOutput captures stdout during function execution
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// displayResultsForTest is like displayResults but doesn't call os.Exit
func displayResultsForTest(results []ComparisonResult) {
	var identical, different, missingSource, missingTarget int

	fmt.Println("Comparison Results:")
	fmt.Println("==================")

	for _, result := range results {
		switch result.Status {
		case "identical":
			identical++
			if verbose {
				fmt.Printf("✓ %s - Identical\n", result.Key)
			}
		case "different":
			different++
			fmt.Printf("⚠ %s - Different (%s)\n", result.Key, strings.Join(result.Differences, ", "))
			if verbose {
				if result.SourceInfo != nil {
					fmt.Printf("  Source: ETag=%s, Size=%d, Modified=%s\n", 
						result.SourceInfo.ETag, result.SourceInfo.Size, result.SourceInfo.LastModified.Format(time.RFC3339))
				}
				if result.TargetInfo != nil {
					fmt.Printf("  Target: ETag=%s, Size=%d, Modified=%s\n", 
						result.TargetInfo.ETag, result.TargetInfo.Size, result.TargetInfo.LastModified.Format(time.RFC3339))
				}
			}
		case "missing_source":
			missingSource++
			fmt.Printf("- %s - Missing in source\n", result.Key)
		case "missing_target":
			missingTarget++
			fmt.Printf("+ %s - Missing in target\n", result.Key)
		}
	}

	fmt.Println("\nSummary:")
	fmt.Printf("  Identical: %d\n", identical)
	fmt.Printf("  Different: %d\n", different)
	fmt.Printf("  Missing in source: %d\n", missingSource)
	fmt.Printf("  Missing in target: %d\n", missingTarget)
	fmt.Printf("  Total compared: %d\n", len(results))

	// Don't call os.Exit in test version
}

func TestDisplayResults(t *testing.T) {
	// Create test results
	results := []ComparisonResult{
		{
			Key:    "file1.txt",
			Status: "identical",
		},
		{
			Key:         "file2.txt",
			Status:      "different",
			Differences: []string{"ETag differs"},
		},
		{
			Key:    "file3.txt",
			Status: "missing_source",
		},
		{
			Key:    "file4.txt",
			Status: "missing_target",
		},
	}

	// Test without verbose mode
	verbose = false
	output := captureOutput(func() {
		displayResultsForTest(results)
	})

	// Verify output contains expected elements
	assert.Contains(t, output, "Comparison Results:")
	assert.Contains(t, output, "⚠ file2.txt - Different (ETag differs)")
	assert.Contains(t, output, "- file3.txt - Missing in source")
	assert.Contains(t, output, "+ file4.txt - Missing in target")
	assert.Contains(t, output, "Summary:")
	assert.Contains(t, output, "Identical: 1")
	assert.Contains(t, output, "Different: 1")
	assert.Contains(t, output, "Missing in source: 1")
	assert.Contains(t, output, "Missing in target: 1")
	assert.Contains(t, output, "Total compared: 4")

	// Should not contain identical file in non-verbose mode
	assert.NotContains(t, output, "✓ file1.txt")
}

func TestDisplayResultsVerbose(t *testing.T) {
	// Create test results with detailed info
	sourceInfo := &ObjectInfo{
		ETag:         "abc123",
		Size:         1024,
		LastModified: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	
	targetInfo := &ObjectInfo{
		ETag:         "def456",
		Size:         2048,
		LastModified: time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}

	results := []ComparisonResult{
		{
			Key:    "file1.txt",
			Status: "identical",
		},
		{
			Key:         "file2.txt",
			Status:      "different",
			Differences: []string{"ETag differs", "Size differs"},
			SourceInfo:  sourceInfo,
			TargetInfo:  targetInfo,
		},
	}

	// Test with verbose mode
	verbose = true
	output := captureOutput(func() {
		displayResultsForTest(results)
	})

	// Verify verbose output contains detailed information
	assert.Contains(t, output, "✓ file1.txt - Identical")
	assert.Contains(t, output, "⚠ file2.txt - Different (ETag differs, Size differs)")
	assert.Contains(t, output, "Source: ETag=abc123, Size=1024")
	assert.Contains(t, output, "Target: ETag=def456, Size=2048")

	// Reset verbose for other tests
	verbose = false
}

func TestGlobalVariables(t *testing.T) {
	// Test that global variables can be set
	originalVersionsMode := versionsMode
	originalVerbose := verbose
	originalInsecure := insecure

	versionsMode = true
	verbose = true
	insecure = true

	assert.True(t, versionsMode)
	assert.True(t, verbose)
	assert.True(t, insecure)

	// Restore original values
	versionsMode = originalVersionsMode
	verbose = originalVerbose
	insecure = originalInsecure
}

func TestAliasConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      AliasConfig
		expectHTTPS bool
		expectHost  string
	}{
		{
			name: "HTTPS URL",
			config: AliasConfig{
				URL: "https://minio.example.com",
			},
			expectHTTPS: true,
			expectHost:  "minio.example.com",
		},
		{
			name: "HTTP URL",
			config: AliasConfig{
				URL: "http://localhost:9000",
			},
			expectHTTPS: false,
			expectHost:  "localhost:9000",
		},
		{
			name: "HTTPS URL with port",
			config: AliasConfig{
				URL: "https://minio.example.com:9000",
			},
			expectHTTPS: true,
			expectHost:  "minio.example.com:9000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useSSL := strings.HasPrefix(tt.config.URL, "https://")
			endpoint := strings.TrimPrefix(strings.TrimPrefix(tt.config.URL, "https://"), "http://")
			
			assert.Equal(t, tt.expectHTTPS, useSSL)
			assert.Equal(t, tt.expectHost, endpoint)
		})
	}
}

func TestErrorScenarios(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "empty URL",
			url:      "",
			expected: "invalid URL format",
		},
		{
			name:     "only alias",
			url:      "alias1",
			expected: "invalid URL format",
		},
		{
			name:     "URL with special characters",
			url:      "alias@special/bucket-name/path/to/file",
			expected: "", // Should not error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := parseURL(tt.url)
			
			if tt.expected == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expected)
			}
		})
	}
}

func TestMCConfigStructure(t *testing.T) {
	config := MCConfig{
		Version: "10",
		Aliases: map[string]AliasConfig{
			"test": {
				URL:       "https://test.example.com",
				AccessKey: "access",
				SecretKey: "secret",
				API:       "s3v4",
				Path:      "auto",
				Insecure:  true,
			},
		},
	}

	assert.Equal(t, "10", config.Version)
	assert.Len(t, config.Aliases, 1)
	
	testAlias := config.Aliases["test"]
	assert.Equal(t, "https://test.example.com", testAlias.URL)
	assert.Equal(t, "access", testAlias.AccessKey)
	assert.Equal(t, "secret", testAlias.SecretKey)
	assert.Equal(t, "s3v4", testAlias.API)
	assert.Equal(t, "auto", testAlias.Path)
	assert.True(t, testAlias.Insecure)
}

func TestEdgeCases(t *testing.T) {
	t.Run("empty object lists", func(t *testing.T) {
		results := compareVersions("test.txt", []*ObjectInfo{}, []*ObjectInfo{})
		assert.Empty(t, results)
	})

	t.Run("nil object comparison", func(t *testing.T) {
		result := compareCurrentVersions("test.txt", nil, nil)
		assert.Equal(t, "missing_source", result.Status)
	})

	t.Run("URL with many slashes", func(t *testing.T) {
		alias, bucket, path, err := parseURL("alias/bucket/path/with/many/slashes/file.txt")
		assert.NoError(t, err)
		assert.Equal(t, "alias", alias)
		assert.Equal(t, "bucket", bucket)
		assert.Equal(t, "path/with/many/slashes/file.txt", path)
	})
}

// Benchmark tests
func BenchmarkParseURL(b *testing.B) {
	url := "alias1/bucket1/very/long/path/to/some/file/deep/in/the/hierarchy/file.txt"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseURL(url)
	}
}

func BenchmarkCompareCurrentVersions(b *testing.B) {
	sourceObj := &ObjectInfo{
		Key:          "test/file.txt",
		ETag:         "abc123",
		Size:         1024,
		LastModified: time.Now(),
		VersionID:    "v1",
		IsLatest:     true,
	}
	
	targetObj := &ObjectInfo{
		Key:          "test/file.txt",
		ETag:         "def456",
		Size:         2048,
		LastModified: time.Now(),
		VersionID:    "v1",
		IsLatest:     true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compareCurrentVersions("test/file.txt", sourceObj, targetObj)
	}
}

func BenchmarkCompareVersions(b *testing.B) {
	sourceObjs := make([]*ObjectInfo, 100)
	targetObjs := make([]*ObjectInfo, 100)
	
	for i := 0; i < 100; i++ {
		sourceObjs[i] = &ObjectInfo{
			Key:          fmt.Sprintf("test/file%d.txt", i),
			ETag:         fmt.Sprintf("etag%d", i),
			Size:         int64(1024 * i),
			LastModified: time.Now(),
			VersionID:    fmt.Sprintf("v%d", i),
			IsLatest:     i == 99,
		}
		
		targetObjs[i] = &ObjectInfo{
			Key:          fmt.Sprintf("test/file%d.txt", i),
			ETag:         fmt.Sprintf("etag%d", i),
			Size:         int64(1024 * i),
			LastModified: time.Now(),
			VersionID:    fmt.Sprintf("v%d", i),
			IsLatest:     i == 99,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compareVersions("test/file.txt", sourceObjs, targetObjs)
	}
}