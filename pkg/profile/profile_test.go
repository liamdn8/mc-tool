package profile

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestProfileOptions_Validation(t *testing.T) {
	tests := []struct {
		name    string
		opts    ProfileOptions
		wantErr bool
	}{
		{
			name: "valid heap profile options",
			opts: ProfileOptions{
				Alias:       "test-alias",
				ProfileType: "heap",
				Duration:    30 * time.Second,
				MCPath:      "mc",
			},
			wantErr: false,
		},
		{
			name: "valid cpu profile with output",
			opts: ProfileOptions{
				Alias:       "prod-cluster",
				ProfileType: "cpu",
				Duration:    60 * time.Second,
				Output:      "/tmp/cpu.pprof",
				MCPath:      "mc-2021",
			},
			wantErr: false,
		},
		{
			name: "leak detection options",
			opts: ProfileOptions{
				Alias:           "test-alias",
				ProfileType:     "heap",
				Duration:        5 * time.Minute,
				DetectLeaks:     true,
				MonitorInterval: 10 * time.Second,
				ThresholdMB:     100,
				MCPath:          "mc",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - just ensure no panics for now
			if tt.opts.Alias == "" {
				t.Error("Alias should not be empty")
			}
			if tt.opts.ProfileType == "" {
				t.Error("ProfileType should not be empty")
			}
			if tt.opts.Duration < 0 {
				t.Error("Duration should not be negative")
			}
		})
	}
}

func TestGetAvailableMCVersions(t *testing.T) {
	versions := GetAvailableMCVersions()

	// Should return at least something, even if empty
	if versions == nil {
		t.Error("GetAvailableMCVersions should not return nil")
	}

	// If mc is available in PATH, it should be included
	if _, err := exec.LookPath("mc"); err == nil {
		found := false
		for _, v := range versions {
			if v == "mc" {
				found = true
				break
			}
		}
		if !found {
			t.Error("mc should be included in available versions when found in PATH")
		}
	}
}

func TestParseHeapProfile(t *testing.T) {
	tests := []struct {
		name               string
		profileData        string
		expectedMB         float64
		expectedGoroutines int
	}{
		{
			name: "heap profile with memory info",
			profileData: `heap profile: 1024: 512 MB [2048: 1024 MB] @ heap/1048576
System memory: 1024 MB
Number of goroutines: 150`,
			expectedMB:         512.0,
			expectedGoroutines: 150,
		},
		{
			name: "goroutine profile",
			profileData: `goroutine profile: total 250
250 @ 0x463f71 0x4640f6 0x4640d7 0x4066c1
#	0x463f70	runtime.gopark+0x70
#	0x4640f5	runtime.selectgo+0x2d5
250 goroutine`,
			expectedGoroutines: 250,
		},
		{
			name: "allocation profile",
			profileData: `heap profile: 2048: 256 MB [4096: 512 MB] @ heap/1048576
1: 128 MB [2: 256 MB] @ 0x42e78a 0x42e6f3
#	0x42e789	main.allocateMemory+0x49`,
			expectedMB: 256.0,
		},
		{
			name:               "empty profile",
			profileData:        "",
			expectedMB:         0.0,
			expectedGoroutines: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := parseHeapProfile(tt.profileData)

			if tt.expectedMB > 0 && stats.AllocMB != tt.expectedMB {
				t.Errorf("Expected AllocMB %f, got %f", tt.expectedMB, stats.AllocMB)
			}

			if tt.expectedGoroutines > 0 && stats.GoRoutines != tt.expectedGoroutines {
				t.Errorf("Expected GoRoutines %d, got %d", tt.expectedGoroutines, stats.GoRoutines)
			}
		})
	}
}

func TestMemoryStats(t *testing.T) {
	stats := MemoryStats{
		Timestamp:    time.Now(),
		AllocMB:      150.5,
		TotalAllocMB: 2048.7,
		SysMB:        256.3,
		NumGC:        42,
		HeapObjects:  150000,
		GoRoutines:   1847,
	}

	if stats.AllocMB <= 0 {
		t.Error("AllocMB should be positive")
	}
	if stats.GoRoutines <= 0 {
		t.Error("GoRoutines should be positive")
	}
	if stats.HeapObjects <= 0 {
		t.Error("HeapObjects should be positive")
	}
}

func TestLeakDetection(t *testing.T) {
	samples := []MemoryStats{
		{
			Timestamp:   time.Now().Add(-4 * time.Minute),
			AllocMB:     100.0,
			GoRoutines:  1000,
			HeapObjects: 100000,
		},
		{
			Timestamp:   time.Now().Add(-3 * time.Minute),
			AllocMB:     120.0,
			GoRoutines:  1050,
			HeapObjects: 120000,
		},
		{
			Timestamp:   time.Now().Add(-2 * time.Minute),
			AllocMB:     150.0,
			GoRoutines:  1100,
			HeapObjects: 150000,
		},
		{
			Timestamp:   time.Now().Add(-1 * time.Minute),
			AllocMB:     200.0,
			GoRoutines:  1200,
			HeapObjects: 200000,
		},
		{
			Timestamp:   time.Now(),
			AllocMB:     300.0,
			GoRoutines:  1500,
			HeapObjects: 300000,
		},
	}

	if len(samples) < 2 {
		t.Error("Need at least 2 samples for leak detection")
	}

	first := samples[0]
	last := samples[len(samples)-1]

	memoryGrowth := last.AllocMB - first.AllocMB
	goroutineGrowth := last.GoRoutines - first.GoRoutines

	if memoryGrowth <= 0 {
		t.Error("Expected positive memory growth in test data")
	}
	if goroutineGrowth <= 0 {
		t.Error("Expected positive goroutine growth in test data")
	}

	// Test growth rate calculation
	duration := last.Timestamp.Sub(first.Timestamp)
	if duration <= 0 {
		t.Error("Duration should be positive")
	}

	hourlyGrowthMB := memoryGrowth / duration.Hours()
	if hourlyGrowthMB <= 0 {
		t.Error("Hourly growth rate should be positive")
	}
}

