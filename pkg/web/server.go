package web

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
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
	jobManager         *models.JobManager
	handlers           *handlers.Handlers
}

// NewServer creates a new web server
func NewServer(cfg *config.WebConfig) *Server {
	// Get the current executable path
	execPath, err := os.Executable()
	if err != nil {
		execPath = "mc-tool" // fallback to PATH lookup
	}

	// Initialize services
	jobManager := models.NewJobManager()
	minioService := services.NewMinIOService(execPath)
	replicationService := services.NewReplicationService(minioService)

	// Initialize handlers
	handlersInstance := handlers.NewHandlers(execPath, staticFiles, minioService, replicationService, jobManager)

	return &Server{
		config:             cfg,
		executablePath:     execPath,
		minioService:       minioService,
		replicationService: replicationService,
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

	// Serve new site replication UI by default
	mux.HandleFunc("/", s.handlers.System.HandleIndex)

	// Custom static file handler with proper MIME types
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.FS(staticFS)))
	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/healthz", s.handlers.System.HandleHealthz)
	mux.HandleFunc("/api/health", s.handlers.System.HandleHealth)
	mux.HandleFunc("/api/mc-config", s.handlers.System.HandleMCConfig)

	// Site endpoints
	mux.HandleFunc("/api/aliases", s.handlers.Site.HandleGetAliases)
	mux.HandleFunc("/api/aliases-stats", s.handlers.Site.HandleGetAliasesWithStats)
	mux.HandleFunc("/api/alias-health", s.handlers.Site.HandleAliasHealth)
	mux.HandleFunc("/api/sites", s.handlers.Site.HandleSites)
	mux.HandleFunc("/api/sites/health", s.handlers.Site.HandleSiteHealth)

	// Bucket endpoints
	mux.HandleFunc("/api/buckets", s.handlers.Bucket.HandleGetBuckets)
	mux.HandleFunc("/api/bucket-stats", s.handlers.Bucket.HandleGetBucketStats)

	// Analysis endpoints
	mux.HandleFunc("/api/compare", s.handlers.Analysis.HandleCompare)
	mux.HandleFunc("/api/analyze", s.handlers.Analysis.HandleAnalyze)
	mux.HandleFunc("/api/profile", s.handlers.Analysis.HandleProfile)
	mux.HandleFunc("/api/checklist", s.handlers.Analysis.HandleChecklist)

	// Job endpoints
	mux.HandleFunc("/api/jobs/", s.handlers.System.HandleJobStatus)

	// Site Replication APIs
	mux.HandleFunc("/api/replication/info", s.handlers.Replication.HandleReplicationInfo)
	mux.HandleFunc("/api/replication/status", s.handlers.Replication.HandleReplicationStatus)
	mux.HandleFunc("/api/replication/compare", s.handlers.Replication.HandleReplicationCompare)
	mux.HandleFunc("/api/replication/split-brain-check", s.handlers.Replication.HandleSplitBrainCheck)

	// Site Replication Management APIs
	mux.HandleFunc("/api/replication/add", s.handlers.Replication.HandleReplicationAdd)
	mux.HandleFunc("/api/replication/add-smart", s.handlers.Replication.HandleReplicationAddSmart)
	mux.HandleFunc("/api/replication/remove", s.handlers.Replication.HandleReplicationRemove)
	mux.HandleFunc("/api/replication/remove-site", s.handlers.Replication.HandleReplicationRemoveSite)
	mux.HandleFunc("/api/replication/remove-site-smart", s.handlers.Replication.HandleReplicationRemoveSiteSmart)
	mux.HandleFunc("/api/replication/resync", s.handlers.Replication.HandleReplicationResync)

	// Operations APIs
	mux.HandleFunc("/api/operations/sync-policies", s.handlers.Operations.HandleSyncPolicies)
	mux.HandleFunc("/api/operations/sync-lifecycle", s.handlers.Operations.HandleSyncLifecycle)
	mux.HandleFunc("/api/operations/validate-consistency", s.handlers.Operations.HandleValidateConsistency)
	mux.HandleFunc("/api/operations/health-check", s.handlers.Operations.HandleHealthCheck)
	mux.HandleFunc("/api/operations/compare", s.handlers.Operations.HandleCompare)
	mux.HandleFunc("/api/operations/checklist", s.handlers.Operations.HandleChecklist)
	mux.HandleFunc("/api/operations/buckets", s.handlers.Operations.HandleGetBuckets)
	mux.HandleFunc("/api/operations/path-suggestions", s.handlers.Operations.HandleGetPathSuggestions)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		Handler:      middleware.CORS(middleware.Logging(mux)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
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
