package profile

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ProfileOptions contains options for profiling
type ProfileOptions struct {
	Alias           string
	ProfileType     string // cpu, heap, goroutine, allocs, block, mutex
	Duration        time.Duration
	Output          string
	Verbose         bool
	MCPath          string // Path to mc binary (mc or mc-2021)
	DetectLeaks     bool
	MonitorInterval time.Duration
	ThresholdMB     int // Memory threshold for leak detection
}

// MemoryStats represents memory profiling statistics
type MemoryStats struct {
	Timestamp    time.Time
	AllocMB      float64
	TotalAllocMB float64
	SysMB        float64
	NumGC        int
	HeapObjects  int
	GoRoutines   int
}

// LeakDetection represents potential memory leak information
type LeakDetection struct {
	Detected     bool
	GrowthRateMB float64
	TrendPeriod  time.Duration
	Samples      []MemoryStats
}

// RunProfile executes mc admin profile or mc support profile command
func RunProfile(opts ProfileOptions) error {
	if opts.Verbose {
		fmt.Printf("🔍 Starting %s profile for alias: %s\n", opts.ProfileType, opts.Alias)
		fmt.Printf("⏱️  Duration: %s\n", opts.Duration)
		if opts.Output != "" {
			fmt.Printf("📁 Output: %s\n", opts.Output)
		}
		fmt.Printf("🔧 MC Binary: %s\n", opts.MCPath)
		fmt.Println()
	}

	// Create output directory if specified
	if opts.Output != "" {
		if err := os.MkdirAll(filepath.Dir(opts.Output), 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
	}

	// Try mc support profile first (newer versions), then fall back to mc admin profile
	var args []string

	// First try the newer mc support profile command
	args = []string{"support", "profile", opts.ProfileType, opts.Alias}
	if opts.Duration > 0 {
		args = append(args, fmt.Sprintf("--duration=%s", opts.Duration))
	}

	cmd := exec.Command(opts.MCPath, args...)

	if opts.Verbose {
		fmt.Printf("🚀 Executing: %s %s\n", opts.MCPath, strings.Join(args, " "))
	}

	// Test the command first
	testOutput, testErr := cmd.CombinedOutput()

	// If support profile fails, try admin profile (older versions)
	if testErr != nil {
		if opts.Verbose {
			fmt.Printf("⚠️  mc support profile failed, trying mc admin profile...\n")
		}

		args = []string{"admin", "profile", opts.ProfileType, opts.Alias}
		if opts.Duration > 0 {
			args = append(args, fmt.Sprintf("--duration=%s", opts.Duration))
		}

		cmd = exec.Command(opts.MCPath, args...)

		if opts.Verbose {
			fmt.Printf("🚀 Executing: %s %s\n", opts.MCPath, strings.Join(args, " "))
		}
	} else {
		// support profile worked, continue with it
		cmd = exec.Command(opts.MCPath, args...)
	}

	// Handle output
	var output io.Writer
	if opts.Output != "" {
		file, err := os.Create(opts.Output)
		if err != nil {
			return fmt.Errorf("failed to create output file: %v", err)
		}
		defer file.Close()
		output = file

		if opts.Verbose {
			// Tee output to both file and stdout
			output = io.MultiWriter(file, os.Stdout)
		}
	} else {
		output = os.Stdout
	}

	cmd.Stdout = output
	cmd.Stderr = os.Stderr

	// Execute the command
	if err := cmd.Run(); err != nil {
		// If both commands failed, show specific error messages
		if strings.Contains(string(testOutput), "Deprecated command") {
			return fmt.Errorf("mc admin profile is deprecated in this mc version. Try with mc-2021: --mc-path mc-2021")
		}
		return fmt.Errorf("mc profile failed: %v", err)
	}

	if opts.Verbose {
		fmt.Printf("\n✅ Profile completed successfully\n")
		if opts.Output != "" {
			fmt.Printf("📄 Profile saved to: %s\n", opts.Output)
		}
	}

	return nil
}

// MonitorMemoryLeaks continuously monitors for memory leaks using heap profiling
func MonitorMemoryLeaks(opts ProfileOptions) error {
	if opts.Verbose {
		fmt.Printf("🔍 Starting memory leak detection for alias: %s\n", opts.Alias)
		fmt.Printf("📊 Monitor interval: %s\n", opts.MonitorInterval)
		fmt.Printf("🚨 Memory threshold: %d MB\n", opts.ThresholdMB)
		fmt.Println()
	}

	var samples []MemoryStats

	// Create a context with timeout if duration is specified
	ctx := context.Background()
	if opts.Duration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Duration)
		defer cancel()
	}

	ticker := time.NewTicker(opts.MonitorInterval)
	defer ticker.Stop()

	sampleCount := 0
	for {
		select {
		case <-ctx.Done():
			if opts.Verbose {
				fmt.Printf("\n⏰ Monitoring duration completed\n")
			}
			return analyzeLeakDetection(samples, opts)
		case <-ticker.C:
			sampleCount++
			if opts.Verbose {
				fmt.Printf("📊 Taking sample %d...\n", sampleCount)
			}

			// Take a heap profile sample
			stats, err := takeMemorySample(opts)
			if err != nil {
				fmt.Printf("⚠️  Warning: Failed to take memory sample: %v\n", err)
				continue
			}

			samples = append(samples, stats)

			// Display current stats
			fmt.Printf("[%s] Alloc: %.2f MB, Total: %.2f MB, Sys: %.2f MB, Goroutines: %d, Objects: %d\n",
				stats.Timestamp.Format("15:04:05"),
				stats.AllocMB,
				stats.TotalAllocMB,
				stats.SysMB,
				stats.GoRoutines,
				stats.HeapObjects)

			// Check for immediate leak indicators
			if len(samples) > 2 {
				checkImmediateLeak(samples, opts)
			}
		}
	}
}

