# Infrastructure Validation Web UI Guide

Complete guide for using the Infrastructure Validation feature through the Web UI.

## Overview

The Infrastructure Validation Web UI provides a modern, intuitive interface for comparing Kubernetes namespace configurations across multiple clusters. Built with React and featuring an ArgoCD-style diff viewer, it makes configuration drift detection accessible and efficient.

## Key Features

### 🎯 Comprehensive Resource Tracking
- **All Resources Displayed**: Shows matched, mismatched, not found, AND extra resources
- **Extra Resource Detection**: Identifies resources in target clusters that don't exist in baseline
- **Status Badges**: Clear visual indicators for each resource state
- **Resource Grouping**: Organized by type (Deployment, StatefulSet, ConfigMap, Secret, Service)

### 🔍 Advanced Diff Viewer
- **ArgoCD-Style Interface**: Familiar side-by-side comparison
- **Myers Algorithm**: Accurate content-based diffs (same algorithm as Git)
- **Collapsible Hunks**: Expand/collapse change sections
- **Full View Toggle**: Switch between "Show only differences" and "Show full"
- **Syntax Highlighting**: YAML syntax coloring for readability
- **VIM/Namespace Labels**: Clear header showing what's being compared
- **Copy Functionality**: Separate copy buttons for baseline and target

### 🎨 User Experience
- **Interactive Badges**: Click Match, Mismatch, or Extra badges to view diffs
- **Sidebar Navigation**: Quick jump to resource types with auto-scroll
- **Search & Filter**: Find specific resources instantly
- **Real-time Progress**: Live updates during validation
- **Mobile Responsive**: Works on tablets and phones
- **Contents Panel**: Integrated navigation below main menu

## Getting Started

### Prerequisites

1. **Infrastructure Configuration**:
   ```yaml
   # ~/.mc-tool/infra-config.yaml
   sites:
     site1:
       endpoint: "https://cluster1.example.com"
       token: "eyJhbGci..."
       insecure: false
     site2:
       endpoint: "https://cluster2.example.com"
       token: "eyJhbGci..."
       insecure: false
   ```

2. **Start Web Server**:
   ```bash
   mc-tool web --port 8080
   ```

3. **Access UI**:
   ```
   http://localhost:8080/minio-webtool/validate/infrastructure
   ```

## Step-by-Step Usage

### 1. Configure Validation

#### Select Baseline
1. **Choose VIM**: Select the Kubernetes cluster to use as reference
2. **Select Namespace**: Pick the namespace containing baseline configuration

The VIM list is automatically loaded from your `~/.mc-tool/infra-config.yaml`.

#### Add Targets
1. Click **"Add Target"** button
2. **Choose VIM**: Select target cluster
3. **Select Namespace**: Pick namespace to compare against baseline
4. **Add Multiple**: Click "Add Target" again for more comparisons
5. **Remove**: Use trash icon to remove unwanted targets

### 2. Run Validation

Click **"Validate Infrastructure"** button.

**What Happens:**
- Job starts on backend
- Real-time progress indicator appears
- Validation compares all configured resource types:
  - Deployments
  - StatefulSets
  - DaemonSets
  - ConfigMaps
  - Secrets (keys-only comparison)
  - Services

### 3. Review Results

#### Overview Cards

At the top, you'll see summary statistics:

```
┌─────────────┬─────────────┬─────────────┬─────────────┬─────────────┐
│   Total     │    Match    │  Mismatch   │ Not Found   │    Extra    │
│     15      │      8      │      3      │      2      │      2      │
└─────────────┴─────────────┴─────────────┴─────────────┴─────────────┘
```

- **Total**: All resources compared
- **Match**: Identical configurations
- **Mismatch**: Configuration differences detected
- **Not Found**: In baseline but missing in target
- **Extra**: In target but not in baseline

#### Navigation Panel

Sidebar shows resource types with status indicators:
- 🟢 **Green checkmark**: All resources match
- 🔴 **Red X**: Drift detected (mismatch/not found/extra)

