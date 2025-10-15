package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReplicationInfoAPI tests the /api/replication/info endpoint
func TestReplicationInfoAPI(t *testing.T) {
	tests := []struct {
		name                string
		mockReplicationInfo map[string]interface{}
		expectedResponse    map[string]interface{}
		expectedStatusCode  int
	}{
		{
			name: "Valid replication group info",
			mockReplicationInfo: map[string]interface{}{
				"enabled": true,
				"sites": []interface{}{
					map[string]interface{}{
						"name":         "site1",
						"deploymentID": "deployment-1",
						"endpoint":     "https://site1.example.com:9000",
					},
					map[string]interface{}{
						"name":         "site2",
						"deploymentID": "deployment-2",
						"endpoint":     "https://site2.example.com:9000",
					},
				},
			},
			expectedResponse: map[string]interface{}{
				"enabled":    true,
				"totalSites": float64(2),
				"sites": []interface{}{
					map[string]interface{}{
						"name":         "site1",
						"deploymentID": "deployment-1",
						"endpoint":     "https://site1.example.com:9000",
						"status":       "healthy",
					},
					map[string]interface{}{
						"name":         "site2",
						"deploymentID": "deployment-2",
						"endpoint":     "https://site2.example.com:9000",
						"status":       "healthy",
					},
				},
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "No replication configured",
			mockReplicationInfo: map[string]interface{}{
				"enabled": false,
			},
			expectedResponse: map[string]interface{}{
				"enabled":    false,
				"totalSites": float64(0),
				"sites":      []interface{}{},
				"message":    "No site replication configured",
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &MockCommandExecutor{
				replicationInfo: tt.mockReplicationInfo,
			}

			req := httptest.NewRequest("GET", "/api/replication/info", nil)
			w := httptest.NewRecorder()

			handler := createAPITestHandler(mockExecutor)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedResponse["enabled"], response["enabled"])
			assert.Equal(t, tt.expectedResponse["totalSites"], response["totalSites"])
		})
	}
}

// TestReplicationAddAPI tests the /api/replication/add endpoint
func TestReplicationAddAPI(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        map[string]interface{}
		mockExecuteSuccess bool
		mockExecuteError   error
		expectedStatusCode int
		expectedResponse   map[string]interface{}
	}{
		{
			name: "Add sites successfully",
			requestBody: map[string]interface{}{
				"aliases": []string{"site1", "site2", "site3"},
			},
			mockExecuteSuccess: true,
			expectedStatusCode: http.StatusOK,
			expectedResponse: map[string]interface{}{
				"success": true,
				"message": "Site replication configured successfully",
				"sites":   []string{"site1", "site2", "site3"},
			},
		},
		{
			name: "Insufficient sites (less than 2)",
			requestBody: map[string]interface{}{
				"aliases": []string{"site1"},
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse: map[string]interface{}{
				"error": "At least 2 sites are required for replication",
			},
		},
		{
			name: "MinIO command fails",
			requestBody: map[string]interface{}{
				"aliases": []string{"site1", "site2"},
			},
			mockExecuteSuccess: false,
			mockExecuteError:   fmt.Errorf("connection refused"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse: map[string]interface{}{
				"error": "Failed to configure site replication",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &MockCommandExecutor{
				removeSuccess: tt.mockExecuteSuccess,
				executeError:  tt.mockExecuteError,
			}

			reqJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/replication/add", bytes.NewBuffer(reqJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := createAPITestHandler(mockExecutor)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedResponse["success"] != nil {
				assert.Equal(t, tt.expectedResponse["success"], response["success"])
				assert.Equal(t, tt.expectedResponse["message"], response["message"])
			} else {
				assert.Contains(t, response["error"], tt.expectedResponse["error"])
			}
		})
	}
}