// takeMemorySample takes a single memory profile sample
func takeMemorySample(opts ProfileOptions) (MemoryStats, error) {
	// Try mc support profile first, then mc admin profile
	args := []string{"support", "profile", "heap", opts.Alias, "--duration=1s"}

	cmd := exec.Command(opts.MCPath, args...)
	output, err := cmd.Output()

	// If support profile fails, try admin profile
	if err != nil {
		args = []string{"admin", "profile", "heap", opts.Alias, "--duration=1s"}
		cmd = exec.Command(opts.MCPath, args...)
		output, err = cmd.Output()

		if err != nil {
			return MemoryStats{}, fmt.Errorf("failed to get heap profile: %v", err)
		}
	}

	// Parse the heap profile output to extract memory statistics
	stats := parseHeapProfile(string(output))
	stats.Timestamp = time.Now()

	return stats, nil
}

// parseHeapProfile parses heap profile output to extract memory statistics
func parseHeapProfile(profileData string) MemoryStats {
	stats := MemoryStats{}

	// Look for runtime.MemStats information in the profile
	// This is a simplified parser - in a real implementation you'd use pprof tools
	lines := strings.Split(profileData, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse memory allocation information from heap profile
		if strings.Contains(line, "MB") {
			// Extract numeric values from lines containing memory information
			re := regexp.MustCompile(`(\d+\.?\d*)\s*MB`)
			matches := re.FindAllStringSubmatch(line, -1)

			if len(matches) > 0 {
				if val, err := strconv.ParseFloat(matches[0][1], 64); err == nil {
					if strings.Contains(line, "heap profile:") {
						// First number in heap profile line
						stats.AllocMB = val
					} else if strings.Contains(line, "alloc") || (strings.Contains(line, "heap") && stats.AllocMB == 0) {
						stats.AllocMB = val
					} else if strings.Contains(line, "sys") || strings.Contains(line, "System") {
						stats.SysMB = val
					} else if strings.Contains(line, "Total") {
						stats.TotalAllocMB = val
					}
				}
			}
		}

		// Extract goroutine count - try multiple patterns
		if strings.Contains(line, "goroutine") {
			// Pattern 1: "goroutine profile: total 250"
			re := regexp.MustCompile(`total\s+(\d+)`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if count, err := strconv.Atoi(matches[1]); err == nil {
					stats.GoRoutines = count
				}
			}

			// Pattern 2: "Number of goroutines: 150"
			re = regexp.MustCompile(`Number of goroutines:\s*(\d+)`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if count, err := strconv.Atoi(matches[1]); err == nil {
					stats.GoRoutines = count
				}
			}

			// Pattern 3: "250 goroutine"
			re = regexp.MustCompile(`(\d+)\s+goroutine`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if count, err := strconv.Atoi(matches[1]); err == nil {
					stats.GoRoutines = count
				}
			}
		}

		// Extract object count
		if strings.Contains(line, "objects") {
			re := regexp.MustCompile(`(\d+)\s+objects`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if count, err := strconv.Atoi(matches[1]); err == nil {
					stats.HeapObjects = count
				}
			}
		}
	}

	return stats
}

// checkImmediateLeak checks for immediate leak indicators
func checkImmediateLeak(samples []MemoryStats, opts ProfileOptions) {
	if len(samples) < 3 {
		return
	}

	latest := samples[len(samples)-1]
	previous := samples[len(samples)-2]

	// Check for significant memory growth
	growthMB := latest.AllocMB - previous.AllocMB
	if growthMB > float64(opts.ThresholdMB) {
		fmt.Printf("🚨 POTENTIAL LEAK DETECTED: Memory grew by %.2f MB in the last interval\n", growthMB)
	}

	// Check for goroutine growth
	goroutineGrowth := latest.GoRoutines - previous.GoRoutines
	if goroutineGrowth > 50 {
		fmt.Printf("🚨 GOROUTINE LEAK DETECTED: %d new goroutines in the last interval\n", goroutineGrowth)
	}
}

