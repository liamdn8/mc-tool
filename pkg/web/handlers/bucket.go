package handlers

import (
	"fmt"
	"net/http"

	"github.com/liamdn8/mc-tool/pkg/web/services"
)

// BucketHandler handles bucket-related requests
type BucketHandler struct {
	BaseHandler
	minioService *services.MinIOService
}

// NewBucketHandler creates a new bucket handler
func NewBucketHandler(minioService *services.MinIOService) *BucketHandler {
	return &BucketHandler{
		minioService: minioService,
	}
}

// HandleBuckets handles GET /api/buckets
func (h *BucketHandler) HandleBuckets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	aliasName := r.URL.Query().Get("alias")
	if aliasName == "" {
		h.RespondError(w, http.StatusBadRequest, "Alias parameter is required")
		return
	}

	buckets, err := h.minioService.ListBuckets(aliasName)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list buckets: %v", err))
		return
	}

	h.RespondJSON(w, map[string]interface{}{
		"buckets": buckets,
	})
}

// HandleGetBuckets handles GET /api/buckets (alternative endpoint)
func (h *BucketHandler) HandleGetBuckets(w http.ResponseWriter, r *http.Request) {
	h.HandleBuckets(w, r) // Delegate to the main handler
}

// HandleBucketStats handles GET /api/bucket-stats
func (h *BucketHandler) HandleBucketStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	aliasName := r.URL.Query().Get("alias")
	bucketName := r.URL.Query().Get("bucket")

	if aliasName == "" || bucketName == "" {
		h.RespondError(w, http.StatusBadRequest, "Both alias and bucket parameters are required")
		return
	}

	stats := h.minioService.GetBucketStats(aliasName, bucketName)

	h.RespondJSON(w, stats)
}

// HandleGetBucketStats handles GET /api/bucket-stats (alternative endpoint)
func (h *BucketHandler) HandleGetBucketStats(w http.ResponseWriter, r *http.Request) {
	h.HandleBucketStats(w, r) // Delegate to the main handler
}
