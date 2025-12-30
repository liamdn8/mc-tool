package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/liamdn8/mc-tool/pkg/analyze"
	"github.com/liamdn8/mc-tool/pkg/client"
	"github.com/liamdn8/mc-tool/pkg/compare"
	"github.com/liamdn8/mc-tool/pkg/config"
	"github.com/liamdn8/mc-tool/pkg/logger"
	"github.com/liamdn8/mc-tool/pkg/perftest"
	"github.com/liamdn8/mc-tool/pkg/profile"
	"github.com/liamdn8/mc-tool/pkg/trace"
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

	// Trace command flags
	traceDuration        string
	traceMCPath          string
	traceStatusCodes     []int
	traceErrorFilters    []string
	traceGroupByAPI      bool
	traceGroupByClient   bool
	traceGroupByVersions bool

	// Validate config command flags
	validateLifecycleOnly bool
	validateEventsOnly    bool

	// Perftest command flags
	perftestSize       string
	perftestBucket     string
	perftestPath       string
	perftestCount      int
	perftestOverride   int
	perftestMode       string
	perftestInterval   string
	perftestIterations int
	perftestParallel   int
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

	validateCmd := &cobra.Command{
		Use:   "validate <bucket-name> <alias1> <alias2> [alias3...]",
		Short: "Validate bucket configuration consistency across multiple aliases",
		Long: `Compare bucket lifecycle and event notification configurations across multiple MinIO aliases.

This command helps ensure consistent configuration across replicated or multi-site MinIO deployments.
The first alias is used as the reference, and all other aliases are compared against it.

Configurations checked:
- Bucket lifecycle (ILM) policies
- Event notifications

Examples:
  # Validate lifecycle and events for test-bucket across 3 sites
  mc-tool validate test-bucket site1 site2 site3

  # Validate only lifecycle configuration
  mc-tool validate test-bucket site1 site2 --lifecycle-only

  # Validate only event notifications
  mc-tool validate test-bucket site1 site2 site3 --events-only

  # Verbose output with detailed configuration display
  mc-tool validate test-bucket site1 site2 site3 --verbose

  # Skip TLS verification for self-signed certificates
  mc-tool validate test-bucket site1 site2 site3 --insecure`,
		Args: cobra.MinimumNArgs(2),
		Run:  runValidate,
	}

	debugCmd := &cobra.Command{
		Use:   "profile <alias>",
		Short: "Profile MinIO server using mc admin profile start/stop",
		Long: `Profile MinIO server by running 'mc admin profile start', waiting, then 'mc admin profile stop'.
Automatically extracts profile.zip and provides go tool pprof commands.

Profile Types (can combine multiple with comma):
- cpu: CPU profiling to identify performance bottlenecks
- mem: Memory heap profiling for memory leak detection  
- goroutines: Goroutine profiling to find goroutine leaks
- block: Blocking profiling for synchronization issues
- mutex: Mutex contention profiling
- trace: Execution trace
- threads: Thread profiling

Features:
- Automatically runs profile start, waits specified duration, then profile stop
- Downloads and extracts profile.zip to timestamped directory
- Displays ready-to-use 'go tool pprof' commands
- Supports --insecure for self-signed certificates
- Works with mc21 (MinIO Client 2021 version)

Examples:
  # Profile with default types (cpu,mem,block,goroutines) for 30 seconds
  mc-tool profile minio-prod

  # CPU and memory profile for 60 seconds
  mc-tool profile minio-prod --type cpu,mem --duration 60s

  # Profile with custom output directory
  mc-tool profile minio-prod --output /tmp/profiles

  # Profile with insecure mode for self-signed certs
  mc-tool profile minio-prod --insecure

  # Use specific mc binary path
  mc-tool profile minio-prod --mc-path /usr/local/bin/mc21`,
		Args: cobra.ExactArgs(1),
		Run:  runProfile,
	}

	traceCmd := &cobra.Command{
		Use:   "trace <alias>",
		Short: "Capture mc admin trace errors and summarize repeated object failures",
		Long: `Capture mc admin trace error output for a fixed duration and aggregate repeated object failures.

Examples:
  mc-tool trace alias
  mc-tool trace alias --duration 10s
  mc-tool trace alias --duration 2m --mc-path /usr/local/bin/mc`,
		Args: cobra.ExactArgs(1),
		Run:  runTrace,
	}

	// Configure flags
	compareCmd.Flags().BoolVar(&versionsMode, "versions", false, "Compare all object versions (default: compare current versions only)")
	compareCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	compareCmd.Flags().BoolVar(&insecure, "insecure", false, "Skip TLS certificate verification (overrides config setting)")

	analyzeCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	analyzeCmd.Flags().BoolVar(&insecure, "insecure", false, "Skip TLS certificate verification (overrides config setting)")

	checklistCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	checklistCmd.Flags().BoolVar(&insecure, "insecure", false, "Skip TLS certificate verification (overrides config setting)")

	validateCmd.Flags().BoolVar(&validateLifecycleOnly, "lifecycle-only", false, "Only validate lifecycle configuration")
	validateCmd.Flags().BoolVar(&validateEventsOnly, "events-only", false, "Only validate event notifications")
	validateCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output with detailed configuration")
	validateCmd.Flags().BoolVar(&insecure, "insecure", false, "Skip TLS certificate verification for self-signed certificates")

	debugCmd.Flags().StringVar(&profileType, "type", "cpu,mem,block,goroutines", "Profile types to collect (comma-separated): cpu,mem,block,mutex,trace,threads,goroutines")
	debugCmd.Flags().StringVar(&profileDuration, "duration", "30s", "Profile duration (e.g., 30s, 1m, 5m)")
	debugCmd.Flags().StringVar(&profileOutput, "output", "", "Output directory for extracted profile data (default: /tmp/profile-<timestamp>)")
	debugCmd.Flags().StringVar(&profileMCPath, "mc-path", "mc21", "Path to mc binary (mc21, mc-2021, or custom path)")
	debugCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	debugCmd.Flags().BoolVar(&insecure, "insecure", false, "Skip TLS certificate verification (pass --insecure to mc)")

	traceCmd.Flags().StringVar(&traceDuration, "duration", "5s", "Trace capture duration between 1s and 5m")
	traceCmd.Flags().StringVar(&traceMCPath, "mc-path", "mc", "Path to mc binary (mc, mc-2021, or custom path)")
	traceCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	traceCmd.Flags().BoolVar(&insecure, "insecure", false, "Pass --insecure to mc admin trace")
	traceCmd.Flags().IntSliceVar(&traceStatusCodes, "status", []int{}, "Only include events with matching HTTP status code (repeatable)")
	traceCmd.Flags().StringSliceVar(&traceErrorFilters, "error-contains", []string{}, "Only include events whose error message contains the provided substring (repeatable)")
	traceCmd.Flags().BoolVar(&traceGroupByAPI, "group-by-api", false, "Include API-level aggregation in the summary output")
	traceCmd.Flags().BoolVar(&traceGroupByClient, "group-by-client", false, "Include client-level aggregation in the summary output")
	traceCmd.Flags().BoolVar(&traceGroupByVersions, "versions", false, "Group by object and version instead of object only")

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

	// Perftest command
	perftestCmd := &cobra.Command{
		Use:   "perftest <site-alias> [flags]",
		Short: "Run performance test for MinIO PUT operations",
		Long: `Execute automated performance tests for MinIO PUT operations.

This tool measures upload performance with different patterns and configurations.

Upload Modes:
  all      - Upload all files at once with parallelism (fastest)
  interval - Upload in rounds with time intervals (rate-limited testing)

Interval Mode:
  --count: number of objects per round
  --iterations: number of rounds to upload
  --interval: time to wait between rounds
  Example: --count 10 --iterations 5 --interval 5s
    => Upload 5 rounds, each with 10 objects, wait 5s between rounds
    => Total: 50 objects

Override Feature:
  Use --override N to upload each object N times (tests versioning/overwrite scenarios)
  Example: --count 10 --override 3 will upload 10 objects, each 3 times = 30 total uploads

Examples:
  # Upload 100 small objects at once with 10 parallel workers
  mc-tool perftest site1 --bucket test-bucket --count 100 --size small --mode all --parallel 10

  # Upload in intervals: 10 objects per round, 5 rounds, wait 5s between rounds
  mc-tool perftest site1 --bucket test-bucket --count 10 --iterations 5 --mode interval --interval 5s

  # Test overwriting: upload 10 objects, override each 5 times
  mc-tool perftest site1 --bucket test-bucket --count 10 --override 5 --size large

  # Custom path
  mc-tool perftest site1 --bucket test-bucket --path testdata/ --count 200

Object Size Presets:
  small  - Random 1-10 KiB
  medium - Random 100-500 KiB
  large  - Random 1-5 MiB`,
		Args: cobra.ExactArgs(1),
		Run:  runPerftest,
	}

	perftestCmd.Flags().StringVar(&perftestSize, "size", "small", "Object size preset: small, medium, large")
	perftestCmd.Flags().StringVar(&perftestBucket, "bucket", "", "Bucket name for testing (required)")
	perftestCmd.Flags().StringVar(&perftestPath, "path", "", "Object path prefix (auto-generated with timestamp if not specified)")
	perftestCmd.Flags().IntVar(&perftestCount, "count", 10, "Number of objects per upload (per round for interval mode)")
	perftestCmd.Flags().IntVar(&perftestOverride, "override", 0, "Number of times to override each object (0 = no override)")
	perftestCmd.Flags().StringVar(&perftestMode, "mode", "all", "Upload mode: all or interval")
	perftestCmd.Flags().StringVar(&perftestInterval, "interval", "5s", "Time between upload rounds (interval mode only)")
	perftestCmd.Flags().IntVar(&perftestIterations, "iterations", 5, "Number of upload rounds (interval mode only)")
	perftestCmd.Flags().IntVar(&perftestParallel, "parallel", 5, "Number of parallel workers for 'all' mode")
	perftestCmd.MarkFlagRequired("bucket")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(compareCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(checklistCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(debugCmd)
	rootCmd.AddCommand(traceCmd)
	rootCmd.AddCommand(perftestCmd)
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

func runValidate(cmd *cobra.Command, args []string) {
	bucket := args[0]
	aliases := args[1:]

	if len(aliases) < 2 {
		log.Fatal("At least 2 aliases required for comparison")
	}

	// Determine what to check
	checkLifecycle := !validateEventsOnly
	checkEvents := !validateLifecycleOnly

	// Load MinIO configuration
	cfg, err := config.LoadMCConfig()
	if err != nil {
		log.Fatalf("Error loading MC config: %v", err)
	}

	// Validate all aliases exist
	for _, alias := range aliases {
		if _, ok := cfg.Aliases[alias]; !ok {
			log.Fatalf("Alias '%s' not found in MC config", alias)
		}
	}

	validator := validation.NewBucketValidator(bucket, aliases, verbose, insecure)

	fmt.Printf("╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  Bucket Configuration Validation                            ║\n")
	fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n\n")
	fmt.Printf("🪣 Bucket: %s\n", bucket)
	fmt.Printf("📍 Reference: %s\n", validator.ReferenceAlias)
	fmt.Printf("🔍 Comparing: %s\n\n", strings.Join(aliases, ", "))

	// Check bucket existence on all aliases
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📦 Bucket Existence Check")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for _, alias := range aliases {
		exists := validator.CheckBucketExists(alias)
		status := "✅ Found"
		if !exists {
			status = "❌ Missing"
		}
		fmt.Printf("  %-20s %s\n", alias+":", status)
	}
	fmt.Println()

	// Validate lifecycle if requested
	if checkLifecycle {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("♻️  Lifecycle Configuration")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		results, err := validator.ValidateLifecycle()
		if err != nil {
			log.Fatalf("Error validating lifecycle: %v", err)
		}

		printValidationResults(results)
		fmt.Println()
	}

	// Validate events if requested
	if checkEvents {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("📢 Event Notifications")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		results, err := validator.ValidateEvents()
		if err != nil {
			log.Fatalf("Error validating events: %v", err)
		}

		printValidationResults(results)
		fmt.Println()
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ Validation complete")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func printValidationResults(results []validation.ValidationResult) {
	matchCount := 0
	mismatchCount := 0

	for _, result := range results {
		if result.Status == "reference" {
			// Print reference
			if result.Configured {
				count := result.RuleCount
				if count == 0 {
					count = result.EventCount
				}
				if count > 0 {
					fmt.Printf("  📍 %s (Reference): ✅ Configured (%d rule(s))\n", result.Alias, count)
				} else {
					fmt.Printf("  📍 %s (Reference): ✅ Configured\n", result.Alias)
				}
				if verbose {
					fmt.Printf("     %s\n", strings.TrimSpace(result.ConfigRaw))
				}
			} else {
				fmt.Printf("  📍 %s (Reference): ⚠️  Not configured\n", result.Alias)
			}
		} else {
			// Print comparison results
			switch result.Status {
			case "match":
				count := result.RuleCount
				if count == 0 {
					count = result.EventCount
				}
				if count > 0 {
					fmt.Printf("  %-20s ✅ Match (%d rule(s))\n", result.Alias+":", count)
				} else {
					fmt.Printf("  %-20s ✅ Match (not configured)\n", result.Alias+":")
				}
				if verbose && result.Configured {
					fmt.Printf("     %s\n", strings.TrimSpace(result.ConfigRaw))
				}
				matchCount++
			case "mismatch":
				count := result.RuleCount
				if count == 0 {
					count = result.EventCount
				}
				if result.Configured {
					fmt.Printf("  %-20s ⚠️  Mismatch (%d rule(s), differs from reference)\n", result.Alias+":", count)
				} else {
					fmt.Printf("  %-20s ⚠️  Mismatch (not configured)\n", result.Alias+":")
				}
				if verbose {
					fmt.Printf("     %s\n", strings.TrimSpace(result.ConfigRaw))
				}
				mismatchCount++
			case "error":
				fmt.Printf("  %-20s ❌ Error: %v\n", result.Alias+":", result.Error)
				mismatchCount++
			}
		}
	}

	// Summary
	total := len(results) - 1 // Exclude reference
	if matchCount == total {
		fmt.Printf("\n  Summary: ✅ All %d aliases match reference\n", total)
	} else if matchCount > 0 {
		fmt.Printf("\n  Summary: ⚠️  %d/%d match, %d mismatch\n", matchCount, total, mismatchCount)
	} else {
		fmt.Printf("\n  Summary: ❌ All %d aliases differ from reference\n", total)
	}
}

func runTrace(cmd *cobra.Command, args []string) {
	alias := args[0]

	duration, err := time.ParseDuration(traceDuration)
	if err != nil {
		log.Fatalf("Invalid duration: %v", err)
	}

	if duration < time.Second || duration > 5*time.Minute {
		log.Fatalf("Duration must be between 1s and 5m")
	}

	cfg, err := config.LoadMCConfig()
	if err != nil {
		log.Fatalf("Error loading MC config: %v", err)
	}

	if _, ok := cfg.Aliases[alias]; !ok {
		log.Fatalf("Alias '%s' not found in MC config", alias)
	}

	mcPath := resolveMCBinary(traceMCPath, verbose)

	if verbose {
		fmt.Printf("📡 Capturing trace for alias: %s\n", alias)
		fmt.Printf("⏱️  Duration: %s\n", duration)
		fmt.Printf("🔧 mc binary: %s\n", mcPath)
		if insecure {
			fmt.Println("⚠️  Passing --insecure to mc admin trace")
		}
	}

	result, err := trace.Run(trace.Options{
		Alias:               alias,
		MCPath:              mcPath,
		Duration:            duration,
		Insecure:            insecure,
		Verbose:             verbose,
		StatusCodeFilters:   traceStatusCodes,
		ErrorMessageFilters: traceErrorFilters,
		GroupByAPI:          traceGroupByAPI,
		GroupByClient:       traceGroupByClient,
		GroupByVersions:     traceGroupByVersions,
	})
	if err != nil {
		log.Fatalf("Trace capture failed: %v", err)
	}

	displayTraceResult(result)
}

func resolveMCBinary(desired string, verboseMode bool) string {
	if desired != "mc" && desired != "mc-2021" {
		if _, err := os.Stat(desired); err != nil {
			log.Fatalf("MC binary not found at: %s", desired)
		}
		return desired
	}

	if _, err := exec.LookPath(desired); err == nil {
		return desired
	}

	versions := profile.GetAvailableMCVersions()
	if len(versions) == 0 {
		log.Fatalf("No mc binary found. Please install MinIO client or specify custom path with --mc-path")
	}
	if verboseMode {
		fmt.Printf("Using fallback mc binary: %s\n", versions[0])
	}
	return versions[0]
}

func displayTraceResult(result trace.Result) {
	window := result.Duration.Round(100 * time.Millisecond)
	if window == 0 {
		window = result.Duration
	}

	fmt.Printf("=== Trace Summary (%s captured) ===\n", window)

	if result.TotalEvents == 0 {
		fmt.Println("No error events were captured within the selected window.")
		return
	}

	fmt.Printf("Events captured: %d\n", result.TotalEvents)
	fmt.Printf("Distinct error patterns: %d\n", len(result.ErrorStats))
	fmt.Printf("Objects with errors: %d\n\n", len(result.Stats))

	const topErrorLimit = 5
	const topObjectLimit = 5
	const topGroupObjectLimit = 10

	if len(result.ErrorStats) > 0 {
		fmt.Println("Top error patterns:")
		for i, errStat := range result.ErrorStats {
			if i >= topErrorLimit {
				break
			}
			fmt.Printf("  %d. %s — %d events\n", i+1, errStat.Message, errStat.Count)
			for j, obj := range errStat.Objects {
				if j >= topObjectLimit {
					break
				}
				fmt.Printf("       • %s (%d)\n", obj.Name, obj.Count)
			}
		}
		fmt.Println()
	} else {
		fmt.Println("No error patterns detected.")
		fmt.Println()
	}

	if len(result.Stats) > 0 {
		fmt.Println("Top objects with errors:")
		for i, stat := range result.Stats {
			if i >= topObjectLimit {
				break
			}
			fmt.Printf("  %d. %s — %d events\n", i+1, stat.Name, stat.Count)
			if detail := formatTopErrorCounts(stat.ErrorCounts, 3); detail != "" {
				fmt.Printf("       errors: %s\n", detail)
			} else if len(stat.SampleErrors) > 0 {
				fmt.Printf("       sample: %s\n", strings.Join(stat.SampleErrors, "; "))
			}
		}
		fmt.Println()
	} else {
		fmt.Println("No objects recorded in trace output.")
		fmt.Println()
	}

	if len(result.APIStats) > 0 {
		fmt.Println("APIs with errors:")
		for i, apiStat := range result.APIStats {
			if i >= topObjectLimit {
				break
			}
			errorSummary := formatTopErrorCounts(apiStat.ErrorCounts, 3)
			fmt.Printf("  %d. %s — %d events across %d objects\n", i+1, apiStat.Name, apiStat.Count, len(apiStat.Objects))
			if len(apiStat.Objects) > 0 {
				fmt.Printf("       objects: %s\n", formatObjectCounts(apiStat.Objects, topGroupObjectLimit))
			}
			if errorSummary != "" {
				fmt.Printf("       errors: %s\n", errorSummary)
			}
		}
		fmt.Println()
	}

	if len(result.ClientStats) > 0 {
		fmt.Println("Clients with errors:")
		for i, clientStat := range result.ClientStats {
			if i >= topObjectLimit {
				break
			}
			errorSummary := formatTopErrorCounts(clientStat.ErrorCounts, 3)
			fmt.Printf("  %d. %s — %d events across %d objects\n", i+1, clientStat.Name, clientStat.Count, len(clientStat.Objects))
			if len(clientStat.Objects) > 0 {
				fmt.Printf("       objects: %s\n", formatObjectCounts(clientStat.Objects, topGroupObjectLimit))
			}
			if errorSummary != "" {
				fmt.Printf("       errors: %s\n", errorSummary)
			}
		}
		fmt.Println()
	}

	if verbose && len(result.Stats) > topObjectLimit {
		fmt.Printf("Additional objects with errors (%d more):\n", len(result.Stats)-topObjectLimit)
		for _, stat := range result.Stats[topObjectLimit:] {
			fmt.Printf("  - %s (%d events)\n", stat.Name, stat.Count)
		}
		fmt.Println()
	}
}

func formatTopErrorCounts(counts map[string]int, limit int) string {
	if len(counts) == 0 {
		return ""
	}

	type entry struct {
		name  string
		count int
	}

	items := make([]entry, 0, len(counts))
	for name, count := range counts {
		items = append(items, entry{name: name, count: count})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].count == items[j].count {
			return items[i].name < items[j].name
		}
		return items[i].count > items[j].count
	})

	maxItems := limit
	if maxItems > len(items) {
		maxItems = len(items)
	}

	parts := make([]string, 0, maxItems)
	for i := 0; i < maxItems; i++ {
		parts = append(parts, fmt.Sprintf("\n         • %s (%d)", items[i].name, items[i].count))
	}

	return strings.Join(parts, ", ")
}

