# Documents Domain: Frontend

**Generated:** 2026-02-01
**Domain:** Documents and Attachments

---

## Overview

The frontend handles documents and attachments through:
1. LokiJS record interfaces and models
2. File upload service using Active Storage Direct Upload
3. Vue components for display and management
4. Mixin for document associations on records

---

## 1. Record Interfaces

### DocumentRecordsInterface

**Location:** `/vue/src/shared/interfaces/document_records_interface.js`

**Purpose:**
Provides API methods for fetching documents.

**Methods:**

#### fetchByModel(model)

Fetches documents for a specific model instance.

**Request:**
- GET /api/v1/documents
- Param: {model_type}_id = model.id

**Usage:**
When loading a discussion, poll, or group's documents.

#### fetchByDiscussion(discussion)

Fetches all documents related to a discussion.

**Request:**
- GET /api/v1/documents/for_discussion
- Param: discussion_key

**Purpose:**
Gets documents from discussion, its polls, and its comments in one request.

### AttachmentRecordsInterface

**Location:** `/vue/src/shared/interfaces/attachment_records_interface.js`

**Purpose:**
Basic interface for attachment records - no custom methods beyond base CRUD.

**Confidence: 5/5** - Interfaces are minimal and clear.

---

## 2. Frontend Models

### DocumentModel

**Location:** `/vue/src/shared/models/document_model.js`

**Indices:** modelId, authorId

**Relationships:**
- belongsTo author (from users by authorId)
- belongsTo group

**Methods:**

#### model()

Returns the parent model by looking up:
- Records[modelType.toLowerCase() + 's'].find(modelId)

Example: If modelType is "Discussion", returns Records.discussions.find(modelId)

#### modelTitle()

Returns display title based on parent model type:
- Group: group name
- Discussion: discussion title
- Outcome: parent poll title
- Comment: parent discussion title
- Poll: poll title

#### authorName()

Returns author's name with title if author exists.

#### isAnImage()

Returns true if icon property is "image".

### AttachmentModel

**Location:** `/vue/src/shared/models/attachment_model.js`

**Indices:** recordType, recordId

**Relationships:**
- belongsTo author (from users)

**Methods:**

#### model()

Returns parent model using eventTypeMap for type resolution.

#### group()

Returns parent model's group.

#### isAnImage()

Returns true if icon property is "image".

**Confidence: 5/5** - Models are simple data wrappers.

---

## 3. File Uploader Service

**Location:** `/vue/src/shared/services/file_uploader.js`

### Purpose

Wraps Active Storage Direct Upload for client-side file uploads.

### Constructor

Accepts onProgress callback for upload progress tracking.

### upload(file) Method

**Flow:**
1. Create DirectUpload instance with file and /direct_uploads URL
2. Configure XHR progress listener
3. Return promise that resolves with blob data

**Progress Callback:**
Called with XHR progress event when lengthComputable.

**Return Value:**
Promise resolving to blob object with:
- signed_id: For attaching to records
- download_url: For immediate display
- preview_url: For image thumbnails (if representable)

**Usage Pattern:**
Pseudo-code for using the uploader:
```
uploader = new FileUploader({ onProgress: updateProgressBar })
blob = await uploader.upload(file)
// Use blob.signed_id when saving record
```

**Confidence: 5/5** - Standard Active Storage integration.

---

## 4. Attachment Service

**Location:** `/vue/src/shared/services/attachment_service.js`

### Purpose

Provides action definitions for attachment and document deletion.

### actions(attachment) Method

Returns action definitions for the given attachment/document.

#### delete_attachment Action

**Condition:** Item is an Attachment and user can administer the group.

**Action:**
Opens ConfirmModal with delete confirmation, then calls attachment.destroy().

#### delete_document Action

**Condition:** Item is a Document and user can administer the group.

**Action:**
Opens ConfirmModal with delete confirmation, then calls document.destroy().

