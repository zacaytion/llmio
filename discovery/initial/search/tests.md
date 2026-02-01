# Search Domain - Tests

**Generated:** 2026-02-01
**Confidence:** 4/5

---

## Controller Tests

**Location:** `/spec/controllers/api/v1/search_controller_spec.rb`

The search controller has comprehensive RSpec tests covering visibility and access control scenarios.

---

## Test Setup

### Actors

- **user** - Normal user who will perform searches
- **group** - Primary group where user has membership
- **visible_subgroup** - Subgroup visible to parent members
- **other_group** - Group where user has no membership

### Test Data

Creates complete content hierarchies for testing visibility:

**In user's group:**
- discussion, comment, poll, stance, outcome (all findable)
- anonymous_poll with anonymous_stance (hidden while open)
- hidden_open_poll with hidden_open_stance (hidden while open)

**Discarded content:**
- discarded_discussion with discarded_comment, poll, stance, outcome (all excluded)

**Direct discussions (no group):**
- io_discussion (invite-only) with io_comment, io_poll, io_stance, io_outcome
- User is added as guest to this discussion

**Guest access to other group:**
- guest_discussion in other_group with guest_comment, poll, stance, outcome
- User is added as guest to this discussion

**Inaccessible content:**
- other_discussion in other_group with other_comment, poll, stance, outcome
- User has no access to this content

---

## Test Cases

### "returns any visible records"

Searches with query "findme" and no group filter.

**Verifies:**
- Returns discussion from user's group
- Returns discussion from visible subgroup
- Returns direct discussion (io_discussion)
- Returns guest discussion (guest_discussion)
- Correct counts by type: 4 Discussions, 3 Comments, 5 Polls, 3 Stances, 3 Outcomes

**Implicitly verifies:**
- Does not return discarded content
- Does not return other_group content user cannot see
- Stances from anonymous/hidden polls are not counted (explains why only 3 stances)

### "returns group records"

Searches with query "findme" filtered to user's group.

**Verifies:**
- Returns the group's discussion, comment, poll, stance, outcome
- Does not return content from other scopes
- Correct counts: 1 Discussion, 1 Comment, 3 Polls (includes anonymous/hidden), 1 Stance, 1 Outcome

### "does not return other group records"

Searches with query "findme" filtered to other_group.

**Verifies:**
- Returns zero results for each content type
- Confirms user cannot search groups they're not a member of
- Even filtering to a specific group_id does not bypass access control

### "returns invite-only records"

Searches with group_id=0 (direct discussions only).

**Verifies:**
- Returns io_discussion and all its nested content
- Only returns content from discussions where user is a guest
- Correct counts: 1 of each type

---

## Test Patterns

### Factory Usage

Tests use FactoryBot factories:
- `:group` - Creates a group
- `:discussion` - Creates a discussion (requires group or can be nil)
- `:comment` - Creates a comment (requires discussion)
- `:poll` - Creates a poll (can have discussion)
- `:stance` - Creates a stance/vote (requires poll, cast_at, choice)
- `:outcome` - Creates an outcome (requires poll)

### Search Indexing

All test records are created with `let!` (eager loading) to ensure:
1. Records exist before tests run
2. pg_search documents are created via multisearchable callbacks

### Response Parsing

Tests parse JSON response and filter results:

```pseudo
results = JSON.parse(response.body)['search_results']
results.filter { |r| r['searchable_type'] == 'Discussion' && r['searchable_id'] == discussion.id }
```

### Count Verification

Tests verify exact counts to ensure no unexpected records leak through:

```pseudo
type_counts = {}
results.each { |result| type_counts[result['searchable_type']] += 1 }
expect(type_counts['Discussion']).to eq 4
```

---

## Missing Test Coverage

The test file has commented-out test cases indicating planned but unimplemented tests:

```pseudo
# it 'returns guest discussions when group_id' do
# end

# it 'does not return hidden stances' do
# end
```

### Suggested Additional Tests

Based on the codebase analysis, these scenarios would benefit from explicit tests:

1. **Tag filtering** - Verify search with tag parameter
2. **Type filtering** - Verify search with type parameter
3. **Order parameter** - Verify authored_at_asc/desc sorting
4. **Visible subgroups** - Explicit test that visible_subgroup content is included
5. **Anonymous poll visibility after close** - Stances become searchable after poll closes
6. **Hidden results poll visibility after close** - Stances become searchable after poll closes
7. **No query parameter** - Error or empty response handling
8. **Highlight content** - Verify highlight attribute contains search terms

---

## E2E Test Status

**Confidence:** 3/5

No dedicated Nightwatch E2E tests found for search functionality. Search is mentioned in `vue/tests/e2e/specs/poll.js` but appears to be incidental rather than focused testing.

### Recommended E2E Scenarios

1. Open search modal from navbar
2. Enter search query and verify results appear
3. Click result and verify navigation
4. Filter by content type
5. Filter by group/organization
6. Filter by tag
7. Verify empty results message
8. Verify pagination/limit behavior
