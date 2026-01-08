package infravalidation

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	// Check baseline
	if config.Baseline.Site == "" {
		return fmt.Errorf("baseline.site is required")
	}
	if config.Baseline.Namespace == "" {
		return fmt.Errorf("baseline.namespace is required")
	}

	// Check targets
	if len(config.Targets) == 0 {
		return fmt.Errorf("at least one target is required")
	}

	for i, target := range config.Targets {
		if target.Site == "" {
			return fmt.Errorf("targets[%d].site is required", i)
		}
		if target.Namespace == "" {
			return fmt.Errorf("targets[%d].namespace is required", i)
		}
	}

	// Check resource types
	if len(config.ResourceTypes) == 0 {
		// Default to common resource types
		config.ResourceTypes = []ResourceType{
			ResourceDeployment,
			ResourceStatefulSet,
			ResourceConfigMap,
			ResourceSecret,
			ResourceService,
		}
	}

	// Validate resource types
	validTypes := map[ResourceType]bool{
		ResourceDeployment:  true,
		ResourceStatefulSet: true,
		ResourceDaemonSet:   true,
		ResourceConfigMap:   true,
		ResourceSecret:      true,
		ResourceService:     true,
	}

	for i, rt := range config.ResourceTypes {
		if !validTypes[rt] {
			return fmt.Errorf("resourceTypes[%d]: invalid resource type %s", i, rt)
		}
	}

	// Default mode
	if config.Mode == "" {
		config.Mode = ModeA
	}

	// Validate mode
	if config.Mode != ModeA && config.Mode != ModeB {
		return fmt.Errorf("invalid mode: %s (must be %s or %s)", config.Mode, ModeA, ModeB)
	}

	// Default secret comparison
	if config.SecretComparison == "" {
		config.SecretComparison = SecretCompareKeys
	}

	// Validate secret comparison mode
	if config.SecretComparison != SecretCompareKeys && config.SecretComparison != SecretCompareHashed {
		return fmt.Errorf("invalid secretComparison: %s", config.SecretComparison)
	}

	return nil
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(config *Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
