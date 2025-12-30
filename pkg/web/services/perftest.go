package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/liamdn8/mc-tool/pkg/config"
	"github.com/liamdn8/mc-tool/pkg/perftest"
)

// PerftestService manages performance testing operations
type PerftestService struct {
	mu         sync.RWMutex
	activeTest *perftest.Runner
	lastResult *perftest.TestResult
	running    bool
}

// NewPerftestService creates a new perftest service
func NewPerftestService() *PerftestService {
	return &PerftestService{
		running: false,
	}
}

// StartTest starts a new performance test
func (s *PerftestService) StartTest(cfg *perftest.TestConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("a test is already running")
	}

	// Load MC config
	mcConfig, err := config.LoadMCConfig()
	if err != nil {
		return fmt.Errorf("failed to load MC config: %w", err)
	}

	// Create runner
	runner, err := perftest.NewRunner(cfg, mcConfig)
	if err != nil {
		return fmt.Errorf("failed to create runner: %w", err)
	}

	s.activeTest = runner
	s.running = true

	// Run test in background
	go func() {
		ctx := context.Background()
		result, err := runner.Run(ctx)

		s.mu.Lock()
		defer s.mu.Unlock()

		s.running = false
		if err == nil {
			s.lastResult = result
		}
		s.activeTest = nil
	}()

	return nil
}

// GetStatus returns current test status
func (s *PerftestService) GetStatus() (bool, *perftest.TestStatus) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running || s.activeTest == nil {
		return false, nil
	}

	status := s.activeTest.GetStatus()
	return true, &status
}

// GetLastResult returns the last test result
func (s *PerftestService) GetLastResult() *perftest.TestResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.lastResult
}

// IsRunning checks if a test is currently running
func (s *PerftestService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.running
}
