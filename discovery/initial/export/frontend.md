# Export Domain: Frontend

**Generated:** 2026-02-01
**Domain:** Export functionality for groups, discussions, and polls

---

## Overview

The frontend provides a modal dialog for initiating group exports and direct links for poll exports. Export functionality is primarily admin-focused.

---

## 1. Export Data Modal Component

**Location:** `/vue/src/components/group/export_data_modal.vue`

This is the main UI component for group data export.

### Component Structure

**Props:**
- `group`: The Group object to export

### Template Layout

The modal presents three export format options:

1. **CSV Export Section**
   - Header: "As CSV"
   - Button triggers confirmation modal for CSV export
   - Calls `group.exportCSV()` on confirm

2. **HTML Export Section**
   - Header: "As HTML"
   - Direct link button opening `/g/{group.key}/export.html?export=1` in new tab
   - Synchronous - opens immediately

3. **JSON Export Section**
   - Header: "As JSON"
   - Button triggers confirmation modal for JSON export
   - Calls `group.export()` on confirm

4. **Help Link**
   - Links to external documentation on help.loomio.com

### Confirmation Flow

Both JSON and CSV exports use a ConfirmModal component:
- Title: "group_export_modal.title"
- Help text: "group_export_modal.body"
- Submit button: "group_export_modal.submit"
- Flash message: "group_export_modal.flash"

The confirmation explains that:
- Export will be processed in background
- User will receive email with download link
- Link expires after one week

---

## 2. Group Model Export Methods

**Location:** `/vue/src/shared/models/group_model.js`

### export()

**Purpose:** Initiate JSON export

**Implementation:**
```
Pseudo-code:
  POST to /api/v1/groups/{id}/export
```

**Binding:** Method is bound in constructor for proper `this` context

### exportCSV()

**Purpose:** Initiate CSV export

**Implementation:**
```
Pseudo-code:
  POST to /api/v1/groups/{id}/export_csv
```

**Binding:** Method is bound in constructor for proper `this` context

---

## 3. Poll Export Access

Poll exports are accessed via direct URL rather than Vue component:

**URL Pattern:** `/p/{poll.key}/export.csv`

**Access:** Available to any user who:
- Can view the poll
- Can see poll results

**Trigger:** Typically linked from poll action menus or result pages

---

## 4. I18n Translation Keys

Based on the component usage, these translation keys are needed:

### Export Modal
- `export_data_modal.title` - Modal title
- `export_data_modal.as_csv` - CSV section header
- `export_data_modal.as_html` - HTML section header
- `export_data_modal.as_json` - JSON section header

### Group Page Options
- `group_page.options.export_data_as_csv` - CSV button text
- `group_page.options.export_data_as_html` - HTML button text
- `group_page.options.export_data_as_json` - JSON button text

### Confirmation Modal
- `group_export_modal.title` - Confirmation title
- `group_export_modal.body` - Explanation text
- `group_export_modal.submit` - Submit button text
- `group_export_modal.flash` - Success flash message

---

## 5. UI/UX Flow

### Group Export Flow

```
User opens Group Settings/Options menu
    |
    v
Selects "Export Data" option
    |
    v
ExportDataModal opens
    |
    +---> CSV: Click button -> ConfirmModal -> API call -> Flash "Check email"
    |
    +---> HTML: Click button -> New tab opens with HTML export page
    |
    +---> JSON: Click button -> ConfirmModal -> API call -> Flash "Check email"
```

### Poll Export Flow

```
User views Poll results
    |
    v
Clicks "Export" or "Download CSV" link
    |
    v
Direct download of CSV file
```

---

## 6. Component Dependencies

### Import Dependencies
- `openModal` from `@/shared/helpers/open_modal`
- `AppConfig` from `@/shared/services/app_config`

### Child Components Used
- `ConfirmModal` - For export confirmation
- `DismissModalButton` - Modal close button

### Vuetify Components
- `v-card` - Modal container
- `v-card-text` - Content area
- `v-btn` - Action buttons
- `v-divider` - Section separator
- `v-alert` - Help link container

---

## 7. Access Control on Frontend

### Group Export Modal Access

The export modal is only shown to group admins. This is controlled by:
- Menu visibility in group settings
- AbilityService checks (not explicitly in this component)

The component itself does not perform permission checks - it assumes the parent component/menu only renders it for authorized users.

### Poll Export Link Access

Poll export links are conditionally shown based on:
- Poll visibility
- Results visibility state

---

## 8. Error Handling

### API Call Failures

The group model methods return promises. Error handling depends on how the calling code handles rejections:

- Flash service displays success message on resolve
- Network errors would be caught by global error handler
- 403 errors would indicate permission issues

### No Explicit Error UI

The export modal does not have explicit error handling UI - it relies on the standard flash notification system.

---

## 9. Accessibility Considerations

### Current Implementation
- Uses semantic button elements
- Has descriptive text for each export option
- External link opens in new tab

### Potential Improvements
- Add aria-labels to buttons
- Indicate loading state during export initiation
- Provide accessible feedback for success/failure

---

## 10. Related Components

### Action Menus
Export options appear in group action menus, controlled by templates not specific to export.

### Group Page
The export modal is launched from the group page settings.

### Poll Show Page
Poll export links appear in poll result views.

---

## Confidence Rating: 4/5

The frontend implementation is straightforward but relatively minimal. The modal component is simple and relies heavily on backend processing. Some aspects (like exactly where export options appear in menus) would require tracing through more component code to fully document.
