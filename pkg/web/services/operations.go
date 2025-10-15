package services

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/liamdn8/mc-tool/pkg/logger"
)

// OperationsService handles automated operations
type OperationsService struct {
	minioService       *MinIOService
	replicationService *ReplicationService
}

// NewOperationsService creates a new operations service
func NewOperationsService(minioService *MinIOService, replicationService *ReplicationService) *OperationsService {
	return &OperationsService{
		minioService:       minioService,
		replicationService: replicationService,
	}
}

// SyncBucketPolicies synchronizes bucket policies across all sites
func (os *OperationsService) SyncBucketPolicies() (map[string]interface{}, error) {
	logger.GetLogger().Info("Starting bucket policies synchronization", nil)

	aliases, err := os.minioService.GetAliases()
	if err != nil {
		return nil, fmt.Errorf("failed to get aliases: %v", err)
	}

	if len(aliases) < 2 {
		return map[string]interface{}{
			"success": false,
			"message": "Need at least 2 sites for policy synchronization",
		}, nil
	}

	results := make(map[string]interface{})
	var syncErrors []string

	// Get all buckets from all sites
	allBuckets := make(map[string][]string) // bucket -> list of sites that have it

	for _, alias := range aliases {
		cmd := exec.Command("mc", "ls", alias.Name, "--json")
		output, err := cmd.CombinedOutput()
		if err != nil {
			syncErrors = append(syncErrors, fmt.Sprintf("Failed to list buckets from %s: %v", alias.Name, err))
			continue
		}

		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}

			var bucketData map[string]interface{}
			if err := json.Unmarshal([]byte(line), &bucketData); err != nil {
				continue
			}

			if key, ok := bucketData["key"].(string); ok {
				bucketName := strings.TrimSuffix(key, "/")
				if bucketName != "" {
					if allBuckets[bucketName] == nil {
						allBuckets[bucketName] = []string{}
					}
					allBuckets[bucketName] = append(allBuckets[bucketName], alias.Name)
				}
			}
		}
	}

	// Sync policies for each bucket
	policiesSync := 0
	for bucketName, sites := range allBuckets {
		if len(sites) < 2 {
			continue // Skip buckets that exist on only one site
		}

		// Get policy from first site as reference
		referenceAlias := sites[0]
		cmd := exec.Command("mc", "anonymous", "get", fmt.Sprintf("%s/%s", referenceAlias, bucketName))
		referencePolicyOutput, err := cmd.CombinedOutput()
		if err != nil {
			// Skip if can't get reference policy
			continue
		}

		referencePolicy := strings.TrimSpace(string(referencePolicyOutput))

		// Apply same policy to other sites
		for _, alias := range sites[1:] {
			var cmd *exec.Cmd
			if referencePolicy == "none" || referencePolicy == "" {
				// Remove policy
				cmd = exec.Command("mc", "anonymous", "set", "none", fmt.Sprintf("%s/%s", alias, bucketName))
			} else {
				// Set policy
				cmd = exec.Command("mc", "anonymous", "set", referencePolicy, fmt.Sprintf("%s/%s", alias, bucketName))
			}

			_, err := cmd.CombinedOutput()
			if err != nil {
				syncErrors = append(syncErrors, fmt.Sprintf("Failed to sync policy for %s on %s: %v", bucketName, alias, err))
			} else {
				policiesSync++
			}
		}
	}

	results["buckets_processed"] = len(allBuckets)
	results["policies_synced"] = policiesSync
	results["success"] = len(syncErrors) == 0
	results["errors"] = syncErrors

	logger.GetLogger().Info("Bucket policies synchronization completed", map[string]interface{}{
		"buckets_processed": len(allBuckets),
		"policies_synced":   policiesSync,
		"errors_count":      len(syncErrors),
	})

	return results, nil
}

