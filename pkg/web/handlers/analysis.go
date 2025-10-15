package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/liamdn8/mc-tool/pkg/profile"
	"github.com/liamdn8/mc-tool/pkg/web/models"
)

// AnalysisHandler handles analysis-related requests like compare, analyze, profile, checklist
type AnalysisHandler struct {
	BaseHandler
	executablePath string
	jobManager     *models.JobManager
}

// NewAnalysisHandler creates a new analysis handler
func NewAnalysisHandler(executablePath string, jobManager *models.JobManager) *AnalysisHandler {
	return &AnalysisHandler{
		executablePath: executablePath,
		jobManager:     jobManager,
	}
}

// HandleCompare handles POST /api/compare
func (h *AnalysisHandler) HandleCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
		Recursive   bool   `json:"recursive"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Create job
	job := h.jobManager.CreateJob("compare")
	go h.runCompareJob(job, req.Source, req.Destination, req.Recursive)

	h.RespondJSON(w, map[string]interface{}{
		"job_id": job.ID,
		"status": "started",
	})
}

// HandleAnalyze handles POST /api/analyze
func (h *AnalysisHandler) HandleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Alias  string `json:"alias"`
		Bucket string `json:"bucket"`
		Prefix string `json:"prefix"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Create job
	job := h.jobManager.CreateJob("analyze")
	go h.runAnalyzeJob(job, req.Alias, req.Bucket, req.Prefix)

	h.RespondJSON(w, map[string]interface{}{
		"job_id": job.ID,
		"status": "started",
	})
}

// HandleProfile handles POST /api/profile
func (h *AnalysisHandler) HandleProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
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
		h.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Create job
	job := h.jobManager.CreateJob("profile")
	go h.runProfileJob(job, req)

	h.RespondJSON(w, map[string]interface{}{
		"job_id": job.ID,
		"status": "started",
	})
}

// HandleChecklist handles POST /api/checklist
func (h *AnalysisHandler) HandleChecklist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Alias  string `json:"alias"`
		Bucket string `json:"bucket"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Create job
	job := h.jobManager.CreateJob("checklist")
	go h.runChecklistJob(job, req.Alias, req.Bucket)

	h.RespondJSON(w, map[string]interface{}{
		"job_id": job.ID,
		"status": "started",
	})
}

// Job execution methods
func (h *AnalysisHandler) runCompareJob(job *models.Job, source, destination string, recursive bool) {
	job.UpdateStatus("running", "Starting comparison...")

	// Use mc-tool command
	args := []string{"compare", source, destination}
	if recursive {
		// recursive is default in mc-tool compare
	}

	cmd := exec.Command(h.executablePath, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		job.Fail(fmt.Sprintf("Comparison failed: %v\n%s", err, string(output)))
		return
	}

	job.AddOutput(string(output))

	// Parse output for summary
	resultData := map[string]interface{}{
		"output":  string(output),
		"success": true,
	}

	job.Complete(resultData, "Comparison completed successfully")
}

func (h *AnalysisHandler) runAnalyzeJob(job *models.Job, alias, bucket, prefix string) {
	job.UpdateStatus("running", "Analyzing bucket...")

	// Use mc-tool command
	path := fmt.Sprintf("%s/%s", alias, bucket)
	if prefix != "" {
		path = fmt.Sprintf("%s/%s", path, prefix)
	}

	cmd := exec.Command(h.executablePath, "analyze", path)
	output, err := cmd.CombinedOutput()

	if err != nil {
		job.Fail(fmt.Sprintf("Analysis failed: %v\n%s", err, string(output)))
		return
	}

	job.AddOutput(string(output))

	resultData := map[string]interface{}{
		"output":  string(output),
		"success": true,
	}

	job.Complete(resultData, "Analysis completed successfully")
}

func (h *AnalysisHandler) runProfileJob(job *models.Job, req struct {
	Alias           string `json:"alias"`
	ProfileType     string `json:"profile_type"`
	Duration        string `json:"duration"`
	DetectLeaks     bool   `json:"detect_leaks"`
	MonitorInterval string `json:"monitor_interval"`
	ThresholdMB     int    `json:"threshold_mb"`
}) {
	job.UpdateStatus("running", "Starting profiling...")

	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		job.Fail(fmt.Sprintf("Invalid duration: %v", err))
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
	job.AddOutput(string(output))

	if err != nil {
		job.Fail(fmt.Sprintf("Profiling failed: %v", err))
		return
	}

	resultData := map[string]interface{}{
		"profile_type": req.ProfileType,
		"duration":     req.Duration,
		"output":       string(output),
	}

	job.Complete(resultData, "Profiling completed successfully")
}

func (h *AnalysisHandler) runChecklistJob(job *models.Job, alias, bucket string) {
	job.UpdateStatus("running", "Running checklist...")

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
	job.AddOutput(finalOutput)

	resultData := map[string]interface{}{
		"bucket":        bucket,
		"alias":         alias,
		"output":        finalOutput,
		"checks_passed": checkCount,
		"warnings":      warningCount,
		"errors":        errorCount,
	}

	job.Complete(resultData, "Checklist completed")
}
