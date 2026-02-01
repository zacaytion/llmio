# Discussions Domain - Models

**Generated:** 2026-02-01
**Domain:** Discussions, Comments, and DiscussionReaders
**Confidence:** 5/5 (High - Based on direct model inspection)

---

## Overview

The discussions domain centers on three primary models:
- **Discussion** - The main thread container, associated with a group
- **Comment** - Individual messages within a discussion thread
- **DiscussionReader** - Per-user tracking of read state, volume preferences, and guest access

---

## Discussion Model

**File:** `/app/models/discussion.rb`

### Core Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | Integer | Primary key |
| `key` | String | Unguessable URL key (via ReadableUnguessableUrls) |
| `title` | String | Discussion title (max 150 chars) |
| `description` | Text | Rich text body content |
| `description_format` | String | Either 'html' or 'md' |
| `private` | Boolean | If true, only members can see |
| `group_id` | Integer | FK to Group (nullable for direct discussions) |
| `author_id` | Integer | FK to User who created |
| `closed_at` | Datetime | When discussion was closed (null = open) |
| `closer_id` | Integer | FK to User who closed |
| `pinned_at` | Datetime | When pinned (null = not pinned) |
| `last_activity_at` | Datetime | Timestamp of most recent activity |
| `items_count` | Integer | Counter cache of thread events |
| `ranges_string` | String | Serialized range set of sequence IDs |
| `max_depth` | Integer | Maximum nesting depth for replies |
| `newest_first` | Boolean | Thread ordering preference |
| `discarded_at` | Datetime | Soft delete timestamp |
| `discarded_by` | Integer | FK to User who deleted |

### Included Concerns

The Discussion model uses extensive mixins:
- **ReadableUnguessableUrls** - Generates `key` for URL identification
- **HasRichText** - Rich text handling with sanitization for `description`
- **HasMentions** - @mention parsing for `description`
- **HasEvents** - Activity event tracking
- **HasTags** - Tag association
- **Translatable** - Translation support
- **Reactable** - Emoji reactions
- **Searchable** - Full-text search via pg_search
- **Discard::Model** - Soft delete functionality
- **HasCreatedEvent** - Creates new_discussion Event on creation
- **MessageChannel** - Real-time update channels

### Associations

```
Discussion
  belongs_to :group (optional for direct discussions)
  belongs_to :author (User)
  belongs_to :closer (User, optional)
  has_many :comments (dependent: destroy)
  has_many :polls (dependent: destroy)
  has_many :items (Event - the thread timeline)
  has_many :discussion_readers (dependent: destroy)
  has_many :readers (through: discussion_readers)
  has_many :documents (polymorphic)
```

### Key Methods

**Privacy and Access:**
- `public?` - Returns true if `private` is false
- `members` - Returns all users with access (group members + guests)
- `admins` - Returns users with admin rights (group admins + discussion admins)
- `guests` - Returns users who have guest access but are not group members

**Read State:**
- `ranges` - Parses `ranges_string` into array of sequence ID ranges
- `first_sequence_id` / `last_sequence_id` - Boundary sequence IDs
- `update_sequence_info!` - Recalculates items_count, ranges_string, last_activity_at

**Guest Management:**
- `add_guest!(user, inviter)` - Creates DiscussionReader with guest: true
- `add_admin!(user, inviter)` - Creates DiscussionReader with admin: true

### Counter Caches

The Discussion model maintains several counter caches:
- `closed_polls_count` - Number of closed polls
- `versions_count` - Number of paper trail versions
- `seen_by_count` - Number of readers with last_read_at set
- `members_count` - Number of active discussion_readers

It also updates counters on the parent Group:
- `discussions_count`
- `public_discussions_count`
- `open_discussions_count`
- `closed_discussions_count`

### Validations

- Title, group, and author are required
- Title max length: 150 characters
- Description max length: configured via AppConfig.app_features[:max_message_length]
- Privacy must match group settings (validates public discussions in private-only groups, etc.)

### Paper Trail Versioning

Tracks changes to: title, description, description_format, private, group_id, author_id, tags, closed_at, closer_id, attachments

---

## Comment Model

**File:** `/app/models/comment.rb`

### Core Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | Integer | Primary key |
| `discussion_id` | Integer | FK to Discussion |
| `user_id` | Integer | FK to User (author) |
| `body` | Text | Comment content |
| `body_format` | String | Either 'html' or 'md' |
| `parent_id` | Integer | FK to polymorphic parent |
| `parent_type` | String | Parent model type (Discussion, Comment, Stance) |
| `edited_at` | Datetime | Last edit timestamp |
| `discarded_at` | Datetime | Soft delete timestamp |
| `discarded_by` | Integer | FK to User who deleted |

### Included Concerns

- **HasRichText** - Rich text handling with sanitization for `body`
- **HasMentions** - @mention parsing for `body`
- **HasEvents** - Activity event tracking
- **Translatable** - Translation support
- **Reactable** - Emoji reactions
- **Searchable** - Full-text search
- **Discard::Model** - Soft delete
- **HasCreatedEvent** - Creates new_comment Event

### Threading Model

Comments use a **polymorphic parent association** for threading:
- `parent_type` can be 'Discussion', 'Comment', or 'Stance'
- `parent_id` references the parent record
- If no parent is specified, defaults to the Discussion itself

