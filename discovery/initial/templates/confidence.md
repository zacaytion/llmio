# Templates Domain: QA Confidence Assessment

**Review Date:** 2026-02-01
**Reviewer:** QA Agent
**Documents Reviewed:** models.md, services.md, controllers.md, frontend.md, tests.md

---

## 1. Checklist Results

### Models Documentation (models.md)

| Item | Status | Notes |
|------|--------|-------|
| Model location accurate | PASS | DiscussionTemplate and PollTemplate at expected paths |
| Schema documentation accurate | PASS | Column definitions match actual model attributes |
| Concerns listed correctly | PASS | Discard::Model, HasRichText, CustomCounterCache::Model confirmed |
| Relationships accurate | PASS | belongs_to :author, :group verified |
| Validations documented | PASS | process_name, process_subtitle, poll_type presence validated |
| Enums documented | PASS | notify_on_closing_soon, hide_results, stance_reason_required verified |
| Paper Trail attributes | PASS | Tracked attributes match has_paper_trail declaration |
| System templates YAML location | PASS | config/discussion_templates.yml and config/poll_templates.yml exist |
| Key methods documented | PASS | poll_templates, poll_template_ids, dump_i18n exist |

### Services Documentation (services.md)

| Item | Status | Notes |
|------|--------|-------|
| Service location accurate | PASS | Both services at /app/services/ |
| Create flow documented | PASS | Authorization, author assignment, key hiding logic verified |
| Update flow documented | PASS | Excludes group_id change verified |
| initial_templates categories | PASS | All 4 categories (board, membership, self_managing, other) verified |
| default_templates i18n handling | PASS | _i18n suffix processing confirmed |
| group_templates method exists | PARTIAL | Only exists on PollTemplateService, not DiscussionTemplateService |
| Authorization pattern | PASS | actor.ability.authorize! pattern confirmed |
| Template hiding vs discarding | PASS | Both mechanisms correctly documented |

### Controllers Documentation (controllers.md)

| Item | Status | Notes |
|------|--------|-------|
| Controller locations | PASS | Both at api/v1/ namespace |
| Routes configuration | PASS | All documented routes match config/routes.rb |
| Standard CRUD actions | PASS | index, show, create, update, destroy confirmed |
| Custom actions documented | PASS | positions, discard, undiscard, hide, unhide verified |
| browse_tags method | PASS | Exists and implementation matches description |
| browse method | PASS | Public template search with filtering confirmed |
| settings action (poll only) | PASS | categorize_poll_templates toggle verified |
| Authorization via adminable_groups | PASS | Pattern confirmed in controller code |

### Frontend Documentation (frontend.md)

| Item | Status | Notes |
|------|--------|-------|
| Model locations | PASS | Both models at vue/src/shared/models/ |
| Default values documented | PASS | Match actual defaultValues() methods |
| Relationships documented | PASS | belongsTo author, group confirmed |
| buildDiscussion/buildPoll methods | PASS | Implementation matches documentation |
| pollTemplates() method on DiscussionTemplateModel | PASS | Returns compact mapped results |
| Frontend service actions | PASS | All 8 actions (edit, move, discard, destroy, hide, unhide, etc.) verified |
| Component locations | PASS | All documented components exist at stated paths |
| Routes in routes.js | ASSUMED | Not directly verified but paths align with controller routes |

### Tests Documentation (tests.md)

| Item | Status | Notes |
|------|--------|-------|
| Factory definitions exist | PASS | poll_template and discussion_template factories in spec/factories.rb |
| Spec files existence | PARTIAL | No dedicated template spec files, only mentioned in group_export_service_spec.rb |
| Test patterns inferred | ACKNOWLEDGED | Documentation correctly notes limited coverage |
| E2E patterns described | ASSUMED | Not verified against actual nightwatch specs |

---

## 2. Confidence Scores

