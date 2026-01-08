package infravalidation

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// DisplayFormat represents the output format type
type DisplayFormat string

const (
	FormatSummary DisplayFormat = "summary"
	FormatTable   DisplayFormat = "table"
	FormatJSON    DisplayFormat = "json"
)

// DisplayOptions contains options for displaying the report
type DisplayOptions struct {
	Format     DisplayFormat
	OutputFile string
}

// Display outputs the validation report in the specified format
func Display(report *ValidationReport, opts DisplayOptions) error {
	// Handle output format
	switch opts.Format {
	case FormatJSON:
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal report: %w", err)
		}
		fmt.Println(string(data))
	case FormatTable:
		displayTable(report)
	case FormatSummary:
		displaySummary(report)
	default:
		return fmt.Errorf("invalid format: %s (must be summary, table, or json)", opts.Format)
	}

	// Save to file if requested
	if opts.OutputFile != "" {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal report: %w", err)
		}

		if err := os.WriteFile(opts.OutputFile, data, 0644); err != nil {
			return fmt.Errorf("failed to write report file: %w", err)
		}

		fmt.Printf("\n[OK] Report saved to: %s\n", opts.OutputFile)
	}

	// Return exit code based on results
	if report.Summary.MismatchCount > 0 || report.Summary.NotFoundCount > 0 {
		return fmt.Errorf("configuration drift detected")
	}

	return nil
}

func displaySummary(report *ValidationReport) {
	fmt.Printf("\n[*] Infrastructure Validation Report\n")
	fmt.Printf(strings.Repeat("=", 100) + "\n\n")

	fmt.Printf("[i] Baseline: %s/%s\n", report.Baseline.Site, report.Baseline.Namespace)
	fmt.Printf("[i] Timestamp: %s\n", report.Timestamp)
	fmt.Printf("[i] Mode: %s\n\n", report.Mode)

	// Per-target results
	for i, targetResult := range report.TargetResults {
		if i > 0 {
			fmt.Printf("\n")
		}
		fmt.Printf("[>] Target: %s/%s\n", targetResult.Target.Site, targetResult.Target.Namespace)
		fmt.Printf(strings.Repeat("-", 100) + "\n")

		// Group by resource type
		resourceGroups := make(map[ResourceType][]ComparisonResult)
		for _, result := range targetResult.Results {
			resourceGroups[result.ResourceType] = append(resourceGroups[result.ResourceType], result)
		}

		// Display each resource type
		for resourceType, results := range resourceGroups {
			matches := 0
			mismatches := 0
			notFound := 0
			extra := 0
			errors := 0

			for _, r := range results {
				switch r.Status {
				case StatusMatch:
					matches++
				case StatusMismatch:
					mismatches++
				case StatusNotFound:
					notFound++
				case StatusExtra:
					extra++
				case StatusError:
					errors++
				}
			}

			fmt.Printf("  %s:\n", resourceType)
			if matches > 0 {
				fmt.Printf("    [OK] %d match(es):\n", matches)
				for _, r := range results {
					if r.Status == StatusMatch {
						fmt.Printf("        - %s\n", r.ResourceName)
					}
				}
			}
			if mismatches > 0 {
				fmt.Printf("    [X] %d mismatch(es):\n", mismatches)
				for _, r := range results {
					if r.Status == StatusMismatch {
						fmt.Printf("        - %s\n", r.ResourceName)
					}
				}
			}
			if notFound > 0 {
				fmt.Printf("    [!] %d not found:\n", notFound)
				for _, r := range results {
					if r.Status == StatusNotFound {
						fmt.Printf("        - %s\n", r.ResourceName)
					}
				}
			}
			if extra > 0 {
				fmt.Printf("    [+] %d extra (not in baseline):\n", extra)
				for _, r := range results {
					if r.Status == StatusExtra {
						fmt.Printf("        - %s\n", r.ResourceName)
					}
				}
			}
			if errors > 0 {
				fmt.Printf("    [!] %d error(s):\n", errors)
				for _, r := range results {
					if r.Status == StatusError {
						fmt.Printf("        - %s: %s\n", r.ResourceName, r.Error)
					}
				}
			}
		}
	}

	fmt.Printf("\n" + strings.Repeat("=", 100) + "\n")
	fmt.Printf("[#] Overall Results\n")
	fmt.Printf(strings.Repeat("=", 100) + "\n\n")

	fmt.Printf("  Total Comparisons: %d\n", report.Summary.TotalComparisons)
	if report.Summary.TotalComparisons > 0 {
		fmt.Printf("  [OK] Matches:      %d (%.1f%%)\n",
			report.Summary.MatchCount,
			float64(report.Summary.MatchCount)*100/float64(report.Summary.TotalComparisons))
		fmt.Printf("  [X]  Mismatches:   %d (%.1f%%)\n",
			report.Summary.MismatchCount,
			float64(report.Summary.MismatchCount)*100/float64(report.Summary.TotalComparisons))
		fmt.Printf("  [!]  Not Found:    %d (%.1f%%)\n",
			report.Summary.NotFoundCount,
			float64(report.Summary.NotFoundCount)*100/float64(report.Summary.TotalComparisons))
		if report.Summary.ExtraCount > 0 {
			fmt.Printf("  [+]  Extra:        %d (%.1f%%)\n",
				report.Summary.ExtraCount,
				float64(report.Summary.ExtraCount)*100/float64(report.Summary.TotalComparisons))
		}
		if report.Summary.ErrorCount > 0 {
			fmt.Printf("  [!]  Errors:       %d (%.1f%%)\n",
				report.Summary.ErrorCount,
				float64(report.Summary.ErrorCount)*100/float64(report.Summary.TotalComparisons))
		}
	}

	fmt.Printf("\n" + strings.Repeat("=", 100) + "\n")

	// Status message
	if report.Summary.MismatchCount > 0 || report.Summary.NotFoundCount > 0 || report.Summary.ExtraCount > 0 {
		fmt.Printf("\n[!] Configuration drift detected!\n")
	} else {
		fmt.Printf("\n[OK] All configurations match!\n")
	}
}

