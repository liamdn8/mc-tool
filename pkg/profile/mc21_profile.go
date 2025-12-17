package profile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// MC21ProfileOptions contains options for mc21 admin profile workflow
type MC21ProfileOptions struct {
	Alias       string
	ProfileType string // Comma-separated types: cpu,mem,block,mutex,trace,threads,goroutines
	Duration    time.Duration
	Output      string // Output directory (will create timestamped subdirectory)
	Verbose     bool
	MCPath      string // Path to mc21 binary
	Insecure    bool   // Skip TLS verification
}

// MC21ProfileResult contains the results of profiling
type MC21ProfileResult struct {
	OutputDir  string
	CPUFiles   []string
	MemFiles   []string
	OtherFiles []string
	Success    bool
	Error      error
}

// RunMC21Profile executes the complete mc21 admin profile workflow
func RunMC21Profile(opts MC21ProfileOptions) (*MC21ProfileResult, error) {
	result := &MC21ProfileResult{
		Success: false,
	}

	// Create output directory with timestamp
	timestamp := time.Now().Format("20060102-150405")
	var outputDir string
	if opts.Output != "" {
		outputDir = filepath.Join(opts.Output, fmt.Sprintf("profile-%s", timestamp))
	} else {
		outputDir = filepath.Join("/tmp", fmt.Sprintf("profile-%s", timestamp))
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}
	result.OutputDir = outputDir

	if opts.Verbose {
		fmt.Printf("╔══════════════════════════════════════════════════════════════╗\n")
		fmt.Printf("║  MinIO Profiling                                             ║\n")
		fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n\n")
		fmt.Printf("🔧 Alias: %s\n", opts.Alias)
		fmt.Printf("📊 Profile types: %s\n", opts.ProfileType)
		fmt.Printf("⏱️  Duration: %s\n", opts.Duration)
		fmt.Printf("🔨 MC Binary: %s\n", opts.MCPath)
		fmt.Printf("📁 Output directory: %s\n", outputDir)
		if opts.Insecure {
			fmt.Printf("🔓 Insecure mode: enabled\n")
		}
		fmt.Println()
	}

	// Step 1: Start profiling
	if err := startProfiling(opts); err != nil {
		result.Error = err
		return result, err
	}

	// Step 2: Wait for duration
	if opts.Verbose {
		fmt.Printf("⏳ Waiting %s for profile data collection...\n", opts.Duration)
	}
	time.Sleep(opts.Duration)
	if opts.Verbose {
		fmt.Println()
	}

	// Step 3: Stop profiling and download profile.zip
	if err := stopProfiling(opts, outputDir); err != nil {
		result.Error = err
		return result, err
	}

	// Step 4: Extract profile.zip
	if err := extractProfile(outputDir, opts.Verbose); err != nil {
		result.Error = err
		return result, err
	}

	// Step 5: Find and categorize profile files
	if err := categorizeProfileFiles(result); err != nil {
		result.Error = err
		return result, err
	}

	result.Success = true
	return result, nil
}

func startProfiling(opts MC21ProfileOptions) error {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("▶️  Starting profile collection...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	startArgs := []string{"--config-dir", filepath.Join(os.Getenv("HOME"), ".mc"), "admin", "profile", "start"}
	if opts.Insecure {
		startArgs = append(startArgs, "--insecure")
	}
	startArgs = append(startArgs, "--type", opts.ProfileType, opts.Alias)

	if opts.Verbose {
		fmt.Printf("$ %s %s\n", opts.MCPath, strings.Join(startArgs, " "))
	}

	startCmd := exec.Command(opts.MCPath, startArgs...)
	startOutput, err := startCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start profiling: %v\nOutput: %s", err, string(startOutput))
	}

	fmt.Printf("✅ Profiling started\n")
	if opts.Verbose {
		fmt.Printf("%s\n", strings.TrimSpace(string(startOutput)))
	}
	fmt.Println()

	return nil
}

