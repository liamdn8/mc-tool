package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/liamdn8/mc-tool/pkg/profile"
)

// Playground for testing profile functionality
// This file can be run independently to test various profile scenarios

func main() {
	fmt.Println("=== MC-Tool Profile Playground ===")
	fmt.Println()

	// Test 1: Check available MC versions
	fmt.Println("1. Testing available MC versions:")
	versions := profile.GetAvailableMCVersions()
	if len(versions) == 0 {
		fmt.Println("   ❌ No MC versions found")
		fmt.Println("   💡 Install mc client: https://docs.min.io/docs/minio-client-quickstart-guide.html")
	} else {
		fmt.Printf("   ✅ Found %d MC versions:\n", len(versions))
		for _, v := range versions {
			fmt.Printf("      - %s\n", v)
		}
	}
	fmt.Println()

	// Test 2: Parse heap profile samples
	fmt.Println("2. Testing heap profile parsing:")
	testParseHeapProfile()
	fmt.Println()

	// Test 3: Memory leak detection simulation
	fmt.Println("3. Testing memory leak detection:")
	testMemoryLeakDetection()
	fmt.Println()

	// Test 4: Profile options validation
	fmt.Println("4. Testing profile options:")
	testProfileOptions()
	fmt.Println()

	// Test 5: MC admin profile availability (if alias provided)
	if len(os.Args) > 1 {
		alias := os.Args[1]
		fmt.Printf("5. Testing MC admin profile with alias '%s':\n", alias)
		testMCAdminProfile(alias)
	} else {
		fmt.Println("5. MC admin profile test skipped (no alias provided)")
		fmt.Println("   💡 Run with: go run playground_test.go <alias-name>")
	}
	fmt.Println()

	fmt.Println("=== Playground Tests Completed ===")
}

func testParseHeapProfile() {
	testCases := []struct {
		name        string
		profileData string
		description string
	}{
		{
			name: "heap_with_memory",
			profileData: `heap profile: 1024: 512 MB [2048: 1024 MB] @ heap/1048576
# runtime.MemProfileRecord
# Memory profile, heap statistics
Total: 512 MB
Active: 256 MB
System: 768 MB
Number of goroutines: 1847`,
			description: "Heap profile with memory statistics",
		},
		{
			name: "goroutine_profile",
			profileData: `goroutine profile: total 2156
2156 @ 0x463f71 0x4640f6 0x4640d7 0x4066c1
#	0x463f70	runtime.gopark+0x70
#	0x4640f5	runtime.selectgo+0x2d5

250 @ 0x463f71 0x46e001 0x46dfd7 0x4066c1
#	0x463f70	runtime.gopark+0x70
#	0x46e000	runtime.netpollblock+0x70`,
			description: "Goroutine profile with stack traces",
		},
		{
			name: "allocation_profile",
			profileData: `heap profile: 145672 objects, 278.3 MB
Total allocations: 2.9 GB
Object count: 145672
System memory: 312.4 MB
GC cycles: 42`,
			description: "Memory allocation profile",
		},
	}

	for _, tc := range testCases {
		fmt.Printf("   Testing: %s\n", tc.description)
		stats := profile.ParseHeapProfileForTesting(tc.profileData)
		
		fmt.Printf("      MemoryMB: %.1f, Goroutines: %d, Objects: %d\n",
			stats.AllocMB, stats.GoRoutines, stats.HeapObjects)
	}
}

