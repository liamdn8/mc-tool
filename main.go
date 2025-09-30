package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/cobra"
)

// MCConfig represents the MinIO client configuration
type MCConfig struct {
	Version string                 `json:"version"`
	Aliases map[string]AliasConfig `json:"aliases"`
}

// AliasConfig represents a single alias configuration
type AliasConfig struct {
	URL       string `json:"url"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	API       string `json:"api"`
	Path      string `json:"path"`
	Insecure  bool   `json:"insecure,omitempty"`
}

// ObjectInfo represents information about an object
type ObjectInfo struct {
	Key          string
	ETag         string
	Size         int64
	LastModified time.Time
	VersionID    string
	IsLatest     bool
	IsDeleteMarker bool
	StorageClass string
}

// ComparisonResult represents the result of comparing two objects
type ComparisonResult struct {
	Key         string
	Status      string // "identical", "different", "missing_source", "missing_target"
	SourceInfo  *ObjectInfo
	TargetInfo  *ObjectInfo
	Differences []string
}

var (
	versionsMode bool
	verbose      bool
	insecure     bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "mc-tool",
		Short: "MinIO client based support tool",
		Long:  "A tool for comparing MinIO buckets and objects across different instances",
	}

	compareCmd := &cobra.Command{
		Use:   "compare <source-alias/bucket/path> <target-alias/bucket/path>",
		Short: "Compare two MinIO buckets or paths",
		Long: `Compare objects between two MinIO buckets or paths.
		
Examples:
  mc-tool compare alias1/bucket1 alias2/bucket2
  mc-tool compare alias1/bucket1/folder alias2/bucket2/folder
  mc-tool compare --versions alias1/bucket1 alias2/bucket2
  mc-tool compare --insecure alias1/bucket1 alias2/bucket2`,
		Args: cobra.ExactArgs(2),
		Run:  runCompare,
	}

	analyzeCmd := &cobra.Command{
		Use:   "analyze <alias/bucket/path>",
		Short: "Analyze object distribution and detect hidden objects",
		Long: `Analyze object distribution in a MinIO bucket to detect:
- Current versions vs old versions
- Delete markers
- Incomplete multipart uploads
- Object count and size statistics
		
This command helps identify discrepancies between metrics and visible objects.

Examples:
  mc-tool analyze alias1/bucket1
  mc-tool analyze --verbose alias1/bucket1/folder`,
		Args: cobra.ExactArgs(1),
		Run:  runAnalyze,
	}

	compareCmd.Flags().BoolVar(&versionsMode, "versions", false, "Compare all object versions (default: compare current versions only)")
	compareCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	compareCmd.Flags().BoolVar(&insecure, "insecure", false, "Skip TLS certificate verification (overrides config setting)")

	analyzeCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	analyzeCmd.Flags().BoolVar(&insecure, "insecure", false, "Skip TLS certificate verification (overrides config setting)")

	rootCmd.AddCommand(compareCmd)
	rootCmd.AddCommand(analyzeCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runCompare(cmd *cobra.Command, args []string) {
	sourceURL := args[0]
	targetURL := args[1]

	// Parse source and target URLs
	sourceAlias, sourceBucket, sourcePath, err := parseURL(sourceURL)
	if err != nil {
		log.Fatalf("Error parsing source URL: %v", err)
	}

	targetAlias, targetBucket, targetPath, err := parseURL(targetURL)
	if err != nil {
		log.Fatalf("Error parsing target URL: %v", err)
	}

	// Load MC configuration
	config, err := loadMCConfig()
	if err != nil {
		log.Fatalf("Error loading MC configuration: %v", err)
	}

	// Create MinIO clients
	sourceClient, err := createMinIOClient(config, sourceAlias)
	if err != nil {
		log.Fatalf("Error creating source client: %v", err)
	}

	targetClient, err := createMinIOClient(config, targetAlias)
	if err != nil {
		log.Fatalf("Error creating target client: %v", err)
	}

	// Perform comparison
	results, err := compareObjects(sourceClient, targetClient, sourceBucket, sourcePath, targetBucket, targetPath)
	if err != nil {
		log.Fatalf("Error comparing objects: %v", err)
	}

	// Display results
	displayResults(results)
}

func runAnalyze(cmd *cobra.Command, args []string) {
	url := args[0]

	// Parse URL
	alias, bucket, path, err := parseURL(url)
	if err != nil {
		log.Fatalf("Error parsing URL: %v", err)
	}

	// Load MinIO configuration
	config, err := loadMCConfig()
	if err != nil {
		log.Fatalf("Error loading MC config: %v", err)
	}

	// Create MinIO client
	client, err := createMinIOClient(config, alias)
	if err != nil {
		log.Fatalf("Error creating MinIO client: %v", err)
	}

	ctx := context.Background()

	// Get all objects (including all versions and delete markers)
	objects, err := listObjects(ctx, client, bucket, path)
	if err != nil {
		log.Fatalf("Error listing objects: %v", err)
	}

	// Get incomplete multipart uploads
	incompleteUploads, err := listIncompleteUploads(ctx, client, bucket, path)
	if err != nil {
		log.Fatalf("Error listing incomplete uploads: %v", err)
	}

	// Analyze object distribution
	stats := analyzeObjectDistribution(objects)

	// Display analysis results
	displayAnalysisResults(stats, incompleteUploads, objects)
}

func parseURL(url string) (alias, bucket, path string, err error) {
	parts := strings.SplitN(url, "/", 3)
	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("invalid URL format: %s (expected alias/bucket[/path])", url)
	}

	alias = parts[0]
	bucket = parts[1]
	if len(parts) > 2 {
		path = parts[2]
	}

	return alias, bucket, path, nil
}

func loadMCConfig() (*MCConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	configPath := filepath.Join(homeDir, ".mc", "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read MC config file: %v", err)
	}

	var config MCConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse MC config: %v", err)
	}

	return &config, nil
}

func createMinIOClient(config *MCConfig, alias string) (*minio.Client, error) {
	aliasConfig, exists := config.Aliases[alias]
	if !exists {
		return nil, fmt.Errorf("alias '%s' not found in MC configuration", alias)
	}

	// Parse URL to determine if HTTPS is used
	useSSL := strings.HasPrefix(aliasConfig.URL, "https://")
	endpoint := strings.TrimPrefix(strings.TrimPrefix(aliasConfig.URL, "https://"), "http://")

	// Determine if we should skip certificate verification
	// Priority: command line flag > config setting > default (false)
	skipVerify := insecure || aliasConfig.Insecure

	// Create credentials
	creds := credentials.NewStaticV4(aliasConfig.AccessKey, aliasConfig.SecretKey, "")

	// Create transport with TLS configuration
	transport := &http.Transport{}
	if useSSL {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: skipVerify,
		}
	}

	// Create MinIO client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:     creds,
		Secure:    useSSL,
		Transport: transport,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %v", err)
	}

	if verbose {
		fmt.Printf("Connected to %s (SSL: %v, Skip Verify: %v)\n", aliasConfig.URL, useSSL, skipVerify)
	}

	return client, nil
}

func compareObjects(sourceClient, targetClient *minio.Client, sourceBucket, sourcePath, targetBucket, targetPath string) ([]ComparisonResult, error) {
	ctx := context.Background()
	var results []ComparisonResult

	// Get objects from source (always gets all versions)
	allSourceObjects, err := listObjects(ctx, sourceClient, sourceBucket, sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to list source objects: %v", err)
	}

	// Get objects from target (always gets all versions)
	allTargetObjects, err := listObjects(ctx, targetClient, targetBucket, targetPath)
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

func listObjects(ctx context.Context, client *minio.Client, bucket, prefix string) ([]*ObjectInfo, error) {
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

// listIncompleteUploads detects incomplete multipart uploads that might affect object counts
func listIncompleteUploads(ctx context.Context, client *minio.Client, bucket, prefix string) ([]minio.ObjectMultipartInfo, error) {
	var incompleteUploads []minio.ObjectMultipartInfo

	for uploadInfo := range client.ListIncompleteUploads(ctx, bucket, prefix, true) {
		if uploadInfo.Err != nil {
			return nil, uploadInfo.Err
		}
		incompleteUploads = append(incompleteUploads, uploadInfo)
	}

	return incompleteUploads, nil
}

// analyzeObjectDistribution provides detailed statistics about object versions and states
func analyzeObjectDistribution(objects []*ObjectInfo) map[string]interface{} {
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

func displayResults(results []ComparisonResult) {
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

func displayAnalysisResults(stats map[string]interface{}, incompleteUploads []minio.ObjectMultipartInfo, objects []*ObjectInfo) {
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
		objectsByKey := make(map[string][]*ObjectInfo)
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
