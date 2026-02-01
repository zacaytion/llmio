# Attachments JSONB Default Value Investigation

## Summary

**The database default for `attachments` JSONB columns is `[]` (empty array).**

All tables use the same default value - there is no inconsistency between tables.

## Ground Truth Answers

### 1. What is the actual database default for attachments columns?

**Answer: `[]` (empty JSON array)**

Evidence from `/Users/z/Code/loomio/db/schema.rb`:

| Table | Line | Definition |
|-------|------|------------|
| comments | 189 | `t.jsonb "attachments", default: [], null: false` |
| discussion_templates | 259 | `t.jsonb "attachments", default: [], null: false` |
| discussions | 298 | `t.jsonb "attachments", default: [], null: false` |
| groups | 467 | `t.jsonb "attachments", default: [], null: false` |
| outcomes | 645 | `t.jsonb "attachments", default: [], null: false` |
| polls | 773 | `t.jsonb "attachments", default: [], null: false` |
| stances | 866 | `t.jsonb "attachments", default: [], null: false` |
| users | 1036 | `t.jsonb "attachments", default: [], null: false` |

### 2. Do different tables have different defaults?

**Answer: No - all 8 tables use `default: []`**

All tables are consistent. The documentation discrepancy likely arose from:
- Initial migration used `{}` (object)
- Later migration changed to `[]` (array)

### 3. What does the application code expect (array or object)?

**Answer: Array**

Evidence:

**Backend (Ruby)** - `/Users/z/Code/loomio/app/models/concerns/has_rich_text.rb:92-102`:
```ruby
def build_attachments
  return true unless self.class.column_names.include?('attachments')
  self[:attachments] = files.map do |file|
    i = file.blob.slice(:id, :filename, :content_type, :byte_size)
    # ... builds array of attachment objects
  end
end
```

The `files.map` call explicitly creates an array of attachment objects.

**Frontend (JavaScript)** - All model files define `attachments: []` as default:

| File | Line |
|------|------|
| `vue/src/shared/models/comment_model.js` | 26 |
| `vue/src/shared/models/discussion_model.js` | 57 |
| `vue/src/shared/models/poll_model.js` | 75 |
| `vue/src/shared/models/user_model.js` | 23 |
| `vue/src/shared/models/group_model.js` | 44 |
| `vue/src/shared/models/stance_model.js` | 24 |
| `vue/src/shared/models/outcome_model.js` | 23 |
| `vue/src/shared/models/discussion_template_model.js` | 23 |
| `vue/src/shared/models/poll_template_model.js` | 42 |
| `vue/src/shared/models/null_discussion_model.js` | 25 |
| `vue/src/shared/models/null_group_model.js` | 30 |

### 4. How do serializers output empty attachments?

**Answer: Direct attribute pass-through (outputs `[]`)**

Serializers use simple attribute declaration:
```ruby
# /Users/z/Code/loomio/app/serializers/comment_serializer.rb:13
attributes :id,
           ...
           :attachments,
```

This passes through the database value as-is, which is `[]` for empty attachments.

## Migration History

The change from `{}` to `[]` is documented in migrations:

1. **Initial addition (2019-03-26)**: Used `{}` (object)
   - `/Users/z/Code/loomio/db/migrate/20190326005806_add_attachments_to_comments.rb:3`
   - `/Users/z/Code/loomio/db/migrate/20190326215735_add_attachments_to_rich_text_models.rb:3-8`

2. **Default change (2019-09-26)**: Changed to `[]` (array)
   - `/Users/z/Code/loomio/db/migrate/20190926001607_change_attachments_default_to_array.rb:3-9`
   ```ruby
   change_column :discussions, :attachments, :jsonb, null: false, default: []
   change_column :groups, :attachments, :jsonb, null: false, default: []
   # ... etc for all tables
   ```

3. **New tables (2023+)**: Created with `[]` from start
   - `/Users/z/Code/loomio/db/migrate/20230731005643_create_discussion_templates_table.rb:16`

## UI Usage Confirms Array Expectation

Template components use array methods:

```pug
# /Users/z/Code/loomio/vue/src/components/thread/attachment_list.vue:15-16
.attachment-list.mb-2(v-if="attachments && attachments.length")
  attachment-list-item(v-for="attachment in attachments", ...)
```

Backend views also iterate as arrays:

```haml
# /Users/z/Code/loomio/app/views/event_mailer/common/_attachments.html.haml:14
- if resource.attachments.any?
```

## Conclusion

The documentation inconsistency originated from the historical migration from `{}` to `[]`. The **current and correct default is `[]` (empty array)**, and this is consistent across:
- Database schema (all 8 tables)
- Backend code (HasRichText concern)
- Frontend code (all model defaultValues)
- Serializers (pass-through)
- View templates (array iteration)
