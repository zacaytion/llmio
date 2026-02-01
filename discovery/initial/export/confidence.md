# Export Domain: QA Review and Confidence Assessment

**Review Date:** 2026-02-01
**Reviewer:** QA Agent
**Documents Reviewed:** models.md, services.md, controllers.md, frontend.md, tests.md

---

## 1. Checklist Results

### Models Documentation (models.md)

| Check | Status | Notes |
|-------|--------|-------|
| File paths are correct | PASS | `/app/models/concerns/group_export_relations.rb` and `/app/models/concerns/discussion_export_relations.rb` verified |
| Association names match code | PASS | All documented associations verified in source code |
| Privacy filter logic accurate | PASS | `where("anonymous = false OR closed_at is not null")` confirmed at line 9 of group_export_relations.rb |
| JSON_PARAMS exclusions documented | PASS | Matches constants in group_export_service.rb lines 23-31 |
| Union query methods exist | PASS | `all_users`, `all_tags`, `all_groups`, `all_events`, `all_notifications`, `all_documents`, `all_reactions` all verified |
| Discussion export concern exists | PASS | File exists with correct associations |

### Services Documentation (services.md)

| Check | Status | Notes |
|-------|--------|-------|
| RELATIONS constant accurate | PASS | Matches code in group_export_service.rb lines 2-21 |
| JSON_PARAMS constant accurate | PASS | Verified at lines 23-31 |
| BACK_REFERENCES documented | PASS | Extensive mapping verified at lines 33-108 |
| export() method flow accurate | PASS | Method at lines 155-194 matches documented flow |
| import() method flow accurate | PASS | Method at lines 224-318 matches documented process |
| Batch size documented correctly | PASS | `find_each(batch_size: 20000)` confirmed at line 163 |
| GroupExporter fields accurate | PASS | EXPORT_MODELS at lines 4-12 of group_exporter.rb verified |
| PollExporter functionality accurate | PASS | poll_exporter.rb verified |
| GroupService.export method | PARTIAL | Documentation says requires `:show` permission, but actual code uses `:show` at line 142, not `:export` as stated elsewhere |

### Controllers Documentation (controllers.md)

| Check | Status | Notes |
|-------|--------|-------|
| API routes documented correctly | PASS | Lines 64-73 of api/v1/groups_controller.rb verified |
| Non-API GroupsController export | PASS | Line 22-27 of groups_controller.rb verified |
| PollsController export | PASS | Lines 8-16 of polls_controller.rb verified |
| DiscussionsController empty | PASS | Controller is empty as documented (just class definition) |
| Authorization requirements | PASS | Group export requires `:export` permission (admins only), Poll export requires `:export` permission (show + results visible) |
| Response patterns accurate | PASS | JSON `{ success: :ok }` for async, direct render for sync |

### Frontend Documentation (frontend.md)

| Check | Status | Notes |
|-------|--------|-------|
| Component location correct | PASS | `/vue/src/components/group/export_data_modal.vue` verified |
| Export methods exist on model | PASS | Lines 249-255 of group_model.js verified |
| Three format options documented | PASS | CSV, HTML, JSON all present in template |
| Confirmation modal flow | PASS | `openConfirmModalForJson()` and `openConfirmModalForCSV()` verified |
| HTML export direct link | PASS | Line 61 shows direct link pattern |
| Translation keys documented | PASS | Keys match component usage |

### Tests Documentation (tests.md)

| Check | Status | Notes |
|-------|--------|-------|
| Test file location correct | PASS | `/spec/services/group_export_service_spec.rb` exists |
| Test scenario accurately described | PASS | `create_scenario` method matches documentation at lines 4-71 |
| Round-trip test flow accurate | PASS | Export, truncate, import flow confirmed at lines 80-87 |
| Verification assertions documented | PASS | Assertions at lines 89-147 match documentation |
| Test coverage gaps identified | PASS | CSV, HTML, authorization, attachment tests correctly noted as missing |
| Commented out test noted | PASS | Line 149-150 confirms commented `import on existing tables` test |

---

## 2. Confidence Scores

| Document | Score | Flag |
|----------|-------|------|
| models.md | 5/5 | - |
| services.md | 4/5 | Minor inconsistency |
| controllers.md | 5/5 | - |
| frontend.md | 4/5 | - |
| tests.md | 4/5 | - |

**Overall Domain Confidence: 4.4/5**

---

## 3. Issues Found

### Issue 1: Permission Inconsistency in Documentation (services.md)

**Severity:** Low

