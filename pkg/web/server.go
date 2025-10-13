package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/liamdn8/mc-tool/pkg/profile"
)

//go:embed static/*
var staticFiles embed.FS

// Server represents the web UI server
type Server struct {
	port           int
	httpServer     *http.Server
	jobManager     *JobManager
	executablePath string
}

// JobManager manages background jobs
type JobManager struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

// Job represents a background operation
type Job struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Status    string                 `json:"status"` // pending, running, completed, failed
	Progress  int                    `json:"progress"`
	Message   string                 `json:"message"`
	Result    map[string]interface{} `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	StartTime time.Time              `json:"start_time"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Output    []string               `json:"output"`
	mu        sync.Mutex
}

// NewServer creates a new web server
func NewServer(port int) *Server {
	// Get the current executable path
	execPath, err := os.Executable()
	if err != nil {
		execPath = "mc-tool" // fallback to PATH lookup
	}

	return &Server{
		port:           port,
		executablePath: execPath,
		jobManager: &JobManager{
			jobs: make(map[string]*Job),
		},
	}
}

// Start starts the web server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Serve static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return fmt.Errorf("failed to load static files: %w", err)
	}
	// Serve new site replication UI by default
	mux.HandleFunc("/", s.handleIndex)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// API endpoints
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/aliases", s.handleGetAliases)
	mux.HandleFunc("/api/aliases-stats", s.handleGetAliasesWithStats)
	mux.HandleFunc("/api/alias-health", s.handleAliasHealth)
	mux.HandleFunc("/api/buckets", s.handleGetBuckets)
	mux.HandleFunc("/api/bucket-stats", s.handleGetBucketStats)
	mux.HandleFunc("/api/compare", s.handleCompare)
	mux.HandleFunc("/api/analyze", s.handleAnalyze)
	mux.HandleFunc("/api/profile", s.handleProfile)
	mux.HandleFunc("/api/checklist", s.handleChecklist)
	mux.HandleFunc("/api/jobs/", s.handleJobStatus)
	mux.HandleFunc("/api/mc-config", s.handleMCConfig)

	// Site Replication APIs
	mux.HandleFunc("/api/replication/info", s.handleReplicationInfo)
	mux.HandleFunc("/api/replication/status", s.handleReplicationStatus)
	mux.HandleFunc("/api/replication/compare", s.handleReplicationCompare)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.corsMiddleware(s.loggingMiddleware(mux)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting web UI server on http://localhost:%d", s.port)
	return s.httpServer.ListenAndServe()
}

// Stop stops the web server
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// Middleware
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// API Handlers
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	// Serve the new site replication UI
	indexHTML, err := staticFiles.ReadFile("static/index-new.html")
	if err != nil {
		http.Error(w, "Failed to load index page", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, map[string]interface{}{
		"status":  "ok",
		"version": "1.0.0",
		"time":    time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleGetAliases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	aliases, err := s.getMCInternalAliases()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get aliases: %v", err))
		return
	}

	s.respondJSON(w, map[string]interface{}{
		"aliases": aliases,
	})
}

func (s *Server) handleGetAliasesWithStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	aliases, err := s.getMCInternalAliases()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get aliases: %v", err))
		return
	}

	// Get stats for each alias
	var aliasesWithStats []map[string]interface{}
	for _, alias := range aliases {
		stats := s.getAliasStats(alias["name"])

		aliasData := map[string]interface{}{
			"name":          alias["name"],
			"url":           alias["url"],
			"bucket_count":  stats["bucket_count"],
			"total_size":    stats["total_size"],
			"total_objects": stats["total_objects"],
		}
		aliasesWithStats = append(aliasesWithStats, aliasData)
	}

	s.respondJSON(w, map[string]interface{}{
		"aliases": aliasesWithStats,
	})
}

func (s *Server) handleGetBuckets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	alias := r.URL.Query().Get("alias")
	if alias == "" {
		s.respondError(w, http.StatusBadRequest, "Alias parameter is required")
		return
	}

	buckets, err := s.listBuckets(alias)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list buckets: %v", err))
		return
	}

	s.respondJSON(w, map[string]interface{}{
		"buckets": buckets,
	})
}

