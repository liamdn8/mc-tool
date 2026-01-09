package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/liamdn8/mc-tool/pkg/infravalidation"
	"github.com/xuri/excelize/v2"
)

// DiffDetail represents a specific difference between two resources
type DiffDetail struct {
	Path          string
	BaselineValue interface{}
	TargetValue   interface{}
}

// HandleExportInfraValidation handles GET /api/validate/infrastructure/export?job_id=xxx&format=xlsx
func (h *InfraValidationHandler) HandleExportInfraValidation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	jobID := r.URL.Query().Get("job_id")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "xlsx" // default to Excel
	}

	if jobID == "" {
		h.RespondError(w, http.StatusBadRequest, "job_id parameter is required")
		return
	}

	// Get job result
	job := h.jobManager.GetJob(jobID)
	if job == nil {
		h.RespondError(w, http.StatusNotFound, "Job not found")
		return
	}

	if job.Status != "completed" {
		h.RespondError(w, http.StatusBadRequest, "Job is not completed yet")
		return
	}

	// Extract validation summary from job result
	// Job result structure: job.Result["summary"] contains the validation data
	var summary map[string]interface{}

	if summaryData, ok := job.Result["summary"].(map[string]interface{}); ok {
		summary = summaryData
	} else {
		// Fallback: use job.Result directly if summary is not nested
		summary = job.Result
	}

	// Validate that resource_table exists
	if _, exists := summary["resource_table"]; !exists {
		h.RespondError(w, http.StatusInternalServerError, "resource_table not found in summary")
		return
	}

	// Generate file based on format
	var fileContent []byte
	var filename string
	var contentType string
	var err error

	if format == "xlsx" {
		fileContent, err = h.generateExcelReport(summary)
		if err != nil {
			h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate Excel: %v", err))
			return
		}
		filename = fmt.Sprintf("infra-validation-%s.xlsx", jobID)
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	} else {
		fileContent, err = h.generateCSVReport(summary)
		if err != nil {
			h.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate CSV: %v", err))
			return
		}
		filename = fmt.Sprintf("infra-validation-%s.csv", jobID)
		contentType = "text/csv"
	}

	// Set headers for file download
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileContent)))

	// Write file content
	w.Write(fileContent)
}

