package handlers

import (
	"embed"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/liamdn8/mc-tool/pkg/web/models"
	"github.com/liamdn8/mc-tool/pkg/web/services"
)

// SystemHandler handles system-related requests like health, index, etc.
type SystemHandler struct {
	BaseHandler
	executablePath string
	staticFiles    embed.FS
	jobManager     *models.JobManager
	minioService   *services.MinIOService
}

// NewSystemHandler creates a new system handler
func NewSystemHandler(executablePath string, staticFiles embed.FS, jobManager *models.JobManager, minioService *services.MinIOService) *SystemHandler {
	return &SystemHandler{
		executablePath: executablePath,
		staticFiles:    staticFiles,
		jobManager:     jobManager,
		minioService:   minioService,
	}
}

// HandleIndex serves the React app index.html
func (h *SystemHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	indexHTML, err := h.staticFiles.ReadFile("static/build/index.html")
	if err != nil {
		http.Error(w, "Failed to load index page", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML)
}

// HandleHealth handles GET /api/health
func (h *SystemHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	h.RespondJSON(w, map[string]interface{}{
		"status":  "ok",
		"version": "1.0.0",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// HandleHealthz handles GET /healthz for container health checks
func (h *SystemHandler) HandleHealthz(w http.ResponseWriter, r *http.Request) {
	// Check if mc command is available
	mcAvailable := false
	if h.executablePath != "" {
		cmd := exec.Command(h.executablePath, "version")
		if err := cmd.Run(); err == nil {
			mcAvailable = true
		}
	}

	// Return 200 if healthy, 503 if unhealthy
	statusCode := http.StatusOK
	status := "healthy"
	if !mcAvailable {
		statusCode = http.StatusServiceUnavailable
		status = "unhealthy"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       status,
		"timestamp":    time.Now().Format(time.RFC3339),
		"mc_available": mcAvailable,
	})
}

// HandleMCConfig handles GET /api/mc-config
func (h *SystemHandler) HandleMCConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Check if mc is configured
	cmd := exec.Command("mc", "alias", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		h.RespondJSON(w, map[string]interface{}{
			"configured": false,
			"message":    "MinIO client (mc) is not configured or not installed",
		})
		return
	}

	h.RespondJSON(w, map[string]interface{}{
		"configured": true,
		"output":     string(output),
	})
}

// HandleJobStatus handles GET /api/jobs/:id
func (h *SystemHandler) HandleJobStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	jobID := strings.TrimPrefix(r.URL.Path, "/api/jobs/")
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
