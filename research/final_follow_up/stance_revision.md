# Stance Revision - Follow-Up Analysis

## Executive Summary

Third-party discovery documents provide **substantial new detail** that our baseline research lacked. The discovery correctly identifies the 15-minute threshold but our investigation uncovered additional nuances the discovery missed. Overall, the discovery is **ACCURATE but INCOMPLETE** on edge cases.

---

## Discrepancy Analysis

### 1. Threshold Trigger Condition

| Aspect | Discovery Claim | Our Research | Ground Truth | Status |
|--------|-----------------|--------------|--------------|--------|
| Threshold exists | Yes, 15 minutes | Not documented | Confirmed at `stance_service.rb:39` | **DISCREPANCY** - Our research missed this |
| Threshold value | 15 minutes | Unknown | `15.minutes` hardcoded | **CONFIRMED** |
| Timestamp used | `updated_at` | Not documented | Uses `stance.updated_at` | **DISCOVERY CORRECT** |

**Our research gap**: `research/investigation/models.md` documents the `latest` boolean pattern but not the time-based revision logic.

### 2. Revision Trigger Conditions

| Condition | Discovery | Ground Truth | Status |
|-----------|-----------|--------------|--------|
| Must be a prior vote | Documented (`is_update`) | `!!stance.cast_at` at line 29 | CORRECT |
| Poll in discussion | **NOT DOCUMENTED** | `stance.poll.discussion_id` check at line 37 | **MISSING** |
| Option scores changed | **NOT DOCUMENTED** | `stance.option_scores != new_stance.build_option_scores` at line 38 | **MISSING** |
| Time threshold | Documented (15 min) | `stance.updated_at < 15.minutes.ago` at line 39 | CORRECT |

**Priority: HIGH** - Discovery missed that standalone polls (no `discussion_id`) ALWAYS update in place regardless of time.

### 3. Event Publishing Logic

| Scenario | Discovery Claim | Ground Truth | Status |
|----------|-----------------|--------------|--------|
| Create new record | `Events::StanceCreated` | Line 51 confirms | CORRECT |
| Update in place (is_update) | `Events::StanceUpdated` | Line 61 confirms | CORRECT |
| First cast via update | `Events::StanceCreated` | Line 63 confirms | CORRECT |

**No discrepancy** - Event logic fully verified.

### 4. Version History Mechanism

| Aspect | Discovery | Our Research | Ground Truth | Status |
|--------|-----------|--------------|--------------|--------|
| paper_trail gem | Not mentioned | Not documented | `has_paper_trail` at `stance.rb:67` | **BOTH MISSED** |
| Tracked fields | Not mentioned | Not documented | `[:reason, :option_scores, :revoked_at, :revoker_id, :inviter_id, :attachments]` | **BOTH MISSED** |
| versions_count | Not mentioned | Not documented | Counter cache at `stance.rb:114-156` | **BOTH MISSED** |

**Priority: MEDIUM** - The paper_trail versioning is a SEPARATE mechanism from the `latest` flag pattern. Discovery conflates these two systems.

---

## Contradictions Requiring Resolution

### Contradiction 1: What Constitutes "History Preservation"

**Discovery implies**: Creating a new Stance record with `latest: false` on old = history preservation

**Ground truth reveals**: There are TWO version tracking mechanisms:
1. **Stance `latest` flag**: Creates new DB row, marks old as `latest: false`
2. **paper_trail versions**: Tracks field-level changes in `versions` table

**Question for third party**: Which mechanism did you intend to document? The discovery focuses on the Stance record creation pattern but doesn't mention paper_trail at all.

### Contradiction 2: "Option Scores Changed" Requirement

**Discovery states** (from `services.md` line 89):
> "Compares the vote choices, not just the reason"

**BUT** also states (line 45-46):
> "If only the reason changed, updates in place regardless of time"

**Ground truth**: This is correct but the logic flow in the discovery diagram is incomplete. The condition `stance.option_scores != new_stance.build_option_scores` is evaluated AFTER `build_replacement` and `assign_attributes_and_files`.

**Priority: LOW** - Technically correct but could mislead Go implementation about when comparisons happen.

---

## Areas of Unclear or Incomplete Discovery

### 1. `build_replacement` Method Details

**Discovery reference**: `services.md` line 117-128 shows the method but doesn't explain:
- When `reason_format` is preserved vs. reset
- How `inviter_id` propagates to replacement stances
- Why `latest: true` is set in `build_replacement` (not service)

