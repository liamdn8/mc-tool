# UI/UX Requirements: MinIO Site Replication Management Interface

## Overview

This document outlines the user interface and user experience requirements for the MinIO Site Replication Management console, focusing on the Lucid Icons integration and responsive design principles.

## 1. Design System

### 1.1 Lucid Icons Integration ⭐ **Core Feature**

#### 1.1.1 Implementation Strategy
```html
<!-- CDN Integration -->
<script src="https://unpkg.com/lucide@latest/dist/umd/lucide.js"></script>

<!-- Icon Usage Pattern -->
<i data-lucide="icon-name" width="size" height="size"></i>

<!-- Dynamic Initialization -->
<script>
if (typeof lucide !== 'undefined') {
    lucide.createIcons();
}
</script>
```

#### 1.1.2 Icon Mapping & Standards

**Navigation Icons** (20x20px):
- `layout-dashboard` - Overview/Dashboard page
- `globe` - Sites management page  
- `folder` - Buckets page
- `repeat` - Replication status page
- `check-circle` - Consistency check page
- `settings` - Operations/Settings page

**Header Icons**:
- `package` (32x32px) - Application logo
- `refresh-cw` (20x20px) - Refresh button

**Action Icons** (16x16px):
- `plus` - Add site/Create actions
- `download` - Resync from (pull data)
- `upload` - Resync to (push data)
- `trash-2` - Remove/Delete actions

**Status Icons** (16x16px):
- `check-circle` - Healthy/Success status
- `alert-circle` - Warning status
- `x-circle` - Error/Failed status
- `clock` - Pending/In-progress status

#### 1.1.3 Dynamic Icon Initialization
```javascript
// After dynamic content updates
function refreshIcons() {
    if (typeof lucide !== 'undefined') {
        lucide.createIcons();
    }
}

// Call after AJAX responses
fetch('/api/replication/info')
    .then(response => response.json())
    .then(data => {
        updateSitesContent(data);
        refreshIcons(); // Re-initialize icons for new content
    });
```

### 1.2 Color Palette

#### 1.2.1 Primary Colors
- **Primary Blue**: `#2563eb` - Main actions, links
- **Secondary Gray**: `#6b7280` - Secondary text, borders
- **Background**: `#f9fafb` - Page background
- **Card Background**: `#ffffff` - Content containers

#### 1.2.2 Status Colors
- **Success Green**: `#10b981` - Healthy status, success states
- **Warning Orange**: `#f59e0b` - Warning states, pending
- **Error Red**: `#ef4444` - Error states, failed operations
- **Info Blue**: `#3b82f6` - Information, neutral states

### 1.3 Typography

#### 1.3.1 Font System
- **Primary Font**: `system-ui, -apple-system, sans-serif`
- **Monospace**: `ui-monospace, 'Cascadia Code', 'Source Code Pro', monospace`

#### 1.3.2 Text Hierarchy
- **H1**: 28px, font-weight 700 - Page titles
- **H2**: 24px, font-weight 600 - Section headers  
- **H3**: 20px, font-weight 600 - Card headers
- **Body**: 14px, font-weight 400 - Regular text
- **Small**: 12px, font-weight 400 - Metadata, captions

## 2. Layout Structure

### 2.1 Application Shell

```
┌─────────────────────────────────────────────────────────┐
│ Header (app-header)                                     │
│ ┌─────────────────┐           ┌──────────────────────┐ │
│ │ Logo + Title    │           │ Language + Refresh   │ │
│ └─────────────────┘           └──────────────────────┘ │
├─────────────────────────────────────────────────────────┤
│ ┌─────────────┐ ┌─────────────────────────────────────┐ │
│ │             │ │                                     │ │
│ │ Sidebar     │ │ Main Content                        │ │
│ │ (app-sidebar)│ │ (app-main)                         │ │
│ │             │ │                                     │ │
│ │ Navigation  │ │ Dynamic Page Content                │ │
│ │ + Status    │ │                                     │ │
│ │             │ │                                     │ │
│ └─────────────┘ └─────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### 2.2 Header Component

#### 2.2.1 Structure
```html
<header class="app-header">
    <div class="header-left">
        <i data-lucide="package" class="app-logo" width="32" height="32"></i>
        <div class="app-title">
            <h1>MinIO Site Replication</h1>
            <span class="app-subtitle">Management Console</span>
        </div>
    </div>
    <div class="header-right">
        <select id="languageSelector" class="language-selector">
            <option value="en">🇬🇧 English</option>
            <option value="vi">🇻🇳 Tiếng Việt</option>
        </select>
        <button id="refreshBtn" class="btn-icon" title="Refresh">
            <i data-lucide="refresh-cw" width="20" height="20"></i>
        </button>
    </div>
