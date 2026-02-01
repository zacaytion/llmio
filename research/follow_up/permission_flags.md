# Permission Flags - Follow-up Investigation Brief

## Discrepancy Summary

Discovery and Research document **different counts** of group permission flags:
- Discovery: 10 `members_can_*` boolean flags
- Research: 11 flags (includes `new_threads_max_depth`, `new_threads_newest_first`)

The difference may be semantic (what counts as a "permission flag" vs "configuration setting").

## Discovery Claims

**Source**: `discovery/initial/groups/models.md`

Lists 10 permission flags:
1. `members_can_add_members`
2. `members_can_add_guests`
3. `members_can_announce`
4. `members_can_create_subgroups`
5. `members_can_start_discussions`
6. `members_can_edit_discussions`
7. `members_can_edit_comments`
8. `members_can_delete_comments`
9. `members_can_raise_motions`
10. `admins_can_edit_user_content`

Note: Discovery also mentions `admins_can_edit_user_content` which is an admin flag, not a member flag.

## Our Research Claims

**Source**: `research/investigation/authorization.md`

Lists 11 group permission flags:
1. `members_can_add_members`
2. `members_can_add_guests`
3. `members_can_announce`
4. `members_can_create_subgroups`
5. `members_can_start_discussions`
6. `members_can_edit_discussions`
7. `members_can_edit_comments`
8. `members_can_delete_comments`
9. `members_can_raise_motions`
10. `new_threads_max_depth` (integer)
11. `new_threads_newest_first` (boolean)

Note: Research includes thread configuration settings in the "permission flags" list.

## Ground Truth Needed

1. Complete list of `members_can_*` columns in groups table
2. Are `new_threads_*` columns permission flags or configuration?
3. Does `admins_can_edit_user_content` exist?
4. Are there any other permission-related columns?

## Investigation Targets

- [ ] File: `orig/loomio/db/schema.rb` - Search for `t.boolean` in groups table definition
- [ ] File: `orig/loomio/app/models/group.rb` - Check column definitions and validations
- [ ] Command: `grep -E "members_can_|admins_can_" orig/loomio/app/models/group.rb` - Find all permission columns
- [ ] Command: `grep "new_threads" orig/loomio/app/models/group.rb` - Find thread configuration columns
- [ ] File: `orig/loomio/app/models/ability/group.rb` - Check which flags affect authorization

## Priority

**MEDIUM** - Permission flags affect authorization logic but the core 9 `members_can_*` flags are agreed upon. The discrepancy is mostly about classification.

## Rails Context

### Permission Flags Pattern

Rails models often use boolean columns for feature flags:

```ruby
# Group model with permission flags
class Group < ApplicationRecord
  # Permission flags (affect what members can do)
  attribute :members_can_add_members, :boolean, default: true
  attribute :members_can_start_discussions, :boolean, default: true

  # Configuration settings (affect behavior, not permissions)
  attribute :new_threads_max_depth, :integer, default: 3
  attribute :new_threads_newest_first, :boolean, default: false
end
```

### Semantic Distinction

**Permission flags** typically:
- Are checked in Ability classes
- Control whether an action is allowed
- Have `_can_` in the name

**Configuration settings** typically:
- Affect default values or behavior
- Don't directly control authorization
- May affect UI/UX without blocking actions

### CanCanCan Usage

```ruby
# app/models/ability/discussion.rb
def ability_for_discussion
  can :create, Discussion do |discussion|
    discussion.group.members_can_start_discussions || user.is_admin?
  end
end
```

## Reconciliation

The discrepancy is likely a **classification difference**, not a data difference:

| Column | Type | Discovery | Research | Actual Category |
|--------|------|-----------|----------|-----------------|
| `members_can_add_members` | boolean | Permission | Permission | Permission |
| `members_can_add_guests` | boolean | Permission | Permission | Permission |
| `members_can_announce` | boolean | Permission | Permission | Permission |
| `members_can_create_subgroups` | boolean | Permission | Permission | Permission |
| `members_can_start_discussions` | boolean | Permission | Permission | Permission |
| `members_can_edit_discussions` | boolean | Permission | Permission | Permission |
| `members_can_edit_comments` | boolean | Permission | Permission | Permission |
| `members_can_delete_comments` | boolean | Permission | Permission | Permission |
| `members_can_raise_motions` | boolean | Permission | Permission | Permission |
| `admins_can_edit_user_content` | boolean | Permission | Not listed | Permission |
| `new_threads_max_depth` | integer | Not listed | Permission | Configuration |
| `new_threads_newest_first` | boolean | Not listed | Permission | Configuration |

**Action**: Verify the complete list and establish canonical classification for Go implementation.

## Impact on Go Rewrite

For Go implementation:
- Create a `GroupPermissions` struct or embed in `Group` model
- Clearly separate permission flags from configuration settings
- Ensure all 9+ permission flags are included in authorization checks
- Document which flags affect which operations
