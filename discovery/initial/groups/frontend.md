# Groups Domain: Frontend Components

**Generated:** 2026-02-01
**Confidence:** 4/5

---

## Table of Contents

1. [Component Overview](#component-overview)
2. [Group Management Components](#group-management-components)
3. [Membership Components](#membership-components)
4. [Frontend Models](#frontend-models)
5. [Key User Flows](#key-user-flows)

---

## Component Overview

**Location:** `/vue/src/components/group/`

The groups domain has 26 Vue components handling group display, settings, membership management, and invitations.

### Component List

| Component | Purpose |
|-----------|---------|
| `page.vue` | Main group page layout |
| `form.vue` | Group settings form (profile, privacy, permissions) |
| `new_form.vue` | New group creation form |
| `avatar.vue` | Group logo/avatar display |
| `join_button.vue` | Button for joining open groups |
| `privacy_button.vue` | Privacy indicator button |
| `demo_banner.vue` | Banner for demo/trial groups |
| `plan_banner.vue` | Subscription plan information banner |
| `discussions_panel.vue` | Discussions list tab |
| `polls_panel.vue` | Polls list tab |
| `members_panel.vue` | Members list and management |
| `files_panel.vue` | Attached files list |
| `emails_panel.vue` | Group email settings |
| `requests_panel.vue` | Membership requests list |
| `tags_panel.vue` | Tag management |
| `invitation_form.vue` | Invite members modal |
| `shareable_link_form.vue` | Shareable link modal |
| `membership_dropdown.vue` | Member actions dropdown |
| `membership_modal.vue` | Membership details modal |
| `membership_request.vue` | Single membership request display |
| `membership_request_form.vue` | Request to join form |
| `membership_requests_card.vue` | Pending requests card |
| `user_name_modal.vue` | Set name for invited user |
| `email_to_group_settings.vue` | Email-to-thread settings |
| `export_data_modal.vue` | Data export modal |
| `member_email_alias_modal.vue` | Email alias configuration |

---

## Group Management Components

### GroupForm (`form.vue`)

**Purpose:** Edit existing group or create new group settings.

**Structure:**
- Three-tab interface: Profile, Privacy, Permissions
- Handles both parent groups and subgroups

**Profile Tab:**
- Cover photo upload (2048x512 px)
- Logo upload (256x256 px)
- Group name field
- URL handle field (auto-suggested)
- Rich text description editor

**Privacy Tab:**
- Privacy radio buttons: Open, Closed, Secret
- Explore listing toggle (parent groups only)
- Membership granting: Request, Approval, Invitation
- Shareable URL display with copy button
- Request-to-join prompt field

**Key Privacy Logic:**
```
IF privacy is 'open':
  - discussionPrivacyOptions = 'public_only'
  - parentMembersCanSeeDiscussions = true
ELSE IF privacy is 'closed':
  - parentMembersCanSeeDiscussions based on user selection
ELSE IF privacy is 'secret':
  - parentMembersCanSeeDiscussions = false
```

**Permissions Tab:**
- Parent members can see discussions (subgroups only)
- Members can add members
- Members can add guests
- Members can announce
- Members can create subgroups (parents only)
- Members can start discussions
- Members can edit discussions
- Members can edit comments
- Members can delete comments
- Members can raise motions
- Allow polls without discussions
- Admins can edit user content

**Data Flow:**
1. Load group from Records store
2. User modifies form fields
3. On submit, call `group.save()`
4. Records.groups.remote handles API call
5. Success: emit closeModal, navigate to group

### GroupNewForm (`new_form.vue`)

Located in `/vue/src/components/start_group/page.vue`

**Purpose:** Simplified form for creating new groups.

**Differs from form.vue:**
- Fewer initial options
- Auto-suggest handle on name change
- Creates new Group record before opening

---

## Membership Components

### MembersPanel (`members_panel.vue`)

**Purpose:** List and manage group members.

**Features:**
- Filter dropdown: All, Accepted, Admins, Delegates, Invitations
- Search by name, email, username
- Invite button (if can add members)
- Shareable link button
- Load more pagination

**Member Display:**
- User avatar
- Name with link to profile
- Email (if visible)
- Title if set
- Bot/Admin/Delegate badges
- Email rejection warning
- Join date or invitation status
- Membership dropdown for actions

**Data Loading:**
```
RecordLoader for memberships with:
  - group_id
  - per: 25
  - order: created_at desc
  - subgroups filter
```

**Subgroup Options:**
- Default: Show only this group's members
- "mine": Show members in subgroups user belongs to
- "all": Show members across all subgroups

### InvitationForm (`invitation_form.vue`)

**Purpose:** Modal for inviting new members.

**Features:**
- Recipients autocomplete (emails and existing users)
- Multi-group selection (parent + subgroups user can invite to)
- Custom invitation message
- Invitations remaining counter (if subscription has limits)
- Too many invitations warning

**Data Flow:**
1. User enters emails or selects users
2. Component calls `announcements/new_member_count` to check limits
3. On submit, POST to `/api/v1/announcements` with:
   - group_id
   - invited_group_ids
   - recipient_emails
   - recipient_user_ids
   - recipient_message
4. Success: flash message, close modal

**Subscription Integration:**
- Displays remaining invitations if max_members set
- Disables invite if over limit or subscription inactive
- Shows upgrade link when needed

### ShareableLinkForm (`shareable_link_form.vue`)

**Purpose:** Generate and manage shareable invitation links.

**Features:**
- Display current shareable URL
- Copy to clipboard button
- Reset token button (invalidates old links)
- Instructions for sharing

### MembershipDropdown (`membership_dropdown.vue`)

**Purpose:** Actions menu for individual members.

**Actions (based on permissions):**
- Make admin / Remove admin
- Make delegate / Remove delegate
- Set title
- Resend invitation (pending only)
- Remove from group
- Leave group (self only)

### MembershipRequestForm (`membership_request_form.vue`)

**Purpose:** Form for users to request group membership.

**Fields:**
- Introduction text (why you want to join)
- Submit button

**Displays:**
- Group's request_to_join_prompt if set

### MembershipRequestsCard (`membership_requests_card.vue`)

**Purpose:** Card showing pending membership requests for admins.

**Features:**
- Count of pending requests
- Link to full requests page

### RequestsPanel (`requests_panel.vue`)

**Purpose:** Full list of membership requests.

**Features:**
- Pending requests list
- Previous (responded) requests list
- Approve/Ignore actions for each request

---

## Frontend Models

### GroupModel

**Location:** `/vue/src/shared/models/group_model.js`

**Key Methods:**
- `parentOrSelf()` - Returns parent group or self
- `selfAndSubgroups()` - Returns array of self and all subgroups
- `membershipFor(user)` - Finds membership for user
- `members()` - Returns array of member User records
- `admins()` / `adminIds()` - Returns admin users/IDs
- `pendingMembershipRequests()` - Filters pending requests
- `privacyIsOpen/Closed/Secret()` - Privacy checks
- `archive()` - Archives the group
- `export()` / `exportCSV()` - Triggers exports
- `fetchToken()` / `resetToken()` - Token management

**Default Values:**
```javascript
{
  parentId: null,
  groupPrivacy: 'secret',
  discussionPrivacyOptions: 'private_only',
  membershipGrantedUpon: 'approval',
  membersCanAddMembers: true,
  membersCanStartDiscussions: true,
  // ... etc
}
```

**Relationships:**
- `hasMany discussions, polls, membershipRequests, memberships, chatbots, subgroups`
- `belongsTo parent, translation, creator`

### MembershipModel

**Location:** `/vue/src/shared/models/membership_model.js`

**Key Methods:**
- `userName()` - Returns user name with title
- `groupName()` - Returns group name
- `saveVolume(volume, applyToAll)` - Updates notification volume
- `resend()` - Resends invitation
- `isMuted()` - Checks if volume is mute

**Default Values:**
```javascript
{
  userId: null,
  groupId: null,
  archivedAt: null,
  inviterId: null,
  volume: null
}
```

**Relationships:**
- `belongsTo group, user, inviter`

### GroupRecordsInterface

**Location:** `/vue/src/shared/interfaces/group_records_interface.js`

**Key Methods:**
- `fuzzyFind(id)` - Finds by id, key, or handle
- `findOrFetch(id)` - Returns cached or fetches from API
- `fetchByParent(parentGroup)` - Fetches subgroups
- `fetchExploreGroups(query, options)` - Search explore directory
- `getExploreResultsCount(query)` - Count explore results
- `getHandle({name, parentHandle})` - Get suggested handle

---

## Key User Flows

### Creating a Group

1. User navigates to `/g/new`
2. `StartGroupPage` component loads
3. User fills name, description, privacy
4. On submit, `group.save()` called
5. API creates group, returns with key
6. Router navigates to `/g/{key}`

### Inviting Members

1. User clicks "Invite" on members panel
2. `GroupInvitationForm` modal opens
3. User enters emails or selects users
4. User selects which groups to invite to
5. User writes custom message
6. On submit, POST to `/api/v1/announcements`
7. API creates memberships, sends invitations
8. Flash success, modal closes
9. Members panel refreshes

### Requesting to Join

1. User views public group page
2. User clicks "Ask to join"
3. `MembershipRequestForm` modal opens
4. User writes introduction
5. On submit, POST to `/api/v1/membership_requests`
6. Event notifies admins
7. User sees pending request status

### Admin Approving Request

1. Admin sees notification of new request
2. Admin navigates to group's requests panel
3. Admin clicks Approve or Ignore
4. API updates request, creates membership if approved
5. Event notifies requestor

### Accepting Invitation

1. User receives invitation email with token link
2. User clicks link, opens app
3. App calls `MembershipService.redeem`
4. All pending invitations in org are accepted
5. User redirected to group page

### Changing Notification Volume

1. User opens membership dropdown
2. User selects "Change notification settings"
3. Volume modal opens
4. User selects volume level
5. Optionally applies to all groups
6. `membership.saveVolume()` called
7. API updates preferences

---

## Component Dependencies

### External Services Used

- `Records` - LokiJS record store
- `AbilityService` - Permission checking
- `EventBus` - Modal management, component communication
- `Flash` - Toast notifications
- `Session` - Current user access
- `AppConfig` - Application configuration

### Common Mixins

- `UrlFor` - URL generation for records
- `WatchRecords` - Reactive record watching

---

## Open Questions

1. **Subgroups UI:** The subgroup management interface is not fully explored - how do users navigate between parent and subgroups?

2. **Billing UI:** The plan/subscription management UI components are referenced but their full implementation is unclear.

3. **Email Aliases:** The email alias functionality (`member_email_alias_modal.vue`) needs further investigation.

**Confidence Breakdown:**
- Group form components: 5/5
- Membership components: 4/5
- Frontend models: 4/5
- User flows: 4/5
