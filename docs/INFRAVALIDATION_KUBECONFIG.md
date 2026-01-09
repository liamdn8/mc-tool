# Infrastructure Validation - Kubeconfig Configuration

## Overview

Infrastructure Validation now uses **standard kubeconfig** format for Kubernetes cluster access, providing seamless integration with kubectl and other Kubernetes tools.

## Configuration Methods

### Method 1: Default Kubeconfig (Recommended)

Load contexts from `~/.kube/config` automatically:

```bash
mc-tool web --port 8080
```

### Method 2: Custom Kubeconfig Path

Specify a custom kubeconfig file:

```bash
mc-tool web --port 8080 --config-dir /path/to/custom/kubeconfig
```

You can also set the `KUBECONFIG` environment variable:

```bash
export KUBECONFIG=/path/to/custom/kubeconfig
mc-tool web --port 8080
```

## API Endpoints

### List Available VIMs (Contexts)

**Endpoint:** `GET /api/validate/infrastructure/vims`

**Response:**
```json
{
  "vims": [
    {
      "name": "docker-desktop",
      "context": "docker-desktop",
      "current": false
    },
    {
      "name": "kind-kind-infra-test",
      "context": "kind-kind-infra-test",
      "current": true
    }
  ],
  "count": 2,
  "currentContext": "kind-kind-infra-test",
  "configPath": "~/.kube/config (default)"
}
```

### List Namespaces in Context

**Endpoint:** `GET /api/validate/infrastructure/namespaces?vim=<context-name>`

**Example:**
```bash
curl "http://localhost:8080/minio-webtool/api/validate/infrastructure/namespaces?vim=kind-kind-infra-test"
```

**Response:**
```json
{
  "namespaces": [
    "default",
    "kube-system",
    "kube-public",
    "local-path-storage"
  ]
}
```

### Search Namespaces Across VIMs

**Endpoint:** `GET /api/validate/infrastructure/search-namespaces?keyword=<search>&exactMatch=<bool>`

**Example:**
```bash
# Search for namespaces containing "prod"
curl "http://localhost:8080/minio-webtool/api/validate/infrastructure/search-namespaces?keyword=prod&exactMatch=false"

# Exact match for namespace "production"
curl "http://localhost:8080/minio-webtool/api/validate/infrastructure/search-namespaces?keyword=production&exactMatch=true"
```

**Response:**
```json
{
  "matches": [
    {
      "vim": "site-a-prod",
      "namespace": "app-production",
      "exact": true
    },
    {
      "vim": "site-b-staging",
      "namespace": "app-prod-mirror",
      "exact": false
    }
  ],
  "count": 2
}
```

### Discover Resources in Namespace

**Endpoint:** `GET /api/validate/infrastructure/discover-resources?vim=<context>&namespace=<ns>`

**Example:**
```bash
curl "http://localhost:8080/minio-webtool/api/validate/infrastructure/discover-resources?vim=kind-kind-infra-test&namespace=default"
```

**Response:**
```json
{
  "resources": [
    "pods",
    "services",
    "configmaps",
    "secrets",
    "deployments",
    "replicasets",
    "statefulsets",
    "daemonsets",
    "jobs",
    "cronjobs",
    "ingresses",
    "persistentvolumeclaims",
    "customresource.example.com"
  ],
  "count": 13
}
```

## Usage Examples

### Quick Validation with Context

1. **Start web UI:**
   ```bash
   mc-tool web --port 8080
   ```

2. **Get available contexts:**
   ```bash
   curl http://localhost:8080/minio-webtool/api/validate/infrastructure/vims | jq .
   ```

3. **List namespaces in context:**
   ```bash
   curl "http://localhost:8080/minio-webtool/api/validate/infrastructure/namespaces?vim=kind-test" | jq .
   ```

4. **Discover resources:**
   ```bash
   curl "http://localhost:8080/minio-webtool/api/validate/infrastructure/discover-resources?vim=kind-test&namespace=default" | jq .
   ```

### Using Custom Kubeconfig

