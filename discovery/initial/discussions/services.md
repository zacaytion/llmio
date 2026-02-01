# Discussions Domain - Services

**Generated:** 2026-02-01
**Domain:** DiscussionService, CommentService, EventService
**Confidence:** 5/5 (High - Based on direct service inspection)

---

## Overview

The discussions domain uses three main service classes:
- **DiscussionService** - All discussion mutations (create, update, move, close, etc.)
- **CommentService** - Comment creation, editing, and deletion
- **EventService** - Thread repair and comment moving operations

All mutations go through services following Loomio's strict architectural pattern.

---

## DiscussionService

**File:** `/app/services/discussion_service.rb`

### create(discussion:, actor:, params:)

Creates a new discussion.

**Flow:**
1. Authorize actor can create discussion
2. Authorize any recipients (user_ids, emails, audience)
3. Set discussion author to actor
4. Validate discussion
5. Within transaction:
   - Save discussion
   - Create DiscussionReader for author (with admin: true, guest if no group)
   - Add any recipient users via add_users
6. Broadcast 'discussion_create' to EventBus
7. Publish Events::NewDiscussion with recipient info

**Return:** The Event created, or false if invalid

**Key Logic:**
- Author automatically becomes admin of their discussion
- If discussion has no group (direct discussion), author marked as guest
- Recipients can be specified by user_ids, emails, or audience keyword

### update(discussion:, actor:, params:)

Updates discussion attributes.

**Flow:**
1. Authorize actor can update discussion
2. Authorize any new recipients
3. Assign attributes (except group_id)
4. Validate discussion
5. Within transaction:
   - Save discussion
   - Update versions count
   - If max_depth changed, queue RepairThreadWorker
   - Add any new recipients
6. Broadcast 'discussion_update' to EventBus
7. Publish Events::DiscussionEdited

**Key Logic:**
- Group cannot be changed via update (use move instead)
- Changing max_depth triggers thread reorganization in background

### close(discussion:, actor:)

Closes a discussion.

**Flow:**
1. Authorize actor can update discussion
2. Set closed_at to now, closer_id to actor
3. Publish models to message channel

**Effect:** Closed discussions prevent new comments and voting.

### reopen(discussion:, actor:)

Reopens a closed discussion.

**Flow:**
1. Authorize actor can update discussion
2. Clear closed_at and closer_id
3. Publish models to message channel

### move(discussion:, params:, actor:)

Moves discussion to a different group.

**Flow:**
1. Load destination group from params
2. Authorize actor can move discussions TO destination
3. Authorize actor can move this discussion
4. Within transaction:
   - Update discussion.group to destination
   - Adjust privacy based on destination's settings
   - Move all polls to destination group
   - Update attachments' group_id
5. Queue PollService.group_members_added for poll voter updates
6. Queue SearchService.reindex_by_discussion_id
7. Publish Events::DiscussionMoved

**Privacy Adjustment Logic:**
- If destination is `public_only` -> discussion becomes public
- If destination is `private_only` -> discussion becomes private
- Otherwise -> preserve current privacy setting

### pin(discussion:, actor:) / unpin(discussion:, actor:)

Pins or unpins discussion.

**Flow:**
1. Authorize actor can pin discussion
2. Set pinned_at to now (or nil for unpin)
3. Broadcast 'discussion_pin' to EventBus

**Effect:** Pinned discussions appear at top of group discussion list.

### discard(discussion:, actor:)

Soft-deletes a discussion.

**Flow:**
1. Authorize actor can discard discussion
2. Within transaction:
   - Set discarded_at and discarded_by
   - Discard all associated polls
   - Queue search reindex
3. Broadcast 'discussion_discard'
4. Return created_event

### invite(discussion:, actor:, params:)

Invites users to an existing discussion.

**Flow:**
1. Authorize recipients
2. Within transaction:
   - Add users via add_users
   - For active polls with anyone_can_vote, create stances
3. Publish Events::DiscussionAnnounced

### add_users(discussion:, actor:, user_ids:, emails:, audience:)

Internal method to add users to discussion.

**Flow:**
1. Find or create users from user_ids, emails, audience
2. Get volume preferences from existing memberships
3. Unrevoke any previously revoked readers
4. Build DiscussionReader records for new users:
   - Set guest: true if user is not a group member
   - Set admin: true if discussion has no group
   - Set volume from membership or user default
5. Bulk import new readers
6. Update members_count counter

**Audience Keywords:**
- `group` - All group members

### mark_as_read(discussion:, params:, actor:)

Marks events as read.

**Flow:**
1. Check actor can mark_as_read
2. Parse ranges from params
3. Mark notifications as viewed for those sequence_ids
4. Update DiscussionReader with new read ranges
5. Broadcast 'discussion_mark_as_read'

### mark_as_seen(discussion:, actor:)

Marks discussion as seen (visited).

**Flow:**
1. Authorize actor can mark_as_seen
2. Get DiscussionReader for actor
3. Call viewed! on reader
4. Publish models to channel
5. Broadcast 'discussion_mark_as_seen'

### dismiss(discussion:, params:, actor:) / recall(discussion:, params:, actor:)

Dismisses or recalls discussion from inbox.

**Flow:**
1. Authorize actor can dismiss
2. Get DiscussionReader
3. Set dismissed_at to now (or nil for recall)
4. Broadcast event

