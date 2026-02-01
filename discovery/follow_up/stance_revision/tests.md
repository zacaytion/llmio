# Stance Revision - Test Coverage Analysis

## Summary

**Test Coverage: INCOMPLETE**

The 15-minute threshold behavior is **NOT directly tested**. Existing tests cover basic CRUD operations but do not validate the time-based branching logic.

---

## Existing Test Files

### 1. `/Users/z/Code/loomio/spec/services/stance_service_spec.rb`

**Coverage**: `create` and `redeem` methods only.

| Method | Tested | Notes |
|--------|--------|-------|
| `create` | YES | Basic creation, authorization, poll score update |
| `update` | NO | No tests exist |
| `uncast` | NO | No unit tests (tested via controller) |
| `redeem` | YES | Guest stance redemption |

**Key Gap**: No `describe 'update'` block exists.

### 2. `/Users/z/Code/loomio/spec/controllers/api/v1/stances_controller_spec.rb`

**Coverage**: Controller actions, but not time-based service logic.

| Test | Relevant Lines | What It Tests |
|------|----------------|---------------|
| `'updates existing stances'` | 364-370 | Verifies update doesn't create new record - but doesn't manipulate time |
| `'uncast' tests` | 155-212 | Uncast functionality via controller |
| Various creation tests | 214-385 | Authorization, validation |

**Key observation from line 367**:
```ruby
expect { post :update, params: {id: old_stance.id, stance: stance_params } }.to change { Stance.count }.by(0)
```
This tests that update doesn't create new records, but:
- Doesn't test the 15-minute boundary
- Doesn't test polls in discussions vs standalone
- Doesn't test option_scores change detection

### 3. `/Users/z/Code/loomio/spec/models/stance_spec.rb`

**Coverage**: Model validations only.

- Stance choice validation
- Reason length validation
- Choice shorthand syntax

**No coverage** of `build_replacement` or `latest` flag behavior.

### 4. `/Users/z/Code/loomio/vue/tests/e2e/specs/stance.js`

**Coverage**: Happy path voting flows only.

- Guest invitation to vote
- Member invitation to vote
- No vote revision scenarios

---

## Missing Test Scenarios

### Critical Missing Tests

1. **15-minute threshold boundary**
   ```ruby
   # Should update in place when updated_at is recent
   stance.update_columns(updated_at: 10.minutes.ago)
   StanceService.update(stance: stance, ...)
   expect(Stance.count).to eq(initial_count)

   # Should create new record when updated_at is old
   stance.update_columns(updated_at: 20.minutes.ago)
   StanceService.update(stance: stance, ...)
   expect(Stance.count).to eq(initial_count + 1)
   ```

2. **Poll in discussion vs standalone**
   ```ruby
   # Standalone poll - always updates in place
   poll.update(discussion_id: nil)
   stance.update_columns(updated_at: 1.hour.ago)
   StanceService.update(stance: stance, ...)
   expect(stance.reload.latest).to eq(true)
   ```

3. **Option scores unchanged**
   ```ruby
   # Only reason changed - should update in place
   stance.update_columns(updated_at: 1.hour.ago)
   StanceService.update(stance: stance, params: {reason: "new reason"})
   expect(Stance.count).to eq(initial_count)
   ```

4. **Latest flag management**
   ```ruby
   # After creating new stance, old should have latest: false
   old_stance = stance
   StanceService.update(stance: stance, ...)
   expect(old_stance.reload.latest).to eq(false)
   expect(Stance.latest.find_by(participant: user)).to_not eq(old_stance)
   ```

5. **Event type verification**
   ```ruby
   # Within 15 minutes - StanceUpdated event
   expect(Events::StanceUpdated).to receive(:publish!)
   StanceService.update(stance: recent_stance, ...)

   # After 15 minutes - StanceCreated event
   expect(Events::StanceCreated).to receive(:publish!)
   StanceService.update(stance: old_stance, ...)
   ```

---

## Recommended Test Additions

### For `/spec/services/stance_service_spec.rb`

```ruby
describe 'update' do
  let(:poll_in_discussion) { create :poll, discussion: discussion }
  let(:standalone_poll) { create :poll, discussion: nil }
  let(:stance) { create :stance, poll: poll_in_discussion, participant: user, cast_at: 1.day.ago }

  context 'when updated within 15 minutes' do
    before { stance.update_columns(updated_at: 5.minutes.ago) }

    it 'updates the existing stance in place' do
      expect {
        StanceService.update(stance: stance, actor: user, params: new_params)
      }.not_to change { Stance.count }
    end

    it 'publishes StanceUpdated event' do
      expect(Events::StanceUpdated).to receive(:publish!).with(stance)
      StanceService.update(stance: stance, actor: user, params: new_params)
    end
  end

  context 'when updated after 15 minutes' do
    before { stance.update_columns(updated_at: 20.minutes.ago) }

    context 'with changed option_scores' do
      it 'creates a new stance record' do
        expect {
          StanceService.update(stance: stance, actor: user, params: different_choice_params)
        }.to change { Stance.count }.by(1)
      end

      it 'marks old stance as not latest' do
        StanceService.update(stance: stance, actor: user, params: different_choice_params)
        expect(stance.reload.latest).to eq(false)
      end

      it 'publishes StanceCreated event for new stance' do
        expect(Events::StanceCreated).to receive(:publish!)
        StanceService.update(stance: stance, actor: user, params: different_choice_params)
      end
    end

    context 'with unchanged option_scores' do
      it 'updates in place when only reason changed' do
        expect {
          StanceService.update(stance: stance, actor: user, params: same_choice_new_reason)
        }.not_to change { Stance.count }
      end
    end
  end

  context 'for standalone polls (no discussion)' do
    let(:stance) { create :stance, poll: standalone_poll, participant: user }

    it 'always updates in place regardless of time' do
      stance.update_columns(updated_at: 1.hour.ago)
      expect {
        StanceService.update(stance: stance, actor: user, params: new_params)
      }.not_to change { Stance.count }
    end
  end
end
```

---

## Test Risk Assessment

| Risk Level | Description |
|------------|-------------|
| **HIGH** | 15-minute threshold could regress undetected |
| **MEDIUM** | `latest` flag logic could break silently |
| **MEDIUM** | Event type (Created vs Updated) could be wrong |
| **LOW** | Basic CRUD is well-tested |

---

## File References

| File | Status | Gap |
|------|--------|-----|
| `/Users/z/Code/loomio/spec/services/stance_service_spec.rb` | Incomplete | Missing `update` tests |
| `/Users/z/Code/loomio/spec/controllers/api/v1/stances_controller_spec.rb` | Partial | Time-based logic not tested |
| `/Users/z/Code/loomio/spec/models/stance_spec.rb` | Partial | No `build_replacement` tests |
| `/Users/z/Code/loomio/vue/tests/e2e/specs/stance.js` | Minimal | No revision scenarios |
