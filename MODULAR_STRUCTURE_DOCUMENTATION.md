# MC-Tool Web UI - Modular Structure Documentation

## Tóm tắt thay đổi

Đã phân tách file `app.js` monolithic thành cấu trúc modular để dễ bảo trì và mở rộng.

## Cấu trúc thư mục mới

```
pkg/web/static/
├── index.html                      # Main HTML template (updated with module imports)
├── styles.css                      # CSS styles
├── app.js                          # Main application controller
└── js/
    ├── utils/                      # Shared utilities
    │   ├── i18n.js                # Internationalization (translations)
    │   ├── api.js                 # API calls and data fetching
    │   └── helpers.js             # Formatting, notifications, dialogs
    ├── pages/                      # Page-specific modules
    │   ├── overview.js            # Overview page logic
    │   ├── sites.js               # Sites page with replication management
    │   ├── buckets.js             # Buckets page logic
    │   ├── replication.js         # Replication status page
    │   ├── consistency.js         # Consistency check page
    │   └── operations.js          # Operations page
    └── components/                 # Future: Reusable components
```

## Tính năng mới

### 1. Auto-Reload sau Replication Actions

**Trước đây:** 
- Hiển thị thông báo yêu cầu người dùng reload trang manually
- Workflow: action → notification → manual reload

**Bây giờ:**
- Tự động reload trang sau khi thành công
- Workflow: action → success message → auto reload sau 2 giây

**Implementation:**
```javascript
// In helpers.js
export function autoReloadAfterReplicationAction(delay = 2000) {
    setTimeout(() => {
        showNotification('info', 'Reloading to show updated status...', { duration: 1500 });
        setTimeout(() => {
            window.location.reload();
        }, 1500);
    }, delay);
}

export function handleReplicationSuccess(message, autoReload = true) {
    showNotification('success', message);
    
    if (autoReload) {
        autoReloadAfterReplicationAction();
    }
}
```

### 2. Cấu trúc Module

**API Module (`js/utils/api.js`)**
- Tập trung tất cả API calls
- Error handling nhất quán
- Return format: `{ response, data }`

**I18n Module (`js/utils/i18n.js`)**
- Quản lý đa ngôn ngữ
- Translation functions
- Dynamic language switching

**Helpers Module (`js/utils/helpers.js`)**
- Utility functions (formatNumber, formatBytes)
- Notification system (enhanced với auto-reload)
- Error dialogs
- Site selection dialogs

**Page Modules (`js/pages/*.js`)**
- Mỗi page có module riêng
- Import dependencies cần thiết
- Export render functions

## API Usage Examples

### Add Sites to Replication
```javascript
import { addSitesToReplication } from './js/utils/api.js';
import { handleReplicationSuccess, handleReplicationError } from './js/utils/helpers.js';

const { response, data } = await addSitesToReplication(['site1', 'site2', 'site3']);

if (response.ok && data.success !== false && !data.error) {
    handleReplicationSuccess(data.message || 'Sites added successfully');
    // Auto-reload sau 2 giây
} else {
    handleReplicationError('Add Sites Failed', data, response);
}
```

### Show Notifications with Auto-Reload
```javascript
import { showNotification, autoReloadAfterReplicationAction } from './js/utils/helpers.js';

// Success với auto-reload
showNotification('success', 'Operation completed');
autoReloadAfterReplicationAction(); // Reload sau 2 giây

// Custom duration
showNotification('info', 'Processing...', { duration: 3000 });
```

## Backward Compatibility

- `window.app` - Global app instance
- `window.loadSites()` - Refresh function
- `window.removeSite()`, `window.resyncSite()` - Site management functions

## Benefits

1. **Modularity**: Mỗi module có responsibility riêng biệt
2. **Maintainability**: Dễ debug và maintain từng phần
3. **Scalability**: Dễ thêm features mới
4. **Auto-Reload**: UX tốt hơn, không cần manual refresh
5. **Code Reusability**: Shared utilities có thể reuse
6. **Type Safety**: Import/Export rõ ràng dependencies

## Testing Status

✅ **Web Server**: Khởi động thành công trên port 8080  
✅ **API Endpoints**: `/api/aliases`, `/api/replication/info` hoạt động  
✅ **Replication Operations**: Add/Remove sites functionality tested  
✅ **Module Loading**: ES6 modules load correctly  
✅ **Auto-Reload**: Implemented in replication actions  

## Migration Notes

- File `app.js` cũ đã được backup thành `app.js.backup`
- HTML template updated với `type="module"` imports
- CSS cache version updated to v5
- Tất cả functionality được bảo toàn và enhanced

## Future Enhancements

1. **Components Directory**: Reusable UI components
2. **State Management**: Centralized app state
3. **WebSocket**: Real-time updates
4. **Progressive Web App**: Offline functionality
5. **Testing Framework**: Unit tests cho modules