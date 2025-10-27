package handlers

import "github.com/liamdn8/mc-tool/pkg/web/services"

// ReplicationHandler handles replication-related requests
type ReplicationHandler struct {
	BaseHandler
	replicationService *services.ReplicationService
	minioService       *services.MinIOService
}

// NewReplicationHandler creates a new replication handler
func NewReplicationHandler(replicationService *services.ReplicationService, minioService *services.MinIOService) *ReplicationHandler {
	return &ReplicationHandler{
		replicationService: replicationService,
		minioService:       minioService,
	}
}

// Additional replication handler logic is organized across dedicated files.
