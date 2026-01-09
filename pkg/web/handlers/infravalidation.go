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
	kubeconfigPath string // Custom kubeconfig path from CLI or env
}

// NewInfraValidationHandler creates a new infrastructure validation handler
func NewInfraValidationHandler(executablePath string, jobManager *models.JobManager, kubeconfigPath string) *InfraValidationHandler {
	return &InfraValidationHandler{
		executablePath: executablePath,
		jobManager:     jobManager,
		kubeconfigPath: kubeconfigPath,
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
// Returns list of VIMs (contexts) from kubeconfig
func (h *InfraValidationHandler) HandleGetInfraVIMs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Load contexts from kubeconfig (use custom path if provided, otherwise default ~/.kube/config)
	contexts, err := infravalidation.LoadKubeconfigContexts(h.kubeconfigPath)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError,
			fmt.Sprintf("Failed to load kubeconfig contexts: %v", err))
		return
	}

	// Get current context
	currentContext, _ := infravalidation.GetCurrentKubeconfigContext(h.kubeconfigPath)

	// Convert contexts to VIM format
	vims := []map[string]interface{}{}
	for _, ctx := range contexts {
		vim := map[string]interface{}{
			"name":    ctx,
			"context": ctx,
			"current": ctx == currentContext,
		}
		vims = append(vims, vim)
	}

	configPath := h.kubeconfigPath
	if configPath == "" {
		configPath = "~/.kube/config (default)"
	}

	h.RespondJSON(w, map[string]interface{}{
		"vims":           vims,
		"count":          len(vims),
		"currentContext": currentContext,
		"configPath":     configPath,
	})
}

// HandleGetNamespaces handles GET /api/validate/infrastructure/namespaces?vim=vim1
// vim parameter is the kubeconfig context name
func (h *InfraValidationHandler) HandleGetNamespaces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	contextName := r.URL.Query().Get("vim")
	if contextName == "" {
		h.RespondError(w, http.StatusBadRequest, "vim parameter (context name) is required")
		return
	}

	// Create site config from context
	siteConfig := infravalidation.SiteConfig{
		Context:        contextName,
		KubeconfigPath: h.kubeconfigPath,
	}

	// Create K8s client
	client, err := infravalidation.NewK8sClient(siteConfig)
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

// HandleSearchNamespaces handles GET /api/validate/infrastructure/search-namespaces?keyword=app&exact=false
func (h *InfraValidationHandler) HandleSearchNamespaces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	keyword := r.URL.Query().Get("keyword")
	exactMatch := r.URL.Query().Get("exact") == "true"

	if keyword == "" {
		h.RespondError(w, http.StatusBadRequest, "keyword parameter is required")
		return
	}

	// Load kubeconfig contexts
	contexts, err := infravalidation.LoadKubeconfigContexts(h.kubeconfigPath)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to load kubeconfig contexts: %v", err))
		return
	}

	// Search across all contexts
	type NamespaceMatch struct {
		Vim       string `json:"vim"`
		Namespace string `json:"namespace"`
		Context   string `json:"context"`
	}

	var matches []NamespaceMatch

	for _, contextName := range contexts {
		// Create site config from context
		siteConfig := infravalidation.SiteConfig{
			Context:        contextName,
			KubeconfigPath: h.kubeconfigPath,
		}

		// Create K8s client
		client, err := infravalidation.NewK8sClient(siteConfig)
		if err != nil {
			// Skip contexts that fail to connect
			continue
		}

		// List namespaces
		namespaces, err := client.ListNamespaces()
		if err != nil {
			// Skip contexts that fail to list namespaces
			continue
		}

		// Filter namespaces
		for _, ns := range namespaces {
			matched := false
			if exactMatch {
				matched = ns == keyword
			} else {
				matched = strings.Contains(strings.ToLower(ns), strings.ToLower(keyword))
			}

			if matched {
				matches = append(matches, NamespaceMatch{
					Vim:       contextName,
					Namespace: ns,
					Context:   contextName,
				})
			}
		}
	}

	h.RespondJSON(w, map[string]interface{}{
		"matches": matches,
		"count":   len(matches),
	})
}

// HandleDiscoverResources handles GET /api/validate/infrastructure/discover-resources?vim=xxx&namespace=yyy
func (h *InfraValidationHandler) HandleDiscoverResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	contextName := r.URL.Query().Get("vim")
	namespace := r.URL.Query().Get("namespace")

	if contextName == "" || namespace == "" {
		h.RespondError(w, http.StatusBadRequest, "vim (context name) and namespace parameters are required")
		return
	}

	// Create site config from context
	siteConfig := infravalidation.SiteConfig{
		Context:        contextName,
		KubeconfigPath: h.kubeconfigPath,
	}

	// Create K8s client
	client, err := infravalidation.NewK8sClient(siteConfig)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create K8s client: %v", err))
		return
	}

	// Discover all resources in the namespace
	resourceMap, err := client.GetAllResources(r.Context(), namespace)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to discover resources: %v", err))
		return
	}

	// Format response
	type ResourceSummary struct {
		Kind  string `json:"kind"`
		Count int    `json:"count"`
	}

	var summaries []ResourceSummary
	for kind, resources := range resourceMap {
		summaries = append(summaries, ResourceSummary{
			Kind:  kind,
			Count: len(resources),
		})
	}

	h.RespondJSON(w, map[string]interface{}{
		"resources": summaries,
		"total":     len(summaries),
	})
}

