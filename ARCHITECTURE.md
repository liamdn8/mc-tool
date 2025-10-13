# MinIO Site Replication Management Tool - Architecture

## Overview
Tool này được thiết kế để quản lý và maintain các MinIO instances trong một Site Replication setup.

## Core Concepts

### 1. Replication Group (Site Replication Cluster)
- Một nhóm các MinIO sites được cấu hình để replicate với nhau
- Mỗi site có thể có 1 hoặc nhiều nodes
- Tất cả sites trong group phải có cùng cấu hình về:
  - Buckets
  - Bucket policies
  - IAM policies
  - Users & Groups
  - Lifecycle policies
  - Encryption settings

### 2. Site (MinIO Deployment)
- Một MinIO deployment (có thể là standalone hoặc distributed)
- Được đại diện bởi một mc alias
- Ví dụ: site1, site2, site3

### 3. Management Operations

#### A. Discovery & Health Check
- Tự động discover tất cả sites trong replication group
- Health check từng site
- Monitor replication lag
- Kiểm tra connectivity giữa các sites

#### B. Configuration Sync
- So sánh cấu hình giữa các sites
- Tự động detect inconsistencies
- Sync configurations:
  - Bucket policies
  - Lifecycle rules
  - IAM policies
  - User/Group permissions

#### C. Bucket Operations
- List buckets across all sites
- Verify bucket replication status
- Check object count consistency
- Compare bucket configurations

#### D. Troubleshooting
- Identify replication failures
- Show sync lag metrics
- Detect configuration drift
- Generate diagnostic reports

## Workflow

```
┌─────────────────────────────────────────────────────────┐
│              MC Tool - Replication Manager              │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
        ┌───────────────────────────────────────┐
        │   Auto-discover Replication Group     │
        │   (via mc admin replicate info)       │
        └───────────────────────────────────────┘
                            │
                ┌───────────┴───────────┐
                ▼                       ▼
        ┌──────────────┐        ┌──────────────┐
        │   Site 1     │◄──────►│   Site 2     │
        │ (minio1:9000)│        │ (minio2:9000)│
        └──────────────┘        └──────────────┘
                │                       │
                └───────────┬───────────┘
                            ▼
            ┌───────────────────────────────┐
            │   Configuration Comparison     │
            │   - Buckets                   │
            │   - Policies                  │
            │   - IAM                       │
            │   - Lifecycle                 │
            └───────────────────────────────┘
                            │
                            ▼
            ┌───────────────────────────────┐
            │   Detect Inconsistencies      │
            │   & Suggest Fixes             │
            └───────────────────────────────┘
                            │
                            ▼
            ┌───────────────────────────────┐
            │   Auto-sync (if enabled)      │
            └───────────────────────────────┘
```

## Key Features to Implement

### Phase 1: Discovery & Monitoring
- [ ] Auto-detect replication group from any site
- [ ] List all sites in replication group
- [ ] Health check per site
- [ ] Show replication status matrix

### Phase 2: Bucket Management
- [ ] List all buckets across sites
- [ ] Show which buckets are replicated
- [ ] Identify missing buckets on any site
- [ ] Compare bucket configurations

### Phase 3: Configuration Sync
- [ ] Compare bucket policies
- [ ] Compare lifecycle rules
- [ ] Sync policies across sites
- [ ] Sync lifecycle rules

### Phase 4: Advanced Operations
- [ ] IAM policy sync
- [ ] User/Group management
- [ ] Encryption settings verification
- [ ] Replication metrics & analytics

## API Design

### GET /api/replication/info
Get site replication group information
```json
{
  "enabled": true,
  "sites": [
    {
      "name": "site1",
      "endpoint": "http://localhost:9001",
      "deploymentID": "xxx",
      "healthy": true
    },
    {
      "name": "site2", 
      "endpoint": "http://localhost:9002",
      "deploymentID": "yyy",
      "healthy": true
    }
  ]
}
```

### GET /api/replication/status
Get replication status between sites
```json
{
  "status": "healthy",
  "sites": {
    "site1": {
      "replicatedBuckets": 5,
      "pendingObjects": 0,
      "failedObjects": 0,
      "lastSyncTime": "2025-10-10T10:00:00Z"
    }
  }
}
```

### POST /api/replication/sync
Trigger configuration sync
```json
{
  "type": "bucket-policy|lifecycle|iam",
  "source": "site1",
  "targets": ["site2"],
  "bucket": "mybucket" // optional, for bucket-specific ops
}
```

### GET /api/replication/compare
Compare configurations across sites
```json
{
  "buckets": {
    "mybucket": {
      "existsOn": ["site1", "site2"],
      "policy": {
        "consistent": false,
        "differences": [...]
      },
      "lifecycle": {
        "consistent": true
      }
    }
  }
}
```

## MC Commands Used

### Discovery
```bash
# Get replication info
mc admin replicate info ALIAS

# Get site info
mc admin info ALIAS
```

### Bucket Operations
```bash
# List buckets
mc ls ALIAS

# Get bucket policy
mc policy get ALIAS/bucket

# Get lifecycle
mc ilm ls ALIAS/bucket
```

### Sync Operations
```bash
# Set bucket policy
mc policy set-json policy.json ALIAS/bucket

# Add lifecycle rule
mc ilm add --expiry-days 30 ALIAS/bucket
```

## UI Flow

1. **Dashboard** - Overview của replication group
   - Total sites
   - Health status
   - Recent sync activities
   - Alerts/Warnings

2. **Sites View** - Chi tiết từng site
   - Endpoint, version
   - Storage stats
   - Buckets count
   - Replication lag

3. **Buckets View** - Quản lý buckets
   - List all buckets (grouped by name)
   - Show replication status per bucket
   - Configuration comparison
   - Quick sync actions

4. **Configuration Sync** - So sánh & đồng bộ
   - Visual diff của configurations
   - One-click sync
   - Sync history
   - Rollback capability

5. **Operations** - Automated tasks
   - Scheduled sync jobs
   - Health monitoring
   - Alerts configuration
   - Audit logs
