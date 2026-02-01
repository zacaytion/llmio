# Permission Flags - Follow-up Items

## Executive Summary

Comparison of third-party discovery documents against our research reveals several discrepancies requiring resolution. The third-party investigation provides more detailed source-level verification but contains some errors in default values that need clarification.

---

## Discrepancies

### 1. Default Value for `members_can_add_guests`

**Priority: MEDIUM**

| Source | Claimed Default |
|--------|-----------------|
| Our Research (`authorization.md`) | `true` |
| Third-party (`findings.md`) | `true` |
| Actual Schema (`db/schema.rb:474`) | `true` |

**Status:** ALIGNED - Both sources agree and match ground truth.

---

### 2. Default Value for `members_can_delete_comments`

**Priority: HIGH**

| Source | Claimed Default |
|--------|-----------------|
| Our Research (`authorization.md`) | `false` |
| Third-party (`findings.md`) | `true` |
| Actual Schema (`db/schema.rb:475`) | `true` |

**Discrepancy:** Our research claims default is `false`, but both third-party and schema show `true`.

**Action Required:** Verify the schema line and update our research document.

```ruby
# db/schema.rb:475
t.boolean "members_can_delete_comments", default: true, null: false
```

**Questions for Third Party:**
1. Can you confirm this was verified directly from the schema file?

---

### 3. Default Value for `members_can_add_members`

**Priority: MEDIUM**

| Source | Claimed Default |
|--------|-----------------|
| Our Research (`authorization.md`) | `true` |
| Third-party (`findings.md`) | `false` |
| Actual Schema (`db/schema.rb:427`) | `false` |

**Discrepancy:** Our research claims default is `true`, but ground truth shows `false`.

**Action Required:** Update our research - this is a security-sensitive setting where the restrictive default (`false`) is correct.

---

### 4. Missing Flag: `members_can_vote`

**Priority: LOW**

| Source | Documented? |
|--------|-------------|
| Our Research | Not mentioned |
| Third-party | Documented as UNUSED legacy flag |
| Schema | Present at line 440 |

**Discrepancy:** Our research doesn't mention this legacy flag at all.

**Third-party Claim:** Flag exists but is never checked in ability files. Migration to remove it was commented out.

**Questions for Third Party:**
1. Did you verify there are absolutely no ability checks for `members_can_vote`?
2. Is the flag used anywhere outside of `record_cloner.rb`?

**Files to Investigate:**
- `/orig/loomio/db/migrate/20201009024231_remove_groups_members_can_vote.rb` - Why was removal commented out?
- `/orig/loomio/app/services/record_cloner.rb:122` - What is the cloning behavior?

---

### 5. Classification of `new_threads_*` Columns

**Priority: MEDIUM**

| Column | Our Research Classification | Third-party Classification |
|--------|----------------------------|---------------------------|
| `new_threads_max_depth` | Permission flag | Configuration setting |
| `new_threads_newest_first` | Permission flag | Configuration setting |

**Resolution:** Third-party classification is correct. These columns:
- Are NOT checked in any ability file
- Set default values for new discussions
- Are used in `Null::Group` to provide defaults

**Action Required:** Update our research to reclassify these as configuration settings.

---

### 6. Missing Flag: `parent_members_can_see_discussions`

**Priority: HIGH**

| Source | Documented? |
|--------|-------------|
| Our Research (`authorization.md`) | Not in flag table |
| Third-party (`findings.md`) | Documented as visibility/permission flag |
| Schema | Present at line 419 |

**Discrepancy:** Our research doesn't list this flag in the permission flags table.

**Third-party Claim:** This flag is used in `DiscussionQuery` (not in Ability classes) to control visibility of subgroup discussions to parent group members.

**Questions for Third Party:**
1. Is this correctly classified as a "permission" flag even though it's used in a query rather than an ability class?
2. Does this affect authorization (blocking access) or just visibility (filtering)?

**Files to Investigate:**
- `/orig/loomio/app/queries/discussion_query.rb:42` - Exact usage pattern
- `/orig/loomio/app/controllers/api/v1/attachments_controller.rb:7` - Secondary usage

---

### 7. Paper Trail Tracking Discrepancy

**Priority: LOW**

Third-party claims `members_can_add_guests` is NOT tracked in paper_trail, but the schema shows it exists and is an active permission.

**Questions for Third Party:**
1. Can you verify the paper_trail `only:` array in `group.rb` lines 132-158?
2. Is `members_can_add_guests` intentionally excluded from versioning?

**Files to Investigate:**
- `/orig/loomio/app/models/group.rb:132-158` - Verify paper_trail configuration

---

### 8. Null::Group Inconsistency

**Priority: MEDIUM**

Third-party discovered `Null::Group` concern at `/orig/loomio/app/models/concerns/null/group.rb`. This defines default permission values for "null" group contexts (invite-only discussions without a group).

**Observation:** `Null::Group` has BOTH `members_can_add_guests` in `true_methods` (line 71) AND `false_methods` (line 93).

