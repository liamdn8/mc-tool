package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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
	SourceAlias    string `json:"sourceAlias"`
	DestAlias      string `json:"destAlias"`
	Path           string `json:"path"`
	CompareVersion bool   `json:"compareVersion"` // Include version comparison
	Insecure       bool   `json:"insecure"`       // Skip TLS certificate verification
}

// TraceRequest represents the request payload for trace capture
type TraceRequest struct {
	Alias           string   `json:"alias"`
	Duration        string   `json:"duration"`
	StatusCodes     []int    `json:"statusCodes"`
	ErrorContains   []string `json:"errorContains"`
	GroupByAPI      bool     `json:"groupByApi"`
	GroupByClient   bool     `json:"groupByClient"`
	GroupByVersions bool     `json:"groupByVersions"`
	Insecure        bool     `json:"insecure"` // Skip TLS certificate verification
}

// ProfileRequest represents the request payload for profile capture
type ProfileRequest struct {
	Alias       string `json:"alias"`
	Duration    string `json:"duration"`
	ProfileType string `json:"profileType"` // cpu,mem,block,mutex,goroutines
	Insecure    bool   `json:"insecure"`    // Skip TLS certificate verification
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

	result, err := h.operationsService.CompareBuckets(req.SourceAlias, req.DestAlias, req.Path, req.CompareVersion, req.Insecure)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.RespondJSON(w, result)
}

// HandleValidate handles POST /api/operations/validate
func (h *OperationsHandler) HandleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	result, err := h.operationsService.ConfigurationValidation()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.RespondJSON(w, result)
}

// HandleValidateBucketConfig handles POST /api/operations/validate-bucket-config
func (h *OperationsHandler) HandleValidateBucketConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Aliases        []string `json:"aliases"`
		Buckets        []string `json:"buckets"`
		CheckLifecycle bool     `json:"check_lifecycle"`
		CheckEvents    bool     `json:"check_events"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	if len(req.Aliases) == 0 {
		h.RespondError(w, http.StatusBadRequest, "At least one alias is required")
		return
	}

	if len(req.Buckets) == 0 {
		h.RespondError(w, http.StatusBadRequest, "At least one bucket is required")
		return
	}

	result, err := h.operationsService.ValidateBucketConfiguration(req.Aliases, req.Buckets, req.CheckLifecycle, req.CheckEvents)
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

// HandleGetBucketVersioning handles GET /api/operations/bucket-versioning?sourceAlias=<alias>&destAlias=<alias>&bucket=<bucket>
func (h *OperationsHandler) HandleGetBucketVersioning(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	sourceAlias := r.URL.Query().Get("sourceAlias")
	destAlias := r.URL.Query().Get("destAlias")
	bucket := r.URL.Query().Get("bucket")

	if sourceAlias == "" || destAlias == "" || bucket == "" {
		h.RespondError(w, http.StatusBadRequest, "sourceAlias, destAlias, and bucket parameters are required")
		return
	}

	// Check versioning status for both aliases
	sourceVersioning, err := h.operationsService.GetBucketVersioningStatus(sourceAlias, bucket)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check source versioning: %v", err))
		return
	}

	destVersioning, err := h.operationsService.GetBucketVersioningStatus(destAlias, bucket)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to check destination versioning: %v", err))
		return
	}

	result := map[string]interface{}{
		"sourceAlias":         sourceAlias,
		"destAlias":           destAlias,
		"bucket":              bucket,
		"sourceVersioning":    sourceVersioning,
		"destVersioning":      destVersioning,
		"bothVersioned":       sourceVersioning && destVersioning,
		"versioningSupported": sourceVersioning || destVersioning,
	}

	h.RespondJSON(w, result)
}

// HandleTrace handles POST /api/operations/trace
func (h *OperationsHandler) HandleTrace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req TraceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Alias) == "" {
		h.RespondError(w, http.StatusBadRequest, "Alias is required")
		return
	}

	durationStr := strings.TrimSpace(req.Duration)
	if durationStr == "" {
		durationStr = "10s"
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid duration: %v", err))
		return
	}

	if duration < time.Second || duration > 5*time.Minute {
		h.RespondError(w, http.StatusBadRequest, "Duration must be between 1s and 5m")
		return
	}

	cleanErrors := make([]string, 0, len(req.ErrorContains))
	for _, filter := range req.ErrorContains {
		trimmed := strings.TrimSpace(filter)
		if trimmed != "" {
			cleanErrors = append(cleanErrors, trimmed)
		}
	}

	options := services.TraceCaptureOptions{
		Alias:           strings.TrimSpace(req.Alias),
		Duration:        duration,
		StatusCodes:     req.StatusCodes,
		ErrorFilters:    cleanErrors,
		GroupByAPI:      req.GroupByAPI,
		GroupByClient:   req.GroupByClient,
		GroupByVersions: req.GroupByVersions,
		Insecure:        req.Insecure,
	}

	result, err := h.operationsService.RunTraceCapture(options)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.RespondJSON(w, result)
}

// HandleProfile handles POST /api/operations/profile
func (h *OperationsHandler) HandleProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req ProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Alias) == "" {
		h.RespondError(w, http.StatusBadRequest, "Alias is required")
		return
	}

	durationStr := strings.TrimSpace(req.Duration)
	if durationStr == "" {
		durationStr = "30s"
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid duration: %v", err))
		return
	}

	if duration < time.Second || duration > 5*time.Minute {
		h.RespondError(w, http.StatusBadRequest, "Duration must be between 1s and 5m")
		return
	}

	profileType := strings.TrimSpace(req.ProfileType)
	if profileType == "" {
		profileType = "cpu,mem"
	}

	options := services.ProfileCaptureOptions{
		Alias:       strings.TrimSpace(req.Alias),
		Duration:    duration,
		ProfileType: profileType,
		Insecure:    req.Insecure,
	}

	result, err := h.operationsService.RunProfileCapture(options)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.RespondJSON(w, result)
}
