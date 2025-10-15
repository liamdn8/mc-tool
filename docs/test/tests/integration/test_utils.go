package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// TestEnvironment represents a test environment setup
type TestEnvironment struct {
	Sites       []TestSite
	TempDir     string
	McConfigDir string
}

// TestSite represents a MinIO site for testing
type TestSite struct {
	Name      string
	Alias     string
	Endpoint  string
	AccessKey string
	SecretKey string
	Port      int
}

// SetupTestEnvironment creates a complete test environment with multiple MinIO sites
func SetupTestEnvironment(siteCount int) (*TestEnvironment, error) {
	tempDir, err := os.MkdirTemp("", "mc-tool-test-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %v", err)
	}

	mcConfigDir := filepath.Join(tempDir, ".mc")
	if err := os.MkdirAll(mcConfigDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create mc config directory: %v", err)
	}

	env := &TestEnvironment{
		Sites:       make([]TestSite, siteCount),
		TempDir:     tempDir,
		McConfigDir: mcConfigDir,
	}

	// Create test sites
	basePort := 9000
	for i := 0; i < siteCount; i++ {
		site := TestSite{
			Name:      fmt.Sprintf("site%d", i+1),
			Alias:     fmt.Sprintf("site%d", i+1),
			Endpoint:  fmt.Sprintf("http://127.0.0.1:%d", basePort+i),
			AccessKey: "testuser",
			SecretKey: "testpass123",
			Port:      basePort + i,
		}
		env.Sites[i] = site
	}

	return env, nil
}

// StartMinIOSites starts MinIO server instances for testing
func (env *TestEnvironment) StartMinIOSites() error {
	for i, site := range env.Sites {
		dataDir := filepath.Join(env.TempDir, fmt.Sprintf("data-site%d", i+1))
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return fmt.Errorf("failed to create data directory for %s: %v", site.Name, err)
		}

		// Start MinIO server in background
		cmd := exec.Command("minio", "server",
			"--address", fmt.Sprintf(":%d", site.Port),
			"--console-address", fmt.Sprintf(":%d", site.Port+100),
			dataDir)

		cmd.Env = append(os.Environ(),
			fmt.Sprintf("MINIO_ROOT_USER=%s", site.AccessKey),
			fmt.Sprintf("MINIO_ROOT_PASSWORD=%s", site.SecretKey),
		)

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start MinIO for %s: %v", site.Name, err)
		}

		// Wait for server to start
		if err := env.waitForMinIOServer(site); err != nil {
			return fmt.Errorf("MinIO server %s failed to start: %v", site.Name, err)
		}
	}

	return nil
}

// ConfigureMCClient configures mc client with test sites
func (env *TestEnvironment) ConfigureMCClient() error {
	// Set MC_CONFIG_DIR environment variable
	os.Setenv("MC_CONFIG_DIR", env.McConfigDir)

	for _, site := range env.Sites {
		cmd := exec.Command("mc", "alias", "set", site.Alias,
			site.Endpoint, site.AccessKey, site.SecretKey)
		cmd.Env = append(os.Environ(), fmt.Sprintf("MC_CONFIG_DIR=%s", env.McConfigDir))

		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to configure mc alias for %s: %v, output: %s",
				site.Name, err, string(output))
		}
	}

	return nil
}

// SetupReplication configures site replication for test sites
func (env *TestEnvironment) SetupReplication() error {
	if len(env.Sites) < 2 {
		return fmt.Errorf("at least 2 sites required for replication")
	}

	// Create aliases array
	aliases := make([]string, len(env.Sites))
	for i, site := range env.Sites {
		aliases[i] = site.Alias
	}

	// Setup replication
	args := append([]string{"admin", "replicate", "add"}, aliases...)
	cmd := exec.Command("mc", args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("MC_CONFIG_DIR=%s", env.McConfigDir))

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to setup replication: %v, output: %s", err, string(output))
	}

	return nil
}

// CreateTestBuckets creates test buckets with sample data
func (env *TestEnvironment) CreateTestBuckets() error {
	buckets := []string{"test-bucket-1", "test-bucket-2", "shared-bucket"}

	for _, site := range env.Sites {
		for _, bucket := range buckets {
			// Create bucket
			cmd := exec.Command("mc", "mb", fmt.Sprintf("%s/%s", site.Alias, bucket))
			cmd.Env = append(os.Environ(), fmt.Sprintf("MC_CONFIG_DIR=%s", env.McConfigDir))

			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to create bucket %s on %s: %v, output: %s",
					bucket, site.Name, err, string(output))
			}

			// Add sample files
			testFile := filepath.Join(env.TempDir, fmt.Sprintf("test-%s-%s.txt", site.Name, bucket))
			content := fmt.Sprintf("Test content for %s on %s\nTimestamp: %s",
				bucket, site.Name, time.Now().Format(time.RFC3339))

			if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to create test file: %v", err)
			}

			// Upload file
			cmd = exec.Command("mc", "cp", testFile,
				fmt.Sprintf("%s/%s/", site.Alias, bucket))
			cmd.Env = append(os.Environ(), fmt.Sprintf("MC_CONFIG_DIR=%s", env.McConfigDir))

			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to upload test file to %s/%s: %v, output: %s",
					site.Name, bucket, err, string(output))
			}
		}
	}

	return nil
}