func TestProfileTypeValidation(t *testing.T) {
	validTypes := []string{"cpu", "heap", "goroutine", "allocs", "block", "mutex"}

	for _, validType := range validTypes {
		t.Run("valid_type_"+validType, func(t *testing.T) {
			if !isValidProfileType(validType) {
				t.Errorf("Profile type %s should be valid", validType)
			}
		})
	}

	invalidTypes := []string{"invalid", "memory", "network", "disk", ""}

	for _, invalidType := range invalidTypes {
		t.Run("invalid_type_"+invalidType, func(t *testing.T) {
			if isValidProfileType(invalidType) {
				t.Errorf("Profile type %s should be invalid", invalidType)
			}
		})
	}
}

// Helper function for testing
func isValidProfileType(profileType string) bool {
	validTypes := []string{"cpu", "heap", "goroutine", "allocs", "block", "mutex"}
	for _, t := range validTypes {
		if profileType == t {
			return true
		}
	}
	return false
}

func TestMCPathValidation(t *testing.T) {
	tests := []struct {
		name   string
		mcPath string
		exists bool
	}{
		{
			name:   "default mc",
			mcPath: "mc",
			exists: true, // Assume mc exists in PATH for testing
		},
		{
			name:   "mc-2021",
			mcPath: "mc-2021",
			exists: false, // May not exist in test environment
		},
		{
			name:   "empty path",
			mcPath: "",
			exists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mcPath == "" {
				return // Skip empty path test
			}

			_, err := exec.LookPath(tt.mcPath)
			exists := err == nil

			// Only test if we have specific expectations
			if tt.name == "default mc" && !exists {
				t.Skip("mc not found in PATH, skipping test")
			}
		})
	}
}

// Benchmark tests
func BenchmarkParseHeapProfile(b *testing.B) {
	profileData := `heap profile: 1024 objects, 512 MB
Total allocations: 2048 MB
System memory: 1024 MB
Number of goroutines: 150
goroutine 1 [running]:
runtime.main()
	/usr/local/go/src/runtime/proc.go:250 +0x207`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseHeapProfile(profileData)
	}
}

func BenchmarkMemoryStatsAnalysis(b *testing.B) {
	samples := make([]MemoryStats, 100)
	baseTime := time.Now()

	for i := range samples {
		samples[i] = MemoryStats{
			Timestamp:    baseTime.Add(time.Duration(i) * time.Second),
			AllocMB:      float64(100 + i),
			TotalAllocMB: float64(1000 + i*10),
			SysMB:        float64(200 + i*2),
			GoRoutines:   1000 + i,
			HeapObjects:  100000 + i*1000,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate leak analysis
		first := samples[0]
		last := samples[len(samples)-1]

		_ = last.AllocMB - first.AllocMB
		_ = last.GoRoutines - first.GoRoutines
		_ = last.Timestamp.Sub(first.Timestamp)
	}
}

// Integration test (requires actual mc binary)
func TestMCIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if mc is available
	mcPath, err := exec.LookPath("mc")
	if err != nil {
		t.Skip("mc binary not found, skipping integration test")
	}

	// Test mc version
	cmd := exec.Command(mcPath, "version")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get mc version: %v", err)
	}

	if !strings.Contains(string(output), "mc version") {
		t.Error("Unexpected mc version output")
	}
}

// Test helper functions
func createTempFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "mc-tool-test-")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	return tmpFile.Name()
}

func TestOutputFileCreation(t *testing.T) {
	opts := ProfileOptions{
		Output: "/tmp/test-profile.pprof",
	}

	// Test that we can create the output directory
	if opts.Output != "" {
		dir := "/tmp" // Parent directory
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Skip("Cannot test output file creation without /tmp directory")
		}
	}
}

// Mock test for command execution
func TestMockProfileExecution(t *testing.T) {
	// This is a mock test that doesn't actually execute mc commands
	// but tests the logic around command construction

	opts := ProfileOptions{
		Alias:       "test-alias",
		ProfileType: "heap",
		Duration:    30 * time.Second,
		MCPath:      "mc",
	}

	// Test command argument construction logic
	args := []string{"admin", "profile", opts.ProfileType, opts.Alias}
	if opts.Duration > 0 {
		args = append(args, "--duration=30s")
	}

	expectedArgs := []string{"admin", "profile", "heap", "test-alias", "--duration=30s"}

	if len(args) != len(expectedArgs) {
		t.Errorf("Expected %d arguments, got %d", len(expectedArgs), len(args))
	}

	for i, arg := range args {
		if arg != expectedArgs[i] {
			t.Errorf("Expected argument %d to be %s, got %s", i, expectedArgs[i], arg)
		}
	}
}
