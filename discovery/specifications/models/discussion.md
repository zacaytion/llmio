# Discussion Model Specification

**Generated:** 2026-02-01
**Source Files:**
- `/app/models/discussion.rb`
- `/app/models/discussion_reader.rb`
- `/app/models/null_discussion.rb`
- `/discovery/schemas/database_schema.md`

---

## Overview

The Discussion model represents threaded conversation containers within groups. Discussions can contain comments, polls, and track per-user read state via DiscussionReader. Discussions support rich text, mentions, tags, and can be templates.

---

## Attributes

### Core Identity

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `id` | serial | auto | NO | PK | Primary key |
| `title` | string(255) | - | YES | max 150, presence required | Discussion title |
| `description` | text | - | YES | max via AppConfig | Rich text body |
| `description_format` | string(10) | "md" | NO | "md" or "html" | Description format |
| `key` | string(255) | - | YES | UNIQUE | Public URL key (8 chars) |

### Relationships

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `group_id` | integer | - | YES | FK to groups | Container group (null = direct) |
| `author_id` | integer | - | YES | FK to users | Discussion creator |
| `closer_id` | integer | - | YES | FK to users | User who closed |

### Privacy & Status

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `private` | boolean | true | NO | - | Visibility flag |
| `closed_at` | datetime | - | YES | - | When discussion was closed |
| `discarded_at` | datetime | - | YES | - | Soft delete timestamp |
| `discarded_by` | integer | - | YES | FK to users | Who deleted |
| `pinned_at` | datetime | - | YES | - | Pinned timestamp |

### Template Settings

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `template` | boolean | false | NO | - | Is a template |
| `discussion_template_id` | integer | - | YES | FK | Source template |
| `discussion_template_key` | string | - | YES | - | Template key reference |

### Thread Settings

| Column | Type | Default | Null | Constraints | Description |
|--------|------|---------|------|-------------|-------------|
| `max_depth` | integer | 2 | - | - | Reply nesting depth |
| `newest_first` | boolean | false | - | - | Sort order |

### Activity Tracking

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `last_activity_at` | datetime | - | YES | Last event time |
| `last_comment_at` | datetime | - | YES | Last comment time |
| `first_sequence_id` | integer | 0 | - | First event sequence |
| `last_sequence_id` | integer | 0 | - | Latest event sequence |
| `items_count` | integer | 0 | - | Event count |
| `ranges_string` | string | - | YES | Compact sequence ranges |

### Counters

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `seen_by_count` | integer | 0 | Unique viewers |
| `members_count` | integer | - | Participant count |
| `closed_polls_count` | integer | 0 | Closed polls |
| `anonymous_polls_count` | integer | 0 | Anonymous polls |
| `versions_count` | integer | 0 | Paper trail versions |

### Other Fields

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `importance` | integer | 0 | Priority/importance |
| `content_locale` | string | - | Content locale |
| `iframe_src` | string(255) | - | Embedded iframe URL |
| `tags` | string[] | [] | Tag array |
| `attachments` | jsonb | [] | Rich text attachments |
| `link_previews` | jsonb | [] | Cached link previews |
| `info` | jsonb | {} | Extensible metadata |

---

## Validations

| Field | Validation | Condition |
|-------|------------|-----------|
| `title` | presence required | always |
| `title` | max length 150 | always |
| `description` | max length via AppConfig | always |
| `group` | presence required | always |
| `author` | presence required | always |
| `private` | must be private | if group is `private_discussions_only?` |
| `private` | must be public | if group is `public_discussions_only?` |
| `description_format` | inclusion in ['html', 'md'] | always |

**Custom Validation:**
```ruby
validate :privacy_is_permitted_by_group
# Ensures discussion privacy matches group settings
```

**Confidence: HIGH** - Validations directly extracted from model code.

---

## Associations

### Belongs To

| Association | Class | Options | Description |
|-------------|-------|---------|-------------|
| `group` | Group | - | Container group |
| `author` | User | - | Discussion creator |
| `user` | User | foreign_key: 'author_id' | Alias for author |
| `closer` | User | foreign_key: 'closer_id' | User who closed |

### Has Many