// Cleanup cleans up the test environment
func (env *TestEnvironment) Cleanup() error {
	// Stop MinIO processes (they should stop when temp dir is removed)
	// Clean up temp directory
	if err := os.RemoveAll(env.TempDir); err != nil {
		return fmt.Errorf("failed to clean up temp directory: %v", err)
	}

	return nil
}

// GetReplicationInfo gets current replication information
func (env *TestEnvironment) GetReplicationInfo() (map[string]interface{}, error) {
	if len(env.Sites) == 0 {
		return nil, fmt.Errorf("no sites configured")
	}

	cmd := exec.Command("mc", "admin", "replicate", "info", env.Sites[0].Alias, "--json")
	cmd.Env = append(os.Environ(), fmt.Sprintf("MC_CONFIG_DIR=%s", env.McConfigDir))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get replication info: %v, output: %s", err, string(output))
	}

	var info map[string]interface{}
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse replication info: %v", err)
	}

	return info, nil
}

// RemoveSiteFromReplication removes a site from replication using smart removal logic
func (env *TestEnvironment) RemoveSiteFromReplication(siteAlias string) error {
	info, err := env.GetReplicationInfo()
	if err != nil {
		return err
	}

	// Get remaining sites
	var remainingSites []string
	if sitesList, ok := info["sites"].([]interface{}); ok {
		for _, site := range sitesList {
			if siteMap, ok := site.(map[string]interface{}); ok {
				if siteName, ok := siteMap["name"].(string); ok && siteName != siteAlias {
					remainingSites = append(remainingSites, siteName)
				}
			}
		}
	}

	var cmd *exec.Cmd
	if len(remainingSites) == 1 {
		// Remove entire replication config
		cmd = exec.Command("mc", "admin", "replicate", "rm", siteAlias, "--all", "--force")
	} else {
		// Remove specific site
		cmd = exec.Command("mc", "admin", "replicate", "rm", remainingSites[0], siteAlias, "--force")
	}

	cmd.Env = append(os.Environ(), fmt.Sprintf("MC_CONFIG_DIR=%s", env.McConfigDir))

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove site from replication: %v, output: %s", err, string(output))
	}

	return nil
}

// VerifyReplicationStatus verifies the replication status across sites
func (env *TestEnvironment) VerifyReplicationStatus() (bool, error) {
	for _, site := range env.Sites {
		cmd := exec.Command("mc", "admin", "replicate", "status", site.Alias, "--json")
		cmd.Env = append(os.Environ(), fmt.Sprintf("MC_CONFIG_DIR=%s", env.McConfigDir))

		output, err := cmd.CombinedOutput()
		if err != nil {
			return false, fmt.Errorf("failed to get replication status for %s: %v", site.Name, err)
		}

		var status map[string]interface{}
		if err := json.Unmarshal(output, &status); err != nil {
			return false, fmt.Errorf("failed to parse replication status for %s: %v", site.Name, err)
		}

		// Check if replication is healthy
		if enabled, ok := status["enabled"].(bool); !ok || !enabled {
			return false, nil
		}
	}

	return true, nil
}

// WaitForReplicationSync waits for replication to sync across all sites
func (env *TestEnvironment) WaitForReplicationSync(timeout time.Duration) error {
	start := time.Now()

	for time.Since(start) < timeout {
		synced, err := env.checkReplicationSync()
		if err != nil {
			return err
		}

		if synced {
			return nil
		}

		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("replication sync timeout after %v", timeout)
}

// checkReplicationSync checks if all sites are synchronized
func (env *TestEnvironment) checkReplicationSync() (bool, error) {
	// Compare bucket listings across all sites
	if len(env.Sites) < 2 {
		return true, nil
	}

	baseSite := env.Sites[0]
	cmd := exec.Command("mc", "ls", baseSite.Alias, "--json")
	cmd.Env = append(os.Environ(), fmt.Sprintf("MC_CONFIG_DIR=%s", env.McConfigDir))

	baseOutput, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to list buckets for %s: %v", baseSite.Name, err)
	}

	// Compare with other sites
	for i := 1; i < len(env.Sites); i++ {
		site := env.Sites[i]
		cmd := exec.Command("mc", "ls", site.Alias, "--json")
		cmd.Env = append(os.Environ(), fmt.Sprintf("MC_CONFIG_DIR=%s", env.McConfigDir))

		output, err := cmd.CombinedOutput()
		if err != nil {
			return false, fmt.Errorf("failed to list buckets for %s: %v", site.Name, err)
		}

		if string(baseOutput) != string(output) {
			return false, nil // Not yet synchronized
		}
	}

	return true, nil
}