func (s *Server) handleAliasHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	alias := r.URL.Query().Get("alias")
	if alias == "" {
		s.respondError(w, http.StatusBadRequest, "Alias parameter is required")
		return
	}

	// Try to ping the alias using mc admin info
	cmd := exec.Command("mc", "admin", "info", alias, "--json")
	output, err := cmd.CombinedOutput()

	healthy := false
	message := "Unknown"

	if err == nil {
		// Parse JSON output to check if server is responding
		var result map[string]interface{}
		if json.Unmarshal(output, &result) == nil {
			if result["status"] != nil {
				healthy = true
				message = "Connected"
			}
		}
	} else {
		// Try simple ls command as fallback
		cmd = exec.Command("mc", "ls", alias)
		if cmd.Run() == nil {
			healthy = true
			message = "Connected (limited)"
		} else {
			message = "Unreachable"
		}
	}

	s.respondJSON(w, map[string]interface{}{
		"healthy": healthy,
		"message": message,
	})
}

func (s *Server) handleCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
		Recursive   bool   `json:"recursive"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Create job
	job := s.jobManager.createJob("compare")
	go s.runCompareJob(job, req.Source, req.Destination, req.Recursive)

	s.respondJSON(w, map[string]interface{}{
		"job_id": job.ID,
		"status": "started",
	})
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Alias  string `json:"alias"`
		Bucket string `json:"bucket"`
		Prefix string `json:"prefix"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Create job
	job := s.jobManager.createJob("analyze")
	go s.runAnalyzeJob(job, req.Alias, req.Bucket, req.Prefix)

	s.respondJSON(w, map[string]interface{}{
		"job_id": job.ID,
		"status": "started",
	})
}

func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Alias           string `json:"alias"`
		ProfileType     string `json:"profile_type"`
		Duration        string `json:"duration"`
		DetectLeaks     bool   `json:"detect_leaks"`
		MonitorInterval string `json:"monitor_interval"`
		ThresholdMB     int    `json:"threshold_mb"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Create job
	job := s.jobManager.createJob("profile")
	go s.runProfileJob(job, req)

	s.respondJSON(w, map[string]interface{}{
		"job_id": job.ID,
		"status": "started",
	})
}

func (s *Server) handleChecklist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Alias  string `json:"alias"`
		Bucket string `json:"bucket"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Create job
	job := s.jobManager.createJob("checklist")
	go s.runChecklistJob(job, req.Alias, req.Bucket)

	s.respondJSON(w, map[string]interface{}{
		"job_id": job.ID,
		"status": "started",
	})
}

func (s *Server) handleJobStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	jobID := strings.TrimPrefix(r.URL.Path, "/api/jobs/")
	if jobID == "" {
		s.respondError(w, http.StatusBadRequest, "Job ID is required")
		return
	}

	job := s.jobManager.getJob(jobID)
	if job == nil {
		s.respondError(w, http.StatusNotFound, "Job not found")
		return
	}

	s.respondJSON(w, job)
}

func (s *Server) handleMCConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Check if mc is configured
	cmd := exec.Command("mc", "alias", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.respondJSON(w, map[string]interface{}{
			"configured": false,
			"message":    "MinIO client (mc) is not configured or not installed",
		})
		return
	}

	s.respondJSON(w, map[string]interface{}{
		"configured": true,
		"output":     string(output),
	})
}

