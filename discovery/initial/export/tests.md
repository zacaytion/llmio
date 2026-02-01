# Export Domain: Tests

**Generated:** 2026-02-01
**Domain:** Export functionality for groups, discussions, and polls

---

## Overview

The export domain has one comprehensive backend test file covering the JSON export/import cycle. Frontend E2E tests do not appear to cover export functionality.

---

## 1. Backend Tests

### GroupExportService Spec

**Location:** `/spec/services/group_export_service_spec.rb`

This is an integration test that verifies the complete export/import round-trip.

#### Test Setup: create_scenario

Creates a comprehensive test dataset including:

**Users:**
- `admin@example.com` - Group admin
- `member@example.com` - Group member
- `another_user@example.com` - Unrelated user (should NOT be exported)

**Groups:**
- `group` - Main group with admin as creator
- `subgroup` - Child of main group
- `another_group` - Unrelated group (should NOT be exported)

**Memberships:**
- Admin and member added to both group and subgroup
- Admin role in both groups

**Templates:**
- Discussion template attached to group
- Poll template attached to group

**Tags:**
- Tag with name and color attached to group

**Discussions:**
- Discussion in main group with tag
- Discussion in subgroup

**Comments:**
- Comments in both discussions

**Polls:**
- Poll in main group
- Poll in subgroup
- Both with poll options (agree/disagree)

**Stances/Votes:**
- Admin and member vote in main poll
- Admin and member added as voters (but no vote cast) in subgroup poll
- Polls closed to generate stance receipts

**Events:**
- Discussion created event
- Comment event

**Discussion Readers:**
- Reader tracking for admin

**Notifications:**
- Notification created from discussion event

**Reactions:**
- Reactions on discussion, poll, and comment

#### Test Flow: 'export, truncate, import'

1. **Setup:** Create complete test scenario
2. **Export:** Call `GroupExportService.export(group.all_groups, group.name)`
3. **Truncate:** Clear all relevant database tables
4. **Import:** Call `GroupExportService.import(filename)`
5. **Verify:** Assert all data was correctly restored

#### Verification Assertions

**Users:**
- Admin and member exist with correct emails
- Unrelated user does NOT exist

**Groups:**
- Main group and subgroup exist
- Unrelated group does NOT exist

**Memberships:**
- All 4 memberships restored (admin in both, member in both)
- Admin roles preserved

**Discussions:**
- Both discussions restored with correct associations
- Tags preserved on tagged discussion

**Tags:**
- Tag restored with name, group, and color

**Comments:**
- Both comments restored with correct associations

**Polls:**
- Both polls restored
- Stance counts accurate after update_counts!

**Stances:**
- All 4 stances restored
- Correct participant associations

**Stance Receipts:**
- All 4 receipts restored
- vote_cast flag correctly preserved (true for cast votes, false for uncast)

**Templates:**
- Discussion template restored
- Poll template restored

**Reactions:**
- All 3 reactions restored with correct associations

#### Helper Methods

**truncate_tables:**
Clears all relevant tables to ensure clean import test:
- StanceReceipt, Group, Membership, User, Discussion, DiscussionTemplate
- Comment, Poll, PollTemplate, PollOption, Stance, StanceChoice
- Reaction, Event, Notification, Document, DiscussionReader, Tag

---

## 2. Test Coverage Analysis

### Well Covered

| Area | Coverage |
|------|----------|
| JSON export file generation | Implicit via round-trip |
| JSON import with ID remapping | Explicit via round-trip |
| User data export | Verified |
| Group/subgroup relationships | Verified |
| Membership preservation | Verified |
| Poll data with options | Verified |
| Stance/vote data | Verified |
| Reaction data | Verified |
| Tag preservation | Verified |
| Template preservation | Verified |
| Notification preservation | Implicit |
| Event preservation | Implicit |

### Gaps in Coverage

