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
