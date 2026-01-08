package infravalidation

import (
	"context"
	"fmt"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
	"k8s.io/apimachinery/pkg/api/errors"
)

// Comparator handles resource comparison logic
type Comparator struct {
	normalizer *Normalizer
}

// NewComparator creates a new comparator
func NewComparator(normalizer *Normalizer) *Comparator {
	return &Comparator{
		normalizer: normalizer,
	}
}

// Compare compares a baseline resource with a target resource
func (c *Comparator) Compare(baseline, target *Resource) ComparisonResult {
	result := ComparisonResult{
		ResourceType: ResourceType(baseline.Kind),
		ResourceName: baseline.Name,
	}

	// Normalize both resources
	normalizedBaseline, err := c.normalizer.Normalize(baseline)
	if err != nil {
		result.Status = StatusError
		result.Error = fmt.Sprintf("failed to normalize baseline: %v", err)
		return result
	}

	normalizedTarget, err := c.normalizer.Normalize(target)
	if err != nil {
		result.Status = StatusError
		result.Error = fmt.Sprintf("failed to normalize target: %v", err)
		return result
	}

	// Convert to JSON for comparison
	baselineJSON, err := ToJSON(normalizedBaseline)
	if err != nil {
		result.Status = StatusError
		result.Error = fmt.Sprintf("failed to marshal baseline: %v", err)
		return result
	}

	targetJSON, err := ToJSON(normalizedTarget)
	if err != nil {
		result.Status = StatusError
		result.Error = fmt.Sprintf("failed to marshal target: %v", err)
		return result
	}

	// Compare
	if baselineJSON == targetJSON {
		result.Status = StatusMatch
		return result
	}

	// Generate diff
	result.Status = StatusMismatch
	result.Diff = generateDiff(baselineJSON, targetJSON)

	return result
}

// generateDiff creates a human-readable diff
func generateDiff(baseline, target string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(baseline, target, false)
	return dmp.DiffPrettyText(diffs)
}

// Validator orchestrates the validation process
type Validator struct {
	config     *Config
	comparator *Comparator
}

// NewValidator creates a new validator
func NewValidator(config *Config) *Validator {
	normalizer := NewNormalizer(config.IgnoreRules, config.SecretComparison)
	comparator := NewComparator(normalizer)

	return &Validator{
		config:     config,
		comparator: comparator,
	}
}

// Validate performs the complete validation
func (v *Validator) Validate(ctx context.Context) (*ValidationReport, error) {
	report := &ValidationReport{
		Baseline:      v.config.Baseline,
		Timestamp:     time.Now().Format(time.RFC3339),
		Mode:          v.config.Mode,
		TargetResults: make([]NamespaceComparisonResult, 0, len(v.config.Targets)),
		Summary: ValidationSummary{
			StatusBreakdown: make(map[string]int),
		},
	}

	// Get baseline site config
	baselineSiteConfig, ok := v.config.SiteConfigs[v.config.Baseline.Site]
	if !ok {
		return nil, fmt.Errorf("site config not found for baseline site: %s", v.config.Baseline.Site)
	}

	// Create baseline client
	baselineClient, err := NewK8sClient(baselineSiteConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create baseline client: %w", err)
	}

	// Fetch baseline resources
	baselineResources, err := v.fetchAllResources(ctx, baselineClient, v.config.Baseline.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch baseline resources: %w", err)
	}

	// Compare with each target
	for _, target := range v.config.Targets {
		targetResult, err := v.compareWithTarget(ctx, baselineResources, target)
		if err != nil {
			return nil, fmt.Errorf("failed to compare with target %s/%s: %w", target.Site, target.Namespace, err)
		}

		report.TargetResults = append(report.TargetResults, *targetResult)

		// Update summary
		for _, result := range targetResult.Results {
			report.Summary.TotalComparisons++
			report.Summary.StatusBreakdown[string(result.Status)]++

			switch result.Status {
			case StatusMatch:
				report.Summary.MatchCount++
			case StatusMismatch:
				report.Summary.MismatchCount++
			case StatusNotFound:
				report.Summary.NotFoundCount++
			case StatusError:
				report.Summary.ErrorCount++
			}
		}
	}

	return report, nil
}

// fetchAllResources fetches all configured resource types from a namespace
func (v *Validator) fetchAllResources(ctx context.Context, client *K8sClient, namespace string) (map[ResourceType]map[string]*Resource, error) {
	resources := make(map[ResourceType]map[string]*Resource)

	for _, resourceType := range v.config.ResourceTypes {
		resourceList, err := client.GetResources(ctx, namespace, resourceType)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s: %w", resourceType, err)
		}

		resourceMap := make(map[string]*Resource)
		for i := range resourceList {
			resourceMap[resourceList[i].Name] = &resourceList[i]
		}
		resources[resourceType] = resourceMap
	}

	return resources, nil
}

// compareWithTarget compares baseline resources with a target namespace
func (v *Validator) compareWithTarget(
	ctx context.Context,
	baselineResources map[ResourceType]map[string]*Resource,
	target ClusterNamespace,
) (*NamespaceComparisonResult, error) {
	result := &NamespaceComparisonResult{
		Target:  target,
		Results: make([]ComparisonResult, 0),
	}

	// Get target site config
	targetSiteConfig, ok := v.config.SiteConfigs[target.Site]
	if !ok {
		return nil, fmt.Errorf("site config not found for target site: %s", target.Site)
	}

	// Create target client
	targetClient, err := NewK8sClient(targetSiteConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create target client: %w", err)
	}

	// Compare each resource type
	for _, resourceType := range v.config.ResourceTypes {
		baselineResourceMap := baselineResources[resourceType]

		for name, baselineResource := range baselineResourceMap {
			// Fetch target resource
			targetResource, err := targetClient.GetResource(ctx, target.Namespace, name, resourceType)

			if err != nil {
				if errors.IsNotFound(err) {
					result.Results = append(result.Results, ComparisonResult{
						ResourceType: resourceType,
						ResourceName: name,
						Status:       StatusNotFound,
					})
					continue
				}

				result.Results = append(result.Results, ComparisonResult{
					ResourceType: resourceType,
					ResourceName: name,
					Status:       StatusError,
					Error:        err.Error(),
				})
				continue
			}

			// Compare
			comparisonResult := v.comparator.Compare(baselineResource, targetResource)
			result.Results = append(result.Results, comparisonResult)
		}
	}

	return result, nil
}
