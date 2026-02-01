# Full-Text Search Indexing - Follow-up Analysis

## Executive Summary

Full-text search using PostgreSQL `pg_search_documents` was identified in baseline research but **not included in the third-party follow-up investigation**. This document captures findings from source code verification and open questions.

---

## Source Code Verification

### Searchable Concern

Located at `app/models/concerns/searchable.rb`:

```ruby
module Searchable
  extend ActiveSupport::Concern
  include PgSearch::Model

  included do
    multisearchable
  end
end

module PgSearch::Multisearchable
  def update_pg_search_document
    PgSearch::Document.where(searchable: self).delete_all
    ActiveRecord::Base.connection.execute(self.class.pg_search_insert_statement(id: self.id))
  end
end
```

### Models Using Searchable

| Model | Include Location | INSERT Statement |
|-------|------------------|------------------|
| Discussion | `app/models/discussion.rb:16` | Lines 18-48 |
| Comment | `app/models/comment.rb:12` | Lines 12-42 |
| Poll | `app/models/poll.rb:16` | Lines 16-46 |
| Stance | `app/models/stance.rb:13` | Lines 13-43 |
| Outcome | `app/models/outcome.rb:13` | Lines 13-43 |

### pg_search_documents Schema

From `research/schema_dump.sql`:

```sql
CREATE TABLE pg_search_documents (
    id bigserial PRIMARY KEY,
    content text,
    ts_content tsvector,          -- Full-text search vector
    searchable_type varchar,      -- Polymorphic type
    searchable_id bigint,         -- Polymorphic ID
    group_id bigint,              -- Filter: search within group
    discussion_id bigint,         -- Filter: search within discussion
    poll_id bigint,               -- Filter: search within poll
    author_id bigint,             -- Filter: search by author
    authored_at timestamp,        -- Sort by relevance/recency
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL
);

CREATE INDEX index_pg_search_documents_on_ts_content
    ON pg_search_documents USING gin(ts_content);
```

### Index Population Mechanism

**Verified:** Uses `multisearchable` from `pg_search` gem which:
1. Adds after_save callback to `update_pg_search_document`
2. Deletes existing document, then re-inserts via custom SQL
3. Uses `to_tsvector('simple', content)` for indexing

---

## Open Questions for Third Party

### HIGH Priority

| # | Question | Impact | Investigation Target |
|---|----------|--------|---------------------|
| 1 | **What triggers bulk reindexing?** | Data migration | Is there a scheduled job? Rake task? |
| 2 | **Are search results filtered by visibility?** | Security | Does search respect discussion privacy? |
| 3 | **What happens to search index on soft delete?** | Data integrity | `discarded_at IS NULL` filter in INSERT |

### MEDIUM Priority

| # | Question | Impact | Investigation Target |
|---|----------|--------|---------------------|
| 4 | Is there a search ranking formula? | Relevance | Default `ts_rank()` or custom? |
| 5 | Are there stop words configured? | Search quality | `'simple'` config vs English |
| 6 | How does search handle non-ASCII content? | i18n | Unicode handling in tsvector |

### LOW Priority

| # | Question | Impact | Investigation Target |
|---|----------|--------|---------------------|
| 7 | Is search index size monitored? | Operations | Large groups may have slow search |
| 8 | Are there search analytics? | Product | Query logging for popular terms? |

---

## Confirmed Implementation Details

### Content Extraction

From `discussion.rb`:
```ruby
content_str = "regexp_replace(CONCAT_WS(' ', discussions.title, discussions.description, users.name), E'<[^>]+>', '', 'gi')"
```

**Pattern:**
- Concatenate searchable fields with space separator
- Strip HTML tags via regex
- Include author name for author-aware search

### Filter Support

Search documents include:
- `group_id` - Search within group
- `discussion_id` - Search within thread
- `poll_id` - Search within poll (stances/outcomes)
- `author_id` - Search by author

### Text Configuration

Uses `'simple'` stemmer, not language-specific:
```sql
to_tsvector('simple', #{content_str}) as ts_content
```

This means:
- No stemming (e.g., "running" won't match "run")
- No stop word removal
- Case-insensitive matching

---

## Discrepancies

### Missing: Search Query Implementation

**Not found in investigation:** How does the API expose search?

Expected endpoint: `/api/v1/search` or similar
Expected parameters: `q`, `group_id`, `author_id`

**Need to locate:** Search controller and query construction.

### Missing: Real-time Index Updates

**Question:** When a discussion is edited, is the search index updated synchronously or via background job?

From `searchable.rb`, it appears **synchronous** (in after_save callback), which could impact write latency.

---

## Files Requiring Investigation

| File | Purpose | Priority |
|------|---------|----------|
| `app/controllers/api/v1/search_controller.rb` | Search API | HIGH |
| `config/initializers/pg_search.rb` | Search config | MEDIUM |
| `lib/tasks/search.rake` | Reindex tasks | MEDIUM |
| `app/models/concerns/searchable.rb` | Indexing logic | VERIFIED |

---

## Priority Assessment

| Area | Priority | Blocking? |
|------|----------|-----------|
| Index population mechanism | HIGH | Yes - affects data sync |
| Search API implementation | HIGH | Yes - user-facing feature |
| Visibility filtering | HIGH | Yes - security concern |
| Bulk reindex strategy | MEDIUM | No - can add later |