func stopProfiling(opts MC21ProfileOptions, outputDir string) error {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("⏹️  Stopping profile and downloading data...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Change to output directory to download profile.zip there
	originalDir, _ := os.Getwd()
	if err := os.Chdir(outputDir); err != nil {
		return fmt.Errorf("failed to change to output directory: %v", err)
	}
	defer os.Chdir(originalDir)

	stopArgs := []string{"--config-dir", filepath.Join(os.Getenv("HOME"), ".mc"), "admin", "profile", "stop"}
	if opts.Insecure {
		stopArgs = append(stopArgs, "--insecure")
	}
	stopArgs = append(stopArgs, opts.Alias)

	if opts.Verbose {
		fmt.Printf("$ cd %s\n", outputDir)
		fmt.Printf("$ %s %s\n", opts.MCPath, strings.Join(stopArgs, " "))
	}

	stopCmd := exec.Command(opts.MCPath, stopArgs...)
	stopOutput, err := stopCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop profiling: %v\nOutput: %s", err, string(stopOutput))
	}

	fmt.Printf("✅ Profile data downloaded\n")
	if opts.Verbose {
		fmt.Printf("%s\n", strings.TrimSpace(string(stopOutput)))
	}
	fmt.Println()

	return nil
}

func extractProfile(outputDir string, verbose bool) error {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📦 Extracting profile.zip...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	profileZip := filepath.Join(outputDir, "profile.zip")
	if _, err := os.Stat(profileZip); os.IsNotExist(err) {
		return fmt.Errorf("profile.zip not found in %s", outputDir)
	}

	unzipCmd := exec.Command("unzip", "-o", "profile.zip")
	unzipCmd.Dir = outputDir
	unzipOutput, err := unzipCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to extract profile.zip: %v\nOutput: %s", err, string(unzipOutput))
	}

	fmt.Printf("✅ Profile data extracted\n")
	if verbose {
		fmt.Printf("%s\n", strings.TrimSpace(string(unzipOutput)))
	}
	fmt.Println()

	return nil
}

func categorizeProfileFiles(result *MC21ProfileResult) error {
	// Find all .pprof files
	files, err := filepath.Glob(filepath.Join(result.OutputDir, "*.pprof"))
	if err != nil {
		return fmt.Errorf("failed to list profile files: %v", err)
	}

	for _, file := range files {
		basename := filepath.Base(file)
		if strings.Contains(basename, "-cpu.pprof") {
			result.CPUFiles = append(result.CPUFiles, file)
		} else if strings.Contains(basename, "-mem.pprof") && !strings.Contains(basename, "-before") {
			result.MemFiles = append(result.MemFiles, file)
		} else if !strings.Contains(basename, "-before") {
			result.OtherFiles = append(result.OtherFiles, file)
		}
	}

	return nil
}

// PrintProfileCommands displays useful pprof commands for analyzing the profiles
func PrintProfileCommands(result *MC21ProfileResult) {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📊 Profile Data Ready")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📁 Location: %s\n\n", result.OutputDir)

	// List all profile files
	allFiles := append(result.CPUFiles, result.MemFiles...)
	allFiles = append(allFiles, result.OtherFiles...)

	if len(allFiles) == 0 {
		fmt.Println("⚠️  No .pprof files found")
		return
	}

	fmt.Println("📋 Available profile files:")
	fmt.Println()

	for _, file := range allFiles {
		basename := filepath.Base(file)
		fmt.Printf("  • %s\n", basename)

		// Provide appropriate pprof command
		var profileTypeDetected string
		if strings.Contains(basename, "-cpu.pprof") {
			profileTypeDetected = "CPU"
		} else if strings.Contains(basename, "-mem.pprof") || strings.Contains(basename, "-heap.pprof") {
			profileTypeDetected = "Memory (Heap)"
		} else if strings.Contains(basename, "-block.pprof") {
			profileTypeDetected = "Block"
		} else if strings.Contains(basename, "-mutex.pprof") {
			profileTypeDetected = "Mutex"
		} else if strings.Contains(basename, "-goroutines.pprof") {
			profileTypeDetected = "Goroutines"
		}

		if profileTypeDetected != "" {
			fmt.Printf("    Type: %s\n", profileTypeDetected)
		}
	}

	fmt.Println()
	fmt.Println("🔍 To analyze profiles, use these commands:")
	fmt.Println()

	// CPU profiles
	if len(result.CPUFiles) > 0 {
		printCPUCommands(result.CPUFiles)
	}

	// Memory profiles
	if len(result.MemFiles) > 0 {
		printMemoryCommands(result.MemFiles, result.OutputDir)
	}

	// Other profiles
	if len(result.OtherFiles) > 0 {
		printOtherCommands(result.OtherFiles)
	}

	// Additional commands
	printAdditionalCommands(result.OutputDir, result.CPUFiles)

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("✅ Profiling completed successfully!\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func printCPUCommands(cpuFiles []string) {
	fmt.Println("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  🔥 CPU Profile Analysis")
	fmt.Println("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	for _, file := range cpuFiles {
		fmt.Println()
		fmt.Printf("  # Web UI (recommended):\n")
		fmt.Printf("  go tool pprof -http=:8080 %s\n\n", file)
		fmt.Printf("  # Top 10 CPU-intensive functions:\n")
		fmt.Printf("  go tool pprof -top -nodecount=10 %s\n\n", file)
		fmt.Printf("  # Flamegraph SVG:\n")
		fmt.Printf("  go tool pprof -svg %s > cpu-flamegraph.svg\n", file)
	}
	fmt.Println()
}