func formatObjectCounts(items []trace.ObjectCount, limit int) string {
	if len(items) == 0 {
		return ""
	}

	maxItems := limit
	if maxItems > len(items) {
		maxItems = len(items)
	}

	parts := make([]string, 0, maxItems)
	for i := 0; i < maxItems; i++ {
		parts = append(parts, fmt.Sprintf("\n         • %s (%d)", items[i].Name, items[i].Count))
	}

	if len(items) > maxItems {
		parts = append(parts, fmt.Sprintf("\n         +%d more", len(items)-maxItems))
	}

	return strings.Join(parts, ", ")
}

func runProfile(cmd *cobra.Command, args []string) {
	alias := args[0]

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

	// Check if mc binary exists
	mcBinary := profileMCPath
	if _, err := exec.LookPath(mcBinary); err != nil {
		log.Fatalf("MC binary not found: %s\nPlease install mc21 or specify path with --mc-path", mcBinary)
	}

	// Create profile options
	opts := profile.MC21ProfileOptions{
		Alias:       alias,
		ProfileType: profileType,
		Duration:    duration,
		Output:      profileOutput,
		Verbose:     verbose,
		MCPath:      mcBinary,
		Insecure:    insecure,
	}

	// Run profiling
	result, err := profile.RunMC21Profile(opts)
	if err != nil {
		log.Fatalf("Profiling failed: %v", err)
	}

	// Display analysis commands
	profile.PrintProfileCommands(result)
}