func testMemoryLeakDetection() {
	// Simulate memory samples over time showing a potential leak
	samples := []profile.MemoryStats{
		{
			Timestamp:    time.Now().Add(-5 * time.Minute),
			AllocMB:      145.2,
			TotalAllocMB: 2100.0,
			SysMB:        234.5,
			GoRoutines:   1847,
			HeapObjects:  145672,
		},
		{
			Timestamp:    time.Now().Add(-4 * time.Minute),
			AllocMB:      167.8,
			TotalAllocMB: 2200.0,
			SysMB:        245.1,
			GoRoutines:   1923,
			HeapObjects:  167834,
		},
		{
			Timestamp:    time.Now().Add(-3 * time.Minute),
			AllocMB:      198.4,
			TotalAllocMB: 2400.0,
			SysMB:        267.2,
			GoRoutines:   2156,
			HeapObjects:  198423,
		},
		{
			Timestamp:    time.Now().Add(-2 * time.Minute),
			AllocMB:      234.7,
			TotalAllocMB: 2600.0,
			SysMB:        289.8,
			GoRoutines:   2387,
			HeapObjects:  234891,
		},
		{
			Timestamp:    time.Now().Add(-1 * time.Minute),
			AllocMB:      278.3,
			TotalAllocMB: 2900.0,
			SysMB:        312.4,
			GoRoutines:   2634,
			HeapObjects:  278456,
		},
	}

	fmt.Printf("   Analyzing %d memory samples...\n", len(samples))
	
	first := samples[0]
	last := samples[len(samples)-1]
	
	memoryGrowth := last.AllocMB - first.AllocMB
	goroutineGrowth := last.GoRoutines - first.GoRoutines
	objectGrowth := last.HeapObjects - first.HeapObjects
	duration := last.Timestamp.Sub(first.Timestamp)
	
	hourlyGrowthMB := memoryGrowth / duration.Hours()
	
	fmt.Printf("   📈 Memory Growth: %.1f MB over %v (%.1f MB/hour)\n", 
		memoryGrowth, duration, hourlyGrowthMB)
	fmt.Printf("   🔄 Goroutine Growth: %d\n", goroutineGrowth)
	fmt.Printf("   📦 Object Growth: %d\n", objectGrowth)
	
	// Leak detection thresholds
	if hourlyGrowthMB > 100.0 {
		fmt.Printf("   🚨 MEMORY LEAK DETECTED: High growth rate\n")
	} else {
		fmt.Printf("   ✅ Memory growth within normal range\n")
	}
	
	if goroutineGrowth > 500 {
		fmt.Printf("   🚨 GOROUTINE LEAK DETECTED: High goroutine growth\n")
	} else {
		fmt.Printf("   ✅ Goroutine growth within normal range\n")
	}
}

func testProfileOptions() {
	testCases := []struct {
		name string
		opts profile.ProfileOptions
		desc string
	}{
		{
			name: "basic_heap_profile",
			opts: profile.ProfileOptions{
				Alias:       "minio-prod",
				ProfileType: "heap",
				Duration:    30 * time.Second,
				MCPath:      "mc",
			},
			desc: "Basic heap profiling",
		},
		{
			name: "cpu_profile_with_output",
			opts: profile.ProfileOptions{
				Alias:       "minio-staging",
				ProfileType: "cpu",
				Duration:    60 * time.Second,
				Output:      "/tmp/cpu.pprof",
				MCPath:      "mc-2021",
				Verbose:     true,
			},
			desc: "CPU profiling with file output",
		},
		{
			name: "leak_detection",
			opts: profile.ProfileOptions{
				Alias:           "minio-dev",
				ProfileType:     "heap",
				Duration:        10 * time.Minute,
				DetectLeaks:     true,
				MonitorInterval: 15 * time.Second,
				ThresholdMB:     75,
				MCPath:          "mc",
				Verbose:         true,
			},
			desc: "Memory leak detection monitoring",
		},
	}

	for _, tc := range testCases {
		fmt.Printf("   %s:\n", tc.desc)
		fmt.Printf("      Alias: %s, Type: %s, Duration: %v\n",
			tc.opts.Alias, tc.opts.ProfileType, tc.opts.Duration)
		
		if tc.opts.DetectLeaks {
			fmt.Printf("      Leak Detection: enabled (threshold: %d MB, interval: %v)\n",
				tc.opts.ThresholdMB, tc.opts.MonitorInterval)
		}
		
		if tc.opts.Output != "" {
			fmt.Printf("      Output: %s\n", tc.opts.Output)
		}
		
		fmt.Printf("      MC Path: %s, Verbose: %v\n", tc.opts.MCPath, tc.opts.Verbose)
		fmt.Println()
	}
}

