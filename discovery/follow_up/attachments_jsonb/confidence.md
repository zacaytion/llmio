# Verification Checklist - Attachments JSONB Default

## Overall Confidence: 5/5

All claims verified against authoritative source code with no contradictions found.

---

## Claim 1: Database default is `[]` (empty array)

**Status: PASS**

| Verification | Source | Result |
|--------------|--------|--------|
| schema.rb check | `/Users/z/Code/loomio/db/schema.rb` lines 189, 259, 298, 467, 645, 773, 866, 1036 | All show `default: []` |
| Migration history | `20190926001607_change_attachments_default_to_array.rb` | Confirms intentional change from `{}` to `[]` |

**Confidence: 5/5** - Schema is definitive source of truth for database state.

---

## Claim 2: All tables use identical default

**Status: PASS**

| Table | Line in schema.rb | Default |
|-------|-------------------|---------|
| comments | 189 | `[]` |
| discussion_templates | 259 | `[]` |
| discussions | 298 | `[]` |
| groups | 467 | `[]` |
| outcomes | 645 | `[]` |
| polls | 773 | `[]` |
| stances | 866 | `[]` |
| users | 1036 | `[]` |

**Confidence: 5/5** - Exhaustive check of all 8 tables with attachments column.

---

## Claim 3: Application code expects array

**Status: PASS**

### Backend Evidence

| Check | File | Line | Evidence |
|-------|------|------|----------|
| Array creation | `app/models/concerns/has_rich_text.rb` | 95 | `self[:attachments] = files.map do |file|` - `.map` returns array |
| Array iteration | `app/views/event_mailer/common/_attachments.html.haml` | 14 | `resource.attachments.any?` - array method |

### Frontend Evidence

| Model | File | Default Value |
|-------|------|---------------|
| CommentModel | `vue/src/shared/models/comment_model.js:26` | `attachments: []` |
| DiscussionModel | `vue/src/shared/models/discussion_model.js:57` | `attachments: []` |
| PollModel | `vue/src/shared/models/poll_model.js:75` | `attachments: []` |
| UserModel | `vue/src/shared/models/user_model.js:23` | `attachments: []` |
| GroupModel | `vue/src/shared/models/group_model.js:44` | `attachments: []` |
| StanceModel | `vue/src/shared/models/stance_model.js:24` | `attachments: []` |
| OutcomeModel | `vue/src/shared/models/outcome_model.js:23` | `attachments: []` |

**Confidence: 5/5** - Consistent across backend and all frontend models.

---

## Claim 4: Serializers pass through database value

**Status: PASS**

| Check | Evidence |
|-------|----------|
| Simple attribute inclusion | All serializers use `:attachments` in attributes list without custom methods |
| No transformation | No `def attachments` override found in any serializer |

Verified in:
- `/Users/z/Code/loomio/app/serializers/comment_serializer.rb:13`
- `/Users/z/Code/loomio/app/serializers/discussion_serializer.rb:35`
- `/Users/z/Code/loomio/app/serializers/poll_serializer.rb:4,67`
- `/Users/z/Code/loomio/app/serializers/group_serializer.rb:39`
- `/Users/z/Code/loomio/app/serializers/stance_serializer.rb:15`
- `/Users/z/Code/loomio/app/serializers/outcome_serializer.rb:10`
- `/Users/z/Code/loomio/app/serializers/user_serializer.rb:9`

**Confidence: 5/5** - Pattern is consistent and simple.

---

## Claim 5: Documentation discrepancy explained by migration history

**Status: PASS**

| Date | Migration | Default Used |
|------|-----------|--------------|
| 2019-03-26 | `add_attachments_to_comments.rb` | `default: {}` |
| 2019-03-26 | `add_attachments_to_rich_text_models.rb` | `default: {}` |
| 2019-09-26 | `change_attachments_default_to_array.rb` | `default: []` |
| 2023-07-31 | `create_discussion_templates_table.rb` | `default: []` |

**Confidence: 5/5** - Migration history clearly shows the evolution.

---

## Summary

| Claim | Confidence | Status |
|-------|------------|--------|
| Database default is `[]` | 5/5 | PASS |
| All tables use same default | 5/5 | PASS |
| Application expects array | 5/5 | PASS |
| Serializers pass through value | 5/5 | PASS |
| Discrepancy explained | 5/5 | PASS |

**Total Verification Score: 25/25 (100%)**

### Confidence Justification

Rating 5/5 because:
1. All claims verified against actual source code
2. No contradictions found
3. Multiple independent sources confirm each claim
4. Historical migration explains documentation discrepancy
5. Consistent behavior across backend, frontend, and database
