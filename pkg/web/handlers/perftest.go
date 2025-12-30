package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/liamdn8/mc-tool/pkg/perftest"
	"github.com/liamdn8/mc-tool/pkg/web/services"
)

// PerftestHandler handles perftest related requests
type PerftestHandler struct {
	service *services.PerftestService
}

// NewPerftestHandler creates a new perftest handler
func NewPerftestHandler(service *services.PerftestService) *PerftestHandler {
	return &PerftestHandler{
		service: service,
	}
}

// StartTestRequest represents a request to start a test
type StartTestRequest struct {
	SiteAlias      string `json:"site_alias"`
	Bucket         string `json:"bucket"`
	ObjectPath     string `json:"object_path"`
	ObjectSizeType string `json:"object_size_type"`
	ObjectCount    int    `json:"object_count"`
	OverrideCount  int    `json:"override_count"`
	UploadMode     string `json:"upload_mode"`
	UploadInterval string `json:"upload_interval"`
	Iterations     int    `json:"iterations"`
	Parallelism    int    `json:"parallelism"`
	Insecure       bool   `json:"insecure"`
}

// HandleStartTest starts a new performance test
func (h *PerftestHandler) HandleStartTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StartTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.SiteAlias == "" || req.Bucket == "" {
		http.Error(w, "site_alias and bucket are required", http.StatusBadRequest)
		return
	}

	// Auto-generate path if not specified
	if req.ObjectPath == "" {
		timestamp := time.Now().Format("20060102-150405")
		req.ObjectPath = "mc-test/" + timestamp + "/"
	}

	// Parse size type
	var sizeType perftest.ObjectSizeType
	switch req.ObjectSizeType {
	case "small":
		sizeType = perftest.ObjectSizeSmall
	case "medium":
		sizeType = perftest.ObjectSizeMedium
	case "large":
		sizeType = perftest.ObjectSizeLarge
	default:
		sizeType = perftest.ObjectSizeSmall
	}

	// Parse upload mode
	var uploadMode perftest.UploadMode
	switch req.UploadMode {
	case "interval":
		uploadMode = perftest.UploadModeInterval
	default:
		uploadMode = perftest.UploadModeAll
	}

	// Parse interval
	uploadInterval, _ := time.ParseDuration(req.UploadInterval)
	if uploadInterval == 0 {
		uploadInterval = 5 * time.Second
	}

	// Set defaults
	if req.ObjectCount <= 0 {
		req.ObjectCount = 10
	}
	if req.Parallelism <= 0 {
		req.Parallelism = 5
	}
	if req.Iterations <= 0 {
		req.Iterations = 5
	}

	// Create test config
	cfg := &perftest.TestConfig{
		SiteAlias:      req.SiteAlias,
		Bucket:         req.Bucket,
		ObjectPath:     req.ObjectPath,
		ObjectSizeType: sizeType,
		ObjectCount:    req.ObjectCount,
		OverrideCount:  req.OverrideCount,
		UploadMode:     uploadMode,
		UploadInterval: uploadInterval,
		Iterations:     req.Iterations,
		Parallelism:    req.Parallelism,
		Insecure:       req.Insecure,
	}

	// Start test
	if err := h.service.StartTest(cfg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Test started successfully",
	})
}

// HandleGetStatus returns current test status
func (h *PerftestHandler) HandleGetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	running, status := h.service.GetStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"running": running,
		"status":  status,
	})
}

// HandleGetResult returns the last test result
func (h *PerftestHandler) HandleGetResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result := h.service.GetLastResult()
	if result == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "No test results available",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"result":  result,
	})
}
