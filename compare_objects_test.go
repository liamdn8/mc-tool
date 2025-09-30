package main

import (
	"context"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockMinIOClient implements a mock MinIO client for testing
type MockMinIOClient struct {
	objects map[string][]minio.ObjectInfo
	err     error
}

// ListObjects mocks the MinIO ListObjects method
func (m *MockMinIOClient) ListObjects(ctx context.Context, bucket string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	ch := make(chan minio.ObjectInfo)
	
	go func() {
		defer close(ch)
		
		if m.err != nil {
			ch <- minio.ObjectInfo{Err: m.err}
			return
		}
		
		key := bucket + "/" + opts.Prefix
		if objects, exists := m.objects[key]; exists {
			for _, obj := range objects {
				// Filter by prefix if specified
				if opts.Prefix == "" || len(obj.Key) >= len(opts.Prefix) && obj.Key[:len(opts.Prefix)] == opts.Prefix {
					ch <- obj
				}
			}
		}
	}()
	
	return ch
}

// Helper function to create mock objects
func createMockObject(key, etag string, size int64, versionID string, isLatest bool) minio.ObjectInfo {
	return minio.ObjectInfo{
		Key:          key,
		ETag:         etag,
		Size:         size,
		LastModified: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		VersionID:    versionID,
		IsLatest:     isLatest,
	}
}

// Since compareObjects expects *minio.Client, we need to create a wrapper
// or modify our approach. Let's create a testable version of compareObjects
func compareObjectsTestable(
	sourceObjects, targetObjects []*ObjectInfo,
	versionsMode bool,
) []ComparisonResult {
	var results []ComparisonResult

	// Create maps for easy lookup
	sourceMap := make(map[string][]*ObjectInfo)
	targetMap := make(map[string][]*ObjectInfo)

	for _, obj := range sourceObjects {
		sourceMap[obj.Key] = append(sourceMap[obj.Key], obj)
	}

	for _, obj := range targetObjects {
		targetMap[obj.Key] = append(targetMap[obj.Key], obj)
	}

	// Get all unique keys
	allKeys := make(map[string]bool)
	for key := range sourceMap {
		allKeys[key] = true
	}
	for key := range targetMap {
		allKeys[key] = true
	}

	// Compare objects
	for key := range allKeys {
		sourceObjs := sourceMap[key]
		targetObjs := targetMap[key]

		if versionsMode {
			// Compare all versions
			result := compareVersions(key, sourceObjs, targetObjs)
			results = append(results, result...)
		} else {
			// Compare only current versions (latest)
			var sourceLatest, targetLatest *ObjectInfo

			for _, obj := range sourceObjs {
				if obj.IsLatest {
					sourceLatest = obj
					break
				}
			}

			for _, obj := range targetObjs {
				if obj.IsLatest {
					targetLatest = obj
					break
				}
			}

			result := compareCurrentVersions(key, sourceLatest, targetLatest)
			results = append(results, result)
		}
	}

	return results
}

func TestCompareObjectsTestable(t *testing.T) {
	tests := []struct {
		name          string
		sourceObjects []*ObjectInfo
		targetObjects []*ObjectInfo
		versionsMode  bool
		expectedCount int
		expectedTypes map[string]int // map of status -> count
	}{
		{
			name: "identical objects current version only",
			sourceObjects: []*ObjectInfo{
				{
					Key:       "file1.txt",
					ETag:      "abc123",
					Size:      1024,
					VersionID: "v1",
					IsLatest:  true,
				},
				{
					Key:       "file2.txt",
					ETag:      "def456",
					Size:      2048,
					VersionID: "v1",
					IsLatest:  true,
				},
			},
			targetObjects: []*ObjectInfo{
				{
					Key:       "file1.txt",
					ETag:      "abc123",
					Size:      1024,
					VersionID: "v1",
					IsLatest:  true,
				},
				{
					Key:       "file2.txt",
					ETag:      "def456",
					Size:      2048,
					VersionID: "v1",
					IsLatest:  true,
				},
			},
			versionsMode:  false,
			expectedCount: 2,
			expectedTypes: map[string]int{
				"identical": 2,
			},
		},
		{
			name: "different objects current version only",
			sourceObjects: []*ObjectInfo{
				{
					Key:       "file1.txt",
					ETag:      "abc123",
					Size:      1024,
					VersionID: "v1",
					IsLatest:  true,
				},
			},
			targetObjects: []*ObjectInfo{
				{
					Key:       "file1.txt",
					ETag:      "xyz789",
					Size:      2048,
					VersionID: "v1",
					IsLatest:  true,
				},
			},
			versionsMode:  false,
			expectedCount: 1,
			expectedTypes: map[string]int{
				"different": 1,
			},
		},
		{
			name: "missing objects",
			sourceObjects: []*ObjectInfo{
				{
					Key:       "file1.txt",
					ETag:      "abc123",
					Size:      1024,
					VersionID: "v1",
					IsLatest:  true,
				},
			},
			targetObjects: []*ObjectInfo{
				{
					Key:       "file2.txt",
					ETag:      "def456",
					Size:      2048,
					VersionID: "v1",
					IsLatest:  true,
				},
			},
			versionsMode:  false,
			expectedCount: 2,
			expectedTypes: map[string]int{
				"missing_target": 1,
				"missing_source": 1,
			},
		},
		{
			name: "multiple versions comparison",
			sourceObjects: []*ObjectInfo{
				{
					Key:       "file1.txt",
					ETag:      "abc123",
					Size:      1024,
					VersionID: "v1",
					IsLatest:  false,
				},
				{
					Key:       "file1.txt",
					ETag:      "def456",
					Size:      2048,
					VersionID: "v2",
					IsLatest:  true,
				},
			},
			targetObjects: []*ObjectInfo{
				{
					Key:       "file1.txt",
					ETag:      "abc123",
					Size:      1024,
					VersionID: "v1",
					IsLatest:  false,
				},
				{
					Key:       "file1.txt",
					ETag:      "xyz789",
					Size:      2048,
					VersionID: "v2",
					IsLatest:  true,
				},
			},
			versionsMode:  true,
			expectedCount: 2,
			expectedTypes: map[string]int{
				"identical": 1,
				"different": 1,
			},
		},
		{
			name: "mixed scenarios with versions",
			sourceObjects: []*ObjectInfo{
				{
					Key:       "file1.txt",
					ETag:      "abc123",
					Size:      1024,
					VersionID: "v1",
					IsLatest:  true,
				},
				{
					Key:       "file2.txt",
					ETag:      "def456",
					Size:      2048,
					VersionID: "v1",
					IsLatest:  false,
				},
				{
					Key:       "file2.txt",
					ETag:      "ghi789",
					Size:      3072,
					VersionID: "v2",
					IsLatest:  true,
				},
			},
			targetObjects: []*ObjectInfo{
				{
					Key:       "file1.txt",
					ETag:      "abc123",
					Size:      1024,
					VersionID: "v1",
					IsLatest:  true,
				},
				{
					Key:       "file2.txt",
					ETag:      "def456",
					Size:      2048,
					VersionID: "v1",
					IsLatest:  false,
				},
				{
					Key:       "file3.txt",
					ETag:      "jkl012",
					Size:      4096,
					VersionID: "v1",
					IsLatest:  true,
				},
			},
			versionsMode:  true,
			expectedCount: 4,
			expectedTypes: map[string]int{
				"identical":      2, // file1.txt v1, file2.txt v1
				"missing_source": 1, // file3.txt v1
				"missing_target": 1, // file2.txt v2
			},
		},
		{
			name:          "empty source and target",
			sourceObjects: []*ObjectInfo{},
			targetObjects: []*ObjectInfo{},
			versionsMode:  false,
			expectedCount: 0,
			expectedTypes: map[string]int{},
		},
		{
			name: "only latest versions when not in versions mode",
			sourceObjects: []*ObjectInfo{
				{
					Key:       "file1.txt",
					ETag:      "old123",
					Size:      512,
					VersionID: "v1",
					IsLatest:  false,
				},
				{
					Key:       "file1.txt",
					ETag:      "new456",
					Size:      1024,
					VersionID: "v2",
					IsLatest:  true,
				},
			},
			targetObjects: []*ObjectInfo{
				{
					Key:       "file1.txt",
					ETag:      "old123",
					Size:      512,
					VersionID: "v1",
					IsLatest:  false,
				},
				{
					Key:       "file1.txt",
					ETag:      "new456",
					Size:      1024,
					VersionID: "v2",
					IsLatest:  true,
				},
			},
			versionsMode:  false,
			expectedCount: 1,
			expectedTypes: map[string]int{
				"identical": 1, // Only v2 (latest) should be compared
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := compareObjectsTestable(tt.sourceObjects, tt.targetObjects, tt.versionsMode)
			
			assert.Equal(t, tt.expectedCount, len(results), "Expected %d results, got %d", tt.expectedCount, len(results))
			
			// Count results by status
			statusCounts := make(map[string]int)
			for _, result := range results {
				statusCounts[result.Status]++
			}
			
			for expectedStatus, expectedCount := range tt.expectedTypes {
				actualCount := statusCounts[expectedStatus]
				assert.Equal(t, expectedCount, actualCount, 
					"Expected %d results with status %s, got %d", expectedCount, expectedStatus, actualCount)
			}
			
			// Verify all results have proper keys
			for _, result := range results {
				assert.NotEmpty(t, result.Key, "Result key should not be empty")
				assert.Contains(t, []string{"identical", "different", "missing_source", "missing_target"}, 
					result.Status, "Result status should be valid")
			}
		})
	}
}

func TestCompareObjectsEdgeCases(t *testing.T) {
	t.Run("objects with same key but no latest version", func(t *testing.T) {
		sourceObjects := []*ObjectInfo{
			{
				Key:       "file1.txt",
				ETag:      "abc123",
				Size:      1024,
				VersionID: "v1",
				IsLatest:  false,
			},
		}
		
		targetObjects := []*ObjectInfo{
			{
				Key:       "file1.txt",
				ETag:      "def456",
				Size:      2048,
				VersionID: "v2",
				IsLatest:  false,
			},
		}
		
		results := compareObjectsTestable(sourceObjects, targetObjects, false)
		
		// When no latest version is found, both should be nil
		require.Len(t, results, 1)
		assert.Equal(t, "missing_source", results[0].Status)
	})
	
	t.Run("multiple objects with same key different versions", func(t *testing.T) {
		sourceObjects := []*ObjectInfo{
			{
				Key:       "file1.txt",
				ETag:      "v1-etag",
				Size:      1024,
				VersionID: "v1",
				IsLatest:  false,
			},
			{
				Key:       "file1.txt",
				ETag:      "v2-etag",
				Size:      2048,
				VersionID: "v2",
				IsLatest:  false,
			},
			{
				Key:       "file1.txt",
				ETag:      "v3-etag",
				Size:      3072,
				VersionID: "v3",
				IsLatest:  true,
			},
		}
		
		targetObjects := []*ObjectInfo{
			{
				Key:       "file1.txt",
				ETag:      "v1-etag",
				Size:      1024,
				VersionID: "v1",
				IsLatest:  false,
			},
			{
				Key:       "file1.txt",
				ETag:      "v3-etag",
				Size:      3072,
				VersionID: "v3",
				IsLatest:  true,
			},
		}
		
		// Test versions mode
		results := compareObjectsTestable(sourceObjects, targetObjects, true)
		assert.Len(t, results, 3) // v1, v2, v3
		
		statusCounts := make(map[string]int)
		for _, result := range results {
			statusCounts[result.Status]++
		}
		
		assert.Equal(t, 2, statusCounts["identical"]) // v1 and v3
		assert.Equal(t, 1, statusCounts["missing_target"]) // v2
		
		// Test non-versions mode (only latest)
		results = compareObjectsTestable(sourceObjects, targetObjects, false)
		assert.Len(t, results, 1) // Only latest version
		assert.Equal(t, "identical", results[0].Status)
	})
}

func TestCompareObjectsVersionMode(t *testing.T) {
	sourceObjects := []*ObjectInfo{
		{
			Key:       "file1.txt",
			ETag:      "abc123",
			Size:      1024,
			VersionID: "v1",
			IsLatest:  false,
		},
		{
			Key:       "file1.txt",
			ETag:      "def456",
			Size:      2048,
			VersionID: "v2",
			IsLatest:  true,
		},
	}
	
	targetObjects := []*ObjectInfo{
		{
			Key:       "file1.txt",
			ETag:      "abc123",
			Size:      1024,
			VersionID: "v1",
			IsLatest:  false,
		},
		{
			Key:       "file1.txt",
			ETag:      "xyz789", // Different ETag
			Size:      2048,
			VersionID: "v2",
			IsLatest:  true,
		},
		{
			Key:       "file1.txt",
			ETag:      "new999",
			Size:      4096,
			VersionID: "v3", // Extra version in target
			IsLatest:  false,
		},
	}
	
	// Test with versions mode enabled
	results := compareObjectsTestable(sourceObjects, targetObjects, true)
	
	assert.Len(t, results, 3) // v1, v2, v3
	
	// Check each version result
	resultMap := make(map[string]ComparisonResult)
	for _, result := range results {
		if result.Key == "file1.txt (version: v1)" {
			resultMap["v1"] = result
		} else if result.Key == "file1.txt (version: v2)" {
			resultMap["v2"] = result
		} else if result.Key == "file1.txt (version: v3)" {
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