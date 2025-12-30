#!/bin/bash

# Example: Using mc-tool with environment variables
# This demonstrates using MC_HOST_* environment variables for configuration

echo "╔════════════════════════════════════════════════════════════════╗"
echo "║  mc-tool Environment Variable Configuration Example            ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo ""

# Example 1: Configure aliases using environment variables
echo "1. Configure MinIO aliases using environment variables:"
echo ""
echo "export MC_HOST_site1=https://accesskey:secretkey@minio1.example.com:9000"
echo "export MC_HOST_site2=http://minioadmin:minioadmin@localhost:9000"
echo "export MC_HOST_staging=https://key:secret@staging.minio.local:9000?insecure=true"
echo ""

# Example 2: These aliases work with all mc-tool commands
echo "2. Use aliases with mc-tool commands:"
echo ""
echo "# Compare buckets"
echo "mc-tool compare site1/bucket site2/bucket"
echo ""
echo "# Validate configuration"
echo "mc-tool validate mybucket site1 site2 staging"
echo ""
echo "# Profile with mc21"
echo "mc-tool profile site1 --mc-path mc21 --duration 30s"
echo ""

# Example 3: Environment variables work with mc21
echo "3. mc21 compatibility:"
echo ""
echo "When mc-tool runs mc21 commands, it automatically passes the"
echo "environment variables so mc21 can access the same alias configurations."
echo ""
echo "This means you can:"
echo "  - Use --config-dir ~/.mc to read existing mc config files"
echo "  - Use MC_HOST_* environment variables to add or override aliases"
echo "  - mc21 will see both sources of configuration"
echo ""

# Example 4: Docker/Kubernetes usage
echo "4. Perfect for containers:"
echo ""
cat << 'YAML'
apiVersion: v1
kind: Pod
metadata:
  name: mc-tool
spec:
  containers:
  - name: mc-tool
    image: mc-tool:latest
    env:
    - name: MC_HOST_site1
      valueFrom:
        secretKeyRef:
          name: minio-credentials
          key: site1-url
    - name: MC_HOST_site2
      valueFrom:
        secretKeyRef:
          name: minio-credentials
          key: site2-url
YAML
echo ""

# Example 5: CI/CD Pipeline
echo "5. CI/CD Pipeline example:"
echo ""
cat << 'SCRIPT'
#!/bin/bash
# GitLab CI / GitHub Actions

export MC_HOST_prod=${PROD_MINIO_URL}
export MC_HOST_staging=${STAGING_MINIO_URL}

# Run validation
mc-tool validate mybucket staging prod

# Run comparison
mc-tool compare staging/backup prod/backup --verbose
SCRIPT
echo ""

echo "╔════════════════════════════════════════════════════════════════╗"
echo "║  For more information, see docs/CONFIG.md                      ║"
echo "╚════════════════════════════════════════════════════════════════╝"
