# Full-Text Search Indexing - Implementation Synthesis

## Executive Summary

Loomio uses PostgreSQL full-text search via the `pg_search` gem. Search documents are stored in `pg_search_documents` table and updated synchronously via ActiveRecord callbacks. Go must implement equivalent index maintenance and query functionality.

---

## Confirmed Architecture

### pg_search_documents Schema

```sql
CREATE TABLE pg_search_documents (
    id bigserial PRIMARY KEY,
    content text,                    -- Plain text content
    ts_content tsvector,             -- Full-text search vector
    searchable_type varchar,         -- 'Discussion', 'Comment', 'Poll', etc.
    searchable_id bigint,            -- ID of the searchable record
    group_id bigint,                 -- Filter: search within group
    discussion_id bigint,            -- Filter: search within discussion
    poll_id bigint,                  -- Filter: search within poll
    author_id bigint,                -- Filter: search by author
    authored_at timestamp,           -- Sort by recency
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL
);

-- GIN index for fast full-text search
CREATE INDEX index_pg_search_documents_on_ts_content
    ON pg_search_documents USING gin(ts_content);

-- Filter indexes
CREATE INDEX index_pg_search_documents_on_group_id
    ON pg_search_documents(group_id);
CREATE INDEX index_pg_search_documents_on_author_id
    ON pg_search_documents(author_id);
CREATE INDEX index_pg_search_documents_on_discussion_id
    ON pg_search_documents(discussion_id);
```

### Searchable Models

| Model | Content Fields | Additional Filters |
|-------|----------------|-------------------|
| Discussion | title, description, author.name | group_id |
| Comment | body, author.name | group_id, discussion_id |
| Poll | title, details, author.name | group_id, discussion_id |
| Stance | reason, author.name | group_id, discussion_id, poll_id |
| Outcome | statement, author.name | group_id, discussion_id, poll_id |

---

## Index Population Mechanism

### Rails Implementation

From `app/models/concerns/searchable.rb`:

```ruby
module Searchable
  extend ActiveSupport::Concern
  include PgSearch::Model

  included do
    multisearchable  # Adds after_save callback
  end
end

module PgSearch::Multisearchable
  def update_pg_search_document
    PgSearch::Document.where(searchable: self).delete_all
    ActiveRecord::Base.connection.execute(self.class.pg_search_insert_statement(id: self.id))
  end
end
```

### Discussion INSERT Statement

From `app/models/discussion.rb`:

```ruby
def self.pg_search_insert_statement(id: nil, author_id: nil)
  content_str = "regexp_replace(CONCAT_WS(' ', discussions.title, discussions.description, users.name), E'<[^>]+>', '', 'gi')"
  <<~SQL.squish
    INSERT INTO pg_search_documents (
      searchable_type, searchable_id, group_id, discussion_id,
      author_id, authored_at, content, ts_content, created_at, updated_at
    )
    SELECT 'Discussion' AS searchable_type,
      discussions.id AS searchable_id,
      discussions.group_id as group_id,
      discussions.id AS discussion_id,
      discussions.author_id AS author_id,
      discussions.created_at AS authored_at,
      #{content_str} AS content,
      to_tsvector('simple', #{content_str}) as ts_content,
      now() AS created_at,
      now() AS updated_at
    FROM discussions
      LEFT JOIN users ON users.id = discussions.author_id
    WHERE discarded_at IS NULL
      #{id ? " AND discussions.id = #{id.to_i} LIMIT 1" : ''}
      #{author_id ? " AND discussions.author_id = #{author_id.to_i}" : ''}
  SQL
end
```

---

## Go Implementation

### Index Update Service

