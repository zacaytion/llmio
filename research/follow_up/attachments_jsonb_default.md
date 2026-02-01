# Attachments JSONB Default - Follow-up Investigation Brief

## Discrepancy Summary

Documentation is **inconsistent** about the default value for `attachments` JSONB columns:
- Some documents say `[]` (empty array)
- Some documents say `{}` (empty object)
- Research review corrected to `'{}'::jsonb`

This affects how the Go implementation initializes attachment fields.

## Discovery Claims

**Source**: `discovery/initial/documents/models.md`

Does not explicitly specify the default value for attachments JSONB.

**Source**: `discovery/initial/synthesis/llm_context.md`

Lists attachments structure but not default value.

## Our Research Claims

**Source**: `research/investigation/database.md`

Initially stated:
> "Attachments: id, signed_id, filename, content_type, byte_size, preview_url, download_url, icon (default: `[]`)"

**Source**: `research/initial_investigation_review.md`

Correction:
> "**Actual Default:** `'{}'::jsonb` (empty object, not empty array)"

**Source**: `research/initial_meta_analysis.md`

Notes this as an unresolved three-way conflict.

## Ground Truth Needed

1. What is the actual database default for attachments columns?
2. Do different tables have different defaults?
3. What does the application code expect (array or object)?
4. How do serializers output empty attachments?

## Investigation Targets

- [ ] File: `orig/loomio/db/schema.rb` - Search for `attachments` column definitions with defaults
- [ ] Command: `grep -n "attachments.*default" orig/loomio/db/migrate/*.rb` - Check migration files
- [ ] Command: `grep -n "attachments" orig/loomio/db/schema.rb` - Find all attachments columns
- [ ] File: `orig/loomio/app/models/concerns/has_rich_text.rb` - Check attachment handling
- [ ] File: `orig/loomio/app/serializers/*_serializer.rb` - Check how empty attachments are serialized

## Priority

**LOW** - This is a minor data initialization detail. Either `[]` or `{}` can work with proper handling, but consistency matters for:
- Database migrations
- JSON parsing in Go
- API response format

## Rails Context

### JSONB Column Defaults

PostgreSQL JSONB columns can have either default:

```sql
-- Empty array default (for list of attachments)
CREATE TABLE discussions (
  attachments jsonb DEFAULT '[]'::jsonb
);

-- Empty object default (for keyed attachments or metadata)
CREATE TABLE discussions (
  attachments jsonb DEFAULT '{}'::jsonb
);
```

### Rails Migration Pattern

```ruby
# Migration with explicit default
add_column :discussions, :attachments, :jsonb, default: []

# Or
add_column :discussions, :attachments, :jsonb, default: {}
```

### Model Usage

The actual usage reveals intent:

```ruby
# Array pattern - list of file metadata
class Discussion < ApplicationRecord
  def add_attachment(file)
    self.attachments ||= []
    self.attachments << {
      id: file.blob_id,
      filename: file.filename,
      content_type: file.content_type
    }
  end
end

# Object pattern - keyed metadata
class Discussion < ApplicationRecord
  def set_attachment(key, file)
    self.attachments ||= {}
    self.attachments[key] = {
      filename: file.filename,
      content_type: file.content_type
    }
  end
end
```

### HasRichText Concern

Based on documentation, attachments are used with Tiptap editor:

```ruby
# app/models/concerns/has_rich_text.rb
module HasRichText
  extend ActiveSupport::Concern

  def attachments_for_editor
    # Returns array of attachment metadata for Tiptap
    attachments.map { |a| AttachmentSerializer.new(a) }
  end
end
```

This suggests **array** semantics (list of attachments).

## Reconciliation Hypothesis

The conflict may arise from:

1. **Database default** is `{}` (object) - defensive empty state
2. **Application code** treats it as array - `attachments ||= []`
3. **Serializers** output as array - `"attachments": []`

Ruby's `||=` and type coercion may mask the difference:
```ruby
# {} is truthy in Ruby, so ||= doesn't trigger
attachments ||= []  # Won't change {} to []

# But array methods would fail on {}
attachments.map { }  # NoMethodError if {}
```

**Action**: Verify actual schema default AND application behavior.

## Impact on Go Rewrite

For Go implementation:
- Check schema dump for actual default
- Implement consistent initialization in models
- Handle both `[]` and `{}` gracefully in deserialization
- Output `[]` (empty array) in API responses for consistency

```go
type Discussion struct {
    // Use pointer to distinguish nil (not set) from empty
    Attachments []Attachment `json:"attachments"`
}

// Ensure empty array in JSON, not null
func (d *Discussion) MarshalJSON() ([]byte, error) {
    if d.Attachments == nil {
        d.Attachments = []Attachment{}
    }
    // ... marshal
}
```
