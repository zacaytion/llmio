# Events Domain: Tests

**Generated:** 2026-02-01
**Confidence:** 5/5

---

## Overview

The events domain has comprehensive test coverage across model behavior, service functionality, controller endpoints, and integration scenarios.

---

## Model Tests

**File:** `/spec/models/event_spec.rb`

### Purpose

Tests event publishing, notification routing, and email delivery based on volume settings.

### Test Setup

The spec creates a complex scenario with users at different volume levels:

```
PSEUDO: Setup users
- user_left_group: Left the group (should never receive)
- user_thread_loud: Thread volume = loud (receives all)
- user_thread_normal: Thread volume = normal (receives in-app + important emails)
- user_thread_quiet: Thread volume = quiet (receives in-app only)
- user_thread_mute: Thread volume = mute (receives nothing)
- user_membership_loud/normal/quiet/mute: Group membership volumes
- user_mentioned: Will be @mentioned
- user_motion_closing_soon: Has email_when_proposal_closing_soon enabled
```

### Test Cases

#### `new_comment`

Tests that NewComment event:
- Sends emails to subscribed recipients
- Includes thread_loud and membership_loud users
- Excludes left_group users
- Excludes normal/quiet/mute users from subscriber emails
- Excludes comment author
- Excludes mentioned users (they get separate notification)
- Excludes parent comment author (gets comment_replied_to)

#### `user_mentioned`

Tests mention notifications:
- UserMentioned event is created for @mentioned users
- Email recipients include users with email_when_mentioned enabled
- Notification recipients include mentioned users
- Expects exactly 1 user in each recipient list

#### `new_discussion`

Tests discussion creation:
- Mentioned users receive notifications
- Subscribers receive emails
- UserMentioned event contains correct user_ids

#### `poll_created`

Tests poll creation:
- Mentioned users in poll details receive notifications
- Subscriber emails are sent
- Webhook is called if configured

#### `poll_edited`

Tests poll editing:
- Newly mentioned users receive UserMentioned event
- Previously mentioned users are not re-notified

#### `poll_closing_soon`

Tests closing soon notifications with different settings:

**notify_on_closing_soon = 'voters':**
- Notifies all poll voters
- Respects individual volume settings
- Emails loud/normal voters
- In-app notifies loud/normal/quiet voters
- Excludes muted voters

**notify_on_closing_soon = 'undecided_voters':**
- Only notifies voters who haven't cast a vote
- Excludes voters with cast_at set

**notify_on_closing_soon = 'author':**
- Only notifies poll author
- Sends exactly 1 email

**notify_on_closing_soon = 'nobody':**
- Sends no notifications
- Sends no emails

#### `poll_expired`

Tests automatic poll closing:
- Creates notification for poll author
- Emails author if their volume allows
- Does not email author if volume is quiet

#### `outcome_created`

Tests outcome publishing:
- Notifies mentioned users
- Notifies explicit recipient_user_ids
- Sends subscriber emails
- Email includes author, mentioned, and loud subscribers

#### `stance_created`

Tests vote creation:
- Notifies poll author if their volume is loud
- Does not notify author if volume is normal
- Does not notify deactivated users

#### `announcement_created` (via PollAnnounced)

Tests poll announcements:
- Does not email users with quiet stance volume
- Sends invitations to recipients
- Can include iCal attachments for meeting outcomes

---

## EventBus Tests

**File:** `/spec/extras/event_bus_spec.rb`

### Purpose

Tests the pub/sub EventBus mechanism.

### Test Cases

#### `listen`

- Activates listener when event name matches
- Passes parameters to listener block
- Does not activate for different event names
- Can register for multiple events at once

#### `deafen`

- Silences specific listener
- Does not silence other events
- Can deafen multiple events at once

#### `clear`

- Removes all registered listeners

---

## EventService Tests

**File:** `/spec/services/event_service_spec.rb`

### Purpose

Tests the repair_discussion functionality for different max_depth settings.

### Test Setup

Creates a discussion with:
- 3 nested comments (comment1 -> comment2 -> comment3)
- A poll with a stance

### Test Cases

#### `repair_discussion` with max_depth: 1 (flat)

