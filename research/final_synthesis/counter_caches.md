# Counter Caches - Implementation Synthesis

## Executive Summary

Loomio uses extensive counter cache columns to optimize query performance. This document provides a complete inventory and implementation guidance for maintaining these counters in Go.

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

## Go Implementation

### Counter Cache Service

```go
package counters

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"
)

type CounterService struct {
    db *pgxpool.Pool
}

func NewCounterService(db *pgxpool.Pool) *CounterService {
    return &CounterService{db: db}
}

// Increment increments a counter column
func (s *CounterService) Increment(ctx context.Context, table string, id int64, column string) error {
    query := fmt.Sprintf(
        "UPDATE %s SET %s = %s + 1 WHERE id = $1",
        table, column, column,
    )
    _, err := s.db.Exec(ctx, query, id)
    return err
}

// Decrement decrements a counter column (min 0)
func (s *CounterService) Decrement(ctx context.Context, table string, id int64, column string) error {
    query := fmt.Sprintf(
        "UPDATE %s SET %s = GREATEST(%s - 1, 0) WHERE id = $1",
        table, column, column,
    )
    _, err := s.db.Exec(ctx, query, id)
    return err
}

// Set sets a counter to a specific value
func (s *CounterService) Set(ctx context.Context, table string, id int64, column string, value int) error {
    query := fmt.Sprintf("UPDATE %s SET %s = $1 WHERE id = $2", table, column)
    _, err := s.db.Exec(ctx, query, value, id)
    return err
}
```

### Group Counter Updates

```go
// UpdateGroupCounters recalculates all group counters
func (s *CounterService) UpdateGroupCounters(ctx context.Context, groupID int64) error {
    _, err := s.db.Exec(ctx, `
        UPDATE groups SET
            memberships_count = (
                SELECT COUNT(*) FROM memberships WHERE group_id = $1 AND revoked_at IS NULL
            ),
            admin_memberships_count = (
                SELECT COUNT(*) FROM memberships WHERE group_id = $1 AND admin = true AND revoked_at IS NULL
            ),
            discussions_count = (
                SELECT COUNT(*) FROM discussions WHERE group_id = $1 AND discarded_at IS NULL
            ),
            open_discussions_count = (
                SELECT COUNT(*) FROM discussions WHERE group_id = $1 AND discarded_at IS NULL AND closed_at IS NULL
            ),
            closed_discussions_count = (
                SELECT COUNT(*) FROM discussions WHERE group_id = $1 AND discarded_at IS NULL AND closed_at IS NOT NULL
            ),
            polls_count = (
                SELECT COUNT(*) FROM polls WHERE group_id = $1 AND discarded_at IS NULL
            ),
            closed_polls_count = (
                SELECT COUNT(*) FROM polls WHERE group_id = $1 AND discarded_at IS NULL AND closed_at IS NOT NULL
            ),
            subgroups_count = (
                SELECT COUNT(*) FROM groups WHERE parent_id = $1
            )
        WHERE id = $1
    `, groupID)
    return err
}
```

### Discussion Counter Updates

```go
func (s *CounterService) UpdateDiscussionCounters(ctx context.Context, discussionID int64) error {
    _, err := s.db.Exec(ctx, `
        UPDATE discussions SET
            items_count = (
                SELECT COUNT(*) FROM events WHERE discussion_id = $1
            ),
            closed_polls_count = (
                SELECT COUNT(*) FROM polls WHERE discussion_id = $1 AND closed_at IS NOT NULL
            ),
            anonymous_polls_count = (
                SELECT COUNT(*) FROM polls WHERE discussion_id = $1 AND anonymous = true
            )
        WHERE id = $1
    `, discussionID)
    return err
}
```

### Poll Counter Updates

```go
func (s *CounterService) UpdatePollCounters(ctx context.Context, pollID int64) error {
    _, err := s.db.Exec(ctx, `
        UPDATE polls SET
            voters_count = (
                SELECT COUNT(DISTINCT participant_id) FROM stances WHERE poll_id = $1 AND latest = true
            ),
            undecided_voters_count = (
                SELECT COUNT(*) FROM poll_unvoted_members WHERE poll_id = $1
            )
        WHERE id = $1
    `, pollID)
    return err
}

// UpdatePollStanceCounts updates JSONB stance/score counts
func (s *CounterService) UpdatePollStanceCounts(ctx context.Context, pollID int64) error {
    _, err := s.db.Exec(ctx, `
        UPDATE polls SET
            stance_counts = (
                SELECT COALESCE(
                    jsonb_agg(jsonb_build_object('option_id', option_id, 'count', cnt)),
                    '[]'::jsonb
                )
                FROM (
                    SELECT po.id as option_id, COUNT(sc.stance_id) as cnt
                    FROM poll_options po
                    LEFT JOIN stance_choices sc ON sc.poll_option_id = po.id
                    LEFT JOIN stances s ON s.id = sc.stance_id AND s.latest = true
                    WHERE po.poll_id = $1
                    GROUP BY po.id
                ) counts
            )
        WHERE id = $1
    `, pollID)
    return err
}
```