// HandleGetDiff handles GET /api/validate/infrastructure/diff?baseline=site1/ns1&target=site2/ns2&resource_type=Deployment&resource_name=app
func (h *InfraValidationHandler) HandleGetDiff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	baseline := r.URL.Query().Get("baseline")
	target := r.URL.Query().Get("target")
	resourceType := r.URL.Query().Get("resource_type")
	resourceName := r.URL.Query().Get("resource_name")

	if baseline == "" || target == "" || resourceType == "" || resourceName == "" {
		h.RespondError(w, http.StatusBadRequest, "baseline, target, resource_type, and resource_name are required")
		return
	}

	// Parse baseline and target
	baselineSite, baselineNS, err := infravalidation.ParseSiteNamespace(baseline)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid baseline format: %v", err))
		return
	}

	targetSite, targetNS, err := infravalidation.ParseSiteNamespace(target)
	if err != nil {
		h.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid target format: %v", err))
		return
	}

	// Create site configs from context names
	baselineSiteConfig := infravalidation.SiteConfig{
		Context:        baselineSite,
		KubeconfigPath: h.kubeconfigPath,
	}

	targetSiteConfig := infravalidation.SiteConfig{
		Context:        targetSite,
		KubeconfigPath: h.kubeconfigPath,
	}

	// Create K8s clients
	baselineClient, err := infravalidation.NewK8sClient(baselineSiteConfig)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create baseline client: %v", err))
		return
	}

	targetClient, err := infravalidation.NewK8sClient(targetSiteConfig)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create target client: %v", err))
		return
	}

	// Get resources
	baselineResource, err := baselineClient.GetResource(r.Context(), baselineNS, resourceName, infravalidation.ResourceType(resourceType))
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get baseline resource: %v", err))
		return
	}

	targetResource, err := targetClient.GetResource(r.Context(), targetNS, resourceName, infravalidation.ResourceType(resourceType))
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get target resource: %v", err))
		return
	}

	// Normalize resources
	normalizer := infravalidation.NewNormalizer(nil, infravalidation.SecretCompareKeys)
	normalizedBaseline, err := normalizer.Normalize(baselineResource)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to normalize baseline: %v", err))
		return
	}

	normalizedTarget, err := normalizer.Normalize(targetResource)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to normalize target: %v", err))
		return
	}

	// Convert to YAML
	baselineYAML, err := infravalidation.ToJSON(normalizedBaseline)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to marshal baseline: %v", err))
		return
	}

	targetYAML, err := infravalidation.ToJSON(normalizedTarget)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to marshal target: %v", err))
		return
	}

	h.RespondJSON(w, map[string]interface{}{
		"baseline": baselineYAML,
		"target":   targetYAML,
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

// findOrCreateResourceRow finds existing resource in table or creates new one
func findOrCreateResourceRow(resourceTable []map[string]interface{}, resourceType, resourceName string, baseline string) (map[string]interface{}, []map[string]interface{}) {
	// Look for existing row with same resource type and name
	for _, row := range resourceTable {
		if row["resource_type"] == resourceType && row["resource_name"] == resourceName {
			return row, resourceTable
		}
	}

	// Not found - create new row
	newRow := map[string]interface{}{
		"resource_type": resourceType,
		"resource_name": resourceName,
		"baseline":      baseline,
	}

	// Add baseline column with default status
	baselineKey := fmt.Sprintf("%v", baseline)
	newRow[baselineKey] = map[string]interface{}{
		"status": "-",
	}

	resourceTable = append(resourceTable, newRow)
	return newRow, resourceTable
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

		// Parse baseline - support both old format and new [i] format
		if strings.HasPrefix(line, "[i] Baseline:") {
			result["baseline"] = strings.TrimSpace(strings.TrimPrefix(line, "[i] Baseline:"))
		} else if strings.HasPrefix(line, "Baseline:") {
			result["baseline"] = strings.TrimSpace(strings.TrimPrefix(line, "Baseline:"))
		}

		// Parse target - support both old format and new [>] format
		if strings.HasPrefix(line, "[>] Target:") {
			currentTarget = strings.TrimSpace(strings.TrimPrefix(line, "[>] Target:"))
			if result["targets"] == nil {
				result["targets"] = []string{}
			}
			result["targets"] = append(result["targets"].([]string), currentTarget)
		} else if strings.HasPrefix(line, "Target:") && !strings.Contains(line, "Targets:") {
			currentTarget = strings.TrimSpace(strings.TrimPrefix(line, "Target:"))
			if result["targets"] == nil {
				result["targets"] = []string{}
			}
			result["targets"] = append(result["targets"].([]string), currentTarget)
		}

		// Parse statistics
		if strings.Contains(line, "Total Comparisons:") {
			var count int
			fmt.Sscanf(line, "Total Comparisons: %d", &count)
			result["totalComparisons"] = count
		} else if strings.Contains(line, "[OK] Matches:") {
			// New format: "[OK] Matches:      8 (57.1%)"
			var count int
			fmt.Sscanf(line, "[OK] Matches: %d", &count)
			result["matchCount"] = count
		} else if strings.Contains(line, "Matches:") && !strings.Contains(line, "[OK]") {
			var count int
			fmt.Sscanf(line, "Matches: %d", &count)
			result["matchCount"] = count
		} else if strings.Contains(line, "[X]  Mismatches:") {
			// New format: "[X]  Mismatches:   5 (35.7%)"
			var count int
			fmt.Sscanf(line, "[X] Mismatches: %d", &count)
			result["mismatchCount"] = count
		} else if strings.Contains(line, "Mismatches:") && !strings.Contains(line, "[X]") {
			var count int
			fmt.Sscanf(line, "Mismatches: %d", &count)
			result["mismatchCount"] = count
		} else if strings.Contains(line, "[!]  Not Found:") {
			// New format: "[!]  Not Found:    1 (7.1%)"
			var count int
			fmt.Sscanf(line, "[!] Not Found: %d", &count)
			result["notFoundCount"] = count
		} else if strings.Contains(line, "Not Found:") && !strings.Contains(line, "[!]") {
			var count int
			fmt.Sscanf(line, "Not Found: %d", &count)
			result["notFoundCount"] = count
		}

		// Parse status
		if strings.Contains(line, "[OK] All configurations match!") || strings.Contains(line, "All configurations match!") {
			result["status"] = "success"
		} else if strings.Contains(line, "[!] Configuration drift detected!") || strings.Contains(line, "Configuration drift detected!") {
			result["status"] = "drift"
		}

		// Parse resource types - detect both indentation styles
		// Resource type lines have exactly 2 spaces of indent and end with ":"
		// Examples: "  Deployment:", "  StatefulSet:", "  ConfigMap:"
		if strings.HasPrefix(originalLine, "  ") && !strings.HasPrefix(originalLine, "   ") && strings.HasSuffix(line, ":") {
			// Resource type line pattern: "  ResourceType:"
			resourceType := strings.TrimSuffix(line, ":")
			resourceType = strings.TrimSpace(resourceType)

			// Filter out non-resource lines and status lines
			if len(resourceType) > 0 &&
				resourceType != "Mode" &&
				resourceType != "Overall Results" &&
				!strings.Contains(resourceType, "match") &&
				!strings.Contains(resourceType, "mismatch") &&
				!strings.Contains(resourceType, "not found") &&
				!strings.Contains(resourceType, "Baseline") &&
				!strings.Contains(resourceType, "Target") {
				currentResourceType = resourceType
			}
		}

		// Parse match resources
		// New format: "    [OK] 1 match(es):" followed by "        - resource-name"
		if (strings.HasPrefix(originalLine, "    [OK]") || strings.HasPrefix(originalLine, "    ✅")) && strings.Contains(line, "match") {
			// Extract matched resource names from following lines
			for j := i + 1; j < len(lines) && j < i+50; j++ {
				nextOrigLine := lines[j]
				nextLine := strings.TrimSpace(nextOrigLine)

				if strings.HasPrefix(nextOrigLine, "      - ") || strings.HasPrefix(nextOrigLine, "        - ") {
					// Resource name (6 or 8 space indent + dash)
					resourceName := strings.TrimPrefix(nextLine, "- ")

					// Find or create resource row
					resourceTable := result["resource_table"].([]map[string]interface{})
					resourceRow, updatedTable := findOrCreateResourceRow(resourceTable, currentResourceType, resourceName, fmt.Sprintf("%v", result["baseline"]))
					result["resource_table"] = updatedTable

					// Update baseline status to match
					baselineKey := fmt.Sprintf("%v", result["baseline"])
					resourceRow[baselineKey] = map[string]interface{}{
						"status": "match",
					}

					// Add/update status for target (match)
					if currentTarget != "" {
						resourceRow[currentTarget] = map[string]interface{}{
							"status": "match",
						}
					}
				} else if strings.TrimSpace(nextOrigLine) != "" && !strings.HasPrefix(nextOrigLine, "      ") && !strings.HasPrefix(nextOrigLine, "        ") {
					// Not a resource name line, stop parsing this block
					break
				}
			}
		}

		// Parse mismatch resources
		// Format: "    [X] 1 mismatch(es):" followed by "        - resource-name"
		if (strings.HasPrefix(originalLine, "    [X]") || strings.HasPrefix(originalLine, "    ✗")) && strings.Contains(line, "mismatch") {
			// Mismatch line - extract resource names from following lines
			for j := i + 1; j < len(lines) && j < i+50; j++ {
				nextOrigLine := lines[j]
				nextLine := strings.TrimSpace(nextOrigLine)

				if strings.HasPrefix(nextOrigLine, "      - ") || strings.HasPrefix(nextOrigLine, "        - ") {
					// Resource name (6 or 8 space indent + dash)
					resourceName := strings.TrimPrefix(nextLine, "- ")

					// Find or create resource row
					resourceTable := result["resource_table"].([]map[string]interface{})
					resourceRow, updatedTable := findOrCreateResourceRow(resourceTable, currentResourceType, resourceName, fmt.Sprintf("%v", result["baseline"]))
					result["resource_table"] = updatedTable

					// Baseline: keep configured marker
					baselineKey := fmt.Sprintf("%v", result["baseline"])
					resourceRow[baselineKey] = map[string]interface{}{
						"status": "-",
					}

					// Target mismatch
					if currentTarget != "" {
						resourceRow[currentTarget] = map[string]interface{}{
							"status": "mismatch",
						}
					}
				} else if strings.TrimSpace(nextOrigLine) != "" && !strings.HasPrefix(nextOrigLine, "      ") && !strings.HasPrefix(nextOrigLine, "        ") {
					// Not a resource name line, stop parsing this block
					break
				}
			}
		}

		// Parse "not found" resources
		// Format: "    [!] 1 not found:" followed by "        - resource-name"
		if (strings.HasPrefix(originalLine, "    [!]") || strings.HasPrefix(originalLine, "    ⚠")) && strings.Contains(line, "not found") {
			for j := i + 1; j < len(lines) && j < i+50; j++ {
				nextOrigLine := lines[j]
				nextLine := strings.TrimSpace(nextOrigLine)

				if strings.HasPrefix(nextOrigLine, "      - ") || strings.HasPrefix(nextOrigLine, "        - ") {
					resourceName := strings.TrimPrefix(nextLine, "- ")

					// Find or create resource row
					resourceTable := result["resource_table"].([]map[string]interface{})
					resourceRow, updatedTable := findOrCreateResourceRow(resourceTable, currentResourceType, resourceName, fmt.Sprintf("%v", result["baseline"]))
					result["resource_table"] = updatedTable

					baselineKey := fmt.Sprintf("%v", result["baseline"])
					resourceRow[baselineKey] = map[string]interface{}{
						"status": "-",
					}

					if currentTarget != "" {
						resourceRow[currentTarget] = map[string]interface{}{
							"status": "not_found",
						}
					}
				} else if strings.TrimSpace(nextOrigLine) != "" && !strings.HasPrefix(nextOrigLine, "      ") && !strings.HasPrefix(nextOrigLine, "        ") {
					break
				}
			}
		}

		// Parse "extra" resources (in target but not in baseline)
		// Format: "    [+] 1 extra (not in baseline):" followed by "        - resource-name"
		if (strings.HasPrefix(originalLine, "    [+]") || strings.HasPrefix(originalLine, "    ➕")) && strings.Contains(line, "extra") {
			for j := i + 1; j < len(lines) && j < i+50; j++ {
				nextOrigLine := lines[j]
				nextLine := strings.TrimSpace(nextOrigLine)

				if strings.HasPrefix(nextOrigLine, "      - ") || strings.HasPrefix(nextOrigLine, "        - ") {
					resourceName := strings.TrimPrefix(nextLine, "- ")

					// Find or create resource row
					resourceTable := result["resource_table"].([]map[string]interface{})
					resourceRow, updatedTable := findOrCreateResourceRow(resourceTable, currentResourceType, resourceName, fmt.Sprintf("%v", result["baseline"]))
					result["resource_table"] = updatedTable

					baselineKey := fmt.Sprintf("%v", result["baseline"])
					resourceRow[baselineKey] = map[string]interface{}{
						"status": "not_found",
					}

					if currentTarget != "" {
						resourceRow[currentTarget] = map[string]interface{}{
							"status": "extra",
						}
					}
				} else if strings.TrimSpace(nextOrigLine) != "" && !strings.HasPrefix(nextOrigLine, "      ") && !strings.HasPrefix(nextOrigLine, "        ") {
					break
				}
			}
		}
	}

	return result
}