func runPerftest(cmd *cobra.Command, args []string) {
	// Get site alias from args
	siteAlias := args[0]

	// Validate required flags
	if perftestBucket == "" {
		log.Fatal("Error: --bucket flag is required")
	}

	// Validate size preset
	var sizeType perftest.ObjectSizeType
	switch strings.ToLower(perftestSize) {
	case "small":
		sizeType = perftest.ObjectSizeSmall
	case "medium":
		sizeType = perftest.ObjectSizeMedium
	case "large":
		sizeType = perftest.ObjectSizeLarge
	default:
		log.Fatalf("Invalid size preset '%s'. Valid options: small, medium, large", perftestSize)
	}

	// Validate upload mode
	var uploadMode perftest.UploadMode
	switch strings.ToLower(perftestMode) {
	case "all":
		uploadMode = perftest.UploadModeAll
	case "interval":
		uploadMode = perftest.UploadModeInterval
	default:
		log.Fatalf("Invalid upload mode '%s'. Valid options: all, interval", perftestMode)
	}

	// Parse interval duration
	uploadInterval, err := time.ParseDuration(perftestInterval)
	if err != nil {
		log.Fatalf("Invalid interval: %v", err)
	}

	// Auto-generate path with timestamp if not specified
	if perftestPath == "" {
		timestamp := time.Now().Format("20060102-150405")
		perftestPath = fmt.Sprintf("mc-test/%s/", timestamp)
	} else if !strings.HasSuffix(perftestPath, "/") {
		// Ensure path ends with /
		perftestPath += "/"
	}

	// Load MC configuration
	cfg, err := config.LoadMCConfig()
	if err != nil {
		log.Fatalf("Error loading MC configuration: %v", err)
	}

	// Validate site exists
	if _, exists := cfg.Aliases[siteAlias]; !exists {
		log.Fatalf("Site alias '%s' not found in MC config", siteAlias)
	}

	fmt.Printf("🚀 Starting MinIO Performance Test\n\n")
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Site:         %s\n", siteAlias)
	fmt.Printf("  Bucket:       %s\n", perftestBucket)
	fmt.Printf("  Path:         %s\n", perftestPath)
	fmt.Printf("  Object Size:  %s\n", sizeType)
	fmt.Printf("  Object Count: %d\n", perftestCount)
	fmt.Printf("  Override:     %d times\n", perftestOverride)
	fmt.Printf("  Upload Mode:  %s\n", uploadMode)
	if uploadMode == perftest.UploadModeAll {
		fmt.Printf("  Parallelism:  %d\n", perftestParallel)
	} else {
		fmt.Printf("  Iterations:   %d rounds\n", perftestIterations)
		fmt.Printf("  Interval:     %v\n", uploadInterval)
		fmt.Printf("  Total:        %d objects (%d per round)\n", perftestCount*perftestIterations, perftestCount)
	}
	fmt.Printf("\n")

	// Create test configuration
	testConfig := &perftest.TestConfig{
		SiteAlias:      siteAlias,
		Bucket:         perftestBucket,
		ObjectPath:     perftestPath,
		ObjectSizeType: sizeType,
		ObjectCount:    perftestCount,
		OverrideCount:  perftestOverride,
		UploadMode:     uploadMode,
		UploadInterval: uploadInterval,
		Iterations:     perftestIterations,
		Parallelism:    perftestParallel,
	}

	// Create test runner
	runner, err := perftest.NewRunner(testConfig, cfg)
	if err != nil {
		log.Fatalf("Failed to create test runner: %v", err)
	}
	defer runner.Cleanup()

	// Run test
	ctx := context.Background()
	result, err := runner.Run(ctx)
	if err != nil {
		log.Fatalf("Test failed: %v", err)
	}

	// Display results
	displayPerftestResults(result)
}