| Area | Status |
|------|--------|
| CSV export format | No tests |
| HTML export rendering | No tests |
| Poll-specific export (PollExporter) | No tests |
| Controller authorization | No tests |
| Worker job execution | No tests |
| Email delivery | No tests |
| Attachment handling | No tests |
| Anonymous poll privacy | No explicit test |
| Sensitive field exclusion | No explicit test |
| Large dataset performance | No tests |
| Import conflict handling | Partial (user email conflict) |
| reset_keys option | No tests |

---

## 3. Frontend E2E Tests

### Export-Related Files Searched

The Nightwatch E2E tests in `/vue/tests/e2e/specs/` were searched for export-related tests.

**Result:** No dedicated export tests found.

The word "export" appears in test files only as part of JavaScript module syntax (`module.exports = {`), not as test content.

### Missing E2E Coverage

- Group export modal interaction
- CSV download initiation
- JSON export initiation
- HTML export page access
- Poll CSV download
- Authorization rejection scenarios

---

## 4. Recommended Additional Tests

### Service Layer

```
Pseudo-tests:

Test: "excludes sensitive user fields from export"
  - Create user with password, tokens
  - Export group containing user
  - Parse export file
  - Assert password/tokens not present

Test: "excludes group token from export"
  - Create group with invitation token
  - Export group
  - Assert token field not in group record

Test: "only exports non-anonymous OR closed polls"
  - Create open anonymous poll
  - Create open non-anonymous poll
  - Create closed anonymous poll
  - Export group
  - Assert only 2 polls exported (non-anon + closed)

Test: "handles large datasets with batching"
  - Create group with 100k comments
  - Export should complete without memory issues
  - Measure memory usage stays bounded
```

### Controller Layer

```
Pseudo-tests:

Test: "export requires admin permission"
  - Sign in as regular member
  - POST to export endpoint
  - Assert 403 response

Test: "export_csv requires admin permission"
  - Sign in as regular member
  - POST to export_csv endpoint
  - Assert 403 response

Test: "poll export requires results visibility"
  - Create poll with hide_results_until_closed
  - Try to export before closing
  - Assert permission denied
```

### Worker Layer

```
Pseudo-tests:

Test: "GroupExportWorker creates document and sends email"
  - Enqueue worker
  - Assert Document created
  - Assert UserMailer called
  - Assert DestroyRecordWorker scheduled for 1 week

Test: "GroupExportCsvWorker creates CSV document"
  - Enqueue worker
  - Assert Document with .csv extension created
  - Assert CSV content is valid
```

---

## 5. Test Data Factories

The test uses inline record creation rather than factories. Relevant FactoryBot factories would include:

- `:user`
- `:group`
- `:membership`
- `:discussion`
- `:discussion_template`
- `:comment`
- `:poll`
- `:poll_template`
- `:poll_option`
- `:stance`
- `:event`
- `:notification`
- `:reaction`
- `:tag`
- `:discussion_reader`

---

## 6. Running Export Tests

### Backend Tests

```
Run single export test file:
  bundle exec rspec spec/services/group_export_service_spec.rb

Run with verbose output:
  bundle exec rspec spec/services/group_export_service_spec.rb --format documentation
```

### Note on Test Duration

The export/import test is relatively slow because it:
- Creates extensive test data
- Writes to filesystem
- Truncates tables
- Reimports all data
- Validates relationships

Consider this when running full test suite.

---

## 7. Import Conflict Handling

The spec has a commented-out test case:

```
# it "import on existing tables" do
# end
```

This suggests import behavior with pre-existing data may not be fully tested or specified.

Current behavior for conflicts:
- Users: Looks up existing user by email, remaps to existing ID
- Other records: Raises error on unique constraint violation

---

## Confidence Rating: 4/5

The existing test is comprehensive for the JSON export/import cycle but leaves gaps in CSV export, HTML export, controller authorization, and edge cases like anonymous polls. The test methodology (round-trip verification) is sound for validating data integrity.
