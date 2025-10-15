package services

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// MinIOService handles MinIO operations
type MinIOService struct {
	executablePath string
}

// NewMinIOService creates a new MinIO service
func NewMinIOService(executablePath string) *MinIOService {
	return &MinIOService{
		executablePath: executablePath,
	}
}

// Alias represents a MinIO alias
type Alias struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// GetAliases retrieves all configured MinIO aliases
func (ms *MinIOService) GetAliases() ([]Alias, error) {
	cmd := exec.Command("mc", "alias", "list", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var aliases []Alias
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		var aliasData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &aliasData); err == nil {
			if aliasName, ok := aliasData["alias"].(string); ok {
				if url, ok := aliasData["URL"].(string); ok {
					aliases = append(aliases, Alias{
						Name: aliasName,
						URL:  url,
					})
				}
			}
		}
	}

	return aliases, nil
}

// ListBuckets lists all buckets for an alias
func (ms *MinIOService) ListBuckets(alias string) ([]string, error) {
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
		if err := json.Unmarshal([]byte(line), &data); err == nil {
			if key, ok := data["key"].(string); ok {
				buckets = append(buckets, strings.TrimSuffix(key, "/"))
			}
		}
	}

	return buckets, nil
}

// GetBucketStats returns statistics for a specific bucket
func (ms *MinIOService) GetBucketStats(alias, bucket string) map[string]interface{} {
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
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &data); err == nil {
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

	return map[string]interface{}{
		"name":    bucket,
		"size":    int64(0),
		"objects": int64(0),
	}
}

// GetAliasHealth checks the health status of an alias
func (ms *MinIOService) GetAliasHealth(alias string) map[string]interface{} {
	cmd := exec.Command("mc", "admin", "info", alias, "--json")
	output, err := cmd.CombinedOutput()

	healthy := false
	message := "Unknown"
	objectCount := int64(0)
	totalSize := int64(0)
	bucketCount := int64(0)
	serverCount := 0

	if err == nil {
		var result map[string]interface{}
		if json.Unmarshal(output, &result) == nil {
			if status, ok := result["status"].(string); ok && status == "success" {
				healthy = true
				message = "Connected"

				if info, ok := result["info"].(map[string]interface{}); ok {
					if objects, ok := info["objects"].(map[string]interface{}); ok {
						if count, ok := objects["count"].(float64); ok {
							objectCount = int64(count)
						}
					}

					if usage, ok := info["usage"].(map[string]interface{}); ok {
						if size, ok := usage["size"].(float64); ok {
							totalSize = int64(size)
						}
					}

					if buckets, ok := info["buckets"].(map[string]interface{}); ok {
						if count, ok := buckets["count"].(float64); ok {
							bucketCount = int64(count)
						}
					}

					if servers, ok := info["servers"].([]interface{}); ok {
						serverCount = len(servers)
					}
				}
			}
		}
	} else {
		cmd = exec.Command("mc", "ls", alias)
		if cmd.Run() == nil {
			healthy = true
			message = "Connected (limited)"
		} else {
			message = "Unreachable"
		}
	}

	return map[string]interface{}{
		"healthy":     healthy,
		"message":     message,
		"objectCount": objectCount,
		"totalSize":   totalSize,
		"bucketCount": bucketCount,
		"serverCount": serverCount,
	}
}