The threading allows comments to be direct replies to:
1. The discussion context (top-level comments)
2. Other comments (nested replies)
3. Stance records (replies to votes)

### Key Methods

- `parent_event` - Returns the Event record of the parent for positioning
- `is_most_recent?` - Checks if this is the last comment in discussion
- `is_edited?` - Returns true if edited_at is present
- `should_pin` - Suggests pinning if body contains headings (h1-h3)

### Validations

- User required (unless discarded)
- Parent must belong to same discussion
- Must have either body content or file attachment

---

## DiscussionReader Model

**File:** `/app/models/discussion_reader.rb`

### Purpose

DiscussionReader tracks **per-user read state** and **notification preferences** for each discussion. It is the bridge between users and discussions.

### Core Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | Integer | Primary key |
| `discussion_id` | Integer | FK to Discussion |
| `user_id` | Integer | FK to User |
| `volume` | Integer (enum) | Notification level (mute=0, quiet=1, normal=2, loud=3) |
| `last_read_at` | Datetime | When user last viewed |
| `last_read_sequence_id` | Integer | Last sequence_id read (legacy) |
| `read_ranges_string` | String | Serialized ranges of read sequence IDs |
| `dismissed_at` | Datetime | When user dismissed from inbox |
| `guest` | Boolean | True if user has guest access |
| `admin` | Boolean | True if user has admin rights |
| `inviter_id` | Integer | FK to User who invited |
| `accepted_at` | Datetime | When invitation was accepted |
| `revoked_at` | Datetime | When access was revoked |
| `token` | String | Unique token for guest access links |

### Volume System

The volume enum controls notification behavior:
- **mute (0)** - No notifications
- **quiet (1)** - In-app notifications only
- **normal (2)** - Email and in-app notifications
- **loud (3)** - All notifications plus replies

Volume cascades: DiscussionReader volume > Membership volume > User default

### Read State Tracking

Read state uses a **range-based system** to efficiently track which events have been read:

- `read_ranges_string` stores serialized ranges like "1-5,8-10,15-20"
- `read_ranges` parses the string into array format like [[1,5], [8,10], [15,20]]
- `has_read?(ranges)` checks if given ranges are contained
- `mark_as_read(ranges)` adds new ranges and reduces/merges overlapping

The range system is efficient because:
1. It compresses sequential IDs (1,2,3,4,5 becomes 1-5)
2. Operations are set-based rather than record-based
3. Syncs cleanly between client and server

### Key Methods

**Factory Methods:**
- `DiscussionReader.for(user:, discussion:)` - Find or initialize a reader for a user
- `DiscussionReader.for_model(model, actor)` - Get reader for a model's discussion

**Read State:**
- `viewed!(ranges, persist:)` - Mark ranges as read and update last_read_at
- `unread_ranges` - Calculate which ranges are unread
- `read_items_count` / `unread_items_count` - Count read/unread items

**Inbox Management:**
- `dismiss!(persist:)` - Mark as dismissed from inbox
- `recall!(persist:)` - Un-dismiss back to inbox

**Volume:**
- `computed_volume` - Returns effective volume (own or inherited from membership)
- `set_volume!(volume, persist:)` - Set notification volume

### Scopes

- `active` - Where revoked_at IS NULL
- `guests` - Active readers with guest: true
- `admins` - Active readers with admin: true
- `redeemable` - Guests who have not accepted invitation

---

## NullDiscussion

**File:** `/app/models/null_discussion.rb`

A null object pattern implementation for when a discussion does not exist. Returns safe default values to avoid nil checks:
- Returns nil for id, key, content_locale, etc.
- Returns empty arrays for member_ids
- Returns empty relations for admins, members, readers

Used to simplify code that might otherwise need to check for nil discussions.

---

## Event Integration

Discussions and comments create Events for the activity timeline:

**Discussion Events:**
- `new_discussion` - Created when discussion is started
- `discussion_edited` - Created when title/description changes
- `discussion_closed` / `discussion_reopened` - State changes
- `discussion_moved` - When moved to another group
- `discussion_forked` - When comments are forked to new discussion
- `discussion_announced` - When users are invited to discussion

**Comment Events:**
- `new_comment` - Created when comment is posted
- `comment_edited` - Created when comment is updated
- `comment_replied_to` - (Notification event for reply authors)

Events have:
- `sequence_id` - Unique sequence within discussion
- `position_key` - Hierarchical position string for nested display
- `parent_id` - Reference to parent event for threading
- `depth` - Nesting depth level

---

## max_depth and Reply Nesting

The `max_depth` attribute on Discussion controls how deeply replies can nest:
- Events check the discussion's max_depth when determining their parent
- When an event would exceed max_depth, it's reparented to the max-depth ancestor
- Changing max_depth triggers `RepairThreadWorker` to reorganize the thread

The `max_depth_adjusted_parent` method on Event handles this:

```
PSEUDO-CODE:
find_parent_event (based on kind - comments reply to their parent event)
if discussion.max_depth equals parent.depth
  return parent's parent (flatten one level)
else
  return original parent
```

This allows admins to flatten overly-deep threads while preserving logical relationships.
