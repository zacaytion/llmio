# Documents Domain: Services

**Generated:** 2026-02-01
**Domain:** Documents and Attachments

---

## Overview

The documents domain has minimal service layer involvement. The DocumentService handles CRUD operations, while attachment handling is distributed across group export functionality and inline concern methods.

---

## 1. DocumentService

**Location:** `/app/services/document_service.rb`

### Purpose

Provides standard CRUD operations for the Document model with authorization and EventBus broadcasting.

### Methods

#### create(document:, actor:)

**Flow:**
1. Authorize actor can create the document
2. Assign author to actor
3. Set title to filename if not provided
4. Validate document
5. Save document
6. Broadcast "document_create" event

**Authorization Check:**
- If document has a model: actor must be able to update that model
- If no model: actor must have verified email

**Return:** None (void method)

#### update(document:, params:, actor:)

**Flow:**
1. Authorize actor can update the document
2. Assign permitted params (url, title, model_id, model_type)
3. Validate document
4. Save document
5. Broadcast "document_update" event

**Return:** None (void method)

#### destroy(document:, actor:)

**Flow:**
1. Authorize actor can destroy the document
2. Hard delete the document
3. Broadcast "document_destroy" event

**Return:** None (void method)

### Notable Patterns

**No Event Publishing:**
Unlike most Loomio services, DocumentService does not publish Events (the notification/activity kind). It only broadcasts via EventBus for side effects. This means document operations do not appear in activity timelines.

**Hard Delete:**
Documents are hard deleted, not soft deleted (discarded). This is different from most Loomio models.

**Simple Authorization:**
The service delegates authorization to ability modules but doesn't do complex permission checks itself.

**Confidence: 5/5** - Small, straightforward service.

---

## 2. GroupExportService

**Location:** `/app/services/group_export_service.rb`

### Purpose

Handles exporting and importing group data including attachments. While not document-specific, it's the primary service for attachment handling.

### Attachment Export

#### export(groups, group_name)

Collects attachments from:
- User avatars
- Group cover photos and logos
- Group direct files and image_files
- Comment files and image_files
- Discussion files and image_files
- Poll files and image_files
- Outcome files and image_files
- Subgroup attachments

For each attachment, writes a JSON record with:
- id, host, record_type, record_id
- name, filename, content_type
- path (relative), url (absolute)

#### puts_attachment(attachment, file)

Generates the attachment export record with:
- Download path from Rails blob path helper
- Full URL constructed from CANONICAL_HOST environment variable

### Attachment Import

#### import(filename_or_url, reset_keys:)

When importing data that includes attachments:
1. Parse all JSON records
2. Import non-attachment records first
3. Build ID migration map (old ID to new ID)
4. For each attachment record:
   - Find the new record ID from migration map
   - Enqueue DownloadAttachmentWorker with record data and new ID

#### download_attachment(record_data, new_id)

Called by DownloadAttachmentWorker:
1. Find the model by new ID
2. Open the URL to download file
3. Create blob with create_and_upload!
4. Attach blob to model's appropriate field
5. Rebuild attachments JSON if model supports it

### Back References for Attachments

The BACK_REFERENCES constant includes "attachments: user_id" under users, indicating attachment ownership transfers during user imports.

**Confidence: 4/5** - Complex service, but attachment portions are clear.

---

## 3. HasRichText Concern Methods

**Location:** `/app/models/concerns/has_rich_text.rb`

While technically a concern, these methods function as service-level attachment handling.

### build_attachments

Called before_save on models with rich text fields.

**Purpose:**
Generates the attachments JSONB column value containing metadata for all attached files.

**Flow:**
1. Check if attachments column exists (migration safety)
2. Map each attached file to metadata hash
3. For representable files, generate preview URL
4. Add download URL, icon, and signed_id
5. Store in attachments column

**Output Structure per File:**
- id: Blob ID
- filename: Original filename
- content_type: MIME type
- byte_size: File size in bytes
- preview_url: Variant URL (images only)
- download_url: Rails blob path
- icon: Doctype icon
- signed_id: For frontend operations

### assign_attributes_and_files(params)

**Purpose:**
Safely assign attributes while preventing accidental attachment removal.

**Logic:**
- If files or image_files keys are present but nil, remove them from params
- This prevents form submissions that don't include files from clearing existing attachments

### attachment_icon(name)

**Purpose:**
Look up icon from doctypes.yml for a given content type or filename.

**Confidence: 5/5** - Methods are well-documented in code.

---

## 4. Workers

### DownloadAttachmentWorker

**Location:** `/app/workers/download_attachment_worker.rb`

**Purpose:**
Background job to download attachments during group import.

**Flow:**
1. Receive record data hash and new ID
2. Delegate to GroupExportService.download_attachment

**Usage:**
Only called during group import process to fetch attachments from source URLs.

### AttachDocumentWorker

**Location:** `/app/workers/attach_document_worker.rb`

**Purpose:**
Migrate documents from legacy URL storage to Active Storage.

**Flow:**
1. Find document by ID
2. Skip if file already attached
3. Parse URL to get S3 path
4. Get S3 object
5. Create blob with object metadata
6. Attach blob to document

**Usage:**
Migration utility for transitioning from URL-based storage to Active Storage blobs.

**Confidence: 5/5** - Workers are simple delegation wrappers.

---

## 5. Attachment Cleanup

### Purging Mechanism

The AttachmentsController destroy action calls:
- attachment.purge_later

This uses Active Storage's built-in purge mechanism which:
1. Enqueues ActiveStorage::PurgeJob
2. Deletes the blob and file from storage backend
3. Deletes the attachment record

### User Redaction

**Location:** `/app/workers/redact_user_worker.rb`

When redacting users, their uploaded avatar is purged:
- user.uploaded_avatar.purge_later

### Orphan Handling

There is no explicit orphan cleanup mechanism for attachments. Orphaned attachments can occur if:
- Rich text content is edited to remove file references
- Parent records are deleted without attachment cleanup

Active Storage does not automatically clean orphaned blobs. Loomio relies on:
- dependent: :detach on has_many_attached (prevents cascade delete)
- Explicit purge_later calls for intentional deletion

**Confidence: 3/5** - Limited evidence of comprehensive cleanup; may need further investigation.

---

## Summary

The documents service layer is minimal by design:
- DocumentService provides thin CRUD wrappers
- GroupExportService handles bulk attachment operations
- HasRichText provides inline attachment building
- Workers handle async download operations

The lack of Event publishing for documents means they're treated as supporting data rather than primary activity items. Attachments are managed as part of their parent records' lifecycle.
