package debug

import (
	"strings"
	"testing"
	"time"
)

func TestParseGoroutineProfile(t *testing.T) {
	// Sample pprof goroutine output
	sampleData := `goroutine 1 [running]:
main.main()
	/app/main.go:10 +0x25

goroutine 2 [syscall, 2 minutes]:
runtime.gopark(0x0, 0x0, 0x0, 0x0, 0x0)
	/usr/local/go/src/runtime/proc.go:363 +0xd6
syscall.Syscall(0x0, 0x0, 0x0, 0x0)
	/usr/local/go/src/syscall/syscall_linux.go:191 +0x4a

goroutine 3 [chan receive]:
main.worker()
	/app/worker.go:15 +0x89

goroutine 4 [select]:
runtime.gopark(0x0, 0x0, 0x0, 0x0, 0x0)
	/usr/local/go/src/runtime/proc.go:363 +0xd6
`

	stats, err := parseGoroutineProfile(sampleData, "http://localhost:9000")
	if err != nil {
		t.Fatalf("Failed to parse goroutine profile: %v", err)
	}

	// Check basic stats
	if stats.Total != 4 {
		t.Errorf("Expected 4 goroutines, got %d", stats.Total)
	}

	// Check state breakdown
	if stats.ByState["running"] != 1 {
		t.Errorf("Expected 1 running goroutine, got %d", stats.ByState["running"])
	}

	if stats.ByState["syscall"] != 1 {
		t.Errorf("Expected 1 syscall goroutine, got %d", stats.ByState["syscall"])
	}

	// Check long-running detection
	if len(stats.LongRunning) != 1 {
		t.Errorf("Expected 1 long-running goroutine, got %d", len(stats.LongRunning))
	}

	// Check potential leaks
	if len(stats.PotentialLeaks) == 0 {
		t.Error("Expected to detect potential leaks")
	}
}

func TestIsLikelyLeak(t *testing.T) {
	testCases := []struct {
		name     string
		goroutine GoroutineInfo
		expected bool
	}{
		{
			name: "Goroutine with gopark",
			goroutine: GoroutineInfo{
				Stack: "runtime.gopark(0x0, 0x0, 0x0, 0x0, 0x0)",
			},
			expected: true,
		},
		{
			name: "Normal goroutine",
			goroutine: GoroutineInfo{
				Stack: "main.process()",
			},
			expected: false,
		},
		{
			name: "Time.Sleep goroutine",
			goroutine: GoroutineInfo{
				Stack: "time.Sleep(1000000000)",
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isLikelyLeak(&tc.goroutine)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestGetTopFunctions(t *testing.T) {
	funcCounts := map[string]int{
		"main.main":    10,
		"runtime.main": 5,
		"net.dial":     3,
		"io.copy":      7,
		"http.serve":   2,
	}

	top := getTopFunctions(funcCounts, 3)

	if len(top) != 3 {
		t.Errorf("Expected 3 functions, got %d", len(top))
	}

	// Should be sorted by count descending
	if top[0].Function != "main.main" || top[0].Count != 10 {
		t.Errorf("Expected main.main with count 10, got %s with count %d", top[0].Function, top[0].Count)
	}

	if top[1].Function != "io.copy" || top[1].Count != 7 {
		t.Errorf("Expected io.copy with count 7, got %s with count %d", top[1].Function, top[1].Count)
	}
}

func TestDebugOptions(t *testing.T) {
	opts := DebugOptions{
		Endpoint:        "http://localhost:9000",
		AccessKey:       "minioadmin",
		SecretKey:       "minioadmin",
		MonitorDuration: 5 * time.Minute,
		Interval:        10 * time.Second,
		ThresholdCount:  50,
		OutputFormat:    "json",
	}

	if opts.Endpoint != "http://localhost:9000" {
		t.Errorf("Unexpected endpoint: %s", opts.Endpoint)
	}

	if opts.MonitorDuration != 5*time.Minute {
		t.Errorf("Unexpected monitor duration: %v", opts.MonitorDuration)
	}
}