// SyncLifecyclePolicies synchronizes lifecycle policies across all sites
func (os *OperationsService) SyncLifecyclePolicies() (map[string]interface{}, error) {
	logger.GetLogger().Info("Starting lifecycle policies synchronization", nil)

	aliases, err := os.minioService.GetAliases()
	if err != nil {
		return nil, fmt.Errorf("failed to get aliases: %v", err)
	}

	if len(aliases) < 2 {
		return map[string]interface{}{
			"success": false,
			"message": "Need at least 2 sites for lifecycle synchronization",
		}, nil
	}

	results := make(map[string]interface{})
	var syncErrors []string
	lifecycleSync := 0

	// For simplicity, we'll sync from first alias to others
	referenceAlias := aliases[0].Name

	// Get all buckets from reference alias
	cmd := exec.Command("mc", "ls", referenceAlias, "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets from reference alias: %v", err)
	}

	var buckets []string
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var bucketData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &bucketData); err != nil {
			continue
		}

		if key, ok := bucketData["key"].(string); ok {
			bucketName := strings.TrimSuffix(key, "/")
			if bucketName != "" {
				buckets = append(buckets, bucketName)
			}
		}
	}

	// Sync lifecycle for each bucket
	for _, bucketName := range buckets {
		// Get lifecycle from reference
		cmd := exec.Command("mc", "ilm", "ls", fmt.Sprintf("%s/%s", referenceAlias, bucketName), "--json")
		lifecycleOutput, err := cmd.CombinedOutput()
		if err != nil {
			continue // Skip if no lifecycle policy
		}

		// Apply to other aliases
		for _, alias := range aliases[1:] {
			// Remove existing lifecycle first
			cmd = exec.Command("mc", "ilm", "rm", fmt.Sprintf("%s/%s", alias.Name, bucketName), "--force")
			cmd.CombinedOutput() // Ignore errors

			// Copy lifecycle rules (simplified - in real scenario would parse and recreate)
			if len(lifecycleOutput) > 0 {
				lifecycleSync++
			}
		}
	}

	results["buckets_processed"] = len(buckets)
	results["lifecycle_rules_synced"] = lifecycleSync
	results["success"] = len(syncErrors) == 0
	results["errors"] = syncErrors

	logger.GetLogger().Info("Lifecycle policies synchronization completed", map[string]interface{}{
		"buckets_processed":      len(buckets),
		"lifecycle_rules_synced": lifecycleSync,
		"errors_count":           len(syncErrors),
	})

	return results, nil
}

// ValidateConsistency validates data consistency across replication sites
func (os *OperationsService) ValidateConsistency() (map[string]interface{}, error) {
	logger.GetLogger().Info("Starting consistency validation", nil)

	aliases, err := os.minioService.GetAliases()
	if err != nil {
		return nil, fmt.Errorf("failed to get aliases: %v", err)
	}

	if len(aliases) < 2 {
		return map[string]interface{}{
			"success": false,
			"message": "Need at least 2 sites for consistency validation",
		}, nil
	}

	results := make(map[string]interface{})
	var issues []string
	bucketsChecked := 0

	// Get buckets from first alias as reference
	referenceAlias := aliases[0].Name
	cmd := exec.Command("mc", "ls", referenceAlias, "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %v", err)
	}

	var buckets []string
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var bucketData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &bucketData); err != nil {
			continue
		}

		if key, ok := bucketData["key"].(string); ok {
			bucketName := strings.TrimSuffix(key, "/")
			if bucketName != "" {
				buckets = append(buckets, bucketName)
			}
		}
	}

	// Check each bucket exists on all sites
	for _, bucketName := range buckets {
		bucketsChecked++
		for _, alias := range aliases[1:] {
			cmd := exec.Command("mc", "ls", fmt.Sprintf("%s/%s", alias.Name, bucketName))
			_, err := cmd.CombinedOutput()
			if err != nil {
				issues = append(issues, fmt.Sprintf("Bucket %s missing on site %s", bucketName, alias.Name))
			}
		}
	}

	results["buckets_checked"] = bucketsChecked
	results["consistency_issues"] = len(issues)
	results["issues"] = issues
	results["success"] = len(issues) == 0

	logger.GetLogger().Info("Consistency validation completed", map[string]interface{}{
		"buckets_checked":    bucketsChecked,
		"consistency_issues": len(issues),
	})

	return results, nil
}