| Association | Class | Options | Description |
|-------------|-------|---------|-------------|
| `polls` | Poll | dependent: :destroy | Discussion polls |
| `active_polls` | Poll | `-> { where(closed_at: nil) }` | Open polls |
| `comments` | Comment | dependent: :destroy | Discussion comments |
| `commenters` | User | through: :comments, `-> { uniq }` | Unique commenters |
| `documents` | Document | as: :model, dependent: :destroy | Attached documents |
| `poll_documents` | Document | through: :polls | Poll documents |
| `comment_documents` | Document | through: :comments | Comment documents |
| `items` | Event | `-> { includes(:user) }`, dependent: :destroy | Timeline events |
| `discussion_readers` | DiscussionReader | dependent: :destroy | Read tracking |
| `readers` | User | through: :discussion_readers, `-> { merge(active) }` | Users with access |

### Concern Associations (HasEvents, HasMentions, Reactable, HasTags, Translatable)

| Association | Through | Description |
|-------------|---------|-------------|
| `events` | HasEvents | Eventable events |
| `notifications` | HasEvents | Through events |
| `reactions` | Reactable | Emoji reactions |
| `translations` | Translatable | Content translations |
| `tasks` | HasRichText | Embedded tasks |

---

## Scopes

```ruby
scope :visible_to_public, -> { kept.where(private: false) }
scope :not_visible_to_public, -> { kept.where(private: true) }
scope :is_open, -> { kept.where(closed_at: nil) }
scope :is_closed, -> { kept.where('closed_at is not null') }
scope :recent, -> { where('last_activity_at > ?', 6.weeks.ago) }

scope :in_organisation, ->(group) {
  includes(:author).where(group_id: group.id_and_subgroup_ids)
}

scope :last_activity_after, ->(time) { where('last_activity_at > ?', time) }
scope :order_by_latest_activity, -> { order(last_activity_at: :desc) }
scope :order_by_pinned_then_latest_activity, -> { order('pinned_at, last_activity_at DESC') }

scope :search_for, ->(q) { kept.where('discussions.title ilike ?', "%#{q}%") }

scope :dangling, -> {
  joins('left join groups g on discussions.group_id = g.id')
    .where('group_id is not null and g.id is null')
}

# From HasTimeframe concern
scope :within, ->(since, till, field = nil) {
  where("#{table_name}.#{field || :created_at} BETWEEN ? AND ?",
        since || 100.years.ago, till || 100.years.from_now)
}
```

---

## Callbacks

### After Create
- `set_last_activity_at_to_created_at` - Initializes `last_activity_at` from `created_at`

### After Destroy
- `drop_sequence_id_sequence` - Cleans up PostgreSQL sequence

### Before Save (from concerns)
- `sanitize_description!` - HTML sanitization (HasRichText)
- `update_content_locale` - Language detection (HasRichText)
- `build_attachments` - Build attachment metadata (HasRichText)
- `sanitize_link_previews` - Sanitize preview data (HasRichText)

### After Save (from concerns)
- `parse_and_update_tasks_description!` - Extract tasks from body (HasRichText)
- `update_group_tags` - Update tag counts (HasTags)

### After Discard (Discard::Model)
- Discards associated tasks

---

## Instance Methods

### Member Access

```ruby
def members
  # Returns users who are group members OR discussion guests
  User.active
    .joins("LEFT OUTER JOIN discussion_readers dr ON dr.discussion_id = #{id || 0} AND dr.user_id = users.id")
    .joins("LEFT OUTER JOIN memberships m ON m.user_id = users.id AND m.group_id = #{group_id || 0}")
    .where('(m.id IS NOT NULL AND m.revoked_at IS NULL) OR
            (dr.id IS NOT NULL AND dr.guest = TRUE AND dr.revoked_at IS NULL)')
end

def admins
  # Returns group admins OR discussion admins
  User.active
    .joins(...)
    .where('(m.admin = TRUE AND m.id IS NOT NULL AND m.revoked_at IS NULL) OR
            (dr.admin = TRUE AND dr.id IS NOT NULL AND dr.revoked_at IS NULL)')
end

def guests
  # Returns users who are discussion guests but NOT group members
  User.active
    .joins(...)
    .where('(m.id IS NULL OR m.revoked_at IS NOT NULL) AND
            (dr.id IS NOT NULL AND dr.guest = TRUE AND dr.revoked_at IS NULL)')
end

def guest_ids
  guests.pluck(:id)
end

def existing_member_ids
  reader_ids
end
```

### Guest Management

```ruby
def add_guest!(user, inviter)
  if (dr = discussion_readers.find_by(user: user))
    dr.update(guest: true, inviter: inviter)
  else
    discussion_readers.create!(
      user: user,
      inviter: inviter,
      guest: true,
      volume: DiscussionReader.volumes[:normal]
    )
  end
end

def add_admin!(user, inviter)
  if (dr = discussion_readers.find_by(user: user))
    dr.update(inviter: inviter, admin: true)
  else
    discussion_readers.create!(
      user: user,
      inviter: inviter,
      admin: true,
      volume: DiscussionReader.volumes[:normal]
    )
  end
end
```

