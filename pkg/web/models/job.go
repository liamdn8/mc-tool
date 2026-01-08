package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const jobsDir = "/tmp/mc-tool-jobs"

// Job represents a background operation
type Job struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Status    string                 `json:"status"` // pending, running, completed, failed
	Progress  int                    `json:"progress"`
	Message   string                 `json:"message"`
	Result    map[string]interface{} `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	StartTime time.Time              `json:"start_time"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Output    []string               `json:"output"`
	mu        sync.Mutex
}

// JobManager manages background jobs
type JobManager struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

// NewJobManager creates a new job manager
func NewJobManager() *JobManager {
	// Ensure jobs directory exists
	os.MkdirAll(jobsDir, 0755)

	return &JobManager{
		jobs: make(map[string]*Job),
	}
}

// CreateJob creates a new job
func (jm *JobManager) CreateJob(jobType string) *Job {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	job := &Job{
		ID:        generateJobID(jobType),
		Type:      jobType,
		Status:    "pending",
		Progress:  0,
		StartTime: time.Now(),
		Output:    []string{},
	}

	jm.jobs[job.ID] = job
	job.persist() // Save to disk
	return job
}

// GetJob retrieves a job by ID
func (jm *JobManager) GetJob(id string) *Job {
	jm.mu.RLock()
	job := jm.jobs[id]
	jm.mu.RUnlock()

	// If not in memory, try loading from disk
	if job == nil {
		job = loadJobFromDisk(id)
		if job != nil {
			jm.mu.Lock()
			jm.jobs[id] = job
			jm.mu.Unlock()
		}
	}

	return job
}

// GetAllJobs returns all jobs
func (jm *JobManager) GetAllJobs() map[string]*Job {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	// Return a copy to avoid concurrent access issues
	result := make(map[string]*Job)
	for id, job := range jm.jobs {
		result[id] = job
	}
	return result
}

// GetJobHistory returns all jobs from disk (including completed ones)
func (jm *JobManager) GetJobHistory(jobType string, limit int) []*Job {
	files, err := os.ReadDir(jobsDir)
	if err != nil {
		return []*Job{}
	}

	jobs := []*Job{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		if filepath.Ext(filename) != ".json" {
			continue
		}

		// Filter by job type if specified
		jobID := filename[:len(filename)-5] // Remove .json extension
		if jobType != "" && len(jobID) > len(jobType) {
			if jobID[:len(jobType)] != jobType {
				continue
			}
		}

		if job := loadJobFromDisk(jobID); job != nil {
			jobs = append(jobs, job)
		}

		if limit > 0 && len(jobs) >= limit {
			break
		}
	}

	return jobs
}

// Job methods
func (j *Job) UpdateStatus(status, message string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = status
	j.Message = message
}

func (j *Job) UpdateProgress(progress int) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Progress = progress
}

func (j *Job) AddOutput(output string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Output = append(j.Output, output)
}

func (j *Job) Complete(result map[string]interface{}, message string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = "completed"
	j.Progress = 100
	j.Message = message
	j.Result = result
	now := time.Now()
	j.EndTime = &now
	j.persistUnlocked() // Save to disk without locking
}

func (j *Job) Fail(error string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = "failed"
	j.Error = error
	now := time.Now()
	j.EndTime = &now
	j.persistUnlocked() // Save to disk without locking
}

// Helper function to generate job ID
func generateJobID(jobType string) string {
	return fmt.Sprintf("%s-%d", jobType, time.Now().Unix())
}

// persist saves job to disk (with locking)
func (j *Job) persist() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.persistUnlocked()
}

// persistUnlocked saves job to disk without locking (caller must hold lock)
func (j *Job) persistUnlocked() {
	filename := filepath.Join(jobsDir, j.ID+".json")
	data, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(filename, data, 0644)
}

// loadJobFromDisk loads a job from disk
func loadJobFromDisk(jobID string) *Job {
	filename := filepath.Join(jobsDir, jobID+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil
	}

	var job Job
	if err := json.Unmarshal(data, &job); err != nil {
		return nil
	}

	return &job
}
