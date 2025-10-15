package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/liamdn8/mc-tool/pkg/web/services"
)

// ReplicationHandler handles replication-related requests
type ReplicationHandler struct {
	BaseHandler
	replicationService *services.ReplicationService
	minioService       *services.MinIOService
}

// NewReplicationHandler creates a new replication handler
func NewReplicationHandler(replicationService *services.ReplicationService, minioService *services.MinIOService) *ReplicationHandler {
	return &ReplicationHandler{
		replicationService: replicationService,
		minioService:       minioService,
	}
}

// HandleReplicationInfo handles GET /api/replication/info
func (h *ReplicationHandler) HandleReplicationInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	info, err := h.replicationService.GetReplicationInfo()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get replication info: %v", err))
		return
	}

	h.RespondJSON(w, info)
}

// HandleReplicationStatus handles GET /api/replication/status
func (h *ReplicationHandler) HandleReplicationStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	aliases, err := h.getMCInternalAliases()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get aliases: %v", err))
		return
	}

	status := make(map[string]interface{})
	status["status"] = "healthy"
	sites := make(map[string]interface{})

	for _, alias := range aliases {
		// Get bucket list
		cmd := exec.Command("mc", "ls", alias["name"], "--json")
		output, err := cmd.CombinedOutput()

		bucketCount := 0
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			for _, line := range lines {
				if line != "" {
					bucketCount++
				}
			}
		}

		sites[alias["name"]] = map[string]interface{}{
			"replicatedBuckets": bucketCount,
			"pendingObjects":    0,
			"failedObjects":     0,
			"lastSyncTime":      time.Now().Format(time.RFC3339),
			"healthy":           true,
		}
	}

	status["sites"] = sites
	h.RespondJSON(w, status)
}

// HandleReplicationCompare handles GET /api/replication/compare
func (h *ReplicationHandler) HandleReplicationCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	aliases, err := h.getMCInternalAliases()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get aliases: %v", err))
		return
	}

	if len(aliases) < 2 {
		h.RespondJSON(w, map[string]interface{}{
			"buckets": map[string]interface{}{},
			"message": "Need at least 2 sites to compare",
		})
		return
	}

	// Collect all buckets from all sites
	type BucketInfo struct {
		Sites      []string
		Policy     map[string]string
		Lifecycle  map[string]interface{}
		Versioning map[string]string
	}

	allBuckets := make(map[string]*BucketInfo)

	for _, alias := range aliases {
		cmd := exec.Command("mc", "ls", alias["name"], "--json")
		output, err := cmd.CombinedOutput()
		if err != nil {
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

			bucketName := ""
			if key, ok := bucketData["key"].(string); ok {
				bucketName = strings.TrimSuffix(key, "/")
			}

			if bucketName == "" {
				continue
			}

			if _, exists := allBuckets[bucketName]; !exists {
				allBuckets[bucketName] = &BucketInfo{
					Sites:      []string{},
					Policy:     make(map[string]string),
					Lifecycle:  make(map[string]interface{}),
					Versioning: make(map[string]string),
				}
			}

			allBuckets[bucketName].Sites = append(allBuckets[bucketName].Sites, alias["name"])

			// Get bucket policy
			policyCmd := exec.Command("mc", "anonymous", "get", fmt.Sprintf("%s/%s", alias["name"], bucketName))
			policyOutput, _ := policyCmd.CombinedOutput()
			allBuckets[bucketName].Policy[alias["name"]] = string(policyOutput)

			// Get lifecycle
			ilmCmd := exec.Command("mc", "ilm", "ls", fmt.Sprintf("%s/%s", alias["name"], bucketName), "--json")
			ilmOutput, _ := ilmCmd.CombinedOutput()
			if ilmOutput != nil {
				var ilmData interface{}
				json.Unmarshal(ilmOutput, &ilmData)
				allBuckets[bucketName].Lifecycle[alias["name"]] = ilmData
			}

			// Get versioning
			versionCmd := exec.Command("mc", "version", "info", fmt.Sprintf("%s/%s", alias["name"], bucketName), "--json")
			versionOutput, _ := versionCmd.CombinedOutput()
			allBuckets[bucketName].Versioning[alias["name"]] = string(versionOutput)
		}
	}

	// Build comparison result
	result := make(map[string]interface{})
	for bucketName, info := range allBuckets {
		bucketResult := map[string]interface{}{
			"existsOn": info.Sites,
			"policy": map[string]interface{}{
				"consistent": h.checkConsistency(info.Policy),
				"values":     info.Policy,
			},
			"lifecycle": map[string]interface{}{
				"consistent": h.checkConsistency(info.Lifecycle),
				"values":     info.Lifecycle,
			},
			"versioning": map[string]interface{}{
				"consistent": h.checkConsistency(info.Versioning),
				"values":     info.Versioning,
			},
		}
		result[bucketName] = bucketResult
	}

	h.RespondJSON(w, map[string]interface{}{
		"buckets": result,
	})
}