</header>
```

#### 2.2.2 Styling Requirements
- **Height**: 64px fixed
- **Background**: White with bottom border
- **Responsive**: Collapse title on mobile (<768px)
- **Alignment**: Logo/title left, controls right

### 2.3 Sidebar Navigation

#### 2.3.1 Structure
```html
<aside class="app-sidebar">
    <nav class="sidebar-nav">
        <a href="#" class="nav-link active" data-page="overview">
            <i data-lucide="layout-dashboard" width="20" height="20"></i>
            <span data-i18n="overview">Overview</span>
        </a>
        <!-- Additional nav items -->
    </nav>
    
    <div class="sidebar-footer">
        <div class="mc-status">
            <span class="status-indicator" id="mcStatusIndicator"></span>
            <span id="mcStatusText">MC Ready</span>
        </div>
    </div>
</aside>
```

#### 2.3.2 Responsive Behavior
- **Desktop**: 240px width, always visible
- **Tablet** (768px-1024px): 200px width, collapsible
- **Mobile** (<768px): Overlay, hidden by default

#### 2.3.3 Active State
- **Visual**: Left border + background color change
- **Icon**: Same color as text
- **Animation**: Smooth transition (200ms)

### 2.4 Main Content Area

#### 2.4.1 Page Header Pattern
```html
<div class="page-header">
    <h2 data-i18n="page_title">Page Title</h2>
    <div class="page-actions">
        <button class="btn-primary" id="primaryAction">
            <i data-lucide="plus" width="16" height="16"></i>
            <span data-i18n="action">Action</span>
        </button>
    </div>
</div>
```

#### 2.4.2 Content Grid System
```css
.content-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 1.5rem;
    margin-top: 1.5rem;
}
```

## 3. Component Library

### 3.1 Cards

#### 3.1.1 Info Card
```html
<div class="info-card">
    <div class="info-card-header">
        <h3 data-i18n="card_title">Card Title</h3>
        <span class="badge badge-success">Status</span>
    </div>
    <div class="info-card-body">
        <!-- Card content -->
    </div>
</div>
```

#### 3.1.2 Site Card
```html
<div class="site-card" data-alias="site1">
    <div class="site-header">
        <div class="site-info">
            <h4 class="site-name">Site 1</h4>
            <span class="site-url">http://192.168.1.10:9000</span>
        </div>
        <div class="site-status">
            <i data-lucide="check-circle" width="16" height="16" class="status-healthy"></i>
        </div>
    </div>
    <div class="site-actions">
        <button class="btn-action" data-action="resync-from" title="Resync From">
            <i data-lucide="download" width="16" height="16"></i>
        </button>
        <button class="btn-action" data-action="resync-to" title="Resync To">
            <i data-lucide="upload" width="16" height="16"></i>
        </button>
        <button class="btn-action btn-danger" data-action="remove" title="Remove">
            <i data-lucide="trash-2" width="16" height="16"></i>
        </button>
    </div>
</div>
```

### 3.2 Buttons

#### 3.2.1 Button Variants
```html
<!-- Primary Button -->
<button class="btn-primary">
    <i data-lucide="plus" width="16" height="16"></i>
    <span>Add Site</span>
</button>

<!-- Secondary Button -->
<button class="btn-secondary">
    <i data-lucide="refresh-cw" width="16" height="16"></i>
    <span>Refresh</span>
</button>

<!-- Icon Only Button -->
<button class="btn-icon" title="Remove">
    <i data-lucide="trash-2" width="16" height="16"></i>
</button>

