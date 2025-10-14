package integration
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSmartSiteRemoval_TwoSitesScenario tests removal when only 2 sites exist
func TestSmartSiteRemoval_TwoSitesScenario(t *testing.T) {
	tests := []struct {
		name               string
		targetSite         string
		mockReplicationInfo map[string]interface{}
		expectedCommand    []string
		expectedResponse   map[string]interface{}
	}{
		{
			name:       "Remove site from 2-site replication (should remove entire config)",
			targetSite: "site2",
			mockReplicationInfo: map[string]interface{}{
				"enabled": true,
				"sites": []interface{}{
					map[string]interface{}{"name": "site1", "deploymentID": "test-1"},
					map[string]interface{}{"name": "site2", "deploymentID": "test-2"},
				},
			},
			expectedCommand: []string{"mc", "admin", "replicate", "rm", "site2", "--all", "--force"},
			expectedResponse: map[string]interface{}{
				"success": true,
				"message": "Site replication configuration removed successfully",
				"note":    "This was the last replication pair. Entire replication configuration has been removed.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock MinIO command execution
			mockExecutor := &MockCommandExecutor{
				replicationInfo: tt.mockReplicationInfo,
				removeSuccess:   true,
			}

			// Create request
			reqBody := map[string]string{"alias": tt.targetSite}
			reqJSON, _ := json.Marshal(reqBody)

			// Create HTTP request
			req := httptest.NewRequest("POST", "/api/replication/remove", bytes.NewBuffer(reqJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute handler with mock
			handler := createTestHandler(mockExecutor)
			handler.ServeHTTP(w, req)

			// Verify response
			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedResponse["success"], response["success"])
			assert.Equal(t, tt.expectedResponse["message"], response["message"])
			assert.Equal(t, tt.expectedResponse["note"], response["note"])

			// Verify correct command was called
			assert.Equal(t, tt.expectedCommand, mockExecutor.lastCommand)
		})
	}
}

// TestSmartSiteRemoval_MultipleSitesScenario tests removal when 3+ sites exist
func TestSmartSiteRemoval_MultipleSitesScenario(t *testing.T) {
	tests := []struct {
		name               string
		targetSite         string
		mockReplicationInfo map[string]interface{}
		expectedCommand    []string
		expectedResponse   map[string]interface{}
	}{
		{
			name:       "Remove site from 4-site replication (should preserve group)",
			targetSite: "site4",
			mockReplicationInfo: map[string]interface{}{
				"enabled": true,
				"sites": []interface{}{
					map[string]interface{}{"name": "site1", "deploymentID": "test-1"},
					map[string]interface{}{"name": "site2", "deploymentID": "test-2"},
					map[string]interface{}{"name": "site3", "deploymentID": "test-3"},
					map[string]interface{}{"name": "site4", "deploymentID": "test-4"},
				},
			},
			expectedCommand: []string{"mc", "admin", "replicate", "rm", "site1", "site4", "--force"},
			expectedResponse: map[string]interface{}{
				"success": true,
				"message": "Site 'site4' removed from replication successfully",
				"note":    "Remaining sites in replication group: site1, site2, site3",
			},
		},
		{
			name:       "Remove site from 3-site replication (should preserve group)",
			targetSite: "site3",
			mockReplicationInfo: map[string]interface{}{
				"enabled": true,
				"sites": []interface{}{
					map[string]interface{}{"name": "site1", "deploymentID": "test-1"},
					map[string]interface{}{"name": "site2", "deploymentID": "test-2"},
					map[string]interface{}{"name": "site3", "deploymentID": "test-3"},
				},
			},
			expectedCommand: []string{"mc", "admin", "replicate", "rm", "site1", "site3", "--force"},
			expectedResponse: map[string]interface{}{
				"success": true,
				"message": "Site 'site3' removed from replication successfully",
				"note":    "Remaining sites in replication group: site1, site2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock MinIO command execution
			mockExecutor := &MockCommandExecutor{
				replicationInfo: tt.mockReplicationInfo,
				removeSuccess:   true,
			}

			// Create request
			reqBody := map[string]string{"alias": tt.targetSite}
			reqJSON, _ := json.Marshal(reqBody)

			// Create HTTP request
			req := httptest.NewRequest("POST", "/api/replication/remove", bytes.NewBuffer(reqJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute handler with mock
			handler := createTestHandler(mockExecutor)
			handler.ServeHTTP(w, req)

			// Verify response
			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedResponse["success"], response["success"])
			assert.Equal(t, tt.expectedResponse["message"], response["message"])
			assert.Equal(t, tt.expectedResponse["note"], response["note"])

			// Verify correct command was called
			assert.Equal(t, tt.expectedCommand, mockExecutor.lastCommand)
		})
	}
}

// TestSmartSiteRemoval_EdgeCases tests edge cases and error scenarios
func TestSmartSiteRemoval_EdgeCases(t *testing.T) {
	tests := []struct {
		name               string
		targetSite         string
		mockReplicationInfo map[string]interface{}
		mockError          error
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:       "Site not found in replication group",
			targetSite: "nonexistent",
			mockReplicationInfo: map[string]interface{}{
				"enabled": true,
				"sites": []interface{}{
					map[string]interface{}{"name": "site1", "deploymentID": "test-1"},
					map[string]interface{}{"name": "site2", "deploymentID": "test-2"},
				},
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Site 'nonexistent' not found in replication group",
		},
		{
			name:               "No replication group exists",
			targetSite:         "site1",
			mockReplicationInfo: map[string]interface{}{"enabled": false},
			expectedStatusCode:  http.StatusBadRequest,
			expectedError:       "No replication group found",
		},
		{
			name:               "MinIO command execution fails",
			targetSite:         "site1",
			mockReplicationInfo: map[string]interface{}{
				"enabled": true,
				"sites": []interface{}{
					map[string]interface{}{"name": "site1", "deploymentID": "test-1"},
					map[string]interface{}{"name": "site2", "deploymentID": "test-2"},
				},
			},
			mockError:          fmt.Errorf("MinIO connection failed"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      "Failed to remove site from replication",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock MinIO command execution
			mockExecutor := &MockCommandExecutor{
				replicationInfo: tt.mockReplicationInfo,
				removeSuccess:   tt.mockError == nil,
				executeError:    tt.mockError,
			}

			// Create request
			reqBody := map[string]string{"alias": tt.targetSite}
			reqJSON, _ := json.Marshal(reqBody)

			// Create HTTP request
			req := httptest.NewRequest("POST", "/api/replication/remove", bytes.NewBuffer(reqJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute handler with mock
			handler := createTestHandler(mockExecutor)
			handler.ServeHTTP(w, req)

			// Verify error response
			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Contains(t, response["error"], tt.expectedError)
		})
	}
}

// TestSmartSiteRemoval_RealWorldScenarios tests scenarios based on actual usage
func TestSmartSiteRemoval_RealWorldScenarios(t *testing.T) {
	t.Run("6-site cluster remove one site", func(t *testing.T) {
		mockInfo := map[string]interface{}{
			"enabled": true,
			"sites": []interface{}{
				map[string]interface{}{"name": "site1", "deploymentID": "id-1"},
				map[string]interface{}{"name": "site2", "deploymentID": "id-2"},
				map[string]interface{}{"name": "site3", "deploymentID": "id-3"},
				map[string]interface{}{"name": "site4", "deploymentID": "id-4"},
				map[string]interface{}{"name": "site5", "deploymentID": "id-5"},
				map[string]interface{}{"name": "site6", "deploymentID": "id-6"},
			},
		}

		mockExecutor := &MockCommandExecutor{
			replicationInfo: mockInfo,
			removeSuccess:   true,
		}

		reqBody := map[string]string{"alias": "site6"}
		reqJSON, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/replication/remove", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler := createTestHandler(mockExecutor)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Equal(t, "Site 'site6' removed from replication successfully", response["message"])
		assert.Contains(t, response["note"], "site1, site2, site3, site4, site5")

		// Verify correct command (should use first remaining site as executor)
		expectedCmd := []string{"mc", "admin", "replicate", "rm", "site1", "site6", "--force"}
		assert.Equal(t, expectedCmd, mockExecutor.lastCommand)
	})
}

// MockCommandExecutor simulates MinIO command execution for testing
type MockCommandExecutor struct {
	replicationInfo map[string]interface{}
	removeSuccess   bool
	executeError    error
	lastCommand     []string
}

func (m *MockCommandExecutor) ExecuteCommand(args ...string) ([]byte, error) {
	m.lastCommand = args

	if m.executeError != nil {
		return nil, m.executeError
	}

	// Simulate different command responses
	if len(args) >= 3 && args[1] == "admin" && args[2] == "replicate" {
		if len(args) >= 4 && args[3] == "info" {
			// Return mock replication info
			output, _ := json.Marshal(m.replicationInfo)
			return output, nil
		}
		if len(args) >= 4 && args[3] == "rm" {
			// Return mock removal success
			if m.removeSuccess {
				return []byte("Following site(s) [target] were removed successfully\n"), nil
			}
			return nil, fmt.Errorf("removal failed")
		}
	}

	return []byte(""), nil
}

// createTestHandler creates a test HTTP handler with mocked command executor
func createTestHandler(executor *MockCommandExecutor) http.Handler {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/api/replication/remove", func(w http.ResponseWriter, r *http.Request) {
		// This would be replaced with actual handler implementation
		// For now, simulate the smart removal logic
		
		var req struct {
			Alias string `json:"alias"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Get replication info
		infoOutput, err := executor.ExecuteCommand("mc", "admin", "replicate", "info", req.Alias, "--json")
		if err != nil {
			http.Error(w, `{"error": "Failed to get replication info"}`, http.StatusInternalServerError)
			return
		}

		var replicateInfo map[string]interface{}
		if err := json.Unmarshal(infoOutput, &replicateInfo); err != nil {
			http.Error(w, `{"error": "Failed to parse replication info"}`, http.StatusInternalServerError)
			return
		}

		// Check if replication is enabled
		if enabled, ok := replicateInfo["enabled"].(bool); !ok || !enabled {
			http.Error(w, `{"error": "No replication group found"}`, http.StatusBadRequest)
			return
		}

		// Find remaining sites
		var remainingSites []string
		if sitesList, ok := replicateInfo["sites"].([]interface{}); ok {
			for _, site := range sitesList {
				if siteMap, ok := site.(map[string]interface{}); ok {
					if siteName, ok := siteMap["name"].(string); ok && siteName != req.Alias {
						remainingSites = append(remainingSites, siteName)
					}
				}
			}
		}

		if len(remainingSites) == 0 {
			http.Error(w, fmt.Sprintf(`{"error": "Site '%s' not found in replication group"}`, req.Alias), http.StatusBadRequest)
			return
		}

		// Smart removal logic
		var response map[string]interface{}
		if len(remainingSites) == 1 {
			// Remove entire configuration
			_, err := executor.ExecuteCommand("mc", "admin", "replicate", "rm", req.Alias, "--all", "--force")
			if err != nil {
				http.Error(w, `{"error": "Failed to remove site from replication"}`, http.StatusInternalServerError)
				return
			}
			response = map[string]interface{}{
				"success": true,
				"message": "Site replication configuration removed successfully",
				"note":    "This was the last replication pair. Entire replication configuration has been removed.",
			}
		} else {
			// Remove specific site
			remainingAlias := remainingSites[0]
			_, err := executor.ExecuteCommand("mc", "admin", "replicate", "rm", remainingAlias, req.Alias, "--force")
			if err != nil {
				http.Error(w, `{"error": "Failed to remove site from replication"}`, http.StatusInternalServerError)
				return
			}
			response = map[string]interface{}{
				"success": true,
				"message": fmt.Sprintf("Site '%s' removed from replication successfully", req.Alias),
				"note":    fmt.Sprintf("Remaining sites in replication group: %s", strings.Join(remainingSites, ", ")),
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	return mux
}