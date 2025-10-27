package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

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

			policyCmd := exec.Command("mc", "anonymous", "get", fmt.Sprintf("%s/%s", alias["name"], bucketName))
			policyOutput, _ := policyCmd.CombinedOutput()
			allBuckets[bucketName].Policy[alias["name"]] = string(policyOutput)

			ilmCmd := exec.Command("mc", "ilm", "ls", fmt.Sprintf("%s/%s", alias["name"], bucketName), "--json")
			ilmOutput, _ := ilmCmd.CombinedOutput()
			if len(ilmOutput) > 0 {
				var ilmData interface{}
				json.Unmarshal(ilmOutput, &ilmData)
				allBuckets[bucketName].Lifecycle[alias["name"]] = ilmData
			}

			versionCmd := exec.Command("mc", "version", "info", fmt.Sprintf("%s/%s", alias["name"], bucketName), "--json")
			versionOutput, _ := versionCmd.CombinedOutput()
			allBuckets[bucketName].Versioning[alias["name"]] = string(versionOutput)
		}
	}

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
