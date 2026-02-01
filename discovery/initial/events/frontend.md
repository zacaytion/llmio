# Events Domain: Frontend

**Generated:** 2026-02-01
**Confidence:** 4/5

---

## Overview

The frontend handles events through a LokiJS-based record store, event models, and Vue components for timeline/thread display. Events are the fundamental unit for rendering discussion threads.

---

## Record Interface

**File:** `/vue/src/shared/interfaces/event_records_interface.js`

### Purpose

Provides the interface between the API and the LokiJS store for events.

### Methods

#### `fetchByDiscussion(discussionKey, options)`

Fetches events for a discussion:

```
PSEUDO:
Set options.discussion_key = discussionKey
Call base fetch with params
```

#### `findByDiscussionAndSequenceId(discussion, sequenceId)`

Finds a specific event by sequence:

```
PSEUDO:
Chain LokiJS queries:
  - Find by discussionId
  - Find by sequenceId
  - Return first match
```

---

## Event Model

**File:** `/vue/src/shared/models/event_model.js`

### Properties

- `singular`: 'event'
- `plural`: 'events'
- `indices`: ['discussionId', 'sequenceId', 'position', 'depth', 'parentId', 'positionKey']
- `uniqueIndices`: ['id']

### Relationships

```
PSEUDO:
belongsTo 'parent' from 'events'
belongsTo 'actor' from 'users'
belongsTo 'discussion'
hasMany 'notifications'
```

### Default Values

```
PSEUDO:
pinned: false
eventableId: null
eventableType: null
discussionId: null
sequenceId: null
position: 0
showReplyForm: true
```

### Key Methods

#### `parentOrSelf()`

Returns parent if nested, otherwise self.

#### `isNested()`

Returns true if depth > 1.

#### `isSurface()`

Returns true if depth === 1 (top-level reply).

#### `surfaceOrSelf()`

Returns parent if nested, otherwise self. Used for reply threading.

#### `children()`

Finds all events with this event as parent:

```
PSEUDO:
Records.events.find({ parentId: this.id })
```

#### `model()`

Gets the eventable record:

```
PSEUDO:
Use eventTypeMap to find collection name
Find record by eventableId in that collection
```

#### `isPollEvent()`

Returns true if eventableType is Poll, Outcome, or Stance.

#### `isUnread()`

Checks if event hasn't been read:

```
PSEUDO:
Return NOT discussion.hasRead(this.sequenceId)
```

#### `markAsRead()`

Marks event as read via discussion:

```
PSEUDO:
discussion.markAsRead(this.sequenceId)
```

#### `removeFromThread()`

Removes event from discussion timeline:

```
PSEUDO:
PATCH /api/v1/events/{id}/remove_from_thread
Then remove from local store
```

#### `pin(title)`

Pins the event:

```
PSEUDO:
PATCH /api/v1/events/{id}/pin with pinned_title: title
```

#### `unpin()`

Unpins the event:

```
PSEUDO:
PATCH /api/v1/events/{id}/unpin
```

#### `suggestedTitle()`

Generates a title for pinning:

```
PSEUDO:
Get the eventable model
IF model has title:
  Return title with HTML stripped
ELSE:
  Parse body/statement as HTML
  IF has h1/h2/h3:
    Return heading text
  ELSE:
    Return actor name
```

#### `isForking()`

Checks if event is selected for forking:

```
PSEUDO:
Return discussion.forkedEventIds includes this.id OR parent is forking
```

#### `forkingDisabled()`

Checks if event cannot be individually forked:

```
PSEUDO:
Return parent is forking OR parent is poll_created
```

---

## Event Service

**File:** `/vue/src/shared/services/event_service.js`

### Purpose

Provides action definitions for event-related operations in the UI.

### Actions

#### `move_event`

Move event to another discussion:

- **Icon:** mdi-call-split
- **Menu:** true
- **Kinds:** new_discussion, poll_created, new_comment
- **canPerform:** Model not discarded, discussion not closed, user can move thread
- **perform:** Add event ID to discussion.forkedEventIds

#### `pin_event`

Pin an event:

- **Icon:** mdi-pin-outline
- **Menu:** true
- **Kinds:** new_comment, poll_created
- **canPerform:** Model not discarded, user can pin
- **perform:** Open PinEventForm modal

#### `unpin_event`

Remove pin from event:

- **Icon:** mdi-pin-off
- **Menu:** true
- **Kinds:** new_comment, poll_created
- **canPerform:** Model not discarded, user can unpin
- **perform:** Call event.unpin(), show success flash

#### `copy_url`

Copy event URL to clipboard:

- **Icon:** mdi-link
- **Menu:** true
- **Kinds:** new_comment, poll_created, stance_created, stance_updated
- **canPerform:** Model not discarded
- **perform:** Copy URL to clipboard, show success flash

---

## Frontend EventBus

**File:** `/vue/src/shared/services/event_bus.js`