| Document | Score | Assessment |
|----------|-------|------------|
| models.md | **5/5** | Highly accurate. Schema, concerns, relationships, and methods all verified against source code. |
| services.md | **4/5** | Mostly accurate. Minor issue: references DiscussionTemplateService.group_templates which doesn't exist (only PollTemplateService has this method). The controller uses it anyway via a naming error. |
| controllers.md | **4/5** | Accurate. Minor note: hide/unhide actions in DiscussionTemplatesController incorrectly call DiscussionTemplateService.group_templates (which doesn't exist) - this appears to be a bug in the actual code. |
| frontend.md | **4/5** | Good coverage. Minor issue: notes frontend service is "named PollTemplateService in file" for discussion templates - this is accurate but confusing naming pattern. |
| tests.md | **3/5** | Appropriately honest about limited coverage. Documentation acknowledges uncertainty and provides reasonable inferences. |

**Overall Domain Confidence: 4/5**

---

## 3. Issues Found

### Critical Issues

None found.

### Significant Issues

1. **Bug in Controller Code (verified via source):**
   - `/app/controllers/api/v1/discussion_templates_controller.rb` lines 107 and 122 call `DiscussionTemplateService.group_templates(group: @group)` but this method does not exist on DiscussionTemplateService.
   - The method only exists on PollTemplateService.
   - This appears to be a copy-paste error or the method was never added to DiscussionTemplateService.
   - The hide/unhide actions for discussion templates likely produce runtime errors.

2. **Documentation Naming Confusion (frontend.md):**
   - The file `/vue/src/shared/services/discussion_template_service.js` exports a class named `PollTemplateService` (line 12), not `DiscussionTemplateService`.
   - This is confusing and could lead to misunderstandings. Documentation correctly notes this but it reflects a codebase issue.

### Minor Issues

1. **models.md schema table:**
   - Missing `key` column for PollTemplate serializer (documented in serializer but not in models.md schema).
   - Missing `meeting_duration` and `can_respond_maybe` columns which exist in PollTemplate.

2. **tests.md:**
   - The document states "No specific template spec files were found" but doesn't mention the integration tests in `group_export_service_spec.rb` which do test template creation and retrieval.

---

## 4. Uncertainties

1. **E2E Test Coverage:**
   - Could not verify whether Nightwatch E2E tests actually cover template workflows. The patterns described in tests.md are reasonable inferences but not confirmed.

2. **Frontend Routes:**
   - Did not directly verify `/vue/src/routes.js` to confirm all documented routes exist. Assumed correct based on controller routes alignment.

3. **System Template Count:**
   - models.md states "14 built-in poll templates" and "13 built-in discussion templates" - not individually verified, counts appear reasonable from YAML inspection.

4. **NullGroup Template Defaults:**
   - models.md references `/app/models/concerns/null/group.rb` for NullGroup template defaults - did not verify this file exists or contains the documented behavior.

5. **Public Template Gallery:**
   - The browse functionality and gallery seeding logic documented but not tested. The "templates" group creation and helper_bot ownership not verified.

---

## 5. Revision Recommendations

### High Priority

1. **services.md - Fix incorrect method reference:**
   - Line referencing `DiscussionTemplateService.group_templates` should either:
     a) Be removed if the method intentionally doesn't exist, OR
     b) Note that this method is missing and should be added (bug report)

2. **controllers.md - Document the bug:**
   - Add a note that hide/unhide actions in DiscussionTemplatesController may have a runtime error due to missing `group_templates` method.

### Medium Priority

3. **models.md - Complete schema documentation:**
   - Add `key` column to PollTemplate schema table
   - Add `meeting_duration` and `can_respond_maybe` to PollTemplate schema

4. **frontend.md - Clarify naming issue:**
   - Make the `PollTemplateService` naming in discussion_template_service.js more prominent as it's a significant source of confusion.

5. **tests.md - Reference existing tests:**
   - Add reference to `spec/services/group_export_service_spec.rb` which contains template testing examples.
   - Update confidence score rationale to acknowledge these exist.

### Low Priority

6. **Add ability module documentation:**
   - The ability modules (`Ability::DiscussionTemplate`, `Ability::PollTemplate`) exist but aren't documented. These are simple (admin-only for create/update) but should be mentioned.

7. **Verify NullGroup defaults:**
   - Confirm `/app/models/concerns/null/group.rb` exists and document its template-related default values.

---

## Summary

The templates domain documentation is **well-written and largely accurate**. The main finding is a potential bug in the production code where `DiscussionTemplatesController` calls a non-existent method. The documentation accurately reflects what the code is *trying* to do but should note where it may fail.

Test coverage documentation is appropriately cautious given the limited dedicated specs. The frontend documentation is thorough and the models documentation is excellent.

**Recommendation:** Address the high-priority revisions, particularly documenting the `group_templates` method discrepancy, then mark this domain as reviewed and approved.