### Sequence & Range Methods

```ruby
def ranges
  RangeSet.parse(ranges_string)
end

def first_sequence_id
  Array(ranges.first).first.to_i
end

def last_sequence_id
  Array(ranges.last).last.to_i
end

def ranges_string
  # Lazy initialization - updates if nil
  update_sequence_info! if self[:ranges_string].nil?
  self[:ranges_string]
end

def update_sequence_info!
  sequence_ids = discussion.items.order(:sequence_id).pluck(:sequence_id).compact
  discussion.ranges_string = RangeSet.serialize(RangeSet.reduce(RangeSet.ranges_from_list(sequence_ids)))
  discussion.last_activity_at = find_last_activity_at
  update_columns(
    items_count: sequence_ids.count,
    ranges_string: discussion.ranges_string,
    last_activity_at: discussion.last_activity_at
  )
end

def find_last_activity_at
  [
    comments.kept.order('created_at desc'),
    polls.kept.order('created_at desc'),
    Outcome.where(poll_id: poll_ids).order('created_at desc'),
    Stance.latest.where(poll_id: poll_ids).order('updated_at, created_at desc'),
    Discussion.where(id: id)
  ].map { |rel| rel.first&.created_at }.compact.max
end
```

### Privacy Methods

```ruby
def public?
  !private
end
```

### Version Tracking

```ruby
def is_new_version?
  (%w[title description private] & changes.keys).any?
end
```

### Body Aliases

```ruby
def body
  description
end

def body=(val)
  self.description = val
end

def body_format
  description_format
end

def body_format=(val)
  self.description_format = val
end
```

### Event Methods (HasCreatedEvent)

```ruby
def created_event_kind
  :new_discussion
end

def created_event
  events.find_by(kind: created_event_kind)
end
```

### Null Object Handling

```ruby
def group
  super || NullGroup.new
end

def author
  super || AnonymousUser.new
end

def discussion
  self  # SelfReferencing
end
```

### Delegate Methods

```ruby
delegate :name, to: :group, prefix: :group
delegate :name, to: :author, prefix: :author
delegate :users, to: :group, prefix: :group
delegate :full_name, to: :group, prefix: :group
delegate :email, to: :author, prefix: :author
delegate :name_and_email, to: :author, prefix: :author
delegate :locale, to: :author
```

---

## Counter Cache Definitions

```ruby
define_counter_cache(:closed_polls_count) { |d| d.polls.closed.count }
define_counter_cache(:versions_count) { |d| d.versions.count }
define_counter_cache(:seen_by_count) { |d| d.discussion_readers.where('last_read_at is not null').count }
define_counter_cache(:members_count) { |d| d.discussion_readers.where('revoked_at is null').count }
define_counter_cache(:anonymous_polls_count) { |d| d.polls.where(anonymous: true).count }

update_counter_cache :group, :discussions_count
update_counter_cache :group, :public_discussions_count
update_counter_cache :group, :open_discussions_count
update_counter_cache :group, :closed_discussions_count
update_counter_cache :group, :closed_polls_count
```

---

## Search Indexing

```ruby
def self.pg_search_insert_statement(id: nil, author_id: nil)
  # Inserts into pg_search_documents table
  # Content: title + description + author name (HTML stripped)
  # Links: group_id, discussion_id, author_id
  # Indexed: ts_content tsvector
end
```

---

## Concerns Included

| Concern | Purpose |
|---------|---------|
| `CustomCounterCache::Model` | Counter cache definitions |
| `ReadableUnguessableUrls` | 8-char key generation |
| `Translatable` | Translation support |
| `Reactable` | Emoji reactions |
| `HasTimeframe` | Time-based scopes |
| `HasEvents` | Event associations |
| `HasMentions` | @mention extraction |
| `MessageChannel` | Real-time pub/sub |
| `SelfReferencing` | `discussion` and `discussion_id` |
| `HasCreatedEvent` | Created event tracking |
| `HasRichText` | Rich text with sanitization |
| `HasTags` | Tag management |
| `Discard::Model` | Soft delete |
| `Searchable` | Full-text search |
| `DiscussionExportRelations` | Export query optimizations |

---

## Paper Trail Tracking

