package web

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/liamdn8/mc-tool/pkg/config"
	"github.com/liamdn8/mc-tool/pkg/logger"
	"github.com/liamdn8/mc-tool/pkg/web/handlers"
	"github.com/liamdn8/mc-tool/pkg/web/middleware"
	"github.com/liamdn8/mc-tool/pkg/web/models"
	"github.com/liamdn8/mc-tool/pkg/web/services"
)

//go:embed static/build/*
var staticFiles embed.FS

// Server represents the web UI server
type Server struct {
	config             *config.WebConfig
	httpServer         *http.Server
	executablePath     string
	minioService       *services.MinIOService
	replicationService *services.ReplicationService
	terminalService    *services.TerminalService
	jobManager         *models.JobManager
	handlers           *handlers.Handlers
}

// NewServer creates a new web server
func NewServer(cfg *config.WebConfig) *Server {
	// Auto-detect mc-tool executable path
	execPath := services.FindMCToolExecutable()

	logger.GetLogger().Info("Using mc-tool executable", map[string]interface{}{
		"path": execPath,
	})

	// Initialize services
	jobManager := models.NewJobManager()
	minioService := services.NewMinIOService(execPath)
	replicationService := services.NewReplicationService(minioService)
	terminalService := services.NewTerminalService()

	// Initialize handlers
	handlersInstance := handlers.NewHandlers(execPath, staticFiles, minioService, replicationService, jobManager, terminalService)

	return &Server{
		config:             cfg,
		executablePath:     execPath,
		minioService:       minioService,
		replicationService: replicationService,
		terminalService:    terminalService,
		jobManager:         jobManager,
		handlers:           handlersInstance,
	}
}

