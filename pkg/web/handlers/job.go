package handlers

import (
	"net/http"

	"github.com/liamdn8/mc-tool/pkg/web/models"
)

// JobHandler handles job-related requests
type JobHandler struct {
	BaseHandler
	jobManager *models.JobManager
}

// NewJobHandler creates a new job handler
func NewJobHandler(jobManager *models.JobManager) *JobHandler {
	return &JobHandler{
		jobManager: jobManager,
	}
}

// HandleJobStatus handles GET /api/jobs/:id
func (h *JobHandler) HandleJobStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract job ID from URL path
	path := r.URL.Path
	if len(path) < 11 { // "/api/jobs/" = 10 characters
		h.RespondError(w, http.StatusBadRequest, "Job ID is required")
		return
	}

	jobID := path[10:] // Remove "/api/jobs/" prefix
	if jobID == "" {
		h.RespondError(w, http.StatusBadRequest, "Job ID is required")
		return
	}

	job := h.jobManager.GetJob(jobID)
	if job == nil {
		h.RespondError(w, http.StatusNotFound, "Job not found")
		return
	}

	h.RespondJSON(w, job)
}

// HandleJobsList handles GET /api/jobs
func (h *JobHandler) HandleJobsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	jobs := h.jobManager.GetAllJobs()
	h.RespondJSON(w, map[string]interface{}{
		"jobs": jobs,
	})
}