// Site Replication Handlers
func (s *Server) handleReplicationInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get all aliases
	aliases, err := s.getMCInternalAliases()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get aliases: %v", err))
		return
	}

	if len(aliases) == 0 {
		s.respondJSON(w, map[string]interface{}{
			"enabled": false,
			"sites":   []interface{}{},
			"message": "No MinIO aliases configured",
		})
		return
	}

	// Check each alias for site replication configuration
	var sites []map[string]interface{}
	var replicationEnabled = false
	var replicationGroupInfo map[string]interface{}

	for _, alias := range aliases {
		siteInfo := map[string]interface{}{
			"alias":              alias["name"],
			"endpoint":           alias["url"],
			"healthy":            false,
			"replicationEnabled": false,
			"replicationStatus":  "not_configured",
			"deploymentID":       "",
			"siteName":           "",
		}

		// Check if site replication is enabled for this alias
		replicateCmd := exec.Command("mc", "admin", "replicate", "info", alias["name"], "--json")
		replicateOutput, replicateErr := replicateCmd.CombinedOutput()

		if replicateErr == nil {
			var replicateInfo map[string]interface{}
			if json.Unmarshal(replicateOutput, &replicateInfo) == nil {
				// Site replication is configured
				if enabled, ok := replicateInfo["enabled"].(bool); ok && enabled {
					siteInfo["replicationEnabled"] = true
					siteInfo["replicationStatus"] = "configured"
					replicationEnabled = true

					// Store replication group info from first enabled site
					if replicationGroupInfo == nil {
						replicationGroupInfo = replicateInfo
					}

					// Extract this site's information from the peer sites
					if sitesList, ok := replicateInfo["sites"].([]interface{}); ok {
						for _, peerSite := range sitesList {
							if peer, ok := peerSite.(map[string]interface{}); ok {
								// Try to match by endpoint
								if endpoint, ok := peer["endpoint"].(string); ok {
									if strings.Contains(alias["url"], endpoint) || strings.Contains(endpoint, alias["url"]) {
										if name, ok := peer["name"].(string); ok {
											siteInfo["siteName"] = name
										}
										if deployID, ok := peer["deploymentID"].(string); ok {
											siteInfo["deploymentID"] = deployID
										}
										break
									}
								}
							}
						}
					}
				} else {
					siteInfo["replicationStatus"] = "disabled"
				}
			}
		} else {
			siteInfo["replicationStatus"] = "not_configured"
		}

		// Get admin info for health check
		adminCmd := exec.Command("mc", "admin", "info", alias["name"], "--json")
		adminOutput, adminErr := adminCmd.CombinedOutput()

		if adminErr == nil {
			var adminInfo map[string]interface{}
			if json.Unmarshal(adminOutput, &adminInfo) == nil {
				if status, ok := adminInfo["status"].(string); ok {
					siteInfo["healthy"] = (status == "success")
				}

				if info, ok := adminInfo["info"].(map[string]interface{}); ok {
					// Get deployment ID if not already set from replication info
					if siteInfo["deploymentID"] == "" {
						if backend, ok := info["backend"].(map[string]interface{}); ok {
							if backendType, ok := backend["backendType"].(string); ok {
								siteInfo["backendType"] = backendType
							}
						}
					}

					// Get server count
					if servers, ok := info["servers"].([]interface{}); ok {
						siteInfo["serverCount"] = len(servers)
					}
				}
			}
		}

		sites = append(sites, siteInfo)
	}

	s.respondJSON(w, map[string]interface{}{
		"enabled":          replicationEnabled,
		"aliases":          sites,
		"totalAliases":     len(sites),
		"replicationGroup": replicationGroupInfo,
		"configuredSites":  len(sites),
	})
}

func (s *Server) handleReplicationStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	aliases, err := s.getMCInternalAliases()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get aliases: %v", err))
		return
	}

	status := make(map[string]interface{})
	status["status"] = "healthy"
	sites := make(map[string]interface{})

	for _, alias := range aliases {
		// Get bucket list
		cmd := exec.Command("mc", "ls", alias["name"], "--json")
		output, err := cmd.CombinedOutput()

		bucketCount := 0
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			for _, line := range lines {
				if line != "" {
					bucketCount++
				}
			}
		}

		sites[alias["name"]] = map[string]interface{}{
			"replicatedBuckets": bucketCount,
			"pendingObjects":    0,
			"failedObjects":     0,
			"lastSyncTime":      time.Now().Format(time.RFC3339),
			"healthy":           true,
		}
	}

	status["sites"] = sites
	s.respondJSON(w, status)
}

