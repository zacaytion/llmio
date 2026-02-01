# Groups Domain: Controllers

**Generated:** 2026-02-01
**Confidence:** 4/5

---

## Table of Contents

1. [GroupsController](#groupscontroller)
2. [MembershipsController](#membershipscontroller)
3. [MembershipRequestsController](#membershiprequestscontroller)
4. [Query Objects](#query-objects)

---

## GroupsController

**Location:** `/app/controllers/api/v1/groups_controller.rb`
**Inherits:** `Api::V1::RestfulController`

### Standard CRUD Actions

Inherited from RestfulController, delegates to GroupService:
- `create` -> `GroupService.create`
- `update` -> `GroupService.update`
- `destroy` -> `GroupService.destroy`

### Custom Actions

#### `GET /api/v1/groups`

Lists groups in the explore directory.

**Query Parameters:**
- `xids` - Comma-separated group IDs to fetch (format: "1x2x3")
- `q` - Search query (filters by name)
- `order` - Sort order:
  - `memberships_count` / `memberships_count_asc`
  - `created_at` / `created_at_asc`

**Behavior:**
- If `xids` provided, fetches specific groups visible to user
- Otherwise, searches explore groups with ordering
- Requires signed-in user if `restrict_explore_to_signed_in_users` feature enabled

**Records Source:** `Queries::ExploreGroups`

#### `GET /api/v1/groups/:id`

Shows a single group.

**Behavior:**
- Loads and authorizes group
- Auto-accepts pending membership if user has invitation token
- Returns group with associations

#### `GET /api/v1/groups/:id/subgroups`

Lists subgroups of a group.

**Authorization:** Must be able to show the group

**Returns:** Subgroups the current user can view

#### `POST /api/v1/groups/:id/token`

Gets the shareable invitation token for a group.

**Authorization:** `invite_people` permission on group

**Returns:** Group with token included in response

#### `POST /api/v1/groups/:id/reset_token`

Generates a new shareable invitation token.

**Authorization:** `invite_people` permission on group

**Behavior:** Creates new secure token, invalidating the old one

#### `GET /api/v1/groups/suggest_handle`

Suggests a URL handle for a new group.

**Parameters:**
- `name` - Proposed group name
- `parent_handle` - Parent group's handle (optional)

**Returns:** `{ handle: "suggested-handle" }`

#### `GET /api/v1/groups/count_explore_results`

Counts groups matching a search query.

**Parameters:**
- `q` - Search query

**Returns:** `{ count: N }`

#### `POST /api/v1/groups/:id/upload_photo`

Uploads a logo or cover photo.

**Parameters:**
- `file` - Image file
- `kind` - Either 'logo' or 'cover_photo'

**Authorization:** Must be able to update the group

#### `POST /api/v1/groups/:id/export`

Initiates JSON export of group data.

**Authorization:** `export` permission on group

**Behavior:** Queues `GroupExportWorker`

#### `POST /api/v1/groups/:id/export_csv`

Initiates CSV export of group data.

**Authorization:** `export` permission on group

**Behavior:** Queues `GroupExportCsvWorker`

---

## MembershipsController

**Location:** `/app/controllers/api/v1/memberships_controller.rb`
**Inherits:** `Api::V1::RestfulController`

### Standard CRUD Actions

- `destroy` -> `MembershipService.revoke` (revokes rather than hard deletes)
- `update` -> `MembershipService.update`

### Custom Actions

#### `GET /api/v1/memberships`

Lists memberships with filtering and search.

**Query Parameters:**
- `group_id` - Filter by group
- `subgroups` - Include subgroup memberships: 'mine', 'all', or none
- `user_xids` - Filter by user IDs (format: "1x2x3")
- `filter` - Status filter: 'admin', 'delegate', 'pending', 'accepted'
- `q` - Search by name, email, or username
- `order` - Sort order

**Authorization:** Can view memberships for:
- Groups user is a member of
- Subgroups where user is admin of parent

#### `GET /api/v1/memberships/for_user`

Gets memberships for a specific user.

**Parameters:**
- `user_id` - User to get memberships for

**Returns:** Memberships in shared groups and public groups

#### `GET /api/v1/memberships/my_memberships`

Gets current user's memberships.

**Returns:** All memberships for the current user

#### `POST /api/v1/memberships/join_group`

User joins a group directly (for open groups).

**Parameters:**
- `group_id` or `group_key`

**Authorization:** Via `MembershipService.join_group`

**Returns:** The new membership

#### `POST /api/v1/memberships/:id/resend`

Resends invitation email.

**Authorization:** Must be group admin, invitation must be pending

#### `POST /api/v1/memberships/:id/make_admin`

Promotes member to admin.

**Authorization:** Via `MembershipService.make_admin`

#### `POST /api/v1/memberships/:id/remove_admin`

Demotes admin to regular member.

**Authorization:** Via `MembershipService.remove_admin`

#### `POST /api/v1/memberships/:id/make_delegate`

Designates member as delegate.

**Authorization:** Via `MembershipService.make_delegate`

#### `POST /api/v1/memberships/:id/remove_delegate`

Removes delegate designation.

**Authorization:** Via `MembershipService.remove_delegate`

#### `PATCH /api/v1/memberships/:id/set_volume`

Sets notification volume for membership.

**Parameters:**
- `volume` - Volume level: 'mute', 'quiet', 'normal', 'loud'
- `apply_to_all` - Apply to all memberships in organization (boolean)

#### `POST /api/v1/memberships/:id/save_experience`

Saves an experience flag on the membership.

**Parameters:**
- `experience` - Experience key to mark as completed

#### `PATCH /api/v1/memberships/:id/user_name`

Sets name for an unverified user (admin-only feature).

**Parameters:**
- `name` - New name
- `username` - New username

**Authorization:** Actor must be admin of a group the user is in, user must be unverified or have no name

---

## MembershipRequestsController

**Location:** `/app/controllers/api/v1/membership_requests_controller.rb`
**Inherits:** `Api::V1::RestfulController`

### Standard CRUD Actions

- `create` -> `MembershipRequestService.create`

### Custom Actions

#### `GET /api/v1/membership_requests/pending`

Lists pending membership requests for a group.

**Parameters:**
- `group_id` or `group_key`

**Authorization:** `manage_membership_requests` permission on group

#### `GET /api/v1/membership_requests/my_pending`

Lists current user's pending requests for a group.

**Parameters:**
- `group_id` or `group_key`

**Authorization:** Can show the group

#### `GET /api/v1/membership_requests/previous`

Lists responded-to membership requests.

**Parameters:**
- `group_id` or `group_key`

**Authorization:** `manage_membership_requests` permission on group

#### `POST /api/v1/membership_requests/:id/approve`

Approves a pending request, creating membership.

**Authorization:** Via `MembershipRequestService.approve`

#### `POST /api/v1/membership_requests/:id/ignore`

Ignores/rejects a pending request.

**Authorization:** Via `MembershipRequestService.ignore`

---

## Query Objects

### GroupQuery

**Location:** `/app/queries/group_query.rb`

Provides query scopes for groups.

#### `start`

Base query with includes for subscription, creator, and parent.

#### `visible_to(user:, chain:, show_public:)`

Filters to groups visible to a user.

**Visibility Rules:**
- Groups user is a member of
- Groups user has guest discussion access to
- Subgroups visible to parent members (if user is parent member)
- Public groups (if show_public is true)

### MembershipQuery

**Location:** `/app/queries/membership_query.rb`

Provides query scopes for memberships.

#### `start`

Base query with includes and joins, filtered to active memberships.

#### `visible_to(user:, chain:)`

Filters to memberships visible to a user.

**Visibility Rules:**
- Memberships in groups user is a member of
- Memberships in subgroups where user is admin of parent

#### `search(chain:, params:)`

Applies search and filter parameters.

**Parameters:**
- `group_id` - Specific group
- `subgroups` - Include subgroups: 'mine', 'all'
- `user_ids` - Filter by specific users
- `filter` - Status: 'admin', 'delegate', 'pending', 'accepted'
- `q` - Text search on name, email, username

---

## API Response Format

### Group Response

```
{
  "groups": [
    {
      "id": 1,
      "key": "abc123",
      "name": "Group Name",
      "handle": "group-handle",
      "description": "<p>Rich text description</p>",
      "group_privacy": "closed",
      "memberships_count": 50,
      "subscription": {
        "max_members": 100,
        "plan": "pro",
        "active": true
      },
      ...
    }
  ],
  "parent_groups": [...],
  "memberships": [...],
  "tags": [...]
}
```

### Membership Response

```
{
  "memberships": [
    {
      "id": 1,
      "group_id": 1,
      "user_id": 5,
      "admin": true,
      "delegate": false,
      "accepted_at": "2025-01-15T10:30:00Z",
      "volume": "normal",
      ...
    }
  ],
  "users": [...],
  "groups": [...]
}
```

---

## Authorization Summary

| Endpoint | Permission Required |
|----------|---------------------|
| Show group | `show` - public, member, parent member if visible, or valid token |
| Update group | `update` - group admin |
| Destroy group | `destroy` - group admin |
| Invite members | `add_members` - admin or member if allowed |
| Get token | `invite_people` - admin or member if add_members allowed |
| List memberships | Member of group or admin of parent |
| Make admin | Group admin, self if only member, or parent admin |
| Manage requests | `manage_membership_requests` - admin or member if add_members allowed |

---

## Open Questions

1. **ExploreGroups Query:** The `Queries::ExploreGroups` class is referenced but not in the standard queries folder - likely in `/app/extras/queries/`.

2. **RestfulController Magic:** The base controller provides significant magic for loading resources, authorizing, and delegating to services that isn't fully documented here.

3. **Announcement Endpoint:** Member invitations also go through `/api/v1/announcements` endpoint which is not covered in this document.

**Confidence Breakdown:**
- GroupsController: 4/5
- MembershipsController: 4/5
- MembershipRequestsController: 5/5
- Query objects: 4/5
