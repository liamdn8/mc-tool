#!/bin/bash

echo "=== Test Environment Variable Configuration Loading ==="
echo ""

# Test with env var
echo "Test: Load alias from environment variable"
echo "Setting MC_HOST_envtest=https://testkey:testsecret@minio.test.local:9000"
export MC_HOST_envtest=https://testkey:testsecret@minio.test.local:9000

# Simple test using mc-tool
cat > /tmp/test_env.go << 'GOEOF'
package main
import (
    "fmt"
    "github.com/liamdn8/mc-tool/pkg/config"
)
func main() {
    cfg, err := config.LoadMCConfig()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("Loaded %d aliases\n", len(cfg.Aliases))
    if alias, exists := cfg.Aliases["envtest"]; exists {
        fmt.Printf("✅ envtest: %s (key: %s)\n", alias.URL, alias.AccessKey)
    }
}
GOEOF

go run /tmp/test_env.go
rm /tmp/test_env.go
echo "=== Test Complete ==="