func (s *Server) handleReplicationCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	aliases, err := s.getMCInternalAliases()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get aliases: %v", err))
		return
	}

	if len(aliases) < 2 {
		s.respondJSON(w, map[string]interface{}{
			"buckets": map[string]interface{}{},
			"message": "Need at least 2 sites to compare",
		})
		return
	}

	// Collect all buckets from all sites
	type BucketInfo struct {
		Sites      []string
		Policy     map[string]string
		Lifecycle  map[string]interface{}
		Versioning map[string]string
	}

	allBuckets := make(map[string]*BucketInfo)

	for _, alias := range aliases {
		cmd := exec.Command("mc", "ls", alias["name"], "--json")
		output, err := cmd.CombinedOutput()
		if err != nil {
			continue
		}

		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}

			var bucketData map[string]interface{}
			if err := json.Unmarshal([]byte(line), &bucketData); err != nil {
				continue
			}

			bucketName := ""
			if key, ok := bucketData["key"].(string); ok {
				bucketName = strings.TrimSuffix(key, "/")
			}

			if bucketName == "" {
				continue
			}

			if _, exists := allBuckets[bucketName]; !exists {
				allBuckets[bucketName] = &BucketInfo{
					Sites:      []string{},
					Policy:     make(map[string]string),
					Lifecycle:  make(map[string]interface{}),
					Versioning: make(map[string]string),
				}
			}

			allBuckets[bucketName].Sites = append(allBuckets[bucketName].Sites, alias["name"])

			// Get bucket policy
			policyCmd := exec.Command("mc", "anonymous", "get", fmt.Sprintf("%s/%s", alias["name"], bucketName))
			policyOutput, _ := policyCmd.CombinedOutput()
			allBuckets[bucketName].Policy[alias["name"]] = string(policyOutput)

			// Get lifecycle
			ilmCmd := exec.Command("mc", "ilm", "ls", fmt.Sprintf("%s/%s", alias["name"], bucketName), "--json")
			ilmOutput, _ := ilmCmd.CombinedOutput()
			if ilmOutput != nil {
				var ilmData interface{}
				json.Unmarshal(ilmOutput, &ilmData)
				allBuckets[bucketName].Lifecycle[alias["name"]] = ilmData
			}

			// Get versioning
			versionCmd := exec.Command("mc", "version", "info", fmt.Sprintf("%s/%s", alias["name"], bucketName), "--json")
			versionOutput, _ := versionCmd.CombinedOutput()
			allBuckets[bucketName].Versioning[alias["name"]] = string(versionOutput)
		}
	}

	// Build comparison result
	result := make(map[string]interface{})
	for bucketName, info := range allBuckets {
		bucketResult := map[string]interface{}{
			"existsOn": info.Sites,
			"policy": map[string]interface{}{
				"consistent": s.checkConsistency(info.Policy),
				"values":     info.Policy,
			},
			"lifecycle": map[string]interface{}{
				"consistent": s.checkConsistency(info.Lifecycle),
				"values":     info.Lifecycle,
			},
			"versioning": map[string]interface{}{
				"consistent": s.checkConsistency(info.Versioning),
				"values":     info.Versioning,
			},
		}
		result[bucketName] = bucketResult
	}

	s.respondJSON(w, map[string]interface{}{
		"buckets": result,
	})
}

func (s *Server) checkConsistency(data interface{}) bool {
	switch v := data.(type) {
	case map[string]string:
		if len(v) <= 1 {
			return true
		}
		var firstValue string
		first := true
		for _, value := range v {
			if first {
				firstValue = value
				first = false
			} else if value != firstValue {
				return false
			}
		}
		return true
	case map[string]interface{}:
		if len(v) <= 1 {
			return true
		}
		var firstValue string
		first := true
		for _, value := range v {
			valueStr := fmt.Sprintf("%v", value)
			if first {
				firstValue = valueStr
				first = false
			} else if valueStr != firstValue {
				return false
			}
		}
		return true
	}
	return true
}

// Helper methods
func (s *Server) respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func (s *Server) getMCInternalAliases() ([]map[string]string, error) {
	// Try to get aliases using mc command
	cmd := exec.Command("mc", "alias", "list", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var aliases []map[string]string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		var aliasData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &aliasData); err == nil {
			if aliasName, ok := aliasData["alias"].(string); ok {
				if url, ok := aliasData["URL"].(string); ok {
					aliases = append(aliases, map[string]string{
						"name": aliasName,
						"url":  url,
					})
				}
			}
		}
	}

	return aliases, nil
}

func (s *Server) listBuckets(alias string) ([]string, error) {
	cmd := exec.Command("mc", "ls", alias, "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var buckets []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(line), &data); err == nil {
			if key, ok := data["key"].(string); ok {
				buckets = append(buckets, strings.TrimSuffix(key, "/"))
			}
		}
	}

	return buckets, nil
}

