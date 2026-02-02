# Permission Flags Analysis

## Executive Summary

This document provides a comprehensive analysis of permission flags in Loomio's Group model, including paper_trail tracking configuration, Null::Group behavior resolution, and a complete flag behavior matrix.

---

## 1. Paper Trail Configuration

### Source: `/Users/z/Code/loomio/app/models/group.rb:132-158`

The Group model uses paper_trail with an explicit `only:` whitelist:

```ruby
has_paper_trail only: [:name,
                       :parent_id,
                       :description,
                       :description_format,
                       :handle,
                       :archived_at,
                       :parent_members_can_see_discussions,
                       :key,
                       :is_visible_to_public,
                       :is_visible_to_parent_members,
                       :discussion_privacy_options,
                       :members_can_add_members,
                       :membership_granted_upon,
                       :members_can_edit_discussions,
                       :members_can_edit_comments,
                       :members_can_delete_comments,
                       :members_can_raise_motions,
                       :members_can_start_discussions,
                       :members_can_create_subgroups,
                       :creator_id,
                       :subscription_id,
                       :members_can_announce,
                       :new_threads_max_depth,
                       :new_threads_newest_first,
                       :admins_can_edit_user_content,
                       :listed_in_explore,
                       :attachments]
```

### Tracked Permission Flags (10 of 12)
| Flag | Tracked |
|------|---------|
| `members_can_add_members` | YES |
| `members_can_edit_discussions` | YES |
| `members_can_edit_comments` | YES |
| `members_can_delete_comments` | YES |
| `members_can_raise_motions` | YES |
| `members_can_start_discussions` | YES |
| `members_can_create_subgroups` | YES |
| `members_can_announce` | YES |
| `members_can_add_guests` | **NO** |
| `members_can_vote` | **NO** |
| `admins_can_edit_user_content` | YES |
| `parent_members_can_see_discussions` | YES |

### Why `members_can_add_guests` is Excluded

**Finding**: `members_can_add_guests` was added later (see schema line 474) and was simply **not added** to the paper_trail `only:` list. This appears to be an oversight rather than intentional exclusion.

**Evidence**:
- The column was added at `/Users/z/Code/loomio/db/schema.rb:474`
- It follows the same naming pattern as other tracked flags
- It IS used in ability checks (`/Users/z/Code/loomio/app/models/ability/group.rb:45-48`)
- No comment or migration explains the exclusion

**Confidence: HIGH** - The paper_trail configuration is explicit and `members_can_add_guests` is simply absent from the list.

### Why `members_can_vote` is Excluded

**Finding**: `members_can_vote` is a **deprecated/unused flag**. A migration was written to remove it but was never executed.

**Evidence**:
- `/Users/z/Code/loomio/db/migrate/20201009024231_remove_groups_members_can_vote.rb:3-4`:
  ```ruby
  # actually remove after 2020-11-01
  # remove_column :groups, :members_can_vote
  ```
- The flag exists in schema (line 440) but is not used in any ability checks
- `Grep` of `app/models/ability/*.rb` for `members_can_vote` returns zero results
- It IS still copied by RecordCloner (`/Users/z/Code/loomio/app/services/record_cloner.rb:122`) for backward compatibility

**Confidence: HIGH** - The migration comment and lack of ability usage confirm this is dead code.

---

## 2. Null::Group Contradiction Resolution

### Source Files
- `/Users/z/Code/loomio/app/models/null_group.rb`
- `/Users/z/Code/loomio/app/models/concerns/null/group.rb`
- `/Users/z/Code/loomio/app/models/concerns/null/object.rb`

### The Apparent Contradiction

In `/Users/z/Code/loomio/app/models/concerns/null/group.rb`:

**true_methods** (lines 62-73):
```ruby
def true_methods
  %w[
    private_discussions_only?
    members_can_raise_motions
    members_can_edit_comments
    members_can_delete_comments
    discussion_private_default
    members_can_announce
    members_can_edit_discussions
    members_can_add_guests
  ]
end
```

**false_methods** (lines 88-99):
```ruby
def false_methods
  %w(
    public_discussions_only?
    is_visible_to_parent_members
    members_can_add_members
    members_can_add_guests        # <-- DUPLICATE
    members_can_create_subgroups
    members_can_edit_discussions  # <-- DUPLICATE
    members_can_start_discussions
    admins_can_edit_user_content
  )
end
```

