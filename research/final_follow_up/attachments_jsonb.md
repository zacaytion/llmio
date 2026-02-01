# Attachments JSONB - Follow-up Analysis

## Executive Summary

Third-party discovery findings **resolve** the inconsistency documented in our research. Their investigation is thorough and well-evidenced. However, a few minor areas require clarification or additional investigation.

---

## Discrepancy Analysis

### Resolved: Default Value Inconsistency

| Source | Claimed Default | Status |
|--------|-----------------|--------|
| Our research (database.md) | `[]` (empty array) | Correct |
| Our research (review correction) | `'{}'::jsonb` (empty object) | **Incorrect** - was overcorrected |
| Third-party discovery | `[]` (empty array) | **Confirmed correct** |
| Schema dump (schema_dump.sql) | `'[]'::jsonb` | **Ground truth** |

**Resolution**: The third-party discovery correctly identifies that:
1. Initial migrations (March 2019) used `{}` (object)
2. Migration `20190926001607_change_attachments_default_to_array.rb` changed default to `[]`
3. Current schema.rb shows `default: []` for all 8 tables

Our review document overcorrected to `'{}'::jsonb` based on stale migration files rather than current schema.

---

## Contradictions Requiring Resolution

### None Found

Third-party findings align with:
- PostgreSQL schema dump (`research/schema_dump.sql`)
- Rails schema.rb (`orig/loomio/db/schema.rb`)
- Model code (`orig/loomio/app/models/concerns/has_rich_text.rb`)
- Frontend models (all use `attachments: []`)

---

## Areas of Incomplete/Unclear Findings

### LOW: Attachment Object Schema

**Issue**: Third-party discovery documents the source of attachment objects but does not provide a complete field schema.

**Evidence from code** (`orig/loomio/app/models/concerns/has_rich_text.rb:92-102`):
```ruby
def build_attachments
  self[:attachments] = files.map do |file|
    i = file.blob.slice(:id, :filename, :content_type, :byte_size)
    i.merge!({ preview_url: Rails.application.routes.url_helpers.rails_representation_path(...) }) if file.representable?
    i.merge!({ download_url: Rails.application.routes.url_helpers.rails_blob_path(file, only_path: true) })
    i.merge!({ icon: attachment_icon(file.content_type || file.filename) })
    i.merge!({ signed_id: file.signed_id })
    i
  end
end
```

**Questions for third party**:
1. Is `preview_url` optional (only for representable files)?
2. What determines if a file is "representable"? (Image/video types?)
3. What are the possible values for the `icon` field?

**Investigation targets**:
- [ ] `orig/loomio/config/app_config.yml` - Check `doctypes` configuration for icon mapping
- [ ] Test database to examine actual attachment JSON structures

---

### LOW: attachments_count Column

**Issue**: Third-party mentions `attachments_count` counter cache on comments table but does not investigate its relationship to the JSONB column.

**Evidence from schema**:
```ruby
# orig/loomio/db/schema.rb:185
t.integer "attachments_count", default: 0, null: false
```

**Questions**:
1. Is `attachments_count` maintained automatically or manually?
2. Does it count JSONB array elements or associated files?
3. Is it used for query optimization?

**Investigation targets**:
- [ ] `orig/loomio/app/models/comment.rb` - Check for counter_cache declarations
- [ ] Check if `attachments_count` is kept in sync with `attachments.length`

---

### LOW: Legacy Attachment Tables

**Issue**: Schema shows legacy `attachments` and `documents` tables separate from JSONB columns.

**Evidence from** `research/schema_investigation.md`:
```
| documents | Legacy file attachments |
| attachments | Legacy comment attachments |
```

**Questions**:
1. Are legacy tables still in use?
2. Was data migrated from tables to JSONB columns?
3. Should Go rewrite support both patterns?

**Investigation targets**:
- [ ] `orig/loomio/db/migrate/` - Look for data migration from legacy tables
- [ ] `orig/loomio/app/models/attachment.rb` - Check if model exists and is used

---

## Specific Questions for Third Party

1. **Icon field values**: Can you provide the complete list of icon values from `AppConfig.doctypes`?

2. **ActiveStorage integration**: How does the `files` association (has_many_attached) relate to the `attachments` JSONB? Is JSONB a denormalization of ActiveStorage blob data?

3. **Representable check**: The code shows `file.representable?` - what file types return true for this?

4. **Counter cache**: Is `attachments_count` on comments table still used and maintained?

---

## Priority Summary

| Item | Priority | Rationale |
|------|----------|-----------|
| Default value resolved | N/A | No action needed |
| Attachment object schema | LOW | Can be derived from code; affects Go struct definition |
| attachments_count column | LOW | Minor counter cache; may not need migration |
| Legacy tables | LOW | Likely deprecated; confirm before excluding |

---

## Files Requiring Investigation

| File | Purpose | Priority |
|------|---------|----------|
| `orig/loomio/config/app_config.yml` | Icon type mapping | LOW |
| `orig/loomio/app/models/comment.rb` | Counter cache config | LOW |
| `orig/loomio/app/models/attachment.rb` | Legacy model status | LOW |
| `orig/loomio/db/migrate/*attachments*` | Migration history | LOW |

---

## Verification Commands

```bash
# Find all attachment-related migrations
ls -la orig/loomio/db/migrate/*attachment*

# Check if legacy Attachment model exists
cat orig/loomio/app/models/attachment.rb

# Find doctypes configuration
grep -r "doctypes" orig/loomio/config/

# Check counter cache on Comment
grep -n "attachments_count" orig/loomio/app/models/comment.rb
```
