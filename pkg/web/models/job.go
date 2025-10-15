package models

import (
	"fmt"
	"sync"
	"time"
)

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
	return job
}

// GetJob retrieves a job by ID
func (jm *JobManager) GetJob(id string) *Job {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	return jm.jobs[id]
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
}

func (j *Job) Fail(error string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = "failed"
	j.Error = error
	now := time.Now()
	j.EndTime = &now
}

// Helper function to generate job ID
func generateJobID(jobType string) string {
	return fmt.Sprintf("%s-%d", jobType, time.Now().Unix())
}
