# Events Domain: Models

**Generated:** 2026-02-01
**Confidence:** 5/5

---

## Overview

The Event system is the backbone of Loomio's activity tracking, notifications, and timeline display. It uses Single Table Inheritance (STI) with 42 event subclasses, each implementing specific notification and side-effect behaviors through composable concerns.

---

## Core Event Model

**File:** `/app/models/event.rb`

### Purpose

The Event base class represents any trackable activity in the system. Events are polymorphically associated with an "eventable" (Discussion, Poll, Comment, Stance, Outcome, Membership, etc.) and belong to a discussion for threading purposes.

### Key Associations

- `belongs_to :eventable` (polymorphic) - The model that triggered the event
- `belongs_to :discussion` (optional) - For threading within a discussion
- `belongs_to :user` (optional) - The actor who caused the event
- `belongs_to :parent` (self-referential) - Parent event for nested threading
- `has_many :children` - Child events (replies, stances on polls)
- `has_many :notifications` - Notifications created for this event

### Custom Fields

Events store flexible data in a JSONB `custom_fields` column:
- `pinned_title` - Custom title when event is pinned
- `recipient_user_ids` - List of user IDs to notify
- `recipient_chatbot_ids` - List of chatbot IDs to notify
- `recipient_message` - Custom message for the notification
- `recipient_audience` - Audience type (group, voters, undecided, etc.)
- `stance_ids` - For poll announcements

### Sequence and Position System

Events within a discussion use a hierarchical positioning system:

1. **sequence_id** - Global sequential order within a discussion (1, 2, 3...). Used for read tracking and pagination.

2. **position** - Position among siblings (children of the same parent). Starts at 0 for root, increments for each child.

3. **position_key** - Hierarchical string for sorting, computed as zero-padded positions joined by dashes. Example: "00001-00003-00002" means:
   - Position 1 at depth 0
   - Position 3 at depth 1
   - Position 2 at depth 2

4. **depth** - Nesting level (0 = root/new_discussion, 1 = top-level replies, 2+ = nested)

### Counter Caches

The Event model maintains two counter caches on parent events:
- `child_count` - Direct children count
- `descendant_count` - All nested descendants count (computed via position_key prefix matching)

### STI Lookup

The `sti_find` class method ensures events are instantiated as their proper subclass:
```
PSEUDO: Given an event ID
  1. Find the generic Event record
  2. Get the "kind" column value
  3. Constantize to "Events::{Kind.classify}"
  4. Return the properly-typed instance
```

### Publishing Events

The `Event.publish!` class method is the standard way to create and process events:

```
PSEUDO: Event.publish!(eventable, **args)
  1. Build new event with:
     - kind: derived from subclass name (e.g., "poll_created")
     - eventable: the triggering model
     - eventable_version_id: for Paper Trail audit trail
     - any additional args (user, discussion, recipients, etc.)
  2. Save the event to database
  3. Enqueue PublishEventWorker.perform_async(event.id)
  4. Return the saved event
```

### Parent Event Resolution

The `find_parent_event` method determines the correct parent based on event kind:

| Event Kind | Parent Is |
|------------|-----------|
| new_comment | comment's parent event (reply threading) |
| poll_created | poll's parent event (discussion created event) |
| stance_created | stance's poll's created event |
| poll_closed_by_user | poll's created event |
| outcome_created | outcome's poll's created event |
| discussion_edited | discussion's created event |
| discussion_closed | discussion's created event |

### Max Depth Adjustment

When creating events, the parent is adjusted if the discussion's `max_depth` setting would be exceeded:

```
PSEUDO: max_depth_adjusted_parent
  IF discussion.max_depth equals original_parent.depth
    RETURN original_parent.parent  (move up one level)
  ELSE
    RETURN original_parent
```

This flattens deep threads to respect discussion settings.

### Recipient Calculation

Recipients are determined by combining explicit lists with volume-based queries:

- `all_recipient_user_ids` - Returns unique, compact list from `recipient_user_ids`
- `email_recipients` - Filters by volume (normal/loud) via UsersByVolumeQuery
- `notification_recipients` - Filters by volume (quiet+) and excludes the actor

### The Trigger Method

The base `trigger!` method broadcasts to EventBus:
```
PSEUDO: trigger!
  EventBus.broadcast("{kind}_event", self)
```

Subclasses extend this via concerns (LiveUpdate, Notify::InApp, etc.) that call `super` to chain behaviors.

---

## Notification Model

**File:** `/app/models/notification.rb`

### Purpose

