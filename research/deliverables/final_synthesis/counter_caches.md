# Counter Caches - Implementation Synthesis

## Executive Summary

Loomio uses extensive counter cache columns to optimize query performance. This document provides a complete inventory for maintaining these counters.

---

## Counter Cache Inventory

### Groups Table (17 counters)

| Column | Purpose | Increment On | Decrement On |
|--------|---------|--------------|--------------|
| `memberships_count` | Total members | Membership created | Membership deleted |
| `admin_memberships_count` | Admin count | Membership created (admin) | Membership deleted (admin) |
| `pending_memberships_count` | Pending invites | Invitation sent | Invitation accepted/rejected |
| `invitations_count` | Total invites sent | Invitation created | Never |
| `discussions_count` | Total discussions | Discussion created | Discussion deleted |
| `open_discussions_count` | Non-closed discussions | Discussion created | Discussion closed/deleted |
| `closed_discussions_count` | Closed discussions | Discussion closed | Discussion reopened/deleted |
| `public_discussions_count` | Public discussions | Discussion made public | Discussion made private/deleted |
| `polls_count` | Total polls | Poll created | Poll deleted |
| `closed_polls_count` | Closed polls | Poll closed | Poll reopened/deleted |
| `closed_motions_count` | Closed proposals (legacy) | Proposal closed | Proposal reopened |
| `proposal_outcomes_count` | Outcomes | Outcome created | Outcome deleted |
| `subgroups_count` | Subgroups | Subgroup created | Subgroup deleted |
| `recent_activity_count` | Recent events | Event created (filtered) | Aged out |
| `discussion_templates_count` | Discussion templates | Template created | Template deleted |
| `poll_templates_count` | Poll templates | Template created | Template deleted |
| `delegates_count` | Delegates | Membership.delegate = true | Membership.delegate = false |

### Discussions Table (7 counters)

| Column | Purpose | Increment On | Decrement On |
|--------|---------|--------------|--------------|
| `items_count` | Comments + events | Item created | Item deleted |
| `versions_count` | Edit history | Version created | Never |
| `closed_polls_count` | Closed polls in thread | Poll closed | Poll reopened |
| `anonymous_polls_count` | Anonymous polls | Anon poll created | Anon poll deleted |
| `seen_by_count` | Unique readers | Discussion read | Never |
| `members_count` | Participants | New participant | Never (optional: recalc) |

### Comments Table (3 counters)

| Column | Purpose | Increment On | Decrement On |
|--------|---------|--------------|--------------|
| `comment_votes_count` | Reactions | Reaction created | Reaction deleted |
| `attachments_count` | Files attached | Attachment added | Attachment removed |
| `versions_count` | Edit history | Version created | Never |

### Polls Table (6 counters)

| Column | Purpose | Increment On | Decrement On |
|--------|---------|--------------|--------------|
| `voters_count` | Total voters | Stance created | Stance deleted |
| `undecided_voters_count` | Not yet voted | Voter added | Stance created |
| `versions_count` | Edit history | Version created | Never |

### Poll Options Table (1 counter)

| Column | Purpose | Increment On | Decrement On |
|--------|---------|--------------|--------------|
| `none_of_the_above_count` | NOTA votes | NOTA stance | Stance changed |

### Events Table (2 counters)

| Column | Purpose | Increment On | Decrement On |
|--------|---------|--------------|--------------|
| `child_count` | Direct children | Child event created | Child deleted |
| `descendant_count` | All descendants | Descendant created | Descendant deleted |

### Stances Table (1 counter)

| Column | Purpose | Increment On | Decrement On |
|--------|---------|--------------|--------------|
| `versions_count` | Edit history | Version created | Never |

### Outcomes Table (1 counter)

| Column | Purpose | Increment On | Decrement On |
|--------|---------|--------------|--------------|
| `versions_count` | Edit history | Version created | Never |

### Tags Table (1 counter)

| Column | Purpose | Increment On | Decrement On |
|--------|---------|--------------|--------------|
| `taggings_count` | Tagged records | Tagging created | Tagging deleted |

### Users Table (1 counter)

| Column | Purpose | Increment On | Decrement On |
|--------|---------|--------------|--------------|
| `memberships_count` | Groups joined | Membership created | Membership deleted |

---

## JSONB Counters

### Polls Table

| Column | Type | Purpose | Structure |
|--------|------|---------|-----------|
| `stance_counts` | `jsonb` | Votes per option | `[{"option_id": 1, "count": 5}, ...]` |
| `matrix_counts` | `jsonb` | Score matrix | `[{"option_id": 1, "scores": {...}}, ...]` |
| `score_counts` | `jsonb` | Score distribution | `{"1": 5, "2": 3, ...}` |

---

## Best Practices

1. **Use transactions** - Counter updates should be in same transaction as the operation
2. **Handle race conditions** - Use `UPDATE ... SET x = x + 1` not `SELECT` then `UPDATE`
3. **Periodic reconciliation** - Run reconciliation job daily to fix drift
4. **Log mismatches** - Alert when reconciliation finds discrepancies
5. **JSONB counters** - Recalculate entirely rather than incrementing (simpler, safer)
