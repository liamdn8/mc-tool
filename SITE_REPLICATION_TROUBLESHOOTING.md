# Site Replication Troubleshooting Guide

## Common Issue: Localhost Endpoint Error

### Error Message
```
❌ Unable to add sites for replication. 
Error received when contacting a peer site (unable to create admin client for site1: 
Remote service endpoint http://localhost:9001 not available
remote target is offline for endpoint http://localhost:9001).
```

### Root Cause
MinIO Site Replication requires all sites to be able to connect to each other directly using their configured endpoints. When sites are configured with `localhost` or `127.0.0.1` endpoints:

1. **Site A** (`http://localhost:9001`) tries to tell **Site B** to connect back
2. **Site B** receives the instruction to connect to `http://localhost:9001`
3. **Site B** tries to connect to **its own** localhost:9001, not Site A's
4. Connection fails because each server's "localhost" points to itself

```
┌─────────────────┐                    ┌─────────────────┐
│   Site A        │                    │   Site B        │
│  localhost:9001 │ ───────X──────────>│  localhost:9002 │
│                 │  Can't connect!    │                 │
│                 │<──────X────────────│                 │
└─────────────────┘                    └─────────────────┘
```

---

## Solution 1: Use Accessible IP Addresses

### Step-by-Step Fix

#### 1. Check Current IP Address

```bash
# Linux/Mac
ip addr show | grep "inet "
# or
hostname -I

# Windows
ipconfig
```

Example output:
```
192.168.1.100  # Use this IP
```

#### 2. Reconfigure MinIO Servers

**Site 1:** (on 192.168.1.100)
```bash
# Stop MinIO if running
killall minio

# Set server URL environment variable
export MINIO_SERVER_URL="http://192.168.1.100:9000"

# Start MinIO
minio server /data --console-address ":9001"
```

**Site 2:** (on 192.168.1.101)
```bash
# Stop MinIO if running
killall minio

# Set server URL environment variable
export MINIO_SERVER_URL="http://192.168.1.101:9000"

# Start MinIO
minio server /data --console-address ":9001"
```

#### 3. Update mc Aliases

```bash
# Remove old aliases
mc alias remove site1
mc alias remove site2

# Add new aliases with accessible IPs
mc alias set site1 http://192.168.1.100:9000 minioadmin minioadmin
mc alias set site2 http://192.168.1.101:9000 minioadmin minioadmin

# Verify connectivity
mc admin info site1
mc admin info site2
```

#### 4. Setup Site Replication

Now you can use the web interface to add sites to replication, or use command line:

```bash
mc admin replicate add site1 site2
```

---

## Solution 2: Use Docker with Host Network

If running in Docker, use host network mode:

### Docker Compose Example

**Site 1:** (docker-compose-site1.yml)
```yaml
version: '3.8'
services:
  minio:
    image: minio/minio
    network_mode: host
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
      MINIO_SERVER_URL: http://192.168.1.100:9000
    command: server /data --console-address ":9001"
    volumes:
      - /mnt/data1:/data
```

**Site 2:** (docker-compose-site2.yml)
```yaml
version: '3.8'
services:
  minio:
    image: minio/minio
    network_mode: host
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
      MINIO_SERVER_URL: http://192.168.1.101:9000
    command: server /data --console-address ":9001"
    volumes:
      - /mnt/data2:/data
```

---

## Solution 3: Use Domain Names

### Setup with DNS/Hosts

#### 1. Edit /etc/hosts (or C:\Windows\System32\drivers\etc\hosts)

```
192.168.1.100  minio-site1.local
192.168.1.101  minio-site2.local
```

#### 2. Configure MinIO with Domain Names

**Site 1:**
```bash
export MINIO_SERVER_URL="http://minio-site1.local:9000"
minio server /data --console-address ":9001"
```

**Site 2:**
```bash
export MINIO_SERVER_URL="http://minio-site2.local:9000"
minio server /data --console-address ":9001"
```

#### 3. Update mc Aliases

```bash
mc alias set site1 http://minio-site1.local:9000 minioadmin minioadmin
mc alias set site2 http://minio-site2.local:9000 minioadmin minioadmin
```

---

## Verification Checklist

Before attempting to setup site replication, verify:

### ✅ Network Connectivity

```bash
# From Site 1, ping Site 2
ping 192.168.1.101

# From Site 2, ping Site 1
ping 192.168.1.100
```

### ✅ Port Accessibility

```bash
# From Site 1, check if Site 2's port is open
nc -zv 192.168.1.101 9000

# From Site 2, check if Site 1's port is open
nc -zv 192.168.1.100 9000
```

Or use telnet:
```bash
telnet 192.168.1.101 9000
```

### ✅ MinIO Server Configuration

```bash
# Check Site 1
mc admin info site1 --json | jq .info.servers

# Check Site 2
mc admin info site2 --json | jq .info.servers
```

The endpoints should show accessible IPs, **not localhost**.

### ✅ Firewall Rules

**Linux (ufw):**
```bash
sudo ufw allow 9000/tcp
sudo ufw allow 9001/tcp
```

