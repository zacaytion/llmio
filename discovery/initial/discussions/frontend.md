# Discussions Domain - Frontend

**Generated:** 2026-02-01
**Domain:** Discussion and Comment components, models, and services
**Confidence:** 4/5 (High - Based on model and interface inspection)

---

## Overview

The frontend uses Vue 3 with LokiJS for in-memory data storage. The discussions domain includes:
- **Models** - DiscussionModel, DiscussionReaderModel, CommentModel
- **Interfaces** - Record store integrations
- **Components** - Thread display, comment forms, strand navigation

---

## Frontend Models

### DiscussionModel

**File:** `/vue/src/shared/models/discussion_model.js`

#### Default Values

```
PSEUDO-CODE:
defaultValues:
  id: null
  key: null
  private: true
  title: ''
  description: ''
  descriptionFormat: 'html'
  forkedEventIds: []
  ranges: []
  readRanges: []
  newestFirst: false
  files: null
  imageFiles: null
  attachments: []
  linkPreviews: []
  tags: []
  recipientMessage: null
  recipientAudience: null
  recipientUserIds: []
  recipientChatbotIds: []
  recipientEmails: []
  notifyRecipients: true
  groupId: null
  pinnedAt: null
  pollTemplateKeysOrIds: []
```

#### Relationships

```
PSEUDO-CODE:
relationships:
  hasMany: polls (sorted by createdAt desc, excludes discarded)
  belongsTo: group
  belongsTo: author (from users)
  belongsTo: closer (from users)
  belongsTo: translation
  hasMany: discussionReaders
```

#### Key Methods

**Read State Methods:**
- `unreadItemsCount()` - Returns items_count minus read items count
- `readItemsCount()` - Uses RangeSet.length on readRanges
- `hasRead(id)` - Checks if sequence_id is in readRanges
- `markAsRead(id)` - Adds id to readRanges and syncs to server
- `updateReadRanges()` - Throttled PATCH to server (2 second debounce)
- `unreadRanges()` - Subtracts readRanges from ranges
- `firstUnreadSequenceId()` - First unread sequence ID

**Inbox Methods:**
- `isUnread()` - True if not dismissed and has unread items
- `isDismissed()` - True if dismissedAt >= lastActivityAt
- `hasUnreadActivity()` - Combines isUnread and unreadItemsCount
- `dismiss()` - Sets dismissedAt and PATCHes server
- `recall()` - Clears dismissedAt and PATCHes server

**Volume Methods:**
- `volume()` - Returns discussionReaderVolume
- `saveVolume(volume, applyToAll)` - Saves volume, optionally to membership
- `isMuted()` - True if volume is 'mute'

**State Methods:**
- `close()` - PATCHes close endpoint
- `reopen()` - PATCHes reopen endpoint
- `savePin()` / `saveUnpin()` - Pin/unpin endpoints

**Forking:**
- `forkedEvents()` - Returns sorted events by forkedEventIds
- `forkTarget()` - Returns first forked event's model
- `moveComments()` - PATCHes move_comments with forkedEventIds

**Event Methods:**
- `createdEvent()` - Finds new_discussion event for this discussion
- `forkedEvent()` - Finds discussion_forked event if exists

**Access Methods:**
- `members()` - All users in group plus discussion readers
- `membersInclude(user)` - Check if user is member/guest
- `adminsInclude(user)` - Check if user is admin

### DiscussionReaderModel

**File:** `/vue/src/shared/models/discussion_reader_model.js`

Simple model tracking reader state:

```
PSEUDO-CODE:
defaultValues:
  discussionId: null
  userId: null
  guest: false

relationships:
  belongsTo: user
  belongsTo: discussion
```

Most reader data is merged into the discussion response, not fetched separately.

### CommentModel

**File:** `/vue/src/shared/models/comment_model.js`

#### Default Values

```
PSEUDO-CODE:
defaultValues:
  discussionId: null
  files: null
  imageFiles: null
  attachments: []
  linkPreviews: []
  body: ''
  bodyFormat: 'html'
  mentionedUsernames: []
```

#### Relationships

```
PSEUDO-CODE:
relationships:
  belongsTo: author (from users)
  belongsTo: discussion
  belongsTo: translation
```

#### Key Methods

- `createdEvent()` - Finds new_comment event for this comment
- `isReply()` - True if parentId is set
- `isBlank()` - True if body is empty
- `parent()` - Returns parent record (Comment, Discussion, or Stance)
- `reactions()` - Returns reactions for this comment
- `group()` - Delegates to discussion.group()
- `beforeDestroy()` - Removes associated events from store

---

## Record Interfaces

### DiscussionRecordsInterface

**File:** `/vue/src/shared/interfaces/discussion_records_interface.js`

Extends BaseRecordsInterface with:
- `nullModel()` - Returns NullDiscussionModel for nil cases
- `search(groupKey, fragment, options)` - Search endpoint
- `fetchInbox(options)` - Dashboard endpoint

### CommentRecordsInterface

**File:** `/vue/src/shared/interfaces/comment_records_interface.js`

Standard interface with no custom methods.

---

## RangeSet Service

**File:** `/vue/src/shared/services/range_set.js`

Utility for managing read state as ranges:

