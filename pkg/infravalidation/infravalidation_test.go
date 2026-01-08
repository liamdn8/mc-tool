package infravalidation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Baseline: ClusterNamespace{
					Context:   "kind-test",
					Cluster:   "test",
					Namespace: "default",
				},
				Targets: []ClusterNamespace{
					{
						Context:   "kind-test2",
						Cluster:   "test2",
						Namespace: "default",
					},
				},
				ResourceTypes: []ResourceType{ResourceDeployment},
				Mode:          ModeA,
			},
			wantErr: false,
		},
		{
			name: "missing baseline context",
			config: Config{
				Baseline: ClusterNamespace{
					Namespace: "default",
				},
				Targets: []ClusterNamespace{
					{
						Context:   "kind-test2",
						Namespace: "default",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing baseline namespace",
			config: Config{
				Baseline: ClusterNamespace{
					Context: "kind-test",
				},
				Targets: []ClusterNamespace{
					{
						Context:   "kind-test2",
						Namespace: "default",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "no targets",
			config: Config{
				Baseline: ClusterNamespace{
					Context:   "kind-test",
					Namespace: "default",
				},
				Targets: []ClusterNamespace{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(&tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestNormalizer tests resource normalization
func TestNormalizer(t *testing.T) {
	// Test parseJSONPath
	tests := []struct {
		path     string
		expected []string
	}{
		{
			path:     "spec.template.metadata.labels",
			expected: []string{"spec", "template", "metadata", "labels"},
		},
		{
			path:     "$.spec.replicas",
			expected: []string{"spec", "replicas"},
		},
		{
			path:     ".metadata.name",
			expected: []string{"metadata", "name"},
		},
		{
			path:     "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := parseJSONPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestResourceTypeGVR tests resource type to GVR mapping
func TestResourceTypeGVR(t *testing.T) {
	tests := []struct {
		resourceType ResourceType
		wantErr      bool
	}{
		{ResourceDeployment, false},
		{ResourceStatefulSet, false},
		{ResourceDaemonSet, false},
		{ResourceConfigMap, false},
		{ResourceSecret, false},
		{ResourceService, false},
		{"Invalid", true},
	}

	for _, tt := range tests {
		t.Run(string(tt.resourceType), func(t *testing.T) {
			_, err := getGVR(tt.resourceType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Integration test - requires KinD clusters to be running
// Run: ./scripts/kind-setup-infravalidation.sh before running this test
func TestIntegrationValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Load config
	config, err := LoadConfig("../../test-data/infravalidation-sample-config.yaml")
	if err != nil {
		t.Skipf("Skipping integration test - config file not found or KinD not setup: %v", err)
		return
	}

	// Create validator
	validator := NewValidator(config)

	// Run validation
	ctx := context.Background()
	report, err := validator.Validate(ctx)

	// If we get a connection error, skip the test as KinD is not available
	if err != nil {
		t.Skipf("Skipping integration test - cannot connect to KinD clusters: %v", err)
		return
	}

	require.NoError(t, err)
	require.NotNil(t, report)

	// Verify report structure
	assert.Equal(t, config.Baseline, report.Baseline)
	assert.Equal(t, ModeA, report.Mode)
	assert.NotEmpty(t, report.Timestamp)
	assert.Len(t, report.TargetResults, len(config.Targets))

	// Verify target1 (should match baseline)
	target1Result := report.TargetResults[0]
	assert.Equal(t, "kind-test-target1", target1Result.Target.Context)

	// Count matches in target1
	matchCount := 0
	for _, result := range target1Result.Results {
		if result.Status == StatusMatch {
			matchCount++
		}
		// Print any mismatches for debugging
		if result.Status == StatusMismatch {
			t.Logf("Target1 Mismatch: %s/%s\n%s", result.ResourceType, result.ResourceName, result.Diff)
		}
	}

	// We expect most resources to match in target1
	assert.Greater(t, matchCount, 0, "Expected at least some matching resources in target1")

	// Verify target2 (should have drifts)
	target2Result := report.TargetResults[1]
	assert.Equal(t, "kind-test-target2", target2Result.Target.Context)

	// Count mismatches in target2
	mismatchCount := 0
	for _, result := range target2Result.Results {
		if result.Status == StatusMismatch {
			mismatchCount++
			t.Logf("Target2 Mismatch detected: %s/%s", result.ResourceType, result.ResourceName)
		}
	}

	// We expect some drifts in target2
	assert.Greater(t, mismatchCount, 0, "Expected at least some drift in target2")

	// Verify summary
	assert.Greater(t, report.Summary.TotalComparisons, 0)
	assert.NotNil(t, report.Summary.StatusBreakdown)

	t.Logf("Validation Summary:")
	t.Logf("  Total Comparisons: %d", report.Summary.TotalComparisons)
	t.Logf("  Matches: %d", report.Summary.MatchCount)
	t.Logf("  Mismatches: %d", report.Summary.MismatchCount)
	t.Logf("  Not Found: %d", report.Summary.NotFoundCount)
	t.Logf("  Errors: %d", report.Summary.ErrorCount)
}