func displayTable(report *ValidationReport) {
	fmt.Printf("\n[*] Infrastructure Validation Report (Table Format)\n")
	fmt.Printf(strings.Repeat("=", 100) + "\n\n")

	fmt.Printf("[i] Baseline: %s/%s\n", report.Baseline.Site, report.Baseline.Namespace)
	fmt.Printf("[i] Timestamp: %s\n\n", report.Timestamp)

	// Collect all resource names and types
	resourceMap := make(map[ResourceType]map[string]map[string]ComparisonStatus)

	for _, targetResult := range report.TargetResults {
		targetKey := fmt.Sprintf("%s/%s", targetResult.Target.Site, targetResult.Target.Namespace)

		for _, result := range targetResult.Results {
			if resourceMap[result.ResourceType] == nil {
				resourceMap[result.ResourceType] = make(map[string]map[string]ComparisonStatus)
			}
			if resourceMap[result.ResourceType][result.ResourceName] == nil {
				resourceMap[result.ResourceType][result.ResourceName] = make(map[string]ComparisonStatus)
			}
			resourceMap[result.ResourceType][result.ResourceName][targetKey] = result.Status
		}
	}

	// Get target keys
	targetKeys := []string{}
	for _, targetResult := range report.TargetResults {
		targetKeys = append(targetKeys, fmt.Sprintf("%s/%s", targetResult.Target.Site, targetResult.Target.Namespace))
	}

	// Calculate column width for targets based on longest target name
	targetColWidth := 20
	for _, key := range targetKeys {
		if len(key) > targetColWidth {
			targetColWidth = len(key)
		}
	}
	if targetColWidth > 50 {
		targetColWidth = 50
	}

	// Display table for each resource type
	for _, resourceType := range []ResourceType{
		ResourceDeployment,
		ResourceStatefulSet,
		ResourceDaemonSet,
		ResourceConfigMap,
		ResourceSecret,
		ResourceService,
	} {
		resources := resourceMap[resourceType]
		if len(resources) == 0 {
			continue
		}

		fmt.Printf("[%s]\n", resourceType)
		fmt.Printf(strings.Repeat("-", 50) + "\n")

		// Header
		fmt.Printf("%-50s", "Resource Name")
		for _, targetKey := range targetKeys {
			fmt.Printf(" %-*s", targetColWidth, truncateString(targetKey, targetColWidth))
		}
		fmt.Printf("\n")
		fmt.Printf(strings.Repeat("-", 50))
		for range targetKeys {
			fmt.Printf(" " + strings.Repeat("-", targetColWidth))
		}
		fmt.Printf("\n")

		// Rows
		for resourceName, statuses := range resources {
			fmt.Printf("%-50s", truncateString(resourceName, 100))
			for _, targetKey := range targetKeys {
				status := statuses[targetKey]
				var statusStr string
				switch status {
				case StatusMatch:
					statusStr = "[OK] Match"
				case StatusMismatch:
					statusStr = "[X] Mismatch"
				case StatusNotFound:
					statusStr = "[!] Not Found"
				case StatusError:
					statusStr = "[!] Error"
				default:
					statusStr = "-"
				}
				fmt.Printf(" %-*s", targetColWidth, statusStr)
			}
			fmt.Printf("\n")
		}
		fmt.Printf("\n")
	}

	fmt.Printf(strings.Repeat("=", 50) + "\n")
	fmt.Printf("[#] Overall Results\n")
	fmt.Printf(strings.Repeat("=", 50) + "\n\n")

	fmt.Printf("  Total Comparisons: %d\n", report.Summary.TotalComparisons)
	if report.Summary.TotalComparisons > 0 {
		fmt.Printf("  [OK] Matches:      %d (%.1f%%)\n",
			report.Summary.MatchCount,
			float64(report.Summary.MatchCount)*100/float64(report.Summary.TotalComparisons))
		fmt.Printf("  [X]  Mismatches:   %d (%.1f%%)\n",
			report.Summary.MismatchCount,
			float64(report.Summary.MismatchCount)*100/float64(report.Summary.TotalComparisons))
		fmt.Printf("  [!]  Not Found:    %d (%.1f%%)\n",
			report.Summary.NotFoundCount,
			float64(report.Summary.NotFoundCount)*100/float64(report.Summary.TotalComparisons))
		if report.Summary.ErrorCount > 0 {
			fmt.Printf("  [!]  Errors:       %d (%.1f%%)\n",
				report.Summary.ErrorCount,
				float64(report.Summary.ErrorCount)*100/float64(report.Summary.TotalComparisons))
		}
	}

	fmt.Printf("\n" + strings.Repeat("=", 50) + "\n")

	// Status message
	if report.Summary.MismatchCount > 0 || report.Summary.NotFoundCount > 0 {
		fmt.Printf("\n[!] Configuration drift detected!\n")
	} else {
		fmt.Printf("\n[OK] All configurations match!\n")
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
