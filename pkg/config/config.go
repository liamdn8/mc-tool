package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// MCConfig represents the MinIO client configuration
type MCConfig struct {
	Version string                 `json:"version"`
	Aliases map[string]AliasConfig `json:"aliases"`
}

// AliasConfig represents a single alias configuration
type AliasConfig struct {
	URL       string `json:"url"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	API       string `json:"api"`
	Path      string `json:"path"`
	Insecure  bool   `json:"insecure,omitempty"`
}

// LoadMCConfig loads the MinIO client configuration from the default location
// and merges with environment variables
func LoadMCConfig() (*MCConfig, error) {
	config := &MCConfig{
		Version: "10",
		Aliases: make(map[string]AliasConfig),
	}

	// Try to load from file first
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".mc", "config.json")
		data, err := os.ReadFile(configPath)
		if err == nil {
			var fileConfig MCConfig
			if err := json.Unmarshal(data, &fileConfig); err == nil {
				config.Version = fileConfig.Version
				config.Aliases = fileConfig.Aliases
			}
		}
	}

	// Load aliases from environment variables (MC_HOST_<alias> format)
	// Environment variables take precedence over file configuration
	envAliases, err := loadAliasesFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load aliases from environment: %v", err)
	}

	// Merge environment aliases (they override file config)
	for alias, aliasConfig := range envAliases {
		config.Aliases[alias] = aliasConfig
	}

	// If no aliases found at all, return error
	if len(config.Aliases) == 0 {
		return nil, fmt.Errorf("no MinIO aliases configured in ~/.mc/config.json or MC_HOST_* environment variables")
	}

	return config, nil
}

// loadAliasesFromEnv loads alias configurations from MC_HOST_<alias> environment variables
// Format: MC_HOST_<alias>=https://accessKey:secretKey@hostname:port
// or: MC_HOST_<alias>=http://accessKey:secretKey@hostname:port
func loadAliasesFromEnv() (map[string]AliasConfig, error) {
	aliases := make(map[string]AliasConfig)

	for _, env := range os.Environ() {
		// Check if it's an MC_HOST_* variable
		if !strings.HasPrefix(env, "MC_HOST_") {
			continue
		}

		// Split into key=value
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		// Extract alias name (everything after MC_HOST_)
		aliasName := strings.TrimPrefix(parts[0], "MC_HOST_")
		aliasName = strings.ToLower(aliasName) // Normalize to lowercase
		if aliasName == "" {
			continue
		}

		// Parse the URL
		aliasValue := parts[1]
		aliasConfig, err := parseAliasURL(aliasValue)
		if err != nil {
			return nil, fmt.Errorf("failed to parse MC_HOST_%s: %v", aliasName, err)
		}

		aliases[aliasName] = aliasConfig
	}

	return aliases, nil
}

// parseAliasURL parses MinIO alias URL in format: scheme://accessKey:secretKey@hostname:port
func parseAliasURL(rawURL string) (AliasConfig, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return AliasConfig{}, fmt.Errorf("invalid URL: %v", err)
	}

	// Extract credentials
	var accessKey, secretKey string
	if parsedURL.User != nil {
		accessKey = parsedURL.User.Username()
		secretKey, _ = parsedURL.User.Password()
	}

	// Build base URL without credentials
	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	// Determine if insecure (self-signed certs)
	insecure := false
	if parsedURL.Query().Get("insecure") == "true" {
		insecure = true
	}

	return AliasConfig{
		URL:       baseURL,
		AccessKey: accessKey,
		SecretKey: secretKey,
		API:       "s3v4",
		Path:      "auto",
		Insecure:  insecure,
	}, nil
}

// GetMCEnvironment returns environment variables for mc/mc21 commands
// This allows mc commands to read alias configurations from environment variables
func GetMCEnvironment(config *MCConfig) []string {
	// Start with current environment
	env := os.Environ()

	// Add MC_HOST_<alias> variables for each configured alias
	for aliasName, aliasConfig := range config.Aliases {
		// Format: MC_HOST_<ALIAS>=https://accessKey:secretKey@hostname:port
		aliasURL := fmt.Sprintf("%s://%s:%s@%s",
			getSchemeFromURL(aliasConfig.URL),
			aliasConfig.AccessKey,
			aliasConfig.SecretKey,
			getHostFromURL(aliasConfig.URL))

		// Add insecure query parameter if needed
		if aliasConfig.Insecure {
			aliasURL += "?insecure=true"
		}

		envVar := fmt.Sprintf("MC_HOST_%s=%s", strings.ToUpper(aliasName), aliasURL)
		env = append(env, envVar)
	}

	return env
}

// getSchemeFromURL extracts scheme from URL
func getSchemeFromURL(rawURL string) string {
	if strings.HasPrefix(rawURL, "https://") {
		return "https"
	}
	return "http"
}

// getHostFromURL extracts host from URL
func getHostFromURL(rawURL string) string {
	rawURL = strings.TrimPrefix(rawURL, "https://")
	rawURL = strings.TrimPrefix(rawURL, "http://")
	return rawURL
}
