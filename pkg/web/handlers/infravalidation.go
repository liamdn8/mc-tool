package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/liamdn8/mc-tool/pkg/infravalidation"
	"github.com/liamdn8/mc-tool/pkg/web/models"
)

// InfraValidationHandler handles infrastructure validation requests
type InfraValidationHandler struct {
	BaseHandler
	executablePath string
	jobManager     *models.JobManager
}

// NewInfraValidationHandler creates a new infrastructure validation handler
func NewInfraValidationHandler(executablePath string, jobManager *models.JobManager) *InfraValidationHandler {
	return &InfraValidationHandler{
		executablePath: executablePath,
		jobManager:     jobManager,
	}
}

// HandleInfraValidate handles POST /api/validate/infrastructure
func (h *InfraValidationHandler) HandleInfraValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Baseline string   `json:"baseline"` // e.g., "site1/app-prod"
		Targets  []string `json:"targets"`  // e.g., ["site2/app-staging", "site3/app-dev"]
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	if req.Baseline == "" {
		h.RespondError(w, http.StatusBadRequest, "Baseline is required")
		return
	}

	if len(req.Targets) == 0 {
		h.RespondError(w, http.StatusBadRequest, "At least one target is required")
		return
	}

	// Create job
	job := h.jobManager.CreateJob("infra-validate")
	go h.runInfraValidateJob(job, req.Baseline, req.Targets)

	h.RespondJSON(w, map[string]interface{}{
		"job_id": job.ID,
		"status": "started",
	})
}

// HandleGetInfraVIMs handles GET /api/validate/infrastructure/vims
func (h *InfraValidationHandler) HandleGetInfraVIMs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Load infra config from ~/.mc-tool/infra-config.yaml
	config, err := infravalidation.LoadDefaultInfraConfig()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to load infra config: %v", err))
		return
	}

	// Convert config to response format
	vims := []map[string]interface{}{}
	for name, siteConfig := range config.Sites {
		vims = append(vims, map[string]interface{}{
			"name":     name,
			"endpoint": siteConfig.Endpoint,
			"insecure": siteConfig.Insecure,
		})
	}

	h.RespondJSON(w, map[string]interface{}{
		"vims": vims,
	})
}

// HandleGetNamespaces handles GET /api/validate/infrastructure/namespaces?vim=vim1
func (h *InfraValidationHandler) HandleGetNamespaces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	vimName := r.URL.Query().Get("vim")
	if vimName == "" {
		h.RespondError(w, http.StatusBadRequest, "vim parameter is required")
		return
	}

	// Load VIM config
	vimConfig, err := infravalidation.LoadSiteConfig(vimName)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to load VIM config: %v", err))
		return
	}

	// Create K8s client
	client, err := infravalidation.NewK8sClient(vimConfig)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create K8s client: %v", err))
		return
	}

	// List namespaces
	namespaces, err := client.ListNamespaces()
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list namespaces: %v", err))
		return
	}

	h.RespondJSON(w, map[string]interface{}{
		"namespaces": namespaces,
	})
}

// HandleGetInfraHistory handles GET /api/validate/infrastructure/history?limit=10
func (h *InfraValidationHandler) HandleGetInfraHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	limit := 20 // Default limit
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		fmt.Sscanf(limitParam, "%d", &limit)
	}

	jobs := h.jobManager.GetJobHistory("infra-validate", limit)

	h.RespondJSON(w, map[string]interface{}{
		"jobs": jobs,
	})
}

func (h *InfraValidationHandler) runInfraValidateJob(job *models.Job, baseline string, targets []string) {
	defer func() {
		if r := recover(); r != nil {
			job.Fail(fmt.Sprintf("Panic: %v", r))
		}
	}()

	// Build command: mc-tool validate-infra baseline target1 target2 ...
	args := []string{"validate-infra", baseline}
	args = append(args, targets...)

	cmd := exec.Command(h.executablePath, args...)
	output, err := cmd.CombinedOutput()

	// Note: validate-infra exits with code 1 if there are mismatches, but still produces valid output
	// We only fail if there's an error AND no output
	if err != nil && len(output) == 0 {
		job.Fail(fmt.Sprintf("Command failed: %v", err))
		return
	}

	// Try to parse output as validation summary
	lines := strings.Split(string(output), "\n")
	summary := h.parseInfraValidationOutput(lines)

	job.Complete(map[string]interface{}{
		"output":  string(output),
		"summary": summary,
	}, "Validation completed")
}

