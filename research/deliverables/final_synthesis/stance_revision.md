# Stance Revision - Implementation Synthesis

## Summary

This document consolidates confirmed findings from both discovery and research for implementing stance revision logic. All claims have been verified against source code at `orig/loomio/app/services/stance_service.rb` and `orig/loomio/app/models/stance.rb`.

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

### Implementation Note

Both mechanisms are needed:

1. **Stance `latest` pattern**: Implement in application code with proper transaction handling
2. **Version history**: Either:
   - Use database triggers (PostgreSQL)
   - Use an auditing library
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

## Test Scenarios

### Unit Tests Required

1. **first cast via update publishes StanceCreated**
   - stance.CastAt is nil
   - Expect StanceCreated event

2. **update within 15 minutes updates in place**
   - stance.UpdatedAt = 5 minutes ago
   - Expect same stance.ID, StanceUpdated event

3. **update after 15 minutes with changed scores creates new**
   - stance.UpdatedAt = 20 minutes ago
   - poll.DiscussionID != nil
   - Different option_scores
   - Expect new stance.ID, old.Latest=false, StanceCreated event

4. **standalone poll always updates in place**
   - poll.DiscussionID = nil
   - stance.UpdatedAt = 1 hour ago
   - Expect update in place regardless of time

5. **reason-only change updates in place**
   - Same option_scores, different reason
   - stance.UpdatedAt = 1 hour ago
   - Expect update in place

6. **transaction rollback on save failure**
   - Simulate new_stance.Save() failure
   - Expect old stance still has latest=true

### Integration Tests Required

1. **full revision flow creates correct timeline**
   - User casts vote -> stance1 (latest=true)
   - User updates within 15 min -> stance1 modified (latest=true)
   - Wait 16 minutes (mock time)
   - User updates again -> stance2 (latest=true), stance1 (latest=false)
   - Verify: ListStanceHistoryByPollAndParticipant returns [stance2, stance1]

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