**Operations:**
- `parse(string)` - "1-5,8-10" -> [[1,5], [8,10]]
- `serialize(ranges)` - [[1,5], [8,10]] -> "1-5,8-10"
- `reduce(ranges)` - Merge overlapping ranges
- `length(ranges)` - Count total items in ranges
- `includesValue(ranges, value)` - Check if value is in any range
- `subtractRanges(whole, parts)` - Remove parts from whole
- `intersectRanges(readRanges, ranges)` - Keep only valid items

**Used For:**
- Tracking which events have been read
- Computing unread counts
- Syncing read state with server

---

## Component Structure

### Thread Page Components

**Location:** `/vue/src/components/thread/`

| Component | Purpose |
|-----------|---------|
| `form_page.vue` | Discussion creation/edit form |
| `comment_form.vue` | New comment input |
| `edit_comment_form.vue` | Comment editing |
| `move_thread_form.vue` | Move to different group |
| `preview.vue` | Discussion preview card |
| `preview_collection.vue` | List of preview cards |
| `attachment_list.vue` | File attachments display |
| `link_preview.vue` | URL preview cards |
| `current_poll_banner.vue` | Active poll indicator |
| `arrangement_form.vue` | Thread ordering settings |
| `pin_event_form.vue` | Pin event to top |

### Strand Components

**Location:** `/vue/src/components/strand/`

The "strand" is the threaded event timeline within a discussion.

| Component | Purpose |
|-----------|---------|
| `page.vue` | Main thread page container |
| `list.vue` | Renders event list |
| `wall.vue` | Full thread display |
| `title.vue` | Discussion title display |
| `toc_nav.vue` | Table of contents navigation |
| `load_more.vue` | Pagination controls |
| `reply_form.vue` | Inline reply input |
| `members.vue` | Discussion members panel |
| `members_list.vue` | Members list display |
| `seen_by_modal.vue` | Who has seen dialog |
| `actions_panel.vue` | Thread actions menu |

### Strand Item Components

**Location:** `/vue/src/components/strand/item/`

Individual event type renderers:

| Component | Purpose |
|-----------|---------|
| `new_comment.vue` | Comment display |
| `new_discussion.vue` | Discussion context |
| `discussion_edited.vue` | Edit event |
| `poll_created.vue` | Poll in thread |
| `poll_edited.vue` | Poll edit event |
| `stance_created.vue` | Vote display |
| `stance_updated.vue` | Vote change |
| `outcome_created.vue` | Poll outcome |
| `headline.vue` | Event headline |
| `collapsed.vue` | Collapsed event |
| `removed.vue` | Discarded item placeholder |
| `stem_wrapper.vue` | Threading line |
| `intersection_wrapper.vue` | Visibility observer |
| `other_kind.vue` | Fallback for unknown events |

---

## Data Flow

### Loading a Discussion

```
PSEUDO-CODE:
1. Router navigates to /d/:key
2. Page component fetches /api/v1/discussions/:key
3. Response includes:
   - discussion record with merged reader data
   - created_event
   - group
   - author
4. Records imported to LokiJS stores
5. Discussion model has access to reader fields:
   - discussionReaderId
   - discussionReaderVolume
   - lastReadAt
   - dismissedAt
   - readRanges
   - guest
   - admin
```

### Loading Thread Events

```
PSEUDO-CODE:
1. Page component fetches /api/v1/events?discussion_id=X
2. Response includes events with:
   - sequenceId
   - positionKey
   - depth
   - parentId
3. Events rendered in positionKey order
4. Nesting determined by depth and parentId
5. Read state checked via discussion.hasRead(sequenceId)
```

### Marking as Read

```
PSEUDO-CODE:
1. IntersectionObserver detects event in viewport
2. Calls discussion.markAsRead(sequenceId)
3. Adds [sequenceId, sequenceId] to readRanges
4. Reduces ranges (merges adjacent)
5. Throttled PATCH to /api/v1/discussions/:key/mark_as_read
6. Server updates DiscussionReader.read_ranges_string
```

### Posting a Comment

```
PSEUDO-CODE:
1. CommentForm creates new CommentModel:
   - discussionId: discussion.id
   - parentId: parent.id
   - parentType: parent.constructor.name
   - body: editor content
   - bodyFormat: 'html'
2. comment.save() POSTs to /api/v1/comments
3. Server creates comment and new_comment Event
4. Response includes event with:
   - sequenceId
   - positionKey
   - parentId
5. Event imported, appears in strand
6. Flash message shown
```

---

## Key UI Patterns

### Read State Visualization

Events are styled based on read state:
- Unread events have visual indicator (e.g., blue dot)
- Read events are normal
- IntersectionObserver auto-marks as read when visible

### Threading Display

Events display with visual threading:
- `stem_wrapper.vue` draws connecting lines
- Depth determines indentation
- Parent references enable reply chains
- max_depth limits visual nesting

### Inbox/Dashboard

Discussion previews show:
- Title and last activity
- Unread count badge
- Active poll indicators
- Pinned status

Users can:
- Dismiss from inbox (dismissedAt)
- Recall back to inbox
- Change volume settings

### Forking Flow

When forking comments to new discussion:
1. User selects events via checkbox
2. Opens "Fork to new discussion" form
3. Sets forkedEventIds on new discussion
4. POST creates discussion and moves events
5. Original discussion shows "forked" indicator