// TestReplicationStatusAPI tests the /api/replication/status endpoint
func TestReplicationStatusAPI(t *testing.T) {
	tests := []struct {
		name               string
		mockStatusResponse map[string]interface{}
		expectedStatusCode int
	}{
		{
			name: "Healthy replication status",
			mockStatusResponse: map[string]interface{}{
				"replicatedBuckets": float64(5),
				"pendingObjects":    float64(0),
				"failedObjects":     float64(0),
				"lastSyncTime":      "2025-10-14T10:30:00Z",
				"sites": map[string]interface{}{
					"site1": map[string]interface{}{
						"status":            "healthy",
						"objectsReplicated": float64(1000),
						"bytesReplicated":   float64(1048576),
					},
					"site2": map[string]interface{}{
						"status":            "healthy",
						"objectsReplicated": float64(1000),
						"bytesReplicated":   float64(1048576),
					},
				},
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &MockCommandExecutor{
				statusResponse: tt.mockStatusResponse,
			}

			req := httptest.NewRequest("GET", "/api/replication/status", nil)
			w := httptest.NewRecorder()

			handler := createAPITestHandler(mockExecutor)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.mockStatusResponse["replicatedBuckets"], response["replicatedBuckets"])
			assert.Equal(t, tt.mockStatusResponse["pendingObjects"], response["pendingObjects"])
		})
	}
}

// TestReplicationResyncAPI tests the /api/replication/resync endpoint
func TestReplicationResyncAPI(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        map[string]interface{}
		mockExecuteSuccess bool
		expectedStatusCode int
		expectedResponse   map[string]interface{}
	}{
		{
			name: "Resync from source successfully",
			requestBody: map[string]interface{}{
				"sourceAlias": "site1",
				"targetAlias": "site2",
				"direction":   "resync-from",
			},
			mockExecuteSuccess: true,
			expectedStatusCode: http.StatusOK,
			expectedResponse: map[string]interface{}{
				"success": true,
				"message": "Resync started successfully",
			},
		},
		{
			name: "Resync to target successfully",
			requestBody: map[string]interface{}{
				"sourceAlias": "site1",
				"targetAlias": "site2",
				"direction":   "resync-to",
			},
			mockExecuteSuccess: true,
			expectedStatusCode: http.StatusOK,
			expectedResponse: map[string]interface{}{
				"success": true,
				"message": "Resync started successfully",
			},
		},
		{
			name: "Invalid direction",
			requestBody: map[string]interface{}{
				"sourceAlias": "site1",
				"targetAlias": "site2",
				"direction":   "invalid",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse: map[string]interface{}{
				"error": "Invalid direction. Must be 'resync-from' or 'resync-to'",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &MockCommandExecutor{
				removeSuccess: tt.mockExecuteSuccess,
			}

			reqJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/replication/resync", bytes.NewBuffer(reqJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := createAPITestHandler(mockExecutor)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedResponse["success"] != nil {
				assert.Equal(t, tt.expectedResponse["success"], response["success"])
			} else {
				assert.Contains(t, response["error"], tt.expectedResponse["error"])
			}
		})
	}
}

// TestReplicationCompareAPI tests the /api/replication/compare endpoint
func TestReplicationCompareAPI(t *testing.T) {
	tests := []struct {
		name                string
		mockCompareResponse map[string]interface{}
		expectedStatusCode  int
	}{
		{
			name: "Consistent configuration across sites",
			mockCompareResponse: map[string]interface{}{
				"consistent": true,
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
				"inconsistencies": []interface{}{},
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "Inconsistent configuration detected",
			mockCompareResponse: map[string]interface{}{
				"consistent": false,
				"buckets": map[string]interface{}{
					"bucket1": map[string]interface{}{
						"versioning": map[string]interface{}{
							"site1": "Enabled",
							"site2": "Suspended",
						},
					},
				},
				"inconsistencies": []interface{}{
					map[string]interface{}{
						"bucket": "bucket1",
						"type":   "versioning",
						"issue":  "Versioning mismatch between sites",
					},
				},
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &MockCommandExecutor{
				compareResponse: tt.mockCompareResponse,
			}

			req := httptest.NewRequest("GET", "/api/replication/compare", nil)
			w := httptest.NewRecorder()

			handler := createAPITestHandler(mockExecutor)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.mockCompareResponse["consistent"], response["consistent"])
		})
	}
}