func printMemoryCommands(memFiles []string, outputDir string) {
	fmt.Println("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  💾 Memory Profile Analysis")
	fmt.Println("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	for _, file := range memFiles {
		fmt.Println()
		fmt.Printf("  # Web UI (recommended):\n")
		fmt.Printf("  go tool pprof -http=:8080 %s\n\n", file)
		fmt.Printf("  # Top 10 memory allocations:\n")
		fmt.Printf("  go tool pprof -top -nodecount=10 -alloc_space %s\n\n", file)
		fmt.Printf("  # Top 10 in-use memory:\n")
		fmt.Printf("  go tool pprof -top -nodecount=10 -inuse_space %s\n", file)
	}
	fmt.Println()
}

func printOtherCommands(otherFiles []string) {
	fmt.Println("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  📊 Other Profile Analysis")
	fmt.Println("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	for _, file := range otherFiles {
		basename := filepath.Base(file)
		fmt.Println()

		if strings.Contains(basename, "-block.pprof") {
			fmt.Printf("  # Block - Web UI:\n")
			fmt.Printf("  go tool pprof -http=:8080 %s\n\n", file)
			fmt.Printf("  # Block - Top 10:\n")
			fmt.Printf("  go tool pprof -top -nodecount=10 %s\n", file)
		} else if strings.Contains(basename, "-mutex.pprof") {
			fmt.Printf("  # Mutex - Web UI:\n")
			fmt.Printf("  go tool pprof -http=:8080 %s\n\n", file)
			fmt.Printf("  # Mutex - Top 10:\n")
			fmt.Printf("  go tool pprof -top -nodecount=10 %s\n", file)
		} else if strings.Contains(basename, "-goroutines.pprof") {
			fmt.Printf("  # Goroutines - Web UI:\n")
			fmt.Printf("  go tool pprof -http=:8080 %s\n\n", file)
			fmt.Printf("  # Goroutines - Top 10:\n")
			fmt.Printf("  go tool pprof -top -nodecount=10 %s\n", file)
		} else {
			fmt.Printf("  # Web UI:\n")
			fmt.Printf("  go tool pprof -http=:8080 %s\n\n", file)
			fmt.Printf("  # Top functions:\n")
			fmt.Printf("  go tool pprof -top %s\n", file)
		}
		fmt.Println()
	}
}

func printAdditionalCommands(outputDir string, cpuFiles []string) {
	fmt.Println("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  🛠️  Additional")
	fmt.Println("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Printf("  # View cluster info:\n")
	fmt.Printf("  cat %s/cluster.info\n\n", outputDir)
	fmt.Printf("  # List all profiles:\n")
	fmt.Printf("  ls -lh %s/*.pprof\n", outputDir)
	fmt.Println()
}
