# CSS Classes Guide - Web UI

Hướng dẫn sử dụng các CSS classes chuẩn hóa cho dự án mc-tool web UI.

## Form Controls

### Input Fields
```jsx
// Basic input
<input className="input" type="text" />

// Small input
<input className="input input-sm" type="text" />

// Large input
<input className="input input-lg" type="text" />

// Disabled input
<input className="input" type="text" disabled />
```

### Select Boxes
```jsx
// Basic select
<select className="select">
  <option>Option 1</option>
</select>

// Small select
<select className="select-sm">
  <option>Option 1</option>
</select>
```

### Checkboxes
```jsx
// Checkbox with label
<label className="checkbox-label">
  <input type="checkbox" className="checkbox" />
  Label text
</label>
```

### Radio Buttons
```jsx
// Radio with label
<label className="radio-label">
  <input type="radio" className="radio" name="group" />
  Label text
</label>
```

### Textarea
```jsx
<textarea className="textarea" rows="4"></textarea>
```

## Buttons

### Button Variants
```jsx
// Primary button
<button className="btn btn-primary">Primary</button>

// Secondary button
<button className="btn btn-secondary">Secondary</button>

// Success button
<button className="btn btn-success">Success</button>

// Warning button
<button className="btn btn-warning">Warning</button>

// Danger button
<button className="btn btn-danger">Danger</button>

// Info button
<button className="btn btn-info">Info</button>

// Neutral button
<button className="btn btn-neutral">Neutral</button>

// Outline button
<button className="btn btn-outline-primary">Outline</button>
```

### Button Sizes
```jsx
// Small button
<button className="btn btn-primary btn-sm">Small</button>

// Icon button
<button className="btn-icon">
  <Icon size={16} />
</button>

// Icon only button
<button className="btn-icon-only">
  <Icon size={16} />
</button>
```

### Button States
```jsx
// Disabled button
<button className="btn btn-primary" disabled>Disabled</button>
```

## Layout Utilities

### Flex
```jsx
// Flex container
<div className="flex">...</div>

// Flex column
<div className="flex flex-col">...</div>

// Flex with alignment
<div className="flex items-center justify-between">...</div>

// Flex with gap
<div className="flex gap-3">...</div>

// Flex wrap
<div className="flex flex-wrap gap-2">...</div>
```

### Spacing

#### Padding
```jsx
<div className="p-4">padding: 16px</div>
<div className="px-4">padding-left & right: 16px</div>
<div className="py-4">padding-top & bottom: 16px</div>
<div className="pt-4 pb-4">padding-top & bottom: 16px each</div>

// Values: 0 (0px), 1 (4px), 2 (8px), 3 (12px), 4 (16px), 5 (20px), 6 (24px)
```

#### Margin
```jsx
<div className="m-4">margin: 16px</div>
<div className="mt-4">margin-top: 16px</div>
<div className="mb-4">margin-bottom: 16px</div>
<div className="mx-auto">margin-left & right: auto</div>

// Values: 0 (0px), 1 (4px), 2 (8px), 3 (12px), 4 (16px), 5 (20px), 6 (24px)
```

#### Gap
```jsx
<div className="flex gap-1">gap: 4px</div>
<div className="flex gap-2">gap: 8px</div>
<div className="flex gap-3">gap: 12px</div>
<div className="flex gap-4">gap: 16px</div>
<div className="flex gap-5">gap: 20px</div>
<div className="flex gap-6">gap: 24px</div>
```

## Typography

### Font Sizes
```jsx
<div className="text-xs">12px</div>
<div className="text-sm">13px</div>
<div className="text-base">14px</div>
<div className="text-lg">16px</div>
<div className="text-xl">18px</div>
<div className="text-2xl">20px</div>
```

### Font Weights
```jsx
<div className="font-normal">400</div>
<div className="font-medium">500</div>
<div className="font-semibold">600</div>
<div className="font-bold">700</div>
```