<!-- Danger Button -->
<button class="btn-danger">
    <i data-lucide="trash-2" width="16" height="16"></i>
    <span>Remove</span>
</button>
```

#### 3.2.2 Button States
- **Default**: Normal appearance
- **Hover**: Slightly darker background
- **Active**: Pressed state visual feedback
- **Disabled**: Reduced opacity, no interactions
- **Loading**: Show spinner, disable interactions

### 3.3 Status Indicators

#### 3.3.1 Badge Components
```html
<span class="badge badge-success">Healthy</span>
<span class="badge badge-warning">Warning</span>
<span class="badge badge-error">Error</span>
<span class="badge badge-info">Info</span>
```

#### 3.3.2 Status Dots
```html
<div class="status-indicator status-healthy"></div>
<div class="status-indicator status-warning"></div>
<div class="status-indicator status-error"></div>
<div class="status-indicator status-info"></div>
```

## 4. Page-Specific Requirements

### 4.1 Overview Page

#### 4.1.1 Replication Group Summary
- **Group Status**: Active/Inactive badge
- **Sites Count**: Number with icon
- **Last Updated**: Timestamp with relative time
- **Quick Actions**: Add Site, Refresh buttons

#### 4.1.2 Stats Grid
```html
<div class="stats-grid">
    <div class="stat-item">
        <i data-lucide="globe" width="24" height="24"></i>
        <div class="stat-value">4</div>
        <div class="stat-label">Sites</div>
    </div>
    <div class="stat-item">
        <i data-lucide="folder" width="24" height="24"></i>
        <div class="stat-value">12</div>
        <div class="stat-label">Buckets</div>
    </div>
    <!-- Additional stats -->
</div>
```

### 4.2 Sites Management Page

#### 4.2.1 Sites Grid Layout
- **Responsive Grid**: 1-4 columns based on screen size
- **Site Cards**: Consistent height and spacing
- **Actions**: Hover reveal for secondary actions
- **Empty State**: Friendly message with add site CTA

#### 4.2.2 Add Site Modal
```html
<div class="modal" id="addSiteModal">
    <div class="modal-content">
        <div class="modal-header">
            <h3>Add Sites to Replication</h3>
            <button class="btn-close">
                <i data-lucide="x" width="20" height="20"></i>
            </button>
        </div>
        <div class="modal-body">
            <!-- Site selection interface -->
        </div>
        <div class="modal-footer">
            <button class="btn-secondary" data-action="cancel">Cancel</button>
            <button class="btn-primary" data-action="add">
                <i data-lucide="plus" width="16" height="16"></i>
                Add Sites
            </button>
        </div>
    </div>
</div>
```

### 4.3 Replication Status Page

#### 4.3.1 Status Table
```html
<div class="status-table">
    <div class="table-header">
        <div class="th">Site</div>
        <div class="th">Buckets</div>
        <div class="th">Objects</div>
        <div class="th">Status</div>
        <div class="th">Last Sync</div>
    </div>
    <div class="table-body">
        <div class="tr">
            <div class="td">
                <i data-lucide="globe" width="16" height="16"></i>
                <span>Site 1</span>
            </div>
            <!-- Additional columns -->
        </div>
    </div>
