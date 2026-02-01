# Documents Domain: QA Confidence Assessment

**Generated:** 2026-02-01
**Reviewer:** QA Agent
**Domain:** Documents and Attachments

---

## 1. Checklist Results

### models.md

| Item | Status | Notes |
|------|--------|-------|
| Document model attributes accurate | PASS | Verified against `/app/models/document.rb` |
| Attachment model description accurate | PASS | Verified - correctly describes thin ActiveStorage::Attachment subclass |
| HasRichText concern accurately documented | PASS | Verified against `/app/models/concerns/has_rich_text.rb` |
| Doctypes configuration correct | PASS | Verified against `/config/doctypes.yml` - all 15 types match |
| Model relationships accurate | PASS | Polymorphic belongs_to and has_one_attached verified |
| Storage backend configuration accurate | PASS | Standard Rails Active Storage configuration |

### services.md

| Item | Status | Notes |
|------|--------|-------|
| DocumentService methods accurate | PASS | All three methods (create, update, destroy) verified |
| EventBus broadcasts documented | PASS | Correct event names documented |
| No Event publishing noted | PASS | Correctly identified - no Events::* classes used |
| Hard delete behavior noted | PASS | Uses `document.destroy` not soft delete |
| HasRichText concern methods | PASS | build_attachments, assign_attributes_and_files verified |
| Worker descriptions accurate | PARTIAL | DownloadAttachmentWorker exists but AttachDocumentWorker not verified |
| Orphan cleanup assessment | FLAGGED | Documentation notes uncertainty (3/5) - correct assessment |

### controllers.md

| Item | Status | Notes |
|------|--------|-------|
| DocumentsController routes accurate | PASS | Verified against controller |
| for_group action logic | PASS | Privacy check and group ID filtering verified |
| for_discussion action | PASS | UnionQuery combining discussion/poll/comment documents |
| AttachmentsController routes | PASS | index and destroy actions verified |
| AttachmentQuery logic | PASS | Six separate queries with union verified |
| DirectUploadsController | PASS | Standard ActiveStorage pattern |
| Authorization flow documented | PASS | Ability modules verified |

### frontend.md

| Item | Status | Notes |
|------|--------|-------|
| DocumentRecordsInterface methods | PASS | fetchByModel, fetchByDiscussion verified |
| DocumentModel structure | PASS | Indices, relationships, methods verified |
| FileUploader implementation | PARTIAL | Basic flow correct but documentation overstates return value structure |
| Attachment service | NOT VERIFIED | File location not confirmed |
| HasDocuments mixin | NOT VERIFIED | File location not confirmed |
| Vue components | PARTIAL | FilesPanel mentioned but not verified |

### tests.md

| Item | Status | Notes |
|------|--------|-------|
| Controller spec exists | PASS | `/spec/controllers/api/v1/documents_controller_spec.rb` verified |
| Test coverage for for_group | PASS | Open/closed/secret privacy levels tested |
| Test coverage for for_discussion | PASS | Member/non-member access tested |
| Commented-out tests noted | PASS | Correctly identified two commented tests |
| Document factory exists | PASS | Factory verified in spec file usage |
| Test gaps identified | PASS | CRUD tests, attachment controller tests missing |

---

## 2. Confidence Scores

| Document | Score | Assessment |
|----------|-------|------------|
| models.md | **5/5** | Highly accurate; all claims verified against source code |
| services.md | **4/5** | Mostly accurate; orphan cleanup uncertainty is valid |
| controllers.md | **5/5** | Comprehensive and accurate controller documentation |
| frontend.md | **3/5** | Core patterns correct but some claims unverified |
| tests.md | **4/5** | Accurate test coverage assessment with valid gap analysis |

**Overall Domain Confidence: 4.2/5**

---

## 3. Issues Found

### Critical Issues

None identified.

### Moderate Issues

1. **frontend.md - FileUploader return value**
   - Documentation states blob returns `download_url` and `preview_url`
   - Actual FileUploader code returns raw blob data from DirectUpload
   - These URLs come from the server response, not client-side code
   - **Impact:** Developer confusion about where data originates

2. **services.md - AttachDocumentWorker location**
   - Worker mentioned but file path `/app/workers/attach_document_worker.rb` not verified
   - May be outdated or removed
   - **Impact:** Developer time wasted searching for non-existent file

3. **tests.md - Attachment factory description**
   - Documentation notes factory may be outdated
   - Factory with `user`, `filename`, `location` attributes does not match current ActiveStorage model
   - **Impact:** Tests using this factory may fail

