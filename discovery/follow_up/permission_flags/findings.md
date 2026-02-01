# Permission Flags Investigation: Complete Enumeration

## Investigation Summary

This document provides the definitive enumeration of permission-related columns on the `groups` table, resolving the discrepancy between Discovery (10 flags) and Research (11 flags) documentation.

## Ground Truth: All Permission-Related Columns in Groups Table

Source: `/Users/z/Code/loomio/db/schema.rb` lines 409-492

### True Permission Flags (12 total)

These columns directly control authorization decisions in Ability classes:

| Column | Type | Default | Schema Line | Authorization Use |
|--------|------|---------|-------------|-------------------|
| `members_can_add_members` | boolean | false | 427 | Group, MembershipRequest abilities |
| `members_can_edit_discussions` | boolean | true | 429 | Discussion, Outcome abilities |
| `members_can_edit_comments` | boolean | true | 438 | Comment ability |
| `members_can_raise_motions` | boolean | true | 439 | Poll ability (create polls) |
| `members_can_vote` | boolean | true | 440 | **UNUSED** - exists but not checked |
| `members_can_start_discussions` | boolean | true | 441 | Group, Discussion abilities |
| `members_can_create_subgroups` | boolean | false | 442 | Group ability |
| `members_can_announce` | boolean | true | 465 | Group, Discussion, Poll abilities |
| `members_can_add_guests` | boolean | true | 474 | Group, Discussion, Poll abilities |
| `members_can_delete_comments` | boolean | true | 475 | Comment ability |
| `admins_can_edit_user_content` | boolean | true | 471 | Comment ability |
| `parent_members_can_see_discussions` | boolean | false | 419 | DiscussionQuery (visibility) |

**Count: 11 `members_can_*` flags + 1 `admins_can_*` flag = 12 permission flags**

### Configuration Settings (NOT Permission Flags)

These columns affect default values or behavior but are NOT checked in Ability classes:

| Column | Type | Default | Schema Line | Purpose |
|--------|------|---------|-------------|---------|
| `new_threads_max_depth` | integer | 3 | 469 | Default thread nesting depth |
| `new_threads_newest_first` | boolean | false | 470 | Default thread sort order |
| `can_start_polls_without_discussion` | boolean | false | 482 | Feature toggle |
| `listed_in_explore` | boolean | false | 472 | Visibility in explore page |

## Key Findings

### 1. `members_can_vote` is a Legacy Flag

The column exists (line 440) but is **never used in ability checks**. A migration attempted to remove it but was commented out:

```ruby
# db/migrate/20201009024231_remove_groups_members_can_vote.rb
# remove_column :groups, :members_can_vote
```

This flag is only referenced in:
- `app/services/record_cloner.rb` (line 122) - for cloning groups
- `app/models/group.rb` - in ransackable_attributes (line 434)

### 2. `new_threads_*` Are Configuration, Not Permissions

These columns set **default values** for new discussions:
- `new_threads_max_depth`: Default value for `discussions.max_depth`
- `new_threads_newest_first`: Default value for `discussions.newest_first`

They are used in `Null::Group` to provide defaults when no group is present.

### 3. Discovery vs Research Discrepancy Resolution

**Discovery counted 10** - likely excluding `parent_members_can_see_discussions` and `admins_can_edit_user_content`

**Research counted 11 including `new_threads_*`** - incorrectly classified configuration as permissions

**Actual count:**
- 11 `members_can_*` boolean flags (including one unused)
- 1 `admins_can_*` boolean flag
- 1 visibility flag (`parent_members_can_see_discussions`)
- 2 thread default settings (NOT permissions)

## Complete Schema Extract

From `db/schema.rb` lines 419-482:

```ruby
t.boolean "parent_members_can_see_discussions", default: false, null: false  # line 419
t.boolean "members_can_add_members", default: false, null: false             # line 427
t.boolean "members_can_edit_discussions", default: true, null: false         # line 429
t.boolean "members_can_edit_comments", default: true                         # line 438
t.boolean "members_can_raise_motions", default: true, null: false            # line 439
t.boolean "members_can_vote", default: true, null: false                     # line 440 (UNUSED)
t.boolean "members_can_start_discussions", default: true, null: false        # line 441
t.boolean "members_can_create_subgroups", default: false, null: false        # line 442
t.boolean "members_can_announce", default: true, null: false                 # line 465
t.integer "new_threads_max_depth", default: 3, null: false                   # line 469 (CONFIG)
t.boolean "new_threads_newest_first", default: false, null: false            # line 470 (CONFIG)
t.boolean "admins_can_edit_user_content", default: true, null: false         # line 471
t.boolean "listed_in_explore", default: false, null: false                   # line 472 (CONFIG)
t.boolean "members_can_add_guests", default: true, null: false               # line 474
t.boolean "members_can_delete_comments", default: true, null: false          # line 475
t.boolean "can_start_polls_without_discussion", default: false, null: false  # line 482 (CONFIG)
```