func (s *Server) getAliasStats(alias string) map[string]interface{} {
	stats := map[string]interface{}{
		"bucket_count":  0,
		"total_size":    int64(0),
		"total_objects": int64(0),
		"buckets":       []map[string]interface{}{},
	}

	// Get list of buckets
	buckets, err := s.listBuckets(alias)
	if err != nil {
		return stats
	}

	stats["bucket_count"] = len(buckets)

	// Get stats for each bucket
	var bucketStats []map[string]interface{}
	var totalSize int64
	var totalObjects int64

	for _, bucket := range buckets {
		bucketStat := s.getBucketStats(alias, bucket)
		if bucketStat != nil {
			bucketStats = append(bucketStats, bucketStat)
			if size, ok := bucketStat["size"].(int64); ok {
				totalSize += size
			}
			if objects, ok := bucketStat["objects"].(int64); ok {
				totalObjects += objects
			}
		}
	}

	stats["buckets"] = bucketStats
	stats["total_size"] = totalSize
	stats["total_objects"] = totalObjects

	return stats
}

func (s *Server) getBucketStats(alias, bucket string) map[string]interface{} {
	// Use mc du command to get bucket statistics
	cmd := exec.Command("mc", "du", fmt.Sprintf("%s/%s", alias, bucket), "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If du fails, try to get basic info
		return map[string]interface{}{
			"name":    bucket,
			"size":    int64(0),
			"objects": int64(0),
		}
	}

	// Parse the last line of output which contains the total
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 {
		return map[string]interface{}{
			"name":    bucket,
			"size":    int64(0),
			"objects": int64(0),
		}
	}

	// Parse the JSON output
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &data); err == nil {
		size := int64(0)
		objects := int64(0)

		if sizeFloat, ok := data["size"].(float64); ok {
			size = int64(sizeFloat)
		}
		if objCount, ok := data["objects"].(float64); ok {
			objects = int64(objCount)
		}

		return map[string]interface{}{
			"name":    bucket,
			"size":    size,
			"objects": objects,
		}
	}

	return map[string]interface{}{
		"name":    bucket,
		"size":    int64(0),
		"objects": int64(0),
	}
}

func (s *Server) handleGetBucketStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	alias := r.URL.Query().Get("alias")
	bucket := r.URL.Query().Get("bucket")

	if alias == "" {
		s.respondError(w, http.StatusBadRequest, "Alias parameter is required")
		return
	}

	if bucket == "" {
		s.respondError(w, http.StatusBadRequest, "Bucket parameter is required")
		return
	}

	stats := s.getBucketStats(alias, bucket)
	s.respondJSON(w, stats)
}

// Job execution methods
func (s *Server) runCompareJob(job *Job, source, destination string, recursive bool) {
	job.updateStatus("running", "Starting comparison...")

	// Use mc-tool command
	args := []string{"compare", source, destination}
	if recursive {
		// recursive is default in mc-tool compare
	}

	cmd := exec.Command(s.executablePath, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		job.fail(fmt.Sprintf("Comparison failed: %v\n%s", err, string(output)))
		return
	}

	job.addOutput(string(output))

	// Parse output for summary
	resultData := map[string]interface{}{
		"output":  string(output),
		"success": true,
	}

	job.complete(resultData, "Comparison completed successfully")
}

func (s *Server) runAnalyzeJob(job *Job, alias, bucket, prefix string) {
	job.updateStatus("running", "Analyzing bucket...")

	// Use mc-tool command
	path := fmt.Sprintf("%s/%s", alias, bucket)
	if prefix != "" {
		path = fmt.Sprintf("%s/%s", path, prefix)
	}

	cmd := exec.Command(s.executablePath, "analyze", path)
	output, err := cmd.CombinedOutput()

	if err != nil {
		job.fail(fmt.Sprintf("Analysis failed: %v\n%s", err, string(output)))
		return
	}

	job.addOutput(string(output))

	resultData := map[string]interface{}{
		"output":  string(output),
		"success": true,
	}

	job.complete(resultData, "Analysis completed successfully")
}

func (s *Server) runProfileJob(job *Job, req struct {
	Alias           string `json:"alias"`
	ProfileType     string `json:"profile_type"`
	Duration        string `json:"duration"`
	DetectLeaks     bool   `json:"detect_leaks"`
	MonitorInterval string `json:"monitor_interval"`
	ThresholdMB     int    `json:"threshold_mb"`
}) {
	job.updateStatus("running", "Starting profiling...")

	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		job.fail(fmt.Sprintf("Invalid duration: %v", err))
		return
	}

	monitorInterval, err := time.ParseDuration(req.MonitorInterval)
	if err != nil {
		monitorInterval = 10 * time.Second
	}

	opts := profile.ProfileOptions{
		Alias:           req.Alias,
		ProfileType:     req.ProfileType,
		Duration:        duration,
		Verbose:         true,
		MCPath:          "mc",
		DetectLeaks:     req.DetectLeaks,
		MonitorInterval: monitorInterval,
		ThresholdMB:     req.ThresholdMB,
	}

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = profile.RunProfile(opts)

	w.Close()
	os.Stdout = oldStdout

	output, _ := io.ReadAll(r)
	job.addOutput(string(output))

	if err != nil {
		job.fail(fmt.Sprintf("Profiling failed: %v", err))
		return
	}

	resultData := map[string]interface{}{
		"profile_type": req.ProfileType,
		"duration":     req.Duration,
		"output":       string(output),
	}

	job.complete(resultData, "Profiling completed successfully")
}