**Linux (firewalld):**
```bash
sudo firewall-cmd --permanent --add-port=9000/tcp
sudo firewall-cmd --permanent --add-port=9001/tcp
sudo firewall-cmd --reload
```

**Windows:**
```powershell
New-NetFirewallRule -DisplayName "MinIO API" -Direction Inbound -LocalPort 9000 -Protocol TCP -Action Allow
New-NetFirewallRule -DisplayName "MinIO Console" -Direction Inbound -LocalPort 9001 -Protocol TCP -Action Allow
```

---

## Common Scenarios

### Scenario 1: All Sites on Same Machine (Testing)

For testing on a single machine, use different ports:

```bash
# Site 1
export MINIO_SERVER_URL="http://127.0.0.1:9000"
minio server /data1 --address ":9000" --console-address ":9001"

# Site 2
export MINIO_SERVER_URL="http://127.0.0.1:9002"
minio server /data2 --address ":9002" --console-address ":9003"

# Configure aliases
mc alias set site1 http://127.0.0.1:9000 minioadmin minioadmin
mc alias set site2 http://127.0.0.1:9002 minioadmin minioadmin
```

**Note:** This works because both sites see the same "127.0.0.1" address.

### Scenario 2: Sites Across Different Networks

For sites across internet or different networks:

1. Use public IPs or domain names
2. Ensure port forwarding is configured
3. Consider VPN or secure tunnel
4. Use HTTPS (recommended for production)

**Example with public domains:**
```bash
mc alias set site1 https://minio1.example.com minioadmin minioadmin
mc alias set site2 https://minio2.example.com minioadmin minioadmin
```

### Scenario 3: Sites in Kubernetes

Use Kubernetes Services with LoadBalancer or NodePort:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: minio-site1
spec:
  type: LoadBalancer
  ports:
    - port: 9000
      targetPort: 9000
      name: api
  selector:
    app: minio-site1
```

Then use the LoadBalancer IP or external domain.

---

## Testing Site Replication

After configuration, test the setup:

### 1. Check Replication Info

```bash
mc admin replicate info site1
```

Expected output:
```json
{
  "enabled": true,
  "sites": [
    {
      "name": "site1",
      "endpoint": "http://192.168.1.100:9000",
      "deploymentID": "..."
    },
    {
      "name": "site2", 
      "endpoint": "http://192.168.1.101:9000",
      "deploymentID": "..."
    }
  ]
}
```

### 2. Create Test Bucket

```bash
mc mb site1/test-replication
mc ls site2  # Should show test-replication bucket
```

### 3. Upload Test Object

```bash
echo "test" > test.txt
mc cp test.txt site1/test-replication/
mc ls site2/test-replication/  # Should show test.txt
```

---

## Advanced Configuration

### TLS/HTTPS Setup

For production, use HTTPS:

```bash
# Generate certificates (or use Let's Encrypt)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/minio/certs/private.key \
  -out /etc/minio/certs/public.crt

# Start MinIO with HTTPS
export MINIO_SERVER_URL="https://192.168.1.100:9000"
minio server /data --certs-dir /etc/minio/certs
```

Update mc alias:
```bash
mc alias set site1 https://192.168.1.100:9000 minioadmin minioadmin
```

### Load Balancer Configuration

When using load balancer, ensure:
1. Sticky sessions enabled
2. Health check configured: `GET /minio/health/live`
3. Timeout settings appropriate (300s+)

---

## FAQ

**Q: Can I change endpoints after replication is setup?**
A: No, you must remove and re-add sites with new endpoints.

**Q: Do all sites need the same credentials?**
A: No, each site can have different root credentials.

**Q: Can I mix localhost and IP addresses?**
A: No, all sites must use mutually accessible endpoints.

**Q: What if sites are in different availability zones?**
A: Use VPN or dedicated network connection for better performance and security.

**Q: Can I use IPv6 addresses?**
A: Yes, but ensure all sites support IPv6.

---

## Error Messages Reference

| Error | Cause | Solution |
|-------|-------|----------|
| `Remote service endpoint http://localhost:9001 not available` | Localhost endpoint used | Use accessible IP/domain |
| `connection refused` | Port not open or service down | Check firewall and service status |
| `no route to host` | Network unreachable | Check network connectivity |
| `timeout` | Slow network or firewall blocking | Check network and increase timeout |
| `certificate verify failed` | SSL/TLS certificate issue | Update certificates or skip verify (testing only) |

---

## Getting Help

If you still face issues:

1. Check MinIO logs: `journalctl -u minio -f`
2. Enable debug mode: `export MINIO_DEBUG=on`
3. Check network: `tcpdump -i any port 9000`
4. Review MinIO documentation: https://min.io/docs/minio/linux/operations/install-deploy-manage/multi-site-replication.html

---

## Production Best Practices

1. ✅ Use DNS names instead of IPs
2. ✅ Enable HTTPS/TLS
3. ✅ Use dedicated network for replication traffic
4. ✅ Monitor replication lag
5. ✅ Regular backup and DR testing
6. ✅ Document your network topology
7. ✅ Use consistent MinIO versions across sites
8. ✅ Set up alerts for replication failures

---

**Last Updated:** October 13, 2025  
**mc-tool Version:** 1.0.0
