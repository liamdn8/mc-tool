# Infrastructure Validation Feature

## Overview

The **Infrastructure Validation** feature enables comparison of Kubernetes namespace configurations across multiple clusters. This helps ensure consistent infrastructure configuration across multi-cluster Kubernetes deployments, detecting configuration drift early.

## Key Features

- ✅ **Multi-cluster namespace comparison** - Compare baseline vs multiple target namespaces
- ✅ **Resource-level configuration diff** - Supports Deployments, StatefulSets, DaemonSets, ConfigMaps, Secrets, Services
- ✅ **Comprehensive resource tracking** - Shows ALL resources (matched, mismatched, not found, and extra)
- ✅ **Extra resource detection** - Identifies resources present in target but missing in baseline
- ✅ **Normalized comparison** - Automatically ignores runtime metadata and ephemeral fields
- ✅ **Flexible Secret handling** - Compare keys-only or hashed values (never plain text)
- ✅ **Advanced diff visualization** - ArgoCD-style diff viewer with Myers algorithm, collapsible hunks, and full/differences toggle
- ✅ **Interactive Web UI** - Modern React-based interface with clickable badges and inline diff viewer
- ✅ **JSON report output** - Machine-readable reports for automation
- ✅ **Mode A implemented** - Structural configuration diff

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Infrastructure Validation                 │
└─────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  K8s Client  │    │  Normalizer  │    │  Comparator  │
└──────────────┘    └──────────────┘    └──────────────┘
        │                   │                   │
        │                   │                   │
        └───────────────────┴───────────────────┘
                            │
                            ▼
                    ┌───────────────┐
                    │  Validator    │
                    └───────────────┘
                            │
                            ▼
                    ┌───────────────┐
                    │    Report     │
                    └───────────────┘
```

### Components

1. **K8s Client** (`k8s_client.go`)
   - Connects to Kubernetes clusters using kubeconfig contexts
   - Fetches resources using dynamic client
   - Supports all configured resource types

2. **Normalizer** (`normalizer.go`)
   - Removes runtime metadata (resourceVersion, uid, timestamps, etc.)
   - Removes status fields
   - Applies resource-specific normalization rules
   - Handles secrets according to comparison mode
   - Sorts arrays for consistent comparison

3. **Comparator** (`comparator.go`)
   - Compares normalized resources
   - Generates human-readable diffs
   - Returns comparison status (match, mismatch, not-found, error)

4. **Validator** (`comparator.go`)
   - Orchestrates the validation process
   - Fetches resources from baseline and targets
   - Coordinates comparison
   - Generates validation report

## Configuration

### YAML Configuration File

```yaml
baseline:
  context: kind-test-baseline        # kubeconfig context
  cluster: test-baseline             # cluster name (descriptive)
  namespace: app-prod                # namespace to use as baseline

targets:
  - context: kind-test-target1
    cluster: test-target1
    namespace: app-prod
  - context: kind-test-target2
    cluster: test-target2
    namespace: app-prod-replica      # namespace names can differ

resourceTypes:
  - Deployment
  - StatefulSet
  - DaemonSet
  - ConfigMap
  - Secret
  - Service

mode: mode-a                          # mode-a: structural diff, mode-b: future policy validation

secretComparison: keys-only           # keys-only or hashed-values

ignoreRules:
  - resourceType: Deployment
    jsonPath: "spec.template.metadata.annotations.deployment.kubernetes.io/revision"
  - resourceType: Service
    jsonPath: "spec.clusterIP"
```

### Configuration Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `baseline` | object | Yes | Baseline cluster/namespace configuration |
| `baseline.context` | string | Yes | Kubeconfig context name |
| `baseline.cluster` | string | No | Descriptive cluster name |
| `baseline.namespace` | string | Yes | Namespace name |
| `targets` | array | Yes | List of target cluster/namespace pairs |
| `targets[].context` | string | Yes | Kubeconfig context name |
| `targets[].cluster` | string | No | Descriptive cluster name |
| `targets[].namespace` | string | Yes | Namespace name (can differ from baseline) |
| `resourceTypes` | array | No | Resource types to compare (defaults to all supported) |
| `mode` | string | No | Comparison mode: `mode-a` (default) or `mode-b` (future) |
| `secretComparison` | string | No | Secret comparison mode: `keys-only` (default) or `hashed-values` |
| `ignoreRules` | array | No | JSONPath patterns to ignore during comparison |

## Usage

### Basic Usage

```bash
# Run validation with config file
mc-tool infravalidate --config config.yaml