### Resolution: Order of Application Matters

Looking at `/Users/z/Code/loomio/app/models/concerns/null/object.rb:3-9`:

```ruby
def apply_null_methods!
  apply_null_method :nil_methods,   nil
  apply_null_method :false_methods, false
  apply_null_method :empty_methods, []
  apply_null_method :hash_methods,  {}
  apply_null_method :true_methods,  true    # <-- Applied LAST
  apply_null_method :zero_methods,  0
  apply_null_method :none_methods,  ->(model) { ... }
end
```

**Key Insight**: `true_methods` is applied **AFTER** `false_methods`, so the later definition wins.

The `apply_null_method` function (lines 12-17) uses `define_method`, which **overwrites** any previously defined method of the same name:

```ruby
def apply_null_method(name, value)
  send(name).each do |method, model|
    self.class.send :define_method, method, ->(*args) {
      value.respond_to?(:call) ? value.call(model) : value
    }
  end
end
```

### Final Values for Contradicting Flags in NullGroup

| Flag | In false_methods | In true_methods | **Final Value** |
|------|-----------------|-----------------|-----------------|
| `members_can_add_guests` | YES | YES | **true** |
| `members_can_edit_discussions` | YES | YES | **true** |

### Semantic Intent

The NullGroup represents a "direct discussion" (no group context). The final values make semantic sense:
- `members_can_add_guests = true`: In direct discussions, participants CAN invite others
- `members_can_edit_discussions = true`: In direct discussions, members CAN edit
- `members_can_add_members = false`: There's no "membership" concept in direct discussions
- `members_can_start_discussions = false`: Can't start sub-discussions from a null group
- `admins_can_edit_user_content = false`: No admin role in direct discussions

**Confidence: HIGH** - Ruby's method redefinition semantics are deterministic, and the code execution order is explicit.

---

## 3. Complete Permission Flag Behavior Matrix

### All 12 Permission Flags

| Flag | DB Default | Schema Line | Used in Ability | Paper Trail | NullGroup Value |
|------|-----------|-------------|-----------------|-------------|-----------------|
| `members_can_add_members` | `false` | 427 | YES: group.rb:56 | YES | `false` |
| `members_can_edit_discussions` | `true` | 429 | YES: discussion.rb:55 | YES | `true` |
| `members_can_edit_comments` | `true` | 438 | YES: comment.rb:13 | YES | `true` |
| `members_can_delete_comments` | `true` | 475 | YES: comment.rb:33 | YES | `true` |
| `members_can_raise_motions` | `true` | 439 | YES: poll.rb:29-30 | YES | `true` |
| `members_can_vote` | `true` | 440 | **NO** (deprecated) | NO | N/A |
| `members_can_start_discussions` | `true` | 441 | YES: discussion.rb:25, group.rb:42 | YES | `false` |
| `members_can_create_subgroups` | `false` | 442 | YES: group.rb:71, 87 | YES | `false` |
| `members_can_announce` | `true` | 465 | YES: group.rb:62, discussion.rb:32, poll.rb:41 | YES | `true` |
| `members_can_add_guests` | `true` | 474 | YES: group.rb:47, discussion.rb:45, poll.rb:55 | **NO** | `true` |
| `admins_can_edit_user_content` | `true` | 471 | YES: comment.rb:14 | YES | `false` |
| `parent_members_can_see_discussions` | `false` | 419 | NO (query-level only) | YES | N/A |

### Ability Check Details

#### `members_can_add_members` - `/Users/z/Code/loomio/app/models/ability/group.rb:50-59`
```ruby
can [:add_members, :invite_people, :announce, :manage_membership_requests], ::Group do |group|
  user.is_admin ||
  (
    ((group.members_can_add_members? && group.members.exists?(user.id)) ||
     group.admins.exists?(user.id))
  )
end
```
**Grants**: `:add_members`, `:invite_people`, `:announce`, `:manage_membership_requests` on Group

#### `members_can_edit_discussions` - `/Users/z/Code/loomio/app/models/ability/discussion.rb:51-56`
```ruby
can [:update, :move, :move_comments, :pin], ::Discussion do |discussion|
  discussion.discarded_at.nil? &&
  (discussion.author == user ||
  discussion.admins.exists?(user.id) ||
  (discussion.group.members_can_edit_discussions && discussion.members.exists?(user.id)))
end
```
**Grants**: `:update`, `:move`, `:move_comments`, `:pin` on Discussion

