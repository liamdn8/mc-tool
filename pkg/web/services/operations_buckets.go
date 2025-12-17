package services

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/liamdn8/mc-tool/pkg/logger"
)

// bucketLister lists buckets for a given alias.
type bucketLister interface {
	ListBuckets(alias string) ([]string, error)
}

// bucketStatsProvider fetches statistics for a specific bucket.
type bucketStatsProvider interface {
	GetBucketStats(alias, bucket string) map[string]interface{}
}

// pathInspector inspects bucket content to provide path suggestions.
type pathInspector interface {
	inspectPaths(alias, bucket string) ([]string, error)
}

// bucketVersionChecker checks if versioning is enabled for a bucket.
type bucketVersionChecker interface {
	checkVersioning(alias, bucket string) (bool, error)
}

// bucketService orchestrates bucket-related operations.
type bucketService struct {
	lister         bucketLister
	statsProvider  bucketStatsProvider
	pathInspector  pathInspector
	versionChecker bucketVersionChecker
}

func newBucketService(l bucketLister, sp bucketStatsProvider, pi pathInspector, vc bucketVersionChecker) *bucketService {
	return &bucketService{
		lister:         l,
		statsProvider:  sp,
		pathInspector:  pi,
		versionChecker: vc,
	}
}

// ListBuckets returns buckets for the given alias.
func (s *bucketService) ListBuckets(alias string) ([]string, error) {
	return s.lister.ListBuckets(alias)
}

// BucketStats returns aggregated statistics for the alias buckets.
func (s *bucketService) BucketStats(alias string) ([]map[string]interface{}, error) {
	buckets, err := s.ListBuckets(alias)
	if err != nil {
		return nil, err
	}

	stats := make([]map[string]interface{}, 0, len(buckets))
	for _, bucket := range buckets {
		data := s.statsProvider.GetBucketStats(alias, bucket)
		stats = append(stats, data)
	}
	return stats, nil
}

// SuggestPaths returns common prefixes within the bucket.
func (s *bucketService) SuggestPaths(alias, bucket string) ([]string, error) {
	return s.pathInspector.inspectPaths(alias, bucket)
}

// CheckVersioning reports whether versioning is enabled for alias/bucket.
func (s *bucketService) CheckVersioning(alias, bucket string) (bool, error) {
	return s.versionChecker.checkVersioning(alias, bucket)
}

// mcBucketInspector implements pathInspector using mc find.
type mcBucketInspector struct{}

func (mcBucketInspector) inspectPaths(alias, bucket string) ([]string, error) {
	target := fmt.Sprintf("%s/%s", alias, bucket)
	cmd := exec.Command("mc", "find", target, "--name", "*", "--type", "d", "--max-depth", "2", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Return empty array instead of error - bucket might be empty or permissions issue
		logger.GetLogger().Warn("Failed to inspect paths", map[string]interface{}{
			"alias":  alias,
			"bucket": bucket,
			"error":  err.Error(),
			"output": string(output),
		})
		return []string{}, nil
	}

	var paths []string
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		var data map[string]interface{}
		if json.Unmarshal([]byte(line), &data) != nil {
			continue
		}
		if key, ok := data["key"].(string); ok {
			paths = append(paths, key)
		}
	}
	return paths, nil
}

// mcVersionChecker implements bucketVersionChecker using mc version info.
type mcVersionChecker struct{}

func (mcVersionChecker) checkVersioning(alias, bucket string) (bool, error) {
	cmd := exec.Command("mc", "version", "info", fmt.Sprintf("%s/%s", alias, bucket), "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}

	var data map[string]interface{}
	if json.Unmarshal(output, &data) != nil {
		return false, nil
	}

	if status, ok := data["status"].(string); ok && status == "success" {
		if enabled, ok := data["versioning"].(map[string]interface{}); ok {
			if state, ok := enabled["status"].(string); ok {
				return strings.EqualFold(state, "enabled"), nil
			}
		}
	}

	return false, nil
}
