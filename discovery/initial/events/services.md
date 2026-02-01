# Events Domain: Services

**Generated:** 2026-02-01
**Confidence:** 5/5

---

## Overview

The events domain uses several services for event publishing, notification management, sequence handling, and real-time communication.

---

## EventService

**File:** `/app/services/event_service.rb`

### Purpose

Provides utilities for event manipulation, thread repair, and comment movement. Unlike other services, EventService is primarily for maintenance and repair operations.

### Methods

#### `remove_from_thread(event:, actor:)`

Removes an event from a discussion thread:

```
PSEUDO:
1. Validate event is a discussion_edited event
2. Authorize actor can remove_events from discussion
3. Set event.discussion_id to nil
4. Call discussion.thread_item_destroyed!
5. Reindex discussion for search
6. Broadcast 'event_remove_from_thread' on EventBus
```

#### `move_comments(discussion:, actor:, params:)`

Moves comments from one discussion to another:

```
PSEUDO:
1. Extract forked_event_ids from params
2. Find source discussion from first event
3. Authorize actor can move_comments in source AND destination
4. Enqueue MoveCommentsWorker with event IDs and discussion IDs
```

#### `repair_discussion(discussion_id)`

Repairs a discussion's event tree and sequences:

```
PSEUDO:
1. Find discussion (return if missing)
2. Ensure created_event exists (create if missing)
3. Set sequence_id on events missing it
4. Reset all events to depth 1 under created_event
5. Recalculate parent/depth for each event
6. Reset child positions for created_event and all parent events
7. Update descendant_count and child_count via SQL
8. Update discussion sequence info
9. Fix all DiscussionReader read_ranges
```

#### `reset_child_positions(parent_id, parent_position_key)`

Recalculates position and position_key for children of a parent event:

```
PSEUDO:
1. Build position_key SQL based on parent_position_key
2. Execute UPDATE to set position and position_key
   - Uses row_number() window function ordered by sequence_id
   - Zero-pads position to 5 digits
3. Drop the sequence counter for this parent
```

#### `repair_all_threads`

Enqueues repair for all discussions:

```
PSEUDO:
FOR each discussion ID:
  Enqueue RepairThreadWorker.perform_async(id)
```

---

## NotificationService

**File:** `/app/services/notification_service.rb`

### Purpose

Manages notification state (marking as viewed) and publishes updates to clients.

### Methods

#### `mark_as_read(eventable_type, eventable_id, actor_id)`

Marks notifications as read for a specific eventable:

```
PSEUDO:
1. Find unviewed notifications for actor where event matches eventable
2. Update all to viewed: true
3. Publish updated notifications to actor's channel
```

Used when: User creates a comment reply (marks parent notifications read)

#### `viewed_events(actor_id:, discussion_id:, sequence_ids:)`

Marks notifications as viewed for events at specific sequence positions:

```
PSEUDO:
1. Find events at given sequence_ids in discussion
2. Find reactions on those events' eventables
3. Collect event IDs from reaction events
4. For each eventable type (Comment, Discussion, Poll, Stance, Outcome):
   - Find notifications for events with those eventable IDs
5. Update all found notifications to viewed: true
6. Publish updated notifications to actor's channel
```

Used when: User views a portion of a discussion thread

#### `viewed(user:)`

Marks all of a user's notifications as viewed:

```
PSEUDO:
1. Update all unviewed notifications for user to viewed: true
2. Load recent 30 notifications with associations
3. Publish notifications to user's channel
```

Used when: User opens notifications panel

---

## SequenceService

**File:** `/app/services/sequence_service.rb`

### Purpose

Manages atomic sequence generation using a `partition_sequences` table to avoid race conditions when multiple events are created simultaneously.

### Methods

#### `seq_present?(key, id)`

Checks if a sequence exists:

```
PSEUDO:
SELECT 0 FROM partition_sequences WHERE key = {key} AND id = {id}
RETURN true if row exists
```

#### `create_seq!(key, id, start)`

Creates a new sequence counter:

```
PSEUDO:
INSERT INTO partition_sequences (key, id, counter)
VALUES ({key}, {id}, {start})
ON CONFLICT DO NOTHING
```

#### `next_seq!(key, id)`

Atomically increments and returns the next value:

```
PSEUDO:
UPDATE partition_sequences SET counter = counter + 1
WHERE key = {key} AND id = {id}
RETURNING counter
```

#### `drop_seq!(key, id)`

Removes a sequence counter:

```
PSEUDO:
DELETE FROM partition_sequences WHERE key = {key} AND id = {id}
```

### Sequence Keys

Two types of sequences are managed:

1. **discussions_sequence_id** (id = discussion_id)
   - Tracks global sequence_id within a discussion
   - Used by `Event#next_sequence_id!`

2. **events_position** (id = parent_id)
   - Tracks position among siblings
   - Used by `Event#next_position!`

---

## MessageChannelService

**File:** `/app/services/message_channel_service.rb`

### Purpose

Publishes real-time updates to connected clients via Redis pub/sub.

### Methods

#### `publish_models(models, serializer:, scope:, root:, group_id:, user_id:)`

Serializes and publishes models to a channel:

