# Permission Flags - Final Synthesis

## Summary

This document provides the implementation-ready specification for permission flags in Loomio. It synthesizes findings from our research, third-party discovery, and direct source verification.

---

## Complete Permission Flag Enumeration

### Member Permission Flags (11 total)

| Flag | Type | Default | Schema Line | Ability Files | Status |
|------|------|---------|-------------|---------------|--------|
| `members_can_add_members` | boolean | `false` | 427 | `group.rb`, `membership_request.rb` | Active |
| `members_can_add_guests` | boolean | `true` | 474 | `group.rb`, `discussion.rb`, `poll.rb` | Active |
| `members_can_announce` | boolean | `true` | 465 | `group.rb`, `discussion.rb`, `poll.rb` | Active |
| `members_can_create_subgroups` | boolean | `false` | 442 | `group.rb` | Active |
| `members_can_start_discussions` | boolean | `true` | 441 | `group.rb`, `discussion.rb` | Active |
| `members_can_edit_discussions` | boolean | `true` | 429 | `discussion.rb`, `outcome.rb` | Active |
| `members_can_edit_comments` | boolean | `true` | 438 | `comment.rb` | Active |
| `members_can_delete_comments` | boolean | `true` | 475 | `comment.rb` | Active |
| `members_can_raise_motions` | boolean | `true` | 439 | `poll.rb` | Active |
| `members_can_vote` | boolean | `true` | 440 | **NONE** | Legacy/Unused |
| `parent_members_can_see_discussions` | boolean | `false` | 419 | N/A (query-based) | Active |

### Admin Permission Flags (1 total)

| Flag | Type | Default | Schema Line | Ability Files | Status |
|------|------|---------|-------------|---------------|--------|
| `admins_can_edit_user_content` | boolean | `true` | 471 | `comment.rb` | Active |

### Configuration Settings (NOT Permission Flags)

These columns affect defaults/behavior but are NOT checked in ability classes:

| Column | Type | Default | Schema Line | Purpose |
|--------|------|---------|-------------|---------|
| `new_threads_max_depth` | integer | 3 | 469 | Default nesting depth for new discussions |
| `new_threads_newest_first` | boolean | `false` | 470 | Default sort order for new discussions |
| `can_start_polls_without_discussion` | boolean | `false` | 482 | Feature toggle for standalone polls |
| `listed_in_explore` | boolean | `false` | 472 | Visibility in public explore page |

---

## Permission Inheritance/Cascade Logic

### Key Finding: Permissions Do NOT Cascade

Permission flags are **independent per group**. Subgroups do NOT inherit permission settings from parent groups.

The only inheritance-related behavior is:
1. **`parent_members_can_see_discussions`**: When `true` on a subgroup, members of the parent can VIEW (but not participate in) discussions
2. **`members_can_create_subgroups`**: Controls whether parent group members can CREATE subgroups

### Evidence

```ruby
# From ability/group.rb:87 - Subgroup creation check
can [:create], ::Group do |group|
  group.parent_id &&
  (user_is_member_of?(group.parent_id) && group.parent.members_can_create_subgroups?)
end
```

This checks the PARENT's `members_can_create_subgroups`, not the new subgroup's setting.

### Visibility Cascade

```ruby
# From discussion_query.rb:42
OR (groups.parent_members_can_see_discussions = TRUE
    AND groups.parent_id IN (:user_group_ids))
```

When `parent_members_can_see_discussions = true`:
- Parent group members can SEE discussions in the subgroup
- They CANNOT participate (comment, vote, etc.)
- This is visibility filtering, not ability authorization

---

## Authorization Pattern Details

### Standard Pattern

All permission checks follow this pattern:

```ruby
can [:action], ::Model do |model|
  # Admin override (most cases)
  model.group.admins.exists?(user.id) ||
  # Member can if: flag enabled AND user is member
  (model.group.members_can_X && model.group.members.exists?(user.id))
end
```

### Permission-Specific Authorization

#### `members_can_start_discussions`

**Checked in:** `ability/group.rb:42`, `ability/discussion.rb:25`

**Actions controlled:** `:move_discussions_to`, `:create` discussion

```ruby
can [:create], ::Discussion do |discussion|
  discussion.group.members.exists?(user.id) &&
  (discussion.group.members_can_start_discussions || discussion.group.admins.exists?(user.id))
end
```

---

#### `members_can_edit_discussions`