# Save JSON report
mc-tool infravalidate --config config.yaml --output report.json
```

### KinD Testing

The feature includes complete KinD (Kubernetes in Docker) test environment:

```bash
# 1. Setup test clusters
./scripts/kind-setup-infravalidation.sh

# This creates:
#   - kind-test-baseline (namespace: app-prod) - baseline cluster
#   - kind-test-target1 (namespace: app-prod) - matching cluster
#   - kind-test-target2 (namespace: app-prod-replica) - drifted cluster

# 2. Run validation
mc-tool infravalidate --config test-data/infravalidation-sample-config.yaml

# 3. Cleanup
./scripts/kind-cleanup-infravalidation.sh
```

### Integration Tests

```bash
# Run unit tests
go test -v ./pkg/infravalidation/... -short

# Run integration tests (requires KinD clusters)
./scripts/kind-setup-infravalidation.sh
go test -v ./pkg/infravalidation/...
./scripts/kind-cleanup-infravalidation.sh
```

## Output

### Console Output

```
🔍 Starting infrastructure validation...
Baseline: kind-test-baseline/app-prod
Targets: 2 cluster(s)
Mode: mode-a

===============================================================================
Infrastructure Validation Summary
===============================================================================

Baseline: kind-test-baseline/app-prod
Timestamp: 2026-01-08T10:30:45Z
Mode: mode-a

Overall Results:
  Total Comparisons: 10
  ✓ Matches:         5 (50.0%)
  ✗ Mismatches:      3 (30.0%)
  ⚠ Not Found:       2 (20.0%)

Target: kind-test-target1/app-prod
───────────────────────────────────────────────────────────────────────────────
  Deployment:
    ✓ 1 match(es)
  ConfigMap:
    ✓ 1 match(es)
  Secret:
    ✓ 1 match(es)

Target: kind-test-target2/app-prod-replica
───────────────────────────────────────────────────────────────────────────────
  Deployment:
    ✗ 1 mismatch(es):
      - app-deployment
  ConfigMap:
    ✗ 1 mismatch(es):
      - app-config
  Service:
    ⚠ 1 not found:
      - app-service

===============================================================================

