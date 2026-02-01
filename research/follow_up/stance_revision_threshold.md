# Stance Revision Threshold - Follow-up Investigation Brief

## Discrepancy Summary

Discovery documents a **15-minute vote revision window** that determines whether updating a vote creates a new stance record (preserving history) or updates in place. This threshold is **not documented in our research**.

## Discovery Claims

**Source**: `discovery/initial/polls/services.md`

> "If updating a cast vote in a discussion, and option_scores changed, and more than 15 minutes since last save: Create new stance record (preserves history)"

The service logic flow described:
1. User casts initial vote → creates Stance with `latest: true`
2. User updates vote within 15 minutes → updates existing Stance in place
3. User updates vote after 15 minutes → creates new Stance, marks old as `latest: false`

This implies vote history preservation behavior depends on timing.

## Our Research Claims

**Source**: `research/investigation/models.md`, `research/investigation/database.md`

Our research documents:
- Stance model with `latest: boolean` field
- Partial unique index on `(poll_id, participant_id) WHERE latest = true`
- `option_scores` JSONB for vote values

But does NOT document:
- The 15-minute threshold
- When new stance records are created vs. updated
- Vote history preservation logic

## Ground Truth Needed

1. Does a 15-minute threshold exist?
2. What is the exact threshold value (if configurable)?
3. What logic determines create-new vs. update-existing?
4. Is this behavior documented/tested in the codebase?

## Investigation Targets

- [ ] File: `orig/loomio/app/services/stance_service.rb` - Find vote update logic
- [ ] Command: `grep -n "15\|minute\|threshold" orig/loomio/app/services/stance_service.rb` - Find time-based logic
- [ ] Command: `grep -n "latest" orig/loomio/app/services/stance_service.rb` - Find latest flag handling
- [ ] File: `orig/loomio/app/models/stance.rb` - Check for time-based validations
- [ ] File: `orig/loomio/spec/services/stance_service_spec.rb` - Find tests for revision behavior

## Priority

**MEDIUM** - This affects:
- Vote audit trail accuracy
- Data model understanding for Go rewrite
- Potential edge cases in vote handling

## Rails Context

### Stance Latest Pattern

The `latest` boolean with partial unique index is a common pattern for "current version" tracking:

```ruby
# app/models/stance.rb
class Stance < ApplicationRecord
  belongs_to :poll
  belongs_to :participant, class_name: 'User'

  scope :latest, -> { where(latest: true) }

  # Partial unique index ensures only one "latest" per user per poll
  # CREATE UNIQUE INDEX ... ON stances (poll_id, participant_id) WHERE latest = true
end
```

### Expected Service Logic

```ruby
# app/services/stance_service.rb
class StanceService
  def self.update(stance:, actor:, params:)
    # Check if significant time has passed
    if stance.cast_at && stance.cast_at < 15.minutes.ago
      # Create new stance, preserve history
      create_new_stance(stance, params)
    else
      # Update in place
      stance.update!(params)
    end
  end

  private

  def self.create_new_stance(old_stance, params)
    old_stance.update!(latest: false)
    Stance.create!(
      poll: old_stance.poll,
      participant: old_stance.participant,
      option_scores: params[:option_scores],
      latest: true,
      cast_at: Time.current
    )
  end
end
```

### Vote History Use Cases

Why preserve vote history:
1. **Audit trail**: See how opinions evolved
2. **Transparency**: Show vote changes in timeline
3. **Analytics**: Track engagement patterns

Why allow in-place updates (within window):
1. **Typo correction**: User immediately fixes mistake
2. **UI feedback**: Quick adjustments shouldn't clutter history
3. **Performance**: Fewer records

## Impact on Go Rewrite

For Go implementation:
- Implement the same threshold logic (verify exact value)
- Ensure partial unique index on `(poll_id, participant_id) WHERE latest`
- Consider making threshold configurable via environment variable
- Document the behavior for API consumers

```go
const StanceRevisionThreshold = 15 * time.Minute

func (s *StanceService) Update(stance *Stance, params StanceParams) error {
    if stance.CastAt != nil && time.Since(*stance.CastAt) > StanceRevisionThreshold {
        // Create new stance, mark old as not latest
        return s.createNewStance(stance, params)
    }
    // Update in place
    return s.db.UpdateStance(stance.ID, params)
}
```