#### `members_can_edit_comments` - `/Users/z/Code/loomio/app/models/ability/comment.rb:11-16`
```ruby
can [:update], ::Comment do |comment|
  !comment.discussion.closed_at && (
    (comment.discussion.members.exists?(user.id) && comment.author == user && comment.group.members_can_edit_comments) ||
    (comment.discussion.admins.exists?(user.id) && comment.group.admins_can_edit_user_content)
  )
end
```
**Grants**: `:update` on Comment (author only + flag required)

#### `members_can_delete_comments` - `/Users/z/Code/loomio/app/models/ability/comment.rb:26-35`
```ruby
can [:destroy], ::Comment do |comment|
  !comment.discussion.closed_at &&
  Comment.where(parent: comment).count == 0 &&
  (
    comment.discussion.admins.exists?(user.id) ||
    (comment.author == user &&
     comment.discussion.members.exists?(user.id) &&
     comment.group.members_can_delete_comments)
  )
end
```
**Grants**: `:destroy` on Comment (author only + no replies + flag required)

#### `members_can_raise_motions` - `/Users/z/Code/loomio/app/models/ability/poll.rb:22-36`
```ruby
can [:create], ::Poll do |poll|
  ...
  (poll.group_id &&
    (
     (poll.group.admins.exists?(user.id) ||
     (poll.group.members_can_raise_motions && poll.group.members.exists?(user.id)) ||
     (poll.group.members_can_raise_motions && poll.discussion.present? && poll.discussion.guests.exists?(user.id)))
    )
  ) ...
end
```
**Grants**: `:create` on Poll (for members and guests if flag is true)

#### `members_can_start_discussions` - `/Users/z/Code/loomio/app/models/ability/discussion.rb:20-27`
```ruby
can :create, ::Discussion do |discussion|
  user.email_verified? &&
  (
    (discussion.group.blank? && user.group_ids.any?) ||
    discussion.group.admins.exists?(user.id) ||
    (discussion.group.members_can_start_discussions && discussion.group.members.exists?(user.id))
  )
end
```
Also used in `/Users/z/Code/loomio/app/models/ability/group.rb:39-43`:
```ruby
can [:move_discussions_to], ::Group do |group|
  user.email_verified? &&
  (group.admins.exists?(user.id) ||
  (group.members_can_start_discussions? && group.members.exists?(user.id)))
end
```
**Grants**: `:create` on Discussion, `:move_discussions_to` on Group

#### `members_can_create_subgroups` - `/Users/z/Code/loomio/app/models/ability/group.rb:67-72, 79-88`
```ruby
can [:add_subgroup], ::Group do |group|
  user.email_verified? &&
  (group.is_parent? &&
  group.members.exists?(user.id) &&
  (group.members_can_create_subgroups? || group.admins.exists?(user.id)))
end

can :create, ::Group do |group|
  ...
  ( user_is_admin_of?(group.parent_id) ||
    (user_is_member_of?(group.parent_id) && group.parent.members_can_create_subgroups?) )
end
```
**Grants**: `:add_subgroup`, `:create` (for subgroups) on Group

#### `members_can_announce` - `/Users/z/Code/loomio/app/models/ability/group.rb:61-63`
```ruby
can [:notify], ::Group do |group|
  (group.members_can_announce && group.members.exists?(user.id)) || group.admins.exists?(user.id)
end
```
Also in `/Users/z/Code/loomio/app/models/ability/discussion.rb:29-36`:
```ruby
can [:announce], ::Discussion do |discussion|
  if discussion.group_id
    discussion.group.admins.exists?(user.id) ||
    (discussion.group.members_can_announce && discussion.members.exists?(user.id))
  else
    discussion.admins.exists?(user.id)
  end
end
```
Also in `/Users/z/Code/loomio/app/models/ability/poll.rb:38-46`:
```ruby
can [:announce, :remind], ::Poll do |poll|
  if poll.group_id
    poll.group.admins.exists?(user.id) ||
    (poll.group.members_can_announce && poll.admins.exists?(user.id))
  else
    ...
  end
end
```
**Grants**: `:notify` on Group, `:announce` on Discussion, `:announce`/`:remind` on Poll