func (h *InfraValidationHandler) parseInfraValidationOutput(lines []string) map[string]interface{} {
	result := map[string]interface{}{
		"baseline":         "",
		"targets":          []string{},
		"totalComparisons": 0,
		"matchCount":       0,
		"mismatchCount":    0,
		"notFoundCount":    0,
		"status":           "unknown",
		"resource_table":   []map[string]interface{}{},
	}

	currentTarget := ""
	currentResourceType := ""

	for i, originalLine := range lines {
		line := strings.TrimSpace(originalLine)

		// Skip empty lines and separator lines
		if line == "" || strings.Contains(line, "════") || strings.Contains(line, "────") {
			continue
		}

		if strings.HasPrefix(line, "Baseline:") {
			result["baseline"] = strings.TrimSpace(strings.TrimPrefix(line, "Baseline:"))
		} else if strings.HasPrefix(line, "Target:") {
			currentTarget = strings.TrimSpace(strings.TrimPrefix(line, "Target:"))
			if result["targets"] == nil {
				result["targets"] = []string{}
			}
			result["targets"] = append(result["targets"].([]string), currentTarget)
		} else if strings.Contains(line, "Total Comparisons:") {
			var count int
			fmt.Sscanf(line, "Total Comparisons: %d", &count)
			result["totalComparisons"] = count
		} else if strings.Contains(line, "Matches:") {
			var count int
			fmt.Sscanf(line, "Matches: %d", &count)
			result["matchCount"] = count
		} else if strings.Contains(line, "Mismatches:") {
			var count int
			fmt.Sscanf(line, "Mismatches: %d", &count)
			result["mismatchCount"] = count
		} else if strings.Contains(line, "Not Found:") {
			var count int
			fmt.Sscanf(line, "Not Found: %d", &count)
			result["notFoundCount"] = count
		} else if strings.Contains(line, "All configurations match!") {
			result["status"] = "success"
		} else if strings.Contains(line, "Configuration drift detected!") {
			result["status"] = "drift"
		} else if strings.HasPrefix(originalLine, "  ") && strings.HasSuffix(line, ":") && !strings.Contains(line, "Baseline") && !strings.Contains(line, "Target") && !strings.Contains(line, "match(es)") {
			// Resource type line (e.g., "  Deployment:", "  ConfigMap:")
			// Has 2-space indent and ends with colon, but not a match/mismatch count line
			resourceType := strings.TrimSuffix(line, ":")
			if len(resourceType) > 0 && resourceType != "Mode" && resourceType != "Overall Results" {
				currentResourceType = resourceType
			}
		} else if strings.HasPrefix(originalLine, "    ") && strings.Contains(line, "match(es)") && !strings.Contains(line, "mismatch") {
			// Skip match lines for now - we focus on mismatches
		} else if strings.HasPrefix(originalLine, "    ") && strings.Contains(line, "mismatch(es)") {
			// Mismatch line - extract resource names from following lines
			for j := i + 1; j < len(lines) && j < i+50; j++ {
				nextOrigLine := lines[j]
				nextLine := strings.TrimSpace(nextOrigLine)

				if strings.HasPrefix(nextOrigLine, "      - ") {
					// Resource name with 6-space indent
					resourceName := strings.TrimPrefix(nextLine, "- ")

					// Create resource table entry
					resourceRow := map[string]interface{}{
						"resource_type": currentResourceType,
						"resource_name": resourceName,
						"baseline":      result["baseline"],
					}

					// Add status for baseline (always exists)
					baselineKey := fmt.Sprintf("%v", result["baseline"])
					resourceRow[baselineKey] = map[string]interface{}{
						"status": "configured",
					}

					// Add status for target (mismatch)
					if currentTarget != "" {
						resourceRow[currentTarget] = map[string]interface{}{
							"status": "mismatch",
						}
					}

					if result["resource_table"] == nil {
						result["resource_table"] = []map[string]interface{}{}
					}
					result["resource_table"] = append(result["resource_table"].([]map[string]interface{}), resourceRow)
				} else if strings.TrimSpace(nextOrigLine) != "" && !strings.HasPrefix(nextOrigLine, "      ") {
					// Not a resource name line, stop parsing this block
					break
				}
			}
		}
	}

	return result
}
