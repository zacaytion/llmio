# Polls Domain: Services

**Generated:** 2026-02-01
**Confidence:** 5/5

---

## Overview

The polls domain has three main service classes that handle all mutations:

- **PollService**: Poll lifecycle (create, update, close, reopen, invite, remind)
- **StanceService**: Voting operations (create, update, uncast)
- **OutcomeService**: Outcome management (create, update, invite)

---

## PollService

**File:** `/app/services/poll_service.rb`

### Core Operations

#### create(poll:, actor:, params:)

Creates a new poll with the following flow:

1. Authorize actor can create the poll
2. Assign author to actor
3. Prioritize poll options (sort by name for meetings)
4. Validate poll
5. Save poll in transaction
6. Update counts
7. If not specified_voters_only, create stances for all eligible voters
8. Broadcast poll_create via EventBus
9. Publish PollCreated event

Stances are created for group members when `specified_voters_only = false`.

#### update(poll:, params:, actor:)

Updates an existing poll:

1. Authorize actor can update
2. Authorize any new recipients for announcement
3. Assign attributes (excludes poll_type, discussion_id, poll_template_id)
4. Re-authorize (group_id may have changed)
5. Validate and save
6. Update counts
7. Trigger search reindex
8. Handle new member stances if group changed
9. Create/find recipient users
10. Publish PollEdited event

#### invite(poll:, actor:, params:)

Invites users to vote in a poll:

1. Authorize invitation permissions
2. In transaction:
   - If poll has discussion, add users there too
   - Create stances for invited users
   - If notify_recipients, publish PollAnnounced event
3. Return created stances

#### remind(poll:, actor:, params:)

Sends reminders to existing voters:

1. Authorize remind permission
2. Find existing users (from audience or user_ids)
3. Publish PollReminder event

### Closing Operations

#### close(poll:, actor:)

Manually closes a poll:

1. Authorize close permission
2. Execute do_closing_work
3. Publish PollClosedByUser event

#### do_closing_work(poll:)

Internal method that handles poll closing mechanics:

1. Return early if already closed
2. Delete existing StanceReceipts, then create new ones
3. If anonymous, scrub participant_id from all stances
4. If hide_results was until_closed, reveal stance events in discussion
5. Set closed_at to now
6. Trigger search reindex

This is called by both manual close and automatic expiration.

#### expire_lapsed_polls

Class method that closes all overdue polls:

1. Find polls where closing_at < now and closed_at is null
2. For each, enqueue CloseExpiredPollWorker

The worker calls do_closing_work and publishes PollExpired event.

#### reopen(poll:, params:, actor:)

Reopens a closed poll:

1. Authorize reopen permission
2. Set new closing_at, clear closed_at
3. Validate
4. Save and broadcast poll_reopen
5. Publish PollReopened event

Note: Cannot reopen anonymous polls.

### Voter Management

#### create_stances(poll:, actor:, user_ids:, emails:, audience:, include_actor:)

Creates stance records for voters:

1. Find existing voter IDs to avoid duplicates
2. Find or create users from ids/emails/audience
3. Get volume preferences from DiscussionReader or Membership
4. Handle reinvited users (clear revoked_at)
5. Build new stance records for new users
6. Bulk import stances
7. Reset latest stances and update counts

Volumes cascade: DiscussionReader > Membership > User default.

#### create_anyone_can_vote_stances(poll)

For open polls (specified_voters_only = false):

1. Get group member IDs
2. Get discussion guest IDs
3. Exclude already invited and revoked users
4. Create stances for remaining users

Called when group members are added or poll is created.

#### group_members_added(group_id)

Called when new members join a group:

1. Find all active, open-voting polls in the group
2. For each, create stances for new members

#### group_members_removed(group_id, removed_user_ids, actor_id, revoked_at)

Called when members leave a group:

1. Find all active polls in the group
2. Revoke stances for removed users
3. Update poll counts

### Results Calculation

#### calculate_results(poll, poll_options)

Computes poll results for display:

1. Sort options by priority or total_score (based on poll type)
2. For each option, calculate:
   - score_percent: Option score / total poll score
   - voter_percent: Option voters / total voters
   - max_score_percent: Option score / highest option score
   - target_percent: For count polls with agree_target
   - average: Total score / voter count
   - test_result: Whether threshold test passes
3. Add "none of the above" entry if enabled
4. Add "undecided" entry if not a meeting poll

Returns array of result objects with id, name, icon, rank, score, percentages, voter_ids, color.

### Other Operations

#### discard(poll:, actor:)

Soft deletes a poll:

1. Authorize destroy permission
2. Set discarded_at and discarded_by
3. Remove stance events from discussion
4. Clear created_event user and child_count
5. Update discussion sequence info
6. Publish via MessageChannelService

#### add_to_thread(poll:, params:, actor:)

Adds a standalone poll to a discussion:

1. Authorize update on both poll and discussion
2. Update poll with discussion_id and group_id
3. Reparent poll's created_event under discussion
4. Recalculate sequences
5. Add stance events to discussion timeline
6. Trigger search reindex

