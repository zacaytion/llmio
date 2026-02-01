# Search Domain - QA Confidence Report

**Reviewed:** 2026-02-01
**Reviewer:** QA Agent

---

## 1. Checklist Results

### Models Documentation (`models.md`)

| Item | Status | Notes |
|------|--------|-------|
| SearchResult model location | PASS | Correctly documented at `/app/models/search_result.rb` |
| SearchResult attributes | PASS | All 18 attributes verified against actual model |
| SearchResult accessor methods | PASS | `poll` and `author` methods confirmed |
| Searchable concern location | PASS | `/app/models/concerns/searchable.rb` verified |
| Searchable behavior | PASS | `multisearchable` call and `PgSearch::Model` inclusion confirmed |
| `update_pg_search_document` override | PASS | Override in `PgSearch::Multisearchable` module confirmed |
| Five searchable models | PASS | Discussion, Comment, Poll, Stance, Outcome all include Searchable |
| Discussion indexed fields | PASS | title, description, author name confirmed in SQL |
| Stance privacy filtering | PASS | Anonymous and hide_results conditions verified in SQL |
| pg_search_documents schema | PASS | All columns verified against migration |
| pg_search_documents indexes | PARTIAL | 6 indexes documented, migration shows 7 indexes but poll_id index is noted as `references` (auto-indexed) |
| pg_search configuration | PARTIAL | Documentation mentions 'simple' dictionary but actual initializer does not explicitly set dictionary - it uses pg_search defaults with tsvector_column |

### Services Documentation (`services.md`)

| Item | Status | Notes |
|------|--------|-------|
| SearchService location | PASS | `/app/services/search_service.rb` verified |
| `reindex_everything` method | PASS | All 5 model statements executed, verified |
| `reindex_by_author_id` method | PASS | Delete and insert for all 5 models verified |
| `reindex_by_discussion_id` method | PASS | Delete and insert verified, uses `id:` param for Discussion |
| `reindex_by_poll_id` method | PASS | Delete and insert for Poll, Stance, Outcome verified |
| `reindex_by_comment_id` method | PASS | Method exists, documented as unused (accurate) |
| Trigger points table | PASS | All triggers verified against codebase grep results |
| GenericWorker usage | PASS | All async calls confirmed in service files and workers |

### Controllers Documentation (`controllers.md`)

| Item | Status | Notes |
|------|--------|-------|
| SearchController location | PASS | `/app/controllers/api/v1/search_controller.rb` verified |
| API endpoint | PASS | `GET /api/v1/search` confirmed |
| Request parameters | PASS | query, group_id, org_id, type, order, tag all verified |
| Response structure | PASS | search_results array with documented fields confirmed |
| Visibility mode 1 (group_id=0) | PASS | Direct discussions via `guest_discussion_ids` confirmed |
| Visibility mode 2 (group/org specified) | PASS | `browseable_group_ids` intersection confirmed |
| Visibility mode 3 (no filter) | PASS | Both `browseable_group_ids` and `guest_discussion_ids` confirmed |
| `browseable_group_ids` definition | PASS | Verified includes parent groups and visible subgroups |
| Tag filtering logic | PASS | Discussion and Poll tag queries confirmed |
| Type filtering | PASS | 5 types validated, where clause confirmed |
| Order options | PASS | authored_at_desc and authored_at_asc confirmed |
| Result limit | PASS | `limit(20)` confirmed |
| Event loading for navigation | PASS | poll_events and stance_events queries verified |
| SearchResultSerializer | PASS | All attributes and associations confirmed |
| Excluded types | PASS | 5 excluded types confirmed |

### Frontend Documentation (`frontend.md`)

| Item | Status | Notes |
|------|--------|-------|
| SearchModal location | PASS | `/vue/src/components/search/modal.vue` verified |
| Component props | PASS | All 4 props with correct types verified |
| Filter controls | PASS | 5 selectors (org, subgroup, tag, type, order) verified |
| Type items | PASS | 6 items including "All content" confirmed |
| Order items | PASS | 3 items (Best match, Newest, Oldest) confirmed |
| Data flow/fetch method | PASS | Records.remote.get('search', params) confirmed |
| Result navigation URLs | PASS | urlForResult() logic for all content types verified |
| Entry points | PASS | navbar, discussions_panel, polls_panel verified |
| Modal launcher integration | PASS | SearchModal registered in launcher.vue |
| Route path watcher | PASS | `'$route.path': 'closeModal'` confirmed |
| Tag integration | PASS | `updateTagItems()` method and `tagsByName()` call verified |

### Tests Documentation (`tests.md`)