func testMCAdminProfile(alias string) {
	versions := profile.GetAvailableMCVersions()
	
	if len(versions) == 0 {
		fmt.Printf("   ❌ No MC versions available for testing\n")
		return
	}

	for _, mcPath := range versions {
		fmt.Printf("   Testing with %s...\n", mcPath)
		
		err := profile.TestMCAdminProfile(mcPath, alias)
		if err != nil {
			fmt.Printf("      ❌ Failed: %v\n", err)
			
			// Provide specific guidance based on error message
			if err.Error() == "Deprecated command" {
				fmt.Printf("      💡 Try older mc version: mc-2021\n")
			} else {
				fmt.Printf("      💡 Check alias configuration: mc alias list\n")
			}
		} else {
			fmt.Printf("      ✅ Success\n")
		}
	}
}

// Test data generation functions
func generateMemoryLeakScenario() []profile.MemoryStats {
	var samples []profile.MemoryStats
	baseTime := time.Now().Add(-30 * time.Minute)
	
	// Simulate a memory leak over 30 minutes
	for i := 0; i < 30; i++ {
		// Exponential memory growth simulating a leak
		leakFactor := 1.0 + (float64(i) * 0.02) // 2% growth per minute
		
		sample := profile.MemoryStats{
			Timestamp:    baseTime.Add(time.Duration(i) * time.Minute),
			AllocMB:      100.0 * leakFactor,
			TotalAllocMB: 1000.0 + (float64(i) * 50.0),
			SysMB:        200.0 + (float64(i) * 5.0),
			GoRoutines:   1000 + (i * 10), // Goroutine leak too
			HeapObjects:  100000 + (i * 5000),
		}
		
		samples = append(samples, sample)
	}
	
	return samples
}

func generateNormalScenario() []profile.MemoryStats {
	var samples []profile.MemoryStats
	baseTime := time.Now().Add(-30 * time.Minute)
	
	// Simulate normal memory usage with small variations
	for i := 0; i < 30; i++ {
		// Small random variations around baseline
		variation := float64(i%5 - 2) * 2.0 // -4 to +4 MB variation
		
		sample := profile.MemoryStats{
			Timestamp:    baseTime.Add(time.Duration(i) * time.Minute),
			AllocMB:      150.0 + variation,
			TotalAllocMB: 2000.0 + (float64(i) * 10.0),
			SysMB:        250.0 + (float64(i) * 1.0),
			GoRoutines:   1500 + (i % 10), // Stable goroutine count
			HeapObjects:  150000 + (i * 1000),
		}
		
		samples = append(samples, sample)
	}
	
	return samples
}

// Example usage demonstrations
func demonstrateProfileCommands() {
	fmt.Println("=== Example Profile Commands ===")
	
	examples := []struct {
		command     string
		description string
	}{
		{
			command:     "mc-tool profile heap minio-prod --duration 1m",
			description: "Basic heap profile for 1 minute",
		},
		{
			command:     "mc-tool profile cpu minio-prod --duration 60s --output /tmp/cpu.pprof",
			description: "CPU profile saved to file",
		},
		{
			command:     "mc-tool profile heap minio-prod --detect-leaks --duration 30m",
			description: "Memory leak detection for 30 minutes",
		},
		{
			command:     "mc-tool profile goroutine minio-prod --mc-path mc-2021",
			description: "Goroutine profile with older mc version",
		},
		{
			command:     "mc-tool profile allocs minio-prod --threshold-mb 200 --verbose",
			description: "Allocation profile with custom threshold",
		},
	}

	for _, example := range examples {
		fmt.Printf("# %s\n", example.description)
		fmt.Printf("%s\n\n", example.command)
	}
}

// Performance testing helper
func benchmarkProfileParsing() {
	fmt.Println("=== Performance Benchmarking ===")
	
	sampleData := `heap profile: 1024: 512 MB [2048: 1024 MB] @ heap/1048576
Total allocations: 2048 MB
System memory: 1024 MB
Number of goroutines: 150
goroutine 1 [running]:
runtime.main()
	/usr/local/go/src/runtime/proc.go:250 +0x207`

	start := time.Now()
	iterations := 10000
	
	for i := 0; i < iterations; i++ {
		profile.ParseHeapProfileForTesting(sampleData)
	}
	
	duration := time.Since(start)
	fmt.Printf("Parsed %d profiles in %v (%.2f μs per parse)\n", 
		iterations, duration, float64(duration.Microseconds())/float64(iterations))
}

// Main demonstration runner
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}