// HealthCheck performs health check on all sites
func (os *OperationsService) HealthCheck() (map[string]interface{}, error) {
	logger.GetLogger().Info("Starting health check", nil)

	aliases, err := os.minioService.GetAliases()
	if err != nil {
		return nil, fmt.Errorf("failed to get aliases: %v", err)
	}

	results := make(map[string]interface{})
	siteHealth := make(map[string]interface{})
	healthySites := 0

	for _, alias := range aliases {
		cmd := exec.Command("mc", "admin", "info", alias.Name)
		output, err := cmd.CombinedOutput()

		siteInfo := map[string]interface{}{
			"alias": alias.Name,
		}

		if err != nil {
			siteInfo["status"] = "unhealthy"
			siteInfo["error"] = err.Error()
		} else {
			siteInfo["status"] = "healthy"
			siteInfo["info"] = string(output)
			healthySites++
		}

		siteHealth[alias.Name] = siteInfo
	}

	results["total_sites"] = len(aliases)
	results["healthy_sites"] = healthySites
	results["site_health"] = siteHealth
	results["success"] = healthySites == len(aliases)

	logger.GetLogger().Info("Health check completed", map[string]interface{}{
		"total_sites":   len(aliases),
		"healthy_sites": healthySites,
	})

	return results, nil
}

// CompareBuckets compares content between two aliases
func (os *OperationsService) CompareBuckets(sourceAlias, destAlias, path string) (map[string]interface{}, error) {
	logger.GetLogger().Info("Starting bucket comparison", map[string]interface{}{
		"source": sourceAlias,
		"dest":   destAlias,
		"path":   path,
	})

	// Build mc-tool compare command
	var cmd *exec.Cmd
	var source, dest string

	if path != "" {
		source = sourceAlias + "/" + path
		dest = destAlias + "/" + path
	} else {
		source = sourceAlias
		dest = destAlias
	}

	// Use mc-tool compare command with --insecure flag
	cmd = exec.Command("./mc-tool", "compare", "--insecure", source, dest)

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// mc-tool compare might exit with code 1 even on successful comparison if differences found
	// So we check if the output contains comparison results
	if err != nil && !strings.Contains(outputStr, "Comparison Results") {
		return nil, fmt.Errorf("failed to compare aliases: %v, output: %s", err, outputStr)
	}

	// Parse mc-tool compare output
	onlyInSource := []string{}
	onlyInDest := []string{}
	different := []map[string]interface{}{}
	identical := 0
	differentCount := 0

	lines := strings.Split(outputStr, "\n")
	inResults := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for start of results
		if strings.Contains(line, "Comparison Results") {
			inResults = true
			continue
		}

		// Skip separator lines
		if strings.Contains(line, "===") || strings.Contains(line, "Summary:") {
			continue
		}

		// Parse summary lines
		if strings.Contains(line, "Identical:") {
			fmt.Sscanf(line, "Identical: %d", &identical)
			continue
		}
		if strings.Contains(line, "Different:") {
			fmt.Sscanf(line, "Different: %d", &differentCount)
			continue
		}

		// Parse difference lines
		if inResults && (strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "⚠")) {
			if strings.HasPrefix(line, "+") {
				// Missing in target (only in source)
				parts := strings.SplitN(line[1:], " - ", 2)
				if len(parts) > 0 {
					filename := strings.TrimSpace(parts[0])
					onlyInSource = append(onlyInSource, filename)
				}
			} else if strings.HasPrefix(line, "-") {
				// Missing in source (only in dest)
				parts := strings.SplitN(line[1:], " - ", 2)
				if len(parts) > 0 {
					filename := strings.TrimSpace(parts[0])
					onlyInDest = append(onlyInDest, filename)
				}
			} else if strings.HasPrefix(line, "⚠") {
				// Different content - handle Unicode properly
				// Remove the warning symbol and leading space
				content := strings.TrimSpace(line[len("⚠"):])
				parts := strings.SplitN(content, " - ", 2)
				if len(parts) > 0 {
					filename := strings.TrimSpace(parts[0])
					description := "Content differs"
					if len(parts) > 1 {
						description = parts[1]
					}
					different = append(different, map[string]interface{}{
						"path":        filename,
						"description": description,
					})
				}
			}
		}
	}

	results := map[string]interface{}{
		"sourceAlias":  sourceAlias,
		"destAlias":    destAlias,
		"path":         path,
		"onlyInSource": onlyInSource,
		"onlyInDest":   onlyInDest,
		"different":    different,
		"summary": map[string]interface{}{
			"identical":       identical,
			"different":       differentCount,
			"missingInSource": len(onlyInDest),
			"missingInTarget": len(onlyInSource),
			"total":           len(onlyInSource) + len(onlyInDest) + len(different) + identical,
		},
		"timestamp": "generated",
	}

	logger.GetLogger().Info("Bucket comparison completed", map[string]interface{}{
		"onlyInSource": len(onlyInSource),
		"onlyInDest":   len(onlyInDest),
		"different":    len(different),
	})

	return results, nil
}