### Text Colors
```jsx
<div className="text-primary">Primary text</div>
<div className="text-secondary">Secondary text</div>
<div className="text-muted">Muted text</div>
<div className="text-success">Success text</div>
<div className="text-warning">Warning text</div>
<div className="text-danger">Danger text</div>
<div className="text-info">Info text</div>
<div className="text-white">White text</div>
```

### Text Alignment
```jsx
<div className="text-left">Left aligned</div>
<div className="text-center">Center aligned</div>
<div className="text-right">Right aligned</div>
```

## Background & Borders

### Background Colors
```jsx
<div className="bg-primary">White background</div>
<div className="bg-secondary">#F5F5F5 background</div>
<div className="bg-gray-50">#f9fafb background</div>
<div className="bg-gray-100">#f3f4f6 background</div>
<div className="bg-success-light">Success light background</div>
<div className="bg-warning-light">Warning light background</div>
<div className="bg-danger-light">Danger light background</div>
```

### Borders
```jsx
<div className="border">Border all sides</div>
<div className="border-t">Border top</div>
<div className="border-b">Border bottom</div>
<div className="border-l">Border left</div>
<div className="border-r">Border right</div>
```

### Border Radius
```jsx
<div className="rounded">6px</div>
<div className="rounded-sm">4px</div>
<div className="rounded-lg">8px</div>
<div className="rounded-full">50%</div>
```

## Common Patterns

### Empty State
```jsx
<div className="empty-state">
  No data available
</div>
```

### Table Responsive
```jsx
<div className="table-responsive">
  <table className="table">...</table>
</div>
```

### Pagination
```jsx
<div className="pagination">
  <span className="pagination-info">1-10 of 100</span>
  <div className="pagination-controls">
    <button className="pagination-button">Prev</button>
    <span className="pagination-page-info">1 / 10</span>
    <button className="pagination-button">Next</button>
  </div>
</div>
```

### Card with Shadow
```jsx
<div className="card card-shadow">
  Content
</div>
```

## Width & Overflow

```jsx
<div className="w-full">width: 100%</div>
<div className="w-auto">width: auto</div>
<div className="min-w-0">min-width: 0</div>

<div className="overflow-auto">overflow: auto</div>
<div className="overflow-hidden">overflow: hidden</div>
<div className="overflow-x-auto">overflow-x: auto</div>
<div className="overflow-y-auto">overflow-y: auto</div>
```

## Migration Examples

### Before (Inline Styles)
```jsx
<div style={{ 
  display: 'flex', 
  alignItems: 'center', 
  gap: '12px',
  padding: '16px',
  backgroundColor: '#f9fafb',
  borderRadius: '8px'
}}>
  Content
</div>
```

### After (CSS Classes)
```jsx
<div className="flex items-center gap-3 p-4 bg-gray-50 rounded-lg">
  Content
</div>
```

### Before (Complex Button)
```jsx
<button style={{ 
  padding: '6px 10px',
  borderRadius: '6px',
  border: '1px solid #d1d5db',
  backgroundColor: '#ffffff',
  color: '#1f2937',
  fontSize: '12px',
  cursor: 'pointer'
}}>
  Click me
</button>
```

### After (CSS Classes)
```jsx
<button className="pagination-button">
  Click me
</button>
```

## Best Practices

1. **Prefer CSS classes over inline styles**: Sử dụng CSS classes thay vì inline styles khi có thể
2. **Combine classes**: Kết hợp nhiều classes để tạo ra style mong muốn
3. **Use semantic naming**: Sử dụng tên class có ý nghĩa (btn-primary, text-secondary)
4. **Consistent spacing**: Sử dụng scale spacing nhất quán (4px increments)
5. **Minimize custom styles**: Chỉ sử dụng inline styles cho những trường hợp đặc biệt (gradient, động)

## Notes

- Tất cả các CSS classes đã được định nghĩa trong `/web/src/styles.css`
- CSS variables được định nghĩa trong `:root` selector
- Responsive breakpoint: 768px cho mobile
- Transition mặc định: `all 0.2s ease`