Click any resource type to scroll to that section.

#### Resource Tables

Resources are grouped by type. Each table shows:

**Header:**
- Resource type name
- Status icon (✓ or ✗)
- Count badges (Match: X, Mismatch: Y, etc.)

**Filters:**
- Search box: Filter by resource name
- Status dropdown: Show all, matched only, mismatches only, etc.

**Table Columns:**
- **Status Icon**: Quick visual indicator
- **Resource Name**: Name of the resource
- **Baseline Column**: Status in baseline namespace
- **Target Column(s)**: Status in each target namespace

**Pagination:**
- 10 resources per page
- Navigation arrows at bottom

### 4. Understanding Status Badges

| Badge | Color | Meaning | Clickable | When to Click |
|-------|-------|---------|-----------|---------------|
| **Match** | Green | Configurations are identical | ✅ Yes | Verify exact match |
| **Configured** | Green | Baseline reference | ❌ No | Shows baseline has this resource |
| **Mismatch** | Warning (Orange) | Configuration differs | ✅ Yes | **View what changed** |
| **Extra** | Warning (Orange) | Only in target, not baseline | ✅ Yes | **Inspect unauthorized resource** |
| **Not Found** | Warning (Orange) | Only in baseline, missing in target | ❌ No | Indicates missing deployment |

### 5. Viewing Diffs

#### Opening Diff Viewer

Click any **clickable badge** (Match, Mismatch, or Extra):
- Diff viewer modal opens
- Loads resource YAML from both sides
- Computes differences using Myers algorithm

#### Diff Viewer Interface

**Header:**
```
[Resource Type] resource-name
Baseline: site1/app-staging  |  Target: site2/app-staging
```

**Toggle Buttons:**
- 🔘 **Show only differences** (default): Collapsed hunks, focus on changes
- 🔘 **Show full**: Complete file content, expanded hunks

**Action Buttons:**
- 📋 **Copy Baseline**: Copy baseline YAML to clipboard
- 📋 **Copy Target**: Copy target YAML to clipboard
- ❌ **Close**: Return to table view

**Diff Display:**

**Collapsed Mode (Show only differences):**
```yaml
┌─────────────────────────────────────────────────────────────────────────┐
│ Hunk 1/3: Lines 15-23 ▼                                                 │
├──────────────────────────────────────┬──────────────────────────────────┤
│ metadata:                            │ metadata:                        │
│   labels:                            │   labels:                        │
│     app: myapp                       │     app: myapp                   │
│ -   version: "1.0"                   │ +   version: "2.0"               │
│     env: staging                     │     env: staging                 │
└──────────────────────────────────────┴──────────────────────────────────┘

● 15 lines unchanged ●

┌─────────────────────────────────────────────────────────────────────────┐
│ Hunk 2/3: Lines 38-45 ▼                                                 │
├──────────────────────────────────────┬──────────────────────────────────┤
│   spec:                              │   spec:                          │
│ -   replicas: 2                      │ +   replicas: 3                  │
│     selector:                        │     selector:                    │
└──────────────────────────────────────┴──────────────────────────────────┘
```

**Expanded Mode (Show full):**
```yaml
┌─────────────────────────────────────────────────────────────────────────┐
│ Lines 1-100 (complete file)                                             │
├──────────────────────────────────────┬──────────────────────────────────┤
│ apiVersion: apps/v1                  │ apiVersion: apps/v1              │
│ kind: Deployment                     │ kind: Deployment                 │
│ metadata:                            │ metadata:                        │
│   name: app-deployment               │   name: app-deployment           │
│   labels:                            │   labels:                        │
│     app: myapp                       │     app: myapp                   │
│ -   version: "1.0"                   │ +   version: "2.0"               │
│     env: staging                     │     env: staging                 │
│ spec:                                │ spec:                            │
│ -   replicas: 2                      │ +   replicas: 3                  │
│   selector:                          │   selector:                      │
│     matchLabels:                     │     matchLabels:                 │
│       app: myapp                     │       app: myapp                 │
│   template:                          │   template:                      │
│     ...                              │     ...                          │
└──────────────────────────────────────┴──────────────────────────────────┘
```

