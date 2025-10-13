# MinIO Site Replication Test Environment

This setup creates 2 MinIO servers for testing site replication features.

## Quick Start

### 1. Start Test Environment

```bash
./setup-test-env.sh
```

This will:
- Start 2 MinIO servers in Docker containers
- Configure mc aliases (site1, site2)
- Create test buckets on both sites
- Add sample data

### 2. Start MC-Tool Web UI

```bash
./mc-tool-new web
```

Then open http://localhost:8080 in your browser.

## MinIO Servers

### Site 1
- **API Endpoint**: http://localhost:9001
- **Console**: http://localhost:9091
- **Buckets**:
  - test-bucket-1 (with file1.txt)
  - test-bucket-2 (with file2.txt)
  - shared-bucket (with shared.txt)

### Site 2
- **API Endpoint**: http://localhost:9002
- **Console**: http://localhost:9092
- **Buckets**:
  - test-bucket-3 (with file3.txt)
  - shared-bucket (with shared2.txt)

## Credentials

Both servers use the same credentials:
- **Username**: minioadmin
- **Password**: minioadmin

## MC Aliases

```bash
# List all aliases
mc alias list

# List buckets on site1
mc ls site1

# List buckets on site2
mc ls site2

# List objects in a bucket
mc ls site1/test-bucket-1
```

## Docker Commands

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f

# Stop and remove volumes (clean slate)
docker-compose down -v
```

## Testing Site Replication Features

### 1. Overview Page
- View total sites (should show 7: gcs, local, play, playground, s3, site1, site2)
- View synced buckets
- View total objects across all sites

### 2. Sites Page
- See list of configured sites
- View bucket count per site
- View object count per site
- Check health status

### 3. Buckets Page
- View all buckets across sites
- See replication status (which buckets exist on which sites)
- "shared-bucket" should appear on both site1 and site2

### 4. Replication Status
- Monitor sync status between sites
- Check replication lag
- View recent sync operations

### 5. Consistency Check
- Compare bucket configurations across sites
- Check if policies are consistent
- Verify lifecycle rules match

### 6. Automated Operations
- **Sync Bucket Policies**: Copy bucket policies from one site to all others
- **Sync Lifecycle Policies**: Copy ILM policies across sites
- **Validate Consistency**: Check all sites have matching configurations
- **Health Check**: Verify all sites are accessible and healthy

## Cleanup

To stop and remove everything:

```bash
docker-compose down -v
```

This will:
- Stop both MinIO containers
- Remove the containers
- Delete the data volumes

## Troubleshooting

### Port Already in Use

If ports 9001, 9002, 9091, or 9092 are already in use, you can modify `docker-compose.yml` to use different ports.

### Cannot Connect to MinIO

Check if containers are running:
```bash
docker ps
```

Check container logs:
```bash
docker-compose logs minio-site1
docker-compose logs minio-site2
```

### MC Alias Not Found

Re-run the setup script:
```bash
./setup-test-env.sh
```

Or manually add aliases:
```bash
mc alias set site1 http://localhost:9001 minioadmin minioadmin
mc alias set site2 http://localhost:9002 minioadmin minioadmin
```
