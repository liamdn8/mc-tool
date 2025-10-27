package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

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

	if err := h.replicationService.AddSiteReplication(req.Aliases); err != nil {
		h.RespondError(w, http.StatusInternalServerError, buildReplicationErrorMessage(err.Error()))
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

	if err := h.replicationService.RemoveEntireReplication(req.Alias); err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.RespondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Site replication configuration removed successfully",
	})
}

// HandleReplicationRemoveSite handles POST /api/replication/remove-site
func (h *ReplicationHandler) HandleReplicationRemoveSite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	aliasesToRemove, err := parseReplicationRemovalRequest(r)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	results, failed, serviceErr := h.replicationService.RemoveIndividualSites(aliasesToRemove)
	if serviceErr != nil {
		h.RespondError(w, http.StatusInternalServerError, serviceErr.Error())
		return
	}

	h.writeReplicationRemovalResponse(w, aliasesToRemove, results, failed)
}

// HandleReplicationRemoveSiteSmart handles POST /api/replication/remove-site-smart
func (h *ReplicationHandler) HandleReplicationRemoveSiteSmart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	aliasesToRemove, err := parseReplicationRemovalRequest(r)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	results, failed, serviceErr := h.replicationService.RemoveIndividualSitesSmart(aliasesToRemove)
	if serviceErr != nil {
		h.RespondError(w, http.StatusInternalServerError, serviceErr.Error())
		return
	}

	h.writeReplicationRemovalResponse(w, aliasesToRemove, results, failed)
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

// HandleReplicationResync handles POST /api/replication/resync
func (h *ReplicationHandler) HandleReplicationResync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		SourceAlias string `json:"source_alias"`
		TargetAlias string `json:"target_alias"`
		Direction   string `json:"direction"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	if err := h.replicationService.ResyncSites(req.SourceAlias, req.TargetAlias, req.Direction); err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.RespondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Resync started successfully",
	})
}

func parseReplicationRemovalRequest(r *http.Request) ([]string, error) {
	var req struct {
		Alias   string   `json:"alias"`
		Aliases []string `json:"aliases"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("Invalid request: %v", err)
	}

	switch {
	case req.Alias != "":
		return []string{req.Alias}, nil
	case len(req.Aliases) > 0:
		return req.Aliases, nil
	default:
		return nil, fmt.Errorf("No alias or aliases specified for removal")
	}
}

func (h *ReplicationHandler) writeReplicationRemovalResponse(w http.ResponseWriter, requested []string, results []map[string]interface{}, failed []string) {
	if len(failed) > 0 {
		h.RespondJSON(w, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to remove %d out of %d sites", len(failed), len(requested)),
			"results": results,
			"failed":  failed,
		})
		return
	}

	h.RespondJSON(w, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Successfully removed %d site(s) from replication cluster", len(requested)),
		"results": results,
	})
}

func buildReplicationErrorMessage(errorMsg string) string {
	switch {
	case strings.Contains(errorMsg, "localhost"), strings.Contains(errorMsg, "127.0.0.1"):
		return "❌ Site Replication Setup Failed\n\n" +
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
	case strings.Contains(errorMsg, "connection refused"):
		return "❌ Site Replication Setup Failed\n\n" +
			"Cannot connect to one or more MinIO servers.\n\n" +
			"Possible causes:\n" +
			"1. MinIO server is not running\n" +
			"2. Firewall blocking connections\n" +
			"3. Wrong port number\n" +
			"4. Network connectivity issues\n\n" +
			"📖 Technical Details:\n" + errorMsg
	default:
		return "Failed to add replication:\n\n" + errorMsg
	}
}