**Checked in:** `ability/discussion.rb:55`, `ability/outcome.rb:11`

**Actions controlled:** `:update`, `:move`, `:move_comments`, `:pin` discussion; `:create`, `:update` outcome

```ruby
can [:update, :move, :move_comments, :pin], ::Discussion do |discussion|
  !discussion.closed_at &&
  (discussion.admins.exists?(user.id) ||
   (discussion.group.members_can_edit_discussions && discussion.members.exists?(user.id)))
end
```

**Note:** This allows ANY member to edit discussions, not just authors (when flag is true).

---

#### `members_can_edit_comments`

**Checked in:** `ability/comment.rb:13`

**Actions controlled:** `:update` comment

```ruby
can [:update], ::Comment do |comment|
  !comment.discussion.closed_at && (
    (comment.discussion.members.exists?(user.id) &&
     comment.author == user &&  # <-- ONLY own comments
     comment.group.members_can_edit_comments) ||
    (comment.discussion.admins.exists?(user.id) &&
     comment.group.admins_can_edit_user_content)
  )
end
```

**Note:** Unlike `members_can_edit_discussions`, this flag only allows editing **own** comments (`comment.author == user`).

---

#### `members_can_delete_comments`

**Checked in:** `ability/comment.rb:33`

**Actions controlled:** `:destroy` comment

```ruby
can [:destroy], ::Comment do |comment|
  !comment.discussion.closed_at &&
  Comment.where(parent: comment).count == 0 &&  # No replies
  (
    comment.discussion.admins.exists?(user.id) ||
    (comment.author == user &&
     comment.discussion.members.exists?(user.id) &&
     comment.group.members_can_delete_comments)
  )
end
```

**Constraints:**
- Comment must have no replies
- Only own comments (when not admin)

---

#### `members_can_raise_motions`

**Checked in:** `ability/poll.rb:29-30`

**Actions controlled:** `:create` poll

```ruby
can [:create], ::Poll do |poll|
  poll.group_id.nil? ||  # No group = allowed
  poll.group.admins.exists?(user.id) ||  # Admin
  (poll.group.members_can_raise_motions && poll.group.members.exists?(user.id)) ||  # Member
  (poll.group.members_can_raise_motions && poll.discussion.present? &&
   poll.discussion.guests.exists?(user.id))  # Discussion guest
end
```

**Note:** Discussion guests can also create polls if the flag is enabled.

---

#### `members_can_add_members`

**Checked in:** `ability/group.rb:56`, `ability/membership_request.rb:16`

**Actions controlled:** `:add_members`, `:invite_people`, `:announce`, `:manage_membership_requests`; `:show`, `:approve`, `:ignore` membership request

```ruby
can [:add_members, :invite_people, :announce, :manage_membership_requests], ::Group do |group|
  !group.archived_at &&
  ((group.members_can_add_members? && group.members.exists?(user.id)) ||
   group.admins.exists?(user.id))
end
```

---

#### `members_can_add_guests`

**Checked in:** `ability/group.rb:47`, `ability/discussion.rb:45`, `ability/poll.rb:55`

**Actions controlled:** `:add_guests` to group/discussion/poll

```ruby
can [:add_guests], ::Group do |group|
  (group.members_can_add_guests && group.members.exists?(user.id)) ||
  group.admins.exists?(user.id)
end
```

---

#### `members_can_announce`

**Checked in:** `ability/group.rb:62`, `ability/discussion.rb:32`, `ability/poll.rb:41`

**Actions controlled:** `:notify` group; `:announce` discussion; `:announce`, `:remind` poll

```ruby
can [:announce], ::Discussion do |discussion|
  !discussion.closed_at &&
  ((discussion.group.members_can_announce && discussion.members.exists?(user.id)) ||
   discussion.admins.exists?(user.id))
end
```

---

#### `members_can_create_subgroups`

**Checked in:** `ability/group.rb:71, 87`

**Actions controlled:** `:add_subgroup`, `:create` subgroup

```ruby
can [:add_subgroup], ::Group do |group|
  !group.archived_at &&
  (group.members_can_create_subgroups? || group.admins.exists?(user.id))
end

can [:create], ::Group do |group|
  group.is_parent? ||
  (user_is_member_of?(group.parent_id) && group.parent.members_can_create_subgroups?)
end
```

---

#### `admins_can_edit_user_content`

**Checked in:** `ability/comment.rb:14`

**Actions controlled:** Admin `:update` on other users' comments

