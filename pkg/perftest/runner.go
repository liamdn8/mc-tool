package perftest

import (
	"context"
	"crypto/rand"
	"fmt"
	mathrand "math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/liamdn8/mc-tool/pkg/config"
	"github.com/liamdn8/mc-tool/pkg/logger"
)

// Runner manages performance test execution
type Runner struct {
	config        *TestConfig
	mcConfig      *config.MCConfig
	tempDir       string
	log           *logger.Logger
	statusLock    sync.RWMutex
	status        TestStatus
	startTime     time.Time
	recentUploads []ObjectResult
	uploadsMutex  sync.Mutex
}

// NewRunner creates a new performance test runner
func NewRunner(cfg *TestConfig, mcConfig *config.MCConfig) (*Runner, error) {
	// Validate site alias exists
	if _, exists := mcConfig.Aliases[cfg.SiteAlias]; !exists {
		return nil, fmt.Errorf("site alias '%s' not found in MC config", cfg.SiteAlias)
	}

	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "perftest-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Calculate total uploads and rounds
	totalUploads := cfg.ObjectCount * (cfg.OverrideCount + 1)
	totalRounds := 1
	if cfg.UploadMode == "interval" {
		totalRounds = cfg.Iterations
		if totalRounds <= 0 {
			totalRounds = 1
		}
		// In interval mode: each round uploads ObjectCount unique files + OverrideCount files repeatedly
		totalUploads = (cfg.ObjectCount + cfg.OverrideCount) * totalRounds
	}

	return &Runner{
		config:   cfg,
		mcConfig: mcConfig,
		tempDir:  tempDir,
		log:      logger.GetLogger(),
		status: TestStatus{
			Running:       false,
			Progress:      0,
			TotalUploads:  totalUploads,
			TotalRounds:   totalRounds,
			RoundDetails:  make([]RoundDetail, 0),
			RecentUploads: make([]ObjectResult, 0, 10),
		},
		recentUploads: make([]ObjectResult, 0, 10),
	}, nil
}

// Run executes the performance test
func (r *Runner) Run(ctx context.Context) (*TestResult, error) {
	totalUploads := r.config.ObjectCount * (r.config.OverrideCount + 1)
	r.startTime = time.Now()
	r.updateStatus(true, 0, "uploading", 0, 0, 0)
	defer r.updateStatus(false, 100, "complete", totalUploads, 0, 0)

	result := &TestResult{
		Config:        *r.config,
		StartTime:     time.Now(),
		ObjectResults: make([]ObjectResult, 0, totalUploads),
		Errors:        make([]string, 0),
	}

	r.log.Info("Starting performance test", map[string]interface{}{
		"site":           r.config.SiteAlias,
		"bucket":         r.config.Bucket,
		"objects":        r.config.ObjectCount,
		"override_count": r.config.OverrideCount,
		"upload_mode":    r.config.UploadMode,
		"size_type":      r.config.ObjectSizeType,
	})

	// Step 1: Calculate actual object size from preset
	if err := r.calculateObjectSize(); err != nil {
		return nil, fmt.Errorf("failed to calculate object size: %w", err)
	}

	// Step 3: Generate object keys
	objectKeys := r.generateObjectKeys()

	// Step 4: Create test data file
	testDataFile, err := r.createTestDataFile()
	if err != nil {
		return nil, fmt.Errorf("failed to create test data: %w", err)
	}
	defer os.Remove(testDataFile)

	// Step 4: Execute uploads based on mode
	var objectResults []ObjectResult
	var uploadErr error

	switch r.config.UploadMode {
	case UploadModeAll:
		objectResults, uploadErr = r.executeUploadAll(ctx, objectKeys, testDataFile)
	case UploadModeInterval:
		objectResults, uploadErr = r.executeUploadInterval(ctx, objectKeys, testDataFile)
	default:
		return nil, fmt.Errorf("unknown upload mode: %s", r.config.UploadMode)
	}

	if uploadErr != nil {
		result.Errors = append(result.Errors, uploadErr.Error())
	}

	result.ObjectResults = objectResults

	// Step 5: Calculate summary
	result.EndTime = time.Now()
	result.TotalDuration = result.EndTime.Sub(result.StartTime)
	result.Summary = r.calculateSummary(objectResults, result.TotalDuration)

	return result, nil
}

