package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/liamdn8/mc-tool/pkg/config"
)

// EnvVarResult represents environment variable validation result
type EnvVarResult struct {
	Alias        string            `json:"alias"`
	Version      string            `json:"version"`
	CommitID     string            `json:"commitID"`
	EnvVars      map[string]string `json:"envVars"`
	FilteredVars map[string]string `json:"filteredVars"`
	Status       string            `json:"status"`
	Error        string            `json:"error,omitempty"`
}

// InfoValidator handles MinIO server information validation
type InfoValidator struct {
	Aliases  []string
	Insecure bool
	mcEnv    []string // Cached MC environment variables
}

// NewInfoValidator creates a new info validator
func NewInfoValidator(aliases []string, insecure bool) *InfoValidator {
	// Load MC config and prepare environment variables
	var mcEnv []string
	mcConfig, err := config.LoadMCConfig()
	if err == nil {
		mcEnv = config.GetMCEnvironment(mcConfig)
	} else {
		mcEnv = os.Environ()
	}

	return &InfoValidator{
		Aliases:  aliases,
		Insecure: insecure,
		mcEnv:    mcEnv,
	}
}

// setupMCCommand sets up an mc command with proper environment variables
func (v *InfoValidator) setupMCCommand(args ...string) *exec.Cmd {
	cmd := exec.Command("mc", args...)
	cmd.Env = v.mcEnv
	return cmd
}

// ValidateEnvironmentVariables validates MinIO environment variables across aliases
func (v *InfoValidator) ValidateEnvironmentVariables() ([]EnvVarResult, error) {
	results := make([]EnvVarResult, 0, len(v.Aliases))

	for _, alias := range v.Aliases {
		result := EnvVarResult{
			Alias:        alias,
			Status:       "success",
			EnvVars:      make(map[string]string),
			FilteredVars: make(map[string]string),
		}

		// Run mc admin info --json
		args := []string{"admin", "info", alias, "--json"}
		if v.Insecure {
			args = append(args, "--insecure")
		}

		cmd := v.setupMCCommand(args...)
		output, err := cmd.CombinedOutput()

		if err != nil {
			result.Status = "error"
			result.Error = fmt.Sprintf("Failed to get admin info: %v", err)
			results = append(results, result)
			continue
		}

		// Parse JSON output
		var adminInfo map[string]interface{}
		if err := json.Unmarshal(output, &adminInfo); err != nil {
			result.Status = "error"
			result.Error = fmt.Sprintf("Failed to parse admin info: %v", err)
			results = append(results, result)
			continue
		}

		// Extract server info
		info, ok := adminInfo["info"].(map[string]interface{})
		if !ok {
			result.Status = "error"
			result.Error = "Invalid admin info structure"
			results = append(results, result)
			continue
		}

		servers, ok := info["servers"].([]interface{})
		if !ok || len(servers) == 0 {
			result.Status = "error"
			result.Error = "No server information found"
			results = append(results, result)
			continue
		}

		server := servers[0].(map[string]interface{})

		// Extract version
		if version, ok := server["version"].(string); ok {
			result.Version = version
		}

		// Extract commitID
		if commitID, ok := server["commitID"].(string); ok {
			result.CommitID = commitID
		}

		// Extract environment variables
		if envVars, ok := server["minio_env_vars"].(map[string]interface{}); ok {
			for key, value := range envVars {
				if strValue, ok := value.(string); ok {
					result.EnvVars[key] = strValue

					// Filter variables - exclude those matching ignore patterns
					if !shouldIgnoreEnvVar(key) {
						result.FilteredVars[key] = strValue
					}
				}
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// shouldIgnoreEnvVar checks if an environment variable should be ignored based on patterns
func shouldIgnoreEnvVar(key string) bool {
	ignorePatterns := []string{
		"_PORT",
		"_SERVICE",
		"PORT_",
		"SERVICE_",
		"MINIO_BROWSER_REDIRECT_URL",
		"MINIO_SERVER_URL",
	}

	keyUpper := strings.ToUpper(key)
	for _, pattern := range ignorePatterns {
		if strings.Contains(keyUpper, pattern) {
			return true
		}
	}
	return false
}

// ValidateEnvironmentVariables is a standalone function for backward compatibility
func ValidateEnvironmentVariables(aliases []string, insecure bool) ([]EnvVarResult, error) {
	validator := NewInfoValidator(aliases, insecure)
	return validator.ValidateEnvironmentVariables()
}
