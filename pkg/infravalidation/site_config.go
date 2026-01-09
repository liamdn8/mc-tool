package infravalidation

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadInfraConfig loads site configurations from YAML file
func LoadInfraConfig(path string) (*InfraConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read infra config file: %w", err)
	}

	var config InfraConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse infra config YAML: %w", err)
	}

	if err := validateInfraConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid infra configuration: %w", err)
	}

	return &config, nil
}

// LoadDefaultInfraConfig loads from ~/.mc-tool/infra-config.yaml
func LoadDefaultInfraConfig() (*InfraConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".mc-tool", "infra-config.yaml")
	return LoadInfraConfig(configPath)
}

// LoadSiteConfig loads a specific site configuration from default config file
func LoadSiteConfig(siteName string) (SiteConfig, error) {
	config, err := LoadDefaultInfraConfig()
	if err != nil {
		return SiteConfig{}, err
	}

	siteConfig, ok := config.Sites[siteName]
	if !ok {
		return SiteConfig{}, fmt.Errorf("site %s not found in config", siteName)
	}

	return siteConfig, nil
}

// validateInfraConfig validates the infra configuration
func validateInfraConfig(config *InfraConfig) error {
	if len(config.Sites) == 0 {
		return fmt.Errorf("no sites configured")
	}

	for name, site := range config.Sites {
		if site.Name == "" {
			site.Name = name
			config.Sites[name] = site
		}

		// Either context or endpoint must be specified
		if site.Context == "" && site.Endpoint == "" {
			return fmt.Errorf("site %s: either context or endpoint must be specified", name)
		}

		// If using legacy endpoint mode, token is required
		if site.Endpoint != "" && site.Token == "" && site.Context == "" {
			return fmt.Errorf("site %s: token is required when using endpoint", name)
		}
	}

	return nil
}

// SaveInfraConfig saves site configurations to YAML file
func SaveInfraConfig(config *InfraConfig, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal infra config: %w", err)
	}

	// Create directory if not exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write infra config file: %w", err)
	}

	return nil
}

// ParseSiteNamespace parses "site/namespace" format
func ParseSiteNamespace(input string) (site, namespace string, err error) {
	parts := splitTwo(input, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid format: expected 'site/namespace', got '%s'", input)
	}

	site = parts[0]
	namespace = parts[1]

	if site == "" || namespace == "" {
		return "", "", fmt.Errorf("invalid format: site and namespace cannot be empty")
	}

	return site, namespace, nil
}

// LoadKubeconfigContexts loads available contexts from kubeconfig file
func LoadKubeconfigContexts(kubeconfigPath string) ([]string, error) {
	if kubeconfigPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		kubeconfigPath = filepath.Join(homeDir, ".kube", "config")
	}

	// Check if file exists
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("kubeconfig file not found: %s", kubeconfigPath)
	}

	// Read kubeconfig file
	data, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	// Parse kubeconfig
	var config struct {
		Contexts []struct {
			Name string `yaml:"name"`
		} `yaml:"contexts"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	var contexts []string
	for _, ctx := range config.Contexts {
		contexts = append(contexts, ctx.Name)
	}

	if len(contexts) == 0 {
		return nil, fmt.Errorf("no contexts found in kubeconfig")
	}

	return contexts, nil
}

// GetCurrentKubeconfigContext gets the current context from kubeconfig
func GetCurrentKubeconfigContext(kubeconfigPath string) (string, error) {
	if kubeconfigPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		kubeconfigPath = filepath.Join(homeDir, ".kube", "config")
	}

	data, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	var config struct {
		CurrentContext string `yaml:"current-context"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	return config.CurrentContext, nil
}

// splitTwo splits string by separator, expecting exactly 2 parts
func splitTwo(s, sep string) []string {
	result := []string{}
	current := ""
	sepFound := false

	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] && !sepFound {
			result = append(result, current)
			current = ""
			sepFound = true
		} else {
			current += string(s[i])
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}