// waitForMinIOServer waits for a MinIO server to become ready
func (env *TestEnvironment) waitForMinIOServer(site TestSite) error {
	timeout := 30 * time.Second
	start := time.Now()

	for time.Since(start) < timeout {
		cmd := exec.Command("mc", "admin", "info", site.Endpoint)
		if err := cmd.Run(); err == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("MinIO server %s did not start within timeout", site.Name)
}

// MockResponseGenerator generates realistic mock responses for testing
type MockResponseGenerator struct{}

// GenerateReplicationInfo generates mock replication info response
func (m *MockResponseGenerator) GenerateReplicationInfo(sites []string) map[string]interface{} {
	sitesList := make([]interface{}, len(sites))
	for i, site := range sites {
		sitesList[i] = map[string]interface{}{
			"name":         site,
			"deploymentID": fmt.Sprintf("deployment-%d", i+1),
			"endpoint":     fmt.Sprintf("https://%s.example.com:9000", site),
		}
	}

	return map[string]interface{}{
		"enabled": true,
		"sites":   sitesList,
	}
}

// GenerateReplicationStatus generates mock replication status response
func (m *MockResponseGenerator) GenerateReplicationStatus(sites []string) map[string]interface{} {
	sitesStatus := make(map[string]interface{})
	for _, site := range sites {
		sitesStatus[site] = map[string]interface{}{
			"status":            "healthy",
			"objectsReplicated": 1000,
			"bytesReplicated":   1048576,
		}
	}

	return map[string]interface{}{
		"replicatedBuckets": 3,
		"pendingObjects":    0,
		"failedObjects":     0,
		"lastSyncTime":      time.Now().UTC().Format(time.RFC3339),
		"sites":             sitesStatus,
	}
}

// GenerateCompareResponse generates mock consistency check response
func (m *MockResponseGenerator) GenerateCompareResponse(consistent bool) map[string]interface{} {
	response := map[string]interface{}{
		"consistent": consistent,
		"buckets": map[string]interface{}{
			"bucket1": map[string]interface{}{
				"versioning": map[string]interface{}{
					"site1": "Enabled",
					"site2": "Enabled",
				},
				"policy": map[string]interface{}{
					"site1": "consistent",
					"site2": "consistent",
				},
			},
		},
	}

	if !consistent {
		response["inconsistencies"] = []interface{}{
			map[string]interface{}{
				"bucket": "bucket1",
				"type":   "versioning",
				"issue":  "Versioning mismatch between sites",
			},
		}
	} else {
		response["inconsistencies"] = []interface{}{}
	}

	return response
}

// TestDataGenerator generates test data for various scenarios
type TestDataGenerator struct{}

// GenerateTestSites generates test site configurations
func (t *TestDataGenerator) GenerateTestSites(count int) []TestSite {
	sites := make([]TestSite, count)
	for i := 0; i < count; i++ {
		sites[i] = TestSite{
			Name:      fmt.Sprintf("site%d", i+1),
			Alias:     fmt.Sprintf("site%d", i+1),
			Endpoint:  fmt.Sprintf("https://site%d.example.com:9000", i+1),
			AccessKey: "testuser",
			SecretKey: "testpass123",
			Port:      9000 + i,
		}
	}
	return sites
}

// GenerateErrorScenarios generates various error scenarios for testing
func (t *TestDataGenerator) GenerateErrorScenarios() []struct {
	Name        string
	Error       error
	ExpectedMsg string
} {
	return []struct {
		Name        string
		Error       error
		ExpectedMsg string
	}{
		{
			Name:        "Connection Refused",
			Error:       fmt.Errorf("connection refused"),
			ExpectedMsg: "Unable to connect to MinIO server",
		},
		{
			Name:        "Access Denied",
			Error:       fmt.Errorf("Access Denied"),
			ExpectedMsg: "Permission denied",
		},
		{
			Name:        "Timeout",
			Error:       fmt.Errorf("timeout"),
			ExpectedMsg: "Request timed out",
		},
		{
			Name:        "Site Not Found",
			Error:       fmt.Errorf("site not in replication group"),
			ExpectedMsg: "Site not found in replication configuration",
		},
	}
}