| Item | Status | Notes |
|------|--------|-------|
| Controller spec location | PASS | `/spec/controllers/api/v1/search_controller_spec.rb` verified |
| Test actors | PASS | user, group, visible_subgroup, other_group confirmed |
| Test data setup | PASS | All content hierarchies verified including anonymous/hidden polls |
| "returns any visible records" test | PASS | Verified with correct type counts |
| "returns group records" test | PASS | Verified |
| "does not return other group records" test | PASS | Verified |
| "returns invite-only records" test | PASS | Verified with group_id=0 |
| Commented-out tests | PASS | Both commented tests accurately documented |
| E2E test status | PASS | Correctly notes no dedicated search E2E tests |

---

## 2. Confidence Scores

| Document | Score | Assessment |
|----------|-------|------------|
| models.md | 5/5 | Excellent accuracy, all claims verified |
| services.md | 5/5 | Excellent accuracy, all triggers confirmed |
| controllers.md | 5/5 | Excellent accuracy, visibility logic correct |
| frontend.md | 5/5 | Excellent accuracy, all components verified |
| tests.md | 4/5 | Good accuracy, minor gaps in missing test coverage analysis |

**Overall Domain Confidence: 5/5**

---

## 3. Issues Found

### Minor Issues

1. **pg_search dictionary configuration (models.md)**
   - Documentation states: "Uses 'simple' dictionary (no stemming, no stop words)"
   - Reality: The initializer (`/config/initializers/pg_search.rb`) does not explicitly set a dictionary. The 'simple' dictionary usage is specified directly in the SQL insert statements via `to_tsvector('simple', ...)`.
   - Impact: Low - the documentation is functionally correct, but the mechanism is slightly different than implied.

2. **poll_id index (models.md)**
   - Documentation lists `index_pg_search_documents_on_poll_id` as a named index
   - Reality: The migration uses `t.references :poll` without explicit index option, meaning the index exists but with Rails-generated name
   - Impact: Negligible - documentation accurately conveys that the column is indexed

3. **Test count discrepancy note (tests.md)**
   - Documentation states "3 Polls" in "returns group records" test should be noted that it includes anonymous_poll and hidden_open_poll which are in the same group
   - Impact: Low - the documentation is accurate but could be clearer about why 3 polls exist

### No Critical Issues Found

All core functionality is accurately documented. The search implementation follows the patterns described in the expert guide (service layer for mutations, query objects for visibility, serializers for API output).

---

## 4. Uncertainties

1. **Pagination behavior**: The controller limits to 20 results but does not implement offset/pagination. The commented code `# results = results.order().offset().limit()` suggests this may have been planned but not implemented. The frontend documentation does not mention pagination either. This appears to be a limitation, not a documentation error.

2. **Empty query handling**: Neither the controller nor frontend documentation explicitly addresses what happens when query parameter is empty or missing. The frontend shows `if (!this.query)` returns empty results, but the backend behavior is not documented.

3. **Concurrent reindex safety**: The services documentation mentions "no full-table locks" but does not address potential race conditions during the brief window between delete and insert. This is a minor operational concern not reflected in documentation.

4. **Author name search**: The documentation correctly notes author names are indexed, but does not clarify that searching for an author name will return all their content (not just discussions they authored) since all content types include author name in the indexed content.

---

## 5. Revision Recommendations

### Low Priority

1. **Clarify dictionary configuration**: Update models.md to specify that the 'simple' dictionary is set in the SQL insert statements rather than the pg_search initializer.

2. **Add pagination caveat**: Note in controllers.md that pagination is not currently implemented (limited to 20 results).

3. **Document empty query behavior**: Add a note about empty/missing query parameter handling in controllers.md.

### Optional Enhancements

1. **Add sequence diagram**: A visual showing the flow from frontend search to API to pg_search_documents to result construction would aid understanding.

2. **Add reindex timing information**: Include approximate timing for full reindex operations and per-discussion reindex to help operational planning.

3. **Cross-reference test gaps**: Link the "Suggested Additional Tests" section to specific documentation sections that would benefit from test coverage.

---

## Summary

The search domain documentation is of high quality with accurate technical details. All five documents (models, services, controllers, frontend, tests) correctly describe the implementation. The search system follows Loomio's architectural patterns consistently, using pg_search for full-text search with proper visibility filtering based on group membership and guest access.

Key strengths:
- Complete coverage of all search components
- Accurate SQL and code references
- Proper security model documentation (visibility filtering)
- Good test documentation including known gaps

The documentation is ready for use by developers with no critical issues requiring immediate correction.