**Color Coding:**
- 🟢 **Green background**: Added lines (+ prefix)
- 🔴 **Red background**: Removed lines (- prefix)
- ⚪ **White background**: Unchanged lines (no prefix)
- 🔵 **Blue header**: Section unchanged (collapsed)

#### Interpreting Diffs

**Match Status:**
- Shows identical content on both sides
- No red/green highlighting
- Use to verify configurations are truly identical

**Mismatch Status:**
- Baseline (left) shows configured values
- Target (right) shows actual values with differences highlighted
- Red lines = removed from baseline
- Green lines = added in target

**Extra Status:**
- Baseline (left) shows "Resource not found in baseline"
- Target (right) shows complete resource YAML
- Helps identify unauthorized or untracked resources

### 6. Workflow Examples

#### Example 1: Pre-Deployment Validation

**Scenario**: Verify staging matches production before promoting

1. Baseline: `production/myapp`
2. Target: `staging/myapp`
3. Run validation
4. **Expected**: All Match
5. **If Mismatch found**:
   - Click badge to view diff
   - Identify what needs updating in staging
   - Fix configuration
   - Re-run validation

#### Example 2: Multi-Region Consistency

**Scenario**: Ensure all regions have identical configuration

1. Baseline: `us-east-1/myapp`
2. Targets:
   - `us-west-2/myapp`
   - `eu-central-1/myapp`
   - `ap-south-1/myapp`
3. Run validation
4. Review each region's table column
5. Click any Mismatch to investigate regional differences

#### Example 3: Drift Detection

**Scenario**: Find unauthorized changes in production

1. Baseline: `gitops-repo/myapp` (assumed baseline)
2. Target: `production/myapp`
3. Run validation
4. **Check for**:
   - **Mismatch**: Configuration changed manually
   - **Extra**: Resources added outside GitOps
   - **Not Found**: Resources deleted
5. Click badges to inspect changes
6. Remediate or update baseline as needed

#### Example 4: Disaster Recovery Verification

**Scenario**: Verify DR environment matches production

1. Baseline: `production/critical-app`
2. Target: `dr-site/critical-app`
3. Run validation
4. **Must be**: 100% Match
5. Any Mismatch = DR not ready
6. Fix immediately and re-validate

## Advanced Features

### Search and Filter

**Search by Name:**
```
[🔍 Search resource name...]
```
- Type resource name
- Table filters in real-time
- Case-insensitive

**Filter by Status:**
```
[Status: All ▼]
  - All
  - Match
  - Mismatch
  - Not Found
  - Extra
```
- Show only specific status
- Useful for focusing on issues

**Combine Filters:**
- Search + Status filter = powerful querying
- Example: Search "database" + Status "Mismatch"

### Pagination

- Default: 10 resources per page
- Navigate with ◀ ▶ arrows
- Shows "Page X of Y"
- Filters apply across all pages

### Real-time Updates

