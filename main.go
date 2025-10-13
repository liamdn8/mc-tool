package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/liamdn8/mc-tool/pkg/analyze"
	"github.com/liamdn8/mc-tool/pkg/client"
	"github.com/liamdn8/mc-tool/pkg/compare"
	"github.com/liamdn8/mc-tool/pkg/config"
	"github.com/liamdn8/mc-tool/pkg/logger"
	"github.com/liamdn8/mc-tool/pkg/profile"
	"github.com/liamdn8/mc-tool/pkg/validation"
	"github.com/liamdn8/mc-tool/pkg/web"
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
	webPort      int

	// Profile command flags
	profileType     string
	profileDuration string
	profileOutput   string
	profileMCPath   string
	detectLeaks     bool
	monitorInterval string
	thresholdMB     int
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

	debugCmd := &cobra.Command{
		Use:   "profile <type> <alias>",
		Short: "Profile MinIO server for performance analysis and memory leak detection",
		Long: `Profile MinIO server using mc admin profile command to analyze performance and detect memory leaks.

Profile Types:
- cpu: CPU profiling to identify performance bottlenecks
- heap: Memory heap profiling for memory leak detection  
- goroutine: Goroutine profiling to find goroutine leaks
- allocs: Allocation profiling for GC pressure analysis
- block: Blocking profiling for synchronization issues
- mutex: Mutex contention profiling

Features:
- Supports both latest mc and mc-2021 versions
- Continuous memory leak monitoring
- Automatic leak detection with configurable thresholds
- Detailed memory growth analysis
- Saves profiles to files for further analysis

Examples:
  # Basic heap profile for memory analysis
  mc-tool profile heap minio-prod

  # CPU profile with custom duration
  mc-tool profile cpu minio-prod --duration 30s

  # Memory leak detection with monitoring
  mc-tool profile heap minio-prod --detect-leaks --monitor-interval 30s --duration 10m

  # Save profile to file with older mc version
  mc-tool profile goroutine minio-prod --output /tmp/goroutine.pprof --mc-path mc-2021

  # Continuous leak monitoring with custom threshold
  mc-tool profile heap minio-prod --detect-leaks --threshold-mb 100 --duration 1h`,
		Args: cobra.ExactArgs(2),
		Run:  runProfile,
	}

	// Configure flags
	compareCmd.Flags().BoolVar(&versionsMode, "versions", false, "Compare all object versions (default: compare current versions only)")
	compareCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	compareCmd.Flags().BoolVar(&insecure, "insecure", false, "Skip TLS certificate verification (overrides config setting)")

	analyzeCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	analyzeCmd.Flags().BoolVar(&insecure, "insecure", false, "Skip TLS certificate verification (overrides config setting)")

	checklistCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	checklistCmd.Flags().BoolVar(&insecure, "insecure", false, "Skip TLS certificate verification (overrides config setting)")

	debugCmd.Flags().StringVar(&profileType, "type", "", "Profile type (auto-detected from command)")
	debugCmd.Flags().StringVar(&profileDuration, "duration", "30s", "Profile duration (e.g., 30s, 1m, 5m)")
	debugCmd.Flags().StringVar(&profileOutput, "output", "", "Output file path for profile data")
	debugCmd.Flags().StringVar(&profileMCPath, "mc-path", "mc", "Path to mc binary (mc, mc-2021, or custom path)")
	debugCmd.Flags().BoolVar(&detectLeaks, "detect-leaks", false, "Enable memory leak detection monitoring")
	debugCmd.Flags().StringVar(&monitorInterval, "monitor-interval", "10s", "Monitoring interval for leak detection")
	debugCmd.Flags().IntVar(&thresholdMB, "threshold-mb", 50, "Memory growth threshold in MB for leak detection")
	debugCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	debugCmd.Flags().BoolVar(&insecure, "insecure", false, "Skip TLS certificate verification")

	// Web UI command
	webCmd := &cobra.Command{
		Use:   "web",
		Short: "Start web UI server",
		Long: `Start a web-based user interface for mc-tool.

The web UI provides an easy-to-use interface for operators who are not familiar with MinIO CLI.

Features:
- Dashboard with MinIO aliases overview
- Visual bucket comparison tool
- Bucket analysis with charts
- Memory profiling and leak detection
- Bucket configuration checklist
- Bilingual support (English and Vietnamese)

Examples:
  # Start web UI on default port 8080
  mc-tool web

  # Start web UI on custom port
  mc-tool web --port 3000`,
		Run: runWeb,
	}

	webCmd.Flags().IntVar(&webPort, "port", 8080, "Web server port")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(compareCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(checklistCmd)
	rootCmd.AddCommand(debugCmd)
	rootCmd.AddCommand(webCmd)

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

func runProfile(cmd *cobra.Command, args []string) {
	profileTypeArg := args[0]
	alias := args[1]

	// Load MC configuration
	cfg, err := config.LoadMCConfig()
	if err != nil {
		log.Fatalf("Error loading MC config: %v", err)
	}

	// Verify alias exists
	_, exists := cfg.Aliases[alias]
	if !exists {
		log.Fatalf("Alias '%s' not found in MC config", alias)
	}

	// Parse duration
	duration, err := time.ParseDuration(profileDuration)
	if err != nil {
		log.Fatalf("Invalid duration: %v", err)
	}

	// Parse monitor interval for leak detection
	var monitorIntervalDuration time.Duration
	if detectLeaks {
		monitorIntervalDuration, err = time.ParseDuration(monitorInterval)
		if err != nil {
			log.Fatalf("Invalid monitor interval: %v", err)
		}
	}

	// Validate profile type
	validTypes := []string{"cpu", "heap", "goroutine", "allocs", "block", "mutex"}
	validType := false
	for _, t := range validTypes {
		if profileTypeArg == t {
			validType = true
			break
		}
	}
	if !validType {
		log.Fatalf("Invalid profile type: %s (valid types: %s)", profileTypeArg, strings.Join(validTypes, ", "))
	}

	// Check if mc binary exists
	if profileMCPath != "mc" && profileMCPath != "mc-2021" {
		// Custom path - check if it exists
		if _, err := os.Stat(profileMCPath); err != nil {
			log.Fatalf("MC binary not found at: %s", profileMCPath)
		}
	} else {
		// Standard path - check if it's available
		if _, err := exec.LookPath(profileMCPath); err != nil {
			// Try to find available versions
			versions := profile.GetAvailableMCVersions()
			if len(versions) == 0 {
				log.Fatalf("No mc binary found. Please install MinIO client or specify custom path with --mc-path")
			}
			fmt.Printf("Available MC versions: %s\n", strings.Join(versions, ", "))
			profileMCPath = versions[0] // Use first available
			fmt.Printf("Using: %s\n", profileMCPath)
		}
	}

	if verbose {
		fmt.Printf("🔧 Using alias: %s\n", alias)
		fmt.Printf("� Profile type: %s\n", profileTypeArg)
		fmt.Printf("⏱️  Duration: %s\n", duration)
		fmt.Printf("� MC Binary: %s\n", profileMCPath)
		if detectLeaks {
			fmt.Printf("🕵️  Leak detection: enabled\n")
			fmt.Printf("📈 Monitor interval: %s\n", monitorIntervalDuration)
			fmt.Printf("🚨 Threshold: %d MB\n", thresholdMB)
		}
		if profileOutput != "" {
			fmt.Printf("� Output: %s\n", profileOutput)
		}
		fmt.Println()
	}

	// Test mc admin profile command availability
	if err := profile.TestMCAdminProfile(profileMCPath, alias); err != nil {
		fmt.Printf("⚠️  Warning: Failed to test mc admin profile: %v\n", err)
		fmt.Printf("💡 Try using mc-2021 version: --mc-path mc-2021\n")
		fmt.Printf("💡 Or check if alias is properly configured: mc alias list\n")

		// Still proceed but with warning
	}

	// Create profile options
	opts := profile.ProfileOptions{
		Alias:           alias,
		ProfileType:     profileTypeArg,
		Duration:        duration,
		Output:          profileOutput,
		Verbose:         verbose,
		MCPath:          profileMCPath,
		DetectLeaks:     detectLeaks,
		MonitorInterval: monitorIntervalDuration,
		ThresholdMB:     thresholdMB,
	}

	// Run profiling
	if detectLeaks && profileTypeArg == "heap" {
		// Use memory leak monitoring for heap profiles
		err = profile.MonitorMemoryLeaks(opts)
	} else {
		// Standard profiling
		err = profile.RunProfile(opts)
	}

	if err != nil {
		log.Fatalf("Profile failed: %v", err)
	}
}

func runWeb(cmd *cobra.Command, args []string) {
	// Load configuration from environment variables
	cfg := config.LoadWebConfig()

	// Override port from CLI flag if provided
	if webPort != 8080 {
		cfg.Port = webPort
	}

	// Initialize logger
	logger.InitGlobalLogger(cfg.LogLevel, cfg.LogFormat)

	logger.GetLogger().Info("Starting MC-Tool Web UI", map[string]interface{}{
		"port":             cfg.Port,
		"refresh_interval": cfg.RefreshInterval,
		"log_level":        cfg.LogLevel,
	})

	fmt.Printf("🚀 Starting MC-Tool Web UI on port %d\n", cfg.Port)
	fmt.Printf("📱 Open your browser at: http://localhost:%d\n", cfg.Port)
	fmt.Printf("🌐 Supported languages: English, Tiếng Việt\n")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop the server")
	fmt.Println()

	server := web.NewServer(cfg)
	if err := server.Start(); err != nil {
		logger.GetLogger().Error("Failed to start web server", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("Failed to start web server: %v", err)
	}
}