// analyzeLeakDetection performs comprehensive leak analysis
func analyzeLeakDetection(samples []MemoryStats, opts ProfileOptions) error {
	if len(samples) < 2 {
		fmt.Printf("❌ Insufficient samples for leak analysis (need at least 2, got %d)\n", len(samples))
		return nil
	}

	fmt.Printf("\n=== Memory Leak Analysis Report ===\n")
	fmt.Printf("📊 Total samples: %d\n", len(samples))
	fmt.Printf("⏱️  Monitoring duration: %s\n", samples[len(samples)-1].Timestamp.Sub(samples[0].Timestamp))

	first := samples[0]
	last := samples[len(samples)-1]

	// Calculate growth rates
	totalGrowthMB := last.AllocMB - first.AllocMB
	durationHours := last.Timestamp.Sub(first.Timestamp).Hours()
	hourlyGrowthMB := totalGrowthMB / durationHours

	goroutineGrowth := last.GoRoutines - first.GoRoutines
	objectGrowth := last.HeapObjects - first.HeapObjects

	fmt.Printf("\n📈 Memory Growth Analysis:\n")
	fmt.Printf("  Initial Memory: %.2f MB\n", first.AllocMB)
	fmt.Printf("  Final Memory: %.2f MB\n", last.AllocMB)
	fmt.Printf("  Total Growth: %.2f MB\n", totalGrowthMB)
	fmt.Printf("  Hourly Growth Rate: %.2f MB/hour\n", hourlyGrowthMB)

	fmt.Printf("\n🔄 Goroutine Analysis:\n")
	fmt.Printf("  Initial Goroutines: %d\n", first.GoRoutines)
	fmt.Printf("  Final Goroutines: %d\n", last.GoRoutines)
	fmt.Printf("  Growth: %d\n", goroutineGrowth)

	fmt.Printf("\n📦 Object Analysis:\n")
	fmt.Printf("  Initial Objects: %d\n", first.HeapObjects)
	fmt.Printf("  Final Objects: %d\n", last.HeapObjects)
	fmt.Printf("  Growth: %d\n", objectGrowth)

	// Leak detection logic
	leakDetected := false
	fmt.Printf("\n🕵️  Leak Detection Results:\n")

	if hourlyGrowthMB > 10.0 {
		fmt.Printf("  🚨 MEMORY LEAK LIKELY: High memory growth rate (%.2f MB/hour)\n", hourlyGrowthMB)
		leakDetected = true
	}

	if goroutineGrowth > 100 {
		fmt.Printf("  🚨 GOROUTINE LEAK DETECTED: Significant goroutine growth (%d)\n", goroutineGrowth)
		leakDetected = true
	}

	if objectGrowth > 100000 {
		fmt.Printf("  🚨 OBJECT LEAK DETECTED: Significant object growth (%d)\n", objectGrowth)
		leakDetected = true
	}

	if !leakDetected {
		if totalGrowthMB < 5.0 && goroutineGrowth < 10 {
			fmt.Printf("  ✅ NO LEAKS DETECTED: Memory usage appears stable\n")
		} else {
			fmt.Printf("  ⚠️  MINOR GROWTH: Some growth detected but within normal ranges\n")
		}
	}

	// Recommendations
	fmt.Printf("\n💡 Recommendations:\n")
	if leakDetected {
		fmt.Printf("  • Take a detailed heap profile for analysis: mc-tool profile heap %s\n", opts.Alias)
		fmt.Printf("  • Consider enabling GC profiling: mc-tool profile allocs %s\n", opts.Alias)
		fmt.Printf("  • Monitor for longer periods to confirm trends\n")
		fmt.Printf("  • Check application logs for error patterns\n")
	} else {
		fmt.Printf("  • Continue periodic monitoring\n")
		fmt.Printf("  • Consider baseline profiling during low load periods\n")
	}

	return nil
}

// GetAvailableMCVersions returns available mc binary versions
func GetAvailableMCVersions() []string {
	var versions []string

	// Check for standard mc
	if _, err := exec.LookPath("mc"); err == nil {
		versions = append(versions, "mc")
	}

	// Check for mc-2021 (older version)
	if _, err := exec.LookPath("mc-2021"); err == nil {
		versions = append(versions, "mc-2021")
	}

	// Check for custom paths
	customPaths := []string{
		"/usr/local/bin/mc",
		"/usr/local/bin/mc-2021",
		"/usr/bin/mc",
		"/usr/bin/mc-2021",
	}

	for _, path := range customPaths {
		if _, err := os.Stat(path); err == nil {
			// Don't add duplicates
			found := false
			for _, v := range versions {
				if v == path {
					found = true
					break
				}
			}
			if !found {
				versions = append(versions, path)
			}
		}
	}

	return versions
}

// TestMCAdminProfile tests if mc admin profile or mc support profile command is available
func TestMCAdminProfile(mcPath string, alias string) error {
	// Test with mc support profile first (newer versions)
	args := []string{"support", "profile", "cpu", alias, "--duration=1s"}
	cmd := exec.Command(mcPath, args...)

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil // support profile works
	}

	// If support profile fails, try admin profile (older versions)
	args = []string{"admin", "profile", "cpu", alias, "--duration=1s"}
	cmd = exec.Command(mcPath, args...)

	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mc profile test failed: %v\nOutput: %s", err, string(output))
	}

	return nil
}

// ParseHeapProfileForTesting exports parseHeapProfile for testing purposes
func ParseHeapProfileForTesting(profileData string) MemoryStats {
	return parseHeapProfile(profileData)
}
