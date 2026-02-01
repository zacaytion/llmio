# Attachments JSONB - Final Synthesis

## Executive Summary

The `attachments` JSONB column stores an **array** of file metadata objects. The default value is `'[]'::jsonb` (empty array). This pattern is consistent across 8 tables and is used for rich text content with embedded files.

---

## Confirmed Findings

### Database Schema

**Default Value**: `'[]'::jsonb` (empty JSON array)

All 8 tables with attachments columns use identical schema:

| Table | Column Definition | Source |
|-------|-------------------|--------|
| comments | `attachments jsonb DEFAULT '[]'::jsonb NOT NULL` | schema_dump.sql:603 |
| discussion_templates | `attachments jsonb DEFAULT '[]'::jsonb NOT NULL` | schema_dump.sql:768 |
| discussions | `attachments jsonb DEFAULT '[]'::jsonb NOT NULL` | schema_dump.sql:831 |
| groups | `attachments jsonb DEFAULT '[]'::jsonb NOT NULL` | schema_dump.sql:1131 |
| outcomes | `attachments jsonb DEFAULT '[]'::jsonb NOT NULL` | schema_dump.sql:1535 |
| polls | `attachments jsonb DEFAULT '[]'::jsonb NOT NULL` | schema_dump.sql:1759 |
| stances | `attachments jsonb DEFAULT '[]'::jsonb NOT NULL` | schema_dump.sql:1977 |
| users | `attachments jsonb DEFAULT '[]'::jsonb NOT NULL` | schema_dump.sql:2330 |

### Migration History

| Date | Migration | Action |
|------|-----------|--------|
| 2019-03-26 | `add_attachments_to_comments.rb` | Added with `default: {}` |
| 2019-03-26 | `add_attachments_to_rich_text_models.rb` | Added to 5 more tables with `default: {}` |
| 2019-09-26 | `change_attachments_default_to_array.rb` | Changed all to `default: []` |
| 2023-07-31 | `create_discussion_templates_table.rb` | Created with `default: []` from start |

---

## JSONB Schema/Structure

### Attachment Object Definition

```typescript
interface Attachment {
  id: number;              // ActiveStorage blob ID
  signed_id: string;       // Signed blob identifier for secure access
  filename: string;        // Original filename with extension
  content_type: string;    // MIME type (e.g., "image/png", "application/pdf")
  byte_size: number;       // File size in bytes
  download_url: string;    // Relative path for downloading
  icon: string;            // UI icon identifier based on file type
  preview_url?: string;    // Optional: relative path for image/video preview
}
```

### Example JSON

```json
[
  {
    "id": 12345,
    "signed_id": "eyJfcmFpbHMiOnsi...",
    "filename": "document.pdf",
    "content_type": "application/pdf",
    "byte_size": 245678,
    "download_url": "/rails/active_storage/blobs/redirect/...",
    "icon": "pdf"
  },
  {
    "id": 12346,
    "signed_id": "eyJfcmFpbHMiOnsi...",
    "filename": "screenshot.png",
    "content_type": "image/png",
    "byte_size": 89012,
    "download_url": "/rails/active_storage/blobs/redirect/...",
    "preview_url": "/rails/active_storage/representations/redirect/...",
    "icon": "image"
  }
]
```

### Field Details

| Field | Type | Required | Source |
|-------|------|----------|--------|
| id | integer | Yes | `file.blob.slice(:id, ...)` |
| filename | string | Yes | `file.blob.slice(:filename, ...)` |
| content_type | string | Yes | `file.blob.slice(:content_type, ...)` |
| byte_size | integer | Yes | `file.blob.slice(:byte_size, ...)` |
| download_url | string | Yes | `rails_blob_path(file, only_path: true)` |
| signed_id | string | Yes | `file.signed_id` |
| icon | string | Yes | `attachment_icon(file.content_type || file.filename)` |
| preview_url | string | No | `rails_representation_path(...)` - only for representable files |

---

## Default Values and Migration Patterns

### PostgreSQL DDL

```sql
-- Creating table with attachments column
CREATE TABLE discussions (
    id BIGSERIAL PRIMARY KEY,
    -- ... other columns ...
    attachments jsonb DEFAULT '[]'::jsonb NOT NULL
);

-- Adding column to existing table
ALTER TABLE comments
ADD COLUMN attachments jsonb DEFAULT '[]'::jsonb NOT NULL;
```

### Rails Migration Pattern

```ruby
# Adding attachments column
add_column :table_name, :attachments, :jsonb, default: [], null: false

# Creating table with attachments
create_table :discussion_templates do |t|
  t.jsonb :attachments, default: [], null: false
end
```

---

## Serialization Behavior

### Rails (Source of Truth)

Serializers pass through the database value directly:

```ruby
# app/serializers/comment_serializer.rb
class CommentSerializer < ApplicationSerializer
  attributes :id,
             # ...
             :attachments,  # Direct attribute - outputs array as-is
             # ...
end
```

### Frontend Expectation

All JavaScript models expect array:

```javascript
// vue/src/shared/models/comment_model.js
defaultValues() {
  return {
    // ...
    attachments: [],
    // ...
  };
}
```

### API Response

Must output `[]` for empty, never `null`:

```json
// Correct
{"attachments": []}

// Incorrect - will break frontend
{"attachments": null}
```

---

## Related Patterns

### HasRichText Concern

Tables with attachments use the `HasRichText` concern which:
1. Declares `has_many_attached :files` (ActiveStorage)
2. Builds `attachments` JSONB from `files` association on save
3. JSONB is a denormalized cache of ActiveStorage metadata

```ruby
# app/models/concerns/has_rich_text.rb
included do
  has_many_attached :files, dependent: :detach
  before_save :build_attachments
end
```

### Tables Using HasRichText

- `comments`
- `discussions`
- `discussion_templates`
- `groups`
- `outcomes`
- `polls`
- `stances`
- `users`

---

## Summary

| Aspect | Value |
|--------|-------|
| Data type | JSONB array |
| Default | `'[]'::jsonb` |
| Nullable | `NOT NULL` |
| Structure | Array of attachment objects |
| Required fields | id, signed_id, filename, content_type, byte_size, download_url, icon |
| Optional fields | preview_url |
| Tables using | 8 (comments, discussions, discussion_templates, groups, outcomes, polls, stances, users) |
| API output | Always `[]` for empty, never `null` |
