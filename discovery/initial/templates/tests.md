# Templates Domain: Tests

**Generated:** 2026-02-01
**Confidence:** 3/5

---

## Overview

The templates domain has limited dedicated test coverage based on the investigation. No specific template spec files were found matching patterns like `*template*spec*`. Testing appears to be covered through:

1. Integration via other domain tests
2. E2E tests via Nightwatch
3. Controller tests for CRUD operations

---

## Discovered Test Patterns

### Service Layer Testing

Based on the service patterns used, expected test structure would be:

**DiscussionTemplateService specs:**
```
PSEUDO-CODE:
describe DiscussionTemplateService:
  describe create:
    context "when user is group admin":
      it "creates the template"
      it "sets author to actor"
      it "hides system template when key present"

    context "when user is not admin":
      it "raises CanCan::AccessDenied"

    context "with invalid attributes":
      it "returns false"

  describe update:
    it "updates allowed attributes"
    it "does not change group_id"

  describe initial_templates:
    it "returns board templates for board category"
    it "returns membership templates for membership category"
    it "returns fallback for unknown category"
    it "returns blank only for subgroups"

  describe default_templates:
    it "loads templates from YAML"
    it "processes i18n keys"
```

**PollTemplateService specs:**
```
PSEUDO-CODE:
describe PollTemplateService:
  describe group_templates:
    it "includes custom templates"
    it "includes system templates"
    it "applies position from group settings"
    it "marks hidden templates as discarded"

  describe create:
    it "creates template"
    it "hides system template being overridden"

  describe update:
    it "updates template attributes"
```

### Controller Testing

Expected structure for API tests:

**DiscussionTemplatesController specs:**
```
PSEUDO-CODE:
describe Api::V1::DiscussionTemplatesController:
  describe GET index:
    it "returns templates for user's group"
    it "initializes templates for new group"
    it "returns 401 for unauthorized"

  describe GET show:
    it "returns template from own group"
    it "returns public template"
    it "returns 404 for private foreign template"

  describe POST create:
    it "creates template for admin"
    it "returns 403 for non-admin"

  describe PATCH update:
    it "updates template"
    it "prevents changing group_id"

  describe DELETE destroy:
    it "deletes template for admin"

  describe POST positions:
    it "updates custom template positions"
    it "updates system template positions in group info"

  describe POST discard/undiscard:
    it "soft deletes and restores templates"

  describe POST hide/unhide:
    it "hides and shows system templates"

  describe GET browse:
    it "returns public templates"
    it "filters by query"
    it "initializes gallery if empty"
```

---

## Factory Requirements

For template testing, factories would need:

**DiscussionTemplate factory:**
```
PSEUDO-CODE:
factory :discussion_template:
  association :author, factory: :user
  association :group
  process_name { "Test Process" }
  process_subtitle { "Test Subtitle" }
  description_format { "html" }
```

**PollTemplate factory:**
```
PSEUDO-CODE:
factory :poll_template:
  association :author, factory: :user
  association :group
  poll_type { "proposal" }
  process_name { "Test Poll Process" }
  process_subtitle { "Test Poll Subtitle" }
  default_duration_in_days { 7 }
```

---

## E2E Test Patterns

Based on Nightwatch conventions, expected tests in `/vue/tests/e2e/specs/`:

**Discussion template tests:**
```
PSEUDO-CODE:
test "admin can create discussion template":
  login as admin
  navigate to /thread_templates?group_id=X
  click new template
  fill form
  save
  verify template appears in list

test "admin can edit discussion template":
  navigate to template edit
  modify values
  save
  verify changes persisted

test "admin can reorder templates":
  open template list
  drag template to new position
  verify new order saved

test "admin can hide/show system templates":
  click hide on system template
  verify hidden
  open settings
  click unhide
  verify visible

test "user can create discussion from template":
  navigate to template list
  click template
  verify form pre-filled
  save discussion
```

**Poll template tests:**
```
PSEUDO-CODE:
test "user can create poll from template":
  open poll chooser
  select template
  verify form pre-filled
  save poll

test "admin can customize poll templates":
  navigate to poll templates
  edit system template
  verify new custom template created
  verify system template hidden
```

---

## Dev Routes for Testing

The `/dev/` routes provide test scenario setup:

Expected nightwatch controller actions for templates:

```
PSEUDO-CODE:
def setup_template_test:
  create group with templates
  create user as admin
  login user
```

---

## Integration Test Coverage

Templates are likely tested indirectly through:

### Discussion creation tests
- Verify discussion_template_id saved
- Verify template attributes copied

### Poll creation tests
- Verify poll_template_id/key saved
- Verify poll options copied from template

### Group setup tests
- Verify initial templates assigned
- Verify category-based selection

---

## Suggested Test Additions

Based on gaps identified:

### Unit Tests Needed

1. **Model validation tests**
   - DiscussionTemplate validations (process_name, process_subtitle required)
   - PollTemplate validations (poll_type, duration required)
   - Quorum percentage normalization

2. **Association tests**
   - poll_template_ids extraction
   - poll_templates relationship

3. **Paper Trail tests**
   - Version tracking for edits
   - Attribute change recording

### Service Tests Needed

1. **Template hiding edge cases**
   - What happens with duplicate keys
   - Hiding already hidden template

2. **Initial template selection**
   - All category mappings
   - Subgroup fallback behavior

3. **Public template gallery**
   - Gallery seeding
   - Query filtering

### Controller Tests Needed

1. **Authorization edge cases**
   - Admin of parent group accessing subgroup
   - Non-member accessing public template

2. **Browse pagination**
   - Large result sets
   - Tag filtering

---

## Confidence Assessment

**Rating: 3/5**

Lower confidence due to:
- No dedicated spec files found for templates
- Test patterns inferred from codebase conventions
- Unable to verify actual test implementation

Higher confidence areas:
- Service patterns well-established, tests likely follow conventions
- E2E patterns visible from other specs
- Factory patterns documented in spec/factories.rb

Recommendation: Add dedicated model and service specs for templates to improve coverage visibility.
