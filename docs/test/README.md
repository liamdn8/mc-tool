# Test Documentation

## Overview

This directory contains comprehensive test documentation and test suites for the MinIO Site Replication Management feature.

## Structure

```
docs/test/
├── TESTCASE_DOCUMENTATION.md     # Complete testcase documentation
├── tests/                        # Test implementation
│   ├── integration/              # Integration test suites
│   │   ├── smart_removal_test.go
│   │   ├── replication_apis_test.go
│   │   ├── ui_integration_test.go
│   │   ├── error_handling_test.go
│   │   └── test_utils.go
│   ├── fixtures/                 # Test data and mock responses
│   │   ├── minio_responses.json
│   │   └── test_configs.json
│   ├── run_integration_tests.sh  # Test runner script
│   └── README.md                 # Test execution instructions
└── README.md                     # This file
```

## Quick Start

### Running Tests

```bash
# Navigate to test directory
cd /home/liamdn/mc-tool/docs/test/tests

# Run all tests
./run_integration_tests.sh

# Run specific test suites
./run_integration_tests.sh --smart-removal
./run_integration_tests.sh --api-only
./run_integration_tests.sh --ui-only
```

### Test Coverage

- ✅ **Smart Site Removal Algorithm**: 6 test functions, 15+ scenarios
- ✅ **Replication Management APIs**: 6 endpoints fully tested
- ✅ **UI/UX Integration**: Lucid Icons, responsive design, accessibility
- ✅ **Error Handling**: Connection failures, permissions, validation
- ✅ **Test Infrastructure**: Environment setup, mocking, cleanup

### Requirements Traceability

All tests are mapped to requirements in `/docs/requirements/`:

1. **Smart Site Removal Logic** → `smart_removal_test.go`
2. **Lucid Icons Integration** → `ui_integration_test.go`  
3. **API Endpoints** → `replication_apis_test.go`
4. **Error Handling** → `error_handling_test.go`

## Documentation

See [TESTCASE_DOCUMENTATION.md](./TESTCASE_DOCUMENTATION.md) for detailed test scenarios, coverage analysis, and execution instructions.

## CI/CD Integration

The test suite is designed for continuous integration with automated reporting and performance benchmarking.

For detailed implementation information, refer to the main tests directory and documentation files.