```ruby
(comment.discussion.admins.exists?(user.id) &&
 comment.group.admins_can_edit_user_content)
```

**Note:** This flag controls whether admins can edit comments written by OTHER users.

---

#### `parent_members_can_see_discussions`

**Checked in:** `discussion_query.rb:42`, `attachments_controller.rb:7`

**Actions controlled:** Query-level visibility filtering (NOT ability authorization)

```ruby
# discussion_query.rb
OR (groups.parent_members_can_see_discussions = TRUE
    AND groups.parent_id IN (:user_group_ids))
```

**Note:** This is NOT an ability check. It filters query results. Parent members can SEE but NOT participate.

---

## Legacy Flag: `members_can_vote`

**Status:** DEPRECATED - Do NOT implement

**Evidence:**
- Column exists in schema (line 440)
- Migration to remove was commented out
- Never checked in any ability file
- Only referenced in `record_cloner.rb` for backward compatibility

**Recommendation:** Exclude from implementation. The original Rails app has effectively abandoned this flag.

---

## Null Group Handling

For discussions without a group (invite-only), use these defaults:

| Flag | Value |
|------|-------|
| MembersCanAddMembers | false |
| MembersCanAddGuests | false |
| MembersCanAnnounce | true |
| MembersCanCreateSubgroups | false |
| MembersCanStartDiscussions | false |
| MembersCanEditDiscussions | false |
| MembersCanEditComments | true |
| MembersCanDeleteComments | true |
| MembersCanRaiseMotions | true |
| AdminsCanEditUserContent | false |

**Note:** There is an inconsistency in the original `Null::Group` concern where `members_can_add_guests` appears in both `true_methods` and `false_methods`. The `false_methods` list is processed after `true_methods`, so `false` takes precedence.

---

## API Serialization

From `group_serializer.rb`, these fields are included in API responses:

```ruby
:parent_members_can_see_discussions,
:members_can_add_members,
:members_can_edit_discussions,
:members_can_edit_comments,
:members_can_delete_comments,
:members_can_raise_motions,
:members_can_start_discussions,
:members_can_create_subgroups,
:members_can_announce,
:members_can_add_guests,
:admins_can_edit_user_content,
:new_threads_max_depth,      # Configuration, not permission
:new_threads_newest_first,   # Configuration, not permission
```

The serializer should match this output for API compatibility.

---

## Paper Trail Versioning

These fields are tracked for audit history:

- parent_members_can_see_discussions
- members_can_add_members
- members_can_edit_discussions
- members_can_edit_comments
- members_can_delete_comments
- members_can_raise_motions
- members_can_start_discussions
- members_can_create_subgroups
- members_can_announce
- admins_can_edit_user_content
- new_threads_max_depth
- new_threads_newest_first
- listed_in_explore

**Note:** `members_can_add_guests` is NOT tracked in paper_trail (potential oversight in original codebase).

---

## Validation Rules

From `group_privacy.rb`:

```ruby
# parent_members_can_see_discussions can only be true if:
# 1. Group is visible to public, OR
# 2. Group is visible to parent members (is_visible_to_parent_members = true)

def parent_members_can_see_discussions_is_valid?
  if is_visible_to_public?
    true
  else
    if parent_members_can_see_discussions?
      is_visible_to_parent_members?
    else
      true
    end
  end
end
```

---

## Summary Table

| Flag | Default | Scope | Edit Own Only | Can Override |
|------|---------|-------|---------------|--------------|
| `members_can_add_members` | false | Group membership | N/A | Admin |
| `members_can_add_guests` | true | Discussion/Poll | N/A | Admin |
| `members_can_announce` | true | Group/Discussion/Poll | N/A | Admin |
| `members_can_create_subgroups` | false | Group hierarchy | N/A | Admin |
| `members_can_start_discussions` | true | Discussion creation | N/A | Admin |
| `members_can_edit_discussions` | true | Any discussion | No | Admin |
| `members_can_edit_comments` | true | Own comments | **Yes** | Admin* |
| `members_can_delete_comments` | true | Own comments | **Yes** | Admin |
| `members_can_raise_motions` | true | Poll creation | N/A | Admin |
| `admins_can_edit_user_content` | true | Other's comments | N/A | N/A |
| `parent_members_can_see_discussions` | false | Query visibility | N/A | Admin |

*Admin can edit others' comments only if `admins_can_edit_user_content` is also true.