### Receipt Building

#### build_receipts(poll)

Creates receipt records for auditing:

For each latest stance, create receipt with:
- poll_id, voter_id, inviter_id, invited_at
- vote_cast: boolean indicating if they voted (null for anonymous quorum not met)

Used by receipts endpoint and closing process.

---

## StanceService

**File:** `/app/services/stance_service.rb`

### Core Operations

#### create(stance:, actor:)

Creates a new vote:

1. Authorize vote_in on poll
2. Set participant to actor
3. Set cast_at to now (if not already set)
4. Clear any revoked state
5. Save stance
6. Update poll counts
7. Publish StanceCreated event

Also marks poll notifications as read for the voter.

#### update(stance:, actor:, params:)

Updates an existing vote:

1. Authorize update on stance
2. Determine if this is an update (already has cast_at)
3. Build replacement stance
4. Apply new attributes

If updating a cast vote in a discussion, and option_scores changed, and more than 15 minutes since last save:
- Create new stance record (preserves history)
- Mark old stance as not latest
- Publish StanceCreated event

Otherwise:
- Update existing stance in place
- Publish StanceUpdated or StanceCreated event

This preserves vote history when users change their position.

#### uncast(stance:, actor:)

Removes a vote (returns to undecided state):

1. Authorize uncast on stance
2. Build replacement stance without cast_at
3. Mark old stance as not latest
4. Save new stance
5. Update poll counts

#### redeem(stance:, actor:)

Transfers a stance to a different user (for guest invitations):

1. Skip if actor already has a stance in this poll
2. Verify stance is redeemable by actor
3. Update participant to actor
4. Set accepted_at

Used when a guest user claims their invitation.

---

## OutcomeService

**File:** `/app/services/outcome_service.rb`

### Core Operations

#### create(outcome:, actor:, params:)

Creates a poll outcome:

1. Authorize create on outcome
2. Authorize recipients for announcement
3. Set author to actor
4. Validate
5. Mark all previous outcomes as not latest
6. Save outcome
7. Find or create recipient users
8. Broadcast outcome_create
9. Publish OutcomeCreated event

#### update(outcome:, actor:, params:)

Updates an existing outcome:

1. Authorize update
2. Authorize any new recipients
3. Apply attributes (review_on, statement, poll_option_id, etc.)
4. Validate and save
5. Update versions count
6. Handle recipients
7. Broadcast outcome_update
8. Publish OutcomeUpdated event

#### invite(outcome:, actor:, params:)

Announces an outcome to users:

1. Authorize announce
2. Authorize recipients
3. Find or create users
4. Publish OutcomeAnnounced event

#### publish_review_due

Scheduled job for outcome reviews:

1. Find outcomes where review_on = today and no review_due event published
2. Publish OutcomeReviewDue event for each

---

## Event Integration

### Poll Events

| Event | When Published | Includes |
|-------|----------------|----------|
| PollCreated | poll.create | Notifies stances if requested |
| PollEdited | poll.update | Notifies new recipients |
| PollAnnounced | poll.invite | Notifies invited users |
| PollReminder | poll.remind | Notifies selected voters |
| PollClosingSoon | 24h before close | Notifies based on notify_on_closing_soon |
| PollClosedByUser | poll.close | Manual close by admin |
| PollExpired | automatic close | Auto-close at closing_at |
| PollReopened | poll.reopen | Notifies voters |

### Stance Events

| Event | When Published |
|-------|----------------|
| StanceCreated | New vote or major vote change |
| StanceUpdated | Minor vote update (within 15 min) |

### Outcome Events

| Event | When Published |
|-------|----------------|
| OutcomeCreated | outcome.create |
| OutcomeUpdated | outcome.update |
| OutcomeAnnounced | outcome.invite |
| OutcomeReviewDue | Scheduled review date reached |

---

## EventBus Broadcasts

The services broadcast to EventBus for cross-cutting concerns:

- `poll_create`: After poll creation
- `poll_update`: After poll update
- `poll_reopen`: After poll reopened
- `outcome_create`: After outcome creation
- `outcome_update`: After outcome update

EventBus listeners (configured in `/config/initializers/event_bus.rb`) handle:
- Updating DiscussionReader states
- Publishing real-time updates

---

## Workers

### CloseExpiredPollWorker

**File:** `/app/workers/close_expired_poll_worker.rb`

Handles automatic poll closing:

1. Find poll by ID
2. Return early if already closed
3. Call PollService.do_closing_work
4. Publish PollExpired event

Enqueued by PollService.expire_lapsed_polls which runs periodically.

### GenericWorker Usage

Several operations use GenericWorker for background processing:

- `SearchService.reindex_by_poll_id`: After poll changes
- `PollService.group_members_added`: After members added to group

---

## Transaction Safety

All service methods that modify multiple records use database transactions:

- Poll creation with stances
- Poll closing with receipt creation and stance scrubbing
- Stance updates with old/new stance swaps
- Outcome creation with marking previous as not latest

This ensures data consistency if any step fails.
