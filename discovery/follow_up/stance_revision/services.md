# StanceService - Vote Update Logic Documentation

## Overview

The `StanceService` class at `/Users/z/Code/loomio/app/services/stance_service.rb` manages all stance (vote) operations. It contains four public methods:

1. `create` - First-time vote creation
2. `update` - Vote modification (revision logic lives here)
3. `uncast` - Withdraw a vote (creates blank replacement)
4. `redeem` - Transfer guest stance to verified user

---

## StanceService.update - The Core Update Logic

### Method Signature

```ruby
def self.update(stance:, actor:, params:)
```

### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `stance` | Stance | The existing stance to update |
| `actor` | User | The user performing the update |
| `params` | Hash | New stance attributes (choices, reason, etc.) |

### Flow Diagram

```
update(stance, actor, params)
    |
    v
authorize!(:update, stance)
    |
    v
is_update = stance.cast_at exists?
    |
    v
build new_stance from replacement template
    |
    v
assign params to new_stance
    |
    +---[if ALL conditions true]---> CREATE NEW RECORD PATH
    |   - is_update (previously cast)
    |   - poll has discussion_id
    |   - option_scores changed
    |   - updated_at > 15 minutes ago
    |           |
    |           v
    |   Transaction:
    |   - old_stance.latest = false
    |   - new_stance.save!
    |           |
    |           v
    |   Events::StanceCreated.publish!
    |
    +---[else]---> UPDATE IN PLACE PATH
                - Clear stance_choices
                - Assign new params
                - Set cast_at if nil
                - Clear revocation flags
                - stance.save!
                        |
                        v
                if is_update:
                  Events::StanceUpdated.publish!
                else:
                  Events::StanceCreated.publish!
```

### Key Decision Points

#### Condition 1: `is_update` (line 29)
```ruby
is_update = !!stance.cast_at
```
- True if the stance has been previously cast (not just invited but unvoted)
- `cast_at` is set when a user submits their vote

#### Condition 2: `stance.poll.discussion_id` (line 37)
- Checks if the poll is attached to a discussion thread
- Standalone polls (not in threads) always update in place
- **This is an important nuance**: The create-new-record behavior only applies to polls within discussions

#### Condition 3: `stance.option_scores != new_stance.build_option_scores` (line 38)
- Compares the vote choices, not just the reason
- If only the reason changed, updates in place regardless of time
- `option_scores` is a JSONB field storing `{poll_option_id: score}` pairs

#### Condition 4: `stance.updated_at < 15.minutes.ago` (line 39)
- The 15-minute threshold
- Uses `updated_at` not `cast_at` - tracks last modification time
- Allows quick corrections without history pollution

---

## Related Methods

### StanceService.create (lines 2-13)

Used for first-time vote creation (not revision):

```ruby
def self.create(stance:, actor:)
  actor.ability.authorize!(:vote_in, stance.poll)

  stance.participant = actor
  stance.cast_at ||= Time.zone.now
  stance.revoked_at = nil
  stance.revoker_id = nil
  stance.save!
  stance.poll.update_counts!

  Events::StanceCreated.publish!(stance)
end
```

### StanceService.uncast (lines 15-25)

Withdraws a vote by creating a blank replacement:

```ruby
def self.uncast(stance:, actor:)
  actor.ability.authorize!(:uncast, stance)

  new_stance = stance.build_replacement
  Stance.transaction do
    stance.update_columns(latest: false)
    new_stance.save!
  end

  new_stance.poll.update_counts!
end
```

Note: Uses `update_columns` to bypass validations and callbacks.

---

## Event Publishing

| Scenario | Event Published |
|----------|-----------------|
| First vote (via create) | `Events::StanceCreated` |
| First vote (via update with no cast_at) | `Events::StanceCreated` |
| Update within 15 mins | `Events::StanceUpdated` |
| Update after 15 mins (with conditions met) | `Events::StanceCreated` (new stance) |
| Uncast vote | No event |

---

## Database Impact

### Create New Record Path (after 15 minutes)

1. Old stance: `UPDATE stances SET latest = false WHERE id = ?`
2. New stance: `INSERT INTO stances ...`
3. Poll counts: `UPDATE polls SET ...`

### Update In Place Path (within 15 minutes)

1. Stance: `UPDATE stances SET ... WHERE id = ?`
2. Poll counts: `UPDATE polls SET ...`

---

## File References

| File | Lines | Purpose |
|------|-------|---------|
| `/Users/z/Code/loomio/app/services/stance_service.rb` | 1-79 | Service class |
| `/Users/z/Code/loomio/app/models/stance.rb` | 120-128 | `build_replacement` method |
| `/Users/z/Code/loomio/app/models/events/stance_created.rb` | 1-40 | StanceCreated event |
| `/Users/z/Code/loomio/app/models/events/stance_updated.rb` | 1-2 | StanceUpdated event (inherits StanceCreated) |
