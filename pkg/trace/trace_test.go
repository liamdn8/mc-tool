package trace

import (
	"encoding/json"
	"testing"
)

func TestExtractObjectFromPath(t *testing.T) {
	sample := []byte(`{
        "path": "/nonexistent-bucket/",
        "statusCode": 404,
        "statusMsg": "Not Found"
    }`)

	entry := map[string]interface{}{}
	if err := json.Unmarshal(sample, &entry); err != nil {
		t.Fatalf("failed to unmarshal sample: %v", err)
	}

	got := extractObject(entry)
	if got == "" {
		t.Fatalf("expected non-empty object, got empty string")
	}

	if got != "nonexistent-bucket/" {
		t.Fatalf("expected 'nonexistent-bucket/', got %q", got)
	}
}

func TestExtractClient(t *testing.T) {
	sample := []byte(`{
        "remoteHost": "192.0.2.10",
        "req": {
            "api": "GetObject",
            "remoteAddr": "192.0.2.10:443"
        }
    }`)

	entry := map[string]interface{}{}
	if err := json.Unmarshal(sample, &entry); err != nil {
		t.Fatalf("failed to unmarshal sample: %v", err)
	}

	got := extractClient(entry)
	if got != "192.0.2.10" {
		t.Fatalf("expected client '192.0.2.10', got %q", got)
	}
}

func TestExtractVersionID(t *testing.T) {
	tests := []struct {
		name     string
		sample   string
		expected string
	}{
		{
			name: "versionId field",
			sample: `{
				"object": "test.txt",
				"versionId": "abc123def456"
			}`,
			expected: "abc123def456",
		},
		{
			name: "version in query params",
			sample: `{
				"object": "test.txt",
				"req": {
					"query": {
						"versionId": "xyz789"
					}
				}
			}`,
			expected: "xyz789",
		},
		{
			name: "version in request headers",
			sample: `{
				"object": "test.txt",
				"req": {
					"headers": {
						"X-Amz-Version-Id": "header-version-123"
					}
				}
			}`,
			expected: "header-version-123",
		},
		{
			name: "version in response headers",
			sample: `{
				"object": "test.txt",
				"resp": {
					"headers": {
						"x-amz-version-id": "resp-version-456"
					}
				}
			}`,
			expected: "resp-version-456",
		},
		{
			name: "no version",
			sample: `{
				"object": "test.txt"
			}`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := map[string]interface{}{}
			if err := json.Unmarshal([]byte(tt.sample), &entry); err != nil {
				t.Fatalf("failed to unmarshal sample: %v", err)
			}

			got := extractVersionID(entry)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestRecordEventWithVersions(t *testing.T) {
	statsMap := map[string]*ObjectStat{}
	errorGroups := map[string]*errorAccumulator{}

	// Test grouping by versions enabled
	recordEvent(statsMap, errorGroups, nil, nil, "test.txt", "v123", "error1", "GetObject", "client1", true)
	recordEvent(statsMap, errorGroups, nil, nil, "test.txt", "v456", "error1", "GetObject", "client1", true)
	recordEvent(statsMap, errorGroups, nil, nil, "test.txt", "v123", "error1", "GetObject", "client1", true)

	// Should have 2 entries: one for each version
	if len(statsMap) != 2 {
		t.Errorf("expected 2 entries in statsMap with versions, got %d", len(statsMap))
	}

	expectedKey1 := "test.txt (version: v123)"
	expectedKey2 := "test.txt (version: v456)"

	if stat, ok := statsMap[expectedKey1]; !ok {
		t.Errorf("expected key %q not found", expectedKey1)
	} else if stat.Count != 2 {
		t.Errorf("expected count 2 for %q, got %d", expectedKey1, stat.Count)
	}

	if stat, ok := statsMap[expectedKey2]; !ok {
		t.Errorf("expected key %q not found", expectedKey2)
	} else if stat.Count != 1 {
		t.Errorf("expected count 1 for %q, got %d", expectedKey2, stat.Count)
	}

	// Test grouping by versions disabled
	statsMap2 := map[string]*ObjectStat{}
	errorGroups2 := map[string]*errorAccumulator{}

	recordEvent(statsMap2, errorGroups2, nil, nil, "test.txt", "v123", "error1", "GetObject", "client1", false)
	recordEvent(statsMap2, errorGroups2, nil, nil, "test.txt", "v456", "error1", "GetObject", "client1", false)
	recordEvent(statsMap2, errorGroups2, nil, nil, "test.txt", "v123", "error1", "GetObject", "client1", false)

	// Should have 1 entry: all versions grouped together
	if len(statsMap2) != 1 {
		t.Errorf("expected 1 entry in statsMap without versions, got %d", len(statsMap2))
	}

	if stat, ok := statsMap2["test.txt"]; !ok {
		t.Errorf("expected key 'test.txt' not found")
	} else if stat.Count != 3 {
		t.Errorf("expected count 3 for 'test.txt', got %d", stat.Count)
	}
}