- Validation runs asynchronously
- Progress indicator shows status
- Results update when complete
- No page refresh needed

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Esc` | Close diff viewer |
| `Ctrl/Cmd + K` | Focus search box |
| `Ctrl/Cmd + C` | Copy (when in diff viewer) |

## Troubleshooting

### No VIMs Available

**Problem**: VIM dropdown is empty

**Solution**:
1. Check `~/.mc-tool/infra-config.yaml` exists
2. Verify sites are configured:
   ```yaml
   sites:
     site1:
       endpoint: "https://..."
       token: "..."
   ```
3. Restart web server

### Namespaces Not Loading

**Problem**: Namespace dropdown shows "No namespaces available"

**Solutions**:
- Verify cluster connectivity
- Check service account has list permissions:
  ```bash
  kubectl auth can-i list namespaces --as=system:serviceaccount:default:infravalidation-sa
  ```
- Check token is valid

### Diff Not Loading

**Problem**: Click badge but diff doesn't appear

**Solutions**:
- Check browser console for errors
- Verify resource exists: `kubectl get <type> <name> -n <namespace>`
- Ensure service account can read resources
- Check backend logs: `/tmp/mc-tool-web.log`

### Slow Performance

**Problem**: Validation takes too long

**Optimizations**:
- Reduce number of resource types in config
- Use namespaces with fewer resources for testing
- Check network latency to clusters
- Ensure clusters are responsive: `kubectl cluster-info`

## Best Practices

### 1. Baseline Selection
✅ **DO**: Use GitOps repo or production as baseline  
❌ **DON'T**: Use dev/test as baseline for prod validation

### 2. Regular Audits
✅ **DO**: Schedule weekly validation runs  
❌ **DON'T**: Only validate after incidents

### 3. Ignore Rules
✅ **DO**: Configure ignoreRules for expected differences  
❌ **DON'T**: Ignore security-critical fields

### 4. Extra Resources
✅ **DO**: Investigate all Extra resources immediately  
❌ **DON'T**: Assume Extra resources are harmless

### 5. Matched Resources
✅ **DO**: Review matches periodically to verify accuracy  
❌ **DON'T**: Blindly trust Match status

## Integration with CI/CD

While Web UI is for manual validation, integrate with CI/CD using CLI:

```yaml
# .gitlab-ci.yml
validate-infrastructure:
  stage: validate
  script:
    - mc-tool validate-infra site1/production site2/production site3/production
    - mc-tool validate-infra --output report.json
  artifacts:
    paths:
      - report.json
  only:
    - main
```

Then review in Web UI:
1. Download `report.json`
2. Open Web UI
3. Upload report (future feature)

## Security Considerations

### Secrets Handling
- **Keys-only comparison**: Default mode shows only secret keys, not values
- **Hashed comparison**: Optional - shows SHA256 hash of values
- **Never plain text**: Secret values never exposed in UI

### Access Control
- Web UI runs locally by default
- For shared deployment:
  - Use reverse proxy with authentication
  - Enable HTTPS
  - Configure CORS
  - Restrict network access

### Token Storage
- Tokens stored in `~/.mc-tool/infra-config.yaml`
- File should be `chmod 600` (owner read/write only)
- Never commit to git

## Comparison with CLI

| Feature | Web UI | CLI |
|---------|--------|-----|
| **Ease of Use** | ✅ Visual, intuitive | ⚠️ Requires command knowledge |
| **Diff Viewer** | ✅ Side-by-side with syntax highlighting | ⚠️ Terminal output |
| **All Resources** | ✅ Shows matched resources | ⚠️ Optional with verbose flag |
| **Extra Detection** | ✅ Automatic | ✅ Automatic |
| **Navigation** | ✅ Click-based | ⚠️ Scroll/search output |
| **CI/CD** | ⚠️ Manual | ✅ Scriptable |
| **JSON Export** | ⚠️ Copy to clipboard | ✅ `--output` flag |
| **Offline** | ⚠️ Requires web server | ✅ Fully offline |

**Recommendation**: Use Web UI for manual validation and investigation, CLI for automation.

## Future Enhancements

Planned features:
- [ ] Report upload and history
- [ ] Scheduled validations
- [ ] Email/Slack notifications
- [ ] Diff export to PDF
- [ ] Resource health checks
- [ ] Policy validation (Mode B)
- [ ] Multi-select comparison
- [ ] Baseline versioning

## Related Documentation

- [Infrastructure Validation Overview](INFRAVALIDATION.md)
- [Quick Start Guide](INFRAVALIDATION_QUICKSTART.md)
- [Web UI Documentation](WEB_UI.md)
- [CLI Reference](../README.md)

## Support

For issues or questions:
1. Check troubleshooting section above
2. Review logs: `/tmp/mc-tool-web.log`
3. Test CLI equivalent command
4. Check GitHub issues
5. Open new issue with:
   - Web UI version
   - Browser/OS
   - Steps to reproduce
   - Console errors
   - Backend logs