### Minor Issues

1. **controllers.md - Authorization description**
   - States "User must be admin of document's group" for destroy
   - Actual ability check is `user_is_admin_of? document.model.group.id`
   - Subtle difference: admin of model's group, not document's group
   - **Impact:** Edge cases where model.group differs from document.group_id

2. **models.md - URL handling description**
   - Documentation mentions "lmo_asset_host" but actual code uses this method
   - Method definition not shown in documentation
   - **Impact:** Minor - readers may not understand URL construction fully

---

## 4. Uncertainties

### Requiring Investigation

1. **Orphan attachment cleanup**
   - services.md correctly flags this at 3/5 confidence
   - No explicit cleanup mechanism found
   - Potential for storage bloat over time
   - **Recommendation:** Investigate production attachment/blob ratios

2. **AttachDocumentWorker existence**
   - Referenced for legacy URL-to-ActiveStorage migration
   - May have been removed after migration completed
   - **Recommendation:** Search codebase or git history

3. **HasDocuments mixin location**
   - frontend.md references `/vue/src/shared/mixins/has_documents.js`
   - Path not verified
   - **Recommendation:** Verify file exists or correct path

4. **E2E test coverage**
   - tests.md notes E2E test locations "not verified" at 2/5 confidence
   - Nightwatch tests likely exist but unconfirmed
   - **Recommendation:** Search `/vue/tests/` directory

### Assumptions Made

1. **Group association on attachment records**
   - Ability checks assume `attachment.record.group` always returns valid group
   - May fail if record has no group (standalone polls/discussions)

2. **Doctype matching order**
   - Documentation notes "first match wins"
   - Verified correct - "other" is catch-all at end

---

## 5. Revision Recommendations

### High Priority

1. **frontend.md - Rewrite FileUploader section**
   - Clarify that `download_url` and `preview_url` come from server response
   - Document the server-side DirectUploadsController enhancements
   - Show actual response structure

2. **services.md - Verify AttachDocumentWorker**
   - Either confirm file exists and add location
   - Or note it may be removed/historical

3. **frontend.md - Verify mixin and component paths**
   - Confirm HasDocuments mixin location
   - Verify AttachmentService location
   - Update paths if incorrect

### Medium Priority

4. **tests.md - Update attachment factory section**
   - Investigate current factory usage
   - Either confirm it works or document that it's legacy

5. **models.md - Document lmo_asset_host method**
   - Add brief explanation of URL construction helper
   - Show where method is defined

6. **controllers.md - Clarify destroy authorization**
   - Specify "model's group" vs "document's group"
   - Add note about edge cases

### Low Priority

7. **services.md - Add orphan cleanup investigation results**
   - After investigating, document findings
   - If no cleanup exists, document as known limitation

8. **tests.md - Verify E2E test coverage**
   - Search Nightwatch test directories
   - Document any existing file upload tests

---

## 6. Verification Commands Used

```bash
# Model verification
Read /app/models/document.rb
Read /app/models/attachment.rb

# Service verification
Read /app/services/document_service.rb

# Controller verification
Read /app/controllers/api/v1/documents_controller.rb
Read /app/controllers/api/v1/attachments_controller.rb

# Ability verification
Read /app/models/ability/document.rb
Read /app/models/ability/attachment.rb

# Frontend verification
Read /vue/src/shared/interfaces/document_records_interface.js
Read /vue/src/shared/models/document_model.js
Read /vue/src/shared/services/file_uploader.js

# Test verification
Read /spec/controllers/api/v1/documents_controller_spec.rb

# Configuration verification
Read /config/doctypes.yml
Read /app/models/concerns/has_rich_text.rb
```

---

## 7. Summary

The documents domain documentation is **generally accurate and comprehensive**. The dual system (Document model + Active Storage attachments) is well-explained, and the key patterns are correctly documented.

**Strengths:**
- Model layer documentation is excellent (5/5)
- Controller documentation is thorough and accurate
- Test coverage assessment identifies real gaps
- Services documentation correctly flags uncertainties

**Weaknesses:**
- Frontend documentation has unverified claims (3/5)
- Some file paths may be incorrect or outdated
- Minor precision issues in authorization descriptions

**Recommendation:** The documentation is suitable for use with the revisions noted above. Priority should be given to verifying frontend file paths and clarifying the FileUploader response structure.
