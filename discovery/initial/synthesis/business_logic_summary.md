# Loomio Business Logic Summary

**Generated:** 2026-02-01
**Purpose:** Document key business logic patterns and workflows

---

## Table of Contents

1. [User Authentication Flows](#1-user-authentication-flows)
2. [Group and Membership Lifecycle](#2-group-and-membership-lifecycle)
3. [Discussion and Comment Workflows](#3-discussion-and-comment-workflows)
4. [Poll Voting Mechanics](#4-poll-voting-mechanics)
5. [Event Publishing and Notifications](#5-event-publishing-and-notifications)
6. [Permission Cascade System](#6-permission-cascade-system)

---

## 1. User Authentication Flows

### 1.1 Password Authentication

```
User submits email + password
    |
    v
SessionsController#create
    |
    +-- Check for pending_login_token in session
    |   (if exists and valid, authenticate that user)
    |
    +-- Check for code parameter (magic link code auth)
    |   (if present, find LoginToken by email+code)
    |
    +-- Fall back to Devise warden authentication
    |
    v
On success:
    - Sign in user via Devise
    - Set signed_in cookie
    - Clear pending_login_token from session
    - Broadcast 'session_create' event
    - Return boot payload (user, groups, memberships, notifications)
```

### 1.2 Magic Link Authentication

```
User submits email
    |
    v
LoginTokensController#create
    |
    v
LoginTokenService.create(actor:, user:)
    |
    +-- Find user by email
    +-- Create LoginToken with:
    |     - token: SecureRandom.hex (24 chars)
    |     - code: Random.rand(100000..999999)
    |     - expires_at: 24 hours from now
    +-- Send email with link and code
    |
    v
User clicks link or enters code
    |
    v
LoginTokensController#show (for link)
    |
    +-- Store token in session as pending_login_token
    +-- Redirect to target path
    |
    v
SessionsController#create
    |
    +-- Find pending_login_token
    +-- Verify token.useable? (not used, not expired)
    +-- Mark token as used
    +-- Sign in user
```

### 1.3 OAuth/SSO Authentication

```
User clicks OAuth provider button
    |
    v
Identities::{Provider}Controller#oauth
    |
    +-- Store return URL in session
    +-- Build OAuth URL with client_id, redirect_uri, scope
    +-- Redirect to provider
    |
    v
Provider authenticates user, redirects back
    |
    v
Identities::{Provider}Controller#create
    |
    +-- Exchange authorization code for access_token
    +-- Fetch user profile (uid, email, name)
    +-- Find existing Identity by (uid, provider_type)
    |
    +-- IF identity exists:
    |     Update attributes, sign in linked user
    |
    +-- IF new identity:
    |     +-- IF current_user signed in:
    |     |     Link identity to current_user
    |     +-- ELSE IF user with verified email exists:
    |     |     Link identity to that user
    |     +-- ELSE:
    |     |     Store identity ID in session as pending
    |     |     User must register to complete
    |
    +-- IF LOOMIO_SSO_FORCE_USER_ATTRS:
    |     Sync name/email from provider to user
    |
    v
Redirect to stored return URL or dashboard
```

### 1.4 Registration Flow

```
User submits name + email
    |
    v
RegistrationsController#create
    |
    +-- Check registration allowed:
    |     AppConfig.features[:create_user] OR
    |     User has pending invitation
    |
    +-- Find or initialize user by email
    |
    +-- Check if email can be auto-verified:
    |     pending_membership.user.email matches OR
    |     pending_login_token.user.email matches OR
    |     pending_identity.email matches
    |
    +-- IF auto-verified:
    |     - Mark email_verified = true
    |     - Sign in user immediately
    |     - Handle pending invitations
    |     - Return boot payload
    |
    +-- ELSE:
    |     - Save user (unverified)
    |     - Create and send LoginToken
    |     - Return { success: ok, signed_in: false }
```

---

## 2. Group and Membership Lifecycle

### 2.1 Group Creation

```
User submits group creation form
    |
    v
GroupService.create(group:, actor:)
    |
    +-- actor.ability.authorize!(:create, group)
    |     Requires: verified email, subscription allows, or parent admin
    |
    +-- Group.transaction do
    |     - Set creator to actor
    |     - Set defaults based on parent (if subgroup)
    |     - Save group
    |     - Create creator's membership as admin
    |     - Initialize default templates
    |
    +-- EventBus.broadcast('group_create', group, actor)
    |
    +-- IF parent group exists:
    |     parent.update_child_counts!
    |
    v
Return created group
```

### 2.2 Membership Invitation Flow

```
Admin invites users (emails or existing users)
    |
    v
AnnouncementService.create (or GroupService.invite)
    |
    v
UserInviter.where_or_create(emails:, actor:, model:)
    |
    +-- For each email:
    |     +-- Find existing user by email
    |     +-- OR create new user (unverified, random password)
    |
    v
MembershipService.add_users_to_group(users:, group:, inviter:)
    |
    +-- For each user:
    |     +-- Find or build membership
    |     +-- Set inviter, created via announcement
    |     +-- Save membership (pending state)
    |
    v
Events::MembershipCreated.publish!(membership, actor:)
    |
    +-- Send invitation email via EventMailer
    +-- Create in-app notification
```

### 2.3 Membership Acceptance

```
Invited user clicks invitation link or signs in
    |
    v
accept_pending_membership (controller before_action)
    |
    +-- Find pending membership by token in session
    +-- OR find by invitation_token in URL
    |
    v
MembershipService.redeem(membership:, actor:, notify:)
    |
    +-- actor.ability.authorize!(:redeem, membership)
    |     Requires: membership is redeemable (pending, not revoked)
    |
    +-- Membership.transaction do
    |     - Set user = actor (if invited by email)
    |     - Set accepted_at = Time.current
    |     - Save membership
    |
    +-- EventBus.broadcast('membership_redeem', membership, actor)
    |     Listeners:
    |     - Update DiscussionReaders for user in group discussions
    |     - Add user as voter to active polls (if open voting)
    |
    +-- Events::InvitationAccepted.publish!(membership)
```

### 2.4 Membership Revocation

```
Admin removes member
    |
    v
MembershipService.revoke(membership:, actor:)
    |
    +-- actor.ability.authorize!(:destroy, membership)
    |     Requires: admin of group OR self-remove OR parent admin
    |
    +-- membership.update!(
    |     revoked_at: Time.current,
    |     revoker: actor
    |   )
    |
    +-- Revoke DiscussionReaders for user
    +-- Revoke Stances for user in group polls
    |
    +-- EventBus.broadcast('membership_destroy', membership, actor)
```

### 2.5 Volume Preferences Cascade

```
When determining notification volume for a user:

1. Check Stance volume (if in poll context)
   |
   v
2. Check DiscussionReader volume (if in discussion context)
   |
   v
3. Check Membership volume (group context)
   |
   v
4. Fall back to User default volume

Volume levels:
  - mute (0): No notifications
  - quiet (1): Mentioned only
  - normal (2): Important updates
  - loud (3): All activity
```

---

## 3. Discussion and Comment Workflows

### 3.1 Discussion Creation

```
User submits discussion form
    |
    v
DiscussionService.create(discussion:, actor:, params:)
    |
    +-- actor.ability.authorize!(:create, discussion)
    |     Requires:
    |       - Email verified
    |       - IF group: admin OR member with start_discussions permission
    |       - IF no group: direct discussion allowed
    |
    +-- discussion.author = actor
    +-- Return false unless discussion.valid?
    |
    +-- Discussion.transaction do
    |     - discussion.save!
    |     - Create DiscussionReaders for invited guests
    |     - Create author's DiscussionReader (volume: loud)
    |
    +-- EventBus.broadcast('discussion_create', discussion, actor)
    |
    +-- Events::NewDiscussion.publish!(discussion, actor:, ...)
    |     - Send notifications to group members (based on volume)
    |     - Notify mentioned users
    |     - Publish to chatbots
```

### 3.2 Discussion Reader and Read State

```
User opens discussion
    |
    v
DiscussionReaders maintained per user per discussion:
    - last_read_at: timestamp of last read
    - read_ranges: compressed ranges of read sequence_ids
    - volume: notification preference for this discussion
    - dismissed_at: when inbox item was dismissed

Marking as read:
    |
    v
DiscussionService.mark_as_read(discussion:, actor:, ranges:)
    |
    +-- Find or create DiscussionReader
    +-- Parse ranges string ("1-10,15-20")
    +-- Merge with existing read_ranges using RangeSet
    +-- Update last_read_at
    +-- Broadcast via MessageChannelService for real-time UI update
```

### 3.3 Comment Creation and Threading

```
User submits comment
    |
    v
CommentService.create(comment:, actor:)
    |
    +-- actor.ability.authorize!(:create, comment)
    |     Requires:
    |       - Discussion is open (not closed)
    |       - User is discussion member
    |
    +-- comment.author = actor
    +-- comment.discussion = discussion
    +-- IF reply: set parent_id and inherit depth
    |
    +-- Comment.transaction do
    |     - comment.save!
    |     - Update discussion counters (comments_count)
    |     - Update discussion.last_activity_at
    |
    +-- EventBus.broadcast('comment_create', comment, actor)
    |     - Update author's DiscussionReader
    |
    +-- Events::NewComment.publish!(comment, ...)
    |     - Notify parent comment author
    |     - Notify mentioned users
    |     - Notify subscribers based on volume
```

### 3.4 Discussion Forking

```
User forks comments to new discussion
    |
    v
DiscussionsController#create with forked_event_ids
    |
    +-- Create new discussion (per normal flow)
    |
    +-- EventService.move_comments(
    |     discussion: new_discussion,
    |     forked_event_ids: [ids],
    |     actor: current_user
    |   )
    |
    v
MoveCommentsWorker.perform_async
    |
    +-- For each event in forked_event_ids:
    |     - Update event.discussion_id
    |     - Update eventable (comment/poll) discussion reference
    |     - Recalculate position_key for new thread
    |
    +-- Repair original discussion thread
    +-- Repair new discussion thread
```

---

## 4. Poll Voting Mechanics

### 4.1 Poll Types and Options

```
Poll Types:
  - proposal: Yes/No/Abstain with optional blocking
  - count: Simple tally (like, agree, etc.)
  - check: Multiple checkboxes
  - question: Open-ended with options
  - poll: Standard multiple choice
  - ranked_choice: Rank options by preference
  - score: Assign numeric scores
  - meeting: Time slot voting
  - dot_vote: Allocate limited dots across options

Poll Options:
  - name: Display text
  - meaning: Semantic meaning (agree, disagree, abstain, block)
  - icon: Visual indicator
  - prompt: Help text
  - priority: Sort order
```

### 4.2 Poll Creation

```
User submits poll form
    |
    v
PollService.create(poll:, actor:, params:)
    |
    +-- actor.ability.authorize!(:create, poll)
    |     Requires:
    |       - Email verified
    |       - IF group: admin OR member with raise_motions permission
    |       - IF standalone: always allowed for verified users
    |
    +-- poll.author = actor
    +-- Build poll_options from poll_options_attributes
    |
    +-- Poll.transaction do
    |     - poll.save!
    |     - Create Stances for invited voters
    |     - Create author's Stance (admin, volume: loud)
    |
    +-- EventBus.broadcast('poll_create', poll, actor)
    |
    +-- Events::PollCreated.publish!(poll, actor:, ...)
```

### 4.3 Casting a Vote (Stance)

```
User submits vote
    |
    v
StanceService.create(stance:, actor:)
    |
    +-- actor.ability.authorize!(:vote_in, poll)
    |     Requires:
    |       - User is logged in
    |       - Poll is active (not closed)
    |       - User is authorized voter:
    |         - Has Stance record (invited), OR
    |         - Is group member AND specified_voters_only = false
    |
    +-- Find existing latest stance for this user
    |
    +-- IF existing AND within 15-minute window:
    |     Update existing stance (append to revision history)
    +-- ELSE:
    |     Create new stance, mark previous as not latest
    |
    +-- Stance.transaction do
    |     - stance.participant = actor
    |     - stance.cast_at = Time.current
    |     - stance.latest = true
    |     - Create stance_choices (option scores)
    |     - stance.save!
    |     - Update poll counters
    |
    +-- EventBus.broadcast('stance_create', stance, actor)
    |
    +-- Events::StanceCreated.publish!(stance, ...)
    |     - Live update to poll UI
    |     - Notify poll author (if configured)
```

### 4.4 Poll Results Calculation

```
PollService.calculate_results(poll)
    |
    +-- Gather all latest, cast stances
    |
    +-- For each poll_option:
    |     - Count stances that selected this option
    |     - Sum scores for this option
    |     - Calculate percentage
    |
    +-- Build results structure:
    |     {
    |       poll_option_id => {
    |         score: total_score,
    |         voter_count: count,
    |         voter_percent: percentage
    |       }
    |     }
    |
    +-- Apply poll-type-specific calculations:
    |     - ranked_choice: Borda count or IRV
    |     - dot_vote: Percentage of total dots
    |     - meeting: Time availability matrix

Results visibility controlled by:
  - hide_results: 'off' | 'until_vote' | 'until_closed'
  - show_results?(voted:) method checks user has voted if until_vote
```

### 4.5 Poll Closing

```
Automatic closing (CloseExpiredPollWorker runs periodically):
    |
    +-- Find polls where closing_at <= Time.current AND closed_at IS NULL
    +-- For each poll: PollService.expire(poll:)
    |
    v
Manual closing:
    |
    v
PollService.close(poll:, actor:)
    |
    +-- actor.ability.authorize!(:close, poll)
    |     Requires: poll admin AND poll is active
    |
    +-- Poll.transaction do
    |     - poll.closed_at = Time.current
    |     - poll.closer = actor
    |     - poll.save!
    |     - do_closing_work(poll)
    |
    v
do_closing_work(poll):
    |
    +-- IF poll.anonymous:
    |     Scrub participant data:
    |     - Clear stance.participant_id
    |     - Clear author names from events
    |     - Delete stance revision history
    |
    +-- Publish PollClosedByUser or PollExpired event
    +-- Notify voters of results
```

---

## 5. Event Publishing and Notifications

### 5.1 Event Model Architecture

```
Event (STI base class)
  |
  +-- kind: string (discriminator)
  +-- eventable: polymorphic (Discussion, Poll, Comment, etc.)
  +-- discussion: belongs_to (for threading)
  +-- user: belongs_to (actor who triggered)
  +-- sequence_id: position in discussion timeline
  +-- position_key: hierarchical position (e.g., "00001-00003-00002")
  +-- custom_fields: JSONB for flexible data

42 Event Subclasses (STI via kind column):
  - Events::NewDiscussion
  - Events::DiscussionEdited
  - Events::NewComment
  - Events::PollCreated
  - Events::StanceCreated
  - Events::OutcomeCreated
  - Events::MembershipCreated
  - ... and more
```

### 5.2 Event Publishing Flow

```
Service performs action
    |
    v
Events::{EventType}.publish!(eventable, actor:, ...)
    |
    +-- Build event record with:
    |     - kind
    |     - eventable
    |     - user (actor)
    |     - discussion (from eventable if applicable)
    |     - sequence_id (next in discussion sequence)
    |     - position_key (calculated from parent)
    |     - custom_fields (recipients, message, etc.)
    |
    +-- event.save!
    |
    +-- PublishEventWorker.perform_async(event.id)
    |
    v
PublishEventWorker
    |
    +-- Event.sti_find(event_id)  # Loads correct subclass
    |
    +-- event.trigger!
    |
    v
Concern-based trigger chain (via super):
    |
    +-- Events::LiveUpdate#trigger!
    |     - Publish to Redis/MessageChannel for real-time UI
    |
    +-- Events::Notify::Mentions#trigger!
    |     - Find mentioned users
    |     - Create notifications
    |
    +-- Events::Notify::InApp#trigger!
    |     - Create Notification records
    |     - Publish to user channels
    |
    +-- Events::Notify::ByEmail#trigger!
    |     - Queue email jobs for email_recipients
    |
    +-- Events::Notify::Chatbots#trigger!
    |     - Send webhooks to configured chatbots
    |
    +-- Events::Notify::Subscribers#trigger!
    |     - Notify based on volume preferences
```

### 5.3 Notification Recipient Calculation

```
event.notification_recipients:
    |
    v
Queries::UsersByVolumeQuery.call(
    model: event.eventable,
    event_type: event.kind,
    actor: event.user
)
    |
    +-- Start with potential recipients:
    |     - Group members (if group context)
    |     - Discussion readers (if discussion context)
    |     - Poll voters (if poll context)
    |
    +-- Filter by volume:
    |     Event has volume_requirement (e.g., :loud for all activity)
    |     User must have volume >= requirement
    |
    +-- Apply cascade:
    |     stance_volume > discussion_reader_volume > membership_volume > user_default
    |
    +-- Exclude:
    |     - Actor (don't notify yourself)
    |     - Users who muted
    |     - Users with email_verified = false
    |
    v
Return filtered user list
```

### 5.4 EventBus for Side Effects

```
EventBus (lib/event_bus.rb):
  - Simple pub/sub system
  - Listeners registered in config/initializers/event_bus.rb

Key listeners:

'comment_create':
  - Update author's DiscussionReader with new read position

'stance_create':
  - Update voter's DiscussionReader with new read position

'membership_redeem':
  - Create DiscussionReaders for existing discussions
  - Add voter Stances to open polls with open voting

'discussion_mark_as_read':
  - Publish real-time update via MessageChannelService
```

---

## 6. Permission Cascade System

### 6.1 CanCanCan Ability Architecture

```
User.ability returns Ability::Base instance

Ability::Base (app/models/ability/base.rb):
  |
  +-- prepend Ability::Comment
  +-- prepend Ability::Discussion
  +-- prepend Ability::Group
  +-- prepend Ability::Poll
  +-- prepend Ability::Stance
  +-- prepend Ability::Membership
  +-- ... (23 total ability modules)

Each module defines can/cannot rules for its model type
```

### 6.2 Group Permission Settings

```
Group settings that affect member permissions:

Boolean flags on Group model:
  - members_can_add_members
  - members_can_add_guests
  - members_can_announce
  - members_can_edit_discussions
  - members_can_edit_comments
  - members_can_delete_comments
  - members_can_raise_motions
  - members_can_start_discussions
  - members_can_create_subgroups

Ability check pattern:

can :create, Discussion do |discussion|
  group = discussion.group
  if group
    is_admin?(group) ||
    (is_member?(group) && group.members_can_start_discussions?)
  else
    user.email_verified?  # Direct discussion
  end
end
```

### 6.3 Discussion Permission Flow

```
Ability::Discussion checks:

:show
  - PollQuery.visible_to includes the discussion
  - Based on: group membership, guest access, public visibility

:create
  - Email verified
  - Group admin OR member with members_can_start_discussions
  - OR no group (direct discussion)

:update
  - Author of discussion
  - OR group admin
  - OR member with members_can_edit_discussions

:close / :reopen
  - Group admin
  - OR discussion author

:announce (invite people)
  - Group admin
  - OR member with members_can_announce

:add_guests
  - Group admin
  - OR member with members_can_add_guests
  - AND subscription allows guests
```

### 6.4 Poll Permission Flow

```
Ability::Poll checks:

:show
  - PollQuery.visible_to includes the poll

:create
  - Email verified
  - Group admin OR member with members_can_raise_motions
  - OR standalone poll (no group)

:update
  - Poll admin (author or made admin via stance)
  - AND poll is not closed

:close
  - Poll admin
  - AND poll is active

:reopen
  - Poll admin
  - AND poll is closed
  - AND poll is not anonymous (can't reopen anonymous)

:vote_in
  - User is logged in
  - Poll is active
  - User is invited voter (has Stance) OR is group member with open voting
```

### 6.5 Membership Admin Cascade

```
Admin permissions cascade from parent groups:

Parent group admin can:
  - View subgroup memberships
  - Manage subgroup memberships (if subgroup allows)
  - Promote subgroup members to admin

Delegate role:
  - Delegates can perform some admin actions
  - Configured per group

Self-administration:
  - Solo member can make themselves admin
  - Members can always remove themselves
```

### 6.6 Guest Access System

```
Guest access provides limited permissions without group membership:

DiscussionReader with guest: true
  - Can view discussion
  - Can comment (if discussion is open)
  - Cannot see other group content

Stance with guest: true (poll voter)
  - Can view poll
  - Can vote
  - Cannot see other group content

Guest invitation flow:
  1. Admin invites email to discussion/poll
  2. User created (if new) or DiscussionReader/Stance created
  3. User receives invitation email
  4. User can access specific content via token
  5. Access persists after registration
```

---

## Appendix: Key Pseudo-Code Patterns

### Service Method Template

```
class ModelService
  def self.action(model:, actor:, params: {})
    # 1. Authorization
    actor.ability.authorize!(:action, model)

    # 2. Validation
    return false unless model.valid?

    # 3. Transaction (if multiple writes)
    Model.transaction do
      # 4. Perform operation
      model.assign_attributes(permitted_params)
      model.save!

      # 5. Related updates
      update_related_records(model)
    end

    # 6. Side effects via EventBus
    EventBus.broadcast('model_action', model, actor)

    # 7. Publish event for notifications
    Events::ModelAction.publish!(model, actor: actor)
  end
end
```

### Query Object Template

```
class ModelQuery
  def self.start
    Model.distinct.kept.includes(:associations)
  end

  def self.visible_to(chain:, user:)
    return chain.none if user.nil?

    chain.where(
      # Direct ownership
      author_id: user.id
    ).or(chain.where(
      # Group membership
      group_id: user.group_ids
    )).or(chain.where(
      # Guest access
      id: user.guest_model_ids
    ))
  end

  def self.filter(chain:, params:)
    chain = chain.where(group_id: params[:group_id]) if params[:group_id]
    chain = chain.where(status: params[:status]) if params[:status]
    chain = chain.search(params[:q]) if params[:q].present?
    chain
  end
end
```

### Event Concern Template

```
module Events::Notify::Custom
  def trigger!
    super  # Call next concern in chain

    # Custom notification logic
    notification_recipients.each do |user|
      # Create notification
      # Send email
      # Publish real-time update
    end
  end

  def notification_recipients
    Queries::UsersByVolumeQuery.call(
      model: eventable,
      event_type: kind,
      actor: user
    )
  end
end
```

---

*End of Business Logic Summary*