Notifications represent in-app notifications shown to users. Each notification belongs to an event and a user.

### Associations

- `belongs_to :user` - Notification recipient
- `belongs_to :actor` - User who caused the notification
- `belongs_to :event` - Source event

### Key Attributes

- `viewed` - Boolean indicating if the notification has been seen
- `url` - Link to the relevant resource
- `translation_values` - JSONB with data for i18n (name, title, poll_type, etc.)

---

## Event Subclasses (42 Total)

### Discussion Events

**Events::NewDiscussion** (`/app/models/events/new_discussion.rb`)
- Concerns: LiveUpdate, InApp, ByEmail, Mentions, Subscribers, Chatbots
- Published when: Discussion is created
- Recipients: Explicitly specified or mentioned users

**Events::DiscussionEdited** (`/app/models/events/discussion_edited.rb`)
- Concerns: LiveUpdate, InApp, ByEmail, Mentions, Chatbots
- Published when: Discussion title or description changes
- Note: Only adds to timeline if recipient_message is set

**Events::DiscussionAnnounced** (`/app/models/events/discussion_announced.rb`)
- Concerns: InApp, ByEmail, Chatbots
- Published when: Discussion is shared/announced to new members

**Events::DiscussionClosed** (`/app/models/events/discussion_closed.rb`)
- Concerns: LiveUpdate only
- Published when: Discussion is closed
- Note: Appears in timeline, no notifications

**Events::DiscussionReopened** (`/app/models/events/discussion_reopened.rb`)
- Concerns: LiveUpdate only
- Published when: Discussion is reopened

**Events::DiscussionMoved** (`/app/models/events/discussion_moved.rb`)
- Concerns: LiveUpdate only
- Published when: Discussion is moved to another group
- Custom fields: source_group_id

**Events::DiscussionForked** (`/app/models/events/discussion_forked.rb`)
- Published when: Comments are forked to a new discussion

**Events::DiscussionTitleEdited** (`/app/models/events/discussion_title_edited.rb`)
- Published when: Only the title changes (vs full edit)

**Events::DiscussionDescriptionEdited** (`/app/models/events/discussion_description_edited.rb`)
- Published when: Only the description changes

### Comment Events

**Events::NewComment** (`/app/models/events/new_comment.rb`)
- Concerns: ByEmail, Mentions, Chatbots, Subscribers, LiveUpdate
- Published when: Comment is created
- Note: Marks parent comment as read for author; respects should_pin setting

**Events::CommentEdited** (`/app/models/events/comment_edited.rb`)
- Concerns: LiveUpdate, Mentions
- Published when: Comment is edited

**Events::CommentRepliedTo** (`/app/models/events/comment_replied_to.rb`)
- Concerns: InApp, ByEmail
- Published when: Someone replies to a user's comment
- Recipients: Parent comment author only

### Poll Events

**Events::PollCreated** (`/app/models/events/poll_created.rb`)
- Concerns: LiveUpdate, Mentions, Chatbots, ByEmail, InApp, Subscribers
- Published when: Poll is created
- Note: Always pinned by default

**Events::PollEdited** (`/app/models/events/poll_edited.rb`)
- Concerns: LiveUpdate, InApp, ByEmail, Mentions, Chatbots, Subscribers
- Published when: Poll is edited
- Note: Only adds to timeline if recipient_message is set

**Events::PollAnnounced** (`/app/models/events/poll_announced.rb`)
- Concerns: InApp, ByEmail, Chatbots
- Published when: Poll is announced to new voters
- Recipients: Users with stances in stance_ids

**Events::PollReminder** (`/app/models/events/poll_reminder.rb`)
- Concerns: InApp, ByEmail, Chatbots
- Published when: Manual reminder sent for poll
- Note: Not added to discussion timeline (discussion_id: nil)

**Events::PollClosingSoon** (`/app/models/events/poll_closing_soon.rb`)
- Concerns: InApp, Author, ByEmail, Chatbots
- Published when: Poll is about to close (scheduled job)
- Recipients: Based on poll.notify_on_closing_soon setting (author, voters, undecided_voters, nobody)

**Events::PollClosedByUser** (`/app/models/events/poll_closed_by_user.rb`)
- Concerns: LiveUpdate, Chatbots
- Published when: User manually closes a poll
- Note: Uses poll.closed_at as created_at

**Events::PollExpired** (`/app/models/events/poll_expired.rb`)
- Concerns: Author, Chatbots, InApp
- Published when: Poll closes automatically at closing_at time
- Note: Not added to timeline; notifies author if volume allows

