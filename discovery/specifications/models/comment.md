# Comment Model Specification

**Generated:** 2026-02-01
**Source Files:**
- `/app/models/comment.rb`
- `/discovery/schemas/database_schema.md`

---

## Overview

The Comment model represents threaded comments within discussions. Comments can be replies to the discussion, other comments, or even stances. Comments support rich text, @mentions, reactions, and versioning via Paper Trail.

---

## Attributes

### Core Identity

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `id` | serial | auto | NO | PK | Primary key |
| `body` | text | "" | YES | - | Rich text content |
| `body_format` | string(10) | "md" | NO | "md" or "html" | Content format |

### Relationships

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `discussion_id` | integer | 0 | YES | FK to discussions | Parent discussion |
| `user_id` | integer | 0 | YES | FK to users | Comment author |
| `parent_id` | integer | - | YES | Polymorphic FK | Reply-to target |
| `parent_type` | string | - | NO | Polymorphic type | Parent model type |

### Status

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `edited_at` | datetime | - | YES | Last edit time |
| `discarded_at` | datetime | - | YES | Soft delete timestamp |
| `discarded_by` | integer | - | YES | FK to users who deleted |

### Counters

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `versions_count` | integer | 0 | Paper Trail versions |
| `comment_votes_count` | integer | 0 | Legacy likes count |
| `attachments_count` | integer | 0 | Legacy attachments count |

### Content Fields

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `content_locale` | string | - | Detected content locale |
| `attachments` | jsonb | [] | Rich text attachments |
| `link_previews` | jsonb | [] | Cached link previews |

### Timestamps

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `created_at` | datetime | - | Creation timestamp |
| `updated_at` | datetime | - | Last update |

---

## Validations

| Field | Validation | Condition |
|-------|------------|-----------|
| `user` | presence required | unless `discarded_at` |
| `body` | has content OR has attachments | unless `discarded_at` |
| `parent` | belongs to same discussion | always |
| `body_format` | inclusion in ['html', 'md'] | always |
| `body` | max length via AppConfig | always |

**Custom Validations:**
```ruby
validate :parent_comment_belongs_to_same_discussion
# Ensures parent comment/discussion has same discussion_id
# Re-parents to discussion if parent is nil

validate :has_body_or_attachment
# Comment must have body text or file attachments
# Ignores if discarded
```

**Confidence: HIGH** - Validations directly extracted from model code.

---

## Associations

### Belongs To

| Association | Class | Options | Description |
|-------------|-------|---------|-------------|
| `discussion` | Discussion | - | Parent discussion |
| `user` | User | - | Comment author |
| `parent` | Polymorphic | polymorphic: true | Reply-to target |

### Has Many

| Association | Class | Options | Description |
|-------------|-------|---------|-------------|
| `documents` | Document | as: :model, dependent: :destroy | Attached documents |

### Aliases

```ruby
alias_method :author, :user
alias_method :author=, :user=
```

### Concern Associations

| Association | Through | Description |
|-------------|---------|-------------|
| `events` | HasEvents | Eventable events |
| `notifications` | HasEvents | Through events |
| `reactions` | Reactable | Emoji reactions |
| `translations` | Translatable | Content translations |
| `tasks` | HasRichText | Embedded tasks |
| `files` | HasRichText | ActiveStorage attachments |
| `image_files` | HasRichText | ActiveStorage image attachments |

---

## Scopes

```ruby
scope :dangling, -> {
  joins('left join discussions on discussion_id = discussions.id')
    .where('discussion_id is not null and discussions.id is null')
}

scope :in_organisation, ->(group) {
  includes(:user, :discussion)
    .joins(:discussion)
    .where("discussions.group_id": group.id_and_subgroup_ids)
}
```

---

## Callbacks

### Before Validation
- `assign_parent_if_nil` - Sets parent to discussion if parent_id is nil

### Before Save (from HasRichText)
- `sanitize_body!` - HTML sanitization with whitelist
- `update_content_locale` - Language detection via CLD
- `build_attachments` - Build attachment metadata
- `sanitize_link_previews` - Sanitize preview data

### After Save (from HasRichText)
- `parse_and_update_tasks_body!` - Extract tasks from body

### After Discard (Discard::Model)
- Discards associated tasks

---

## Instance Methods

### Parent Event Resolution

