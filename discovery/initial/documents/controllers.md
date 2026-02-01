# Documents Domain: Controllers

**Generated:** 2026-02-01
**Domain:** Documents and Attachments

---

## Overview

The documents domain has three controllers:
1. DocumentsController - CRUD for Document model
2. AttachmentsController - Browsing and deleting Active Storage attachments
3. DirectUploadsController - File upload endpoint for Active Storage

---

## 1. DocumentsController

**Location:** `/app/controllers/api/v1/documents_controller.rb`

### Inheritance

Extends Api::V1::RestfulController, providing standard CRUD operations.

### Routes

| Method | Path | Action |
|--------|------|--------|
| POST | /api/v1/documents | create |
| PATCH | /api/v1/documents/:id | update |
| DELETE | /api/v1/documents/:id | destroy |
| GET | /api/v1/documents | index |
| GET | /api/v1/documents/for_group | for_group |
| GET | /api/v1/documents/for_discussion | for_discussion |

### Custom Actions

#### for_group

**Purpose:** List all documents associated with a group.

**Parameters:**
- group_id (required)
- subgroups: "mine", "all", or omitted for current group only
- q: search query for title

**Authorization Logic:**
- If user can see_private_content for group: Return all group documents
- Otherwise: Return only public discussion and comment documents

**Private Group Documents Query:**
1. Determine group IDs based on subgroups param
2. Query Document where group_id in those IDs
3. Order by created_at descending

**Public Group Documents Query:**
Uses UnionQuery to combine:
- Group's direct documents
- Public discussion documents
- Public comment documents

#### for_discussion

**Purpose:** List all documents for a discussion and its related content.

**Parameters:**
- discussion_id (required)

**Authorization:**
Calls load_and_authorize on discussion.

**Query:**
Uses UnionQuery to combine:
- Discussion's direct documents
- Poll documents (from discussion's polls)
- Comment documents (from discussion's comments)

### Accessible Records

The accessible_records method tries to load parent models in priority order:
1. Group
2. Discussion
3. Comment
4. Poll
5. Outcome (required)

Returns the documents association of the first successfully loaded model.

**Confidence: 5/5** - Controller logic is clear and well-structured.

---

## 2. AttachmentsController

**Location:** `/app/controllers/api/v1/attachments_controller.rb`

### Inheritance

Extends Api::V1::RestfulController.

### Routes

| Method | Path | Action |
|--------|------|--------|
| GET | /api/v1/attachments | index |
| DELETE | /api/v1/attachments/:id | destroy |

### Actions

#### index

**Purpose:** Search and list attachments across a group hierarchy.

**Parameters:**
- group_id (required)
- q: search query for filename
- per: results per page (default 20)
- from: offset for pagination

**Authorization Logic:**
1. Find group that current user belongs to
2. Calculate intersection of user's groups with target group hierarchy
3. Add subgroups where parent_members_can_see_discussions is true
4. Query AttachmentQuery with resulting group IDs

**Response:**
Returns attachment collection with total count for pagination.

#### destroy

**Purpose:** Delete an attachment.

**Flow:**
1. Load and authorize attachment for destroy
2. Get the parent record
3. Call purge_later on attachment
4. Save the record (triggers attachments column rebuild)
5. Return serialized parent record

**Why Return Parent Record:**
The parent record's attachments JSON column needs to be updated after purging. Returning the updated parent allows the frontend to update its local state.

### Serialization

Uses AttachmentSerializer with "attachments" as root key.

**Confidence: 5/5** - Controller is straightforward.

---

## 3. DirectUploadsController

**Location:** `/app/controllers/direct_uploads_controller.rb`

### Inheritance

Extends ActiveStorage::DirectUploadsController.

### Purpose

Handles direct upload to storage backend, bypassing Rails server for file transfer.

### Route

| Method | Path | Action |
|--------|------|--------|
| POST | /direct_uploads | create |

### CSRF Handling

- protect_from_forgery with: :exception
- skip_before_action :verify_authenticity_token

This allows JavaScript clients to POST directly without CSRF tokens while maintaining protection for other requests.

### Response Enhancement

Overrides direct_upload_json to add:
- download_url: Rails blob path for immediate access
- preview_url: Variant path for representable files (images)

**Standard Response Fields:**
- signed_id: For attaching to records
- direct_upload.url: Presigned URL for upload
- direct_upload.headers: Required headers for upload request

**Confidence: 5/5** - Standard Active Storage pattern with minor customization.

---

## 4. AttachmentQuery

**Location:** `/app/queries/attachment_query.rb`

### Purpose

Complex query object for finding attachments across multiple record types with group-based visibility.

### Method: find(group_ids, query, limit, offset)

**Strategy:**
Executes six separate queries and unions the results. Each query:
1. Joins active_storage_attachments with active_storage_blobs
2. Joins to specific record type
3. Filters by group membership
4. Filters to files only (not image_files)
5. ILIKE searches on filename
6. Applies limit/offset
7. Returns attachment IDs

**Record Types Queried:**
1. Groups - direct group files
2. Comments - via comments -> discussions -> groups
3. Outcomes - via outcomes -> polls -> groups
4. Stances - via stances -> polls -> groups
5. Discussions - direct discussion files
6. Polls - direct poll files

**Filtering:**
- Only "files" attachments (not image_files)
- Excludes discarded records
- Excludes revoked stances

**Performance:**
Each query returns IDs, then final query fetches full attachment records. This approach avoids complex UNION ALL SQL but may have performance implications at scale.

**Confidence: 4/5** - Query logic is clear but complex; edge cases may exist.

---

## 5. Authorization Flow

### Document Authorization

**Create/Update:**
- If document has parent model: User must be able to update parent
- If no parent model: User must have verified email

**Destroy:**
- User must be admin of document's group

### Attachment Authorization

**Show:**
- User must be member of attachment's record's group

**Destroy:**
- User must be admin of attachment's record's group

### Permission Inheritance

Both Document and Attachment permissions inherit from parent model permissions. A user who can edit a discussion can attach documents to it.

**Confidence: 5/5** - Authorization is straightforward delegation.

---

## 6. API Response Patterns

### Document Response

Root key: "documents"

Per document:
- id, title, icon, color, url, download_url
- web_url, thumb_url (images only)
- model_id, model_type, group_id
- created_at
- Embedded author user

### Attachment Response

Root key: "attachments"

Per attachment:
- id, filename, content_type, byte_size
- icon, preview_url, download_url
- created_at, record_type, record_id
- Embedded author user
- Embedded polymorphic record

**Confidence: 5/5** - Verified in serializer files.

---

## Summary

The controllers follow standard RESTful patterns with:
- DocumentsController for explicit document management
- AttachmentsController for inline attachment browsing
- DirectUploadsController for file uploads

Key patterns:
- Group-based visibility filtering
- Polymorphic parent model resolution
- Direct upload to storage backend
- Attachment purging with parent record update
