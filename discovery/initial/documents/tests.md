# Documents Domain: Tests

**Generated:** 2026-02-01
**Domain:** Documents and Attachments

---

## Overview

The documents domain has controller specs for the DocumentsController. Attachment testing appears to be covered implicitly through integration tests and model specs.

---

## 1. DocumentsController Spec

**Location:** `/spec/controllers/api/v1/documents_controller_spec.rb`

### Test Setup

**Factories Used:**
- :user, :another_user
- :group
- :discussion (public and private)
- :poll
- :comment
- :document

**Key Associations:**
- group_document: Document attached to group
- public_discussion_document: Document attached to public discussion
- private_discussion_document: Document attached to private discussion

### for_group Action Tests

#### Open Group Privacy

**Test: Non-members see all documents**
- Given: Open privacy group
- When: Non-member requests for_group
- Then: Returns all documents (group, public, private)

**Test: Members see all documents**
- Given: Open privacy group, user is member
- When: Member requests for_group
- Then: Returns all documents

**Test: Visitors see all documents**
- Given: Open privacy group
- When: Unauthenticated request for_group
- Then: Returns all documents

#### Closed Group Privacy

**Test: Members see all documents**
- Given: Closed privacy group, user is member
- When: Member requests for_group
- Then: Returns all documents

**Commented-Out Tests:**
Two tests for non-member visibility are commented out:
- Non-members should see only public documents
- Visitors should see only public documents

These may indicate incomplete implementation or changed requirements.

#### Secret Group Privacy

**Test: Non-members get 403**
- Given: Secret privacy group
- When: Non-member requests for_group
- Then: Returns 403 Forbidden

**Test: Members see all documents**
- Given: Secret privacy group, user is member
- When: Member requests for_group
- Then: Returns all documents

**Test: Non-members in closed context get 403**
- Given: Secret privacy group
- When: Non-member requests for_group
- Then: Returns 403 Forbidden

### for_discussion Action Tests

**Setup:**
- Discussion with poll and comment
- Documents attached to each (discussion, poll, comment)
- Another discussion with separate documents

**Test: Returns discussion documents for members**
- Given: User is group member
- When: Request for_discussion
- Then: Returns documents from discussion, poll, and comment
- And: Does not return documents from other discussions

**Test: Non-members get 403**
- Given: User is not group member
- When: Request for_discussion
- Then: Returns 403 Forbidden

**Confidence: 5/5** - Tests are comprehensive for the controller.

---

## 2. Document Factory

**Location:** `/spec/factories.rb`

### Definition

```
factory :document
  association :author, factory: :user
  association :model, factory: :discussion
  title: Faker::Name.name
  url: Faker::Internet.url
```

**Notes:**
- Default model is discussion
- Uses Faker for title and URL
- No file attachment by default

**Confidence: 5/5** - Factory is straightforward.

---

## 3. Attachment Factory

**Location:** `/spec/factories.rb`

### Definition

```
factory :attachment
  user
  filename: Faker::Name.name
  location: Faker::Company.logo
```

**Notes:**
This appears to be a legacy factory. The Attachment model now subclasses ActiveStorage::Attachment, so this factory may not work with current code.

**Confidence: 3/5** - Factory may be outdated.

---

## 4. Inferred Test Coverage

Based on the codebase structure, the following areas likely have test coverage:

### Model Validations

Document model validates:
- title presence
- doctype presence
- color presence

These should be tested in model specs (not found in search but likely exist).

### Service Authorization

DocumentService methods authorize before action:
- create: authorize! :create, document
- update: authorize! :update, document
- destroy: authorize! :destroy, document

Authorization specs likely exist in ability specs.

### HasRichText Concern

The concern includes complex behavior:
- HTML sanitization
- Attachment building
- Link preview sanitization

These may be tested in model specs for models that include HasRichText.

**Confidence: 3/5** - Inferred coverage, not directly verified.

---

## 5. Test Gaps Identified

### Missing Controller Tests

The following actions lack spec coverage:
- create action
- update action
- destroy action
- index action (standard)

### Attachment Controller Tests

No specs found for AttachmentsController:
- index action
- destroy action

### Direct Upload Tests

No specs found for DirectUploadsController:
- create action with file upload

### Service Tests

No specs found for DocumentService:
- create method
- update method
- destroy method

### Integration Tests

E2E tests (Nightwatch) may cover document/attachment flows but would require investigation of the vue/test directory.

**Confidence: 4/5** - Gaps are evident from search results.

---

## 6. Testing Patterns

### Authorization Testing Pattern

The controller spec tests authorization by:
1. Setting up group privacy level
2. Creating user with/without membership
3. Making request
4. Asserting response status or document IDs returned

### Document ID Verification

Tests extract document IDs from response:
```
json = JSON.parse response.body
document_ids = json['documents'].map { |d| d['id'] }
expect(document_ids).to include document.id
```

### Privacy Level Testing

Tests cover all three privacy levels:
- open: Everyone can see
- closed: Members and public content
- secret: Members only

**Confidence: 5/5** - Patterns are clear from existing tests.

---

## 7. Recommended Test Additions

### High Priority

1. **Document CRUD Actions**
   - Test create with file upload
   - Test update title/model
   - Test destroy authorization

2. **Attachment Controller**
   - Test index with group filtering
   - Test destroy with purge confirmation

3. **Document Service**
   - Test authorization flows
   - Test EventBus broadcasts

### Medium Priority

4. **HasRichText Concern**
   - Test attachment building
   - Test sanitization edge cases

5. **Direct Upload**
   - Test blob creation
   - Test preview URL generation

### Low Priority

6. **AttachmentQuery**
   - Test complex joins
   - Test discarded record filtering

---

## 8. E2E Test Locations (Inferred)

Nightwatch tests likely exist for:
- File upload in discussion creation
- Document attachment to polls
- File browser panel

These would be in `/vue/tests/e2e/` or similar directory.

**Confidence: 2/5** - E2E test locations not verified.

---

## Summary

Test coverage for the documents domain is partial:
- Controller specs cover visibility and authorization for list endpoints
- CRUD operations appear untested in specs
- Attachment controller lacks dedicated specs
- Service and concern tests not found

The existing tests focus on the most critical aspect: ensuring document visibility respects group privacy settings.

Recommended focus areas:
1. Add CRUD operation tests for documents
2. Add attachment controller tests
3. Add service-level tests for authorization
