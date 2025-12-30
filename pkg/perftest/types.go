package perftest

import (
	"time"
)

// ObjectSizeType represents predefined object size categories
type ObjectSizeType string

const (
	ObjectSizeSmall  ObjectSizeType = "small"  // 1-10 KiB
	ObjectSizeMedium ObjectSizeType = "medium" // 100-500 KiB
	ObjectSizeLarge  ObjectSizeType = "large"  // 1-5 MiB
)

// UploadMode represents the upload pattern
type UploadMode string

const (
	UploadModeAll      UploadMode = "all"      // Upload all files at once
	UploadModeInterval UploadMode = "interval" // Upload files at intervals
)

// TestConfig contains configuration for performance test
type TestConfig struct {
	SiteAlias      string         // Site alias to upload to
	Bucket         string         // Bucket name for testing
	ObjectPath     string         // Object path prefix
	ObjectSizeType ObjectSizeType // Predefined size category
	ObjectSize     int64          // Actual size in bytes (auto-calculated from type)
	ObjectCount    int            // Number of objects per upload round
	OverrideCount  int            // Number of times to override each object (0 = no override, 1+ = override N times)
	UploadMode     UploadMode     // Upload pattern: all or interval
	UploadInterval time.Duration  // Interval between upload rounds (only for interval mode)
	Iterations     int            // Number of upload rounds (only for interval mode)
	Parallelism    int            // Number of parallel workers (only for 'all' mode)
	Insecure       bool           // Skip TLS certificate verification
}

// TestResult contains results from a performance test
type TestResult struct {
	Config        TestConfig     `json:"config"`
	StartTime     time.Time      `json:"start_time"`
	EndTime       time.Time      `json:"end_time"`
	TotalDuration time.Duration  `json:"total_duration"`
	ObjectResults []ObjectResult `json:"object_results"`
	Summary       TestSummary    `json:"summary"`
	Errors        []string       `json:"errors,omitempty"`
}

// ObjectResult contains result for a single object PUT operation
type ObjectResult struct {
	ObjectKey    string        `json:"object_key"`
	ObjectSize   int64         `json:"object_size"`
	UploadNumber int           `json:"upload_number"` // Which upload attempt (1 = first, 2+ = overrides)
	RoundNumber  int           `json:"round_number"`  // Which upload round (for interval mode)
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Duration     time.Duration `json:"duration"`
	Success      bool          `json:"success"`
	Error        string        `json:"error,omitempty"`
}

// TestSummary contains aggregated statistics
type TestSummary struct {
	TotalUploads         int            `json:"total_uploads"`          // Total number of upload operations
	SuccessfulUploads    int            `json:"successful_uploads"`     // Successful uploads
	FailedUploads        int            `json:"failed_uploads"`         // Failed uploads
	UniqueObjects        int            `json:"unique_objects"`         // Number of unique object keys
	OverriddenObjects    int            `json:"overridden_objects"`     // Objects that were overridden
	TotalDataUploaded    int64          `json:"total_data_uploaded"`    // Total bytes uploaded
	AverageUploadLatency time.Duration  `json:"average_upload_latency"` // Average upload time
	Throughput           float64        `json:"throughput"`             // Uploads per second
	DataThroughput       float64        `json:"data_throughput"`        // Bytes per second
	OverrideDetails      map[string]int `json:"override_details"`       // Object key -> override count
}

// TestStatus represents the current status of an ongoing test
type TestStatus struct {
	Running           bool           `json:"running"`
	Progress          int            `json:"progress"` // Percentage 0-100
	TotalUploads      int            `json:"total_uploads"`
	CompletedUploads  int            `json:"completed_uploads"`
	SuccessfulUploads int            `json:"successful_uploads"`
	FailedUploads     int            `json:"failed_uploads"`
	ElapsedTime       time.Duration  `json:"elapsed_time"`
	CurrentPhase      string         `json:"current_phase"` // "uploading", "complete"
	CurrentRound      int            `json:"current_round,omitempty"`
	TotalRounds       int            `json:"total_rounds,omitempty"`
	RoundDetails      []RoundDetail  `json:"round_details,omitempty"`
	RecentUploads     []ObjectResult `json:"recent_uploads,omitempty"` // Last 10 uploads
}

// RoundDetail contains details for each upload round
type RoundDetail struct {
	Round             int           `json:"round"`
	ObjectsUploaded   int           `json:"objects_uploaded"`
	SuccessfulUploads int           `json:"successful_uploads"`
	FailedUploads     int           `json:"failed_uploads"`
	Duration          time.Duration `json:"duration"`
	StartTime         time.Time     `json:"start_time"`
	EndTime           time.Time     `json:"end_time"`
}