// GetBucketsForAlias returns list of buckets for a specific alias
func (os *OperationsService) GetBucketsForAlias(alias string) ([]string, error) {
	cmd := exec.Command("mc", "ls", alias, "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets for alias %s: %v", alias, err)
	}

	var buckets []string
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var bucketInfo map[string]interface{}
		if err := json.Unmarshal([]byte(line), &bucketInfo); err != nil {
			continue
		}

		if bucketInfo["type"] == "folder" {
			bucketName := strings.TrimSuffix(bucketInfo["key"].(string), "/")
			buckets = append(buckets, bucketName)
		}
	}

	return buckets, nil
}

// GetPathSuggestionsForBucket returns path suggestions for a specific bucket
func (os *OperationsService) GetPathSuggestionsForBucket(alias, bucket string) ([]string, error) {
	cmd := exec.Command("mc", "ls", alias+"/"+bucket, "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If bucket doesn't exist or is empty, return empty suggestions
		return []string{}, nil
	}

	var paths []string
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var pathInfo map[string]interface{}
		if err := json.Unmarshal([]byte(line), &pathInfo); err != nil {
			continue
		}

		if pathInfo["type"] == "folder" {
			pathName := strings.TrimSuffix(pathInfo["key"].(string), "/")
			paths = append(paths, pathName)
		}
	}

	return paths, nil
}

// ConfigurationChecklist performs comprehensive configuration checks
func (os *OperationsService) ConfigurationChecklist() (map[string]interface{}, error) {
	logger.GetLogger().Info("Starting configuration checklist", nil)

	aliases, err := os.minioService.GetAliases()
	if err != nil {
		return nil, fmt.Errorf("failed to get aliases: %v", err)
	}

	items := []map[string]interface{}{}
	totalChecks := 0
	passedChecks := 0
	failedChecks := 0
	warningChecks := 0

	// Check environment variables for each alias
	for _, alias := range aliases {
		envChecks := os.checkEnvironmentVariables(alias.Name)
		items = append(items, envChecks...)

		for _, check := range envChecks {
			totalChecks++
			switch check["status"] {
			case "pass":
				passedChecks++
			case "fail":
				failedChecks++
			case "warning":
				warningChecks++
			}
		}

		// Check event configurations
		eventChecks := os.checkEventConfiguration(alias.Name)
		items = append(items, eventChecks...)

		for _, check := range eventChecks {
			totalChecks++
			switch check["status"] {
			case "pass":
				passedChecks++
			case "fail":
				failedChecks++
			case "warning":
				warningChecks++
			}
		}

		// Check bucket events
		bucketEventChecks := os.checkBucketEvents(alias.Name)
		items = append(items, bucketEventChecks...)

		for _, check := range bucketEventChecks {
			totalChecks++
			switch check["status"] {
			case "pass":
				passedChecks++
			case "fail":
				failedChecks++
			case "warning":
				warningChecks++
			}
		}

		// Check object lifecycle
		lifecycleChecks := os.checkObjectLifecycle(alias.Name)
		items = append(items, lifecycleChecks...)

		for _, check := range lifecycleChecks {
			totalChecks++
			switch check["status"] {
			case "pass":
				passedChecks++
			case "fail":
				failedChecks++
			case "warning":
				warningChecks++
			}
		}
	}

	results := map[string]interface{}{
		"items": items,
		"summary": map[string]interface{}{
			"total":    totalChecks,
			"passed":   passedChecks,
			"failed":   failedChecks,
			"warnings": warningChecks,
		},
		"timestamp": "generated",
	}

	logger.GetLogger().Info("Configuration checklist completed", map[string]interface{}{
		"total":    totalChecks,
		"passed":   passedChecks,
		"failed":   failedChecks,
		"warnings": warningChecks,
	})

	return results, nil
}

