package config

import (
	"os"
	"strconv"
	"time"
)

// WebConfig holds web server configuration
type WebConfig struct {
	Port                int
	RefreshInterval     time.Duration
	AlertWebhookURL     string
	AuthMethod          string // none, basic, oauth2
	LogLevel            string
	LogFormat           string // text, json
	DataDir             string
	ReferenceConfigsDir string
}

// LoadWebConfig loads configuration from environment variables
func LoadWebConfig() *WebConfig {
	return &WebConfig{
		Port:                getEnvInt("WEB_PORT", 8080),
		RefreshInterval:     getEnvDuration("REFRESH_INTERVAL", 60*time.Second),
		AlertWebhookURL:     getEnv("ALERT_WEBHOOK_URL", ""),
		AuthMethod:          getEnv("AUTH_METHOD", "none"),
		LogLevel:            getEnv("LOG_LEVEL", "info"),
		LogFormat:           getEnv("LOG_FORMAT", "text"),
		DataDir:             getEnv("DATA_DIR", "./data"),
		ReferenceConfigsDir: getEnv("REFERENCE_CONFIGS_DIR", "/app/reference-configs"),
	}
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
