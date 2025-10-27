package handlers

import (
	"fmt"
	"net/http"
)

// HandleSites handles GET /api/sites (alias list)
func (h *SiteHandler) HandleSites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	aliases, err := h.minioService.GetAliases()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to get aliases")
		return
	}

	h.RespondJSON(w, map[string]interface{}{
		"sites": aliases,
	})
}

// HandleGetAliases handles GET /api/aliases
func (h *SiteHandler) HandleGetAliases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	aliases, err := h.getMCInternalAliases()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get aliases: %v", err))
		return
	}

	h.RespondJSON(w, map[string]interface{}{
		"aliases": aliases,
	})
}

// HandleGetAliasesWithStats handles GET /api/aliases-stats
func (h *SiteHandler) HandleGetAliasesWithStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	aliases, err := h.getMCInternalAliases()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get aliases: %v", err))
		return
	}

	var aliasesWithStats []map[string]interface{}
	for _, alias := range aliases {
		aliasName, ok := alias["name"].(string)
		if !ok {
			continue
		}

		stats := h.getAliasStats(aliasName)

		aliasData := map[string]interface{}{
			"name":          alias["name"],
			"url":           alias["url"],
			"healthy":       alias["healthy"],
			"status":        alias["status"],
			"bucket_count":  stats["bucket_count"],
			"total_size":    stats["total_size"],
			"total_objects": stats["total_objects"],
		}

		if accessKey, ok := alias["accessKey"]; ok {
			aliasData["accessKey"] = accessKey
		}
		if api, ok := alias["api"]; ok {
			aliasData["api"] = api
		}

		aliasesWithStats = append(aliasesWithStats, aliasData)
	}

	h.RespondJSON(w, map[string]interface{}{
		"aliases": aliasesWithStats,
	})
}
