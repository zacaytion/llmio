# Stance Revision - Implementation Synthesis

## Summary

This document consolidates confirmed findings from both discovery and research for implementing stance revision logic in the Go rewrite. All claims have been verified against source code at `orig/loomio/app/services/stance_service.rb` and `orig/loomio/app/models/stance.rb`.

---

## Confirmed Business Rules

### Core Revision Logic

When a user updates their vote (stance), the system decides between two paths:

1. **Create New Record Path**: Preserves history by creating a new stance and marking the old one as `latest: false`
2. **Update In Place Path**: Modifies the existing stance record directly

### Decision Criteria (ALL must be true for Create New Record)

```
CREATE NEW RECORD when:
  1. is_update         = stance has been previously cast (cast_at is not nil)
  AND
  2. poll_in_discussion = poll.discussion_id is present
  AND
  3. scores_changed    = option_scores differ from new values
  AND
  4. time_exceeded     = updated_at was more than 15 minutes ago
```

If ANY condition is false, the system updates in place.

### Truth Table

| is_update | poll_in_discussion | scores_changed | time_exceeded | Result |
|-----------|-------------------|----------------|---------------|--------|
| false | - | - | - | Update in place (first cast) |
| true | false | - | - | Update in place (standalone poll) |
| true | true | false | - | Update in place (reason-only change) |
| true | true | true | false | Update in place (within 15 min) |
| true | true | true | true | **Create new record** |

---

## Version Tracking Mechanisms

### Mechanism 1: Stance `latest` Flag Pattern

**Purpose**: Track "current" vote while preserving historical votes as separate rows.

**Database schema**:
```sql
-- Stances table
CREATE TABLE stances (
  id BIGSERIAL PRIMARY KEY,
  poll_id BIGINT NOT NULL,
  participant_id BIGINT NOT NULL,
  option_scores JSONB DEFAULT '{}',
  latest BOOLEAN DEFAULT true,
  cast_at TIMESTAMP,
  updated_at TIMESTAMP,
  -- ... other fields
);

-- Partial unique index ensures one "latest" per user per poll
CREATE UNIQUE INDEX idx_stances_latest_unique
  ON stances (poll_id, participant_id)
  WHERE latest = true;
```

**Behavior**:
- New stance always has `latest: true`
- When creating replacement, old stance gets `latest: false` via `update_columns`
- The partial unique index prevents duplicate "latest" stances

### Mechanism 2: paper_trail Versioning

**Purpose**: Track field-level changes within a single stance record.

**Configuration** (from `stance.rb:67`):
```ruby
has_paper_trail only: [:reason, :option_scores, :revoked_at, :revoker_id, :inviter_id, :attachments]
```

**Tracked fields**:
- `reason` - Vote explanation text
- `option_scores` - JSONB vote values
- `revoked_at` - Revocation timestamp
- `revoker_id` - Who revoked the stance
- `inviter_id` - Who invited the voter
- `attachments` - File references

**Counter cache** (from `stance.rb:114-156`):
```ruby
after_save :update_versions_count!

def update_versions_count!
  update_columns(versions_count: versions.count)
end
```

### Implementation Note for Go

The Go rewrite needs BOTH mechanisms:

1. **Stance `latest` pattern**: Implement in application code with proper transaction handling
2. **Version history**: Either:
   - Use database triggers (PostgreSQL)
   - Use a Go auditing library
   - Build custom version table with JSONB diff storage

---

## Event Publishing Logic

### Event Types

| Scenario | Event Published |
|----------|-----------------|
| First vote via `create` | `Events::StanceCreated` |
| First vote via `update` (cast_at was nil) | `Events::StanceCreated` |
| Update in place | `Events::StanceUpdated` |
| Create new record (after 15 min) | `Events::StanceCreated` (for new stance) |
| Uncast vote | No event published |

### Go Implementation