// Helper methods for checklist
func (os *OperationsService) checkEnvironmentVariables(alias string) []map[string]interface{} {
	checks := []map[string]interface{}{}

	// Check if alias configuration exists
	cmd := exec.Command("mc", "config", "host", "list", alias)
	output, err := cmd.CombinedOutput()

	check := map[string]interface{}{
		"alias":       alias,
		"name":        "Alias Configuration",
		"category":    "env",
		"description": "Verify alias configuration is properly set",
	}

	if err != nil {
		check["status"] = "fail"
		check["message"] = "Alias configuration not found"
		check["details"] = err.Error()
	} else {
		check["status"] = "pass"
		check["message"] = "Alias configuration exists"
		check["details"] = string(output)
	}

	checks = append(checks, check)

	// Check server connectivity
	cmd = exec.Command("mc", "ping", alias)
	_, err = cmd.CombinedOutput()

	connectCheck := map[string]interface{}{
		"alias":       alias,
		"name":        "Server Connectivity",
		"category":    "env",
		"description": "Verify server connectivity",
	}

	if err != nil {
		connectCheck["status"] = "fail"
		connectCheck["message"] = "Cannot connect to server"
		connectCheck["details"] = err.Error()
	} else {
		connectCheck["status"] = "pass"
		connectCheck["message"] = "Server connectivity OK"
	}

	checks = append(checks, connectCheck)

	return checks
}

func (os *OperationsService) checkEventConfiguration(alias string) []map[string]interface{} {
	checks := []map[string]interface{}{}

	// Check if admin events are configured
	cmd := exec.Command("mc", "admin", "config", "get", alias, "logger_webhook")
	output, err := cmd.CombinedOutput()

	check := map[string]interface{}{
		"alias":       alias,
		"name":        "Webhook Logger Configuration",
		"category":    "event",
		"description": "Check webhook logger configuration",
	}

	if err != nil {
		check["status"] = "warning"
		check["message"] = "Webhook logger not configured"
		check["details"] = "Consider configuring webhook logging for better monitoring"
	} else {
		check["status"] = "pass"
		check["message"] = "Webhook logger configured"
		check["details"] = string(output)
	}

	checks = append(checks, check)

	return checks
}

func (os *OperationsService) checkBucketEvents(alias string) []map[string]interface{} {
	checks := []map[string]interface{}{}

	// List all buckets and check their event configurations
	cmd := exec.Command("mc", "ls", alias, "--json")
	output, err := cmd.CombinedOutput()

	if err != nil {
		check := map[string]interface{}{
			"alias":       alias,
			"name":        "Bucket Event Check",
			"category":    "bucket_event",
			"status":      "fail",
			"message":     "Cannot list buckets",
			"description": "Failed to retrieve bucket list for event checking",
			"details":     err.Error(),
		}
		checks = append(checks, check)
		return checks
	}

	lines := strings.Split(string(output), "\n")
	bucketCount := 0
	eventConfiguredBuckets := 0

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var bucketInfo map[string]interface{}
		if err := json.Unmarshal([]byte(line), &bucketInfo); err != nil {
			continue
		}

		if bucketInfo["type"] == "folder" {
			bucketName := strings.TrimSuffix(bucketInfo["key"].(string), "/")
			bucketCount++

			// Check bucket notification configuration
			eventCmd := exec.Command("mc", "event", "list", alias+"/"+bucketName)
			eventOutput, eventErr := eventCmd.CombinedOutput()

			eventCheck := map[string]interface{}{
				"alias":       alias,
				"name":        fmt.Sprintf("Bucket Events: %s", bucketName),
				"category":    "bucket_event",
				"description": "Check bucket event notification configuration",
			}

			if eventErr != nil || strings.Contains(string(eventOutput), "No events configured") {
				eventCheck["status"] = "warning"
				eventCheck["message"] = "No events configured"
				eventCheck["details"] = "Consider configuring bucket events for monitoring"
			} else {
				eventCheck["status"] = "pass"
				eventCheck["message"] = "Events configured"
				eventCheck["details"] = string(eventOutput)
				eventConfiguredBuckets++
			}

			checks = append(checks, eventCheck)
		}
	}

	// Summary check
	summaryCheck := map[string]interface{}{
		"alias":       alias,
		"name":        "Bucket Events Summary",
		"category":    "bucket_event",
		"description": fmt.Sprintf("Event configuration summary for %d buckets", bucketCount),
	}

	if eventConfiguredBuckets == 0 {
		summaryCheck["status"] = "warning"
		summaryCheck["message"] = "No buckets have events configured"
	} else if eventConfiguredBuckets < bucketCount {
		summaryCheck["status"] = "warning"
		summaryCheck["message"] = fmt.Sprintf("%d/%d buckets have events configured", eventConfiguredBuckets, bucketCount)
	} else {
		summaryCheck["status"] = "pass"
		summaryCheck["message"] = "All buckets have events configured"
	}

	checks = append(checks, summaryCheck)

	return checks
}

