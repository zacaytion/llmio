# Export Domain: Controllers

**Generated:** 2026-02-01
**Domain:** Export functionality for groups, discussions, and polls

---

## Overview

Export functionality is handled by both API controllers (for initiating background exports) and non-API controllers (for synchronous HTML/CSV downloads).

---

## 1. API Controllers

### Api::V1::GroupsController

**Location:** `/app/controllers/api/v1/groups_controller.rb`

Handles JSON and CSV export requests for groups via background workers.

#### export (JSON)

**Route:** `POST /api/v1/groups/:id/export`

**Authorization:** Requires `:export` permission on group

**Process:**
1. Load group and authorize export action
2. Delegate to GroupService.export
3. Return success response immediately

**Response:** `{ success: :ok }` - Export happens asynchronously

**Permission Check:**
- Uses `load_and_authorize(:group, :export)`
- Only group admins can export (see Ability::Group)

#### export_csv

**Route:** `POST /api/v1/groups/:id/export_csv`

**Authorization:** Requires `:export` permission on group

**Process:**
1. Load group and authorize export action
2. Enqueue GroupExportCsvWorker directly (bypasses service layer)
3. Return success response immediately

**Response:** `{ success: :ok }` - Export happens asynchronously

---

## 2. Non-API Controllers

### GroupsController

**Location:** `/app/controllers/groups_controller.rb`

Handles synchronous HTML export (browser-based viewing).

#### export

**Route:** `GET /g/:key/export` (accessible at `/g/:key/export.html`)

**Authorization:** Requires `:export` permission on group

**Process:**
1. Load group and authorize export action
2. Create GroupExporter instance
3. Render HTML view

**Response:** HTML page with tabular export data (synchronous)

**Use Case:** Users can view/print export data directly in browser

---

### PollsController

**Location:** `/app/controllers/polls_controller.rb`

Handles poll export in HTML and CSV formats.

#### export

**Route:** `GET /p/:key/export`

**Authorization:** Requires `:export` permission on poll

**Process:**
1. Load poll and authorize export action
2. Create PollExporter instance
3. Respond based on requested format:
   - HTML: Renders export template
   - CSV: Sends file download with generated filename

**CSV Response Headers:**
- Content-Type: text/csv
- Content-Disposition: attachment with filename `poll-{id}-{key}-{title}.csv`

**Permission Note:** Poll export requires user can show the poll AND results are visible (`poll.show_results?`)

---

### DiscussionsController

**Location:** `/app/controllers/discussions_controller.rb`

**Route:** `GET /d/:key/export`

**Current State:** The controller is essentially empty - discussion export is not implemented.

```
class DiscussionsController < ApplicationController
end
```

The route exists but has no implementation, which would result in a routing error or missing action error.

---

## 3. Route Definitions

**Location:** `/config/routes.rb`

### API Routes (namespaced under api/v1/groups)
```
POST /api/v1/groups/:id/export      -> Api::V1::GroupsController#export
POST /api/v1/groups/:id/export_csv  -> Api::V1::GroupsController#export_csv
```

### Non-API Routes (public/direct access)
```
GET /g/:key/export -> GroupsController#export (HTML)
GET /p/:key/export -> PollsController#export (HTML/CSV)
GET /d/:key/export -> DiscussionsController#export (NOT IMPLEMENTED)
```

---

## 4. Authorization Summary

### Group Export
- **Permission:** `:export`
- **Requirement:** User must be a group admin
- **Checked via:** `Ability::Group` module
- **Both JSON and CSV exports require the same permission**

### Poll Export
- **Permission:** `:export`
- **Requirements:**
  1. User can `:show` the poll (visibility check)
  2. Poll results are visible (`poll.show_results?`)
- **Checked via:** `Ability::Poll` module
- **Note:** Even public polls require results to be visible (timing-dependent)

---

## 5. Response Patterns

### Asynchronous Exports (API)

```
POST request
    |
    v
Immediate response: { success: :ok }
    |
    v
Background worker processes export
    |
    v
Email notification with download link
    |
    v
File auto-deleted after 1 week
```

### Synchronous Exports (Non-API)

```
GET request
    |
    v
Authorization check
    |
    v
Generate export data
    |
    v
Direct response (HTML page or CSV download)
```

---

## 6. Error Handling

### Authorization Failures
- Handled by SnorlaxBase rescue_from
- Returns 403 Forbidden with error message
- Standard CanCan::AccessDenied exception flow

### Missing Records
- Handled by LoadAndAuthorize concern
- Returns 404 Not Found
- Uses `find_by!` which raises RecordNotFound

### Export Processing Errors
- Background workers: Logged to Sidekiq, may retry
- Synchronous: Standard Rails error handling (500)

---

## 7. Frontend Integration Points

The API endpoints are called from the Vue frontend:

### Group Model Methods (group_model.js)
```
export() - Calls POST /api/v1/groups/:id/export
exportCSV() - Calls POST /api/v1/groups/:id/export_csv
```

### Export Data Modal Component
- Location: `/vue/src/components/group/export_data_modal.vue`
- Provides UI for selecting export format
- JSON and CSV trigger API calls (background)
- HTML opens direct link in new tab (synchronous)

---

## 8. Controller Comparison Table

| Controller | Route | Method | Auth | Processing | Delivery |
|------------|-------|--------|------|------------|----------|
| Api::V1::GroupsController | POST /api/v1/groups/:id/export | export | Admin | Background | Email |
| Api::V1::GroupsController | POST /api/v1/groups/:id/export_csv | export_csv | Admin | Background | Email |
| GroupsController | GET /g/:key/export | export | Admin | Sync | HTML |
| PollsController | GET /p/:key/export | export | Can show + results | Sync | HTML/CSV |
| DiscussionsController | GET /d/:key/export | - | - | NOT IMPL | - |

---

## Confidence Rating: 5/5

The controller layer is straightforward with clear separation between API (background) and non-API (synchronous) exports. The only notable gap is the unimplemented discussion export endpoint.