**Events::PollReopened** (`/app/models/events/poll_reopened.rb`)
- Concerns: Chatbots
- Published when: Closed poll is reopened
- Note: Uses direct create instead of super, broadcasts manually

**Events::PollOptionAdded** (`/app/models/events/poll_option_added.rb`)
- Published when: New options added to an existing poll

### Stance Events

**Events::StanceCreated** (`/app/models/events/stance_created.rb`)
- Concerns: LiveUpdate, InApp, Mentions, Chatbots, Subscribers
- Published when: User casts a vote
- Note: Respects anonymous/hide_results settings for mentions; marks poll as read

**Events::StanceUpdated** (`/app/models/events/stance_updated.rb`)
- Inherits from: StanceCreated (identical behavior)
- Published when: User updates their vote

### Outcome Events

**Events::OutcomeCreated** (`/app/models/events/outcome_created.rb`)
- Concerns: Mentions, InApp, ByEmail, Chatbots, LiveUpdate, Subscribers
- Published when: Poll outcome is published

**Events::OutcomeAnnounced** (`/app/models/events/outcome_announced.rb`)
- Concerns: InApp, ByEmail
- Published when: Outcome is announced to additional recipients

**Events::OutcomeUpdated** (`/app/models/events/outcome_updated.rb`)
- Published when: Outcome is edited

**Events::OutcomeReviewDue** (`/app/models/events/outcome_review_due.rb`)
- Published when: Outcome review date is reached

### Membership Events

**Events::MembershipCreated** (`/app/models/events/membership_created.rb`)
- Concerns: InApp, ByEmail
- Published when: User is invited to a group
- Recipients: Explicitly specified user IDs

**Events::InvitationAccepted** (`/app/models/events/invitation_accepted.rb`)
- Concerns: InApp, LiveUpdate
- Published when: User accepts group invitation
- Recipients: The inviter only

**Events::UserJoinedGroup** (`/app/models/events/user_joined_group.rb`)
- No concerns (just records the event)
- Published when: User joins via open membership

**Events::UserAddedToGroup** (`/app/models/events/user_added_to_group.rb`)
- Published when: Admin adds user to group

**Events::MembershipRequested** (`/app/models/events/membership_requested.rb`)
- Concerns: InApp, ByEmail
- Published when: User requests to join a group
- Recipients: Group admins

**Events::MembershipRequestApproved** (`/app/models/events/membership_request_approved.rb`)
- Published when: Membership request is approved

**Events::MembershipResent** (`/app/models/events/membership_resent.rb`)
- Published when: Invitation email is resent

**Events::NewCoordinator** (`/app/models/events/new_coordinator.rb`)
- Concerns: InApp
- Published when: User is made an admin
- Recipients: The new coordinator

**Events::NewDelegate** (`/app/models/events/new_delegate.rb`)
- Published when: User is made a delegate (limited admin)

### Mention Events

**Events::UserMentioned** (`/app/models/events/user_mentioned.rb`)
- Concerns: InApp, ByEmail
- Published when: User is @mentioned in content
- Recipients: Users in custom_fields['user_ids']
- Note: Email only if user has email_when_mentioned enabled

**Events::GroupMentioned** (`/app/models/events/group_mentioned.rb`)
- Concerns: InApp, ByEmail
- Published when: @group mention is used
- Recipients: Group members (excluding already mentioned/notified)
- Note: Filters by actor's ability to notify the group

### Reaction Events

**Events::ReactionCreated** (`/app/models/events/reaction_created.rb`)
- Concerns: InApp, LiveUpdate
- Published when: User reacts to content (emoji)
- Recipients: Content author only (unless self-reaction or author left group)

### Administrative Events

**Events::AnnouncementResend** (`/app/models/events/announcement_resend.rb`)
- Concerns: ByEmail only
- Published when: Admin resends announcement emails
- Note: Uses different email method ('group_announced')

**Events::UserReactivated** (`/app/models/events/user_reactivated.rb`)
- Published when: Deactivated user account is reactivated

**Events::UnknownSender** (`/app/models/events/unknown_sender.rb`)
- Published when: Email received from unknown sender

---

## Event Concerns

### Events::LiveUpdate

**File:** `/app/models/concerns/events/live_update.rb`

Publishes real-time updates to connected clients:

```
PSEUDO: trigger!
  1. Call super (chain to other concerns)
  2. If eventable has a group:
     - Publish to group channel via MessageChannelService
  3. If eventable has guests:
     - Publish to each guest's user channel
```

### Events::Notify::InApp