func (s *Server) runChecklistJob(job *Job, alias, bucket string) {
	job.updateStatus("running", "Running checklist...")

	var output strings.Builder
	bucketPath := fmt.Sprintf("%s/%s", alias, bucket)

	// 1. Check Bucket Event Notification Configuration
	output.WriteString("=== BUCKET EVENT NOTIFICATION ===\n")
	eventCmd := exec.Command("mc", "event", "list", bucketPath, "--json")
	eventOutput, err := eventCmd.CombinedOutput()

	if err != nil {
		output.WriteString(fmt.Sprintf("❌ Failed to check event configuration: %v\n", err))
	} else {
		eventStr := string(eventOutput)
		if strings.TrimSpace(eventStr) == "" || strings.Contains(eventStr, "no event notification found") {
			output.WriteString("⚠️  No event notifications configured\n")
		} else {
			output.WriteString("✓ Event notifications configured:\n")
			// Parse JSON output to show event details
			decoder := json.NewDecoder(strings.NewReader(eventStr))
			for decoder.More() {
				var event map[string]interface{}
				if err := decoder.Decode(&event); err == nil {
					if arn, ok := event["arn"].(string); ok {
						output.WriteString(fmt.Sprintf("  - ARN: %s\n", arn))
					}
					if events, ok := event["events"].([]interface{}); ok {
						output.WriteString(fmt.Sprintf("    Events: %v\n", events))
					}
					if prefix, ok := event["prefix"].(string); ok && prefix != "" {
						output.WriteString(fmt.Sprintf("    Prefix: %s\n", prefix))
					}
					if suffix, ok := event["suffix"].(string); ok && suffix != "" {
						output.WriteString(fmt.Sprintf("    Suffix: %s\n", suffix))
					}
				}
			}
		}
	}
	output.WriteString("\n")

	// 2. Check Bucket Lifecycle Policy
	output.WriteString("=== BUCKET LIFECYCLE POLICY ===\n")
	lifecycleCmd := exec.Command("mc", "ilm", "ls", bucketPath, "--json")
	lifecycleOutput, err := lifecycleCmd.CombinedOutput()

	if err != nil {
		output.WriteString(fmt.Sprintf("❌ Failed to check lifecycle policy: %v\n", err))
	} else {
		lifecycleStr := string(lifecycleOutput)
		if strings.TrimSpace(lifecycleStr) == "" || strings.Contains(lifecycleStr, "no lifecycle configuration found") {
			output.WriteString("⚠️  No lifecycle policies configured\n")
		} else {
			output.WriteString("✓ Lifecycle policies configured:\n")
			// Parse JSON output to show lifecycle details
			decoder := json.NewDecoder(strings.NewReader(lifecycleStr))
			for decoder.More() {
				var rule map[string]interface{}
				if err := decoder.Decode(&rule); err == nil {
					if id, ok := rule["id"].(string); ok {
						output.WriteString(fmt.Sprintf("  - Rule ID: %s\n", id))
					}
					if status, ok := rule["status"].(string); ok {
						output.WriteString(fmt.Sprintf("    Status: %s\n", status))
					}
					if prefix, ok := rule["prefix"].(string); ok && prefix != "" {
						output.WriteString(fmt.Sprintf("    Prefix: %s\n", prefix))
					}
					if expiration, ok := rule["expiration"].(map[string]interface{}); ok {
						if days, ok := expiration["days"].(float64); ok {
							output.WriteString(fmt.Sprintf("    Expiration: %d days\n", int(days)))
						}
						if date, ok := expiration["date"].(string); ok {
							output.WriteString(fmt.Sprintf("    Expiration Date: %s\n", date))
						}
						if delMarker, ok := expiration["delete_marker"].(bool); ok && delMarker {
							output.WriteString("    Delete Expired Object Delete Markers: Yes\n")
						}
					}
					if noncurrentExpiration, ok := rule["noncurrent_version_expiration"].(map[string]interface{}); ok {
						if days, ok := noncurrentExpiration["noncurrent_days"].(float64); ok {
							output.WriteString(fmt.Sprintf("    Delete Noncurrent Versions After: %d days\n", int(days)))
						}
					}
					if transition, ok := rule["transition"].(map[string]interface{}); ok {
						if days, ok := transition["days"].(float64); ok {
							output.WriteString(fmt.Sprintf("    Transition: %d days\n", int(days)))
						}
						if storageClass, ok := transition["storage_class"].(string); ok {
							output.WriteString(fmt.Sprintf("    Storage Class: %s\n", storageClass))
						}
					}
					if noncurrentTransition, ok := rule["noncurrent_version_transition"].(map[string]interface{}); ok {
						if days, ok := noncurrentTransition["noncurrent_days"].(float64); ok {
							output.WriteString(fmt.Sprintf("    Transition Noncurrent Versions After: %d days\n", int(days)))
						}
						if storageClass, ok := noncurrentTransition["storage_class"].(string); ok {
							output.WriteString(fmt.Sprintf("    Noncurrent Storage Class: %s\n", storageClass))
						}
					}
				}
			}
		}
	}
	output.WriteString("\n")

	// 3. Check Bucket Versioning
	output.WriteString("=== BUCKET VERSIONING ===\n")
	versionCmd := exec.Command("mc", "version", "info", bucketPath, "--json")
	versionOutput, err := versionCmd.CombinedOutput()

	if err != nil {
		output.WriteString(fmt.Sprintf("❌ Failed to check versioning: %v\n", err))
	} else {
		versionStr := string(versionOutput)
		var versionInfo map[string]interface{}
		if err := json.Unmarshal([]byte(versionStr), &versionInfo); err == nil {
			if status, ok := versionInfo["status"].(string); ok {
				if status == "Enabled" {
					output.WriteString("✓ Versioning: Enabled\n")
				} else if status == "Suspended" {
					output.WriteString("⚠️  Versioning: Suspended\n")
				} else {
					output.WriteString("⚠️  Versioning: Disabled\n")
				}
			}
		} else {
			output.WriteString("⚠️  Versioning: Not configured\n")
		}
	}
	output.WriteString("\n")

	// 4. Summary
	output.WriteString("=== SUMMARY ===\n")
	outputStr := output.String()
	checkCount := strings.Count(outputStr, "✓")
	warningCount := strings.Count(outputStr, "⚠️")
	errorCount := strings.Count(outputStr, "❌")

	output.WriteString(fmt.Sprintf("Checks passed: %d\n", checkCount))
	output.WriteString(fmt.Sprintf("Warnings: %d\n", warningCount))
	output.WriteString(fmt.Sprintf("Errors: %d\n", errorCount))

	finalOutput := output.String()
	job.addOutput(finalOutput)

	resultData := map[string]interface{}{
		"bucket":        bucket,
		"alias":         alias,
		"output":        finalOutput,
		"checks_passed": checkCount,
		"warnings":      warningCount,
		"errors":        errorCount,
	}

	job.complete(resultData, "Checklist completed")
}

// JobManager methods
func (jm *JobManager) createJob(jobType string) *Job {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	job := &Job{
		ID:        fmt.Sprintf("%s-%d", jobType, time.Now().Unix()),
		Type:      jobType,
		Status:    "pending",
		Progress:  0,
		StartTime: time.Now(),
		Output:    []string{},
	}

	jm.jobs[job.ID] = job
	return job
}

func (jm *JobManager) getJob(id string) *Job {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	return jm.jobs[id]
}

// Job methods
func (j *Job) updateStatus(status, message string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = status
	j.Message = message
}

func (j *Job) updateProgress(progress int) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Progress = progress
}

func (j *Job) addOutput(output string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Output = append(j.Output, output)
}

func (j *Job) complete(result map[string]interface{}, message string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = "completed"
	j.Progress = 100
	j.Message = message
	j.Result = result
	now := time.Now()
	j.EndTime = &now
}

func (j *Job) fail(error string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = "failed"
	j.Error = error
	now := time.Now()
	j.EndTime = &now
}
