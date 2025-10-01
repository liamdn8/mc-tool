package main

import (
	"os/exec"
	"strings"
	"testing"
)

// Integration tests for command-line interface
// These tests verify that the CLI commands work correctly

func TestCommandsExist(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "help shows all commands",
			args:     []string{"--help"},
			expected: "checklist",
		},
		{
			name:     "checklist command exists",
			args:     []string{"checklist", "--help"},
			expected: "comprehensive validation of MinIO bucket configuration",
		},
		{
			name:     "analyze command exists without config-check",
			args:     []string{"analyze", "--help"},
			expected: "object distribution, versions, and incomplete uploads",
		},
		{
			name:     "compare command still works",
			args:     []string{"compare", "--help"},
			expected: "Compare objects between two MinIO buckets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./mc-tool", tt.args...)
			output, err := cmd.CombinedOutput()
			
			// For help commands, exit code 0 is expected
			if err != nil && !strings.Contains(string(output), tt.expected) {
				t.Fatalf("Command failed: %v, output: %s", err, output)
			}
			
			if !strings.Contains(string(output), tt.expected) {
				t.Errorf("Expected output to contain %q, got: %s", tt.expected, output)
			}
		})
	}
}

func TestChecklistCommandValidation(t *testing.T) {
	// Test that checklist command validates arguments
	cmd := exec.Command("./mc-tool", "checklist")
	output, err := cmd.CombinedOutput()
	
	if err == nil {
		t.Error("Expected checklist command to fail without arguments")
	}
	
	if !strings.Contains(string(output), "accepts 1 arg(s), received 0") {
		t.Errorf("Expected argument validation error, got: %s", output)
	}
}

func TestAnalyzeNoConfigCheck(t *testing.T) {
	// Verify that analyze command help doesn't mention config-check
	cmd := exec.Command("./mc-tool", "analyze", "--help")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		t.Fatalf("Analyze help command failed: %v", err)
	}
	
	if strings.Contains(string(output), "config-check") {
		t.Error("Analyze command should not have config-check flag anymore")
	}
}