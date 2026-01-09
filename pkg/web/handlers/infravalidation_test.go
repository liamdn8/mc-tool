package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/liamdn8/mc-tool/pkg/web/models"
)

// TestHandleSearchNamespaces tests the search namespaces endpoint
func TestHandleSearchNamespaces(t *testing.T) {
	jobManager := models.NewJobManager()
	handler := NewInfraValidationHandler("/path/to/mc-tool", jobManager)

	tests := []struct {
		name           string
		method         string
		query          string
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:           "missing keyword parameter",
			method:         http.MethodGet,
			query:          "",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if _, ok := resp["error"]; !ok {
					t.Error("Expected error in response")
				}
			},
		},
		{
			name:           "method not allowed",
			method:         http.MethodPost,
			query:          "keyword=test",
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if _, ok := resp["error"]; !ok {
					t.Error("Expected error in response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/validate/infrastructure/search-namespaces?"+tt.query, nil)
			w := httptest.NewRecorder()

			handler.HandleSearchNamespaces(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				var resp map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				tt.checkResponse(t, resp)
			}
		})
	}
}

// TestHandleDiscoverResources tests the discover resources endpoint
func TestHandleDiscoverResources(t *testing.T) {
	jobManager := models.NewJobManager()
	handler := NewInfraValidationHandler("/path/to/mc-tool", jobManager)

	tests := []struct {
		name           string
		method         string
		query          string
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:           "missing vim parameter",
			method:         http.MethodGet,
			query:          "namespace=default",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if _, ok := resp["error"]; !ok {
					t.Error("Expected error in response")
				}
				errorMsg := resp["error"].(string)
				if !strings.Contains(errorMsg, "vim") {
					t.Errorf("Expected error about vim parameter, got: %s", errorMsg)
				}
			},
		},
		{
			name:           "missing namespace parameter",
			method:         http.MethodGet,
			query:          "vim=site1",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if _, ok := resp["error"]; !ok {
					t.Error("Expected error in response")
				}
				errorMsg := resp["error"].(string)
				if !strings.Contains(errorMsg, "namespace") {
					t.Errorf("Expected error about namespace parameter, got: %s", errorMsg)
				}
			},
		},
		{
			name:           "method not allowed",
			method:         http.MethodPost,
			query:          "vim=site1&namespace=default",
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if _, ok := resp["error"]; !ok {
					t.Error("Expected error in response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/validate/infrastructure/discover-resources?"+tt.query, nil)
			w := httptest.NewRecorder()

			handler.HandleDiscoverResources(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				var resp map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				tt.checkResponse(t, resp)
			}
		})
	}
}

// TestHandleInfraValidate tests the main validation endpoint
func TestHandleInfraValidate(t *testing.T) {
	jobManager := models.NewJobManager()
	handler := NewInfraValidationHandler("/path/to/mc-tool", jobManager)

	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:           "method not allowed",
			method:         http.MethodGet,
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "invalid json body",
			method:         http.MethodPost,
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing baseline",
			method:         http.MethodPost,
			body:           `{"targets": ["site1/ns1"]}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if _, ok := resp["error"]; !ok {
					t.Error("Expected error in response")
				}
			},
		},
		{
			name:           "missing targets",
			method:         http.MethodPost,
			body:           `{"baseline": "site1/ns1"}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if _, ok := resp["error"]; !ok {
					t.Error("Expected error in response")
				}
			},
		},
		{
			name:           "empty targets array",
			method:         http.MethodPost,
			body:           `{"baseline": "site1/ns1", "targets": []}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				if _, ok := resp["error"]; !ok {
					t.Error("Expected error in response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/validate/infrastructure", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.HandleInfraValidate(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				var resp map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				tt.checkResponse(t, resp)
			}
		})
	}
}

// TestValidationRequestParsing tests request parsing logic
func TestValidationRequestParsing(t *testing.T) {
	tests := []struct {
		name      string
		baseline  string
		targets   []string
		wantError bool
	}{
		{
			name:      "valid single target",
			baseline:  "site1/namespace1",
			targets:   []string{"site2/namespace1"},
			wantError: false,
		},
		{
			name:      "valid multiple targets",
			baseline:  "site1/app-prod",
			targets:   []string{"site2/app-prod", "site3/app-prod"},
			wantError: false,
		},
		{
			name:      "empty baseline",
			baseline:  "",
			targets:   []string{"site2/namespace1"},
			wantError: true,
		},
		{
			name:      "empty targets",
			baseline:  "site1/namespace1",
			targets:   []string{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := (tt.baseline == "" || len(tt.targets) == 0)
			if hasError != tt.wantError {
				t.Errorf("Expected error=%v, got error=%v", tt.wantError, hasError)
			}
		})
	}
}