// GetStatus returns current test status
func (r *Runner) GetStatus() TestStatus {
	r.statusLock.RLock()
	defer r.statusLock.RUnlock()

	status := r.status
	if r.startTime.IsZero() {
		status.ElapsedTime = 0
	} else {
		status.ElapsedTime = time.Since(r.startTime)
	}

	// Copy recent uploads
	r.uploadsMutex.Lock()
	status.RecentUploads = make([]ObjectResult, len(r.recentUploads))
	copy(status.RecentUploads, r.recentUploads)
	r.uploadsMutex.Unlock()

	return status
}

// Cleanup removes temporary files
func (r *Runner) Cleanup() error {
	if r.tempDir != "" {
		return os.RemoveAll(r.tempDir)
	}
	return nil
}

// updateStatus updates the current test status
func (r *Runner) updateStatus(running bool, progress int, phase string, completed, successful, failed int) {
	r.statusLock.Lock()
	defer r.statusLock.Unlock()

	r.status.Running = running
	r.status.Progress = progress
	r.status.CurrentPhase = phase
	r.status.CompletedUploads = completed
	r.status.SuccessfulUploads = successful
	r.status.FailedUploads = failed
}

// calculateObjectSize calculates actual size from preset
func (r *Runner) calculateObjectSize() error {
	var minSize, maxSize int64

	switch r.config.ObjectSizeType {
	case ObjectSizeSmall:
		minSize = 1 * 1024  // 1 KiB
		maxSize = 10 * 1024 // 10 KiB
	case ObjectSizeMedium:
		minSize = 100 * 1024 // 100 KiB
		maxSize = 500 * 1024 // 500 KiB
	case ObjectSizeLarge:
		minSize = 1 * 1024 * 1024 // 1 MiB
		maxSize = 5 * 1024 * 1024 // 5 MiB
	default:
		return fmt.Errorf("unknown size type: %s", r.config.ObjectSizeType)
	}

	// Random size within range
	if maxSize > minSize {
		r.config.ObjectSize = minSize + mathrand.Int63n(maxSize-minSize)
	} else {
		r.config.ObjectSize = minSize
	}

	r.log.Info("Object size calculated", map[string]interface{}{
		"type":  r.config.ObjectSizeType,
		"bytes": r.config.ObjectSize,
		"human": formatBytes(r.config.ObjectSize),
	})

	return nil
}

// generateObjectKeys generates unique object keys
func (r *Runner) generateObjectKeys() []string {
	keys := make([]string, r.config.ObjectCount)
	timestamp := time.Now().Format("20060102-150405")

	for i := 0; i < r.config.ObjectCount; i++ {
		keys[i] = fmt.Sprintf("%sobj-%s-%04d.bin", r.config.ObjectPath, timestamp, i+1)
	}

	return keys
}

// createTestDataFile creates a temporary file with random data
func (r *Runner) createTestDataFile() (string, error) {
	filePath := filepath.Join(r.tempDir, "testdata.bin")

	// Generate random data
	data := make([]byte, r.config.ObjectSize)
	if _, err := rand.Read(data); err != nil {
		return "", fmt.Errorf("failed to generate random data: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write test data: %w", err)
	}

	r.log.Info("Test data file created", map[string]interface{}{
		"path": filePath,
		"size": r.config.ObjectSize,
	})

	return filePath, nil
}

// executeUploadAll uploads all files at once with parallelism
func (r *Runner) executeUploadAll(ctx context.Context, objectKeys []string, testDataFile string) ([]ObjectResult, error) {
	results := make([]ObjectResult, 0)
	resultsMutex := sync.Mutex{}
	successCount := 0
	failCount := 0

	// Create work queue
	type uploadTask struct {
		objectKey    string
		uploadNumber int
	}

	tasks := make([]uploadTask, 0)

	// Generate all tasks (original + overrides)
	for _, key := range objectKeys {
		for overrideNum := 0; overrideNum <= r.config.OverrideCount; overrideNum++ {
			tasks = append(tasks, uploadTask{
				objectKey:    key,
				uploadNumber: overrideNum + 1,
			})
		}
	}

	r.log.Info("Executing upload-all mode", map[string]interface{}{
		"total_uploads": len(tasks),
		"parallelism":   r.config.Parallelism,
	})

	// Execute with parallelism
	taskChan := make(chan uploadTask, len(tasks))
	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	var wg sync.WaitGroup
	for i := 0; i < r.config.Parallelism; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				result := r.uploadObject(ctx, task.objectKey, testDataFile, task.uploadNumber, 0)

				resultsMutex.Lock()
				results = append(results, result)
				if result.Success {
					successCount++
				} else {
					failCount++
				}
				completed := len(results)
				progress := (completed * 100) / len(tasks)
				r.updateStatus(true, progress, "uploading", completed, successCount, failCount)
				resultsMutex.Unlock()
			}
		}()
	}

	wg.Wait()

	return results, nil
}

