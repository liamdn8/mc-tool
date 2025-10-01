package analyze

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/liamdn8/mc-tool/pkg/compare"
)

func TestAnalyzeObjectDistribution(t *testing.T) {
	// Create test objects with various states
	objects := []*compare.ObjectInfo{
		{
			Key:            "file1.txt",
			Size:           100,
			IsLatest:       true,
			IsDeleteMarker: false,
		},
		{
			Key:            "file1.txt",
			Size:           80,
			IsLatest:       false,
			IsDeleteMarker: false,
		},
		{
			Key:            "file2.txt",
			Size:           200,
			IsLatest:       true,
			IsDeleteMarker: false,
		},
		{
			Key:            "file3.txt",
			Size:           0,
			IsLatest:       true,
			IsDeleteMarker: true,
		},
		{
			Key:            "file4.txt",
			Size:           150,
			IsLatest:       false,
			IsDeleteMarker: false,
		},
	}

	stats := AnalyzeObjectDistribution(objects)

	// Verify total counts
	assert.Equal(t, 5, stats["total_objects"])
	assert.Equal(t, 2, stats["current_versions"]) // file1 (latest), file2 (latest) - delete markers don't count as current versions
	assert.Equal(t, 2, stats["old_versions"])     // file1 (old), file4 (old)
	assert.Equal(t, 1, stats["delete_markers"])   // file3
	assert.Equal(t, 4, stats["unique_keys"])      // file1, file2, file3, file4

	// Verify sizes
	assert.Equal(t, int64(530), stats["total_size"])   // 100+80+200+0+150
	assert.Equal(t, int64(300), stats["current_size"]) // 100+200+0 (only current versions)

	// Verify version distribution
	versionDist := stats["version_distribution"].(map[string]int)
	assert.Equal(t, 2, versionDist["file1.txt"])
	assert.Equal(t, 1, versionDist["file2.txt"])
	assert.Equal(t, 1, versionDist["file3.txt"])
	assert.Equal(t, 1, versionDist["file4.txt"])
}

func TestAnalyzeObjectDistributionEmpty(t *testing.T) {
	// Test with empty object list
	objects := []*compare.ObjectInfo{}
	stats := AnalyzeObjectDistribution(objects)

	assert.Equal(t, 0, stats["total_objects"])
	assert.Equal(t, 0, stats["current_versions"])
	assert.Equal(t, 0, stats["old_versions"])
	assert.Equal(t, 0, stats["delete_markers"])
	assert.Equal(t, int64(0), stats["total_size"])
	assert.Equal(t, int64(0), stats["current_size"])
	assert.Equal(t, 0, stats["unique_keys"])

	versionDist := stats["version_distribution"].(map[string]int)
	assert.Len(t, versionDist, 0)
}

func TestAnalyzeObjectDistributionOnlyDeleteMarkers(t *testing.T) {
	// Test with only delete markers
	objects := []*compare.ObjectInfo{
		{
			Key:            "deleted1.txt",
			Size:           0,
			IsLatest:       true,
			IsDeleteMarker: true,
		},
		{
			Key:            "deleted2.txt",
			Size:           0,
			IsLatest:       true,
			IsDeleteMarker: true,
		},
	}

	stats := AnalyzeObjectDistribution(objects)

	assert.Equal(t, 2, stats["total_objects"])
	assert.Equal(t, 0, stats["current_versions"]) // Delete markers don't count as current versions
	assert.Equal(t, 0, stats["old_versions"])
	assert.Equal(t, 2, stats["delete_markers"])
	assert.Equal(t, int64(0), stats["total_size"])
	assert.Equal(t, int64(0), stats["current_size"])
	assert.Equal(t, 2, stats["unique_keys"])
}

func TestAnalyzeObjectDistributionVersioned(t *testing.T) {
	// Test with multiple versions of the same object
	objects := []*compare.ObjectInfo{
		{
			Key:            "versioned.txt",
			Size:           300,
			IsLatest:       true,
			IsDeleteMarker: false,
		},
		{
			Key:            "versioned.txt",
			Size:           250,
			IsLatest:       false,
			IsDeleteMarker: false,
		},
		{
			Key:            "versioned.txt",
			Size:           200,
			IsLatest:       false,
			IsDeleteMarker: false,
		},
	}

	stats := AnalyzeObjectDistribution(objects)

	assert.Equal(t, 3, stats["total_objects"])
	assert.Equal(t, 1, stats["current_versions"])
	assert.Equal(t, 2, stats["old_versions"])
	assert.Equal(t, 0, stats["delete_markers"])
	assert.Equal(t, int64(750), stats["total_size"])
	assert.Equal(t, int64(300), stats["current_size"])
	assert.Equal(t, 1, stats["unique_keys"])

	versionDist := stats["version_distribution"].(map[string]int)
	assert.Equal(t, 3, versionDist["versioned.txt"])
}