**Description:** The services.md states that `GroupService.export` requires `:show` permission, but:
- The controller uses `load_and_authorize(:group, :export)`
- The `Ability::Group` module defines `:export` as requiring admin status
- The actual `GroupService.export` method uses `actor.ability.authorize! :show, group`

This creates a dual-authorization pattern where:
1. Controller checks `:export` permission (admin only)
2. Service checks `:show` permission (redundant)

The documentation should clarify this dual-check pattern.

**Affected Code:**
- `/app/controllers/api/v1/groups_controller.rb` line 65: `load_and_authorize(:group, :export)`
- `/app/services/group_service.rb` line 142: `actor.ability.authorize! :show, group`

### Issue 2: DiscussionExportRelations Bug

**Severity:** Medium

**Description:** The `DiscussionExportRelations` concern has an incorrect foreign key reference:

```ruby
has_many :exportable_polls, -> { where("anonymous = false OR closed_at is not null") },
         class_name: 'Poll', foreign_key: :group_id  # Should be :discussion_id
```

This concern is on Discussion but uses `group_id` as the foreign key. This would export polls belonging to the group, not the discussion. This appears to be a copy-paste error from `GroupExportRelations`.

**Location:** `/app/models/concerns/discussion_export_relations.rb` line 5

### Issue 3: Documentation Claims Feature Not Implemented

**Severity:** Low

**Description:** The controllers.md correctly notes that `/d/:key/export` route exists but the controller action is not implemented. However, this gap is only mentioned in the controllers doc, not in the models or services docs which reference `DiscussionExportRelations`. The domain documentation should be more explicit that discussion-level export is not functional.

---

## 4. Uncertainties

### Uncertainty 1: Attachment Download During Import

The `import` method delegates attachment downloads to `DownloadAttachmentWorker`, but the documentation does not clarify:
- What happens if the source URL is no longer accessible
- Whether there's retry logic for failed downloads
- Whether the import succeeds if attachments fail

### Uncertainty 2: Large Export Handling

Documentation mentions batch processing but does not cover:
- Maximum practical export size
- Disk space requirements for `/tmp` storage
- Memory consumption during attachment processing (not batched)
- Timeout considerations for very large groups

### Uncertainty 3: Import Conflict Behavior

The test has a commented-out test case for "import on existing tables". The documentation notes this but does not fully specify:
- What happens when importing a user that already exists (email conflict handling is documented)
- Behavior when importing groups/discussions with duplicate keys
- Whether `reset_keys: true` is the recommended approach for migration scenarios

### Uncertainty 4: CSV Export Scope

The `GroupExporter` uses `.in_organisation(group)` scope. Documentation does not clarify whether this includes:
- Subgroups and their content
- The same depth as JSON export
- The same privacy filtering as JSON export for anonymous polls

---

## 5. Revision Recommendations

### High Priority

1. **Fix DiscussionExportRelations Bug Documentation**
   - Add a note in models.md that `DiscussionExportRelations` has a bug where polls use wrong foreign key
   - Consider whether this concern is even used anywhere (it may be dead code)

2. **Clarify Permission Model**
   - Update services.md to explain the dual-authorization pattern
   - Document that controller checks `:export` (admin) and service checks `:show` (redundant)

### Medium Priority

3. **Document CSV Export Limitations**
   - Add section in services.md comparing CSV vs JSON export scope
   - Clarify whether CSV includes subgroups
   - Note that CSV does not apply anonymous poll privacy filtering

4. **Expand Test Coverage Gap Analysis**
   - Add specific recommendations for missing controller authorization tests
   - Prioritize anonymous poll privacy test as it's a security-sensitive behavior

### Low Priority

5. **Add Import Edge Cases**
   - Document attachment download failure handling
   - Clarify conflict resolution behavior
   - Add example of `reset_keys: true` use case

6. **Performance Considerations Section**
   - Add guidance on export size limitations
   - Document memory/disk space requirements
   - Note that synchronous HTML exports may timeout on large groups

---

## Summary

The export domain documentation is generally accurate and comprehensive. The main issues are:

1. A likely bug in `DiscussionExportRelations` that uses wrong foreign key
2. Minor inconsistency in permission documentation
3. Some gaps in edge case and error handling documentation

The documentation correctly identifies that test coverage is incomplete, particularly for CSV exports, authorization, and anonymous poll privacy. The core JSON export/import functionality is well-tested and documented.

**Recommendation:** The documentation is suitable for use with the noted caveats. The `DiscussionExportRelations` bug should be investigated to determine if it's actively used or dead code.
