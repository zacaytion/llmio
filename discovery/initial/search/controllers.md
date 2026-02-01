# Search Domain - Controllers

**Generated:** 2026-02-01
**Confidence:** 5/5

---

## SearchController

**Location:** `/app/controllers/api/v1/search_controller.rb`

The SearchController handles full-text search requests from the frontend. It extends RestfulController but only implements the index action for search queries.

---

## API Contract

### Endpoint

```
GET /api/v1/search
```

### Request Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| query | string | Yes | Search query text |
| group_id | integer | No | Limit to specific group |
| org_id | integer | No | Limit to organization and subgroups |
| type | string | No | Filter by content type (Discussion, Comment, Poll, Stance, Outcome) |
| order | string | No | Sort order: authored_at_desc, authored_at_asc, or null for relevance |
| tag | string | No | Filter by tag name |

### Response Structure

```json
{
  "search_results": [
    {
      "id": 123,
      "searchable_type": "Discussion",
      "searchable_id": 456,
      "poll_title": null,
      "discussion_title": "Project Planning",
      "discussion_key": "abc123",
      "highlight": "The <b>meeting</b> is scheduled for...",
      "poll_key": null,
      "poll_id": null,
      "sequence_id": null,
      "group_id": 1,
      "group_handle": "my-group",
      "group_key": "xyz789",
      "group_name": "My Group",
      "author_name": "Jane Smith",
      "author_id": 42,
      "authored_at": "2026-01-15T10:30:00Z",
      "tags": ["planning", "quarterly"]
    }
  ],
  "users": [...],
  "polls": [...]
}
```

---

## Visibility Filtering

The controller implements security-first visibility filtering to ensure users only see content they have access to.

### Scope Determination Logic

Three visibility modes based on group_id and org_id parameters:

**Mode 1: Direct Discussions Only (group_id = 0 or org_id = 0)**
- Searches only discussions where user is a guest
- Uses user's guest_discussion_ids
- Matches documents where group_id is null AND discussion_id is in guest list

**Mode 2: Specific Group/Org (group_id or org_id provided)**
- Searches within specified group hierarchy
- group_ids calculated as intersection of:
  - User's browseable_group_ids (groups they can see)
  - Requested group(s) scope
- For org_id: includes parent and all subgroups
- For group_id: includes only that specific group

**Mode 3: All Content (no group/org specified)**
- Searches all content user can access
- Combines:
  - All user's browseable_group_ids
  - All user's guest_discussion_ids

### browseable_group_ids

Defined on User model, returns IDs of:
- Groups where user is a member
- Subgroups of member groups that are visible to parent members

This ensures users can search subgroup content they have access to even without direct membership.

---

## Query Processing

### Full-Text Search

Uses PgSearch.multisearch(query) which:
1. Parses the query for terms and operators
2. Searches the ts_content tsvector column
3. Applies prefix matching (partial words)
4. Supports negation with minus prefix

### Tag Filtering

When tag parameter is provided:
1. Finds all discussion IDs in accessible groups with matching tag
2. Finds all poll IDs in accessible groups with matching tag
3. Filters search results to those matching either set

### Type Filtering

Accepts one of: Discussion, Comment, Poll, Stance, Outcome
Filters pg_search_documents by searchable_type

### Ordering Options

- `null` (default): PostgreSQL full-text relevance ranking (ts_rank)
- `authored_at_desc`: Newest content first
- `authored_at_asc`: Oldest content first

---

## Result Construction

After querying pg_search_documents, the controller:

1. **Limits results:** Maximum 20 results per query
2. **Loads highlights:** Uses with_pg_search_highlight to get snippets
3. **Batch loads related records:**
   - Groups by group_id
   - Discussions by discussion_id
   - Polls by poll_id
   - Authors (Users) by author_id
4. **Loads navigation events:**
   - Poll events for linking to poll position in discussion
   - Stance events for linking to stance position in discussion
5. **Constructs SearchResult objects** with all denormalized data

### Event Loading for Navigation

For search results that are part of discussions, the controller loads the related Event records to find sequence_id. This allows the frontend to navigate directly to the item's position in the discussion timeline.

---

## Serialization

### SearchResultSerializer

**Location:** `/app/serializers/search_result_serializer.rb`

Serializes all SearchResult attributes plus associated records:
- `author` - Uses AuthorSerializer, rooted under "users"
- `poll` - Uses PollSerializer, rooted under "polls"

### Excluded Types

The controller excludes these types from related record serialization:
- group
- membership
- discussion
- outcome
- event

This reduces response payload by not embedding full records that aren't needed for display.
