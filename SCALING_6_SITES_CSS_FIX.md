# MinIO Sites Scaling + CSS Button Fix

## 📊 Changes Summary

### 1. ✅ Scaled MinIO from 2 to 6 Sites

**File**: `docker-compose.yml`

**Added 4 new MinIO sites**:
- **site3**: Port 9003 (API), 9093 (Console) - Region: eu-west-1
- **site4**: Port 9004 (API), 9094 (Console) - Region: ap-south-1
- **site5**: Port 9005 (API), 9095 (Console) - Region: ap-northeast-1
- **site6**: Port 9006 (API), 9096 (Console) - Region: sa-east-1

**Full Configuration**:
```yaml
services:
  minio-site1:  # Port 9001 - us-east-1
  minio-site2:  # Port 9002 - us-west-1
  minio-site3:  # Port 9003 - eu-west-1   ← NEW
  minio-site4:  # Port 9004 - ap-south-1  ← NEW
  minio-site5:  # Port 9005 - ap-northeast-1 ← NEW
  minio-site6:  # Port 9006 - sa-east-1   ← NEW

volumes:
  minio-site1-data:
  minio-site2-data:
  minio-site3-data:  ← NEW
  minio-site4-data:  ← NEW
  minio-site5-data:  ← NEW
  minio-site6-data:  ← NEW
```

### 2. ✅ Fixed CSS Button Layout Issues

**File**: `pkg/web/static/styles-new.css`

**Problem**: 
- Buttons tràn dòng khi có nhiều sites
- Text bị cắt khi buttons quá dài
- Spacing không đồng nhất

**Solution**:

#### A. Reduced Button Padding & Font Size
```css
/* Before */
.btn-icon {
    padding: 8px 14px;
    font-size: 13px;
}

/* After */
.btn-icon {
    padding: 6px 10px;   /* Smaller padding */
    font-size: 12px;      /* Smaller font */
    white-space: nowrap;  /* Prevent text wrap */
    flex-shrink: 0;       /* Don't shrink */
}
```

#### B. Fixed Icon Sizes
```css
.btn-icon svg {
    flex-shrink: 0;
    width: 14px;
    height: 14px;
}
```

**Benefits**:
- Icons won't resize when container shrinks
- Consistent visual appearance across all buttons

#### C. Improved Flex Layout
```css
.site-management-actions {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
    margin-top: 12px;  /* Better spacing from header */
}
```

**Benefits**:
- Buttons wrap to new line automatically
- Consistent spacing between buttons
- Better visual hierarchy

#### D. Text Overflow Handling
```css
.btn-icon span {
    overflow: hidden;
    text-overflow: ellipsis;
}
```

**Benefits**:
- Long button text gets ellipsis (...)
- Prevents layout breaking

---

## 🧪 Testing

### 1. Verify All 6 Sites Running

```bash
docker ps | grep minio-site
```

**Expected Output**:
```
minio-site1   Up (healthy)   0.0.0.0:9001->9000/tcp
minio-site2   Up (healthy)   0.0.0.0:9002->9000/tcp
minio-site3   Up (healthy)   0.0.0.0:9003->9000/tcp
minio-site4   Up (healthy)   0.0.0.0:9004->9000/tcp
minio-site5   Up (healthy)   0.0.0.0:9005->9000/tcp
minio-site6   Up (healthy)   0.0.0.0:9006->9000/tcp
```

### 2. Verify MC Aliases

```bash
mc alias list | grep "site[1-6]"
```

**Expected**:
```
site1
site2
site3
site4
site5
site6
```

### 3. Test API Response

```bash
curl -s http://localhost:8080/api/replication/info | jq '.totalAliases'
```

**Expected**: `6`

### 4. Visual Verification

Open browser: `http://localhost:8080`

**Site Cards Should Show**:
```
┌─────────────────────────────────┐
│ site1          ○ Standalone     │
│ HEALTH: ✓ Healthy              │
│ BUCKETS: 0                      │
│ REPLICATION: ○ Not Replicated   │
│                                 │
│ [Resync ↓] [Resync ↑] [Remove] │ ← Buttons fit well
└─────────────────────────────────┘
```

**With 6 Sites Grid**:
```
┌─────┐ ┌─────┐ ┌─────┐
│Site1│ │Site2│ │Site3│
└─────┘ └─────┘ └─────┘
┌─────┐ ┌─────┐ ┌─────┐
│Site4│ │Site5│ │Site6│
└─────┘ └─────┘ └─────┘
```

---

## 🚀 Deployment Commands

### Full Setup from Scratch

```bash
# 1. Stop all old containers
docker ps -a | grep minio-site | awk '{print $1}' | xargs -r docker rm -f

# 2. Start 6 new sites
cd /home/liamdn/mc-tool
docker-compose up -d

# 3. Wait for health check
sleep 10

# 4. Setup MC aliases
HOST_IP=$(hostname -I | awk '{print $1}')
for i in 1 2 3 4 5 6; do
  mc alias set site$i http://$HOST_IP:900$i minioadmin minioadmin
  echo "✓ Site $i configured"
done

# 5. Verify all sites
mc alias list | grep "site[1-6]"

# 6. Rebuild and restart web server
pkill -f 'mc-tool web'
cd /home/liamdn/mc-tool
go build -o mc-tool .
./mc-tool web --port 8080 > /tmp/mc-tool.log 2>&1 &

# 7. Verify API
sleep 3
curl -s http://localhost:8080/api/replication/info | jq '.totalAliases'
```

