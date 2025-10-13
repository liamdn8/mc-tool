# MinIO Site Replication Management Tool

A comprehensive web-based tool for managing and maintaining MinIO instances in a Site Replication setup.

## 🎯 Purpose

This tool is designed to help operations teams maintain MinIO Site Replication deployments by providing:

- **Unified Dashboard** - Overview of all sites in your replication group
- **Configuration Sync** - Detect and fix configuration inconsistencies
- **Health Monitoring** - Real-time status of all sites
- **Automated Operations** - One-click sync and maintenance tasks

## 🏗️ Architecture

### What is MinIO Site Replication?

MinIO Site Replication allows you to replicate data across multiple MinIO deployments (sites). Each site can be:
- In different data centers
- Different geographical regions
- Different cloud providers

### How This Tool Helps

```
┌──────────────────────────────────────────────────┐
│         MC-Tool Replication Manager              │
│                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐      │
│  │  Site 1  │  │  Site 2  │  │  Site 3  │      │
│  └──────────┘  └──────────┘  └──────────┘      │
│                                                  │
│  Features:                                       │
│  • Auto-discover all sites                      │
│  • Compare configurations                       │
│  • Sync bucket policies                         │
│  • Monitor replication status                   │
│  • Health checks                                 │
└──────────────────────────────────────────────────┘
```

## 🚀 Quick Start

### Prerequisites

1. **MinIO Client (mc)** must be installed and configured
   ```bash
   # Install mc
   wget https://dl.min.io/client/mc/release/linux-amd64/mc
   chmod +x mc
   sudo mv mc /usr/local/bin/
   
   # Verify installation
   mc --version
   ```

2. **Add your MinIO sites as aliases**
   ```bash
   # Add site1
   mc alias set site1 https://minio1.example.com ACCESS_KEY SECRET_KEY
   
   # Add site2
   mc alias set site2 https://minio2.example.com ACCESS_KEY SECRET_KEY
   
   # Verify
   mc alias list
   ```

### Installation

```bash
# Clone or download the tool
git clone https://github.com/liamdn8/mc-tool.git
cd mc-tool

# Build
go build -ldflags="-s -w" -o mc-tool main.go

# Run
./mc-tool web
```

Then open http://localhost:8080 in your browser.

## 📊 Features

### 1. Overview Dashboard

The main dashboard shows:
- **Total Sites** - Number of MinIO sites configured
- **Synced Buckets** - Buckets replicated across sites
- **Total Objects** - Combined object count
- **Health Status** - Overall replication health

### 2. Sites Management

View all sites in your replication group:
- Site name and endpoint
- Number of buckets
- Object count
- Health status
- Quick actions

### 3. Buckets View

See all buckets across sites:
- Which buckets exist on which sites
- Replication status (fully replicated or partial)
- Quick comparison of bucket configurations

### 4. Replication Status

Monitor replication health:
- Per-site replication metrics
- Pending objects count
- Failed replication count
- Last sync time
- Real-time health indicators

### 5. Consistency Check

Compare configurations across sites:
- **Bucket Policies** - Detect policy differences
- **Lifecycle Rules** - Compare ILM settings
- **Versioning** - Check versioning status
- Visual indicators for consistent/inconsistent configs

### 6. Automated Operations

One-click operations to maintain your sites:

#### Sync Bucket Policies
Copies bucket policies from one site to all others
```bash
# Behind the scenes, this runs:
mc policy get-json site1/bucket > policy.json
mc policy set-json policy.json site2/bucket
```

#### Sync Lifecycle Policies
Synchronizes ILM rules across sites
```bash
# Example operation:
mc ilm export site1/bucket > lifecycle.json
mc ilm import site2/bucket < lifecycle.json
```

#### Validate Consistency
Checks all configurations and reports differences

#### Health Check
Verifies all sites are online and reachable

## 🔧 API Endpoints

The tool provides REST APIs for integration:

### GET /api/replication/info
Get site replication group information

**Response:**
```json
{
  "enabled": true,
  "sites": [
    {
      "name": "site1",
      "endpoint": "https://minio1.example.com",
      "healthy": true
    },
    {
      "name": "site2",
      "endpoint": "https://minio2.example.com",
      "healthy": true
    }
  ],
  "totalSites": 2,
  "replicationGroup": true
}
```

### GET /api/replication/status
Get replication status for each site

**Response:**
```json
{
  "status": "healthy",
  "sites": {
    "site1": {
      "replicatedBuckets": 10,
      "pendingObjects": 0,
      "failedObjects": 0,
      "lastSyncTime": "2025-10-10T10:00:00Z",
      "healthy": true
    },
    "site2": {
      "replicatedBuckets": 10,
      "pendingObjects": 5,
      "failedObjects": 0,
      "lastSyncTime": "2025-10-10T10:00:30Z",
      "healthy": true
    }
  }
}
```

### GET /api/replication/compare
Compare bucket configurations across sites