```ruby
def true_methods
  %w[
    # ...
    members_can_add_guests  # line 71
  ]
end

def false_methods
  %w(
    # ...
    members_can_add_guests  # line 93
  )
end
```

**Questions for Third Party:**
1. Is this a bug in the original codebase?
2. Which takes precedence - `true_methods` or `false_methods`?

---

## Contradictions Requiring Resolution

### Contradiction 1: Permission Flag Count

| Source | Count | Includes |
|--------|-------|----------|
| Our Research (initial) | 11 | Incorrectly includes `new_threads_*` |
| Discovery (initial) | 10 | Missing several flags |
| Third-party (final) | 12 | 11 `members_can_*` + 1 `admins_can_*` |
| Ground Truth | 12 | 11 `members_can_*` + 1 `admins_can_*` |

**Resolution:** Accept third-party count of 12 total permission flags.

---

### Contradiction 2: `members_can_edit_comments` Scope

**Our Research Claim:** "Edit any comment"
**Third-party Claim:** "Edit their own comments"

**Files to Investigate:**
- `/orig/loomio/app/models/ability/comment.rb:13`

```ruby
(comment.discussion.members.exists?(user.id) && comment.author == user && comment.group.members_can_edit_comments)
```

**Resolution:** Third-party is correct. The check includes `comment.author == user`, meaning this flag only allows members to edit **their own** comments, not any comment.

---

## Unclear/Incomplete Areas in Third-party Documentation

### 1. Permission Inheritance Logic

**Priority: HIGH**

Third-party documents individual flags but doesn't clarify:
- Do subgroups inherit permission flag values from parent groups?
- What is the cascade behavior when a parent group changes a permission?
- How does `parent_members_can_see_discussions` interact with other visibility controls?

**Investigation Needed:**
- Check for `before_create` or `after_create` callbacks that copy parent permissions
- Look for any inheritance patterns in group creation/update

---

### 2. Guest Access Permission Interaction

**Priority: MEDIUM**

Third-party documents `members_can_add_guests` as controlling guest invitations, but doesn't explain:
- How does this interact with `discussion_readers` table?
- What happens when a guest is added to a poll vs a discussion?
- Is there separate control for poll guests vs discussion guests?

**Files to Investigate:**
- `/orig/loomio/app/models/ability/discussion.rb:45`
- `/orig/loomio/app/models/ability/poll.rb:55`

---

### 3. Admin Override Behavior

**Priority: MEDIUM**

Third-party shows admins can always perform actions regardless of `members_can_*` flags, but doesn't clarify:
- Does `admins_can_edit_user_content` apply only to comments or also to discussions/polls?
- Are there any actions even admins cannot perform?

---

## Specific Files Requiring Investigation

| File | Line(s) | Reason | Priority |
|------|---------|--------|----------|
| `/orig/loomio/db/schema.rb` | 419-482 | Verify all defaults | HIGH |
| `/orig/loomio/app/models/ability/comment.rb` | 13-14, 33 | Verify edit/delete scope | HIGH |
| `/orig/loomio/app/models/group.rb` | 132-158 | Verify paper_trail tracking | MEDIUM |
| `/orig/loomio/app/queries/discussion_query.rb` | 42 | Understand `parent_members_can_see_discussions` usage | HIGH |
| `/orig/loomio/app/models/concerns/null/group.rb` | 71, 93 | Resolve duplicate method definition | MEDIUM |
| `/orig/loomio/db/migrate/20201009024231_remove_groups_members_can_vote.rb` | All | Why was removal commented out? | LOW |

---

## Questions for Third Party

1. **Default Values:** Can you provide a single authoritative table with column name, default value, and schema line number for all 12 permission flags?

2. **Legacy Flags:** Should `members_can_vote` be included for compatibility, or should it be omitted entirely?

3. **Null::Group Bug:** Is the duplicate `members_can_add_guests` in both `true_methods` and `false_methods` a known issue? How is this handled at runtime?

4. **Permission Inheritance:** Do subgroups inherit ANY permission settings from their parent group, or are all permissions independent?

5. **Visibility vs Authorization:** You classify `parent_members_can_see_discussions` as a "visibility query" flag. Does this mean it only filters results but doesn't block API access if someone has the discussion ID directly?

---

## Priority Summary

| Priority | Item |
|----------|------|
| HIGH | Verify `members_can_delete_comments` default (was `false` in our docs) |
| HIGH | Add `parent_members_can_see_discussions` to our permission flag list |
| HIGH | Clarify permission inheritance patterns |
| MEDIUM | Update `members_can_add_members` default to `false` |
| MEDIUM | Reclassify `new_threads_*` as configuration, not permissions |
| MEDIUM | Resolve `Null::Group` duplicate method issue |
| LOW | Document `members_can_vote` legacy flag status |
| LOW | Verify paper_trail tracking for `members_can_add_guests` |
