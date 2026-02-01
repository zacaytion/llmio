# Full-Text Search Indexing - Implementation Synthesis

## Executive Summary

Loomio uses PostgreSQL full-text search via the `pg_search` gem. Search documents are stored in `pg_search_documents` table and updated synchronously via ActiveRecord callbacks.

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

## Index Update Service Requirements

### Operations Needed

1. **UpdateDiscussion** - Updates search index for a discussion
2. **UpdateComment** - Updates search index for a comment
3. **DeleteDocument** - Removes a document from the search index
4. **RebuildAll** - Full reindex of all searchable content
5. **RebuildGroup** - Reindex all content within a group

### Update Logic

For each update:
1. Delete existing document (WHERE searchable_type = X AND searchable_id = Y)
2. Insert new document with refreshed content
3. Use `to_tsvector('simple', content)` for the tsvector column

---

## Search Query Service

### Parameters

| Parameter | Type | Purpose |
|-----------|------|---------|
| Query | string | Search terms |
| GroupID | int64 (optional) | Filter by group |
| DiscussionID | int64 (optional) | Filter by discussion |
| AuthorID | int64 (optional) | Filter by author |
| Limit | int | Max results (default 20) |
| Offset | int | Pagination offset |

### Query Pattern

```sql
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
```

---

## Integration Points

### Service Integration

Index updates should happen:
1. After creating a searchable record
2. After updating a searchable record
3. After discarding/deleting a searchable record (remove from index)

### Error Handling

Search indexing is non-critical - failures should be logged but not fail the main operation.

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
