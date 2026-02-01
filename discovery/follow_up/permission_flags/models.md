# Group Permission Model Documentation

## Overview

The Group model uses boolean flags to control what actions members and admins can perform. These flags are checked in the `app/models/ability/` classes which implement CanCanCan authorization.

## Permission Flag Details

### Member Permission Flags

#### `members_can_start_discussions`
- **Default:** true
- **Checked in:**
  - `app/models/ability/group.rb:42` - `:move_discussions_to` action
  - `app/models/ability/discussion.rb:25` - `:create` action
- **Behavior:** When true, regular members can create new discussions in the group.

#### `members_can_edit_discussions`
- **Default:** true
- **Checked in:**
  - `app/models/ability/discussion.rb:55` - `:update, :move, :move_comments, :pin` actions
  - `app/models/ability/outcome.rb:11` - `:create, :update` actions
- **Behavior:** When true, members can edit discussions they participate in (not just their own).

#### `members_can_edit_comments`
- **Default:** true
- **Checked in:**
  - `app/models/ability/comment.rb:13` - `:update` action
- **Behavior:** When true, members can edit their own comments.

#### `members_can_delete_comments`
- **Default:** true
- **Checked in:**
  - `app/models/ability/comment.rb:33` - `:destroy` action
- **Behavior:** When true, members can delete their own comments (if no replies).

#### `members_can_raise_motions`
- **Default:** true
- **Checked in:**
  - `app/models/ability/poll.rb:29-30` - `:create` action
- **Behavior:** When true, members can create polls/proposals within the group.

#### `members_can_vote`
- **Default:** true
- **Checked in:** NOWHERE (legacy flag)
- **Status:** DEPRECATED - Column exists but authorization is controlled at the poll level, not group level.
- **Evidence:** Migration `20201009024231_remove_groups_members_can_vote.rb` has removal commented out.

#### `members_can_add_members`
- **Default:** false
- **Checked in:**
  - `app/models/ability/group.rb:56` - `:add_members, :invite_people, :announce, :manage_membership_requests` actions
  - `app/models/ability/membership_request.rb:16` - `:show, :approve, :ignore` actions
- **Behavior:** When true, members (not just admins) can invite new members.

#### `members_can_add_guests`
- **Default:** true
- **Checked in:**
  - `app/models/ability/group.rb:47` - `:add_guests` action
  - `app/models/ability/discussion.rb:45` - `:add_guests` action
  - `app/models/ability/poll.rb:55` - `:add_guests` action
- **Behavior:** When true, members can invite guests to discussions/polls.

#### `members_can_announce`
- **Default:** true
- **Checked in:**
  - `app/models/ability/group.rb:62` - `:notify` action
  - `app/models/ability/discussion.rb:32` - `:announce` action
  - `app/models/ability/poll.rb:41` - `:announce, :remind` actions
- **Behavior:** When true, members can send notifications/announcements to other members.

#### `members_can_create_subgroups`
- **Default:** false
- **Checked in:**
  - `app/models/ability/group.rb:71` - `:add_subgroup` action
  - `app/models/ability/group.rb:87` - `:create` action (for subgroups)
- **Behavior:** When true, members can create subgroups within this group.

### Admin Permission Flags

#### `admins_can_edit_user_content`
- **Default:** true
- **Checked in:**
  - `app/models/ability/comment.rb:14` - `:update` action
- **Behavior:** When true, group admins can edit comments written by other users.

### Visibility Flags (Used in Queries, Not Ability Classes)

#### `parent_members_can_see_discussions`
- **Default:** false
- **Checked in:**
  - `app/queries/discussion_query.rb:42` - visibility filtering
  - `app/controllers/api/v1/attachments_controller.rb:7` - attachment visibility
- **Behavior:** When true on a subgroup, members of the parent group can see discussions in this subgroup.

## Authorization Pattern

All ability checks follow this pattern:

```ruby
can [:action], ::Model do |model|
  # Admin always can (in most cases)
  model.group.admins.exists?(user.id) ||
  # Member can if flag is enabled AND user is a member
  (model.group.members_can_X && model.group.members.exists?(user.id))
end
```

## Model File Locations

| File | Path |
|------|------|
| Group model | `/app/models/group.rb` |
| Group ability | `/app/models/ability/group.rb` |
| Discussion ability | `/app/models/ability/discussion.rb` |
| Comment ability | `/app/models/ability/comment.rb` |
| Poll ability | `/app/models/ability/poll.rb` |
| Outcome ability | `/app/models/ability/outcome.rb` |
| MembershipRequest ability | `/app/models/ability/membership_request.rb` |
| Discussion query | `/app/queries/discussion_query.rb` |

## Paper Trail Versioning

Permission flags tracked in paper_trail (`app/models/group.rb:132-158`):

**Tracked:**
- `parent_members_can_see_discussions`
- `members_can_add_members`
- `members_can_edit_discussions`
- `members_can_edit_comments`
- `members_can_delete_comments`
- `members_can_raise_motions`
- `members_can_start_discussions`
- `members_can_create_subgroups`
- `members_can_announce`
- `new_threads_max_depth` (configuration, but versioned)
- `new_threads_newest_first` (configuration, but versioned)
- `admins_can_edit_user_content`
- `listed_in_explore` (configuration, but versioned)

**NOT Tracked (missing from paper_trail):**
- `members_can_vote` - Legacy/unused flag
- `members_can_add_guests` - Active permission but not versioned
