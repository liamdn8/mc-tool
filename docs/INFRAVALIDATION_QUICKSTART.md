# Infrastructure Validation Quick Start Guide

This guide will walk you through setting up and running your first infrastructure validation.

## Prerequisites

1. **Install KinD** (Kubernetes in Docker)
   ```bash
   # On Linux
   curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
   chmod +x ./kind
   sudo mv ./kind /usr/local/bin/kind
   
   # On macOS
   brew install kind
   ```

2. **Install kubectl**
   ```bash
   # On Linux
   curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
   chmod +x kubectl
   sudo mv kubectl /usr/local/bin/
   
   # On macOS
   brew install kubectl
   ```

3. **Build mc-tool**
   ```bash
   cd /path/to/mc-tool
   go build -o mc-tool .
   ```

## Step 1: Setup Test Environment

Create a test cluster with sample workloads in three namespaces:

```bash
./scripts/kind-setup-infravalidation.sh
```

This will:
- Create 1 KinD cluster (`kind-infra-test`)
- Create 3 namespaces: `app-prod`, `app-staging`, `app-dev`
- Deploy identical resources to `app-prod` and `app-staging`
- Deploy drifted resources to `app-dev`
- Create service accounts and extract tokens
- Generate `~/.mc-tool/infra-config.yaml` configuration

**Expected output:**
```
🚀 Setting up KinD cluster for infravalidation testing
📦 Creating KinD cluster: kind-infra-test
✅ Cluster ready. API Server: https://127.0.0.1:42979
📁 Creating namespaces...
🔑 Creating service accounts...
✅ Service accounts created and tokens extracted
📦 Deploying baseline resources to app-prod...
📦 Deploying matching resources to app-staging...
📦 Deploying drifted resources to app-dev...
✅ KinD test environment setup complete!
```

## Step 2: Verify Cluster

Verify that cluster and namespaces are ready:

```bash
# Check cluster
kubectl get nodes

# Check all namespaces
kubectl get ns | grep app-

# Check resources in each namespace
kubectl get all -n app-prod
kubectl get all -n app-staging
kubectl get all -n app-dev
```

## Step 3: Run Validation

Run infrastructure validation comparing namespaces:

```bash
# Compare prod with staging (should mostly match)
./mc-tool validate-infra site1/app-prod site2/app-staging

# Compare prod with dev (should show drifts)
./mc-tool validate-infra site1/app-prod site3/app-dev

# Compare all three at once
./mc-tool validate-infra site1/app-prod site2/app-staging site3/app-dev
```

**Expected output (prod vs dev):**
```
🔍 Starting infrastructure validation...
Baseline: site1/app-prod
Targets: 1 namespace(s)
  - site3/app-dev

================================================================================
Infrastructure Validation Summary
================================================================================

Baseline: site1/app-prod
Timestamp: 2026-01-08T17:29:48+07:00
Mode: mode-a

Overall Results:
  Total Comparisons: 7
  ✓ Matches:         2 (28.6%)
  ✗ Mismatches:      4 (57.1%)
  ⚠ Not Found:       1 (14.3%)

Target: site3/app-dev
────────────────────────────────────────────────────────────────────────────────
  Deployment:
    ✗ 1 mismatch(es):
      - app-deployment
  StatefulSet:
    ⚠ 1 not found:
      - app-stateful
  ConfigMap:
    ✓ 1 match(es)
    ✗ 1 mismatch(es):
      - app-config
  Secret:
    ✓ 1 match(es)
    ✗ 1 mismatch(es):
      - infravalidation-sa-token
  Service:
    ✗ 1 mismatch(es):
      - app-service
```

## Step 4: Generate JSON Report

Save the validation report to a JSON file:

```bash
./mc-tool validate-infra site1/app-prod site3/app-dev --output drift-report.json

# View the report
cat drift-report.json | jq .

# Extract specific diffs
cat drift-report.json | jq -r '.targetResults[0].results[] | select(.resourceName == "app-config") | .diff'
```

## Step 5: Understand the Configuration

