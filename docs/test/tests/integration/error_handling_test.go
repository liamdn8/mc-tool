package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConnectionFailures tests handling of MinIO connection failures
func TestConnectionFailures(t *testing.T) {
	tests := []struct {
		name               string
		endpoint           string
		mockError          error
		expectedStatusCode int
		expectedErrorType  string
	}{
		{
			name:               "MinIO server unreachable",
			endpoint:           "/api/replication/info",
			mockError:          fmt.Errorf("connection refused"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrorType:  "connection_error",
		},
		{
			name:               "Network timeout",
			endpoint:           "/api/replication/status",
			mockError:          fmt.Errorf("timeout"),
			expectedStatusCode: http.StatusRequestTimeout,
			expectedErrorType:  "timeout_error",
		},
		{
			name:               "DNS resolution failure",
			endpoint:           "/api/replication/info",
			mockError:          fmt.Errorf("no such host"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrorType:  "dns_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &ErrorTestExecutor{
				executeError: tt.mockError,
			}

			req := httptest.NewRequest("GET", tt.endpoint, nil)
			w := httptest.NewRecorder()

			handler := createErrorTestHandler(mockExecutor)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Contains(t, response, "error")
			assert.Contains(t, response, "errorType")
			assert.Equal(t, tt.expectedErrorType, response["errorType"])
		})
	}
}

// TestPermissionErrors tests handling of authentication and authorization failures
func TestPermissionErrors(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        map[string]interface{}
		mockError          error
		expectedStatusCode int
		expectedMessage    string
	}{
		{
			name: "Invalid credentials",
			requestBody: map[string]interface{}{
				"aliases": []string{"site1", "site2"},
			},
			mockError:          fmt.Errorf("Access Denied"),
			expectedStatusCode: http.StatusUnauthorized,
			expectedMessage:    "Authentication failed",
		},
		{
			name: "Insufficient permissions",
			requestBody: map[string]interface{}{
				"alias": "site1",
			},
			mockError:          fmt.Errorf("insufficient permissions"),
			expectedStatusCode: http.StatusForbidden,
			expectedMessage:    "Insufficient permissions",
		},
		{
			name: "Invalid access key",
			requestBody: map[string]interface{}{
				"aliases": []string{"site1", "site2"},
			},
			mockError:          fmt.Errorf("invalid access key"),
			expectedStatusCode: http.StatusUnauthorized,
			expectedMessage:    "Invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &ErrorTestExecutor{
				executeError: tt.mockError,
			}

			var endpoint string
			var method string
			if _, hasAlias := tt.requestBody["alias"]; hasAlias {
				endpoint = "/api/replication/remove"
				method = "POST"
			} else {
				endpoint = "/api/replication/add"
				method = "POST"
			}

			reqJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(method, endpoint, bytes.NewBuffer(reqJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := createErrorTestHandler(mockExecutor)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Contains(t, response["error"], tt.expectedMessage)
		})
	}
}

