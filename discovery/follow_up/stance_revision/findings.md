# Stance Revision Threshold - Findings

## Executive Summary

**CONFIRMED**: A 15-minute threshold exists for stance revision behavior. The threshold is hardcoded in `StanceService.update` at line 39.

---

## Ground Truth Answers

### 1. Does a 15-minute threshold exist?

**YES** - Confirmed at `/Users/z/Code/loomio/app/services/stance_service.rb:39`

```ruby
stance.updated_at < 15.minutes.ago
```

The inline comment at line 40 explicitly documents the intent:
```ruby
# they've changed their position, in a poll in a thread, and it's more than 15 minutes since they last saved it.
```

### 2. What is the exact threshold value (if configurable)?

**15 minutes** - Hardcoded, not configurable.

- No constant or environment variable references this value
- The value `15.minutes` is an inline Rails duration literal
- No configuration mechanism exists to change this threshold

### 3. What logic determines create-new vs. update-existing?

The decision tree in `StanceService.update` (lines 36-65):

**Create NEW stance record when ALL conditions are true:**
1. `is_update` - The stance has been previously cast (`!!stance.cast_at` is true)
2. `stance.poll.discussion_id` - The poll is attached to a discussion thread
3. `stance.option_scores != new_stance.build_option_scores` - The vote choices have changed
4. `stance.updated_at < 15.minutes.ago` - More than 15 minutes since last update

**Update EXISTING stance record when ANY condition is false:**
- First-time vote (not an update)
- Poll is not in a discussion thread
- Vote choices haven't changed (only reason updated)
- Less than 15 minutes since last update

### 4. Is this behavior documented/tested in the codebase?

**Partially documented, NOT tested.**

- **Code comments**: The inline comment explains the intent
- **Unit tests**: No specific tests for the 15-minute threshold behavior exist in:
  - `/Users/z/Code/loomio/spec/services/stance_service_spec.rb` - No update method tests
  - `/Users/z/Code/loomio/spec/controllers/api/v1/stances_controller_spec.rb` - Tests update but not time-based branching
- **E2E tests**: `/Users/z/Code/loomio/vue/tests/e2e/specs/stance.js` - No time-based test scenarios

---

## Discovery Claim Verification

| Claim from Discovery | Verified? | Notes |
|---------------------|-----------|-------|
| User casts initial vote - creates Stance with `latest: true` | PARTIAL | Initial vote goes through `create` method, not `update`. The `build_replacement` sets `latest: true` by default |
| User updates within 15 minutes - updates existing | CORRECT | Falls through to `else` branch, updates in place |
| User updates after 15 minutes - creates new, marks old as `latest: false` | MOSTLY CORRECT | Only if choices changed AND poll is in discussion |

---

## Key Code Evidence

### StanceService.update - Full Logic (lines 27-65)

```ruby
def self.update(stance: , actor: , params: )
  actor.ability.authorize!(:update, stance)
  is_update = !!stance.cast_at

  new_stance = stance.build_replacement
  new_stance.assign_attributes_and_files(params)

  event = Event.where(eventable: stance, discussion_id: stance.poll.discussion_id).order('id desc').first

  if is_update &&
     stance.poll.discussion_id &&
     stance.option_scores != new_stance.build_option_scores &&
     stance.updated_at < 15.minutes.ago
    # they've changed their position, in a poll in a thread, and it's more than 15 minutes since they last saved it.

    new_stance.cast_at = Time.zone.now

    Stance.transaction do
      stance.update_columns(latest: false)
      new_stance.save!
    end

    new_stance.poll.update_counts!
    MessageChannelService.publish_models([stance], group_id: stance.poll.group_id)
    Events::StanceCreated.publish!(new_stance)
  else
    stance.stance_choices = []
    stance.assign_attributes_and_files(params)
    stance.cast_at ||= Time.zone.now
    stance.revoked_at = nil
    stance.revoker_id = nil
    stance.save!
    stance.poll.update_counts!
    if is_update
      Events::StanceUpdated.publish!(stance)
    else
      Events::StanceCreated.publish!(stance)
    end
  end
end
```

### Stance.build_replacement (lines 120-128)

```ruby
def build_replacement
  Stance.new(
    poll_id: poll_id,
    participant_id: participant_id,
    inviter_id: inviter_id,
    reason_format: reason_format,
    latest: true
  )
end
```

---

## Confidence Level

**Confidence: 5/5**

All claims verified through direct code inspection:
- Exact line numbers identified
- No ambiguity in the logic
- Code comments confirm developer intent
