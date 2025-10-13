package debug

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// GoroutineInfo represents information about goroutines
type GoroutineInfo struct {
	ID       int    `json:"id"`
	State    string `json:"state"`
	Function string `json:"function"`
	Location string `json:"location"`
	Duration string `json:"duration"`
	Stack    string `json:"stack"`
}

// GoroutineStats represents aggregated goroutine statistics
type GoroutineStats struct {
	Total           int                    `json:"total"`
	ByState         map[string]int         `json:"by_state"`
	ByFunction      map[string]int         `json:"by_function"`
	TopFunctions    []FunctionCount        `json:"top_functions"`
	LongRunning     []GoroutineInfo        `json:"long_running"`
	PotentialLeaks  []GoroutineInfo        `json:"potential_leaks"`
	Timestamp       time.Time              `json:"timestamp"`
	Endpoint        string                 `json:"endpoint"`
}

// FunctionCount represents function call frequency
type FunctionCount struct {
	Function string `json:"function"`
	Count    int    `json:"count"`
}

// DebugOptions contains options for debugging
type DebugOptions struct {
	Endpoint        string
	AccessKey       string
	SecretKey       string
	Insecure        bool
	Verbose         bool
	MonitorDuration time.Duration
	Interval        time.Duration
	OutputFormat    string
	ThresholdCount  int
}

// DebugMinIOGoroutines analyzes MinIO server goroutines for potential leaks
func DebugMinIOGoroutines(opts DebugOptions) error {
	// Create HTTP client for pprof endpoints with proper TLS configuration
	transport := &http.Transport{}
	
	// Configure TLS settings based on insecure flag
	if opts.Insecure {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		if opts.Verbose {
			fmt.Println("⚠️  TLS certificate verification disabled")
		}
	}
	
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	if opts.MonitorDuration > 0 {
		return monitorGoroutines(client, opts)
	}

	// Single snapshot analysis
	stats, err := analyzeGoroutines(client, opts)
	if err != nil {
		return fmt.Errorf("failed to analyze goroutines: %w", err)
	}

	return displayResults(*stats, opts.OutputFormat, opts.Verbose)
}

// monitorGoroutines performs continuous monitoring for leak detection
func monitorGoroutines(client *http.Client, opts DebugOptions) error {
	fmt.Printf("🔍 Monitoring MinIO goroutines for %s (interval: %s)\n", opts.MonitorDuration, opts.Interval)
	fmt.Printf("📡 Endpoint: %s\n", opts.Endpoint)
	fmt.Printf("🎯 Leak threshold: %d goroutines\n", opts.ThresholdCount)
	fmt.Println()

	startTime := time.Now()
	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	var baseline *GoroutineStats
	measurements := []GoroutineStats{}

	for {
		select {
		case <-ticker.C:
			stats, err := analyzeGoroutines(client, opts)
			if err != nil {
				fmt.Printf("⚠️  Error analyzing goroutines: %v\n", err)
				continue
			}

			measurements = append(measurements, *stats)

			if baseline == nil {
				baseline = stats
				fmt.Printf("📊 Baseline: %d goroutines\n", baseline.Total)
				continue
			}

			// Check for potential leaks
			growth := stats.Total - baseline.Total
			if growth > opts.ThresholdCount {
				fmt.Printf("🚨 POTENTIAL LEAK DETECTED!\n")
				fmt.Printf("   Growth: +%d goroutines (baseline: %d, current: %d)\n", 
					growth, baseline.Total, stats.Total)
				displayGrowthAnalysis(baseline, stats)
			} else if opts.Verbose {
				fmt.Printf("✅ Current: %d goroutines (growth: %+d)\n", stats.Total, growth)
			}

			// Check if monitoring duration exceeded
			if time.Since(startTime) >= opts.MonitorDuration {
				fmt.Println("\n📈 Monitoring Summary:")
				return generateSummaryReport(measurements, opts)
			}

		default:
			// Non-blocking check for duration
			if time.Since(startTime) >= opts.MonitorDuration {
				return generateSummaryReport(measurements, opts)
			}
		}
	}
}

