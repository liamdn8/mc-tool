package handlers

import (
	"github.com/liamdn8/mc-tool/pkg/web/services"
)

// SiteHandler handles site-related requests
type SiteHandler struct {
	BaseHandler
	minioService *services.MinIOService
}

// NewSiteHandler creates a new site handler
func NewSiteHandler(minioService *services.MinIOService) *SiteHandler {
	return &SiteHandler{
		minioService: minioService,
	}
}

// Additional site handler logic is split into dedicated files for readability.