func displayPerftestResults(result *perftest.TestResult) {
	fmt.Printf("\n")
	fmt.Printf("=" + strings.Repeat("=", 79) + "\n")
	fmt.Printf("Test Results\n")
	fmt.Printf("=" + strings.Repeat("=", 79) + "\n\n")

	fmt.Printf("Site:         %s\n", result.Config.SiteAlias)
	fmt.Printf("Bucket:       %s\n", result.Config.Bucket)
	fmt.Printf("Duration:     %v\n", result.TotalDuration)
	fmt.Printf("\n")

	// Upload Summary
	fmt.Printf("Upload Summary:\n")
	fmt.Printf("  Total Uploads:       %d\n", result.Summary.TotalUploads)
	fmt.Printf("  Successful:          %d\n", result.Summary.SuccessfulUploads)
	fmt.Printf("  Failed:              %d\n", result.Summary.FailedUploads)
	fmt.Printf("  Unique Objects:      %d\n", result.Summary.UniqueObjects)
	fmt.Printf("  Overridden Objects:  %d\n", result.Summary.OverriddenObjects)
	fmt.Printf("  Total Data Uploaded: %s\n", formatBytes(result.Summary.TotalDataUploaded))
	fmt.Printf("\n")

	// Performance Metrics
	fmt.Printf("Performance Metrics:\n")
	fmt.Printf("  Average Upload Time: %v\n", result.Summary.AverageUploadLatency)
	fmt.Printf("  Throughput:          %.2f uploads/sec\n", result.Summary.Throughput)
	fmt.Printf("  Data Throughput:     %.2f KB/sec\n", result.Summary.DataThroughput/1024)
	fmt.Printf("\n")

	// Override Details
	if result.Summary.OverriddenObjects > 0 {
		fmt.Printf("Override Details:\n")
		for objKey, count := range result.Summary.OverrideDetails {
			fmt.Printf("  %s: overridden %d times\n", objKey, count)
		}
		fmt.Printf("\n")
	}

	// Errors
	if len(result.Errors) > 0 {
		fmt.Printf("Errors:\n")
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
		fmt.Printf("\n")
	}

	fmt.Printf("Test completed successfully!\n")
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
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
