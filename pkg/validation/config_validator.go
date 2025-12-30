package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/liamdn8/mc-tool/pkg/config"
)

// ValidationResult represents the result of a configuration validation
type ValidationResult struct {
	Alias      string
	Configured bool
	RuleCount  int
	EventCount int
	Status     string // "match", "mismatch", "error"
	Message    string
	ConfigRaw  string
	Error      error
}

// BucketValidator handles bucket configuration validation
type BucketValidator struct {
	Bucket         string
	Aliases        []string
	ReferenceAlias string
	Verbose        bool
	Insecure       bool
	mcEnv          []string // Cached MC environment variables
}

// NewBucketValidator creates a new bucket validator
func NewBucketValidator(bucket string, aliases []string, verbose bool, insecure bool) *BucketValidator {
	// Load MC config and prepare environment variables
	var mcEnv []string
	mcConfig, err := config.LoadMCConfig()
	if err == nil {
		mcEnv = config.GetMCEnvironment(mcConfig)
	} else {
		mcEnv = os.Environ()
	}

	return &BucketValidator{
		Bucket:         bucket,
		Aliases:        aliases,
		ReferenceAlias: aliases[0], // First alias is reference
		Verbose:        verbose,
		Insecure:       insecure,
		mcEnv:          mcEnv,
	}
}

// setupMCCommand sets up an mc command with proper environment variables
func (v *BucketValidator) setupMCCommand(args ...string) *exec.Cmd {
	cmd := exec.Command("mc", args...)
	cmd.Env = v.mcEnv
	return cmd
}

// CheckBucketExists verifies if a bucket exists on an alias
func (v *BucketValidator) CheckBucketExists(alias string) bool {
	args := []string{"ls"}
	if v.Insecure {
		args = append(args, "--insecure")
	}
	args = append(args, fmt.Sprintf("%s/%s", alias, v.Bucket))
	cmd := v.setupMCCommand(args...)
	err := cmd.Run()
	return err == nil
}

