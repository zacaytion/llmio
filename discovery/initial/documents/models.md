# Documents Domain: Models

**Generated:** 2026-02-01
**Domain:** Documents and Attachments

---

## Overview

Loomio has two parallel systems for file management:

1. **Document Model** - A dedicated model for URL-based links and files attached explicitly to groups/discussions/polls/comments
2. **Active Storage Attachments** - Rails native file attachments embedded in rich text content via the HasRichText concern

The distinction:
- Documents are explicit, user-facing file references with metadata (title, icon, color)
- Attachments are inline files embedded within rich text content (body/description fields)

---

## 1. Document Model

**Location:** `/app/models/document.rb`

### Purpose

The Document model represents an explicitly attached file or URL to a model (Group, Discussion, Poll, Comment, Outcome). It provides metadata display, icon classification, and download capabilities.

### Core Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| title | string | Display name for the document (required) |
| url | string | External URL or nil if file attached |
| doctype | string | Classification name from doctypes.yml |
| icon | string | Material Design icon name |
| color | string | Hex color for display |
| model_id | integer | Polymorphic parent ID |
| model_type | string | Polymorphic parent class name |
| group_id | integer | Denormalized for query efficiency |
| author_id | integer | User who uploaded/created |

### Associations

- **Belongs to model (polymorphic):** Can be Group, Discussion, Poll, Comment, or Outcome
- **Belongs to author:** The User who created the document
- **Has one attached file:** Active Storage attachment

### Key Behaviors

**Metadata Auto-Detection:**
Before validation, the model automatically detects document type by matching the file content type or URL against regex patterns in doctypes.yml. Sets doctype, icon, and color accordingly.

**Group ID Denormalization:**
Before save, the group_id is automatically set by traversing the polymorphic parent's group_id method. This enables efficient per-group document queries.

**URL Handling:**
The url method returns:
- Active Storage URL if file is attached
- Absolute URL if url column contains http/https prefix
- URL prefixed with asset host otherwise

**Helper Methods:**
- download_url: Returns Rails blob path for attached files
- is_an_image?: Returns true if metadata icon is "image"
- group/discussion/poll: Proxy methods to traverse polymorphic parent

### Search Scope

Supports case-insensitive ILIKE search on title field via search_for scope.

**Confidence: 5/5** - Model is straightforward with clear responsibilities.

---

## 2. Attachment Model

**Location:** `/app/models/attachment.rb`

### Purpose

A thin subclass of ActiveStorage::Attachment that exists primarily for:
- Custom serialization
- Ability checks
- Query object support

The model itself has no additional behavior - it inherits everything from ActiveStorage::Attachment.

### Relationship to Active Storage

Active Storage uses a polymorphic attachment system:
- active_storage_attachments table links blobs to records
- active_storage_blobs table stores file metadata
- Actual files stored in configured storage backend

**Confidence: 5/5** - Very simple model.

---

## 3. HasRichText Concern

**Location:** `/app/models/concerns/has_rich_text.rb`

### Purpose

Provides rich text handling for models with body/description/details fields. Crucially, this concern adds the file attachment capability that creates inline attachments.

### Attachment Declarations

The concern declares:
- has_many_attached :files (detach on destroy)
- has_many_attached :image_files (detach on destroy)

These are the attachments that appear in the AttachmentQuery searches.

### Build Attachments Method

The build_attachments callback generates a JSON array stored in the attachments column. For each attached file, it stores:
- id, filename, content_type, byte_size from blob
- preview_url (if representable, using PREVIEW_OPTIONS)
- download_url (Rails blob path)
- icon (detected from doctypes.yml)
- signed_id (for frontend operations)

### Preview Options

Image previews are constrained to:
- Maximum 1280x1280 pixels
- Quality 85
- Metadata stripped

### Assignment Protection

The assign_attributes_and_files method prevents accidental attachment removal by deleting nil values for files/image_files from params.

**Confidence: 5/5** - Well-documented concern with clear patterns.

---

## 4. Doctypes Configuration

**Location:** `/config/doctypes.yml`

### Purpose

Defines file type classification rules using regex matching against content type or filename.

### Supported Types

| Name | Pattern | Icon | Use Case |
|------|---------|------|----------|
| youtube_video | youtube.com/watch | video | YouTube embeds |
| image | gif/jpg/jpeg/png/tiff/svg | image | Image files |
| pdf | .pdf extension | file-pdf-box | PDF documents |
| excel | .xls/.xlsx | file-excel-box | Excel spreadsheets |
| document | .doc/.docx | file-word-box | Word documents |
| csv | .csv extension | file-document | CSV files |
| video | mp4/mov/m4a | file-video | Video files |
| text | .txt extension | file-document | Plain text |
| google_doc | docs.google.com/document | text-box | Google Docs links |
| google_sheet | docs.google.com/spreadsheet | google-spreadsheet | Google Sheets links |
| google_drive | drive.google.com | google-drive | Google Drive links |
| pull_request | github.com/pulls | source-pull | GitHub PRs |
| trello_board | trello.com/b/ | collage | Trello boards |
| trello_card | trello.com/c/ | cards | Trello cards |
| other | .* (fallback) | file | Unknown types |

The order matters - first match wins. The "other" type is a catch-all fallback.

**Confidence: 5/5** - Configuration file is self-documenting.

---

## 5. Model Relationships

### Groups

Groups have:
- Direct documents (has_many :documents as model)
- Discussion documents (through discussions)
- Poll documents (through polls)
- Comment documents (through comments)

### Discussions

Discussions have:
- Direct documents
- Poll documents (from discussion's polls)
- Comment documents (from discussion's comments)

### Document Cascade

When a parent model is destroyed:
- Documents are destroyed (dependent: :destroy on has_many)
- Active Storage attachments are detached (dependent: :detach)

**Confidence: 5/5** - Relationships verified in model files.

---

## 6. Storage Backend Configuration

**Location:** `/config/storage.yml`

### Available Storage Services

| Service | Provider | Configuration |
|---------|----------|---------------|
| test | Disk | tmp/storage (testing) |
| local | Disk | storage/ directory |
| amazon | S3 | AWS S3 with standard credentials |
| digitalocean | S3 | DigitalOcean Spaces |
| s3_compatible | S3 | Generic S3-compatible (Minio, etc) |
| google | GCS | Google Cloud Storage |

### Service Selection

The active storage service is selected by:
1. Environment-specific config (test.rb, development.rb)
2. Environment variable ACTIVE_STORAGE_SERVICE (production)

Default is :local if not specified.

### S3-Compatible Configuration

The s3_compatible service supports:
- Custom endpoint
- Force path style (for Minio)
- Custom regions
- Checksum configuration

**Confidence: 5/5** - Standard Rails Active Storage configuration.

---

## Summary

The documents domain uses a dual approach:
1. Document model for explicit, user-managed file references with metadata
2. Active Storage via HasRichText for inline rich text attachments

Both systems share:
- doctype classification from doctypes.yml
- Active Storage for actual file storage
- Group-based visibility and access control

The key difference is lifecycle and visibility:
- Documents are explicitly managed and searchable as standalone entities
- Attachments are embedded in content and discovered through content queries
