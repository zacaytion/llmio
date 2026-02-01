# Stance Revision Threshold - Verification Checklist

## Overall Confidence: 5/5

All ground truth questions answered with direct code evidence.

---

## Claim Verification Matrix

### Discovery Claims

| # | Claim | Status | Evidence | File:Line |
|---|-------|--------|----------|-----------|
| 1 | 15-minute threshold exists | **PASS** | `stance.updated_at < 15.minutes.ago` | `stance_service.rb:39` |
| 2 | Threshold is 15 minutes exactly | **PASS** | Hardcoded `15.minutes` literal | `stance_service.rb:39` |
| 3 | Update within threshold modifies existing record | **PASS** | Falls through to `else` branch with `stance.save!` | `stance_service.rb:52-64` |
| 4 | Update after threshold creates new record | **PASS** | Creates `new_stance`, sets old `latest: false` | `stance_service.rb:42-51` |
| 5 | Old stance marked `latest: false` | **PASS** | `stance.update_columns(latest: false)` | `stance_service.rb:45` |

### Additional Conditions Discovered (Not in Original Discovery)

| # | Condition | Status | Evidence | File:Line |
|---|-----------|--------|----------|-----------|
| 6 | Requires poll to be in discussion | **NEW FINDING** | `stance.poll.discussion_id` check | `stance_service.rb:37` |
| 7 | Requires option_scores to change | **NEW FINDING** | `stance.option_scores != new_stance.build_option_scores` | `stance_service.rb:38` |
| 8 | Threshold not configurable | **CONFIRMED** | No constants, env vars, or settings found | N/A |

---

## Evidence Quality Assessment

| Evidence Type | Quality | Notes |
|---------------|---------|-------|
| Source code | **Excellent** | Direct inspection of service class |
| Code comments | **Good** | Developer intent documented inline |
| Test coverage | **Poor** | No tests for time-based behavior |
| Configuration | **N/A** | Value is hardcoded |
| Documentation | **None** | No external docs found |

---

## Verification Commands Run

| Command | Result | Purpose |
|---------|--------|---------|
| Read `stance_service.rb` | Complete file read | Primary evidence source |
| Grep `15\|minute\|threshold` | Found line 39-40 | Locate threshold logic |
| Grep `latest` in stance_service | Found 3 occurrences | Trace flag management |
| Read `stance.rb` | Complete file read | Model structure and methods |
| Read `stance_service_spec.rb` | No update tests | Test coverage gap confirmed |
| Grep `\.minutes` in app/ | Only 2 hits | Confirm no other thresholds |

---

## Discovery Document Accuracy

### Original Discovery Statement
> "Discovery documents a **15-minute vote revision window** that determines whether updating a vote creates a new stance record (preserving history) or updates in place."

**Assessment**: ACCURATE but INCOMPLETE

The discovery correctly identified:
- The existence of a 15-minute window
- The create-new vs update-existing behavior

The discovery missed:
- The `discussion_id` requirement (standalone polls always update in place)
- The `option_scores` change requirement (reason-only changes always update in place)

---

## Confidence Breakdown by Question

### Q1: Does a 15-minute threshold exist?
- **Confidence: 5/5**
- Direct code evidence at `stance_service.rb:39`
- Inline comment confirms intent at line 40

### Q2: What is the exact threshold value?
- **Confidence: 5/5**
- Value is `15.minutes` (Rails duration)
- No configuration mechanism found
- Grep confirmed no other similar patterns

### Q3: What logic determines create-new vs update-existing?
- **Confidence: 5/5**
- Complete conditional documented
- Four conditions identified and verified
- Transaction boundaries understood

### Q4: Is this behavior documented/tested?
- **Confidence: 5/5**
- Code comment exists (partial documentation)
- No unit tests for threshold (test gap confirmed)
- Controller tests don't cover time manipulation

---

## Recommendations

### Immediate
1. Add unit tests for `StanceService.update` with time manipulation
2. Document the four conditions in code comments or CLAUDE.md

### Future
1. Consider extracting `15.minutes` to a named constant for discoverability
2. Add E2E test for vote revision after 15+ minutes
3. Consider whether standalone polls should have the same behavior as discussion polls

---

## Sign-off

| Item | Status |
|------|--------|
| Ground truth questions answered | YES |
| Code evidence provided | YES |
| Line numbers documented | YES |
| Test gaps identified | YES |
| Confidence justified | YES |
