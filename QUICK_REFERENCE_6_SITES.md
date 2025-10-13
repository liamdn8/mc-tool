# Quick Reference - 6 MinIO Sites Setup

## 🚀 Quick Start

### Start All Sites
```bash
cd /home/liamdn/mc-tool
docker-compose up -d
```

### Stop All Sites
```bash
docker-compose down
```

### Check Status
```bash
docker ps | grep minio-site
```

---

## 📍 Site URLs

### API Endpoints
```
Site 1: http://172.31.85.74:9001
Site 2: http://172.31.85.74:9002
Site 3: http://172.31.85.74:9003
Site 4: http://172.31.85.74:9004
Site 5: http://172.31.85.74:9005
Site 6: http://172.31.85.74:9006
```

### Console URLs
```
Site 1 Console: http://172.31.85.74:9091
Site 2 Console: http://172.31.85.74:9092
Site 3 Console: http://172.31.85.74:9093
Site 4 Console: http://172.31.85.74:9094
Site 5 Console: http://172.31.85.74:9095
Site 6 Console: http://172.31.85.74:9096
```

**Credentials**: `minioadmin` / `minioadmin`

---

## 🌐 Web UI

**Management Tool**: http://localhost:8080

**Features**:
- Overview Dashboard with 6 sites
- Health monitoring
- Bucket management
- Site replication setup
- Resync operations

---

## 🛠️ Common Commands

### Setup Site Replication (All 6 Sites)
```bash
mc admin replicate add site1 site2 site3 site4 site5 site6
```

### Check Replication Status
```bash
mc admin replicate info site1
```

### Resync from Site1 to Site2
```bash
# Get deployment ID
DEPLOY_ID=$(mc admin replicate info site2 --json | jq -r '.sites[0].deploymentID')

# Resync
mc admin replicate resync start --deployment-id $DEPLOY_ID site1
```

### Remove Replication
```bash
mc admin replicate rm site1 --all --force
```

---

## 📊 Testing Setup

### Create Test Bucket on Site1
```bash
mc mb site1/test-bucket
mc cp /path/to/file site1/test-bucket/
```

### Verify Replication to Other Sites
```bash
mc ls site2/test-bucket/
mc ls site3/test-bucket/
mc ls site4/test-bucket/
mc ls site5/test-bucket/
mc ls site6/test-bucket/
```

---

## 🔧 Troubleshooting

### Restart Specific Site
```bash
docker restart minio-site3
```

### View Site Logs
```bash
docker logs -f minio-site3
```

### Check Health
```bash
mc admin info site3
```

### Rebuild MC Aliases
```bash
HOST_IP=$(hostname -I | awk '{print $1}')
for i in 1 2 3 4 5 6; do
  mc alias set site$i http://$HOST_IP:900$i minioadmin minioadmin
done
```

---

## 📁 Data Volumes

**Location**: Docker named volumes

```bash
# List volumes
docker volume ls | grep minio-site

# Inspect volume
docker volume inspect mc-tool_minio-site1-data

# Cleanup (⚠️ DELETES DATA)
docker-compose down -v
```

---

## 🎯 Replication Scenarios

### 2-Site Cluster
```bash
mc admin replicate add site1 site2
```

### 3-Site Cluster
```bash
mc admin replicate add site1 site2 site3
```

### 6-Site Cluster (Full)
```bash
mc admin replicate add site1 site2 site3 site4 site5 site6
```

### Mixed Setup
```bash
# Cluster 1: site1, site2, site3
mc admin replicate add site1 site2 site3

# Cluster 2: site4, site5, site6
mc admin replicate add site4 site5 site6
```

---

## ⚡ Performance Notes

- Each site runs on single disk (test setup)
- Production: Use distributed MinIO (multiple disks)
- Network: All sites on same bridge network (low latency)
- Resources: 6 containers × ~200MB RAM each

---

## 📝 Files

### Main Configuration
- `docker-compose.yml` - 6 MinIO services
- `pkg/web/static/styles-new.css` - UI styles (v=3)
- `pkg/web/static/app-new.js` - UI logic (v=3)

### Documentation
- `SCALING_6_SITES_CSS_FIX.md` - This update details
- `HEALTH_STATUS_FIX.md` - Health status improvements
- `FINAL_FIX_CACHE_BUSTING.md` - Cache control

---

**Quick Test**:
```bash
# 1. Start sites
docker-compose up -d && sleep 10

# 2. Check all healthy
docker ps | grep healthy | wc -l
# Should show: 6

# 3. Open web UI
# http://localhost:8080

# 4. See all 6 sites with green health badges
```

**Version**: v=3 (CSS & JS)  
**Date**: October 13, 2025
