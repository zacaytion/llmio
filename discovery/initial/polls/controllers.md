# Polls Domain: Controllers & API

**Generated:** 2026-02-01
**Confidence:** 5/5

---

## Overview

The polls domain exposes REST APIs through two main controllers:

- **Api::V1::PollsController**: Poll management
- **Api::V1::StancesController**: Voting and voter management

Both inherit from `Api::V1::RestfulController` (via SnorlaxBase) which provides standard CRUD operations.

---

## PollsController

**File:** `/app/controllers/api/v1/polls_controller.rb`

### Standard CRUD

Inherits from RestfulController which provides:

- **GET /api/v1/polls**: List polls
- **GET /api/v1/polls/:key**: Show single poll
- **POST /api/v1/polls**: Create poll
- **PATCH /api/v1/polls/:key**: Update poll
- **DELETE /api/v1/polls/:key**: Destroy poll (not used, use discard)

### Custom Actions

#### GET /api/v1/polls (index)

Lists polls visible to the user with filtering:

```
Parameters:
  group_key: Filter by group
  discussion_key: Filter by discussion
  tags: Pipe-separated tag list
  status: 'active', 'closed', 'recent', 'template', 'vote'
  author_id: Filter by author
  poll_type: Filter by type
  query: Search text
  subgroups: 'none' to exclude subgroups
```

Uses PollQuery.filter for filtering and PollQuery.visible_to for authorization.

#### GET /api/v1/polls/:key (show)

Returns a single poll with associations:
- Author
- Group
- Discussion
- Poll options
- Current outcome
- My stance (if user has voted)
- Created event

Accepts pending membership invitation if applicable.

#### POST /api/v1/polls/:key/close

Closes an active poll early:

```
Response: Poll resource with closed_at set
Events: PollClosedByUser published
```

#### POST /api/v1/polls/:key/reopen

Reopens a closed poll:

```
Parameters:
  poll[closing_at]: New closing date
Response: Poll resource
Events: PollReopened published
```

#### POST /api/v1/polls/:key/remind

Sends reminder notifications:

```
Parameters:
  recipient_user_ids: [array of user IDs]
  recipient_audience: 'undecided', 'voters', etc.
  recipient_message: Custom message
  recipient_chatbot_ids: [chatbot IDs]
Response: { count: number_of_recipients }
```

#### DELETE /api/v1/polls/:key/discard

Soft deletes a poll:

```
Response: Poll resource with discarded_at set
```

#### PATCH /api/v1/polls/:key/add_to_thread

Adds standalone poll to a discussion:

```
Parameters:
  discussion_id: Target discussion
Response: Poll's created_event
```

#### GET /api/v1/polls/:key/voters

Lists users who have voted:

```
Response: Users collection (AuthorSerializer)
Note: Returns empty for anonymous polls
```

#### GET /api/v1/polls/:key/receipts

Returns voting receipts for verification:

```
Response: {
  voters_count: number,
  poll_title: string,
  receipts: [{
    poll_id, voter_id, voter_name, voter_email,
    member_since, inviter_id, inviter_name,
    invited_on, vote_cast
  }]
}
Note: Email partially hidden for non-admins
      Receipts are shuffled for privacy
```

### Visibility Logic

The `accessible_records` method uses PollQuery.visible_to which checks:

1. User is the poll author
2. Group allows public access (discussion_privacy_options = public_only)
3. User has active membership in the group
4. User has guest access via DiscussionReader
5. User has guest access via Stance

---

## StancesController

**File:** `/app/controllers/api/v1/stances_controller.rb`

### Standard CRUD

- **GET /api/v1/stances**: List stances for a poll
- **GET /api/v1/stances/:id**: Show single stance
- **POST /api/v1/stances**: Create stance (vote)
- **PATCH /api/v1/stances/:id**: Update stance (change vote)

### Create with Retry

The create action has special handling for duplicate votes:

If a stance already exists for this user/poll (RecordNotUnique):
1. Find the existing stance
2. Perform update instead of create

This handles race conditions in concurrent voting.

### Custom Actions

#### GET /api/v1/stances (index)

Lists stances for a poll with filtering:

```
Parameters:
  poll_id or poll_key: Required - which poll
  name: Search by voter name
  poll_option_id: Filter by selected option
```

Returns stances ordered by cast_at DESC (voters first, then invited).

Results are hidden based on poll's show_results? setting.

#### GET /api/v1/stances/:poll_id/users

Lists users with their voting status:

```
Parameters:
  query: Search text
Response: Users collection with meta:
  - guest_ids: Users who are guests
  - member_admin_ids: Group admin user IDs
  - stance_admin_ids: Poll admin user IDs
```