func (h *InfraValidationHandler) generateExcelReport(summary map[string]interface{}) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Validation Result"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1") // Remove default sheet

	// Set column headers
	headers := []string{"Site", "Namespace", "Object", "Type", "Key", "Value", "Status"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// Style header row
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#4472C4"}},
	})
	f.SetCellStyle(sheetName, "A1", "G1", headerStyle)

	// Get baseline from summary
	baselineSite, _ := summary["baseline"].(string)

	// Process resource_table
	resourceTable, ok := summary["resource_table"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("resource_table not found in summary")
	}

	row := 2
	for _, item := range resourceTable {
		resource, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		resourceType, _ := resource["resource_type"].(string)
		resourceName, _ := resource["resource_name"].(string)

		// Iterate through all site columns (excluding baseline, resource_type, resource_name)
		for key, value := range resource {
			if key == "baseline" || key == "resource_type" || key == "resource_name" {
				continue
			}

			// This is a site column
			site := key
			namespace := extractNamespace(site)

			statusMap, ok := value.(map[string]interface{})
			if !ok {
				continue
			}

			status, _ := statusMap["status"].(string)

			// Convert status to P/F
			var statusCode string
			switch status {
			case "match":
				statusCode = "P" // Pass
			case "mismatch", "not_found", "extra":
				statusCode = "F" // Fail
			default:
				statusCode = "-"
			}

			// For Pass status, don't show key/value
			// For Fail status, would need to call diff API - for now just mark as F
			if statusCode == "P" {
				// Pass - no key/value details needed
				f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), site)
				f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), namespace)
				f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), resourceName)
				f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), resourceType)
				f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), "") // No key
				f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), "") // No value
				f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), statusCode)
				row++
			} else if statusCode == "F" {
				// Fail - extract diff details if available
				if status == "mismatch" {
					// Try to get diff details
					diffs := h.extractDiffDetails(baselineSite, site, namespace, resourceType, resourceName)
					if len(diffs) > 0 {
						// Add a row for each diff
						for _, diff := range diffs {
							f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), site)
							f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), namespace)
							f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), resourceName)
							f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), resourceType)
							f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), diff.Path)
							f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("Expected: %v, Got: %v", diff.BaselineValue, diff.TargetValue))
							f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), statusCode)
							row++
						}
					} else {
						// No diff details available - show generic failure
						f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), site)
						f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), namespace)
						f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), resourceName)
						f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), resourceType)
						f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), "Mismatch")
						f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), "Check diff API for details")
						f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), statusCode)
						row++
					}
				} else {
					// not_found or extra - show status
					f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), site)
					f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), namespace)
					f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), resourceName)
					f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), resourceType)
					f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("Status: %s", status))
					f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), "-")
					f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), statusCode)
					row++
				}
			}
		}
	}

	// Auto-fit columns
	for i := 0; i < 7; i++ {
		col := string(rune('A' + i))
		f.SetColWidth(sheetName, col, col, 20)
	}

	// Save to buffer
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (h *InfraValidationHandler) generateCSVReport(summary map[string]interface{}) ([]byte, error) {
	var csv string

	// Headers
	csv += "Site,Namespace,Object,Type,Key,Value,Status\n"

	// Get baseline from summary
	baselineSite, _ := summary["baseline"].(string)

	// Process resource_table
	resourceTable, ok := summary["resource_table"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("resource_table not found in summary")
	}

	for _, item := range resourceTable {
		resource, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		resourceType, _ := resource["resource_type"].(string)
		resourceName, _ := resource["resource_name"].(string)

		// Iterate through all site columns
		for key, value := range resource {
			if key == "baseline" || key == "resource_type" || key == "resource_name" {
				continue
			}

			site := key
			namespace := extractNamespace(site)

			statusMap, ok := value.(map[string]interface{})
			if !ok {
				continue
			}

			status, _ := statusMap["status"].(string)

			var statusCode string
			switch status {
			case "match":
				statusCode = "P"
			case "mismatch", "not_found", "extra":
				statusCode = "F"
			default:
				statusCode = "-"
			}

			if statusCode == "P" {
				csv += fmt.Sprintf("%s,%s,%s,%s,,,%s\n", site, namespace, resourceName, resourceType, statusCode)
			} else if statusCode == "F" {
				// For mismatch, try to get diff details
				if status == "mismatch" {
					diffs := h.extractDiffDetails(baselineSite, site, namespace, resourceType, resourceName)
					if len(diffs) > 0 {
						for _, diff := range diffs {
							// Escape commas in values
							valueStr := fmt.Sprintf("Expected: %v, Got: %v", diff.BaselineValue, diff.TargetValue)
							valueStr = strings.ReplaceAll(valueStr, ",", ";")
							csv += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s\n",
								site, namespace, resourceName, resourceType, diff.Path, valueStr, statusCode)
						}
					} else {
						csv += fmt.Sprintf("%s,%s,%s,%s,Mismatch,Check diff API for details,%s\n",
							site, namespace, resourceName, resourceType, statusCode)
					}
				} else {
					csv += fmt.Sprintf("%s,%s,%s,%s,Status: %s,-,%s\n",
						site, namespace, resourceName, resourceType, status, statusCode)
				}
			}
		}
	}

	return []byte(csv), nil
}

// extractNamespace extracts namespace from site/namespace format
func extractNamespace(siteNS string) string {
	// siteNS format: "kind-kind-infra-test/app-prod"
	// Extract after last "/"
	for i := len(siteNS) - 1; i >= 0; i-- {
		if siteNS[i] == '/' {
			return siteNS[i+1:]
		}
	}
	return siteNS
}