// HandleReplicationAdd handles POST /api/replication/add
func (h *ReplicationHandler) HandleReplicationAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Aliases []string `json:"aliases"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	if len(req.Aliases) < 2 {
		h.RespondError(w, http.StatusBadRequest, "At least 2 aliases are required")
		return
	}

	err := h.replicationService.AddSiteReplication(req.Aliases)
	if err != nil {
		// Parse error message to provide better feedback
		errorMsg := err.Error()
		var userFriendlyMsg string

		if strings.Contains(errorMsg, "localhost") || strings.Contains(errorMsg, "127.0.0.1") {
			userFriendlyMsg = "❌ Site Replication Setup Failed\n\n" +
				"The MinIO servers are configured with localhost endpoints and cannot connect to each other.\n\n" +
				"📋 Requirements for Site Replication:\n" +
				"1. Each MinIO server must have a publicly accessible endpoint (not localhost)\n" +
				"2. All sites must be able to reach each other over the network\n" +
				"3. Use IP addresses or domain names instead of localhost\n\n" +
				"🔧 How to fix:\n" +
				"1. Reconfigure your MinIO aliases with accessible endpoints:\n" +
				"   Example: mc alias set site1 http://192.168.1.10:9000 accesskey secretkey\n" +
				"   Example: mc alias set site2 http://192.168.1.11:9000 accesskey secretkey\n\n" +
				"2. Ensure MinIO servers are started with accessible addresses:\n" +
				"   Example: MINIO_SERVER_URL=http://192.168.1.10:9000 minio server /data\n\n" +
				"📖 Technical Details:\n" + errorMsg
		} else if strings.Contains(errorMsg, "connection refused") {
			userFriendlyMsg = "❌ Site Replication Setup Failed\n\n" +
				"Cannot connect to one or more MinIO servers.\n\n" +
				"Possible causes:\n" +
				"1. MinIO server is not running\n" +
				"2. Firewall blocking connections\n" +
				"3. Wrong port number\n" +
				"4. Network connectivity issues\n\n" +
				"📖 Technical Details:\n" + errorMsg
		} else {
			userFriendlyMsg = "Failed to add replication:\n\n" + errorMsg
		}

		h.RespondError(w, http.StatusInternalServerError, userFriendlyMsg)
		return
	}

	h.RespondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Site replication added successfully",
	})
}

// HandleReplicationRemove handles POST /api/replication/remove
func (h *ReplicationHandler) HandleReplicationRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Alias string `json:"alias"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	err := h.replicationService.RemoveEntireReplication(req.Alias)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.RespondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Site replication configuration removed successfully",
	})
}