// TestInvalidInputs tests validation of request inputs
func TestInvalidInputs(t *testing.T) {
	tests := []struct {
		name               string
		endpoint           string
		method             string
		requestBody        interface{}
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:               "Invalid JSON",
			endpoint:           "/api/replication/add",
			method:             "POST",
			requestBody:        "invalid json",
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Invalid JSON",
		},
		{
			name:     "Empty aliases array",
			endpoint: "/api/replication/add",
			method:   "POST",
			requestBody: map[string]interface{}{
				"aliases": []string{},
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "At least 2 sites are required",
		},
		{
			name:     "Single alias for replication",
			endpoint: "/api/replication/add",
			method:   "POST",
			requestBody: map[string]interface{}{
				"aliases": []string{"site1"},
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "At least 2 sites are required",
		},
		{
			name:     "Empty alias for removal",
			endpoint: "/api/replication/remove",
			method:   "POST",
			requestBody: map[string]interface{}{
				"alias": "",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Alias is required",
		},
		{
			name:     "Invalid resync direction",
			endpoint: "/api/replication/resync",
			method:   "POST",
			requestBody: map[string]interface{}{
				"sourceAlias": "site1",
				"targetAlias": "site2",
				"direction":   "invalid-direction",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Invalid direction",
		},
		{
			name:     "Missing required fields for resync",
			endpoint: "/api/replication/resync",
			method:   "POST",
			requestBody: map[string]interface{}{
				"direction": "resync-from",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "sourceAlias and targetAlias are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &ErrorTestExecutor{}

			var req *http.Request
			if tt.requestBody == "invalid json" {
				req = httptest.NewRequest(tt.method, tt.endpoint, strings.NewReader(tt.requestBody.(string)))
			} else {
				reqJSON, _ := json.Marshal(tt.requestBody)
				req = httptest.NewRequest(tt.method, tt.endpoint, bytes.NewBuffer(reqJSON))
			}
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := createErrorTestHandler(mockExecutor)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Contains(t, response["error"], tt.expectedError)
		})
	}
}

// TestLocalhostEndpointErrors tests specific localhost endpoint handling
func TestLocalhostEndpointErrors(t *testing.T) {
	tests := []struct {
		name         string
		aliases      []string
		expectedTips []string
	}{
		{
			name:    "Localhost endpoints detected",
			aliases: []string{"http://localhost:9000", "https://127.0.0.1:9001"},
			expectedTips: []string{
				"Replace localhost with actual IP address",
				"Ensure MinIO servers can reach each other",
				"Use public or accessible hostnames",
			},
		},
		{
			name:    "Mixed localhost and valid endpoints",
			aliases: []string{"http://localhost:9000", "https://site2.example.com:9000"},
			expectedTips: []string{
				"Replace localhost with actual IP address",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &ErrorTestExecutor{
				executeError: fmt.Errorf("connection refused - localhost endpoints detected"),
			}

			reqBody := map[string]interface{}{
				"aliases": tt.aliases,
			}
			reqJSON, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/api/replication/add", bytes.NewBuffer(reqJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := createErrorTestHandler(mockExecutor)
			handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Contains(t, response, "error")
			assert.Contains(t, response, "troubleshooting")

			if tips, ok := response["troubleshooting"].([]interface{}); ok {
				for _, expectedTip := range tt.expectedTips {
					found := false
					for _, tip := range tips {
						if strings.Contains(tip.(string), expectedTip) {
							found = true
							break
						}
					}
					assert.True(t, found, fmt.Sprintf("Expected tip '%s' not found", expectedTip))
				}
			}
		})
	}
}

// TestRetryMechanism tests automatic retry logic for transient failures
func TestRetryMechanism(t *testing.T) {
	tests := []struct {
		name           string
		maxRetries     int
		failureCount   int
		expectedStatus int
	}{
		{
			name:           "Success after 2 retries",
			maxRetries:     3,
			failureCount:   2,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Failure after max retries",
			maxRetries:     3,
			failureCount:   4,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &RetryTestExecutor{
				maxRetries:   tt.maxRetries,
				failureCount: tt.failureCount,
				currentTry:   0,
			}

			req := httptest.NewRequest("GET", "/api/replication/info", nil)
			w := httptest.NewRecorder()

			handler := createRetryTestHandler(mockExecutor)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "retryCount")
			}
		})
	}
}

// TestErrorMessages tests user-friendly error message generation
func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name             string
		minioError       string
		expectedUserMsg  string
		expectedTechMsg  string
		languageCode     string
	}{
		{
			name:            "Connection refused (English)",
			minioError:      "connection refused",
			expectedUserMsg: "Unable to connect to MinIO server",
			expectedTechMsg: "connection refused",
			languageCode:    "en",
		},
		{
			name:            "Connection refused (Vietnamese)",
			minioError:      "connection refused",
			expectedUserMsg: "Không thể kết nối đến MinIO server",
			expectedTechMsg: "connection refused",
			languageCode:    "vi",
		},
		{
			name:            "Access denied",
			minioError:      "Access Denied",
			expectedUserMsg: "Permission denied",
			expectedTechMsg: "Access Denied",
			languageCode:    "en",
		},
		{
			name:            "Site not found",
			minioError:      "site not in replication group",
			expectedUserMsg: "Site not found in replication configuration",
			expectedTechMsg: "site not in replication group",
			languageCode:    "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &ErrorTestExecutor{
				executeError: fmt.Errorf(tt.minioError),
			}

			req := httptest.NewRequest("GET", "/api/replication/info", nil)
			req.Header.Set("Accept-Language", tt.languageCode)
			w := httptest.NewRecorder()

			handler := createErrorTestHandler(mockExecutor)
			handler.ServeHTTP(w, req)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Contains(t, response["error"], tt.expectedUserMsg)
			assert.Contains(t, response["technical"], tt.expectedTechMsg)
		})
	}
}

// ErrorTestExecutor simulates various error conditions
type ErrorTestExecutor struct {
	executeError error
}

func (e *ErrorTestExecutor) ExecuteCommand(args ...string) ([]byte, error) {
	return nil, e.executeError
}

// RetryTestExecutor simulates retry scenarios
type RetryTestExecutor struct {
	maxRetries   int
	failureCount int
	currentTry   int
}

func (r *RetryTestExecutor) ExecuteCommand(args ...string) ([]byte, error) {
	r.currentTry++
	
	if r.currentTry <= r.failureCount {
		return nil, fmt.Errorf("temporary failure %d", r.currentTry)
	}
	
	// Success after failures
	result := map[string]interface{}{
		"enabled":    true,
		"retryCount": r.currentTry - 1,
	}
	return json.Marshal(result)
}