### update_reader(discussion:, params:, actor:)

Updates reader preferences like volume.

**Flow:**
1. Authorize actor can show discussion
2. Get DiscussionReader
3. Update volume from params
4. Also update volume on related Stances
5. Broadcast 'discussion_update_reader'

---

## CommentService

**File:** `/app/services/comment_service.rb`

### create(comment:, actor:)

Creates a new comment.

**Flow:**
1. Authorize actor can create comment
2. Set author to actor
3. Validate comment
4. Save comment
5. Broadcast 'comment_create' to EventBus
6. Publish Events::NewComment

**Return:** The Event created, or false if invalid

### update(comment:, params:, actor:)

Updates a comment.

**Flow:**
1. Authorize actor can update comment
2. Set edited_at to now
3. Assign attributes and files
4. Validate comment
5. Save comment
6. Update versions_count
7. Broadcast 'comment_update' to EventBus
8. Publish Events::CommentEdited

### discard(comment:, actor:)

Soft-deletes a comment.

**Flow:**
1. Authorize actor can discard comment
2. Within transaction:
   - Set discarded_at and discarded_by
   - Unpin the created_event
3. Update discussion sequence_info
4. Return created_event

### undiscard(comment:, actor:)

Restores a discarded comment.

**Flow:**
1. Authorize actor can undiscard
2. Within transaction:
   - Clear discarded_at and discarded_by
   - Restore user_id on created_event
3. Return created_event

### destroy(comment:, actor:)

Permanently deletes a comment.

**Flow:**
1. Authorize actor can destroy comment
2. Note: Cannot delete if comment has replies (checked in ability)
3. Destroy comment
4. Queue RepairThreadWorker

---

## EventService

**File:** `/app/services/event_service.rb`

### move_comments(discussion:, actor:, params:)

Moves comments (events) to a different discussion.

**Flow:**
1. Get event_ids from params (forked_event_ids)
2. Find source discussion from first event
3. Authorize actor can move_comments on source
4. Authorize actor can move_comments on destination
5. Queue MoveCommentsWorker with event_ids, source_id, target_id

**Worker Execution (MoveCommentsWorker):**
1. Sanitize event_ids to ensure they belong to source
2. Find all child events recursively
3. Update eventable.discussion_id for comments and polls
4. Reparent comments whose parent is not in the move set
5. Update event.discussion_id, clear sequence_id
6. Repair both source and target discussions
7. Reindex search for both
8. Update attachment group_ids
9. Publish models to channel

### repair_discussion(discussion_id)

Rebuilds thread structure after moves or deletes.

**Flow:**
1. Ensure created_event exists for discussion
2. Set sequence_id on events that are missing one
3. Reset all event ancestry to discussion.created_event
4. Rebuild parent relationships based on eventable connections
5. Recalculate positions within each parent:
   - Use position = row_number by sequence_id
   - Build position_key from parent's position_key + padded position
6. Recalculate descendant_count and child_count for all events
7. Intersect DiscussionReader read_ranges with valid ranges

### reset_child_positions(parent_id, parent_position_key)

Reorders children of a parent event.

Uses SQL to set:
- `position` = row number ordered by sequence_id
- `position_key` = parent's position_key + padded position

Drops the sequence cache for this parent after updating.

### remove_from_thread(event:, actor:)

Removes an edited discussion event from the thread timeline.

Only works on 'discussion_edited' events. Sets discussion_id to nil.

---

## Background Workers

### RepairThreadWorker

**File:** `/app/workers/repair_thread_worker.rb`

Simple worker that calls `EventService.repair_discussion(discussion_id)`.

Used after:
- max_depth changes
- Comment deletion
- Comment moves

### MoveCommentsWorker

**File:** `/app/workers/move_comments_worker.rb`

Handles the heavy lifting of moving comments between discussions:
1. Collects all descendant events of selected events
2. Moves comments, polls, and events to target discussion
3. Reparents orphaned comments
4. Repairs both discussions
5. Reindexes search
6. Publishes real-time updates

---

## EventBus Integration

Services broadcast events for side effects:

| Broadcast | Listeners |
|-----------|-----------|
| `discussion_create` | Update DiscussionReader state |
| `discussion_update` | Trigger republish |
| `discussion_mark_as_read` | Publish read state to channels |
| `discussion_mark_as_seen` | Update seen counts |
| `comment_create` | Update DiscussionReader volume |
| `comment_update` | Trigger republish |

EventBus listeners are configured in `/config/initializers/event_bus.rb`.

---

## Key Patterns

### Authorization First

Every service method starts with authorization:
```
PSEUDO-CODE:
actor.ability.authorize! :action, resource
```

### Transaction Wrapping

Multi-step operations use transactions:
```
PSEUDO-CODE:
Discussion.transaction do
  save changes
  create events
  update caches
end
```

### Event Publishing

Actions create Event records that trigger notifications:
```
PSEUDO-CODE:
Events::ActionName.publish!(
  eventable,
  recipient_user_ids: [...],
  recipient_audience: '...'
)
```

### Async Workers for Heavy Operations

Thread repair and comment moves are async to avoid blocking:
```
PSEUDO-CODE:
RepairThreadWorker.perform_async(discussion.id)
MoveCommentsWorker.perform_async(event_ids, source_id, target_id)
```
