package handlers

import (
	"net/http"
)

// HandleAliasHealth handles GET /api/alias-health
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

	health := h.minioService.GetAliasHealth(alias)
	h.RespondJSON(w, health)
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