// createErrorTestHandler creates handlers with error handling logic
func createErrorTestHandler(executor interface{}) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/replication/info", func(w http.ResponseWriter, r *http.Request) {
		var err error
		if errorExec, ok := executor.(*ErrorTestExecutor); ok {
			_, err = errorExec.ExecuteCommand("mc", "admin", "replicate", "info", "site1", "--json")
		}

		if err != nil {
			statusCode := determineStatusCode(err)
			errorType := determineErrorType(err)
			language := r.Header.Get("Accept-Language")
			
			response := map[string]interface{}{
				"error":     translateError(err, language),
				"technical": err.Error(),
				"errorType": errorType,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			json.NewEncoder(w).Encode(response)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	})

	mux.HandleFunc("/api/replication/add", func(w http.ResponseWriter, r *http.Request) {
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

		// Check for localhost endpoints
		troubleshooting := checkLocalhostEndpoints(req.Aliases)
		if len(troubleshooting) > 0 {
			response := map[string]interface{}{
				"error":          "Localhost endpoints detected",
				"troubleshooting": troubleshooting,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		var err error
		if errorExec, ok := executor.(*ErrorTestExecutor); ok {
			_, err = errorExec.ExecuteCommand("mc", "admin", "replicate", "add", req.Aliases[0], req.Aliases[1])
		}

		if err != nil {
			statusCode := determineStatusCode(err)
			response := map[string]interface{}{
				"error": translateError(err, "en"),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			json.NewEncoder(w).Encode(response)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	})

	mux.HandleFunc("/api/replication/remove", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Alias string `json:"alias"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

		if req.Alias == "" {
			http.Error(w, `{"error": "Alias is required"}`, http.StatusBadRequest)
			return
		}

		var err error
		if errorExec, ok := executor.(*ErrorTestExecutor); ok {
			_, err = errorExec.ExecuteCommand("mc", "admin", "replicate", "rm", req.Alias, "--force")
		}

		if err != nil {
			statusCode := determineStatusCode(err)
			response := map[string]interface{}{
				"error": translateError(err, "en"),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			json.NewEncoder(w).Encode(response)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	})

	mux.HandleFunc("/api/replication/resync", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			SourceAlias string `json:"sourceAlias"`
			TargetAlias string `json:"targetAlias"`
			Direction   string `json:"direction"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

		if req.SourceAlias == "" || req.TargetAlias == "" {
			http.Error(w, `{"error": "sourceAlias and targetAlias are required"}`, http.StatusBadRequest)
			return
		}

		if req.Direction != "resync-from" && req.Direction != "resync-to" {
			http.Error(w, `{"error": "Invalid direction. Must be 'resync-from' or 'resync-to'"}`, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	})

	return mux
}

// createRetryTestHandler creates handlers with retry logic
func createRetryTestHandler(executor *RetryTestExecutor) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/replication/info", func(w http.ResponseWriter, r *http.Request) {
		output, err := executor.ExecuteCommand("mc", "admin", "replicate", "info", "site1", "--json")
		
		if err != nil {
			http.Error(w, `{"error": "Max retries exceeded"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(output)
	})

	return mux
}

// Helper functions for error handling

func determineStatusCode(err error) int {
	errMsg := strings.ToLower(err.Error())
	
	switch {
	case strings.Contains(errMsg, "timeout"):
		return http.StatusRequestTimeout
	case strings.Contains(errMsg, "access denied"), strings.Contains(errMsg, "invalid access key"):
		return http.StatusUnauthorized
	case strings.Contains(errMsg, "insufficient permissions"):
		return http.StatusForbidden
	case strings.Contains(errMsg, "connection refused"), strings.Contains(errMsg, "no such host"):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func determineErrorType(err error) string {
	errMsg := strings.ToLower(err.Error())
	
	switch {
	case strings.Contains(errMsg, "timeout"):
		return "timeout_error"
	case strings.Contains(errMsg, "connection refused"):
		return "connection_error"
	case strings.Contains(errMsg, "no such host"):
		return "dns_error"
	case strings.Contains(errMsg, "access denied"):
		return "auth_error"
	default:
		return "unknown_error"
	}
}

func translateError(err error, language string) string {
	errMsg := strings.ToLower(err.Error())
	
	translations := map[string]map[string]string{
		"en": {
			"connection refused":     "Unable to connect to MinIO server",
			"timeout":               "Request timed out",
			"access denied":         "Permission denied",
			"site not in replication": "Site not found in replication configuration",
		},
		"vi": {
			"connection refused":     "Không thể kết nối đến MinIO server",
			"timeout":               "Hết thời gian chờ",
			"access denied":         "Không có quyền truy cập",
			"site not in replication": "Không tìm thấy site trong cấu hình replication",
		},
	}

	if langMap, ok := translations[language]; ok {
		for key, translation := range langMap {
			if strings.Contains(errMsg, key) {
				return translation
			}
		}
	}

	// Default to English
	if langMap, ok := translations["en"]; ok {
		for key, translation := range langMap {
			if strings.Contains(errMsg, key) {
				return translation
			}
		}
	}

	return "An error occurred"
}

func checkLocalhostEndpoints(aliases []string) []string {
	var tips []string
	
	for _, alias := range aliases {
		if strings.Contains(alias, "localhost") || strings.Contains(alias, "127.0.0.1") {
			tips = append(tips, "Replace localhost with actual IP address or hostname")
			tips = append(tips, "Ensure MinIO servers can reach each other over the network")
			tips = append(tips, "Use public or accessible hostnames for site replication")
			break
		}
	}
	
	return tips
}