```go
type StanceEventKind string

const (
    EventStanceCreated StanceEventKind = "stance_created"
    EventStanceUpdated StanceEventKind = "stance_updated"
)

func (s *StanceService) determineEvent(isUpdate bool, createdNewRecord bool) StanceEventKind {
    if !isUpdate || createdNewRecord {
        return EventStanceCreated
    }
    return EventStanceUpdated
}
```

---

## Implementation-Ready Go Code

### Constants and Types

```go
package stance

import (
    "time"
)

// StanceRevisionThreshold is the time window during which vote updates
// modify the existing record rather than creating a new one.
// After this duration, updates to a vote's option_scores create a new
// stance record to preserve voting history in discussion timelines.
const StanceRevisionThreshold = 15 * time.Minute

// Stance represents a user's vote in a poll
type Stance struct {
    ID            int64              `json:"id" db:"id"`
    PollID        int64              `json:"poll_id" db:"poll_id"`
    ParticipantID int64              `json:"participant_id" db:"participant_id"`
    InviterID     *int64             `json:"inviter_id" db:"inviter_id"`
    OptionScores  map[string]int     `json:"option_scores" db:"option_scores"`
    Reason        *string            `json:"reason" db:"reason"`
    ReasonFormat  string             `json:"reason_format" db:"reason_format"`
    Latest        bool               `json:"latest" db:"latest"`
    CastAt        *time.Time         `json:"cast_at" db:"cast_at"`
    RevokedAt     *time.Time         `json:"revoked_at" db:"revoked_at"`
    RevokerID     *int64             `json:"revoker_id" db:"revoker_id"`
    VersionsCount int                `json:"versions_count" db:"versions_count"`
    CreatedAt     time.Time          `json:"created_at" db:"created_at"`
    UpdatedAt     time.Time          `json:"updated_at" db:"updated_at"`
}
```

### Service Implementation