```ruby
def parent_event
  # Returns the event to use as parent in thread hierarchy
  if parent.nil? && discussion.present?
    self.parent = self.discussion
    save!(validate: false)
  end

  if parent.is_a? Stance
    # Stances may have updated events, not just created
    Event.where(eventable_type: parent_type, eventable_id: parent_id)
         .where('discussion_id is not null').first
  else
    parent.created_event
  end
end

def assign_parent_if_nil
  self.parent = self.discussion if self.parent_id.nil?
end
```

### Author Methods

```ruby
def author_id
  user_id
end

def author_name
  user.name
end

def real_participant
  author
end

def user
  super || AnonymousUser.new
end
```

### Status Methods

```ruby
def is_most_recent?
  discussion.comments.last == self
end

def is_edited?
  edited_at.present?
end
```

### Content Methods

```ruby
def should_pin
  # Returns true if comment contains headings (h1, h2, h3)
  return false if body_format != "html"
  Nokogiri::HTML(self.body).css("h1,h2,h3").length > 0
end
```

### Event Methods (HasCreatedEvent)

```ruby
def created_event_kind
  :new_comment
end

def created_event
  events.find_by(kind: created_event_kind)
end
```

### Delegation

```ruby
delegate :name, to: :user, prefix: :author
delegate :author, to: :parent, prefix: :parent, allow_nil: true
delegate :group, to: :discussion
delegate :title, to: :discussion
delegate :group_id, to: :discussion, allow_nil: true
delegate :guests, to: :discussion
delegate :members, to: :discussion
```

### Identity Methods

```ruby
def title_model
  discussion
end

def poll
  nil
end

def poll_id
  nil
end
```

---

## Counter Cache Definitions

```ruby
define_counter_cache(:versions_count) { |comment| comment.versions.count }
```

---

## Search Indexing

```ruby
def self.pg_search_insert_statement(id: nil, author_id: nil, discussion_id: nil)
  # Inserts into pg_search_documents table
  # Content: body + author name (HTML stripped)
  # Links: group_id (via discussion), discussion_id, author_id
  # Filters: discarded_at IS NULL for both comment and discussion
end
```

---

## Concerns Included

| Concern | Purpose |
|---------|---------|
| `Discard::Model` | Soft delete support |
| `CustomCounterCache::Model` | Counter cache definitions |
| `Translatable` | Translation support |
| `Reactable` | Emoji reactions |
| `HasMentions` | @mention extraction |
| `HasCreatedEvent` | Created event tracking |
| `HasEvents` | Event associations |
| `HasRichText` | Rich text with sanitization |
| `Searchable` | Full-text search |

---

## Paper Trail Tracking

Tracked fields:
- `body`
- `body_format`
- `user_id`
- `discarded_at`
- `discarded_by`
- `attachments`

---

## Indexes

| Columns | Type | Notes |
|---------|------|-------|
| `discussion_id` | INDEX | |
| `(parent_type, parent_id)` | INDEX | Polymorphic lookup |

---

## Parent Types

Comments can be replies to different parent types:

| Parent Type | Description |
|-------------|-------------|
| `Discussion` | Top-level comment on discussion |
| `Comment` | Reply to another comment |
| `Stance` | Reply to a poll vote/stance |

**Note:** The parent must always belong to the same discussion as the comment.

---

## Rich Text Sanitization

From HasRichText concern, the following HTML tags and attributes are allowed:

**Allowed Tags:**
```
strong, em, b, i, p, s, code, pre, big, div, small, hr, br, span, mark,
h1, h2, h3, ul, ol, li, abbr, a, img, video, audio, blockquote, table,
thead, th, tr, td, iframe, u
```

**Allowed Attributes:**
```
href, src, alt, title, data-type, data-iframe-container, data-done,
data-mention-id, poster, controls, data-author-id, data-uid, data-checked,
data-due-on, data-color, data-remind, width, height, target, colspan,
rowspan, data-text-align
```

---

## Uncertainties

1. **Parent polymorphic type** - Stance replies behavior when parent stance is anonymous
2. **comment_votes_count / attachments_count** - Legacy fields, current usage unclear
3. **Re-parenting behavior** - When parent is deleted, comment re-parents to discussion via validation

**Confidence Level:** HIGH for core functionality, MEDIUM for edge cases around parent handling.
