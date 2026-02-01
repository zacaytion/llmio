# Groups Domain: Services

**Generated:** 2026-02-01
**Confidence:** 4/5

---

## Table of Contents

1. [GroupService](#groupservice)
2. [MembershipService](#membershipservice)
3. [MembershipRequestService](#membershiprequestservice)
4. [UserInviter](#userinviter)
5. [Service Interactions](#service-interactions)

---

## GroupService

**Location:** `/app/services/group_service.rb`

### Purpose

Handles all group lifecycle operations: creation, updates, destruction, member invitations, exports, and merging.

### Methods

#### `create(group:, actor:, skip_authorize: false)`

Creates a new group.

**Flow:**
1. Authorize actor can create the group (unless skipped)
2. Return false if group is invalid
3. If parent group (not subgroup):
   - Attach random cover photo from default set
   - Set actor as creator
   - Create new Subscription record
4. Save the group
5. Add actor as admin of the group
6. Broadcast 'group_create' via EventBus

**Authorization:** `actor.ability.authorize!(:create, group)`

#### `update(group:, params:, actor:)`

Updates group attributes.

**Flow:**
1. Authorize actor can update the group
2. Assign attributes and files from params
3. Apply `group_privacy` setting if provided (triggers privacy cascade)
4. Create PrivacyChange object to track privacy state changes
5. Return false if invalid
6. Save group
7. Commit privacy changes (updates child resources if needed)
8. Broadcast 'group_update' via EventBus

**Authorization:** `actor.ability.authorize!(:update, group)`

#### `destroy(group:, actor:)`

Archives a group with a delayed hard delete.

**Flow:**
1. Authorize actor can destroy the group
2. Send warning email to all group admins
3. Archive the group (sets archived_at)
4. Schedule `DestroyGroupWorker` to run in 2 weeks
5. Broadcast 'group_destroy' via EventBus

**Authorization:** `actor.ability.authorize!(:destroy, group)`

#### `invite(group:, params:, actor:)`

Invites users to group(s) by email or user ID.

**Flow:**
1. Parse invited group IDs (supports inviting to multiple groups at once)
2. Restrict group IDs to single organization (parent + subgroups)
3. Authorize via `UserInviter.authorize_add_members!`:
   - Check subscription is active
   - Check member limits not exceeded
   - Check actor can add members to each group
4. Create/find users via `UserInviter.where_or_create!`
5. For each group:
   - Unrevoke any revoked memberships for these users
   - Create new membership records (ignoring duplicates)
   - Auto-accept invitations for users already in organization
   - Update membership counts
   - Trigger `PollService.group_members_added` to add to active polls
6. Publish `Events::MembershipCreated` event
7. Return active memberships for the primary group

**Parameters:**
- `recipient_emails` - Array of email addresses
- `recipient_user_ids` - Array of user IDs
- `invited_group_ids` - Array of group IDs to invite to (optional)
- `recipient_message` - Custom invitation message

#### `move(group:, parent:, actor:)`

Moves a group to become a subgroup of another group.

**Flow:**
1. Authorize actor can move the group (requires site admin)
2. Update handle to include parent's handle
3. Update parent_id and clear subscription_id
4. Broadcast 'group_move' via EventBus

**Authorization:** Requires site admin (`user.is_admin`)

#### `merge(source:, target:, actor:)`

Merges two groups, moving all content from source to target.

**Flow (in transaction):**
1. Authorize actor can merge both groups (requires site admin)
2. Move all subgroups to target
3. Move all discussions to target
4. Move all polls to target
5. Move all membership requests to target
6. Move memberships (excluding users already in target)
7. Destroy source group

**Authorization:** Requires site admin for both groups

#### `export(group:, actor:)`

Exports group data as JSON.

**Flow:**
1. Authorize actor can show the group
2. Get group IDs user can access within this organization
3. Enqueue `GroupExportWorker` with group IDs

#### `suggest_handle(name:, parent_handle:)`

Generates a unique handle for a new group.

**Flow:**
1. Generate parameterized handle from name
2. Prepend parent handle if provided
3. Check for uniqueness
4. Append incrementing number if collision

---

## MembershipService

**Location:** `/app/services/membership_service.rb`

### Purpose

Handles membership lifecycle: invitation redemption, revocation, role changes, and volume settings.

### Methods

#### `redeem(membership:, actor:, notify: true)`

Accepts a membership invitation.

**Flow:**
1. Raise error if already accepted
2. Set accepted_at timestamp
3. Find all pending memberships for this user in the organization
4. Unrevoke any revoked memberships in invited groups
5. Accept all pending memberships in invited groups
6. Transfer memberships from unverified user to actor if different
7. Trigger `PollService.group_members_added` for each group
8. Remove guest access (convert to member access):
   - Clear guest flag on DiscussionReaders
   - Clear guest flag on Stances
9. Unrevoke stances on active polls
10. Publish `Events::InvitationAccepted` (if notify and not already member)

**Key Behavior:**
- Handles the case where invitation was sent to an email that maps to an unverified user
- Accepts all related invitations in the organization at once
- Converts any existing guest access to member access

#### `revoke(membership:, actor:, revoked_at:)`

Revokes a membership (soft delete).

**Flow:**
1. Authorize actor can revoke the membership
2. Call `revoke_by_id` for all groups in the organization
3. Broadcast 'membership_destroy' via EventBus

**Authorization:** Actor can revoke if:
- Site admin
- Self (leaving group)
- Group admin
- Inviter of pending invitation

#### `revoke_by_id(group_ids, user_id, actor_id, revoked_at)`

Internal method to revoke access across multiple groups.

**Flow:**
1. Revoke DiscussionReaders (set revoked_at, clear guest flag)
2. Clear guest flag on Stances
3. Call `PollService.group_members_removed` for each group (removes from active polls)
4. Set revoked_at on memberships
5. Update membership counts

#### `make_admin(membership:, actor:)` / `remove_admin(membership:, actor:)`

Toggles admin status on a membership.

**Flow:**
1. Authorize action
2. Update admin flag
3. Publish `Events::NewCoordinator` (for make_admin only)

**Authorization for make_admin:**
- Group admin
- Self if only member (bootstrap case)
- Parent group admin making self admin of subgroup

#### `make_delegate(membership:, actor:)` / `remove_delegate(membership:, actor:)`

Toggles delegate status on a membership.

**Flow:**
1. Authorize action (requires group admin)
2. Update delegate flag
3. Update user experiences to reflect delegate status
4. Broadcast user update for UI
5. Publish `Events::NewDelegate` (for make_delegate only)

#### `join_group(group:, actor:)`

User self-joins a group (for open groups or parent admins).

**Flow:**
1. Authorize actor can join the group
2. Add actor as member via `group.add_member!`
3. Broadcast 'membership_join_group' via EventBus
4. Publish `Events::UserJoinedGroup`

**Authorization:** Can join if:
- Can show group
- Email is verified
- Group grants membership upon request, OR
- Actor is admin of parent group

#### `set_volume(membership:, params:, actor:)`

Sets notification volume preference.

**Flow:**
1. Authorize actor can update membership
2. If `apply_to_all`:
   - Update all memberships in organization
   - Update all discussion readers in organization
   - Update all stances in organization
3. Otherwise:
   - Update just this membership
   - Update related discussion readers and stances

#### `update(membership:, params:, actor:)`

Updates membership attributes (currently just title).

**Flow:**
1. Authorize actor can update membership
2. Assign title from params
3. Validate and save
4. Update user experiences with titles
5. Broadcast changes

---

## MembershipRequestService

**Location:** `/app/services/membership_request_service.rb`

### Purpose

Handles the request-to-join workflow for groups that require approval.

### Methods

#### `create(membership_request:, actor:)`

Creates a new membership request.

**Flow:**
1. Set requestor to actor
2. Validate request
3. Authorize actor can create the request
4. Save the request
5. Publish `Events::MembershipRequested`

**Authorization:** Must be able to show the group and be logged in

#### `approve(membership_request:, actor:)`

Approves a pending membership request.

**Flow:**
1. Authorize actor can approve (group admin or members if allowed)
2. Mark request as approved (sets response, responder, responded_at)
3. Convert to membership via `convert_to_membership!`
4. Publish `Events::MembershipRequestApproved` with the new membership

#### `ignore(membership_request:, actor:)`

Ignores/rejects a membership request.

**Flow:**
1. Authorize actor can ignore
2. Mark request as ignored (sets response, responder, responded_at)

---

## UserInviter

**Location:** `/app/extras/user_inviter.rb`

### Purpose

Utility class for creating/finding users during invitation flows and checking authorization.

### Methods

#### `authorize_add_members!(parent_group:, group_ids:, emails:, user_ids:, actor:)`

Validates invitation is allowed.

**Checks:**
1. Subscription is active (raises `Subscription::NotActive` if not)
2. Actor can add members to each selected group
3. New member count won't exceed subscription limit (raises `MaxMembersExceeded` if so)

#### `new_members_count(parent_group:, user_ids:, emails:)`

Calculates how many truly new members are being added (not already in organization).

#### `where_or_create!(emails:, user_ids:, audience:, model:, actor:, include_actor:)`

Creates new users for emails or finds existing users.

**Flow:**
1. Process audience (group members, voters, etc.)
2. Identify existing members
3. Identify invitable guests (users in organization but not in model)
4. Apply rate limiting via ThrottleService
5. Create new User records for unknown emails (with actor's timezone/locale)
6. Return all matched/created users

**Rate Limiting:** Daily limit on invitations per user

---

## Service Interactions

### Invitation Flow Diagram

```
GroupService.invite
    |
    +-> UserInviter.authorize_add_members! (check subscription, permissions)
    |
    +-> UserInviter.where_or_create! (create/find users)
    |
    +-> For each group:
    |       |
    |       +-> Unrevoke revoked memberships
    |       +-> Create new memberships (import)
    |       +-> Auto-accept for existing org members
    |       +-> PollService.group_members_added
    |
    +-> Events::MembershipCreated.publish!
```

### Invitation Redemption Flow

```
MembershipService.redeem
    |
    +-> Accept all pending memberships in org
    |
    +-> Unrevoke any revoked memberships
    |
    +-> Transfer memberships from unverified user
    |
    +-> PollService.group_members_added (for each group)
    |
    +-> Convert guest access to member access
    |
    +-> Unrevoke active poll stances
    |
    +-> Events::InvitationAccepted.publish!
```

### Membership Revocation Flow

```
MembershipService.revoke
    |
    +-> revoke_by_id (for org groups)
            |
            +-> Revoke DiscussionReaders
            +-> Clear guest flags on Stances
            +-> PollService.group_members_removed (each group)
            +-> Revoke Memberships
            +-> Update group counts
```

---

## Events Published

| Event | Trigger | Recipients |
|-------|---------|------------|
| `Events::MembershipCreated` | GroupService.invite | Invited users |
| `Events::InvitationAccepted` | MembershipService.redeem | Inviter |
| `Events::NewCoordinator` | MembershipService.make_admin | New admin |
| `Events::NewDelegate` | MembershipService.make_delegate | New delegate |
| `Events::UserJoinedGroup` | MembershipService.join_group | (no notifications) |
| `Events::MembershipRequested` | MembershipRequestService.create | Group admins |
| `Events::MembershipRequestApproved` | MembershipRequestService.approve | Requestor |
| `Events::MembershipResent` | MembershipService.resend | Invited user |

---

## Open Questions

1. **PrivacyChange Class:** The `PrivacyChange` class used in GroupService.update is not in the services directory - needs investigation of where privacy cascade logic lives.

2. **GenericWorker Pattern:** Services use `GenericWorker.perform_async('PollService', 'group_members_added', ...)` for deferred work - this pattern could use better documentation.

3. **ThrottleService Limits:** The exact rate limits for invitations are defined on the User model (`invitations_rate_limit`) but the configuration is not documented.

**Confidence Breakdown:**
- GroupService: 5/5
- MembershipService: 4/5
- MembershipRequestService: 5/5
- UserInviter: 4/5 (some audience logic is in AnnouncementService)
