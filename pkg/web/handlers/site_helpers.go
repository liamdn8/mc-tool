package handlers

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

func (h *SiteHandler) getMCInternalAliases() ([]map[string]interface{}, error) {
	cmd := exec.Command("mc", "alias", "list", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var aliases []map[string]interface{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var aliasData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &aliasData); err != nil {
			continue
		}

		aliasName, okName := aliasData["alias"].(string)
		aliasURL, okURL := aliasData["URL"].(string)
		if !okName || !okURL {
			continue
		}

		// Return immediately with "checking" status - let frontend check health async
		alias := map[string]interface{}{
			"name":    aliasName,
			"url":     aliasURL,
			"healthy": false,
			"status":  "checking",
		}

		if accessKey, ok := aliasData["accessKey"].(string); ok {
			alias["accessKey"] = accessKey
		}
		if api, ok := aliasData["api"].(string); ok {
			alias["api"] = api
		}
		if path, ok := aliasData["path"].(string); ok {
			alias["path"] = path
		}

		aliases = append(aliases, alias)
	}

	return aliases, nil
}

func (h *SiteHandler) getAliasHealthStatus(alias string) (bool, string) {
	cmd := exec.Command("mc", "admin", "info", alias, "--json")
	output, err := cmd.CombinedOutput()
	if err == nil {
		var result map[string]interface{}
		if json.Unmarshal(output, &result) == nil {
			if status, ok := result["status"].(string); ok && status == "success" {
				return true, "healthy"
			}
		}
	}

	cmd = exec.Command("mc", "ls", alias)
	if cmd.Run() == nil {
		return true, "healthy"
	}

	return false, "unhealthy"
}

func (h *SiteHandler) listBuckets(alias string) ([]string, error) {
	cmd := exec.Command("mc", "ls", alias, "--json")
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
	cmd := exec.Command("mc", "du", fmt.Sprintf("%s/%s", alias, bucket), "--json")
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
