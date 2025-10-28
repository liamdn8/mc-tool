package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

// HandleGetResyncOptions handles GET /api/operations/resync/options
func (h *OperationsHandler) HandleGetResyncOptions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	result, err := h.operationsService.GetResyncOptions()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.RespondJSON(w, result)
}

// resyncRequest represents the payload for starting a resync operation.
type resyncRequest struct {
	SourceAlias string `json:"sourceAlias"`
	TargetAlias string `json:"targetAlias"`
}

// HandleStartResync handles POST /api/operations/resync/start
func (h *OperationsHandler) HandleStartResync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req resyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.SourceAlias == "" || req.TargetAlias == "" {
		h.RespondError(w, http.StatusBadRequest, "sourceAlias and targetAlias are required")
		return
	}

	result, err := h.operationsService.StartReplicationResync(req.SourceAlias, req.TargetAlias)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "must") || strings.Contains(err.Error(), "not part") {
			status = http.StatusBadRequest
		}
		h.RespondError(w, status, err.Error())
		return
	}

	h.RespondJSON(w, result)
}

// HandleGetResyncStatus handles GET /api/operations/resync/status
func (h *OperationsHandler) HandleGetResyncStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	sourceAlias := r.URL.Query().Get("sourceAlias")
	targetAlias := r.URL.Query().Get("targetAlias")

	if sourceAlias == "" || targetAlias == "" {
		h.RespondError(w, http.StatusBadRequest, "sourceAlias and targetAlias query parameters are required")
		return
	}

	result, err := h.operationsService.GetReplicationResyncStatus(sourceAlias, targetAlias)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.RespondJSON(w, result)
}
