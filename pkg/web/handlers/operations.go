package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/liamdn8/mc-tool/pkg/web/services"
)

// OperationsHandler handles operations-related requests
type OperationsHandler struct {
	BaseHandler
	operationsService *services.OperationsService
}

// NewOperationsHandler creates a new operations handler
func NewOperationsHandler(operationsService *services.OperationsService) *OperationsHandler {
	return &OperationsHandler{
		operationsService: operationsService,
	}
}

// HandleSyncPolicies handles POST /api/operations/sync-policies
func (h *OperationsHandler) HandleSyncPolicies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	result, err := h.operationsService.SyncBucketPolicies()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.RespondJSON(w, result)
}

// HandleSyncLifecycle handles POST /api/operations/sync-lifecycle
func (h *OperationsHandler) HandleSyncLifecycle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	result, err := h.operationsService.SyncLifecyclePolicies()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.RespondJSON(w, result)
}

// HandleValidateConsistency handles POST /api/operations/validate-consistency
func (h *OperationsHandler) HandleValidateConsistency(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	result, err := h.operationsService.ValidateConsistency()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.RespondJSON(w, result)
}

// HandleHealthCheck handles POST /api/operations/health-check
func (h *OperationsHandler) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	result, err := h.operationsService.HealthCheck()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.RespondJSON(w, result)
}

// CompareRequest represents the request for comparing aliases
type CompareRequest struct {
	SourceAlias string `json:"sourceAlias"`
	DestAlias   string `json:"destAlias"`
	Path        string `json:"path"`
}

// HandleCompare handles POST /api/operations/compare
func (h *OperationsHandler) HandleCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req CompareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.SourceAlias == "" || req.DestAlias == "" {
		h.RespondError(w, http.StatusBadRequest, "Source and destination aliases are required")
		return
	}

	result, err := h.operationsService.CompareBuckets(req.SourceAlias, req.DestAlias, req.Path)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.RespondJSON(w, result)
}

// HandleChecklist handles POST /api/operations/checklist
func (h *OperationsHandler) HandleChecklist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	result, err := h.operationsService.ConfigurationChecklist()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.RespondJSON(w, result)
}

// HandleGetBuckets handles GET /api/operations/buckets?alias=<alias>
func (h *OperationsHandler) HandleGetBuckets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	alias := r.URL.Query().Get("alias")
	if alias == "" {
		h.RespondError(w, http.StatusBadRequest, "Alias parameter is required")
		return
	}

	buckets, err := h.operationsService.GetBucketsForAlias(alias)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := map[string]interface{}{
		"alias":   alias,
		"buckets": buckets,
	}

	h.RespondJSON(w, result)
}

// HandleGetPathSuggestions handles GET /api/operations/path-suggestions?alias=<alias>&bucket=<bucket>
func (h *OperationsHandler) HandleGetPathSuggestions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	alias := r.URL.Query().Get("alias")
	bucket := r.URL.Query().Get("bucket")

	if alias == "" || bucket == "" {
		h.RespondError(w, http.StatusBadRequest, "Both alias and bucket parameters are required")
		return
	}

	paths, err := h.operationsService.GetPathSuggestionsForBucket(alias, bucket)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := map[string]interface{}{
		"alias":  alias,
		"bucket": bucket,
		"paths":  paths,
	}

	h.RespondJSON(w, result)
}