#### `members_can_add_guests` - `/Users/z/Code/loomio/app/models/ability/group.rb:45-48`
```ruby
can [:add_guests], ::Group do |group|
  user.email_verified? && Subscription.for(group).is_active? &&
  ((group.members_can_add_guests && group.members.exists?(user.id)) || group.admins.exists?(user.id))
end
```
Also in `/Users/z/Code/loomio/app/models/ability/discussion.rb:42-49`:
```ruby
can [:add_guests], ::Discussion do |discussion|
  if discussion.group_id
    Subscription.for(discussion.group).allow_guests &&
    (discussion.group.admins.exists?(user.id) || (discussion.group.members_can_add_guests && discussion.members.exists?(user.id)))
  else
    !discussion.id || discussion.admins.exists?(user.id)
  end
end
```
Also in `/Users/z/Code/loomio/app/models/ability/poll.rb:52-59`:
```ruby
can [:add_guests], ::Poll do |poll|
  if poll.group_id
    Subscription.for(poll.group).allow_guests &&
    (poll.group.admins.exists?(user.id) || (poll.group.members_can_add_guests && poll.admins.exists?(user.id)))
  else
    poll.admins.exists?(user.id)
  end
end
```
**Grants**: `:add_guests` on Group, Discussion, Poll (also requires subscription check)

#### `admins_can_edit_user_content` - `/Users/z/Code/loomio/app/models/ability/comment.rb:11-16`
```ruby
can [:update], ::Comment do |comment|
  !comment.discussion.closed_at && (
    (comment.discussion.members.exists?(user.id) && comment.author == user && comment.group.members_can_edit_comments) ||
    (comment.discussion.admins.exists?(user.id) && comment.group.admins_can_edit_user_content)
  )
end
```
**Grants**: `:update` on Comment (for admins to edit other users' comments)

#### `parent_members_can_see_discussions` - NOT in ability files
Used only in queries: `/Users/z/Code/loomio/app/queries/discussion_query.rb:42`
```ruby
#{'OR (groups.parent_members_can_see_discussions = TRUE AND groups.parent_id IN (:user_group_ids))' if or_subgroups}
```
**Purpose**: Controls visibility of subgroup discussions to parent group members

---

## 4. Summary Table

| Flag | Default | Tracked | Active | Scope |
|------|---------|---------|--------|-------|
| `members_can_add_members` | false | YES | YES | Group invitations |
| `members_can_edit_discussions` | true | YES | YES | Discussion editing |
| `members_can_edit_comments` | true | YES | YES | Comment editing (author) |
| `members_can_delete_comments` | true | YES | YES | Comment deletion (author) |
| `members_can_raise_motions` | true | YES | YES | Poll creation |
| `members_can_vote` | true | NO | **NO** | DEPRECATED |
| `members_can_start_discussions` | true | YES | YES | Discussion creation |
| `members_can_create_subgroups` | false | YES | YES | Subgroup creation |
| `members_can_announce` | true | YES | YES | Notifications |
| `members_can_add_guests` | true | **NO** | YES | Guest invitations |
| `admins_can_edit_user_content` | true | YES | YES | Admin edit others' content |
| `parent_members_can_see_discussions` | false | YES | YES | Subgroup visibility |

---

## Confidence Levels

| Finding | Confidence | Reasoning |
|---------|------------|-----------|
| Paper trail `only:` list | HIGH | Explicit code, lines 132-158 |
| `members_can_add_guests` excluded from paper_trail | HIGH | Absent from explicit list |
| `members_can_vote` deprecated | HIGH | Commented-out migration + no ability usage |
| Null::Group true_methods wins over false_methods | HIGH | Ruby method redefinition semantics + explicit order in apply_null_methods! |
| All ability mappings | HIGH | Direct code references provided |

---

## Recommendations

1. **Add `members_can_add_guests` to paper_trail** - This appears to be an oversight
2. **Remove `members_can_vote` column** - The migration has been waiting since 2020
3. **Clean up Null::Group contradictions** - Remove duplicates from `false_methods` to prevent confusion

---

*Generated: 2026-02-01*
*Files analyzed: 15*
*Total permission flags: 12 (11 active, 1 deprecated)*
