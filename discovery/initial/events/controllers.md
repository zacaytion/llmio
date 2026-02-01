# Events Domain: Controllers

**Generated:** 2026-02-01
**Confidence:** 5/5

---

## Overview

Events are exposed via a single REST controller that provides thread navigation, pinning, and timeline functionality.

---

## EventsController

**File:** `/app/controllers/api/v1/events_controller.rb`

### Inheritance

Extends `Api::V1::RestfulController` (via SnorlaxBase), providing standard REST patterns.

### Actions

#### `index`

Lists events for a discussion with extensive filtering options.

**Route:** `GET /api/v1/events`

**Required params:**
- `discussion_id` - Discussion to fetch events from

**Optional params:**
- `from` - Minimum sequence_id to start from (default: 0)
- `per` - Results per page (default: 30)
- `order` - Sort column: sequence_id, position, or position_key
- `order_by` - Alternative ordering with optional `order_desc` for direction
- `pinned` - Set to 'true' to only return pinned events
- `kind` - Comma-separated list of event kinds to include
- `parent_id` - Filter to children of specific event
- `depth`, `sequence_id`, `position`, `position_key` - Direct field filters
- Comparison suffixes: `_lt`, `_gt`, `_lte`, `_gte`, `_sw` (starts with)
- `sequence_id_in` - Range string (e.g., "1-5_10-15") for specific ranges
- `sequence_id_not_in` - Exclude ranges
- `from_sequence_id_of_position` - Start from event at specific position

**Response:** Standard collection with events, discussions (with reader), and associated records

**Authorization:** User must be able to show the discussion

#### `comment`

Fetches a specific event by comment ID.

**Route:** `GET /api/v1/events/comment`

**Params:**
- `discussion_id` - Discussion containing the comment
- `comment_id` - ID of the comment to find

**Response:** Single event for the new_comment event

**Error:** 404 if comment not found

#### `timeline`

Returns lightweight timeline data for efficient rendering.

**Route:** `GET /api/v1/events/timeline`

**Params:**
- `discussion_id` - Discussion to get timeline for

**Response:** Array of tuples containing:
- position_key
- sequence_id
- created_at
- user_id
- depth
- descendant_count

This endpoint is optimized for timeline navigation without loading full event data.

#### `position_keys`

Returns sorted list of all position_keys in a discussion.

**Route:** `GET /api/v1/events/position_keys`

**Params:**
- `discussion_id` - Discussion to get keys for

**Response:** Sorted array of position_key strings

Used for jump navigation in long threads.

#### `pin`

Pins an event to make it prominent in the timeline.

**Route:** `PATCH /api/v1/events/:id/pin`

**Params:**
- `id` - Event ID
- `pinned_title` - Optional custom title for the pin

**Authorization:** User must be able to pin events in the discussion

**Response:** Updated event with pinned: true

#### `unpin`

Removes pin from an event.

**Route:** `PATCH /api/v1/events/:id/unpin`

**Params:**
- `id` - Event ID

**Authorization:** User must be able to unpin events in the discussion

**Response:** Updated event with pinned: false

#### `remove_from_thread`

Removes an event from the discussion timeline.

**Route:** `PATCH /api/v1/events/:id/remove_from_thread`

**Params:**
- `id` - Event ID

**Authorization:** Event must be discussion_edited kind; user must be able to remove_events

**Response:** Updated event with discussion_id: nil

#### `count`

Returns count of matching events.

**Route:** `GET /api/v1/events/count`

**Params:** Same filters as `index`

**Response:** Integer count

---

## Filtering Logic

The `accessible_records` private method builds the query:

```
PSEUDO:
1. Load and authorize discussion
2. Start with Event.where(discussion_id: discussion.id)
3. Apply order_by if specified (with optional DESC)
4. Otherwise filter by "from" parameter on order column
5. Apply sequence_id_in ranges if specified
6. Apply sequence_id_not_in exclusion ranges if specified
7. Apply pinned: true filter if requested
8. Apply kind filter if specified (comma-separated list)
9. For each position field (parent_id, depth, sequence_id, position, position_key):
   - Apply exact match if param present
   - Apply comparison operators (_lt, _gt, _lte, _gte)
   - Apply prefix match (_sw) for position_key navigation
```

### Range String Format

The `sequence_id_in` and `sequence_id_not_in` parameters accept a compressed range format:

```
Format: "start1-end1_start2-end2_..."
Example: "1-10_25-30_45-50"

PSEUDO: Parse ranges
Split by '_' to get individual ranges
Split each by '-' to get [start, end]
Convert to Ruby Range objects
Apply as WHERE clause
```

---

## Response Structure

Events are serialized with the EventSerializer, returning:

```
{
  "events": [...],
  "discussions": [...],     // Parent discussion with reader
  "parent_events": [...],   // Parent events if fetching children
  "users": [...],           // Actors
  "polls": [...],           // If eventable is poll
  "comments": [...],        // If eventable is comment
  "stances": [...],         // If eventable is stance
  "outcomes": [...],        // If eventable is outcome
  "groups": [...]           // For discussion_moved events
}
```

---

## Pagination

Default pagination:
- `per`: 30 events per page
- `from`: Starting sequence_id (0 for beginning)

The `page_collection` method applies ordering and limit:

```
PSEUDO:
collection.order(order_column).limit(per)
```

Note: Count collection is disabled (`count_collection` returns false).

---

## Authorization

All endpoints require discussion access:

- `load_and_authorize(:discussion)` - Verifies user can show the discussion
- Pin/unpin use `current_user.ability.authorize!(:pin, @event)` / `(:unpin, @event)`
- Remove from thread uses EventService which checks `:remove_events` permission

---

## Related Routes

From `/config/routes.rb`:

```
resources :events, only: [:index] do
  collection do
    get :timeline
    get :position_keys
    get :comment
    get :count
  end
  member do
    patch :pin
    patch :unpin
    patch :remove_from_thread
  end
end
```
