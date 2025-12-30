package handlers

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/liamdn8/mc-tool/pkg/config"
)

func (h *SiteHandler) getMCInternalAliases() ([]map[string]interface{}, error) {
	// Load configuration from both file and environment variables
	mcConfig, err := config.LoadMCConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load MC config: %v", err)
	}

	var aliases []map[string]interface{}

	// Convert config aliases to the expected format
	for aliasName, aliasConfig := range mcConfig.Aliases {
		alias := map[string]interface{}{
			"name":      aliasName,
			"url":       aliasConfig.URL,
			"accessKey": aliasConfig.AccessKey,
			"api":       aliasConfig.API,
			"path":      aliasConfig.Path,
			"insecure":  aliasConfig.Insecure,
			"healthy":   false,
			"status":    "checking",
		}

		aliases = append(aliases, alias)
	}

	return aliases, nil
}

// setupMCCommand creates an mc command with environment variables
func (h *SiteHandler) setupMCCommand(args ...string) *exec.Cmd {
	cmd := exec.Command("mc", args...)

	// Load MC config and set environment variables
	mcConfig, err := config.LoadMCConfig()
	if err == nil {
		cmd.Env = config.GetMCEnvironment(mcConfig)
	}

	return cmd
}

func (h *SiteHandler) getAliasHealthStatus(alias string) (bool, string) {
	cmd := h.setupMCCommand("admin", "info", alias, "--json")
	output, err := cmd.CombinedOutput()
	if err == nil {
		var result map[string]interface{}
		if json.Unmarshal(output, &result) == nil {
			if status, ok := result["status"].(string); ok && status == "success" {
				return true, "healthy"
			}
		}
	}

	cmd = h.setupMCCommand("ls", alias)
	if cmd.Run() == nil {
		return true, "healthy"
	}

	return false, "unhealthy"
}

func (h *SiteHandler) listBuckets(alias string) ([]string, error) {
	cmd := h.setupMCCommand("ls", alias, "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var buckets []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			continue
		}

		if key, ok := data["key"].(string); ok {
			buckets = append(buckets, strings.TrimSuffix(key, "/"))
		}
	}

	return buckets, nil
}

func (h *SiteHandler) getAliasStats(alias string) map[string]interface{} {
	stats := map[string]interface{}{
		"bucket_count":  0,
		"total_size":    int64(0),
		"total_objects": int64(0),
		"buckets":       []map[string]interface{}{},
	}

	buckets, err := h.listBuckets(alias)
	if err != nil {
		return stats
	}

	stats["bucket_count"] = len(buckets)

	var (
		bucketStats  []map[string]interface{}
		totalSize    int64
		totalObjects int64
	)

	for _, bucket := range buckets {
		bucketStat := h.getBucketStats(alias, bucket)
		if bucketStat == nil {
			continue
		}

		bucketStats = append(bucketStats, bucketStat)
		if size, ok := bucketStat["size"].(int64); ok {
			totalSize += size
		}
		if objects, ok := bucketStat["objects"].(int64); ok {
			totalObjects += objects
		}
	}

	stats["buckets"] = bucketStats
	stats["total_size"] = totalSize
	stats["total_objects"] = totalObjects

	return stats
}

func (h *SiteHandler) getBucketStats(alias, bucket string) map[string]interface{} {
	cmd := h.setupMCCommand("du", fmt.Sprintf("%s/%s", alias, bucket), "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{
			"name":    bucket,
			"size":    int64(0),
			"objects": int64(0),
		}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 {
		return map[string]interface{}{
			"name":    bucket,
			"size":    int64(0),
			"objects": int64(0),
		}
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &data); err != nil {
		return map[string]interface{}{
			"name":    bucket,
			"size":    int64(0),
			"objects": int64(0),
		}
	}

	size := int64(0)
	objects := int64(0)

	if sizeFloat, ok := data["size"].(float64); ok {
		size = int64(sizeFloat)
	}
	if objCount, ok := data["objects"].(float64); ok {
		objects = int64(objCount)
	}

	return map[string]interface{}{
		"name":    bucket,
		"size":    size,
		"objects": objects,
	}
}