**File:** `/app/models/concerns/events/notify/in_app.rb`

Creates in-app notifications:

```
PSEUDO: trigger!
  1. Call super
  2. Build notifications for each recipient in notification_recipients
  3. Bulk import notifications
  4. Publish each notification to user's channel
```

Customizable methods:
- `notification_recipients` - Who gets notified (default: volume-filtered recipients excluding actor)
- `notification_actor` - Avatar shown (default: event user)
- `notification_url` - Link destination (default: polymorphic path to eventable)
- `notification_translation_values` - I18n values (name, title, poll_type)

### Events::Notify::ByEmail

**File:** `/app/models/concerns/events/notify/by_email.rb`

Sends email notifications:

```
PSEUDO: trigger!
  1. Call super
  2. For each recipient in email_recipients:
     - Filter to active users without spam complaints
     - Enqueue EventMailer.event(user_id, event_id) as background job
```

### Events::Notify::Mentions

**File:** `/app/models/concerns/events/notify/mentions.rb`

Handles @mention notifications:

```
PSEUDO: trigger!
  1. Call super
  2. Return early if silence_mentions? is true
  3. For newly_mentioned_groups:
     - Publish Events::GroupMentioned
  4. For newly_mentioned_users:
     - Publish Events::UserMentioned
```

Also modifies `email_recipients` to exclude mentioned users (they get separate mention notifications).

### Events::Notify::Chatbots

**File:** `/app/models/concerns/events/notify/chatbots.rb`

Sends to configured webhooks:

```
PSEUDO: trigger!
  1. Call super
  2. Enqueue ChatbotService.publish_event! via GenericWorker
```

### Events::Notify::Subscribers

**File:** `/app/models/concerns/events/notify/subscribers.rb`

Sends emails to "loud" volume subscribers:

```
PSEUDO: trigger!
  1. Call super
  2. Find subscribed_recipients:
     - Users with "loud" volume on the subscribed_eventable
     - Excluding: author, mentioned users, mentioned group members, explicit recipients
  3. Enqueue EventMailer for each subscriber
```

### Events::Notify::Author

**File:** `/app/models/concerns/events/notify/author.rb`

Special handling for notifying the eventable's author:

```
PSEUDO: trigger!
  1. Call super
  2. If notify_author? returns true:
     - Send EventMailer to author
```

Used by PollExpired and PollClosingSoon to notify poll authors.

---

## Serialization

**File:** `/app/serializers/event_serializer.rb`

Events are serialized with:
- Position data: id, sequence_id, position, depth, child_count, descendant_count, position_key
- Associations: actor, eventable (polymorphic), discussion, parent, source_group
- Metadata: kind, created_at, pinned, pinned_title, recipient_message
- Custom fields: Only for poll_edited, discussion_edited, discussion_moved

The serializer uses RecordCache to efficiently load associated records across collections.

---

## Key Patterns

### Concern Composition

Event subclasses compose behaviors by including multiple concerns. The trigger chain works through Ruby's `super`:

```
PSEUDO: When StanceCreated.trigger! is called
  1. StanceCreated has no trigger! override
  2. Subscribers.trigger! calls super then sends emails
  3. Chatbots.trigger! calls super then enqueues webhook
  4. Mentions.trigger! calls super then publishes mention events
  5. InApp.trigger! calls super then creates notifications
  6. LiveUpdate.trigger! calls super then publishes to channels
  7. Event.trigger! broadcasts to EventBus
```

### Event-Specific Overrides

Subclasses can override:
- `email_recipients` / `notification_recipients` - Who receives notifications
- `notification_url` - Link destination
- `notification_translation_values` - I18n data
- `notification_actor` - Avatar user
- `silence_mentions?` - Disable mention processing
- `subscribed_eventable` - What model determines subscription

### Pinning

Events can be pinned to appear prominently:
- `pinned` boolean flag
- `pinned_title` custom title in custom_fields
- PollCreated events are pinned by default
- NewComment events respect `should_pin` from the comment

---

## Database Schema Notes

Events table key columns:
- `id`, `created_at`, `updated_at`
- `kind` (string) - STI discriminator
- `eventable_type`, `eventable_id` - Polymorphic association
- `discussion_id` - For threading (nullable for non-discussion events)
- `user_id` - Actor
- `parent_id`, `depth` - Hierarchy
- `sequence_id`, `position`, `position_key` - Ordering
- `child_count`, `descendant_count` - Counter caches
- `pinned` (boolean)
- `custom_fields` (JSONB)
- `eventable_version_id` - Paper Trail version reference
