package analyze

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"

	"github.com/liamdn8/mc-tool/pkg/compare"
)

// ListIncompleteUploads detects incomplete multipart uploads that might affect object counts
func ListIncompleteUploads(ctx context.Context, client *minio.Client, bucket, prefix string) ([]minio.ObjectMultipartInfo, error) {
	var incompleteUploads []minio.ObjectMultipartInfo

	for uploadInfo := range client.ListIncompleteUploads(ctx, bucket, prefix, true) {
		if uploadInfo.Err != nil {
			return nil, uploadInfo.Err
		}
		incompleteUploads = append(incompleteUploads, uploadInfo)
	}

	return incompleteUploads, nil
}

// AnalyzeObjectDistribution provides detailed statistics about object versions and states
func AnalyzeObjectDistribution(objects []*compare.ObjectInfo) map[string]interface{} {
	stats := make(map[string]interface{})

	var totalObjects int
	var currentVersions int
	var oldVersions int
	var deleteMarkers int
	var totalSize int64
	var currentSize int64

	objectVersionCount := make(map[string]int)

	for _, obj := range objects {
		totalObjects++
		totalSize += obj.Size
		objectVersionCount[obj.Key]++

		if obj.IsDeleteMarker {
			deleteMarkers++
		} else if obj.IsLatest {
			currentVersions++
			currentSize += obj.Size
		} else {
			oldVersions++
		}
	}

	stats["total_objects"] = totalObjects
	stats["current_versions"] = currentVersions
	stats["old_versions"] = oldVersions
	stats["delete_markers"] = deleteMarkers
	stats["total_size"] = totalSize
	stats["current_size"] = currentSize
	stats["unique_keys"] = len(objectVersionCount)
	stats["version_distribution"] = objectVersionCount

	return stats
}

// DisplayAnalysisResults displays the analysis results in a formatted way
func DisplayAnalysisResults(stats map[string]interface{}, incompleteUploads []minio.ObjectMultipartInfo, objects []*compare.ObjectInfo, verbose bool) {
	fmt.Println("Object Distribution Analysis:")
	fmt.Println("============================")

	fmt.Printf("Total Objects (all versions): %d\n", stats["total_objects"])
	fmt.Printf("Current Versions: %d\n", stats["current_versions"])
	fmt.Printf("Old Versions: %d\n", stats["old_versions"])
	fmt.Printf("Delete Markers: %d\n", stats["delete_markers"])
	fmt.Printf("Unique Object Keys: %d\n", stats["unique_keys"])
	fmt.Printf("Total Size (all versions): %d bytes\n", stats["total_size"])
	fmt.Printf("Current Version Size: %d bytes\n", stats["current_size"])

	if len(incompleteUploads) > 0 {
		fmt.Printf("\nIncomplete Multipart Uploads: %d\n", len(incompleteUploads))
		if verbose {
			fmt.Println("\nIncomplete Upload Details:")
			for _, upload := range incompleteUploads {
				fmt.Printf("  - %s (ID: %s, Initiated: %s)\n",
					upload.Key, upload.UploadID, upload.Initiated.Format("2006-01-02 15:04:05"))
			}
		}
	} else {
		fmt.Println("\nIncomplete Multipart Uploads: 0")
	}

	if verbose && len(objects) > 0 {
		fmt.Println("\nDetailed Object Analysis:")
		fmt.Println("========================")

		// Group objects by key
		objectsByKey := make(map[string][]*compare.ObjectInfo)
		for _, obj := range objects {
			objectsByKey[obj.Key] = append(objectsByKey[obj.Key], obj)
		}

		for key, versions := range objectsByKey {
			fmt.Printf("\nObject: %s\n", key)
			fmt.Printf("  Total versions: %d\n", len(versions))

			for i, version := range versions {
				status := ""
				if version.IsLatest {
					status += "[CURRENT]"
				}
				if version.IsDeleteMarker {
					status += "[DELETE_MARKER]"
				}
				if status == "" {
					status = "[OLD_VERSION]"
				}

				fmt.Printf("  %d. %s Size: %d, ETag: %s, VersionID: %s, Modified: %s\n",
					i+1, status, version.Size, version.ETag, version.VersionID,
					version.LastModified.Format("2006-01-02 15:04:05"))
			}
		}
	}

	// Analysis summary
	fmt.Println("\nPotential Discrepancy Sources:")
	fmt.Println("==============================")

	if stats["delete_markers"].(int) > 0 {
		fmt.Printf("⚠ Found %d delete markers that might not be counted in some metrics\n", stats["delete_markers"])
	}

	if len(incompleteUploads) > 0 {
		fmt.Printf("⚠ Found %d incomplete multipart uploads that might affect object counts\n", len(incompleteUploads))
	}

	if stats["old_versions"].(int) > 0 {
		fmt.Printf("ℹ Found %d old versions (these should not affect current object counts)\n", stats["old_versions"])
	}

	currentObjects := stats["current_versions"].(int)
	totalVersions := stats["total_objects"].(int)

	fmt.Printf("\nMetrics Comparison:\n")
	fmt.Printf("- Current objects (should match bucket metrics): %d\n", currentObjects)
	fmt.Printf("- Total storage entries (all versions): %d\n", totalVersions)
	fmt.Printf("- Objects with delete markers as current version: %d\n", stats["delete_markers"])

	if stats["delete_markers"].(int) > 0 || len(incompleteUploads) > 0 {
		fmt.Println("\n🔍 Recommendation: These hidden objects might explain metric discrepancies")
	} else {
		fmt.Println("\n✅ No hidden objects detected - metric discrepancy might be due to other factors")
	}
}