A Vue-based event bus for component communication (separate from backend EventBus):

- `$on(event, handler)` - Register listener
- `$off(event, handler)` - Remove listener
- `$emit(event, data)` - Emit event

Used for:
- `currentComponent` - Track current page component
- `setAnchor` - Set scroll anchor
- `visibleKeys` - Track visible position keys
- `toggleThreadNav` - Toggle navigation drawer

---

## Strand Components

The strand component family renders the discussion thread.

### StrandPage

**File:** `/vue/src/components/strand/page.vue`

Root component for thread display:

```
PSEUDO: Initialization
1. Fetch discussion by key
2. Create ThreadLoader for discussion
3. Respond to route changes (sequence_id, comment_id, query params)
4. Listen for anchor and visibility events
```

Handles:
- Route-based navigation to specific events
- Unread/newest jumping
- Focus mode for highlighting specific items
- Thread navigation panel toggle

### StrandList

**File:** `/vue/src/components/strand/list.vue`

Recursive list component for event tree:

```
PSEUDO: For each object in collection
1. Show load-more if missingEarlier
2. If collapsed: show Collapsed component
3. If not collapsed:
   - Show gutter with avatar or forking checkbox
   - Show StemWrapper for visual threading
   - Show IntersectionWrapper for event content
   - Recursively render children as StrandList
   - Show ReplyForm
4. Show load-more if missingAfter
```

Props:
- `loader` - ThreadLoader instance
- `collection` - Array of event objects with children
- `focusSelector` - CSS selector for focused item

### Event Item Components

Located in `/vue/src/components/strand/item/`:

#### `intersection_wrapper.vue`

Wraps event content, handles visibility tracking for read state.

#### `stem_wrapper.vue`

Renders visual stem lines for threading.

#### `collapsed.vue`

Shows collapsed event count, expands on click.

#### `headline.vue`

Common headline pattern for events showing actor and timestamp.

#### `new_comment.vue`

Renders comment events with body, reactions, actions.

#### `new_discussion.vue`

Renders discussion opening post.

#### `poll_created.vue`

Renders poll with voting interface.

#### `stance_created.vue` / `stance_updated.vue`

Renders vote/stance with reason.

#### `outcome_created.vue`

Renders poll outcome.

#### `discussion_edited.vue`

Renders discussion edit notification.

#### `poll_edited.vue`

Renders poll edit notification.

#### `other_kind.vue`

Generic fallback for other event types.

#### `removed.vue`

Placeholder for deleted/discarded events.

---

## Thread Loading

### ThreadLoader

**File:** `/vue/src/shared/loaders/thread_loader.js`

Manages loading and organizing events for a discussion:

```
PSEUDO: Key responsibilities
1. Fetch events from API with pagination
2. Build tree structure from flat events
3. Track collapsed/expanded state
4. Handle load-more for children and siblings
5. Track visible/missing ranges
```

Key state:
- `events` - Flat list of loaded events
- `collapsed` - Map of event ID to collapsed state
- `visibleRanges` - Which sequence ranges are loaded

---

## Notification Components

### NotificationsCount

**File:** `/vue/src/components/common/notifications_count.vue`

Shows unread notification count badge.

### Notifications

**File:** `/vue/src/components/common/notifications.vue`

Dropdown panel showing notification list with:
- Mark all as read
- Click to navigate
- Real-time updates

---

## Notification Model

**File:** `/vue/src/shared/models/notification_model.js`

### Relationships

```
PSEUDO:
belongsTo 'event'
belongsTo 'user'
belongsTo 'actor' from 'users'
```

### Methods

#### `href()`

Generates navigation URL:

```
PSEUDO:
IF no URL: return '/'
IF membership_requested: return group membership_requests page
IF URL starts with base: strip base and return relative path
Otherwise return URL as-is
```

#### `args()`

Returns translation arguments:

```
PSEUDO:
{
  actor: name,
  reaction: emoji (if reaction_created),
  title: title,
  poll_type: pollType,
  model: model
}
```

#### `isRouterLink()`

Returns true unless URL is an invitation link (needs full page load).

---

## Real-time Updates

Events are received via MessageChannelService through WebSocket/SSE:

```
PSEUDO: Update flow
1. Server publishes to Redis (user or group channel)
2. Channels service pushes to client
3. Client receives on /records channel
4. Records are imported into LokiJS store
5. Vue reactivity updates components
```

Notifications also update via user channel:
- New notifications appear immediately
- Viewed state syncs across tabs

---

## Pinning UI

### PinEventForm

**File:** `/vue/src/components/thread/pin_event_form.vue`

Modal for pinning events:

```
PSEUDO:
1. Show event with suggested title pre-filled
2. Allow editing title
3. On submit: call event.pin(title)
4. Show success message
```

Pinned events appear with:
- Pin icon in timeline
- Custom title if set
- Prominent positioning