// Enhanced MockCommandExecutor for all API tests
type MockCommandExecutor struct {
	replicationInfo map[string]interface{}
	statusResponse  map[string]interface{}
	compareResponse map[string]interface{}
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
		switch {
		case len(args) >= 4 && args[3] == "info":
			output, _ := json.Marshal(m.replicationInfo)
			return output, nil
		case len(args) >= 4 && args[3] == "status":
			output, _ := json.Marshal(m.statusResponse)
			return output, nil
		case len(args) >= 4 && args[3] == "add":
			if m.removeSuccess {
				return []byte("Site replication successfully set up\n"), nil
			}
			return nil, fmt.Errorf("setup failed")
		case len(args) >= 4 && args[3] == "rm":
			if m.removeSuccess {
				return []byte("Site removed successfully\n"), nil
			}
			return nil, fmt.Errorf("removal failed")
		case len(args) >= 4 && args[3] == "resync":
			if m.removeSuccess {
				return []byte("Resync started successfully\n"), nil
			}
			return nil, fmt.Errorf("resync failed")
		}
	}

	// Simulate bucket compare command
	if len(args) >= 2 && args[0] == "mc" && args[1] == "diff" {
		output, _ := json.Marshal(m.compareResponse)
		return output, nil
	}

	return []byte(""), nil
}

// createAPITestHandler creates a test handler for all API endpoints
func createAPITestHandler(executor *MockCommandExecutor) http.Handler {
	mux := http.NewServeMux()

	// GET /api/replication/info
	mux.HandleFunc("/api/replication/info", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		output, err := executor.ExecuteCommand("mc", "admin", "replicate", "info", "site1", "--json")
		if err != nil {
			http.Error(w, `{"error": "Failed to get replication info"}`, http.StatusInternalServerError)
			return
		}

		var info map[string]interface{}
		json.Unmarshal(output, &info)

		response := map[string]interface{}{
			"enabled":    info["enabled"],
			"totalSites": 0,
			"sites":      []interface{}{},
		}

		if enabled, ok := info["enabled"].(bool); ok && enabled {
			if sites, ok := info["sites"].([]interface{}); ok {
				response["totalSites"] = len(sites)
				for _, site := range sites {
					if siteMap, ok := site.(map[string]interface{}); ok {
						siteMap["status"] = "healthy" // Mock status
					}
				}
				response["sites"] = sites
			}
		} else {
			response["message"] = "No site replication configured"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// POST /api/replication/add
	mux.HandleFunc("/api/replication/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Aliases []string `json:"aliases"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

		if len(req.Aliases) < 2 {
			http.Error(w, `{"error": "At least 2 sites are required for replication"}`, http.StatusBadRequest)
			return
		}

		args := append([]string{"mc", "admin", "replicate", "add"}, req.Aliases...)
		_, err := executor.ExecuteCommand(args...)
		if err != nil {
			http.Error(w, `{"error": "Failed to configure site replication"}`, http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"success": true,
			"message": "Site replication configured successfully",
			"sites":   req.Aliases,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// GET /api/replication/status
	mux.HandleFunc("/api/replication/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		output, err := executor.ExecuteCommand("mc", "admin", "replicate", "status", "site1", "--json")
		if err != nil {
			http.Error(w, `{"error": "Failed to get replication status"}`, http.StatusInternalServerError)
			return
		}

		var status map[string]interface{}
		json.Unmarshal(output, &status)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	// POST /api/replication/resync
	mux.HandleFunc("/api/replication/resync", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			SourceAlias string `json:"sourceAlias"`
			TargetAlias string `json:"targetAlias"`
			Direction   string `json:"direction"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

		if req.Direction != "resync-from" && req.Direction != "resync-to" {
			http.Error(w, `{"error": "Invalid direction. Must be 'resync-from' or 'resync-to'"}`, http.StatusBadRequest)
			return
		}

		_, err := executor.ExecuteCommand("mc", "admin", "replicate", "resync", "start", req.SourceAlias, req.TargetAlias)
		if err != nil {
			http.Error(w, `{"error": "Failed to start resync"}`, http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"success": true,
			"message": "Resync started successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// GET /api/replication/compare
	mux.HandleFunc("/api/replication/compare", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		output, err := executor.ExecuteCommand("mc", "diff", "site1", "site2", "--json")
		if err != nil {
			http.Error(w, `{"error": "Failed to compare sites"}`, http.StatusInternalServerError)
			return
		}

		var compare map[string]interface{}
		json.Unmarshal(output, &compare)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(compare)
	})

	return mux
}