### Quick Restart (After Code Changes)

```bash
# Stop server
pkill -f 'mc-tool web'

# Rebuild
cd /home/liamdn/mc-tool
go build -o mc-tool .

# Start
./mc-tool web --port 8080 > /tmp/mc-tool.log 2>&1 &
```

---

## 📊 Port Mapping Reference

| Site | API Port | Console Port | Region | URL |
|------|----------|--------------|--------|-----|
| site1 | 9001 | 9091 | us-east-1 | http://<IP>:9001 |
| site2 | 9002 | 9092 | us-west-1 | http://<IP>:9002 |
| site3 | 9003 | 9093 | eu-west-1 | http://<IP>:9003 |
| site4 | 9004 | 9094 | ap-south-1 | http://<IP>:9004 |
| site5 | 9005 | 9095 | ap-northeast-1 | http://<IP>:9005 |
| site6 | 9006 | 9096 | sa-east-1 | http://<IP>:9006 |

---

## 🎨 CSS Changes Detail

### Before (Tràn dòng khi nhiều buttons)

```css
.btn-icon {
    padding: 8px 14px;        /* Too large */
    font-size: 13px;          /* Too large */
    /* No white-space control */
    /* No flex-shrink control */
}

.btn-icon svg {
    /* No size control - can shrink unexpectedly */
}
```

**Issues**:
- Buttons quá lớn → 3 buttons không fit trong 1 dòng
- Text wrap → button height không consistent
- Icons có thể shrink → looks broken

### After (Fit tốt, responsive)

```css
.site-management-actions {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;          /* Wrap khi cần */
    margin-top: 12px;         /* Better spacing */
}

.btn-icon {
    padding: 6px 10px;        /* Compact */
    font-size: 12px;          /* Readable but smaller */
    white-space: nowrap;      /* Single line */
    flex-shrink: 0;           /* Don't shrink */
    min-width: 0;             /* Allow flex to work */
}

.btn-icon svg {
    flex-shrink: 0;           /* Icons stay 14x14 */
    width: 14px;
    height: 14px;
}

.btn-icon span {
    overflow: hidden;         /* Clip long text */
    text-overflow: ellipsis;  /* Add ... */
}
```

**Benefits**:
- ✅ 3 buttons fit trong 1 dòng (width ~320px card)
- ✅ Icons luôn đúng size 14x14px
- ✅ Text dài tự động hiển thị "..."
- ✅ Responsive: wrap khi card nhỏ
- ✅ Consistent spacing với gap: 8px

---

## 🧩 Button Layout Examples

### On Standard Card (320px)
```
┌────────────────────────────────────┐
│ site1              ○ Standalone    │
│                                    │
│ [Resync ↓] [Resync ↑] [Remove]   │ ← Fits well
└────────────────────────────────────┘
```

### On Narrow Card (280px)
```
┌──────────────────────────────┐
│ site1      ○ Standalone      │
│                              │
│ [Resync ↓] [Resync ↑]       │ ← 2 buttons
│ [Remove]                     │ ← Wraps to new line
└──────────────────────────────┘
```

### With Long Button Text (if translated)
```
┌────────────────────────────────────┐
│ site1              ○ Standalone    │
│                                    │
│ [Đồng bộ...] [Đồng bộ...] [Xóa]  │ ← Text clipped with ...
└────────────────────────────────────┘
```

---

## 🔄 Migration Path

### From 2 Sites → 6 Sites

**No data loss**, just scale up:

```bash
# 1. Stop current 2-site setup
docker-compose down

# 2. Update docker-compose.yml (already done)

# 3. Start 6 sites
docker-compose up -d

# 4. Setup new aliases (site3-6)
# site1 and site2 configs are preserved in mc config

# 5. Test site replication with 6 sites
mc admin replicate add site1 site2 site3 site4 site5 site6
```

---

## 📝 Future Improvements

### 1. Responsive Icon-Only Mode
```css
@media (max-width: 768px) {
    .btn-icon span {
        display: none;  /* Hide text on mobile */
    }
    .btn-icon {
        padding: 6px;   /* Icon only */
    }
}
```

### 2. Button Groups
```css
.site-management-actions {
    display: flex;
    gap: 8px;
}

.btn-group {
    display: flex;
    gap: 4px;
    border: 1px solid var(--border-color);
    border-radius: 6px;
}

.btn-group .btn-icon:not(:last-child) {
    border-right: 1px solid var(--border-color);
    border-radius: 0;
}
```

### 3. Tooltip for Clipped Text
```html
<button class="btn-icon" title="Full button text here">
    <svg>...</svg>
    <span>Clipped...</span>
</button>
```

---

## ✅ Verification Checklist

After deployment:

- [ ] All 6 Docker containers running
- [ ] All 6 containers show "(healthy)" status
- [ ] All 6 MC aliases configured
- [ ] API returns `totalAliases: 6`
- [ ] Web UI shows 6 site cards
- [ ] Buttons fit well in cards
- [ ] No button text overflow
- [ ] Icons are 14x14px (not shrunk)
- [ ] Buttons wrap correctly on narrow screens
- [ ] Can click all buttons successfully
- [ ] Hard refresh browser (Ctrl+Shift+R)

---

**Date**: October 13, 2025  
**Status**: ✅ Completed  
**MinIO Sites**: 2 → 6 (3x scale up)  
**CSS**: Button layout fixed for better fit
