# Groups Domain: Models

**Generated:** 2026-02-01
**Confidence:** 4/5

---

## Table of Contents

1. [Group Model](#group-model)
2. [FormalGroup, GuestGroup, NullGroup](#formalgroup-guestgroup-nullgroup)
3. [Membership Model](#membership-model)
4. [MembershipRequest Model](#membershiprequest-model)
5. [Subscription Model](#subscription-model)
6. [Group Hierarchy and Relationships](#group-hierarchy-and-relationships)

---

## Group Model

**Location:** `/app/models/group.rb`

### Purpose

The Group model is the organizational container for collaborative decision-making. Groups hold members, discussions, polls, and subgroups. Every major activity in Loomio happens within a group context.

### Key Concerns Included

- **HasRichText** - HTML sanitization and formatting for the description field
- **CustomCounterCache::Model** - Maintains denormalized counts for performance
- **ReadableUnguessableUrls** - Generates secure, unguessable URL keys
- **SelfReferencing** - Provides `group` and `group_id` methods returning self
- **MessageChannel** - Real-time updates via Redis pub/sub
- **GroupPrivacy** - Privacy configuration logic (see detailed section below)
- **HasEvents** - Links to Event records for activity tracking
- **Translatable** - Supports translation of name and description

### Primary Associations

| Association | Type | Description |
|-------------|------|-------------|
| `creator` | belongs_to User | The user who created the group |
| `parent` | belongs_to Group | Parent group (nil for top-level groups) |
| `subgroups` | has_many Group | Child groups (non-archived only) |
| `subscription` | belongs_to | Billing/plan subscription (parent groups only) |
| `memberships` | has_many | Active membership records |
| `discussions` | has_many | All discussions in this group |
| `polls` | has_many | All polls in this group |
| `membership_requests` | has_many | Join requests from users |
| `chatbots` | has_many | Integration webhooks |
| `tags` | has_many | Organization-level tags |

### Key Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `name` | string | Group display name (required, max 250 chars) |
| `handle` | string | URL-friendly identifier (unique, parameterized) |
| `description` | text | Rich text group description |
| `is_visible_to_public` | boolean | Whether group appears publicly |
| `is_visible_to_parent_members` | boolean | Subgroup visibility to parent members |
| `discussion_privacy_options` | enum | 'public_only', 'private_only', 'public_or_private' |
| `membership_granted_upon` | enum | 'request', 'approval', 'invitation' |
| `archived_at` | datetime | When group was archived (soft delete) |
| `token` | string | Secure token for shareable invite links |

### Permission Settings

Groups contain boolean settings that control what non-admin members can do:

- `members_can_add_members` - Invite new members
- `members_can_add_guests` - Add guests to discussions/polls
- `members_can_announce` - Send notifications to group
- `members_can_start_discussions` - Create new discussion threads
- `members_can_edit_discussions` - Edit any discussion (not just own)
- `members_can_edit_comments` - Edit any comment
- `members_can_delete_comments` - Delete any comment
- `members_can_raise_motions` - Create polls/proposals
- `members_can_create_subgroups` - Create child groups
- `admins_can_edit_user_content` - Admins can edit member content

### Counter Caches

The model maintains denormalized counts for performance:

- `memberships_count` / `pending_memberships_count` / `admin_memberships_count`
- `delegates_count`
- `discussions_count` / `open_discussions_count` / `closed_discussions_count` / `public_discussions_count`
- `polls_count` / `closed_polls_count`
- `subgroups_count`
- `poll_templates_count` / `discussion_templates_count`

### Key Methods

**Membership Management:**
- `add_member!(user, inviter:)` - Add user as member, unrevoking if previously revoked
- `add_admin!(user)` - Add user and grant admin privileges
- `add_members!(users, inviter:)` - Batch add multiple members
- `membership_for(user)` - Find membership record for a user

**Hierarchy Navigation:**
- `parent_or_self` - Returns parent if subgroup, otherwise self
- `self_and_subgroups` - Returns this group plus all subgroups
- `id_and_subgroup_ids` - Returns array of this group's ID and all subgroup IDs
- `is_parent?` / `is_subgroup?` - Check hierarchy position

**Organization Counts:**
- `org_members_count` - Total unique members across org (parent + subgroups)
- `org_discussions_count` / `org_polls_count` - Aggregate counts

**Archival:**
- `archive!` - Sets archived_at for group and all subgroups
- `unarchive!` - Clears archived_at for group and all subgroups

### Validations

- Name is required, max 250 characters
- Subgroups cannot have their own subscription
- Handle must be unique and start with parent handle if subgroup
- Parent group cannot be more than one level deep (no grandparent groups)

---

## FormalGroup, GuestGroup, NullGroup

### FormalGroup

**Location:** `/app/models/formal_group.rb`

An empty subclass of Group. Historically used for STI differentiation but now simply inherits all Group behavior unchanged. Exists for legacy compatibility.

### GuestGroup

**Location:** `/app/models/guest_group.rb`

An empty subclass of Group. Like FormalGroup, exists for legacy STI purposes. Both FormalGroup and GuestGroup are functionally identical to Group.

### NullGroup

**Location:** `/app/models/null_group.rb`, `/app/models/concerns/null/group.rb`

A null object pattern implementation for when a real group doesn't exist. Used for:
- Direct discussions (discussions without a group)
- Standalone polls (polls without a group)
- Default context when group is optional

**Key Characteristics:**
- `id`, `key`, `parent_id` all return nil
- `name` returns translated "Direct" string
- Returns empty collections for members, memberships, etc.
- Returns sensible defaults for permission booleans (most permissions return true)
- `private_discussions_only?` returns true
- `subscription` returns a mock hash indicating active/unlimited

---

## Membership Model

**Location:** `/app/models/membership.rb`

### Purpose

Represents a user's membership in a group. Tracks role, status, and notification preferences.

### Key Concerns

- **CustomCounterCache::Model** - Updates group/user counters
- **HasVolume** - Notification volume preferences (mute/quiet/normal/loud)
- **HasTimeframe** - Time-based filtering
- **HasExperiences** - User experience tracking
- **FriendlyId** - Token-based lookup

### Associations

| Association | Type | Description |
|-------------|------|-------------|
| `group` | belongs_to | The group this membership is for |
| `user` | belongs_to | The member user |
| `inviter` | belongs_to User | Who invited this member |
| `revoker` | belongs_to User | Who revoked this membership |
| `events` | has_many | Membership-related events |

### Key Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `token` | string | Secure token for invitation redemption |
| `admin` | boolean | Whether member has admin privileges |
| `delegate` | boolean | Whether member is a delegate (representative) |
| `title` | string | Custom title/role within group |
| `accepted_at` | datetime | When invitation was accepted |
| `revoked_at` | datetime | When membership was revoked |
| `volume` | enum | Notification level (mute/quiet/normal/loud) |

### Scopes

- `active` - Where revoked_at is nil
- `pending` - Active but accepted_at is nil
- `accepted` - Where accepted_at is not nil
- `revoked` - Where revoked_at is not nil
- `delegates` - Where delegate is true
- `admin` - Where admin is true
- `email_verified` - Joins user and filters by email verification

### Key Methods

- `make_admin!` / `remove_admin!` - Toggle admin status
- `discussion_readers` - DiscussionReaders for this user in this group
- `stances` - Stances (votes) for this user on polls in this group

### Lifecycle

**Pending State:**
- Membership created with accepted_at nil
- Token sent to user via email
- User can redeem token to accept

**Active State:**
- accepted_at is set
- User appears in group members list
- User can participate in group activities

**Revoked State:**
- revoked_at is set
- User no longer has group access
- Can be reinvited (membership unrevoked)

---

## MembershipRequest Model

**Location:** `/app/models/membership_request.rb`

### Purpose

Represents a user's request to join a group (for groups with approval-required joining).

### Associations

| Association | Type | Description |
|-------------|------|-------------|
| `group` | belongs_to | The group being requested |
| `requestor` | belongs_to User | User requesting membership |
| `responder` | belongs_to User | Admin who responded to request |

### Key Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `introduction` | text | Why user wants to join (max 250 chars) |
| `response` | string | 'approved' or 'ignored' |
| `responded_at` | datetime | When admin responded |

### Scopes

- `pending` - Where response is nil (awaiting decision)
- `responded_to` - Where response is not nil

### Key Methods

- `approve!(responder)` - Mark as approved and set response details
- `ignore!(responder)` - Mark as ignored
- `convert_to_membership!` - Create actual membership from approved request

### Validations

- Cannot request if already a member
- Cannot have duplicate pending requests
- Introduction max 250 characters

---

## Subscription Model

**Location:** `/app/models/subscription.rb`

### Purpose

Manages billing and plan limits for groups. Only parent groups have subscriptions; subgroups inherit from parent.

### Key Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `plan` | string | Plan name (trial, demo, or paid plan name) |
| `state` | string | 'active', 'on_hold', 'pending', 'canceled' |
| `max_members` | integer | Member limit (nil for unlimited) |
| `max_threads` | integer | Discussion limit (nil for unlimited) |
| `expires_at` | datetime | When subscription expires |
| `renews_at` | datetime | When subscription renews |
| `payment_method` | enum | 'chargify', 'manual', 'barter', 'paypal' |
| `owner` | belongs_to User | Account owner for billing |

### Key Methods

- `self.for(group)` - Get subscription for any group (goes to parent if subgroup)
- `is_active?` - Check if subscription is in active state and not expired

### Scopes

- `active` - Active/pending/on_hold and not expired
- `expired` - Active state but past expiration
- `canceled` - State is canceled

### Integration

When inviting members, the UserInviter checks subscription limits:
- If max_members is set and would be exceeded, raises `MaxMembersExceeded`
- If subscription is not active, raises `NotActive`

---

## Group Hierarchy and Relationships

### Parent-Subgroup Structure

Loomio supports a single level of nesting:
- **Parent Groups** - Top-level organizations with their own subscription
- **Subgroups** - Child groups that belong to a parent

Subgroups:
- Cannot have their own subscription (use parent's)
- Handle must start with parent's handle (e.g., "parent-handle-subgroup")
- Can inherit visibility settings from parent
- Can allow parent members to see discussions

### Handle Rules

- Handles are URL-friendly identifiers (parameterized)
- Must be globally unique
- Subgroup handles must begin with parent handle + hyphen
- If parent has no handle, subgroup cannot have one

### Privacy Cascade

The `GroupPrivacy` concern manages three privacy modes:

**Open:**
- `is_visible_to_public = true`
- `discussion_privacy_options = 'public_only'`
- Anyone can see group and discussions
- Membership can be granted upon request

**Closed:**
- `is_visible_to_public = true` (group visible, content private)
- `discussion_privacy_options = 'private_only'` or 'public_or_private'
- Users can see group exists but not discussions
- Membership requires approval or invitation

**Secret:**
- `is_visible_to_public = false`
- `discussion_privacy_options = 'private_only'`
- Group hidden from non-members
- Membership by invitation only

### Subgroup Visibility Options

For subgroups of secret parents:
- Only 'closed' or 'secret' privacy allowed
- Cannot be listed in explore

For subgroups visible to parent members:
- `is_visible_to_parent_members = true`
- `parent_members_can_see_discussions = true` (optional)

---

## Open Questions

1. **GuestGroup Usage:** Unclear where GuestGroup is actually instantiated vs FormalGroup. Both appear to be legacy remnants.

2. **Subscription Plans:** The exact plan tiers and their features are defined in SubscriptionConcern which may be in a separate/private module.

3. **Category Field:** Groups have a `category` and `category_id` field whose purpose and usage is not fully documented.

**Confidence Breakdown:**
- Group model structure: 5/5
- Membership lifecycle: 4/5
- Subscription integration: 3/5 (limited visibility into billing logic)
- Privacy rules: 4/5
