package compare

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

// ObjectInfo represents information about an object
type ObjectInfo struct {
	Key            string
	ETag           string
	Size           int64
	LastModified   time.Time
	VersionID      string
	IsLatest       bool
	IsDeleteMarker bool
	StorageClass   string
}

// ComparisonResult represents the result of comparing two objects
type ComparisonResult struct {
	Key         string
	Status      string // "identical", "different", "missing_source", "missing_target"
	SourceInfo  *ObjectInfo
	TargetInfo  *ObjectInfo
	Differences []string
}

// CompareObjects performs comparison between two MinIO buckets
func CompareObjects(sourceClient, targetClient *minio.Client, sourceBucket, sourcePath, targetBucket, targetPath string, versionsMode bool) ([]ComparisonResult, error) {
	ctx := context.Background()
	var results []ComparisonResult

	// Get objects from source (always gets all versions)
	allSourceObjects, err := ListObjects(ctx, sourceClient, sourceBucket, sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to list source objects: %v", err)
	}

	// Get objects from target (always gets all versions)
	allTargetObjects, err := ListObjects(ctx, targetClient, targetBucket, targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list target objects: %v", err)
	}

	// Filter objects based on comparison mode
	var sourceObjects, targetObjects []*ObjectInfo

	if versionsMode {
		// Include all versions when in versions mode
		sourceObjects = allSourceObjects
		targetObjects = allTargetObjects
	} else {
		// Filter to only current versions (non-delete markers with IsLatest = true)
		for _, obj := range allSourceObjects {
			if obj.IsLatest && !obj.IsDeleteMarker {
				sourceObjects = append(sourceObjects, obj)
			}
		}
		for _, obj := range allTargetObjects {
			if obj.IsLatest && !obj.IsDeleteMarker {
				targetObjects = append(targetObjects, obj)
			}
		}
	}

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

			// When not in versions mode, take the first object (current version)
			if len(sourceObjs) > 0 {
				sourceLatest = sourceObjs[0]
			}

			if len(targetObjs) > 0 {
				targetLatest = targetObjs[0]
			}

			result := compareCurrentVersions(key, sourceLatest, targetLatest)
			results = append(results, result)
		}
	}

	return results, nil
}

// ListObjects lists all objects in a bucket with the given prefix
func ListObjects(ctx context.Context, client *minio.Client, bucket, prefix string) ([]*ObjectInfo, error) {
	var objects []*ObjectInfo

	// Always use versioned listing for comprehensive detection
	opts := minio.ListObjectsOptions{
		Prefix:       prefix,
		Recursive:    true,
		WithVersions: true, // Always list all versions to detect hidden objects
	}

	for objInfo := range client.ListObjects(ctx, bucket, opts) {
		if objInfo.Err != nil {
			return nil, objInfo.Err
		}

		obj := &ObjectInfo{
			Key:            objInfo.Key,
			ETag:           objInfo.ETag,
			Size:           objInfo.Size,
			LastModified:   objInfo.LastModified,
			VersionID:      objInfo.VersionID,
			IsLatest:       objInfo.IsLatest,
			IsDeleteMarker: objInfo.IsDeleteMarker,
			StorageClass:   objInfo.StorageClass,
		}

		objects = append(objects, obj)
	}

	return objects, nil
}

func compareVersions(key string, sourceObjs, targetObjs []*ObjectInfo) []ComparisonResult {
	var results []ComparisonResult

	// Create version maps
	sourceVersions := make(map[string]*ObjectInfo)
	targetVersions := make(map[string]*ObjectInfo)

	for _, obj := range sourceObjs {
		sourceVersions[obj.VersionID] = obj
	}

	for _, obj := range targetObjs {
		targetVersions[obj.VersionID] = obj
	}

	// Get all version IDs
	allVersions := make(map[string]bool)
	for versionID := range sourceVersions {
		allVersions[versionID] = true
	}
	for versionID := range targetVersions {
		allVersions[versionID] = true
	}

	// Compare each version
	for versionID := range allVersions {
		sourceObj := sourceVersions[versionID]
		targetObj := targetVersions[versionID]

		var status string
		var differences []string

		if sourceObj == nil {
			status = "missing_source"
		} else if targetObj == nil {
			status = "missing_target"
		} else if sourceObj.ETag == targetObj.ETag && sourceObj.Size == targetObj.Size {
			status = "identical"
		} else {
			status = "different"
			if sourceObj.ETag != targetObj.ETag {
				differences = append(differences, "ETag differs")
			}
			if sourceObj.Size != targetObj.Size {
				differences = append(differences, "Size differs")
			}
		}

		result := ComparisonResult{
			Key:         fmt.Sprintf("%s (version: %s)", key, versionID),
			Status:      status,
			SourceInfo:  sourceObj,
			TargetInfo:  targetObj,
			Differences: differences,
		}

		results = append(results, result)
	}

	return results
}

func compareCurrentVersions(key string, sourceObj, targetObj *ObjectInfo) ComparisonResult {
	var status string
	var differences []string

	if sourceObj == nil {
		status = "missing_source"
	} else if targetObj == nil {
		status = "missing_target"
	} else if sourceObj.ETag == targetObj.ETag && sourceObj.Size == targetObj.Size {
		status = "identical"
	} else {
		status = "different"
		if sourceObj.ETag != targetObj.ETag {
			differences = append(differences, "ETag differs")
		}
		if sourceObj.Size != targetObj.Size {
			differences = append(differences, "Size differs")
		}
	}

	return ComparisonResult{
		Key:         key,
		Status:      status,
		SourceInfo:  sourceObj,
		TargetInfo:  targetObj,
		Differences: differences,
	}
}

// DisplayResults displays comparison results in a formatted way
func DisplayResults(results []ComparisonResult, verbose bool) {
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

	if different > 0 || missingSource > 0 || missingTarget > 0 {
		os.Exit(1)
	}
}