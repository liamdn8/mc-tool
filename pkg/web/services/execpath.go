package services

import (
	"os"
	"os/exec"
	"path/filepath"
)

// FindMCToolExecutable finds the mc-tool executable path
// It checks in the following order:
// 1. Current executable path (if running as mc-tool)
// 2. ./mc-tool (current directory)
// 3. mc-tool in PATH
// 4. /usr/local/bin/mc-tool (Docker/system install)
func FindMCToolExecutable() string {
	// 1. Try current executable path
	if execPath, err := os.Executable(); err == nil {
		// Resolve symlinks
		if realPath, err := filepath.EvalSymlinks(execPath); err == nil {
			execPath = realPath
		}
		// Check if it's mc-tool itself
		if filepath.Base(execPath) == "mc-tool" || filepath.Base(execPath) == "mc-tool.exe" {
			return execPath
		}
	}

	// 2. Try current directory
	if _, err := os.Stat("./mc-tool"); err == nil {
		if absPath, err := filepath.Abs("./mc-tool"); err == nil {
			return absPath
		}
		return "./mc-tool"
	}

	// 3. Try PATH lookup
	if path, err := exec.LookPath("mc-tool"); err == nil {
		if absPath, err := filepath.Abs(path); err == nil {
			return absPath
		}
		return path
	}

	// 4. Try common installation paths
	commonPaths := []string{
		"/usr/local/bin/mc-tool",
		"/usr/bin/mc-tool",
		"/bin/mc-tool",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Fallback to "mc-tool" and let PATH resolve it
	return "mc-tool"
}