// Start starts the web server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Serve static files from React build
	staticFS, err := fs.Sub(staticFiles, "static/build")
	if err != nil {
		return fmt.Errorf("failed to load static files: %w", err)
	}

	// Base path for the application
	basePath := "/minio-webtool"

	// Redirect root to base path
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, basePath+"/", http.StatusMovedPermanently)
			return
		}
		http.NotFound(w, r)
	})

	// Serve new site replication UI by default
	mux.HandleFunc(basePath+"/", s.handlers.System.HandleIndex)

	// Custom static file handler with proper MIME types
	staticHandler := http.StripPrefix(basePath+"/static/", http.FileServer(http.FS(staticFS)))
	mux.HandleFunc(basePath+"/static/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Set correct MIME type based on file extension
		if strings.HasSuffix(path, ".js") {
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		} else if strings.HasSuffix(path, ".css") {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		} else if strings.HasSuffix(path, ".html") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		} else if strings.HasSuffix(path, ".json") {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		} else if strings.HasSuffix(path, ".png") {
			w.Header().Set("Content-Type", "image/png")
		} else if strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") {
			w.Header().Set("Content-Type", "image/jpeg")
		} else if strings.HasSuffix(path, ".svg") {
			w.Header().Set("Content-Type", "image/svg+xml")
		}

		// Serve the file
		staticHandler.ServeHTTP(w, r)
	})

	// System endpoints
	mux.HandleFunc(basePath+"/healthz", s.handlers.System.HandleHealthz)
	mux.HandleFunc(basePath+"/api/health", s.handlers.System.HandleHealth)
	mux.HandleFunc(basePath+"/api/mc-config", s.handlers.System.HandleMCConfig)

	// Site endpoints
	mux.HandleFunc(basePath+"/api/aliases", s.handlers.Site.HandleGetAliases)
	mux.HandleFunc(basePath+"/api/aliases-stats", s.handlers.Site.HandleGetAliasesWithStats)
	mux.HandleFunc(basePath+"/api/alias-health", s.handlers.Site.HandleAliasHealth)
	mux.HandleFunc(basePath+"/api/alias-health-fast", s.handlers.Site.HandleAliasHealthFast)
	mux.HandleFunc(basePath+"/api/sites", s.handlers.Site.HandleSites)
	mux.HandleFunc(basePath+"/api/sites/health", s.handlers.Site.HandleSiteHealth)

	// Bucket endpoints
	mux.HandleFunc(basePath+"/api/buckets", s.handlers.Bucket.HandleGetBuckets)
	mux.HandleFunc(basePath+"/api/bucket-stats", s.handlers.Bucket.HandleGetBucketStats)

	// Analysis endpoints
	mux.HandleFunc(basePath+"/api/compare", s.handlers.Analysis.HandleCompare)
	mux.HandleFunc(basePath+"/api/analyze", s.handlers.Analysis.HandleAnalyze)
	mux.HandleFunc(basePath+"/api/profile", s.handlers.Analysis.HandleProfile)
	mux.HandleFunc(basePath+"/api/validate", s.handlers.Analysis.HandleValidate)

	// Job endpoints
	mux.HandleFunc(basePath+"/api/jobs/", s.handlers.System.HandleJobStatus)

	// Site Replication APIs
	mux.HandleFunc(basePath+"/api/replication/info", s.handlers.Replication.HandleReplicationInfo)
	mux.HandleFunc(basePath+"/api/replication/status", s.handlers.Replication.HandleReplicationStatus)
	mux.HandleFunc(basePath+"/api/replication/compare", s.handlers.Replication.HandleReplicationCompare)
	mux.HandleFunc(basePath+"/api/replication/split-brain-check", s.handlers.Replication.HandleSplitBrainCheck)

	// Site Replication Management APIs
	mux.HandleFunc(basePath+"/api/replication/add", s.handlers.Replication.HandleReplicationAdd)
	mux.HandleFunc(basePath+"/api/replication/add-smart", s.handlers.Replication.HandleReplicationAddSmart)
	mux.HandleFunc(basePath+"/api/replication/remove", s.handlers.Replication.HandleReplicationRemove)
	mux.HandleFunc(basePath+"/api/replication/remove-site", s.handlers.Replication.HandleReplicationRemoveSite)
	mux.HandleFunc(basePath+"/api/replication/remove-site-smart", s.handlers.Replication.HandleReplicationRemoveSiteSmart)
	mux.HandleFunc(basePath+"/api/replication/resync", s.handlers.Replication.HandleReplicationResync)

	// Operations APIs
	mux.HandleFunc(basePath+"/api/operations/resync/options", s.handlers.Operations.HandleGetResyncOptions)
	mux.HandleFunc(basePath+"/api/operations/resync/start", s.handlers.Operations.HandleStartResync)
	mux.HandleFunc(basePath+"/api/operations/resync/status", s.handlers.Operations.HandleGetResyncStatus)
	mux.HandleFunc(basePath+"/api/operations/sync-policies", s.handlers.Operations.HandleSyncPolicies)
	mux.HandleFunc(basePath+"/api/operations/sync-lifecycle", s.handlers.Operations.HandleSyncLifecycle)
	mux.HandleFunc(basePath+"/api/operations/validate-consistency", s.handlers.Operations.HandleValidateConsistency)
	mux.HandleFunc(basePath+"/api/operations/health-check", s.handlers.Operations.HandleHealthCheck)
	mux.HandleFunc(basePath+"/api/operations/compare", s.handlers.Operations.HandleCompare)
	mux.HandleFunc(basePath+"/api/operations/validate", s.handlers.Operations.HandleValidate)
	mux.HandleFunc(basePath+"/api/operations/validate-bucket-config", s.handlers.Operations.HandleValidateBucketConfig)
	mux.HandleFunc(basePath+"/api/operations/buckets", s.handlers.Operations.HandleGetBuckets)
	mux.HandleFunc(basePath+"/api/operations/path-suggestions", s.handlers.Operations.HandleGetPathSuggestions)
	mux.HandleFunc(basePath+"/api/operations/bucket-versioning", s.handlers.Operations.HandleGetBucketVersioning)
	mux.HandleFunc(basePath+"/api/operations/trace", s.handlers.Operations.HandleTrace)
	mux.HandleFunc(basePath+"/api/operations/profile", s.handlers.Operations.HandleProfile)

	// Infrastructure Validation APIs
	mux.HandleFunc(basePath+"/api/validate/infrastructure", s.handlers.InfraValidation.HandleInfraValidate)
	mux.HandleFunc(basePath+"/api/validate/infrastructure/vims", s.handlers.InfraValidation.HandleGetInfraVIMs)
	mux.HandleFunc(basePath+"/api/validate/infrastructure/namespaces", s.handlers.InfraValidation.HandleGetNamespaces)
	mux.HandleFunc(basePath+"/api/validate/infrastructure/history", s.handlers.InfraValidation.HandleGetInfraHistory)
	mux.HandleFunc(basePath+"/api/validate/infrastructure/diff", s.handlers.InfraValidation.HandleGetDiff)

	// Perftest APIs
	mux.HandleFunc(basePath+"/api/perftest/start", s.handlers.Perftest.HandleStartTest)
	mux.HandleFunc(basePath+"/api/perftest/status", s.handlers.Perftest.HandleGetStatus)
	mux.HandleFunc(basePath+"/api/perftest/result", s.handlers.Perftest.HandleGetResult)

	// Terminal APIs
	mux.HandleFunc(basePath+"/api/terminal/ws", s.handlers.Terminal.HandleWebsocket)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		Handler:      middleware.CORS(middleware.Logging(mux)),
		ReadTimeout:  10 * time.Minute, // Long timeout for operations like profile, trace
		WriteTimeout: 10 * time.Minute, // Long timeout for streaming responses
		IdleTimeout:  60 * time.Second,
	}

	logger.GetLogger().Info("Starting web UI server", map[string]interface{}{
		"port": s.config.Port,
		"url":  fmt.Sprintf("http://localhost:%d", s.config.Port),
	})
	return s.httpServer.ListenAndServe()
}

// Stop stops the web server
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