// executeUploadInterval uploads files in rounds at specified intervals
func (r *Runner) executeUploadInterval(ctx context.Context, objectKeys []string, testDataFile string) ([]ObjectResult, error) {
	results := make([]ObjectResult, 0)
	totalRounds := r.config.Iterations
	if totalRounds <= 0 {
		totalRounds = 1
	}

	// In interval mode: each round uploads ObjectCount unique files + OverrideCount files repeatedly
	totalUploads := (r.config.ObjectCount + r.config.OverrideCount) * totalRounds

	r.log.Info("Executing upload-interval mode", map[string]interface{}{
		"rounds":             totalRounds,
		"unique_per_round":   r.config.ObjectCount,
		"override_per_round": r.config.OverrideCount,
		"total_per_round":    r.config.ObjectCount + r.config.OverrideCount,
		"total_uploads":      totalUploads,
		"interval":           r.config.UploadInterval,
	})

	completed := 0

	// Generate override keys once (these will be uploaded in every round)
	timestamp := time.Now().Format("20060102-150405")
	overrideKeys := make([]string, r.config.OverrideCount)
	for i := 0; i < r.config.OverrideCount; i++ {
		overrideKeys[i] = fmt.Sprintf("%soverride-obj-%s-%04d.bin", r.config.ObjectPath, timestamp, i+1)
	}

	// Upload in rounds
	for round := 1; round <= totalRounds; round++ {
		roundStartTime := time.Now()
		roundSuccessful := 0
		roundFailed := 0

		r.log.Info(fmt.Sprintf("========== Starting Round %d/%d ==========", round, totalRounds), nil)

		// Update status with current round
		r.statusLock.Lock()
		r.status.CurrentRound = round
		r.statusLock.Unlock()

		// Generate unique object keys for this round
		roundTimestamp := time.Now().Format("20060102-150405")
		roundKeys := make([]string, r.config.ObjectCount)
		for i := 0; i < r.config.ObjectCount; i++ {
			roundKeys[i] = fmt.Sprintf("%sround-%d-obj-%s-%04d.bin", r.config.ObjectPath, round, roundTimestamp, i+1)
		}

		// Upload unique objects for this round
		for _, key := range roundKeys {
			// Check context cancellation
			select {
			case <-ctx.Done():
				return results, ctx.Err()
			default:
			}

			result := r.uploadObject(ctx, key, testDataFile, 1, round)
			results = append(results, result)

			if result.Success {
				roundSuccessful++
			} else {
				roundFailed++
			}

			completed++
			progress := (completed * 100) / totalUploads
			r.updateStatus(true, progress, "uploading", completed, len(results)-roundFailed, roundFailed)
		}

		// Upload override objects (same keys across all rounds)
		for i, key := range overrideKeys {
			// Check context cancellation
			select {
			case <-ctx.Done():
				return results, ctx.Err()
			default:
			}

			// Upload number = round (each round is a new version)
			result := r.uploadObject(ctx, key, testDataFile, round, round)
			results = append(results, result)

			if result.Success {
				roundSuccessful++
			} else {
				roundFailed++
			}

			completed++
			progress := (completed * 100) / totalUploads
			r.updateStatus(true, progress, "uploading", completed, len(results)-roundFailed, roundFailed)

			r.log.Info(fmt.Sprintf("[OVERRIDE %d/%d] Uploading version %d", i+1, r.config.OverrideCount, round), map[string]interface{}{
				"object": key,
				"round":  round,
			})
		}

		roundEndTime := time.Now()
		roundDuration := roundEndTime.Sub(roundStartTime)

		// Add round detail to status
		r.statusLock.Lock()
		roundDetail := RoundDetail{
			Round:             round,
			ObjectsUploaded:   r.config.ObjectCount + r.config.OverrideCount,
			SuccessfulUploads: roundSuccessful,
			FailedUploads:     roundFailed,
			Duration:          roundDuration,
			StartTime:         roundStartTime,
			EndTime:           roundEndTime,
		}
		r.status.RoundDetails = append(r.status.RoundDetails, roundDetail)
		r.statusLock.Unlock()

		r.log.Info(fmt.Sprintf("Round %d/%d completed", round, totalRounds), map[string]interface{}{
			"objects_uploaded": r.config.ObjectCount + r.config.OverrideCount,
			"unique":           r.config.ObjectCount,
			"override":         r.config.OverrideCount,
			"successful":       roundSuccessful,
			"failed":           roundFailed,
			"duration":         roundDuration,
		})

		// Wait interval before next round (except for last round)
		if round < totalRounds {
			r.log.Info(fmt.Sprintf("Waiting %v before next round...", r.config.UploadInterval), nil)
			time.Sleep(r.config.UploadInterval)
		}
	}

	return results, nil
}