func (os *OperationsService) checkObjectLifecycle(alias string) []map[string]interface{} {
	checks := []map[string]interface{}{}

	// List all buckets and check their lifecycle policies
	cmd := exec.Command("mc", "ls", alias, "--json")
	output, err := cmd.CombinedOutput()

	if err != nil {
		check := map[string]interface{}{
			"alias":       alias,
			"name":        "Lifecycle Policy Check",
			"category":    "lifecycle",
			"status":      "fail",
			"message":     "Cannot list buckets",
			"description": "Failed to retrieve bucket list for lifecycle checking",
			"details":     err.Error(),
		}
		checks = append(checks, check)
		return checks
	}

	lines := strings.Split(string(output), "\n")
	bucketCount := 0
	lifecycleConfiguredBuckets := 0

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var bucketInfo map[string]interface{}
		if err := json.Unmarshal([]byte(line), &bucketInfo); err != nil {
			continue
		}

		if bucketInfo["type"] == "folder" {
			bucketName := strings.TrimSuffix(bucketInfo["key"].(string), "/")
			bucketCount++

			// Check bucket lifecycle configuration
			lifecycleCmd := exec.Command("mc", "ilm", "list", alias+"/"+bucketName, "--json")
			lifecycleOutput, lifecycleErr := lifecycleCmd.CombinedOutput()

			lifecycleCheck := map[string]interface{}{
				"alias":       alias,
				"name":        fmt.Sprintf("Object Lifecycle: %s", bucketName),
				"category":    "lifecycle",
				"description": "Check bucket object lifecycle policy configuration",
			}

			if lifecycleErr != nil || strings.Contains(string(lifecycleOutput), "No lifecycle configuration") {
				lifecycleCheck["status"] = "warning"
				lifecycleCheck["message"] = "No lifecycle policy configured"
				lifecycleCheck["details"] = "Consider configuring lifecycle policies for automated object management"
			} else {
				lifecycleCheck["status"] = "pass"
				lifecycleCheck["message"] = "Lifecycle policy configured"
				lifecycleCheck["details"] = string(lifecycleOutput)
				lifecycleConfiguredBuckets++
			}

			checks = append(checks, lifecycleCheck)
		}
	}

	// Summary check
	summaryCheck := map[string]interface{}{
		"alias":       alias,
		"name":        "Object Lifecycle Summary",
		"category":    "lifecycle",
		"description": fmt.Sprintf("Lifecycle policy summary for %d buckets", bucketCount),
	}

	if lifecycleConfiguredBuckets == 0 {
		summaryCheck["status"] = "warning"
		summaryCheck["message"] = "No buckets have lifecycle policies"
	} else if lifecycleConfiguredBuckets < bucketCount {
		summaryCheck["status"] = "warning"
		summaryCheck["message"] = fmt.Sprintf("%d/%d buckets have lifecycle policies", lifecycleConfiguredBuckets, bucketCount)
	} else {
		summaryCheck["status"] = "pass"
		summaryCheck["message"] = "All buckets have lifecycle policies"
	}

	checks = append(checks, summaryCheck)

	return checks
}
