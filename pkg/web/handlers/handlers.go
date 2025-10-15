package handlers

import (
	"embed"

	"github.com/liamdn8/mc-tool/pkg/web/models"
	"github.com/liamdn8/mc-tool/pkg/web/services"
)

// Handlers aggregates all handler types
type Handlers struct {
	System      *SystemHandler
	Site        *SiteHandler
	Bucket      *BucketHandler
	Replication *ReplicationHandler
	Job         *JobHandler
	Analysis    *AnalysisHandler
	Operations  *OperationsHandler
}

// NewHandlers creates and initializes all handlers
func NewHandlers(executablePath string, staticFiles embed.FS, minioService *services.MinIOService, replicationService *services.ReplicationService, jobManager *models.JobManager) *Handlers {
	operationsService := services.NewOperationsService(minioService, replicationService)

	return &Handlers{
		System:      NewSystemHandler(executablePath, staticFiles, jobManager, minioService),
		Site:        NewSiteHandler(minioService),
		Bucket:      NewBucketHandler(minioService),
		Replication: NewReplicationHandler(replicationService, minioService),
		Job:         NewJobHandler(jobManager),
		Analysis:    NewAnalysisHandler(executablePath, jobManager),
		Operations:  NewOperationsHandler(operationsService),
	}
}