Tracked fields:
- `title`
- `description`
- `description_format`
- `private`
- `group_id`
- `author_id`
- `tags`
- `closed_at`
- `closer_id`
- `attachments`

---

## Indexes

| Columns | Type | Notes |
|---------|------|-------|
| `key` | UNIQUE | |
| `group_id` | INDEX | |
| `author_id` | INDEX | |
| `last_activity_at` | INDEX (desc) | |
| `created_at` | INDEX | |
| `private` | INDEX | |
| `tags` | GIN | Array search |
| `discarded_at` | PARTIAL | where IS NULL |
| `template` | PARTIAL | where IS TRUE |

---

## DiscussionReader Model

Per-user read state and participation for discussions.

### Attributes

| Column | Type | Default | Null | Description |
|--------|------|---------|------|-------------|
| `id` | serial | auto | NO | Primary key |
| `discussion_id` | integer | - | NO | FK to discussions |
| `user_id` | integer | - | NO | FK to users |
| `inviter_id` | integer | - | YES | FK to users (who invited) |
| `revoker_id` | integer | - | YES | FK to users (who revoked) |
| `last_read_at` | datetime | - | YES | Last read time |
| `last_read_sequence_id` | integer | 0 | NO | Last read event |
| `read_ranges_string` | string | - | YES | Compact read state |
| `volume` | integer | 2 | NO | Notification volume |
| `participating` | boolean | false | NO | Active participant |
| `admin` | boolean | false | NO | Discussion admin |
| `guest` | boolean | false | NO | Guest (not group member) |
| `token` | string | - | YES | Invitation token (UNIQUE) |
| `accepted_at` | datetime | - | YES | Invitation accepted |
| `revoked_at` | datetime | - | YES | Access revoked |
| `dismissed_at` | datetime | - | YES | Dismissed from inbox |

### Scopes

```ruby
scope :active, -> { where('discussion_readers.revoked_at IS NULL') }
scope :guests, -> { active.where('discussion_readers.guest': true) }
scope :admins, -> { active.where('discussion_readers.admin': true) }
scope :redeemable, -> { guests.where('discussion_readers.accepted_at IS NULL') }
scope :redeemable_by, ->(user_id) {
  redeemable.joins(:user).where('user_id = ? OR users.email_verified = false', user_id)
}
```

### Key Methods

```ruby
def self.for(user:, discussion:)
  # Find or initialize reader with appropriate volume default
  if user&.is_logged_in?
    find_or_initialize_by(user_id: user.id, discussion_id: discussion.id) do |dr|
      m = user.memberships.find_by(group_id: discussion.group_id)
      dr.volume = (m && m.volume) || 'normal'
    end
  else
    new(discussion: discussion)
  end
end

def update_reader(ranges: nil, volume: nil, participate: false, dismiss: false)
  viewed!(ranges, persist: false)     if ranges
  set_volume!(volume, persist: false) if volume && (volume != :loud || user.email_on_participation?)
  dismiss!(persist: false)            if dismiss
  save!                               if changed?
  self
end

def viewed!(ranges = [], persist: true)
  mark_as_read(ranges) unless has_read?(ranges)
  assign_attributes(last_read_at: Time.now)
  save if persist
end

def has_read?(ranges = [])
  RangeSet.includes?(read_ranges, ranges)
end

def read_ranges
  RangeSet.parse(read_ranges_string)
end

def unread_ranges
  RangeSet.subtract_ranges(discussion.ranges, read_ranges)
end

def read_items_count
  RangeSet.length(read_ranges)
end

def unread_items_count
  RangeSet.length(unread_ranges)
end

def computed_volume
  # Returns explicit volume, membership volume, or 'normal'
  if persisted?
    volume || membership&.volume || 'normal'
  else
    membership.volume
  end
end
```

### Indexes

| Columns | Type | Notes |
|---------|------|-------|
| `(user_id, discussion_id)` | UNIQUE | |
| `token` | UNIQUE | |
| `discussion_id` | INDEX | |
| `guest` | PARTIAL | where = true |
| `inviter_id` | PARTIAL | where NOT NULL |

---

## Uncertainties

1. **ranges_string format** - Compact encoding of sequence ID ranges, parsed by RangeSet utility
2. **iframe_src field** - Purpose and security implications of embedded iframes unclear
3. **importance field** - Usage for prioritization unclear
4. **SequenceService** - PostgreSQL sequence management for thread ordering

**Confidence Level:** HIGH for core functionality, MEDIUM for sequence/range handling details.
