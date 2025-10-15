package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/liamdn8/mc-tool/pkg/web/services"
)

// SiteHandler handles site-related requests
type SiteHandler struct {
	BaseHandler
	minioService *services.MinIOService
}

// NewSiteHandler creates a new site handler
func NewSiteHandler(minioService *services.MinIOService) *SiteHandler {
	return &SiteHandler{
		minioService: minioService,
	}
}

// HandleSites handles GET /api/sites (alias list)
func (h *SiteHandler) HandleSites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	aliases, err := h.minioService.GetAliases()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to get aliases")
		return
	}

	h.RespondJSON(w, map[string]interface{}{
		"sites": aliases,
	})
}

// HandleGetAliases handles GET /api/aliases
func (h *SiteHandler) HandleGetAliases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	aliases, err := h.getMCInternalAliases()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get aliases: %v", err))
		return
	}

	h.RespondJSON(w, map[string]interface{}{
		"aliases": aliases,
	})
}

// HandleGetAliasesWithStats handles GET /api/aliases-stats
func (h *SiteHandler) HandleGetAliasesWithStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	aliases, err := h.getMCInternalAliases()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get aliases: %v", err))
		return
	}

	// Get stats for each alias
	var aliasesWithStats []map[string]interface{}
	for _, alias := range aliases {
		aliasName, ok := alias["name"].(string)
		if !ok {
			continue
		}

		stats := h.getAliasStats(aliasName)

		aliasData := map[string]interface{}{
			"name":          alias["name"],
			"url":           alias["url"],
			"healthy":       alias["healthy"],
			"status":        alias["status"],
			"bucket_count":  stats["bucket_count"],
			"total_size":    stats["total_size"],
			"total_objects": stats["total_objects"],
		}

		// Add additional info if available
		if accessKey, ok := alias["accessKey"]; ok {
			aliasData["accessKey"] = accessKey
		}
		if api, ok := alias["api"]; ok {
			aliasData["api"] = api
		}

		aliasesWithStats = append(aliasesWithStats, aliasData)
	}

	h.RespondJSON(w, map[string]interface{}{
		"aliases": aliasesWithStats,
	})
}

// HandleAliasHealth handles GET /api/alias-health
func (h *SiteHandler) HandleAliasHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	alias := r.URL.Query().Get("alias")
	if alias == "" {
		h.RespondError(w, http.StatusBadRequest, "Alias parameter is required")
		return
	}

	// Try to get admin info
	cmd := exec.Command("mc", "admin", "info", alias, "--json")
	output, err := cmd.CombinedOutput()

	healthy := false
	message := "Unknown"
	objectCount := int64(0)
	totalSize := int64(0)
	bucketCount := int64(0)
	serverCount := 0

	if err == nil {
		// Parse JSON output to extract detailed info
		var result map[string]interface{}
		if json.Unmarshal(output, &result) == nil {
			if status, ok := result["status"].(string); ok && status == "success" {
				healthy = true
				message = "Connected"

				// Extract info object
				if info, ok := result["info"].(map[string]interface{}); ok {
					// Get object count
					if objects, ok := info["objects"].(map[string]interface{}); ok {
						if count, ok := objects["count"].(float64); ok {
							objectCount = int64(count)
						}
					}

					// Get total size
					if usage, ok := info["usage"].(map[string]interface{}); ok {
						if size, ok := usage["size"].(float64); ok {
							totalSize = int64(size)
						}
					}

					// Get bucket count
					if buckets, ok := info["buckets"].(map[string]interface{}); ok {
						if count, ok := buckets["count"].(float64); ok {
							bucketCount = int64(count)
						}
					}

					// Get server count
					if servers, ok := info["servers"].([]interface{}); ok {
						serverCount = len(servers)
					}
				}
			}
		}
	} else {
		// Try simple ls command as fallback
		cmd = exec.Command("mc", "ls", alias)
		if cmd.Run() == nil {
			healthy = true
			message = "Connected (limited)"
		} else {
			message = "Unreachable"
		}
	}

	h.RespondJSON(w, map[string]interface{}{
		"healthy":     healthy,
		"message":     message,
		"objectCount": objectCount,
		"totalSize":   totalSize,
		"bucketCount": bucketCount,
		"serverCount": serverCount,
	})
}

// HandleSiteHealth handles GET /api/sites/health
func (h *SiteHandler) HandleSiteHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	aliases, err := h.minioService.GetAliases()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to get aliases")
		return
	}

	healthData := make(map[string]interface{})
	for _, alias := range aliases {
		health := h.minioService.GetAliasHealth(alias.Name)
		healthData[alias.Name] = health
	}

	h.RespondJSON(w, healthData)
}

// Helper methods
func (h *SiteHandler) getMCInternalAliases() ([]map[string]interface{}, error) {
	// Try to get aliases using mc command
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
		if err := json.Unmarshal([]byte(line), &aliasData); err == nil {
			if aliasName, ok := aliasData["alias"].(string); ok {
				if url, ok := aliasData["URL"].(string); ok {
					// Get health status for this alias
					healthy, status := h.getAliasHealthStatus(aliasName)

					alias := map[string]interface{}{
						"name":    aliasName,
						"url":     url,
						"healthy": healthy,
						"status":  status,
					}

					// Add additional info if available
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
			}
		}
	}

	return aliases, nil
}

// getAliasHealthStatus checks if an alias is healthy and returns status
func (h *SiteHandler) getAliasHealthStatus(alias string) (bool, string) {
	// Try to get admin info first
	cmd := exec.Command("mc", "admin", "info", alias, "--json")
	output, err := cmd.CombinedOutput()

	if err == nil {
		// Parse JSON output to check status
		var result map[string]interface{}
		if json.Unmarshal(output, &result) == nil {
			if status, ok := result["status"].(string); ok && status == "success" {
				return true, "healthy"
			}
		}
	}

	// Try simple ls command as fallback
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
		if err := json.Unmarshal([]byte(line), &data); err == nil {
			if key, ok := data["key"].(string); ok {
				buckets = append(buckets, strings.TrimSuffix(key, "/"))
			}
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

	// Get list of buckets
	buckets, err := h.listBuckets(alias)
	if err != nil {
		return stats
	}

	stats["bucket_count"] = len(buckets)

	// Get stats for each bucket
	var bucketStats []map[string]interface{}
	var totalSize int64
	var totalObjects int64

	for _, bucket := range buckets {
		bucketStat := h.getBucketStats(alias, bucket)
		if bucketStat != nil {
			bucketStats = append(bucketStats, bucketStat)
			if size, ok := bucketStat["size"].(int64); ok {
				totalSize += size
			}
			if objects, ok := bucketStat["objects"].(int64); ok {
				totalObjects += objects
			}
		}
	}

	stats["buckets"] = bucketStats
	stats["total_size"] = totalSize
	stats["total_objects"] = totalObjects

	return stats
}

func (h *SiteHandler) getBucketStats(alias, bucket string) map[string]interface{} {
	// Use mc du command to get bucket statistics
	cmd := exec.Command("mc", "du", fmt.Sprintf("%s/%s", alias, bucket), "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If du fails, try to get basic info
		return map[string]interface{}{
			"name":    bucket,
			"size":    int64(0),
			"objects": int64(0),
		}
	}

	// Parse the last line of output which contains the total
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 {
		return map[string]interface{}{
			"name":    bucket,
			"size":    int64(0),
			"objects": int64(0),
		}
	}

	// Parse the JSON output
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