```
PSEUDO:
1. Return early if models empty
2. Build RecordCache for the collection
3. Serialize models using provided or auto-detected serializer
4. Call publish_serialized_records with data
```

#### `serialize_models(models, serializer:, scope:, root:)`

Serializes models to JSON format:

```
PSEUDO:
1. Get first model to determine type
2. If model is Event, use EventSerializer
3. Otherwise, use "{ModelClass}Serializer"
4. Return ArraySerializer with scope and root key
```

#### `publish_serialized_records(data, group_id:, user_id:)`

Publishes to Redis:

```
PSEUDO:
1. Determine room from user_id or group_id
   - user-{id} for user channels
   - group-{id} for group channels
2. Publish to Redis /records channel:
   { room: room, records: data }
```

#### `publish_system_notice(notice, reload)`

Publishes system-wide notices:

```
PSEUDO:
Publish to /system_notice channel:
{ version: current_version, notice: notice, reload: reload }
```

### Channel Architecture

The system uses Redis pub/sub with an external channels service:

```
PSEUDO: Real-time update flow
1. Server publishes to Redis channel
2. External channels service subscribes to Redis
3. Channels service pushes to clients via WebSocket/SSE
4. Client receives and updates LokiJS store
```

---

## PublishEventWorker

**File:** `/app/workers/publish_event_worker.rb`

### Purpose

Background worker that triggers event side effects after creation.

### Implementation

```
PSEUDO: perform(event_id)
1. Find event using STI-aware lookup: Event.sti_find(event_id)
2. Call event.trigger! to execute all side effects
```

### Why Background?

- Decouples event creation from notification/email processing
- Allows quick response to user actions
- Side effects (emails, webhooks) can be slow
- Sidekiq provides retry on failure

---

## ChatbotService Integration

Events with the `Events::Notify::Chatbots` concern enqueue:

```
GenericWorker.perform_async('ChatbotService', 'publish_event!', id)
```

The ChatbotService looks up configured webhooks for the event's group and sends appropriate payloads.

---

## EventBus

**File:** `/lib/event_bus.rb`

### Purpose

Simple in-process pub/sub system for decoupled event handling.

### Methods

#### `configure { |config| ... }`

Yields self for configuration block.

#### `broadcast(event, *params)`

Calls all registered listeners for an event:

```
PSEUDO:
FOR each listener in listeners[event]:
  listener.call(*params)
```

#### `listen(*events, &block)`

Registers a listener for one or more events:

```
PSEUDO:
FOR each event in events:
  Add block to listeners[event] set
```

#### `deafen(*events, &block)`

Removes a listener:

```
PSEUDO:
FOR each event in events:
  Remove block from listeners[event] set
```

#### `clear`

Removes all listeners (used in testing).

---

## EventBus Configuration

**File:** `/config/initializers/event_bus.rb`

### Registered Listeners

#### Discussion Reader Updates

Listens to: `new_comment_event`, `new_discussion_event`, `discussion_edited_event`, `poll_created_event`, `poll_edited_event`, `stance_created_event`, `outcome_created_event`, `poll_closed_by_user_event`

```
PSEUDO: When any of these events fire:
  IF event has a discussion:
    Find or create DiscussionReader for (discussion, user or participant)
    Update reader:
      - Add sequence_id to read ranges
      - Set volume to loud
```

This ensures authors/participants have their own activity marked as read.

#### Real-time Read State Updates

Listens to: `discussion_mark_as_read`, `discussion_dismiss`, `discussion_mark_as_seen`

```
PSEUDO: When reader state changes:
  Publish discussion to user's channel
  (Updates other open tabs/devices)
```

---

## UsersByVolumeQuery

**File:** `/app/extras/queries/users_by_volume_query.rb`

### Purpose

Determines which users should receive notifications based on their volume settings.

### Volume Levels

| Level | Value | Receives |
|-------|-------|----------|
| mute | 0 | Nothing |
| quiet | 1 | In-app only |
| normal | 2 | In-app + important emails |
| loud | 3 | In-app + all emails |

### Methods

#### `email_notifications(model)`

Returns users who should receive emails (normal or loud volume):

```
PSEUDO:
users_by_volume(model, '>=', NORMAL)
```

#### `app_notifications(model)`

Returns users who should receive in-app notifications (quiet or higher):

```
PSEUDO:
users_by_volume(model, '>=', QUIET)
```

#### `normal_or_loud(model)`

Alias for `email_notifications`.

#### `mute(model)`, `quiet(model)`, `normal(model)`, `loud(model)`

Returns users at exactly that volume level.

### Volume Cascade

The query joins multiple tables to find effective volume:

```
PSEUDO: Effective volume priority
1. Stance volume (if user has stance on poll)
2. DiscussionReader volume (if user has reader for discussion)
3. Membership volume (if user is group member)
4. Default: normal (2)
```

Query logic:
```
PSEUDO:
1. LEFT JOIN discussion_readers on discussion_id and user_id
2. LEFT JOIN memberships on group_id and user_id
3. LEFT JOIN stances on poll_id and participant_id (latest only)
4. WHERE user has membership OR guest discussion_reader OR guest stance
5. Filter by: COALESCE(stance.volume, reader.volume, membership.volume, 2) {operator} {volume}
```