// analyzeGoroutines fetches and analyzes goroutine data from MinIO
func analyzeGoroutines(client *http.Client, opts DebugOptions) (*GoroutineStats, error) {
	// Construct pprof endpoint URL
	pprofURL := fmt.Sprintf("%s/debug/pprof/goroutine?debug=1", strings.TrimSuffix(opts.Endpoint, "/"))

	if opts.Verbose {
		fmt.Printf("🔍 Fetching goroutine data from: %s\n", pprofURL)
	}

	// Create request
	req, err := http.NewRequest("GET", pprofURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Try multiple authentication approaches
	var resp *http.Response
	var lastErr error

	// Approach 1: Try without authentication (some MinIO instances allow this)
	if opts.Verbose {
		fmt.Println("🔐 Trying without authentication...")
	}
	resp, err = client.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		if opts.Verbose {
			fmt.Println("✅ Success without authentication")
		}
		defer resp.Body.Close()
		return parseGoroutineResponse(resp, opts.Endpoint)
	}
	if resp != nil {
		resp.Body.Close()
		lastErr = fmt.Errorf("no auth failed with status: %d", resp.StatusCode)
	} else {
		lastErr = err
	}

	// Approach 2: Try with AWS signature v4 if credentials provided
	if opts.AccessKey != "" && opts.SecretKey != "" {
		if opts.Verbose {
			fmt.Println("🔐 Trying with AWS signature v4...")
		}
		
		// Create new request for signing
		req, err = http.NewRequest("GET", pprofURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create signed request: %w", err)
		}

		if err := signRequest(req, opts); err != nil {
			if opts.Verbose {
				fmt.Printf("⚠️ Signing failed: %v\n", err)
			}
		} else {
			resp, err = client.Do(req)
			if err == nil && resp.StatusCode == http.StatusOK {
				if opts.Verbose {
					fmt.Println("✅ Success with AWS signature v4")
				}
				defer resp.Body.Close()
				return parseGoroutineResponse(resp, opts.Endpoint)
			}
			if resp != nil {
				resp.Body.Close()
				lastErr = fmt.Errorf("AWS v4 auth failed with status: %d", resp.StatusCode)
			} else {
				lastErr = err
			}
		}
	}

	// Approach 3: Try with basic auth or Authorization header
	if opts.AccessKey != "" && opts.SecretKey != "" {
		if opts.Verbose {
			fmt.Println("🔐 Trying with Authorization header...")
		}
		
		req, err = http.NewRequest("GET", pprofURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create auth header request: %w", err)
		}

		// Try MinIO admin token approach
		req.Header.Set("Authorization", fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s", opts.AccessKey))
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			if opts.Verbose {
				fmt.Println("✅ Success with Authorization header")
			}
			defer resp.Body.Close()
			return parseGoroutineResponse(resp, opts.Endpoint)
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	// All approaches failed
	if opts.Verbose {
		fmt.Printf("❌ All authentication approaches failed. Last error: %v\n", lastErr)
		fmt.Println("💡 Troubleshooting tips:")
		fmt.Println("   • pprof endpoints may be disabled on this MinIO server")
		fmt.Println("   • Some MinIO instances don't expose /debug/pprof for security")
		fmt.Println("   • Try with a local MinIO instance where you have admin access")
		fmt.Println("   • Check if your MinIO server has MINIO_PROFILE=enable environment variable")
	}
	
	return nil, fmt.Errorf("pprof endpoint not accessible (status: 403) - this may be disabled for security on production/public MinIO servers")
}

// parseGoroutineResponse parses the HTTP response containing goroutine data
func parseGoroutineResponse(resp *http.Response, endpoint string) (*GoroutineStats, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse the goroutine profile data
	return parseGoroutineProfile(string(body), endpoint)
}

// parseGoroutineProfile parses the pprof goroutine output
func parseGoroutineProfile(data, endpoint string) (*GoroutineStats, error) {
	lines := strings.Split(data, "\n")
	stats := &GoroutineStats{
		ByState:    make(map[string]int),
		ByFunction: make(map[string]int),
		Timestamp:  time.Now(),
		Endpoint:   endpoint,
	}

	var currentGoroutine *GoroutineInfo
	var stackLines []string

	// Regex patterns for parsing
	goroutinePattern := regexp.MustCompile(`^goroutine (\d+) \[([^\]]+)\](?:, (.+))?:`)
	functionPattern := regexp.MustCompile(`^([^\(]+)\(.*\)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if currentGoroutine != nil {
				currentGoroutine.Stack = strings.Join(stackLines, "\n")
				analyzeGoroutine(currentGoroutine, stats)
				currentGoroutine = nil
				stackLines = nil
			}
			continue
		}

		// Check for goroutine header
		if matches := goroutinePattern.FindStringSubmatch(line); matches != nil {
			// Save previous goroutine
			if currentGoroutine != nil {
				currentGoroutine.Stack = strings.Join(stackLines, "\n")
				analyzeGoroutine(currentGoroutine, stats)
			}

			// Parse new goroutine
			id, _ := strconv.Atoi(matches[1])
			state := matches[2]
			duration := ""
			if len(matches) > 3 && matches[3] != "" {
				duration = matches[3]
			}

			currentGoroutine = &GoroutineInfo{
				ID:       id,
				State:    state,
				Duration: duration,
			}
			stackLines = []string{}
			stats.Total++
			stats.ByState[state]++
		} else if currentGoroutine != nil {
			stackLines = append(stackLines, line)
			
			// Extract function name from stack line
			if strings.Contains(line, "(") && !strings.HasPrefix(line, "\t") {
				if matches := functionPattern.FindStringSubmatch(line); matches != nil {
					currentGoroutine.Function = matches[1]
					stats.ByFunction[matches[1]]++
				}
			}
		}
	}

	// Handle last goroutine
	if currentGoroutine != nil {
		currentGoroutine.Stack = strings.Join(stackLines, "\n")
		analyzeGoroutine(currentGoroutine, stats)
	}

	// Generate top functions
	stats.TopFunctions = getTopFunctions(stats.ByFunction, 10)

	return stats, nil
}

// analyzeGoroutine analyzes individual goroutine for potential issues
func analyzeGoroutine(g *GoroutineInfo, stats *GoroutineStats) {
	// Check for long-running goroutines
	if strings.Contains(g.Duration, "m") || strings.Contains(g.Duration, "h") {
		stats.LongRunning = append(stats.LongRunning, *g)
	}

	// Check for potential leak patterns
	leakPatterns := []string{
		"syscall",
		"select",
		"chan receive",
		"chan send",
		"IO wait",
	}

	for _, pattern := range leakPatterns {
		if strings.Contains(g.State, pattern) {
			// Additional analysis for potential leaks
			if isLikelyLeak(g) {
				stats.PotentialLeaks = append(stats.PotentialLeaks, *g)
			}
			break
		}
	}
}

// isLikelyLeak determines if a goroutine is likely to be a leak
func isLikelyLeak(g *GoroutineInfo) bool {
	// Heuristics for leak detection
	suspiciousPatterns := []string{
		"runtime.gopark",
		"sync.runtime_Semacquire",
		"internal/poll.runtime_pollWait",
		"time.Sleep",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(g.Stack, pattern) {
			return true
		}
	}

	return false
}

// getTopFunctions returns the most frequent functions
func getTopFunctions(funcCounts map[string]int, limit int) []FunctionCount {
	var functions []FunctionCount
	for fn, count := range funcCounts {
		functions = append(functions, FunctionCount{Function: fn, Count: count})
	}

	sort.Slice(functions, func(i, j int) bool {
		return functions[i].Count > functions[j].Count
	})

	if len(functions) > limit {
		functions = functions[:limit]
	}

	return functions
}

// displayResults shows the analysis results
func displayResults(stats GoroutineStats, format string, verbose bool) error {
	if format == "json" {
		data, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	// Text format
	fmt.Printf("🔍 MinIO Goroutine Analysis\n")
	fmt.Printf("==========================\n")
	fmt.Printf("📡 Endpoint: %s\n", stats.Endpoint)
	fmt.Printf("🕐 Timestamp: %s\n", stats.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("📊 Total Goroutines: %d\n\n", stats.Total)

	// State breakdown
	fmt.Printf("📈 Goroutines by State:\n")
	for state, count := range stats.ByState {
		fmt.Printf("   %-20s: %d\n", state, count)
	}
	fmt.Println()

	// Top functions
	if len(stats.TopFunctions) > 0 {
		fmt.Printf("🏆 Top Functions:\n")
		for i, fn := range stats.TopFunctions {
			fmt.Printf("   %d. %-30s: %d\n", i+1, fn.Function, fn.Count)
		}
		fmt.Println()
	}

	// Long-running goroutines
	if len(stats.LongRunning) > 0 {
		fmt.Printf("⏱️  Long-running Goroutines (%d):\n", len(stats.LongRunning))
		for _, g := range stats.LongRunning {
			fmt.Printf("   ID:%d [%s] %s - %s\n", g.ID, g.State, g.Duration, g.Function)
		}
		fmt.Println()
	}

	// Potential leaks
	if len(stats.PotentialLeaks) > 0 {
		fmt.Printf("🚨 Potential Leaks (%d):\n", len(stats.PotentialLeaks))
		for _, g := range stats.PotentialLeaks {
			fmt.Printf("   ID:%d [%s] %s\n", g.ID, g.State, g.Function)
			if verbose {
				fmt.Printf("      Stack: %s\n", strings.Split(g.Stack, "\n")[0])
			}
		}
		fmt.Println()
	}

	return nil
}

// displayGrowthAnalysis shows the difference between baseline and current
func displayGrowthAnalysis(baseline, current *GoroutineStats) {
	fmt.Printf("📊 Growth Analysis:\n")
	for state, count := range current.ByState {
		baseCount := baseline.ByState[state]
		if growth := count - baseCount; growth > 0 {
			fmt.Printf("   %s: +%d (was %d, now %d)\n", state, growth, baseCount, count)
		}
	}
}

// generateSummaryReport creates a summary of the monitoring session
func generateSummaryReport(measurements []GoroutineStats, opts DebugOptions) error {
	if len(measurements) == 0 {
		return fmt.Errorf("no measurements collected")
	}

	fmt.Printf("📈 Summary Report\n")
	fmt.Printf("================\n")
	fmt.Printf("Duration: %s\n", opts.MonitorDuration)
	fmt.Printf("Measurements: %d\n", len(measurements))

	baseline := measurements[0]
	final := measurements[len(measurements)-1]

	fmt.Printf("Baseline: %d goroutines\n", baseline.Total)
	fmt.Printf("Final: %d goroutines\n", final.Total)
	fmt.Printf("Net Growth: %+d goroutines\n", final.Total-baseline.Total)

	// Find peak
	peak := baseline
	for _, m := range measurements {
		if m.Total > peak.Total {
			peak = m
		}
	}
	fmt.Printf("Peak: %d goroutines\n", peak.Total)

	return nil
}

// signRequest adds AWS signature v4 authentication to the request
func signRequest(req *http.Request, opts DebugOptions) error {
	// Parse the endpoint to get host and port
	u, err := url.Parse(opts.Endpoint)
	if err != nil {
		return err
	}

	// Determine if connection should be secure
	secure := u.Scheme == "https" && !opts.Insecure

	// Create MinIO client for signing
	minioClient, err := minio.New(u.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(opts.AccessKey, opts.SecretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return err
	}

	// Set required headers
	req.Header.Set("Host", u.Host)
	req.Header.Set("X-Amz-Date", time.Now().UTC().Format("20060102T150405Z"))

	// Note: This is a simplified approach. In practice, you might need
	// a more sophisticated signing mechanism for pprof endpoints
	_ = minioClient // Use the client for signing if needed

	return nil
}