---

## Integration with Services

### Membership Service

```go
func (s *MembershipService) Create(ctx context.Context, membership *Membership) error {
    // Create membership
    if err := s.repo.Create(ctx, membership); err != nil {
        return err
    }

    // Update group counters
    s.counters.Increment(ctx, "groups", membership.GroupID, "memberships_count")
    if membership.Admin {
        s.counters.Increment(ctx, "groups", membership.GroupID, "admin_memberships_count")
    }

    // Update user counter
    s.counters.Increment(ctx, "users", membership.UserID, "memberships_count")

    return nil
}

func (s *MembershipService) Delete(ctx context.Context, membership *Membership) error {
    // Delete membership
    if err := s.repo.Delete(ctx, membership.ID); err != nil {
        return err
    }

    // Update group counters
    s.counters.Decrement(ctx, "groups", membership.GroupID, "memberships_count")
    if membership.Admin {
        s.counters.Decrement(ctx, "groups", membership.GroupID, "admin_memberships_count")
    }

    // Update user counter
    s.counters.Decrement(ctx, "users", membership.UserID, "memberships_count")

    return nil
}
```

### Stance Service

```go
func (s *StanceService) Create(ctx context.Context, stance *Stance) error {
    // Create stance
    if err := s.repo.Create(ctx, stance); err != nil {
        return err
    }

    // Update poll counters
    s.counters.Increment(ctx, "polls", stance.PollID, "voters_count")
    s.counters.Decrement(ctx, "polls", stance.PollID, "undecided_voters_count")

    // Update JSONB counts
    s.counters.UpdatePollStanceCounts(ctx, stance.PollID)

    return nil
}
```

---

## Counter Reconciliation Job

```go
// ReconcileCountersJob - Run periodically to fix drift
type ReconcileCountersJob struct {
    GroupID int64 `json:"group_id,omitempty"`
}

func (ReconcileCountersJob) Kind() string { return "reconcile_counters" }

type ReconcileCountersWorker struct {
    river.WorkerDefaults[ReconcileCountersJob]
    counters *CounterService
}

func (w *ReconcileCountersWorker) Work(ctx context.Context, job *river.Job[ReconcileCountersJob]) error {
    if job.Args.GroupID != 0 {
        // Reconcile specific group
        return w.counters.UpdateGroupCounters(ctx, job.Args.GroupID)
    }

    // Reconcile all groups (batch)
    rows, err := w.counters.db.Query(ctx, "SELECT id FROM groups")
    if err != nil {
        return err
    }
    defer rows.Close()

    for rows.Next() {
        var groupID int64
        rows.Scan(&groupID)
        if err := w.counters.UpdateGroupCounters(ctx, groupID); err != nil {
            slog.Error("failed to reconcile group", "group_id", groupID, "error", err)
        }
    }

    return nil
}
```

---

## Testing

```go
func TestMembershipCounters(t *testing.T) {
    ctx := context.Background()

    // Create group
    group := &Group{Name: "Test Group"}
    groupRepo.Create(ctx, group)

    // Initial count should be 0
    group, _ = groupRepo.Find(ctx, group.ID)
    assert.Equal(t, 0, group.MembershipsCount)

    // Add member
    membership := &Membership{GroupID: group.ID, UserID: 1}
    membershipService.Create(ctx, membership)

    // Count should be 1
    group, _ = groupRepo.Find(ctx, group.ID)
    assert.Equal(t, 1, group.MembershipsCount)

    // Remove member
    membershipService.Delete(ctx, membership)

    // Count should be 0
    group, _ = groupRepo.Find(ctx, group.ID)
    assert.Equal(t, 0, group.MembershipsCount)
}
```

---

## Best Practices

1. **Use transactions** - Counter updates should be in same transaction as the operation
2. **Handle race conditions** - Use `UPDATE ... SET x = x + 1` not `SELECT` then `UPDATE`
3. **Periodic reconciliation** - Run reconciliation job daily to fix drift
4. **Log mismatches** - Alert when reconciliation finds discrepancies
5. **JSONB counters** - Recalculate entirely rather than incrementing (simpler, safer)