**File to investigate**: `/Users/z/Code/llmio/orig/loomio/app/models/stance.rb:120-128`

### 2. Transaction Boundary Implications

**Discovery notes**: Uses `update_columns` to bypass validations

**Unclear**: What happens if `new_stance.save!` fails after `update_columns(latest: false)`? The transaction should rollback, but:

```ruby
Stance.transaction do
  stance.update_columns(latest: false)  # Bypasses validations
  new_stance.save!                       # Runs validations
end
```

**Question**: Is there test coverage for transaction rollback scenarios?

### 3. MessageChannelService Integration

**Discovery mentions** (line 50):
```ruby
MessageChannelService.publish_models([stance], group_id: stance.poll.group_id)
```

**Question**: Why is the OLD stance published after creating a new one? Is this for WebSocket notification of the `latest: false` change?

### 4. Guest Stance Handling

**Discovery doesn't address**: How does the 15-minute threshold apply to guest stances? The `redeem` method exists but its interaction with `update` is unclear.

---

## Questions for Third Party

### High Priority

1. **Standalone poll behavior**: Your discovery doesn't mention that polls without `discussion_id` always update in place. Was this intentional omission or overlooked?

2. **paper_trail integration**: The Stance model uses `has_paper_trail`. How does this interact with the `latest` flag pattern for audit trails? Are both needed for Go rewrite?

3. **Test coverage claim verification**: You state tests don't cover the 15-minute threshold. Did you check:
   - Integration tests that might use `Timecop` or `travel_to`?
   - Factory defaults that set `updated_at` values?

### Medium Priority

4. **Event query purpose**: Line 34 in `stance_service.rb` queries for an event:
   ```ruby
   event = Event.where(eventable: stance, discussion_id: stance.poll.discussion_id).order('id desc').first
   ```
   But this `event` variable is never used. Is this dead code or a bug?

5. **Threshold configurability**: You confirm 15 minutes is hardcoded. Has there been discussion of making this configurable? Any feature requests or comments in the codebase?

### Low Priority

6. **Anonymous poll handling**: When `poll.anonymous?` is true, how does the revision threshold interact with scrubbed `participant_id`?

7. **Uncast interaction**: After `uncast`, can a user immediately vote again? Does the 15-minute threshold apply to the gap between uncast and re-vote?

---

## Specific Code References Requiring Investigation

| File | Lines | Issue | Priority |
|------|-------|-------|----------|
| `orig/loomio/app/services/stance_service.rb` | 34 | Unused `event` variable - dead code? | MEDIUM |
| `orig/loomio/app/services/stance_service.rb` | 36-39 | Full conditional not diagrammed in discovery | HIGH |
| `orig/loomio/app/models/stance.rb` | 67 | paper_trail config not in discovery | MEDIUM |
| `orig/loomio/app/models/stance.rb` | 114-156 | versions_count mechanism undocumented | LOW |
| `orig/loomio/spec/services/stance_service_spec.rb` | entire file | No `describe 'update'` block | HIGH |
| `orig/loomio/spec/controllers/api/v1/stances_controller_spec.rb` | 364-370 | Update test doesn't manipulate time | MEDIUM |

---

## Priority Summary

| Item | Priority | Rationale |
|------|----------|-----------|
| Standalone poll exception | **HIGH** | Implementation will be wrong without this |
| paper_trail versioning | **MEDIUM** | May need separate Go implementation |
| Test coverage gaps | **HIGH** | Regression risk during rewrite |
| Unused event variable | **MEDIUM** | Possible bug or cleanup needed |
| Guest stance interaction | **LOW** | Edge case, can defer |
| MessageChannelService publish | **LOW** | WebSocket implementation detail |

---

## Verification Commands

To validate discovery claims against source:

```bash
# Verify 15-minute threshold
grep -n "15\.minutes\|minutes\.ago" orig/loomio/app/services/stance_service.rb

# Verify discussion_id condition
grep -n "discussion_id" orig/loomio/app/services/stance_service.rb

# Check for time-based tests
grep -rn "travel_to\|Timecop\|freeze_time" orig/loomio/spec/ | grep -i stance

# Find paper_trail usage
grep -rn "has_paper_trail" orig/loomio/app/models/

# Check versions_count migration
cat orig/loomio/db/migrate/20181129031039_add_versions_count_to_stances.rb
```