**UI Pattern:**
Both actions:
- Use "mdi-delete" icon
- Show "common.action.delete" label
- Require confirmation modal
- Show success flash after deletion

**Confidence: 5/5** - Simple action definitions.

---

## 5. HasDocuments Mixin

**Location:** `/vue/src/shared/mixins/has_documents.js`

### Purpose

Adds document association methods to model instances.

### apply(model, opts) Method

Adds the following to model:

#### newDocumentIds / removedDocumentIds

Arrays for tracking document changes before save.

#### documents()

Returns documents from LokiJS where:
- modelId matches model.id
- modelType matches capitalized model singular name

#### newDocuments()

Returns documents matching newDocumentIds.

#### newAndPersistedDocuments()

Returns union of documents() and newDocuments(), excluding removedDocumentIds.

#### hasDocuments()

Returns true if newAndPersistedDocuments has any items.

#### serialize()

Override that adds document_ids to serialization data.

#### showDocumentTitle

Set from opts.showTitle for display configuration.

**Usage:**
Applied to models that support document attachments (discussions, polls, etc.).

**Confidence: 5/5** - Mixin is well-structured.

---

## 6. Vue Components

### FilesPanel

**Location:** `/vue/src/components/group/files_panel.vue`

**Purpose:**
Group-level file browser showing both documents and attachments.

**Features:**
- Search by filename/title
- Subgroup filtering (mine, all, none)
- Pagination with load more
- Delete action for admins

**Data Loading:**
Uses two RecordLoaders:
- documents loader (for_group path)
- attachments loader

**Display:**
Combines documents and attachments into items array, sorted by createdAt descending.

**Table Columns:**
- Filename with icon
- Uploaded by (avatar)
- Uploaded at (time ago)
- Actions (admin only)

### AttachmentList

**Location:** `/vue/src/components/thread/attachment_list.vue`

**Purpose:**
Simple list wrapper for displaying attachments.

**Props:**
- attachments: Array or Object of attachment items

**Rendering:**
Iterates attachments and renders AttachmentListItem for each.

### AttachmentListItem

**Location:** `/vue/src/components/thread/attachment_list_item.vue`

**Purpose:**
Individual attachment display (not fully analyzed but inferred from usage).

### FilesList

**Location:** `/vue/src/components/lmo_textarea/files_list.vue`

**Purpose:**
Display files within the text editor area (rich text attachments).

**Confidence: 4/5** - Component analysis based on naming and structure.

---

## 7. Upload Flow

### Step 1: User Selects File

Text editor or file input captures file selection.

### Step 2: Direct Upload

FileUploader.upload() is called:
- Creates DirectUpload instance
- Initiates upload to /direct_uploads
- Progress events update UI

### Step 3: Blob Response

Server returns blob data:
- signed_id for record attachment
- URLs for display

### Step 4: Record Save

When parent record is saved:
- files or image_files param includes signed_ids
- Server attaches blobs to record
- build_attachments generates JSON metadata

### Step 5: Display

Attachments appear via:
- attachments column data (inline display)
- AttachmentList component (file listings)

**Confidence: 5/5** - Standard Active Storage Direct Upload pattern.

---

## 8. Document vs Attachment Display

### Documents

Displayed with:
- title (user-provided or filename)
- icon (from doctype)
- color (from doctype)
- Author name
- Download link

### Attachments

Displayed with:
- filename (from blob)
- icon (from doctype)
- preview (if image)
- Author (from parent record)
- Download link

**Key Difference:**
Documents have explicit titles; attachments use original filenames.

**Confidence: 5/5** - Display patterns are clear from components.

---

## Summary

The frontend document/attachment handling follows patterns:
1. Direct Upload for file transfers (bypasses Rails)
2. LokiJS for client-side record management
3. Polymorphic model resolution for parent relationships
4. Action service for delete operations
5. Combined display in files panels

Integration points:
- FileUploader for uploads
- Records.documents and Records.attachments for queries
- HasDocuments mixin for model associations
- AttachmentService for actions
