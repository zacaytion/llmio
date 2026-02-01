# Permission Flags Verification Checklist

## Ground Truth Answers

### Q1: Complete list of `members_can_*` columns in groups table

**Answer: 11 columns**

| # | Column | Verified | Evidence |
|---|--------|----------|----------|
| 1 | `members_can_add_members` | PASS | `db/schema.rb:427` |
| 2 | `members_can_edit_discussions` | PASS | `db/schema.rb:429` |
| 3 | `members_can_edit_comments` | PASS | `db/schema.rb:438` |
| 4 | `members_can_raise_motions` | PASS | `db/schema.rb:439` |
| 5 | `members_can_vote` | PASS | `db/schema.rb:440` |
| 6 | `members_can_start_discussions` | PASS | `db/schema.rb:441` |
| 7 | `members_can_create_subgroups` | PASS | `db/schema.rb:442` |
| 8 | `members_can_announce` | PASS | `db/schema.rb:465` |
| 9 | `members_can_add_guests` | PASS | `db/schema.rb:474` |
| 10 | `members_can_delete_comments` | PASS | `db/schema.rb:475` |
| 11 | `parent_members_can_see_discussions` | PASS | `db/schema.rb:419` |

**Confidence: 5/5** - Direct schema inspection with line numbers.

---

### Q2: Are `new_threads_*` columns permission flags or configuration?

**Answer: CONFIGURATION, not permissions**

| Claim | Status | Evidence |
|-------|--------|----------|
| `new_threads_max_depth` is not in any Ability file | PASS | `grep` found 0 matches in `app/models/ability/` |
| `new_threads_newest_first` is not in any Ability file | PASS | `grep` found 0 matches in `app/models/ability/` |
| They are integer/boolean defaults, not authorization | PASS | Schema shows integer type, used in `Null::Group` for defaults |
| Discussions have their own `max_depth` and `newest_first` | PASS | `db/schema.rb:260-261` (discussions table) |

**Confidence: 5/5** - Exhaustive grep search of ability files confirmed no usage.

---

### Q3: Does `admins_can_edit_user_content` exist?

**Answer: YES**

| Claim | Status | Evidence |
|-------|--------|----------|
| Column exists in schema | PASS | `db/schema.rb:471` |
| Default is true | PASS | `default: true, null: false` |
| Used in Comment ability | PASS | `app/models/ability/comment.rb:14` |
| Allows admins to edit other users' comments | PASS | Code: `comment.discussion.admins.exists?(user.id) && comment.group.admins_can_edit_user_content` |

**Confidence: 5/5** - Found in both schema and ability code with exact line numbers.

---

### Q4: Are there any other permission-related columns?

**Answer: YES, plus configuration columns that were sometimes miscounted**

| Column | Classification | Status | Evidence |
|--------|---------------|--------|----------|
| `admins_can_edit_user_content` | PERMISSION | PASS | Used in ability check |
| `parent_members_can_see_discussions` | PERMISSION (visibility) | PASS | Used in DiscussionQuery |
| `new_threads_max_depth` | CONFIGURATION | PASS | Not in abilities |
| `new_threads_newest_first` | CONFIGURATION | PASS | Not in abilities |
| `can_start_polls_without_discussion` | CONFIGURATION | PASS | Not in abilities |
| `listed_in_explore` | CONFIGURATION | PASS | Not in abilities |

**Confidence: 5/5** - Complete schema analysis with grep verification.

---

## Ability File Cross-Reference

Each permission flag with the ability file(s) that check it:

| Permission Flag | Ability Files | Line Numbers |
|-----------------|---------------|--------------|
| `members_can_start_discussions` | group.rb, discussion.rb | 42, 25 |
| `members_can_edit_discussions` | discussion.rb, outcome.rb | 55, 11 |
| `members_can_edit_comments` | comment.rb | 13 |
| `members_can_delete_comments` | comment.rb | 33 |
| `members_can_raise_motions` | poll.rb | 29-30 |
| `members_can_vote` | **NONE** | N/A (legacy) |
| `members_can_add_members` | group.rb, membership_request.rb | 56, 16 |
| `members_can_add_guests` | group.rb, discussion.rb, poll.rb | 47, 45, 55 |
| `members_can_announce` | group.rb, discussion.rb, poll.rb | 62, 32, 41 |
| `members_can_create_subgroups` | group.rb | 71, 87 |
| `admins_can_edit_user_content` | comment.rb | 14 |
| `parent_members_can_see_discussions` | **QUERY** (not ability) | discussion_query.rb:42 |

---

## Discrepancy Resolution

| Source | Claimed Count | Issues |
|--------|---------------|--------|
| Discovery | 10 `members_can_*` | Likely missed `parent_members_can_see_discussions` or `members_can_add_guests` |
| Research | 11 (with `new_threads_*`) | Incorrectly included configuration as permissions |
| **Actual** | 11 `members_can_*` + 1 `admins_can_*` | Correct count with proper classification |

**Resolution Confidence: 5/5** - All flags traced to source with verification.

---

## Verification Methods Used

1. **Schema grep**: `grep "members_can_\|admins_can_" db/schema.rb`
2. **Ability grep**: `grep -r "members_can_" app/models/ability/`
3. **Direct file reads**: All ability files and schema.rb
4. **Cross-reference**: Each flag verified in both schema and code

---

## Summary

| Question | Answer | Confidence |
|----------|--------|------------|
| Total `members_can_*` columns | 11 | 5/5 |
| `new_threads_*` classification | Configuration | 5/5 |
| `admins_can_edit_user_content` exists | Yes | 5/5 |
| Other permission columns | `parent_members_can_see_discussions` (visibility query) | 5/5 |
| Unused permission flag | `members_can_vote` (legacy) | 5/5 |