```go
package stance

import (
    "context"
    "time"
)

type UpdateParams struct {
    StanceChoices []StanceChoice `json:"stance_choices"`
    Reason        *string        `json:"reason"`
    ReasonFormat  string         `json:"reason_format"`
    Attachments   []string       `json:"attachments"`
}

// Update handles vote modification with revision threshold logic.
// When all four conditions are met, creates a new stance record
// and marks the old one as latest=false to preserve history.
func (s *StanceService) Update(ctx context.Context, stance *Stance, params UpdateParams) (*Stance, error) {
    // Check authorization
    if err := s.authorize(ctx, "update", stance); err != nil {
        return nil, err
    }

    isUpdate := stance.CastAt != nil
    newOptionScores := s.buildOptionScores(params.StanceChoices)
    poll, err := s.pollRepo.FindByID(ctx, stance.PollID)
    if err != nil {
        return nil, err
    }

    // Decision: Create new record or update in place?
    shouldCreateNew := s.shouldCreateNewStance(isUpdate, poll, stance, newOptionScores)

    if shouldCreateNew {
        return s.createReplacementStance(ctx, stance, params, newOptionScores, poll)
    }

    return s.updateStanceInPlace(ctx, stance, params, isUpdate)
}

// shouldCreateNewStance evaluates the four conditions for creating
// a new stance record instead of updating in place.
func (s *StanceService) shouldCreateNewStance(
    isUpdate bool,
    poll *Poll,
    stance *Stance,
    newOptionScores map[string]int,
) bool {
    // Condition 1: Must be an update to a previously cast vote
    if !isUpdate {
        return false
    }

    // Condition 2: Poll must be in a discussion (has discussion_id)
    if poll.DiscussionID == nil {
        return false
    }

    // Condition 3: Option scores must have changed
    if s.optionScoresEqual(stance.OptionScores, newOptionScores) {
        return false
    }

    // Condition 4: More than 15 minutes since last update
    if time.Since(stance.UpdatedAt) <= StanceRevisionThreshold {
        return false
    }

    return true
}

// createReplacementStance creates a new stance record and marks
// the old one as latest=false within a transaction.
func (s *StanceService) createReplacementStance(
    ctx context.Context,
    oldStance *Stance,
    params UpdateParams,
    newOptionScores map[string]int,
    poll *Poll,
) (*Stance, error) {
    now := time.Now()

    newStance := &Stance{
        PollID:        oldStance.PollID,
        ParticipantID: oldStance.ParticipantID,
        InviterID:     oldStance.InviterID,
        ReasonFormat:  oldStance.ReasonFormat,
        Latest:        true,
        CastAt:        &now,
        OptionScores:  newOptionScores,
        Reason:        params.Reason,
    }

    // Apply params to new stance
    if params.Reason != nil {
        newStance.Reason = params.Reason
    }
    if params.ReasonFormat != "" {
        newStance.ReasonFormat = params.ReasonFormat
    }

    // Execute in transaction
    err := s.db.WithTx(ctx, func(tx *sqlx.Tx) error {
        // Mark old stance as not latest (bypass validations)
        if err := s.stanceRepo.UpdateLatest(ctx, tx, oldStance.ID, false); err != nil {
            return err
        }

        // Save new stance
        if err := s.stanceRepo.Create(ctx, tx, newStance); err != nil {
            return err
        }

        return nil
    })

    if err != nil {
        return nil, err
    }

    // Update poll counts
    if err := s.pollRepo.UpdateCounts(ctx, poll.ID); err != nil {
        return nil, err
    }

    // Publish channel update for old stance (WebSocket notification)
    s.messageChannel.PublishModels(ctx, []*Stance{oldStance}, poll.GroupID)

    // Publish event for new stance
    s.events.Publish(ctx, EventStanceCreated, newStance)

    return newStance, nil
}

// updateStanceInPlace modifies the existing stance record directly.
func (s *StanceService) updateStanceInPlace(
    ctx context.Context,
    stance *Stance,
    params UpdateParams,
    isUpdate bool,
) (*Stance, error) {
    now := time.Now()

    // Clear existing choices before applying new ones
    stance.StanceChoices = nil

    // Apply params
    stance.OptionScores = s.buildOptionScores(params.StanceChoices)
    if params.Reason != nil {
        stance.Reason = params.Reason
    }
    if params.ReasonFormat != "" {
        stance.ReasonFormat = params.ReasonFormat
    }

    // Set cast_at if first cast
    if stance.CastAt == nil {
        stance.CastAt = &now
    }

    // Clear revocation
    stance.RevokedAt = nil
    stance.RevokerID = nil

    if err := s.stanceRepo.Update(ctx, stance); err != nil {
        return nil, err
    }

    // Update poll counts
    poll, _ := s.pollRepo.FindByID(ctx, stance.PollID)
    if poll != nil {
        s.pollRepo.UpdateCounts(ctx, poll.ID)
    }

    // Publish appropriate event
    if isUpdate {
        s.events.Publish(ctx, EventStanceUpdated, stance)
    } else {
        s.events.Publish(ctx, EventStanceCreated, stance)
    }

    return stance, nil
}

// optionScoresEqual compares two option score maps for equality.
func (s *StanceService) optionScoresEqual(a, b map[string]int) bool {
    if len(a) != len(b) {
        return false
    }
    for k, v := range a {
        if bv, ok := b[k]; !ok || bv != v {
            return false
        }
    }
    return true
}

// buildOptionScores converts stance choices to the option_scores JSONB format.
func (s *StanceService) buildOptionScores(choices []StanceChoice) map[string]int {
    scores := make(map[string]int)
    for _, choice := range choices {
        scores[fmt.Sprintf("%d", choice.PollOptionID)] = choice.Score
    }
    return scores
}
```

### SQL Queries (sqlc)