// ValidateLifecycle validates lifecycle configuration across all aliases
func (v *BucketValidator) ValidateLifecycle() ([]ValidationResult, error) {
	results := make([]ValidationResult, 0, len(v.Aliases))

	// Get reference lifecycle
	refArgs := []string{"ilm", "ls"}
	if v.Insecure {
		refArgs = append(refArgs, "--insecure")
	}
	refArgs = append(refArgs, fmt.Sprintf("%s/%s", v.ReferenceAlias, v.Bucket), "--json")
	refCmd := v.setupMCCommand(refArgs...)
	refOutput, refErr := refCmd.CombinedOutput()

	var refRulesData []interface{}
	refResult := ValidationResult{
		Alias:      v.ReferenceAlias,
		Configured: false,
		Status:     "reference",
	}

	if refErr == nil {
		refRules := string(refOutput)
		refResult.ConfigRaw = refRules

		var refData map[string]interface{}
		if err := json.Unmarshal([]byte(refRules), &refData); err == nil {
			if status, ok := refData["status"].(string); ok && status == "success" {
				if config, ok := refData["config"].(map[string]interface{}); ok {
					if rules, ok := config["Rules"].([]interface{}); ok && len(rules) > 0 {
						refResult.Configured = true
						refResult.RuleCount = len(rules)
						refRulesData = rules
					}
				}
			}
		}
	} else {
		refResult.Error = refErr
	}

	results = append(results, refResult)

	// Compare with other aliases
	for _, alias := range v.Aliases {
		if alias == v.ReferenceAlias {
			continue
		}

		result := ValidationResult{
			Alias: alias,
		}

		args := []string{"ilm", "ls"}
		if v.Insecure {
			args = append(args, "--insecure")
		}
		args = append(args, fmt.Sprintf("%s/%s", alias, v.Bucket), "--json")
		cmd := v.setupMCCommand(args...)
		output, err := cmd.CombinedOutput()

		if err != nil {
			result.Status = "error"
			result.Error = err
			result.Message = fmt.Sprintf("Failed to get lifecycle: %v", err)
			results = append(results, result)
			continue
		}

		targetRules := string(output)
		result.ConfigRaw = targetRules

		var targetRulesData []interface{}
		var targetData map[string]interface{}
		if err := json.Unmarshal([]byte(targetRules), &targetData); err == nil {
			if status, ok := targetData["status"].(string); ok && status == "success" {
				if config, ok := targetData["config"].(map[string]interface{}); ok {
					if rules, ok := config["Rules"].([]interface{}); ok && len(rules) > 0 {
						result.Configured = true
						result.RuleCount = len(rules)
						targetRulesData = rules
					}
				}
			}
		}

		// Compare rules
		if compareLifecycleRules(refRulesData, targetRulesData) {
			result.Status = "match"
			result.Message = "Lifecycle configuration matches reference"
		} else {
			result.Status = "mismatch"
			if result.Configured {
				result.Message = "Lifecycle configuration differs from reference"
			} else {
				result.Message = "Not configured"
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// ValidateEvents validates event notification configuration across all aliases
func (v *BucketValidator) ValidateEvents() ([]ValidationResult, error) {
	results := make([]ValidationResult, 0, len(v.Aliases))

	// Get reference events
	refArgs := []string{"event", "list"}
	if v.Insecure {
		refArgs = append(refArgs, "--insecure")
	}
	refArgs = append(refArgs, fmt.Sprintf("%s/%s", v.ReferenceAlias, v.Bucket), "--json")
	refCmd := v.setupMCCommand(refArgs...)
	refOutput, refErr := refCmd.CombinedOutput()

	refResult := ValidationResult{
		Alias:      v.ReferenceAlias,
		Configured: false,
		Status:     "reference",
	}

	var refEventsRaw string
	if refErr == nil {
		refEventsRaw = string(refOutput)
		refResult.ConfigRaw = refEventsRaw
		refResult.Configured = strings.TrimSpace(refEventsRaw) != ""
		refResult.EventCount = countEvents(refEventsRaw)
	} else {
		refResult.Error = refErr
	}

	results = append(results, refResult)

	// Compare with other aliases
	for _, alias := range v.Aliases {
		if alias == v.ReferenceAlias {
			continue
		}

		result := ValidationResult{
			Alias: alias,
		}

		args := []string{"event", "list"}
		if v.Insecure {
			args = append(args, "--insecure")
		}
		args = append(args, fmt.Sprintf("%s/%s", alias, v.Bucket), "--json")
		cmd := v.setupMCCommand(args...)
		output, err := cmd.CombinedOutput()

		if err != nil {
			result.Status = "error"
			result.Error = err
			result.Message = fmt.Sprintf("Failed to get events: %v", err)
			results = append(results, result)
			continue
		}

		targetEvents := string(output)
		result.ConfigRaw = targetEvents
		result.Configured = strings.TrimSpace(targetEvents) != ""
		result.EventCount = countEvents(targetEvents)

		// Compare events
		if targetEvents == refEventsRaw {
			result.Status = "match"
			result.Message = "Event configuration matches reference"
		} else {
			result.Status = "mismatch"
			if result.Configured {
				result.Message = "Event configuration differs from reference"
			} else {
				result.Message = "Not configured"
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// compareLifecycleRules compares lifecycle rules ignoring IDs
func compareLifecycleRules(refRules, targetRules []interface{}) bool {
	if len(refRules) != len(targetRules) {
		return false
	}

	if len(refRules) == 0 {
		return true
	}

	normalizeRules := func(rules []interface{}) []interface{} {
		normalized := make([]interface{}, len(rules))
		for i, rule := range rules {
			if ruleMap, ok := rule.(map[string]interface{}); ok {
				normalizedRule := make(map[string]interface{})
				for k, v := range ruleMap {
					if k != "ID" {
						normalizedRule[k] = v
					}
				}
				normalized[i] = normalizedRule
			} else {
				normalized[i] = rule
			}
		}
		return normalized
	}

	refNorm, _ := json.Marshal(normalizeRules(refRules))
	targetNorm, _ := json.Marshal(normalizeRules(targetRules))

	return string(refNorm) == string(targetNorm)
}

// countEvents counts the number of event notifications
func countEvents(eventsOutput string) int {
	count := 0
	lines := strings.Split(strings.TrimSpace(eventsOutput), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			var eventData map[string]interface{}
			if json.Unmarshal([]byte(line), &eventData) == nil {
				if status, ok := eventData["status"].(string); ok && status == "success" {
					count++
				}
			}
		}
	}
	return count
}
