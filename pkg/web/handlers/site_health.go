package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"time"
)

// HandleAliasHealth handles GET /api/alias-health?alias=xxx
func (h *SiteHandler) HandleAliasHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	alias := r.URL.Query().Get("alias")
	if alias == "" {
		h.RespondError(w, http.StatusBadRequest, "Alias parameter is required")
		return
	}

	health := h.checkAliasHealthWithTimeout(alias, 5*time.Second)
	h.RespondJSON(w, health)
}

// checkAliasHealthWithTimeout checks alias health with timeout
func (h *SiteHandler) checkAliasHealthWithTimeout(alias string, timeout time.Duration) map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resultChan := make(chan map[string]interface{}, 1)

	go func() {
		healthy, status := h.getAliasHealthStatus(alias)
		resultChan <- map[string]interface{}{
			"alias":   alias,
			"healthy": healthy,
			"status":  status,
		}
	}()

	select {
	case result := <-resultChan:
		return result
	case <-ctx.Done():
		return map[string]interface{}{
			"alias":   alias,
			"healthy": false,
			"status":  "timeout",
		}
	}
}

// HandleAliasHealthFast handles GET /api/aliases/:alias/health (new fast endpoint)
func (h *SiteHandler) HandleAliasHealthFast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	alias := r.URL.Query().Get("alias")
	if alias == "" {
		h.RespondError(w, http.StatusBadRequest, "Alias parameter is required")
		return
	}

	// Use shorter timeout for fast check
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	healthy, status, err := h.fastHealthCheck(ctx, alias)

	h.RespondJSON(w, map[string]interface{}{
		"alias":   alias,
		"healthy": healthy,
		"status":  status,
		"error":   err,
	})
}

// fastHealthCheck performs a quick health check with context
func (h *SiteHandler) fastHealthCheck(ctx context.Context, alias string) (bool, string, string) {
	// Try mc admin info first (more reliable)
	cmd := exec.CommandContext(ctx, "mc", "admin", "info", alias, "--json")
	output, err := cmd.CombinedOutput()

	if err == nil {
		var result map[string]interface{}
		if json.Unmarshal(output, &result) == nil {
			if status, ok := result["status"].(string); ok && status == "success" {
				return true, "healthy", ""
			}
		}
	}

	// Fallback to mc ls
	cmd = exec.CommandContext(ctx, "mc", "ls", alias)
	if err := cmd.Run(); err == nil {
		return true, "healthy", ""
	}

	// Check if context was cancelled (timeout)
	if ctx.Err() == context.DeadlineExceeded {
		return false, "timeout", "Connection timeout"
	}

	return false, "unhealthy", "Connection failed"
}

// HandleSiteHealth handles GET /api/sites/health
func (h *SiteHandler) HandleSiteHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	aliases, err := h.minioService.GetAliases()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to get aliases")
		return
	}

	healthData := make(map[string]interface{})
	for _, alias := range aliases {
		healthData[alias.Name] = h.minioService.GetAliasHealth(alias.Name)
	}

	h.RespondJSON(w, healthData)
}