```go
package search

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"
)

type IndexService struct {
    db *pgxpool.Pool
}

func NewIndexService(db *pgxpool.Pool) *IndexService {
    return &IndexService{db: db}
}

// UpdateDiscussion updates the search index for a discussion
func (s *IndexService) UpdateDiscussion(ctx context.Context, discussionID int64) error {
    // Delete existing document
    _, err := s.db.Exec(ctx, `
        DELETE FROM pg_search_documents
        WHERE searchable_type = 'Discussion' AND searchable_id = $1
    `, discussionID)
    if err != nil {
        return fmt.Errorf("delete old document: %w", err)
    }

    // Insert new document
    _, err = s.db.Exec(ctx, `
        INSERT INTO pg_search_documents (
            searchable_type, searchable_id, group_id, discussion_id,
            author_id, authored_at, content, ts_content, created_at, updated_at
        )
        SELECT
            'Discussion',
            d.id,
            d.group_id,
            d.id,
            d.author_id,
            d.created_at,
            regexp_replace(concat_ws(' ', d.title, d.description, u.name), E'<[^>]+>', '', 'gi'),
            to_tsvector('simple', regexp_replace(concat_ws(' ', d.title, d.description, u.name), E'<[^>]+>', '', 'gi')),
            now(),
            now()
        FROM discussions d
        LEFT JOIN users u ON u.id = d.author_id
        WHERE d.id = $1 AND d.discarded_at IS NULL
    `, discussionID)
    if err != nil {
        return fmt.Errorf("insert document: %w", err)
    }

    return nil
}

// UpdateComment updates the search index for a comment
func (s *IndexService) UpdateComment(ctx context.Context, commentID int64) error {
    _, err := s.db.Exec(ctx, `
        DELETE FROM pg_search_documents
        WHERE searchable_type = 'Comment' AND searchable_id = $1
    `, commentID)
    if err != nil {
        return fmt.Errorf("delete old document: %w", err)
    }

    _, err = s.db.Exec(ctx, `
        INSERT INTO pg_search_documents (
            searchable_type, searchable_id, group_id, discussion_id,
            author_id, authored_at, content, ts_content, created_at, updated_at
        )
        SELECT
            'Comment',
            c.id,
            c.group_id,
            c.discussion_id,
            c.author_id,
            c.created_at,
            regexp_replace(concat_ws(' ', c.body, u.name), E'<[^>]+>', '', 'gi'),
            to_tsvector('simple', regexp_replace(concat_ws(' ', c.body, u.name), E'<[^>]+>', '', 'gi')),
            now(),
            now()
        FROM comments c
        LEFT JOIN users u ON u.id = c.author_id
        WHERE c.id = $1 AND c.discarded_at IS NULL
    `, commentID)

    return err
}

// DeleteDocument removes a document from the search index
func (s *IndexService) DeleteDocument(ctx context.Context, searchableType string, searchableID int64) error {
    _, err := s.db.Exec(ctx, `
        DELETE FROM pg_search_documents
        WHERE searchable_type = $1 AND searchable_id = $2
    `, searchableType, searchableID)
    return err
}
```

### Search Query Service

```go
package search

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"
)

type QueryService struct {
    db *pgxpool.Pool
}

type SearchParams struct {
    Query        string
    GroupID      *int64
    DiscussionID *int64
    AuthorID     *int64
    Limit        int
    Offset       int
}

type SearchResult struct {
    SearchableType string
    SearchableID   int64
    Rank           float64
    Content        string
    AuthoredAt     time.Time
}

func (s *QueryService) Search(ctx context.Context, params SearchParams) ([]SearchResult, error) {
    if params.Limit == 0 {
        params.Limit = 20
    }

    rows, err := s.db.Query(ctx, `
        SELECT
            searchable_type,
            searchable_id,
            ts_rank(ts_content, plainto_tsquery('simple', $1)) AS rank,
            content,
            authored_at
        FROM pg_search_documents
        WHERE ts_content @@ plainto_tsquery('simple', $1)
          AND ($2::bigint IS NULL OR group_id = $2)
          AND ($3::bigint IS NULL OR discussion_id = $3)
          AND ($4::bigint IS NULL OR author_id = $4)
        ORDER BY rank DESC, authored_at DESC
        LIMIT $5 OFFSET $6
    `, params.Query, params.GroupID, params.DiscussionID, params.AuthorID, params.Limit, params.Offset)
    if err != nil {
        return nil, fmt.Errorf("search query: %w", err)
    }
    defer rows.Close()

    var results []SearchResult
    for rows.Next() {
        var r SearchResult
        if err := rows.Scan(&r.SearchableType, &r.SearchableID, &r.Rank, &r.Content, &r.AuthoredAt); err != nil {
            return nil, fmt.Errorf("scan result: %w", err)
        }
        results = append(results, r)
    }

    return results, nil
}
```

### Integration with Services

```go
// In discussion_service.go
func (s *DiscussionService) Create(ctx context.Context, discussion *Discussion, actor *User) error {
    // ... create discussion ...

    // Update search index (synchronous, matching Rails)
    if err := s.searchIndex.UpdateDiscussion(ctx, discussion.ID); err != nil {
        slog.Error("failed to update search index", "discussion_id", discussion.ID, "error", err)
        // Don't fail the request - search is non-critical
    }

    return nil
}

func (s *DiscussionService) Update(ctx context.Context, discussion *Discussion, actor *User) error {
    // ... update discussion ...

    // Re-index after update
    if err := s.searchIndex.UpdateDiscussion(ctx, discussion.ID); err != nil {
        slog.Error("failed to update search index", "discussion_id", discussion.ID, "error", err)
    }

    return nil
}

func (s *DiscussionService) Discard(ctx context.Context, discussion *Discussion, actor *User) error {
    // ... soft delete discussion ...

    // Remove from search index
    if err := s.searchIndex.DeleteDocument(ctx, "Discussion", discussion.ID); err != nil {
        slog.Error("failed to remove from search index", "discussion_id", discussion.ID, "error", err)
    }

    return nil
}
```

---

## Bulk Reindexing

### Rebuild All Documents