Tests that flattening works:
- All comments have depth 1
- All comments have discussion_event as parent
- Poll and stance also have discussion_event as parent

#### `repair_discussion` with max_depth: 2 (branching)

Tests depth 2 threading:
- comment1 has depth 1, parent = discussion_event
- comment2 has depth 2, parent = comment1_event
- comment3 has depth 2, parent = comment1_event (not comment2)
- poll has depth 1, parent = discussion_event
- stance has depth 2, parent = poll_event

#### `repair_discussion` with max_depth: 3 (deep)

Tests deep threading:
- comment1 depth 1, parent = discussion_event
- comment2 depth 2, parent = comment1_event
- comment3 depth 3, parent = comment2_event
- poll depth 1, stance depth 2

---

## Controller Tests

**File:** `/spec/controllers/api/v1/events_controller_spec.rb`

### Purpose

Tests API endpoints for event retrieval and manipulation.

### Test Setup

Creates group, discussion, users with appropriate memberships.

### Test Cases

#### `pinning`

**pin event:**
- Returns 200 status
- Returns updated event in response
- Sets event.pinned to true

#### `index`

**logged out user:**
- Can access events for public discussions
- Gets 403 for private discussions

**logged in user:**
- Returns events filtered by discussion
- Does not return events from other discussions
- Response includes discussion with reader

**with comment:**
- Can find event by comment_id
- Returns 404 for non-existent comment

**with parent_id:**
- Returns child events for specified parent
- Includes parent event in response
- Excludes unrelated events

**paging:**
- Respects `per` parameter for limit
- Respects `from` parameter for offset
- Correctly handles deleted sequence_ids

---

## Integration Tests

**File:** `/spec/models/discussion_event_integration_spec.rb`

### Purpose

Tests the interaction between Discussion, DiscussionReader, and Events when items are deleted.

### Scenario: Two Comments, First Deleted

Tests that after deleting a comment:
- Discussion items_count is updated
- DiscussionReader read_items_count reflects deletions
- Unread counts remain accurate

**user has seen nothing:**
- After delete, unread count is correct (1 item)

**user sees discussion before comments:**
- After delete, unread count is correct

---

## Test Patterns

### Factory Usage

The specs use FactoryBot factories:
- `:user` - Basic user
- `:group` - Group with settings
- `:discussion` - Discussion with author and group
- `:poll`, `:poll_meeting`, `:poll_proposal` - Different poll types
- `:comment` - Comment with discussion and user
- `:stance` - Vote on a poll
- `:outcome` - Poll outcome
- `:membership` - Group membership
- `:chatbot` - Webhook configuration

### Email Counting

Tests track email delivery:
```
PSEUDO:
ActionMailer::Base.deliveries = []  # Reset before test
expect { ... }.to change { emails_sent }.by(N)

def emails_sent
  ActionMailer::Base.deliveries.count
end
```

### Private Method Testing

Event recipient methods are tested via `.send()`:
```
PSEUDO:
event.send(:email_recipients)
event.send(:notification_recipients)
event.send(:subscribed_recipients)
```

### WebMock for Webhooks

Webhook calls are verified:
```
PSEUDO:
expect(WebMock).to have_requested(:post, webhook.server).at_least_once
```

---

## Coverage Gaps

Based on the test files examined:

1. **Missing coverage for:**
   - Some event types (discussion_forked, discussion_title_edited, etc.)
   - Complex position_key scenarios
   - Sequence collision handling
   - Group mention notifications

2. **Well covered:**
   - Volume-based recipient filtering
   - Major event types (comment, poll, stance, outcome)
   - Poll notification settings
   - Repair/migration scenarios
   - API endpoints

---

## Running Tests

```bash
# All event tests
bundle exec rspec spec/models/event_spec.rb
bundle exec rspec spec/services/event_service_spec.rb
bundle exec rspec spec/controllers/api/v1/events_controller_spec.rb
bundle exec rspec spec/extras/event_bus_spec.rb
bundle exec rspec spec/models/discussion_event_integration_spec.rb

# By tag or pattern
bundle exec rspec --tag events

# Single test by line
bundle exec rspec spec/models/event_spec.rb:78
```