#### GET /api/v1/stances/my_stances

Lists current user's stances across polls:

```
Parameters:
  discussion_id: Filter by discussion
  group_id: Filter by group
Response: Stances with associated polls
```

#### POST /api/v1/stances/:id/uncast

Removes a vote (returns to undecided):

```
Response: Recent stances for the user in this poll
```

#### POST /api/v1/stances/make_admin

Grants poll admin rights to a voter:

```
Parameters:
  participant_id: User to promote
  poll_id: The poll
Response: Updated stance
```

#### POST /api/v1/stances/remove_admin

Removes poll admin rights:

```
Parameters:
  participant_id: User to demote
  poll_id: The poll
Response: Updated stance
```

#### POST /api/v1/stances/revoke

Removes a user's voting access:

```
Parameters:
  participant_id: User to remove
  poll_id: The poll
Response: All stances for that user in the poll
```

Revokes all stances (not just latest) for the user.

### Response Format

Stance responses include the latest stance events for the user, allowing the frontend to update the timeline.

---

## Authorization (Abilities)

### Ability::Poll

**File:** `/app/models/ability/poll.rb`

| Action | Requirements |
|--------|--------------|
| show | PollQuery.visible_to returns the poll |
| create | Group admin, or member if allowed, or standalone if verified |
| update | Poll admin and not closed |
| destroy | Poll admin |
| close | Poll admin and poll is active |
| reopen | Poll admin, poll closed, not anonymous |
| vote_in | Logged in, poll active, is voter or open voting member |
| announce/remind | Group admin, or poll admin if members_can_announce |
| add_voters | Poll admin |
| add_guests | Poll admin (and subscription allows guests) |
| export | Can show and results are visible |

### Ability::Stance

**File:** `/app/models/ability/stance.rb`

| Action | Requirements |
|--------|--------------|
| show | Can show the poll |
| create | Can vote in the poll |
| update | Can vote, is real participant, stance is latest |
| uncast | Can update and has cast vote |
| redeem | Stance is redeemable |
| make_admin/remove_admin/remove | Is poll admin |

---

## Query Objects

### PollQuery

**File:** `/app/queries/poll_query.rb`

#### visible_to(user:, chain:, group_ids:)

Filters polls to those visible to the user:

Uses LEFT OUTER JOINs on:
- discussions (for discussion-based access)
- groups (for group settings)
- memberships (for group membership)
- discussion_readers (for guest access)
- stances (for voter access)

Returns polls where:
- User is author, OR
- Group is public, OR
- User has active membership, OR
- User has discussion guest access, OR
- User has poll guest access

Also supports token-based access via discussion_reader_token or stance_token.

#### filter(chain:, params:)

Applies filter parameters:

- group_key + subgroups: Filter to group(s)
- discussion_key/id: Filter to discussion
- tags: Filter by tags
- status: Apply scope (active/closed/recent/template)
- status='vote': Exclude polls user has voted in
- author_id: Filter by author
- poll_type: Filter by type
- query: Text search

---

## Serialization

### PollSerializer

**File:** `/app/serializers/poll_serializer.rb`

Serializes polls with:

**Core attributes:** id, title, details, poll_type, closing_at, closed_at, anonymous, hide_results, specified_voters_only, etc.

**Computed attributes:**
- results: Calculated from PollService.calculate_results (conditional on show_results)
- stance_counts: (conditional on show_results)
- poll_option_names

**Associations:**
- discussion
- group
- author
- poll_options
- current_outcome
- my_stance (user's own stance)
- created_event

**Conditional includes:**
- `include_results?`: Only if poll.show_results?(voted: true)
- `include_my_stance?`: Only if user has a stance

**Hide when discarded:** Most attributes are nulled when poll is soft-deleted.

### StanceSerializer

Serializes stances with:

- id, reason, option_scores, cast_at, latest
- participant (or anonymous)
- poll reference
- stance_choices

---

## API Response Patterns

### Create/Update Response

Returns the event created by the operation, which contains the updated resource:

```json
{
  "events": [{ "id": 123, "kind": "poll_created", "eventable_id": 456 }],
  "polls": [{ "id": 456, ... }],
  "poll_options": [...],
  "users": [...]
}
```

### Collection Response

Returns array with associated records:

```json
{
  "polls": [...],
  "groups": [...],
  "users": [...],
  "poll_options": [...],
  "meta": { "total": 50, "root": "polls" }
}
```

### Error Response

On validation failure:

```json
{
  "errors": {
    "title": ["can't be blank"],
    "closing_at": ["must be in the future"]
  }
}
```

On authorization failure:

```json
{
  "error": "Access denied"
}
```
(HTTP 403)
