# Testing Documentation for mc-tool

This document describes the comprehensive test suite for the mc-tool MinIO bucket comparison utility.

## Test Coverage

The test suite covers **41.0%** of the codebase with comprehensive unit tests, integration tests, and benchmarks.

## Test Structure

### 1. Unit Tests (`main_test.go`)

**Core Functionality Tests:**
- `TestParseURL` - URL parsing for MinIO aliases and paths
- `TestLoadMCConfig` - MC configuration file loading
- `TestCompareCurrentVersions` - Object comparison logic
- `TestCompareVersions` - Multi-version object comparison
- `TestInsecureFlagLogic` - TLS certificate verification logic

**Test Cases Covered:**
```
✓ Simple bucket URLs (alias/bucket)
✓ Complex path URLs (alias/bucket/path/to/file)
✓ Invalid URL formats and error handling
✓ Identical objects (same ETag and size)
✓ Different objects (different ETag, size, or both)
✓ Missing objects (source or target)
✓ Insecure flag priority (CLI > config > default)
✓ Configuration file parsing and validation
```

### 2. Integration Tests (`integration_test.go`)

**Configuration Workflow Tests:**
- `TestCreateMinIOClientConfiguration` - MinIO client creation with various configurations
- `TestFullConfigurationWorkflow` - End-to-end configuration loading and validation
- `TestErrorHandlingIntegration` - Error scenarios and recovery

**Test Scenarios:**
```
✓ HTTP connections without TLS
✓ HTTPS connections with proper certificates
✓ HTTPS with self-signed certificates (insecure config)
✓ Command-line insecure flag override
✓ Non-existent alias error handling
✓ Invalid JSON configuration handling
✓ Missing configuration file handling
```

### 3. Output Tests (`output_test.go`)

**Display and Output Tests:**
- `TestDisplayResults` - Output formatting and content verification
- `TestDisplayResultsVerbose` - Verbose mode output validation
- `TestGlobalVariables` - Global state management
- `TestAliasConfigValidation` - URL parsing and SSL detection

**Output Validation:**
```
✓ Summary statistics (identical, different, missing counts)
✓ Verbose mode detailed object information
✓ Non-verbose mode filtering
✓ Error messages and formatting
✓ Exit code behavior (mocked for testing)
```

### 4. Benchmark Tests

**Performance Testing:**
- `BenchmarkParseURL` - URL parsing performance
- `BenchmarkCompareCurrentVersions` - Single object comparison performance
- `BenchmarkCompareVersions` - Multi-version comparison performance

**Benchmark Results:**
```
BenchmarkParseURL-4                     15,668,044 ops    78.19 ns/op    48 B/op    1 allocs/op
BenchmarkCompareCurrentVersions-4       11,095,629 ops   101.7 ns/op     48 B/op    2 allocs/op
BenchmarkCompareVersions-4                  16,178 ops 69,698 ns/op  47,874 B/op  335 allocs/op
```

## Test Configuration

### Test Helper Functions

The test suite includes several helper functions:

```go
// TestHelper - Provides utilities for testing with temporary directories
type TestHelper struct {
    TempDir    string
    ConfigPath string
}

// GetTestConfig() - Returns standard test configuration
// captureOutput() - Captures stdout for output testing
// displayResultsForTest() - Non-exiting version of displayResults for testing
```

### Sample Test Configurations

**Insecure Configuration Example:**
```json
{
  "version": "10",
  "aliases": {
    "staging": {
      "url": "https://minio-staging.example.com",
      "accessKey": "staging-access-key",
      "secretKey": "staging-secret-key",
      "api": "s3v4",
      "path": "auto",
      "insecure": true
    }
  }
}
```

## Running Tests

### Basic Test Execution
```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run with race detection
make test-race

# Run benchmark tests
make test-bench

# Run specific test
make test-single TEST=TestParseURL
```

### Coverage Report

The coverage report is generated as `coverage.html` and includes:
- Line-by-line coverage visualization
- Function coverage statistics
- Uncovered code highlighting

## Test Data and Fixtures

### Mock Configuration Files
- Sample MC configuration with multiple aliases
- Invalid JSON configurations for error testing
- Missing configuration scenarios

### Test Object Data
```go
ObjectInfo{
    Key:          "test/file.txt",
    ETag:         "abc123",
    Size:         1024,
    LastModified: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
    VersionID:    "v1",
    IsLatest:     true,
}
```

## Edge Cases Tested

1. **URL Parsing Edge Cases:**
   - Empty URLs
   - URLs with only alias
   - URLs with special characters
   - Deep path hierarchies

2. **Object Comparison Edge Cases:**
   - Nil objects
   - Empty object lists
   - Objects with identical ETags but different sizes
   - Missing version information

3. **Configuration Edge Cases:**
   - Missing home directory
   - Corrupted JSON files
   - Non-existent aliases
   - Mixed HTTP/HTTPS configurations

## Continuous Integration

The test suite is designed to run in CI environments with:
- No external dependencies (uses temporary directories)
- Deterministic results
- Fast execution (< 6 seconds including benchmarks)
- Clear pass/fail indicators

## Test Maintenance

### Adding New Tests

When adding new functionality:
1. Add unit tests to the appropriate test file
2. Include edge cases and error scenarios
3. Add integration tests for end-to-end workflows
4. Update benchmark tests for performance-critical code
5. Verify coverage doesn't decrease significantly

### Test Debugging

Use verbose test output for debugging:
```bash
go test ./... -v -run TestSpecificFunction
```

### Performance Regression Detection

Monitor benchmark results for performance regressions:
```bash
# Run benchmarks multiple times and compare
go test -bench=. -count=5 -benchmem
```

## Future Test Enhancements

1. **Integration with Real MinIO Instances**
   - Docker-based test environment
   - Live bucket comparison testing

2. **Property-Based Testing**
   - Generate random object configurations
   - Verify comparison invariants

3. **Parallel Test Execution**
   - Concurrent comparison testing
   - Race condition detection

4. **Extended Error Scenarios**
   - Network failure simulation
   - Timeout handling testing

## Conclusion

The test suite provides comprehensive coverage of the mc-tool functionality, ensuring reliability and maintainability. The combination of unit tests, integration tests, and benchmarks provides confidence in the tool's correctness and performance across various MinIO deployment scenarios.