</div>
```

## 5. Responsive Design

### 5.1 Breakpoints
- **Mobile**: 0-767px
- **Tablet**: 768px-1023px  
- **Desktop**: 1024px+

### 5.2 Mobile Adaptations

#### 5.2.1 Header
- Logo/title stack vertically
- Language selector becomes dropdown
- Refresh button remains visible

#### 5.2.2 Navigation
- Sidebar becomes overlay
- Hamburger menu button in header
- Touch-friendly tap targets (44px minimum)

#### 5.2.3 Content
- Single column layout
- Cards stack vertically
- Horizontal scroll for tables
- Bottom sheet for actions

### 5.3 Touch Interactions
- **Tap Targets**: Minimum 44px x 44px
- **Hover States**: Convert to touch/tap states
- **Swipe Gestures**: For dismissing modals/notifications
- **Pull to Refresh**: For updating data

## 6. Accessibility

### 6.1 Keyboard Navigation
- **Tab Order**: Logical progression through interface
- **Focus Indicators**: Clear visual focus states
- **Shortcuts**: Alt+N for navigation, Ctrl+R for refresh
- **Skip Links**: Jump to main content

### 6.2 Screen Reader Support
- **ARIA Labels**: Descriptive labels for icon buttons
- **Role Attributes**: Proper semantic roles
- **Live Regions**: Status updates announced
- **Alternative Text**: Meaningful descriptions

### 6.3 Color Accessibility
- **Contrast Ratios**: WCAG AA compliance (4.5:1 minimum)
- **Color Independence**: Information not conveyed by color alone
- **High Contrast Mode**: Support for system preferences

## 7. Animation & Transitions

### 7.1 Micro-interactions
- **Button Hover**: 150ms ease-in-out
- **Page Transitions**: 200ms slide/fade
- **Loading States**: Smooth spinner animations
- **Status Changes**: Color transition over 300ms

### 7.2 Performance Considerations
- **Hardware Acceleration**: Use transform/opacity for animations
- **Reduced Motion**: Respect prefers-reduced-motion
- **60fps Target**: Smooth animations on all devices

## 8. Error States & Empty States

### 8.1 Error Handling
```html
<div class="error-state">
    <i data-lucide="alert-circle" width="48" height="48"></i>
    <h3>Connection Failed</h3>
    <p>Unable to connect to MinIO servers. Please check your configuration.</p>
    <button class="btn-primary">
        <i data-lucide="refresh-cw" width="16" height="16"></i>
        Retry
    </button>
</div>
```

### 8.2 Empty States
```html
<div class="empty-state">
    <i data-lucide="globe" width="48" height="48"></i>
    <h3>No Sites Configured</h3>
    <p>Get started by adding your first MinIO site to the replication group.</p>
    <button class="btn-primary">
        <i data-lucide="plus" width="16" height="16"></i>
        Add Site
    </button>
</div>
```

## 9. Performance Requirements

### 9.1 Loading Performance
- **First Paint**: <1.5s
- **Interactive**: <3s
- **Icon Loading**: <500ms for all icons
- **Page Transitions**: <200ms

### 9.2 Runtime Performance
- **Smooth Scrolling**: 60fps on all devices
- **Memory Usage**: <50MB for typical usage
- **Icon Rendering**: Hardware accelerated where possible

## 10. Testing Requirements

### 10.1 Browser Compatibility
- **Chrome**: Latest 2 versions
- **Firefox**: Latest 2 versions
- **Safari**: Latest 2 versions
- **Edge**: Latest 2 versions

### 10.2 Device Testing
- **Desktop**: 1920x1080, 1366x768
- **Tablet**: iPad (1024x768), Android tablets
- **Mobile**: iPhone (375x667), Android phones (360x640)

### 10.3 Accessibility Testing
- **Screen Readers**: NVDA, JAWS, VoiceOver
- **Keyboard Only**: Full functionality without mouse
- **High Contrast**: Windows high contrast mode

## 11. Implementation Checklist

### 11.1 Lucid Icons Integration
- [ ] CDN script included in HTML
- [ ] All navigation icons implemented with data-lucide
- [ ] All action buttons use consistent icon sizes
- [ ] Dynamic content calls lucide.createIcons()
- [ ] Fallback handling for failed icon loading

### 11.2 Responsive Design
- [ ] Mobile-first CSS approach
- [ ] Sidebar collapses on mobile
- [ ] Touch-friendly interface elements
- [ ] Horizontal scroll for wide tables
- [ ] Consistent spacing across breakpoints

### 11.3 Accessibility
- [ ] Keyboard navigation works throughout
- [ ] Screen reader friendly markup
- [ ] Sufficient color contrast ratios
- [ ] Alternative text for all icons
- [ ] Focus management in modals

### 11.4 Performance
- [ ] Icons load within 500ms
- [ ] Smooth animations at 60fps
- [ ] Optimized image assets
- [ ] Minimal JavaScript bundle size
- [ ] Efficient CSS animations

This comprehensive UI/UX specification ensures a consistent, accessible, and performant interface for the MinIO Site Replication Management console with proper Lucid Icons integration.