Let's look at the site configuration file:

```yaml
# ~/.mc-tool/infra-config.yaml

sites:
  site1:
    name: site1
    endpoint: https://127.0.0.1:42979  # K8s API server endpoint
    token: eyJhbGci...                 # Service account token
    insecure: true                      # Skip TLS verification for KinD
    
  site2:
    name: site2
    endpoint: https://127.0.0.1:42979  # Same cluster, different namespace
    token: eyJhbGci...                 # Different service account token
    insecure: true
    
  site3:
    name: site3
    endpoint: https://127.0.0.1:42979  # Same cluster, different namespace
    token: eyJhbGci...                 # Different service account token
    insecure: true
```

**Note:** In this test setup, all sites point to the same cluster but use different namespaces. In production, you would have:
- Different endpoints for different clusters
- Different tokens for each cluster/namespace
- `insecure: false` for production clusters with valid TLS certificates
  - Service

mode: mode-a                     # Comparison mode

secretComparison: keys-only      # Secret comparison mode
```

## Step 6: Customize for Your Environment

Create your own configuration:

```yaml
# my-infra-config.yaml

baseline:
  context: prod-cluster-1        # Your production cluster
  namespace: my-app

targets:
  - context: prod-cluster-2
    namespace: my-app
  - context: staging-cluster
    namespace: my-app-staging

resourceTypes:
  - Deployment
  - ConfigMap
  - Secret

mode: mode-a
secretComparison: keys-only

# Optional: ignore specific fields
ignoreRules:
  - resourceType: Deployment
    jsonPath: "spec.replicas"    # Ignore replica count differences
```

Then run:

```bash
./mc-tool infravalidate --config my-infra-config.yaml
```

## Step 7: Cleanup

When done testing, cleanup the KinD clusters:

```bash
./scripts/kind-cleanup-infravalidation.sh
```

## Next Steps

### Understanding Results

**Match** ✓
- Configuration is identical between baseline and target

**Mismatch** ✗
- Configuration differs
- Review the diff to understand changes

**Not Found** ⚠
- Resource exists in baseline but not in target
- May indicate missing deployment

**Error** ⚠
- Failed to fetch or compare resource
- Check permissions and connectivity

### Common Use Cases

1. **Pre-deployment validation**
   ```bash
   # Before promoting to production
   mc-tool infravalidate --config staging-to-prod-config.yaml
   ```

2. **Drift detection**
   ```bash
   # Scheduled audit of multiple regions
   mc-tool infravalidate --config multi-region-config.yaml
   ```

3. **Compliance checking**
   ```bash
   # Ensure all clusters follow baseline
   mc-tool infravalidate --config compliance-config.yaml --output audit-report.json
   ```

### Integration with CI/CD

Add to your pipeline:

```yaml
# .gitlab-ci.yml
validate-infrastructure:
  stage: test
  script:
    - mc-tool infravalidate --config infra-config.yaml --output report.json
  artifacts:
    paths:
      - report.json
    when: always
  allow_failure: false  # Fail pipeline if drift detected
```

## Troubleshooting

### "Failed to create client"
- Check kubeconfig context exists: `kubectl config get-contexts`
- Verify cluster is accessible: `kubectl cluster-info`

### "Resource not found"
- Ensure namespace exists in target cluster
- Check RBAC permissions
- Verify resource exists: `kubectl get <resource-type> -n <namespace>`

### "Too many differences"
- Review ignoreRules configuration
- Check if resource types are appropriate
- Verify baseline is the correct reference

## Getting Help

- Full documentation: [docs/INFRAVALIDATION.md](../docs/INFRAVALIDATION.md)
- Requirements spec: [docs/requirements/validate/checklist.md](../docs/requirements/validate/checklist.md)
- Run tests: `go test -v ./pkg/infravalidation/...`

## Summary

You've learned how to:
- ✅ Setup KinD test environment
- ✅ Run infrastructure validation
- ✅ Generate reports
- ✅ Understand results
- ✅ Customize for your environment

Happy validating! 🚀