```bash
# Create custom kubeconfig
kubectl config view --flatten > /tmp/my-clusters.yaml

# Start web UI with custom config
mc-tool web --port 8080 --config-dir /tmp/my-clusters.yaml

# List VIMs from custom config
curl http://localhost:8080/minio-webtool/api/validate/infrastructure/vims | jq .
```

## Web UI Features

The Infrastructure Validation page provides:

- **VIM Selection:** Dropdown with all available kubeconfig contexts
- **Quick Validation Mode:** 
  - Search namespaces across all contexts
  - Exact or fuzzy matching
  - Auto-populate baseline and targets
- **Manual Mode:**
  - Select baseline VIM/namespace
  - Add multiple target VIM/namespace pairs
  - Advanced resource type filtering
- **Auto-Discovery:**
  - Automatically detect all resource types in namespace
  - Include CRDs, DaemonSets, StatefulSets
  - Filter out static pods (managed by Kubelet)

## Kubeconfig Requirements

Your kubeconfig must contain:

1. **Clusters:** Kubernetes API server endpoints
2. **Users:** Authentication credentials (certificates, tokens, etc.)
3. **Contexts:** Cluster + User + Namespace mappings

Example kubeconfig structure:
```yaml
apiVersion: v1
kind: Config
clusters:
- name: site-a-prod
  cluster:
    server: https://api.site-a.example.com:6443
    certificate-authority-data: LS0tLS...
users:
- name: admin-site-a
  user:
    client-certificate-data: LS0tLS...
    client-key-data: LS0tLS...
contexts:
- name: site-a-prod
  context:
    cluster: site-a-prod
    user: admin-site-a
    namespace: default
current-context: site-a-prod
```

## Migration from Legacy Config

If you previously used custom `infra-config.yaml` with endpoint/token format:

**Old format (deprecated):**
```yaml
sites:
  - name: site1
    endpoint: https://127.0.0.1:6443
    token: eyJhbGciOiJSUzI1NiIsImtpZCI6Ik...
    insecure: true
```

**New format (standard kubeconfig):**
```yaml
# Just use your existing ~/.kube/config or create context-based config
apiVersion: v1
kind: Config
contexts:
- name: site1
  context:
    cluster: my-cluster
    user: my-user
```

**Migration steps:**
1. Remove `~/.mc-tool/infra-config.yaml` if exists
2. Use `kubectl config` to manage contexts
3. Start mc-tool web without any custom config (uses default kubeconfig)

## Troubleshooting

### Error: "Failed to load kubeconfig contexts"

**Cause:** Invalid or missing kubeconfig file

**Solution:**
```bash
# Verify kubeconfig is valid
kubectl config view

# Check current context
kubectl config current-context

# List all contexts
kubectl config get-contexts
```

### Error: "Connection refused" when listing namespaces

**Cause:** Kubernetes cluster is not running or unreachable

**Solution:**
```bash
# Test cluster connectivity
kubectl get nodes

# Check cluster status
kubectl cluster-info

# For kind clusters, ensure cluster is running:
kind get clusters
```

### Custom Kubeconfig Not Loaded

**Cause:** Flag name confusion (--config-dir vs KUBECONFIG)

**Solution:**
```bash
# Use --config-dir flag
mc-tool web --config-dir /path/to/kubeconfig

# OR set environment variable
export KUBECONFIG=/path/to/kubeconfig
mc-tool web
```

## Best Practices

1. **Use kubectl contexts:** Manage all clusters through standard kubectl config
2. **Name contexts clearly:** Use descriptive names like `site-a-prod`, `site-b-staging`
3. **Set current context:** Default context will be highlighted in UI
4. **Secure kubeconfig:** Keep credentials safe, use RBAC with minimal permissions
5. **Test connectivity:** Verify `kubectl get nodes` works before using mc-tool

## See Also

- [Infrastructure Validation Overview](INFRAVALIDATION.md)
- [Quick Start Guide](INFRAVALIDATION_QUICKSTART.md)
- [Web UI Guide](INFRAVALIDATION_WEB_UI.md)