// HandleReplicationAddSmart handles POST /api/replication/add-smart
func (h *ReplicationHandler) HandleReplicationAddSmart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Aliases []string `json:"aliases"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	result, err := h.replicationService.AddSiteReplicationSmart(req.Aliases)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.RespondJSON(w, map[string]interface{}{
		"success": true,
		"data":    result,
		"message": "Smart site replication operation completed successfully",
	})
}

// HandleReplicationRemoveSiteSmart handles POST /api/replication/remove-site-smart
func (h *ReplicationHandler) HandleReplicationRemoveSiteSmart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Alias   string   `json:"alias"`   // Single alias to remove
		Aliases []string `json:"aliases"` // Multiple aliases to remove (bulk operation)
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Determine which aliases to remove
	var aliasesToRemove []string
	if req.Alias != "" {
		aliasesToRemove = []string{req.Alias}
	} else if len(req.Aliases) > 0 {
		aliasesToRemove = req.Aliases
	} else {
		h.RespondError(w, http.StatusBadRequest, "No alias or aliases specified for removal")
		return
	}

	results, failed, err := h.replicationService.RemoveIndividualSitesSmart(aliasesToRemove)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Prepare response
	if len(failed) > 0 {
		h.RespondJSON(w, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to remove %d out of %d sites", len(failed), len(aliasesToRemove)),
			"results": results,
			"failed":  failed,
		})
	} else {
		h.RespondJSON(w, map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("Successfully removed %d site(s) from replication cluster", len(aliasesToRemove)),
			"results": results,
		})
	}
}

// HandleReplicationRemoveSite handles POST /api/replication/remove-site
func (h *ReplicationHandler) HandleReplicationRemoveSite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Alias   string   `json:"alias"`   // Single alias to remove
		Aliases []string `json:"aliases"` // Multiple aliases to remove (bulk operation)
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Determine which aliases to remove
	var aliasesToRemove []string
	if req.Alias != "" {
		aliasesToRemove = []string{req.Alias}
	} else if len(req.Aliases) > 0 {
		aliasesToRemove = req.Aliases
	} else {
		h.RespondError(w, http.StatusBadRequest, "No alias or aliases specified for removal")
		return
	}

	results, failed, err := h.replicationService.RemoveIndividualSites(aliasesToRemove)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Prepare response
	if len(failed) > 0 {
		h.RespondJSON(w, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to remove %d out of %d sites", len(failed), len(aliasesToRemove)),
			"results": results,
			"failed":  failed,
		})
	} else {
		h.RespondJSON(w, map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("Successfully removed %d site(s) from replication cluster", len(aliasesToRemove)),
			"results": results,
		})
	}
}

// HandleReplicationResync handles POST /api/replication/resync
func (h *ReplicationHandler) HandleReplicationResync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		SourceAlias string `json:"source_alias"`
		TargetAlias string `json:"target_alias"`
		Direction   string `json:"direction"` // "resync-from" or "resync-to"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	err := h.replicationService.ResyncSites(req.SourceAlias, req.TargetAlias, req.Direction)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.RespondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Resync started successfully",
	})
}

// Helper methods
func (h *ReplicationHandler) getMCInternalAliases() ([]map[string]string, error) {
	// Try to get aliases using mc command
	cmd := exec.Command("mc", "alias", "list", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var aliases []map[string]string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		var aliasData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &aliasData); err == nil {
			if aliasName, ok := aliasData["alias"].(string); ok {
				if url, ok := aliasData["URL"].(string); ok {
					aliases = append(aliases, map[string]string{
						"name": aliasName,
						"url":  url,
					})
				}
			}
		}
	}

	return aliases, nil
}

func (h *ReplicationHandler) checkConsistency(data interface{}) bool {
	switch v := data.(type) {
	case map[string]string:
		if len(v) <= 1 {
			return true
		}
		var firstValue string
		first := true
		for _, value := range v {
			if first {
				firstValue = value
				first = false
			} else if value != firstValue {
				return false
			}
		}
		return true
	case map[string]interface{}:
		if len(v) <= 1 {
			return true
		}
		var firstValue string
		first := true
		for _, value := range v {
			valueStr := fmt.Sprintf("%v", value)
			if first {
				firstValue = valueStr
				first = false
			} else if valueStr != firstValue {
				return false
			}
		}
		return true
	}
	return true
}

// HandleSplitBrainCheck handles GET /api/replication/split-brain-check
func (h *ReplicationHandler) HandleSplitBrainCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	status, err := h.replicationService.CheckSplitBrainStatus()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check split brain status: %v", err))
		return
	}

	h.RespondJSON(w, status)
}