⚠️  Configuration drift detected!
```

### JSON Report

```json
{
  "baseline": {
    "context": "kind-test-baseline",
    "cluster": "test-baseline",
    "namespace": "app-prod"
  },
  "timestamp": "2026-01-08T10:30:45Z",
  "mode": "mode-a",
  "targetResults": [
    {
      "target": {
        "context": "kind-test-target1",
        "cluster": "test-target1",
        "namespace": "app-prod"
      },
      "results": [
        {
          "resourceType": "Deployment",
          "resourceName": "app-deployment",
          "status": "match"
        }
      ]
    }
  ],
  "summary": {
    "totalComparisons": 12,
    "matchCount": 5,
    "mismatchCount": 3,
    "notFoundCount": 2,
    "extraCount": 2,
    "errorCount": 0,
    "statusBreakdown": {
      "match": 5,
      "mismatch": 3,
      "not-found": 2,
      "extra": 2
    }
  }
}
```

## Normalization Rules

### Automatic Normalization

The following fields are automatically normalized (removed) during comparison:

**Metadata:**
- `resourceVersion`
- `uid`
- `generation`
- `creationTimestamp`
- `managedFields`
- `selfLink`

**Runtime Annotations:**
- `kubectl.kubernetes.io/last-applied-configuration`
- `deployment.kubernetes.io/revision`
- `autoscaling.alpha.kubernetes.io/*`

**Resource-specific:**
- Deployment: `spec.revisionHistoryLimit`, `spec.progressDeadlineSeconds`
- DaemonSet: `spec.replicas` (auto-managed)
- Service: `spec.clusterIP`, `spec.clusterIPs`, `spec.ports[].nodePort` (if not NodePort/LoadBalancer)

### Secret Handling

**Keys-only mode (default):**
```yaml
# All secret values replaced with placeholder
data:
  db-password: "***"
  api-key: "***"
```

**Hashed-values mode:**
```yaml
# Secret values hashed with SHA256
data:
  db-password: "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"
  api-key: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
```

## Web UI

### Features

**Enhanced Resource Display:**
- ✅ Shows ALL resources including matched ones (not just drift)
- ✅ Clear baseline status badges ("Configured" instead of "-")
- ✅ Extra resources highlighted with warning badges
- ✅ Interactive table with filtering and pagination
- ✅ Status counts displayed in header badges

**Advanced Diff Viewer:**
- ✅ ArgoCD-style side-by-side comparison
- ✅ Myers diff algorithm for accurate content-based diffs
- ✅ Collapsible hunks with context lines
- ✅ "Show full" vs "Show only differences" toggle
- ✅ VIM/namespace labels in diff header
- ✅ Separate copy buttons for baseline and target
- ✅ Clickable badges to view diff for Match, Mismatch, and Extra resources

**Navigation & UX:**
- ✅ Sidebar navigation with auto-scroll to sections
- ✅ Contents panel below Navigation
- ✅ Resource type grouping with visual status indicators
- ✅ Search and filter capabilities
- ✅ Mobile-responsive design

### Resource Status Badges

| Status | Baseline | Target | Description | Clickable |
|--------|----------|--------|-------------|----------|
| **Match** | "Match" (green) | "Match" (green) | Configuration identical | ✅ Yes - View diff |
| **Mismatch** | "Configured" (green) | "Mismatch" (warning) | Configuration differs | ✅ Yes - View diff |
| **Not Found** | N/A | "Not Found" (warning) | In baseline but missing in target | ❌ No |
| **Extra** | "Not Found" (warning) | "Extra" (warning) | In target but not in baseline | ✅ Yes - View diff |

### Starting Web UI

```bash
# Start on default port 8080
mc-tool web --port 8080

# Access at http://localhost:8080/minio-webtool/validate/infrastructure
```

### Using Web UI

1. **Select Configuration:**
   - Choose baseline VIM and namespace
   - Add one or more target VIM/namespace pairs
   
2. **Run Validation:**
   - Click "Validate Infrastructure"
   - Watch real-time progress
   
3. **Review Results:**
   - Overview cards show summary statistics
   - Navigate using sidebar Contents panel
   - Filter by resource type
   - Search for specific resources
   
4. **Inspect Differences:**
   - Click any badge (Match/Mismatch/Extra) to view diff
   - Toggle between "Show only differences" and "Show full"
   - Copy baseline or target YAML
   - Close diff viewer to return to table

## Exit Codes

- `0` - All configurations match
- `1` - Configuration drift detected (mismatches, not-found, or extra resources)

## Requirements

- Go 1.24+
- Kubernetes 1.23+
- kubeconfig with access to target clusters
- Read permissions on target namespaces

## Future Enhancements (Mode B)

- Policy/invariant validation (OPA/Kyverno-style rules)
- Severity levels
- Compliance reporting
- Export to HTML/Excel
- CI/CD integration
- Scheduled audits

## Troubleshooting

### "Failed to connect to cluster"
- Verify kubeconfig context exists: `kubectl config get-contexts`
- Test cluster connectivity: `kubectl --context <context> get nodes`
- Check RBAC permissions

### "Resource not found"
- Verify namespace exists: `kubectl --context <context> get ns`
- Check if resources exist in namespace: `kubectl --context <context> -n <namespace> get <resource-type>`

### "Config validation failed"
- Ensure all required fields are present
- Verify resource types are valid
- Check YAML syntax

## See Also

- [SRS Document](requirements/validate/checklist.md) - Full requirements specification
- [KinD Documentation](https://kind.sigs.k8s.io/) - Kubernetes in Docker
- [Sample Config](../test-data/infravalidation-sample-config.yaml) - Example configuration