**Response:**
```json
{
  "buckets": {
    "my-bucket": {
      "existsOn": ["site1", "site2"],
      "policy": {
        "consistent": true,
        "values": {
          "site1": "download",
          "site2": "download"
        }
      },
      "lifecycle": {
        "consistent": false,
        "values": {
          "site1": {...},
          "site2": {...}
        }
      }
    }
  }
}
```

## 📝 Common Use Cases

### Use Case 1: Adding a New Bucket

When you create a new bucket on one site, ensure it's replicated:

1. Create bucket on site1: `mc mb site1/new-bucket`
2. Open tool → **Consistency Check**
3. Click "Run Check"
4. If bucket exists only on site1, manually create on other sites
5. Use **Sync Bucket Policies** to copy configuration

### Use Case 2: Policy Drift Detection

Detect when bucket policies have diverged:

1. Navigate to **Consistency Check**
2. Click "Run Check"
3. Review buckets marked as "Inconsistent"
4. Click on bucket to see detailed differences
5. Use automated sync to fix

### Use Case 3: Health Monitoring

Monitor overall replication health:

1. Navigate to **Replication Status**
2. Check each site's health status
3. Review pending/failed objects
4. If issues found, investigate specific site

### Use Case 4: Bulk Configuration Sync

After configuration changes on one site:

1. Navigate to **Operations**
2. Select sync operation (Policies or Lifecycle)
3. Click "Execute"
4. Review results
5. Verify with **Consistency Check**

## 🔍 Troubleshooting

### No sites showing up

**Problem:** Dashboard shows "No Sites Configured"

**Solution:**
```bash
# Verify mc is installed
mc --version

# Check configured aliases
mc alias list

# Add your sites
mc alias set site1 https://minio1.example.com ACCESS_KEY SECRET_KEY
```

### Consistency check shows errors

**Problem:** All buckets show as "Inconsistent"

**Possible causes:**
- Sites are not in same replication group
- Manual configuration changes on one site
- Network issues preventing proper sync

**Solution:**
1. Verify sites are properly configured for replication
2. Use automated sync operations to restore consistency
3. Check MinIO server logs for replication errors

### Cannot connect to site

**Problem:** Site shows as "Unhealthy"

**Solution:**
```bash
# Test connectivity
mc admin info site1

# Check if MinIO server is running
curl https://minio1.example.com/minio/health/live

# Verify credentials
mc ls site1
```

## 🎨 UI Features

### Dashboard Layout

- **Left Sidebar** - Navigation menu
- **Main Content** - Page-specific content
- **Header** - Logo, title, language selector, refresh button

### Language Support

The tool supports:
- 🇬🇧 English
- 🇻🇳 Tiếng Việt

Switch language using the dropdown in the header.

### Real-time Updates

- Health indicators update automatically
- Status badges show current state
- Metrics refresh on page navigation

## 🔐 Security Considerations

1. **MC Configuration** - Stores credentials for all sites
   - Keep `~/.mc/config.json` secure
   - Use restrictive file permissions: `chmod 600 ~/.mc/config.json`

2. **Network Access** - Tool needs network access to all sites
   - Ensure firewall rules allow connections
   - Use HTTPS for production deployments

3. **Authentication** - Currently relies on mc credentials
   - Consider adding web UI authentication for production
   - Use read-only credentials where possible

## 📦 Deployment

### Docker Deployment

```bash
# Build image
docker build -t mc-tool .

# Run container
docker run -d \
  -p 8080:8080 \
  -v ~/.mc:/root/.mc:ro \
  mc-tool
```

### Systemd Service

Create `/etc/systemd/system/mc-tool.service`:

```ini
[Unit]
Description=MinIO Site Replication Management Tool
After=network.target

[Service]
Type=simple
User=minio
WorkingDirectory=/opt/mc-tool
ExecStart=/opt/mc-tool/mc-tool web
Restart=always

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable mc-tool
sudo systemctl start mc-tool
```

## 🤝 Contributing

Contributions are welcome! Areas for improvement:

- [ ] IAM policy sync
- [ ] User/Group management
- [ ] Encryption settings comparison
- [ ] Replication metrics & analytics
- [ ] Scheduled sync jobs
- [ ] Email/Slack notifications
- [ ] Multi-language support expansion

## 📄 License

This tool is provided as-is for MinIO operations teams.

## 🔗 Related Resources

- [MinIO Site Replication Docs](https://min.io/docs/minio/linux/operations/install-deploy-manage/multi-site-replication.html)
- [MinIO Client (mc) Guide](https://min.io/docs/minio/linux/reference/minio-mc.html)
- [MinIO Admin API](https://min.io/docs/minio/linux/reference/minio-mc-admin.html)

## 💡 Tips & Best Practices

1. **Regular Consistency Checks** - Run daily to catch drift early
2. **Monitor Health** - Set up monitoring on replication status page
3. **Document Changes** - Keep track of manual configuration changes
4. **Test Sync Operations** - Test on non-production buckets first
5. **Backup Configurations** - Export policies before bulk changes
6. **Use Read Replicas** - For monitoring, use read-only access when possible

---

**Need Help?** Check the [issues page](https://github.com/liamdn8/mc-tool/issues) or create a new issue.