```go
func (s *IndexService) RebuildAll(ctx context.Context) error {
    // Clear all documents
    _, err := s.db.Exec(ctx, "TRUNCATE pg_search_documents")
    if err != nil {
        return fmt.Errorf("truncate: %w", err)
    }

    // Rebuild discussions
    _, err = s.db.Exec(ctx, `
        INSERT INTO pg_search_documents (
            searchable_type, searchable_id, group_id, discussion_id,
            author_id, authored_at, content, ts_content, created_at, updated_at
        )
        SELECT
            'Discussion',
            d.id,
            d.group_id,
            d.id,
            d.author_id,
            d.created_at,
            regexp_replace(concat_ws(' ', d.title, d.description, u.name), E'<[^>]+>', '', 'gi'),
            to_tsvector('simple', regexp_replace(concat_ws(' ', d.title, d.description, u.name), E'<[^>]+>', '', 'gi')),
            now(),
            now()
        FROM discussions d
        LEFT JOIN users u ON u.id = d.author_id
        WHERE d.discarded_at IS NULL
    `)
    if err != nil {
        return fmt.Errorf("rebuild discussions: %w", err)
    }

    // Similarly for comments, polls, stances, outcomes...
    return nil
}

func (s *IndexService) RebuildGroup(ctx context.Context, groupID int64) error {
    // Delete all documents for this group
    _, err := s.db.Exec(ctx, "DELETE FROM pg_search_documents WHERE group_id = $1", groupID)
    if err != nil {
        return err
    }

    // Rebuild all searchable records for this group
    // ...
    return nil
}
```

---

## sqlc Queries

### queries/search.sql

```sql
-- name: SearchDocuments :many
SELECT
    searchable_type,
    searchable_id,
    ts_rank(ts_content, plainto_tsquery('simple', @query)) AS rank,
    content,
    authored_at
FROM pg_search_documents
WHERE ts_content @@ plainto_tsquery('simple', @query)
  AND (sqlc.narg('group_id')::bigint IS NULL OR group_id = sqlc.narg('group_id'))
  AND (sqlc.narg('discussion_id')::bigint IS NULL OR discussion_id = sqlc.narg('discussion_id'))
  AND (sqlc.narg('author_id')::bigint IS NULL OR author_id = sqlc.narg('author_id'))
ORDER BY rank DESC, authored_at DESC
LIMIT @limit_count OFFSET @offset_count;

-- name: DeleteSearchDocument :exec
DELETE FROM pg_search_documents
WHERE searchable_type = @searchable_type AND searchable_id = @searchable_id;

-- name: CountSearchResults :one
SELECT COUNT(*)
FROM pg_search_documents
WHERE ts_content @@ plainto_tsquery('simple', @query)
  AND (sqlc.narg('group_id')::bigint IS NULL OR group_id = sqlc.narg('group_id'));
```

---

## API Endpoint

### Search Controller

```go
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("q")
    if query == "" {
        http.Error(w, "query required", http.StatusBadRequest)
        return
    }

    params := search.SearchParams{
        Query: query,
        Limit: 20,
    }

    // Parse optional filters
    if groupID := r.URL.Query().Get("group_id"); groupID != "" {
        id, _ := strconv.ParseInt(groupID, 10, 64)
        params.GroupID = &id
    }

    // Check user has access to group
    if params.GroupID != nil && !h.ability.CanViewGroup(r.Context(), *params.GroupID) {
        http.Error(w, "forbidden", http.StatusForbidden)
        return
    }

    results, err := h.searchService.Search(r.Context(), params)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Load full records
    response := h.hydrateResults(r.Context(), results)
    json.NewEncoder(w).Encode(response)
}
```

---

## Configuration

### Text Search Configuration

Using `'simple'` configuration means:
- **No stemming** - "running" won't match "run"
- **No stop words** - "the", "and" are indexed
- **Case insensitive** - "Hello" matches "hello"

For better English search, consider:
```sql
to_tsvector('english', content)  -- Uses English stemmer
```

### Performance Tuning

```sql
-- Increase work_mem for large searches
SET work_mem = '256MB';

-- Analyze table for query planner
ANALYZE pg_search_documents;
```

---

## Testing

```go
func TestSearch(t *testing.T) {
    ctx := context.Background()

    // Create test discussion
    discussion := &Discussion{
        Title:       "Test Discussion",
        Description: "This is about testing search functionality",
        GroupID:     1,
        AuthorID:    1,
    }
    discussionService.Create(ctx, discussion, testUser)

    // Search should find it
    results, err := searchService.Search(ctx, search.SearchParams{
        Query:   "testing search",
        GroupID: ptr(int64(1)),
    })
    require.NoError(t, err)
    require.Len(t, results, 1)
    assert.Equal(t, "Discussion", results[0].SearchableType)
    assert.Equal(t, discussion.ID, results[0].SearchableID)
}
```
