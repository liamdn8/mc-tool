package compare

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCompareCurrentVersions(t *testing.T) {
	// Test identical objects
	obj1 := &ObjectInfo{
		Key:  "test.txt",
		ETag: "abc123",
		Size: 100,
	}
	obj2 := &ObjectInfo{
		Key:  "test.txt",
		ETag: "abc123",
		Size: 100,
	}

	result := compareCurrentVersions("test.txt", obj1, obj2)
	assert.Equal(t, "test.txt", result.Key)
	assert.Equal(t, "identical", result.Status)
	assert.Empty(t, result.Differences)

	// Test different objects
	obj3 := &ObjectInfo{
		Key:  "test.txt",
		ETag: "def456",
		Size: 200,
	}

	result = compareCurrentVersions("test.txt", obj1, obj3)
	assert.Equal(t, "test.txt", result.Key)
	assert.Equal(t, "different", result.Status)
	assert.Contains(t, result.Differences, "ETag differs")
	assert.Contains(t, result.Differences, "Size differs")

	// Test missing source
	result = compareCurrentVersions("test.txt", nil, obj2)
	assert.Equal(t, "test.txt", result.Key)
	assert.Equal(t, "missing_source", result.Status)

	// Test missing target
	result = compareCurrentVersions("test.txt", obj1, nil)
	assert.Equal(t, "test.txt", result.Key)
	assert.Equal(t, "missing_target", result.Status)
}

func TestCompareVersions(t *testing.T) {
	// Create test objects with different versions
	sourceObjs := []*ObjectInfo{
		{
			Key:       "test.txt",
			ETag:      "abc123",
			Size:      100,
			VersionID: "version1",
		},
		{
			Key:       "test.txt",
			ETag:      "def456",
			Size:      150,
			VersionID: "version2",
		},
	}

	targetObjs := []*ObjectInfo{
		{
			Key:       "test.txt",
			ETag:      "abc123",
			Size:      100,
			VersionID: "version1",
		},
		{
			Key:       "test.txt",
			ETag:      "ghi789",
			Size:      200,
			VersionID: "version3",
		},
	}

	results := compareVersions("test.txt", sourceObjs, targetObjs)

	// Should have 3 results: version1 (identical), version2 (missing in target), version3 (missing in source)
	assert.Len(t, results, 3)

	// Find each result type
	var identicalResult, missingTargetResult, missingSourceResult *ComparisonResult
	for i := range results {
		switch results[i].Status {
		case "identical":
			identicalResult = &results[i]
		case "missing_target":
			missingTargetResult = &results[i]
		case "missing_source":
			missingSourceResult = &results[i]
		}
	}

	// Verify identical result
	assert.NotNil(t, identicalResult)
	assert.Contains(t, identicalResult.Key, "version1")
	assert.Equal(t, "identical", identicalResult.Status)

	// Verify missing in target
	assert.NotNil(t, missingTargetResult)
	assert.Contains(t, missingTargetResult.Key, "version2")
	assert.Equal(t, "missing_target", missingTargetResult.Status)

	// Verify missing in source
	assert.NotNil(t, missingSourceResult)
	assert.Contains(t, missingSourceResult.Key, "version3")
	assert.Equal(t, "missing_source", missingSourceResult.Status)
}

func TestObjectInfoStructure(t *testing.T) {
	obj := ObjectInfo{
		Key:            "test/file.txt",
		ETag:           "abc123def456",
		Size:           1024,
		LastModified:   time.Now(),
		VersionID:      "version123",
		IsLatest:       true,
		IsDeleteMarker: false,
		StorageClass:   "STANDARD",
	}

	assert.Equal(t, "test/file.txt", obj.Key)
	assert.Equal(t, "abc123def456", obj.ETag)
	assert.Equal(t, int64(1024), obj.Size)
	assert.Equal(t, "version123", obj.VersionID)
	assert.True(t, obj.IsLatest)
	assert.False(t, obj.IsDeleteMarker)
	assert.Equal(t, "STANDARD", obj.StorageClass)
}

func TestComparisonResultStructure(t *testing.T) {
	sourceObj := &ObjectInfo{
		Key:  "test.txt",
		ETag: "abc123",
		Size: 100,
	}

	targetObj := &ObjectInfo{
		Key:  "test.txt",
		ETag: "def456",
		Size: 200,
	}

	result := ComparisonResult{
		Key:         "test.txt",
		Status:      "different",
		SourceInfo:  sourceObj,
		TargetInfo:  targetObj,
		Differences: []string{"ETag differs", "Size differs"},
	}

	assert.Equal(t, "test.txt", result.Key)
	assert.Equal(t, "different", result.Status)
	assert.Equal(t, sourceObj, result.SourceInfo)
	assert.Equal(t, targetObj, result.TargetInfo)
	assert.Len(t, result.Differences, 2)
	assert.Contains(t, result.Differences, "ETag differs")
	assert.Contains(t, result.Differences, "Size differs")
}