// extractDiffDetails calls the diff API and extracts detailed differences
func (h *InfraValidationHandler) extractDiffDetails(baselineSite, targetSite, namespace, resourceType, resourceName string) []DiffDetail {
	// Construct baseline and target strings
	// Format expected: "context/namespace"
	baselineParts := strings.Split(baselineSite, "/")
	targetParts := strings.Split(targetSite, "/")

	if len(baselineParts) < 1 || len(targetParts) < 1 {
		return nil
	}

	baselineContext := baselineParts[0]
	targetContext := targetParts[0]
	baselineNS := namespace // Use the same namespace from target for baseline
	if len(baselineParts) >= 2 {
		baselineNS = baselineParts[1]
	}
	targetNS := namespace
	if len(targetParts) >= 2 {
		targetNS = targetParts[1]
	}

	// Get diff by calling internal logic
	baselineSiteConfig := infravalidation.SiteConfig{
		Context:        baselineContext,
		KubeconfigPath: h.kubeconfigPath,
	}

	targetSiteConfig := infravalidation.SiteConfig{
		Context:        targetContext,
		KubeconfigPath: h.kubeconfigPath,
	}

	// Create K8s clients
	baselineClient, err := infravalidation.NewK8sClient(baselineSiteConfig)
	if err != nil {
		return nil
	}

	targetClient, err := infravalidation.NewK8sClient(targetSiteConfig)
	if err != nil {
		return nil
	}

	ctx := context.Background()

	// Get resources
	baselineResource, err := baselineClient.GetResource(ctx, baselineNS, resourceName, infravalidation.ResourceType(resourceType))
	if err != nil {
		return nil
	}

	targetResource, err := targetClient.GetResource(ctx, targetNS, resourceName, infravalidation.ResourceType(resourceType))
	if err != nil {
		return nil
	}

	// Normalize resources
	normalizer := infravalidation.NewNormalizer(nil, infravalidation.SecretCompareKeys)
	normalizedBaseline, err := normalizer.Normalize(baselineResource)
	if err != nil {
		return nil
	}

	normalizedTarget, err := normalizer.Normalize(targetResource)
	if err != nil {
		return nil
	}

	// Convert to map for comparison
	baselineJSON, _ := infravalidation.ToJSON(normalizedBaseline)
	targetJSON, _ := infravalidation.ToJSON(normalizedTarget)

	var baselineMap, targetMap map[string]interface{}
	json.Unmarshal([]byte(baselineJSON), &baselineMap)
	json.Unmarshal([]byte(targetJSON), &targetMap)

	// Extract differences
	diffs := []DiffDetail{}
	compareMaps("", baselineMap, targetMap, &diffs)

	return diffs
}

// compareMaps recursively compares two maps and extracts differences
func compareMaps(path string, baseline, target map[string]interface{}, diffs *[]DiffDetail) {
	// Check all keys in baseline
	for key, baseValue := range baseline {
		currentPath := key
		if path != "" {
			currentPath = path + "/" + key
		}

		targetValue, exists := target[key]
		if !exists {
			// Key missing in target
			*diffs = append(*diffs, DiffDetail{
				Path:          currentPath,
				BaselineValue: baseValue,
				TargetValue:   nil,
			})
			continue
		}

		// Compare values
		if !reflect.DeepEqual(baseValue, targetValue) {
			// Check if both are maps - recurse
			baseMap, baseIsMap := baseValue.(map[string]interface{})
			targetMap, targetIsMap := targetValue.(map[string]interface{})
			if baseIsMap && targetIsMap {
				compareMaps(currentPath, baseMap, targetMap, diffs)
			} else {
				// Value difference
				*diffs = append(*diffs, DiffDetail{
					Path:          currentPath,
					BaselineValue: baseValue,
					TargetValue:   targetValue,
				})
			}
		}
	}

	// Check for keys that exist in target but not in baseline
	for key, targetValue := range target {
		if _, exists := baseline[key]; !exists {
			currentPath := key
			if path != "" {
				currentPath = path + "/" + key
			}
			*diffs = append(*diffs, DiffDetail{
				Path:          currentPath,
				BaselineValue: nil,
				TargetValue:   targetValue,
			})
		}
	}
}
