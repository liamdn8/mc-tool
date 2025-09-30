# mc-tool MinIO Playground Test Report

**Date:** Tue Sep 30 11:38:17 +07 2025
**Test Environment:** MinIO Playground (https://play.min.io:9000)

## Test Summary

### Buckets Created
- playground/mc-tool-delete-test1 (with delete markers)
- playground/mc-tool-delete-test2 (with delete markers)
- playground/mc-tool-versions-test (versions test)

### Test Scenarios Covered
1. ✓ Basic object comparison (current versions)
2. ✓ Verbose output testing
3. ✓ Version comparison mode
4. ✓ Delete marker handling
5. ✓ Self-comparison (identical buckets)
6. ✓ Insecure connection testing
7. ✓ Empty vs non-empty bucket comparison
8. ✓ Performance testing with multiple objects

### Delete Marker Testing
- Created delete markers for objects in versioned buckets
- Verified mc-tool handles delete markers correctly
- Confirmed version mode shows all versions including delete markers

### Key Findings
- mc-tool correctly handles delete markers as "missing" objects in current version mode
- Version comparison mode properly shows all object versions
- Tool performs well with multiple objects
- TLS/insecure options work as expected

### Binary Information
- Binary: ./build/mc-tool
- Type:  ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, BuildID[sha1]=aa203053e1827522f8acd80853b5c5ae1d07aa36, with debug_info, not stripped
- Size: 14M

## Conclusion
mc-tool successfully handles all tested scenarios including complex delete marker situations.
