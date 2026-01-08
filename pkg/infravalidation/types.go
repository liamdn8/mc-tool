package infravalidation

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ComparisonMode defines the type of comparison to perform
type ComparisonMode string

const (
	// ModeA performs structural configuration diff
	ModeA ComparisonMode = "mode-a"
	// ModeB performs policy/invariant validation (future)
	ModeB ComparisonMode = "mode-b"
)

// ComparisonStatus represents the outcome of a resource comparison
type ComparisonStatus string

const (
	StatusMatch     ComparisonStatus = "match"
	StatusMismatch  ComparisonStatus = "mismatch"
	StatusNotFound  ComparisonStatus = "not-found"
	StatusNotConfig ComparisonStatus = "not-config"
	StatusError     ComparisonStatus = "error"
)

// ClusterNamespace represents a Kubernetes cluster and namespace pair
type ClusterNamespace struct {
	Site      string `yaml:"site" json:"site"`
	Namespace string `yaml:"namespace" json:"namespace"`
}

// SiteConfig represents a Kubernetes cluster configuration
type SiteConfig struct {
	Name     string `yaml:"name" json:"name"`
	Endpoint string `yaml:"endpoint" json:"endpoint"`
	Token    string `yaml:"token" json:"token"`
	Insecure bool   `yaml:"insecure" json:"insecure"` // Skip TLS verification
}

// InfraConfig represents the sites configuration file
type InfraConfig struct {
	Sites map[string]SiteConfig `yaml:"sites" json:"sites"`
}

// ResourceType represents a Kubernetes resource kind
type ResourceType string

const (
	ResourceDeployment  ResourceType = "Deployment"
	ResourceStatefulSet ResourceType = "StatefulSet"
	ResourceDaemonSet   ResourceType = "DaemonSet"
	ResourceConfigMap   ResourceType = "ConfigMap"
	ResourceSecret      ResourceType = "Secret"
	ResourceService     ResourceType = "Service"
)

// Config represents the tool configuration
type Config struct {
	Baseline         ClusterNamespace      `yaml:"baseline" json:"baseline"`
	Targets          []ClusterNamespace    `yaml:"targets" json:"targets"`
	ResourceTypes    []ResourceType        `yaml:"resourceTypes" json:"resourceTypes"`
	Mode             ComparisonMode        `yaml:"mode" json:"mode"`
	SecretComparison SecretCompareMode     `yaml:"secretComparison" json:"secretComparison"`
	IgnoreRules      []IgnoreRule          `yaml:"ignoreRules" json:"ignoreRules"`
	SiteConfigs      map[string]SiteConfig // Runtime site configs
}

// SecretCompareMode defines how secrets are compared
type SecretCompareMode string

const (
	SecretCompareKeys   SecretCompareMode = "keys-only"
	SecretCompareHashed SecretCompareMode = "hashed-values"
)

// IgnoreRule defines patterns to ignore during comparison
type IgnoreRule struct {
	ResourceType ResourceType `yaml:"resourceType" json:"resourceType"`
	JSONPath     string       `yaml:"jsonPath" json:"jsonPath"`
}

// Resource represents a normalized Kubernetes resource
type Resource struct {
	Kind      string
	Name      string
	Namespace string
	Object    *unstructured.Unstructured
}

// ComparisonResult represents the result of comparing one resource
type ComparisonResult struct {
	ResourceType ResourceType     `json:"resourceType"`
	ResourceName string           `json:"resourceName"`
	Status       ComparisonStatus `json:"status"`
	Diff         string           `json:"diff,omitempty"`
	Error        string           `json:"error,omitempty"`
}

// NamespaceComparisonResult represents comparison results for one target namespace
type NamespaceComparisonResult struct {
	Target  ClusterNamespace   `json:"target"`
	Results []ComparisonResult `json:"results"`
}

// ValidationReport represents the complete validation report
type ValidationReport struct {
	Baseline      ClusterNamespace            `json:"baseline"`
	Timestamp     string                      `json:"timestamp"`
	Mode          ComparisonMode              `json:"mode"`
	TargetResults []NamespaceComparisonResult `json:"targetResults"`
	Summary       ValidationSummary           `json:"summary"`
}

// ValidationSummary provides aggregate statistics
type ValidationSummary struct {
	TotalComparisons int            `json:"totalComparisons"`
	MatchCount       int            `json:"matchCount"`
	MismatchCount    int            `json:"mismatchCount"`
	NotFoundCount    int            `json:"notFoundCount"`
	ErrorCount       int            `json:"errorCount"`
	StatusBreakdown  map[string]int `json:"statusBreakdown"`
}