```sql
-- name: UpdateStanceLatest :exec
-- Bypass-style update for marking stance as not latest
UPDATE stances
SET latest = $2, updated_at = now()
WHERE id = $1;

-- name: CreateStance :one
INSERT INTO stances (
    poll_id, participant_id, inviter_id, option_scores,
    reason, reason_format, latest, cast_at,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, now(), now()
)
RETURNING *;

-- name: GetLatestStanceByPollAndParticipant :one
SELECT * FROM stances
WHERE poll_id = $1
  AND participant_id = $2
  AND latest = true
LIMIT 1;

-- name: ListStanceHistoryByPollAndParticipant :many
SELECT * FROM stances
WHERE poll_id = $1
  AND participant_id = $2
ORDER BY created_at DESC;
```

---

## Database Considerations

### Partial Unique Index

```sql
CREATE UNIQUE INDEX idx_stances_poll_participant_latest
  ON stances (poll_id, participant_id)
  WHERE latest = true;
```

This ensures:
- Only one "latest" stance per (poll, participant) combination
- Historical stances (latest=false) don't conflict
- Database-level enforcement of business rule

### Version History Table (if implementing paper_trail equivalent)

```sql
CREATE TABLE stance_versions (
    id BIGSERIAL PRIMARY KEY,
    stance_id BIGINT NOT NULL REFERENCES stances(id),
    object_changes JSONB NOT NULL,  -- Field-level diffs
    whodunnit BIGINT,               -- User who made change
    created_at TIMESTAMP DEFAULT now()
);

CREATE INDEX idx_stance_versions_stance_id ON stance_versions(stance_id);
```

---

## Test Scenarios for Go Implementation

### Unit Tests Required

```go
func TestStanceService_Update(t *testing.T) {
    t.Run("first cast via update publishes StanceCreated", func(t *testing.T) {
        // stance.CastAt is nil
        // Expect StanceCreated event
    })

    t.Run("update within 15 minutes updates in place", func(t *testing.T) {
        // stance.UpdatedAt = 5 minutes ago
        // Expect same stance.ID, StanceUpdated event
    })

    t.Run("update after 15 minutes with changed scores creates new", func(t *testing.T) {
        // stance.UpdatedAt = 20 minutes ago
        // poll.DiscussionID != nil
        // Different option_scores
        // Expect new stance.ID, old.Latest=false, StanceCreated event
    })

    t.Run("standalone poll always updates in place", func(t *testing.T) {
        // poll.DiscussionID = nil
        // stance.UpdatedAt = 1 hour ago
        // Expect update in place regardless of time
    })

    t.Run("reason-only change updates in place", func(t *testing.T) {
        // Same option_scores, different reason
        // stance.UpdatedAt = 1 hour ago
        // Expect update in place
    })

    t.Run("transaction rollback on save failure", func(t *testing.T) {
        // Simulate new_stance.Save() failure
        // Expect old stance still has latest=true
    })
}
```

### Integration Tests Required

```go
func TestStanceRevision_Integration(t *testing.T) {
    t.Run("full revision flow creates correct timeline", func(t *testing.T) {
        // 1. User casts vote -> stance1 (latest=true)
        // 2. User updates within 15 min -> stance1 modified (latest=true)
        // 3. Wait 16 minutes (mock time)
        // 4. User updates again -> stance2 (latest=true), stance1 (latest=false)
        // Verify: ListStanceHistoryByPollAndParticipant returns [stance2, stance1]
    })
}
```

---

## Edge Cases to Handle

1. **Race condition**: Two updates submitted simultaneously after 15-minute threshold
   - The partial unique index prevents duplicate latest stances
   - Second transaction will fail; handle gracefully

2. **Revoked stance**: User tries to update a revoked stance
   - Check `revoked_at` is nil before proceeding
   - Return appropriate error if revoked

3. **Anonymous poll close**: When poll closes and becomes anonymous
   - Existing version history should be preserved
   - participant_id scrubbing happens at poll close, not stance update

4. **Guest stance redemption**: Guest votes, then logs in
   - Redemption is separate from update
   - 15-minute threshold doesn't apply to redemption
