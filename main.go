package main

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/liamdn8/mc-tool/pkg/analyze"
	"github.com/liamdn8/mc-tool/pkg/client"
	"github.com/liamdn8/mc-tool/pkg/compare"
	"github.com/liamdn8/mc-tool/pkg/config"
	"github.com/liamdn8/mc-tool/pkg/validation"
)

var (
	// Build-time variables
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"

	// Runtime flags
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

	// Version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("mc-tool version %s\n", Version)
			fmt.Printf("Commit: %s\n", Commit)
			fmt.Printf("Built: %s\n", BuildTime)
		},
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
		Short: "Analyze MinIO bucket for object distribution",
		Long: `Analyze a MinIO bucket for object distribution, versions, and incomplete uploads.

Examples:
  mc-tool analyze alias/bucket
  mc-tool analyze --verbose alias/bucket/path
  mc-tool analyze alias/bucket/specific/path`,
		Args: cobra.ExactArgs(1),
		Run:  runAnalyze,
	}

	checklistCmd := &cobra.Command{
		Use:   "checklist <alias/bucket>",
		Short: "Check bucket configuration including event settings and lifecycle",
		Long: `Perform comprehensive validation of MinIO bucket configuration.

Checks include:
- Bucket existence
- Versioning configuration
- Event notifications (Lambda, Topic, Queue)
- Object lifecycle policies
- Server-side encryption
- Bucket policies and security settings

Examples:
  mc-tool checklist alias/bucket
  mc-tool checklist --verbose alias/bucket`,
		Args: cobra.ExactArgs(1),
		Run:  runChecklist,
	}

	// Configure flags
	compareCmd.Flags().BoolVar(&versionsMode, "versions", false, "Compare all object versions (default: compare current versions only)")
	compareCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	compareCmd.Flags().BoolVar(&insecure, "insecure", false, "Skip TLS certificate verification (overrides config setting)")

	analyzeCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	analyzeCmd.Flags().BoolVar(&insecure, "insecure", false, "Skip TLS certificate verification (overrides config setting)")

	checklistCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	checklistCmd.Flags().BoolVar(&insecure, "insecure", false, "Skip TLS certificate verification (overrides config setting)")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(compareCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(checklistCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runCompare(cmd *cobra.Command, args []string) {
	sourceURL := args[0]
	targetURL := args[1]

	// Parse source and target URLs
	sourceAlias, sourceBucket, sourcePath, err := client.ParseURL(sourceURL)
	if err != nil {
		log.Fatalf("Error parsing source URL: %v", err)
	}

	targetAlias, targetBucket, targetPath, err := client.ParseURL(targetURL)
	if err != nil {
		log.Fatalf("Error parsing target URL: %v", err)
	}

	// Load MC configuration
	cfg, err := config.LoadMCConfig()
	if err != nil {
		log.Fatalf("Error loading MC configuration: %v", err)
	}

	// Create MinIO clients
	sourceClient, err := client.CreateMinIOClient(cfg, sourceAlias, insecure, verbose)
	if err != nil {
		log.Fatalf("Error creating source client: %v", err)
	}

	targetClient, err := client.CreateMinIOClient(cfg, targetAlias, insecure, verbose)
	if err != nil {
		log.Fatalf("Error creating target client: %v", err)
	}

	// Perform comparison
	results, err := compare.CompareObjects(sourceClient, targetClient, sourceBucket, sourcePath, targetBucket, targetPath, versionsMode)
	if err != nil {
		log.Fatalf("Error comparing objects: %v", err)
	}

	// Display results
	compare.DisplayResults(results, verbose)
}

func runAnalyze(cmd *cobra.Command, args []string) {
	url := args[0]

	// Parse URL
	alias, bucket, path, err := client.ParseURL(url)
	if err != nil {
		log.Fatalf("Error parsing URL: %v", err)
	}

	// Load MinIO configuration
	cfg, err := config.LoadMCConfig()
	if err != nil {
		log.Fatalf("Error loading MC config: %v", err)
	}

	// Create MinIO client
	minioClient, err := client.CreateMinIOClient(cfg, alias, insecure, verbose)
	if err != nil {
		log.Fatalf("Error creating MinIO client: %v", err)
	}

	ctx := context.Background()

	// Get all objects (including all versions and delete markers)
	objects, err := compare.ListObjects(ctx, minioClient, bucket, path)
	if err != nil {
		log.Fatalf("Error listing objects: %v", err)
	}

	// Get incomplete multipart uploads
	incompleteUploads, err := analyze.ListIncompleteUploads(ctx, minioClient, bucket, path)
	if err != nil {
		log.Fatalf("Error listing incomplete uploads: %v", err)
	}

	// Analyze object distribution
	stats := analyze.AnalyzeObjectDistribution(objects)

	// Display analysis results
	analyze.DisplayAnalysisResults(stats, incompleteUploads, objects, verbose)
}

func runChecklist(cmd *cobra.Command, args []string) {
	url := args[0]

	// Parse URL (only need alias and bucket for checklist)
	alias, bucket, _, err := client.ParseURL(url)
	if err != nil {
		log.Fatalf("Error parsing URL: %v", err)
	}

	// Load MinIO configuration
	cfg, err := config.LoadMCConfig()
	if err != nil {
		log.Fatalf("Error loading MC config: %v", err)
	}

	// Create MinIO client
	minioClient, err := client.CreateMinIOClient(cfg, alias, insecure, verbose)
	if err != nil {
		log.Fatalf("Error creating MinIO client: %v", err)
	}

	ctx := context.Background()

	// Perform bucket configuration validation
	fmt.Printf("=== Bucket Configuration Checklist ===\n")
	err = validation.CheckBucketConfiguration(ctx, minioClient, bucket)
	if err != nil {
		log.Fatalf("Error checking bucket configuration: %v", err)
	}
}