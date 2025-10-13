# MinIO Site Replication Test Environment

## Overview
Scripts để setup 2 MinIO servers bằng Docker cho việc test Site Replication.

## Setup Details

### Architecture
```
┌─────────────────┐         ┌─────────────────┐
│  MinIO Site 1   │         │  MinIO Site 2   │
│                 │         │                 │
│  Port: 9001     │◄───────►│  Port: 9002     │
│  Console: 9011  │  Sync   │  Console: 9012  │
│                 │         │                 │
│  Alias: site1   │         │  Alias: site2   │
└─────────────────┘         └─────────────────┘
        │                           │
        └───────────┬───────────────┘
                    │
            Docker Network
         minio-replication-network
```

### Container Configuration

**Site 1:**
- Container: `minio-site1`
- API Port: `9001` (mapped to container port 9000)
- Console Port: `9011` (mapped to container port 9001)
- Endpoint: `http://<HOST_IP>:9001`
- Data: `./test-data/site1`

**Site 2:**
- Container: `minio-site2`
- API Port: `9002` (mapped to container port 9000)
- Console Port: `9012` (mapped to container port 9001)
- Endpoint: `http://<HOST_IP>:9002`
- Data: `./test-data/site2`

**Credentials (both sites):**
- Username: `minioadmin`
- Password: `minioadmin123`

### Key Features

✅ **Network Accessible Endpoints**
- Uses real host IP address instead of localhost
- Sites can communicate with each other
- Ready for site replication

✅ **Proper MINIO_SERVER_URL**
- Each server has MINIO_SERVER_URL configured
- Required for site replication to work

✅ **Docker Network**
- Both containers on same Docker network
- Can resolve each other by container name

✅ **Persistent Data**
- Data stored in `./test-data/site1` and `./test-data/site2`
- Survives container restart (unless cleaned up)

## Quick Start

### 1. Setup Environment
```bash
# Make scripts executable
chmod +x setup-minio-test.sh cleanup-minio-test.sh

# Run setup script
./setup-minio-test.sh
```

The script will:
1. Detect your host IP address
2. Create Docker network
3. Start 2 MinIO containers
4. Configure mc aliases with accessible endpoints
5. Verify connectivity

### 2. Start mc-tool
```bash
# Build if needed
go build -o mc-tool .

# Start web interface
./mc-tool web --port 8080
```

### 3. Test Site Replication

**Option A: Via Web Interface**
1. Open http://localhost:8080
2. Go to "Sites" page
3. Select both `site1` and `site2` checkboxes
4. Click "Add Sites to Replication"
5. Verify success message

**Option B: Via Command Line**
```bash
# Add sites to replication
mc admin replicate add site1 site2

# Verify replication status
mc admin replicate info site1
mc admin replicate info site2
```

### 4. Test Replication

**Create bucket on site1:**
```bash
mc mb site1/test-bucket
mc cp some-file.txt site1/test-bucket/
```

**Verify bucket appears on site2:**
```bash
mc ls site2
mc ls site2/test-bucket
```

Both sites should show the same buckets and objects!

### 5. Test Remove & Resync

**Via Web Interface:**
- Click "Remove" button on a site
- Click "Resync From" or "Resync To" icons
- Select target site in modal dialog

**Via Command Line:**
```bash
# Remove a site from replication
mc admin replicate rm site1 --force

# Resync from site2 to site1
mc admin replicate resync start site1 --deployment-id <site2-deployment-id>
```

## Cleanup

```bash
# Stop containers and cleanup
./cleanup-minio-test.sh
```

This will:
1. Stop both MinIO containers
2. Remove containers
3. Remove Docker network
4. Optionally delete test data

## Verification Commands

### Check Container Status
```bash
# View running containers
docker ps

# Check site1 logs
docker logs -f minio-site1

# Check site2 logs
docker logs -f minio-site2
```

### Check MinIO Status
```bash
# List aliases
mc alias list

# Check site1 info
mc admin info site1

# Check site2 info
mc admin info site2

# Check replication status
mc admin replicate info site1
mc admin replicate info site2
```

### Check Connectivity
```bash
# Ping site1
curl http://<HOST_IP>:9001/minio/health/live

# Ping site2
curl http://<HOST_IP>:9002/minio/health/live
```

## Common Issues & Solutions

### Issue: "localhost not available"
**Cause:** MINIO_SERVER_URL not set or using localhost
**Solution:** The script automatically uses host IP address

### Issue: Containers not communicating
**Cause:** Firewall blocking ports
**Solution:** 
```bash
# Check if ports are listening
sudo netstat -tlnp | grep -E '9001|9002'

# On some Linux systems, allow ports
sudo ufw allow 9001
sudo ufw allow 9002
```

### Issue: Permission denied on data directories
**Solution:**
```bash
# Fix permissions
sudo chown -R $USER:$USER ./test-data
chmod -R 755 ./test-data
```

### Issue: Port already in use
**Solution:**
```bash
# Find process using port
sudo lsof -i :9001

# Stop existing MinIO
./cleanup-minio-test.sh
```

## Advanced Configuration

### Change Ports
Edit `setup-minio-test.sh`:
```bash
SITE1_PORT="9001"      # Change to desired port
SITE2_PORT="9002"      # Change to desired port
```

### Change Credentials
Edit `setup-minio-test.sh`:
```bash
ROOT_USER="myadmin"
ROOT_PASSWORD="mypassword123"
```

### Add More Sites
Duplicate the site2 block in `setup-minio-test.sh`:
```bash
SITE3_NAME="minio-site3"
SITE3_PORT="9003"
SITE3_CONSOLE_PORT="9013"

docker run -d \
  --name $SITE3_NAME \
  --network $NETWORK_NAME \
  -p $SITE3_PORT:9000 \
  -p $SITE3_CONSOLE_PORT:9001 \
  -e "MINIO_ROOT_USER=$ROOT_USER" \
  -e "MINIO_ROOT_PASSWORD=$ROOT_PASSWORD" \
  -e "MINIO_SERVER_URL=http://${HOST_IP}:${SITE3_PORT}" \
  -v $(pwd)/test-data/site3:/data \
  minio/minio server /data --console-address ":9001"

mc alias set site3 http://${HOST_IP}:${SITE3_PORT} $ROOT_USER $ROOT_PASSWORD
```

## Testing Checklist

- [ ] Both containers running: `docker ps`
- [ ] Both sites accessible: `mc admin info site1`, `mc admin info site2`
- [ ] Endpoints use IP not localhost: `mc alias list`
- [ ] Can add sites to replication: `mc admin replicate add site1 site2`
- [ ] Buckets sync between sites: Create bucket on site1, check site2
- [ ] Objects sync between sites: Upload to site1, verify on site2
- [ ] Can remove site: `mc admin replicate rm site1`
- [ ] Can resync: `mc admin replicate resync start`
- [ ] Web interface shows both sites
- [ ] Can add via web interface
- [ ] Can remove via web interface
- [ ] Can resync via web interface

## Notes

- **Network:** Containers use bridge network for isolation
- **Data:** Test data stored locally, not in containers
- **Performance:** This is a test setup, not for production
- **Security:** Default credentials, change for production use
- **Scalability:** Can add more sites by duplicating configuration

## Resources

- [MinIO Site Replication Guide](https://min.io/docs/minio/linux/operations/install-deploy-manage/multi-site-replication.html)
- [Docker MinIO Documentation](https://min.io/docs/minio/container/index.html)
- [mc Admin Replicate](https://min.io/docs/minio/linux/reference/minio-mc-admin/mc-admin-replicate.html)