// uploadObject uploads a single object
func (r *Runner) uploadObject(ctx context.Context, objectKey, sourceFile string, uploadNumber, roundNumber int) ObjectResult {
	startTime := time.Now()
	target := fmt.Sprintf("%s/%s/%s", r.config.SiteAlias, r.config.Bucket, objectKey)

	r.log.Info(fmt.Sprintf("[PUT] Uploading object"), map[string]interface{}{
		"object": objectKey,
		"upload": uploadNumber,
		"round":  roundNumber,
		"size":   formatBytes(r.config.ObjectSize),
		"target": target,
	})

	// Build mc command with insecure flag if needed
	args := []string{"cp"}
	if r.config.Insecure {
		args = append(args, "--insecure")
	}
	args = append(args, sourceFile, target)

	cmd := exec.CommandContext(ctx, "mc", args...)
	output, err := cmd.CombinedOutput()
	endTime := time.Now()

	result := ObjectResult{
		ObjectKey:    objectKey,
		ObjectSize:   r.config.ObjectSize,
		UploadNumber: uploadNumber,
		RoundNumber:  roundNumber,
		StartTime:    startTime,
		EndTime:      endTime,
		Duration:     endTime.Sub(startTime),
		Success:      err == nil,
	}

	if err != nil {
		result.Error = fmt.Sprintf("upload failed: %v, output: %s", err, string(output))
		r.log.Error("Upload failed", map[string]interface{}{
			"object": objectKey,
			"upload": uploadNumber,
			"round":  roundNumber,
			"error":  result.Error,
		})
	} else {
		r.log.Info("[PUT SUCCESS]", map[string]interface{}{
			"object":   objectKey,
			"upload":   uploadNumber,
			"round":    roundNumber,
			"duration": result.Duration,
		})
	}
	// Add to recent uploads (keep last 10)
	r.uploadsMutex.Lock()
	r.recentUploads = append(r.recentUploads, result)
	if len(r.recentUploads) > 10 {
		r.recentUploads = r.recentUploads[len(r.recentUploads)-10:]
	}
	r.uploadsMutex.Unlock()
	return result
}

// calculateSummary generates summary statistics
func (r *Runner) calculateSummary(results []ObjectResult, totalDuration time.Duration) TestSummary {
	summary := TestSummary{
		TotalUploads:    len(results),
		OverrideDetails: make(map[string]int),
	}

	var totalLatency time.Duration
	var totalDataUploaded int64
	uniqueObjects := make(map[string]bool)

	for _, result := range results {
		uniqueObjects[result.ObjectKey] = true

		if result.Success {
			summary.SuccessfulUploads++
			totalLatency += result.Duration
			totalDataUploaded += result.ObjectSize
		} else {
			summary.FailedUploads++
		}

		// Track overrides
		if result.UploadNumber > 1 {
			summary.OverrideDetails[result.ObjectKey]++
		}
	}

	summary.UniqueObjects = len(uniqueObjects)
	summary.OverriddenObjects = len(summary.OverrideDetails)
	summary.TotalDataUploaded = totalDataUploaded

	// Calculate averages
	if summary.SuccessfulUploads > 0 {
		summary.AverageUploadLatency = totalLatency / time.Duration(summary.SuccessfulUploads)
		summary.Throughput = float64(summary.SuccessfulUploads) / totalDuration.Seconds()
		summary.DataThroughput = float64(totalDataUploaded) / totalDuration.Seconds()
	}

	return summary
}

// formatBytes formats bytes to human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
