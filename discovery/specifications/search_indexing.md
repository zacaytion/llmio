# Search Indexing Analysis

This document details how pg_search full-text search indexing works in Loomio.

## 1. Reindex Triggers

**Confidence: HIGH**

Search index updates are triggered **manually via background jobs**, NOT via automatic ActiveRecord callbacks. The application overrides pg_search's default callback behavior.

### Custom Callback Override

The `Searchable` concern includes `multisearchable` (line 6 of `app/models/concerns/searchable.rb`), but the standard pg_search callback behavior is overridden:

```ruby
# app/models/concerns/searchable.rb:20-24
module PgSearch::Multisearchable
  def update_pg_search_document
    PgSearch::Document.where(searchable: self).delete_all
    ActiveRecord::Base.connection.execute(self.class.pg_search_insert_statement(id: self.id))
  end
end
```

This monkey-patch replaces pg_search's default `update_pg_search_document` method with a custom implementation that:
1. Deletes existing search documents for the record
2. Executes a raw SQL INSERT statement using the model's `pg_search_insert_statement`

### Service-Triggered Reindexing

Reindexing is explicitly triggered by service classes using `GenericWorker` (a Sidekiq worker) to run `SearchService` methods asynchronously:

| Trigger Location | When Called | Method |
|-----------------|-------------|--------|
| `app/services/discussion_service.rb:111` | Discussion discarded | `reindex_by_discussion_id` |
| `app/services/discussion_service.rb:143` | Discussion moved | `reindex_by_discussion_id` |
| `app/services/poll_service.rb:48` | Poll updated | `reindex_by_poll_id` |
| `app/services/poll_service.rb:303` | Poll closed | `reindex_by_poll_id` |
| `app/services/poll_service.rb:348` | Poll added to thread | `reindex_by_discussion_id` |
| `app/services/event_service.rb:9` | Event removed from thread | `reindex_by_discussion_id` |
| `app/services/user_service.rb:49` | User reactivated | `reindex_by_author_id` |
| `app/services/user_service.rb:77` | User name changed | `reindex_by_author_id` |
| `app/workers/move_comments_worker.rb:31-32` | Comments moved | `reindex_by_discussion_id` (sync) |
| `app/workers/redact_user_worker.rb:56` | User redacted | `reindex_by_author_id` (sync) |
| `app/workers/deactivate_user_worker.rb:17` | User deactivated | `reindex_by_author_id` (sync) |

**Note:** Most calls use `GenericWorker.perform_async()` for asynchronous execution, but worker classes call `SearchService` methods directly (synchronously).

### What Does NOT Trigger Reindexing

The following operations do NOT automatically trigger search reindexing:
- Creating new discussions (no reindex call in `DiscussionService.create`)
- Creating new comments (no reindex call in comment creation flow)
- Creating new polls (no reindex call in `PollService.create`)
- Creating/updating stances (no reindex call in stance creation)
- Creating outcomes (no reindex call in outcome creation)

**Implication:** Newly created content relies on the `multisearchable` callback from pg_search being invoked, but the overridden `update_pg_search_document` method must be called. This appears to be handled by pg_search's after_save/after_create hooks that call `update_pg_search_document`.

## 2. Full Reindex Mechanism

**Confidence: HIGH**

### Primary Method: SearchService.reindex_everything

Location: `app/services/search_service.rb:2-11`

```ruby
def self.reindex_everything
  [
    Discussion.pg_search_insert_statement,
    Comment.pg_search_insert_statement,
    Poll.pg_search_insert_statement,
    Stance.pg_search_insert_statement,
    Outcome.pg_search_insert_statement
  ].each do |statement|
    ActiveRecord::Base.connection.execute(statement)
  end
end
```

### How to Run Full Reindex

**Via Rails Console:**
```ruby
SearchService.reindex_everything
```

**Via Migration (used historically):**
The migration `db/migrate/20230819001215_reindex_everything.rb` demonstrates the pattern:
```ruby
if ENV['MIGRATE_DATA_ASYNC']
  GenericWorker.perform_async('SearchService', 'reindex_everything')
else
  SearchService.reindex_everything
end
```

**Via Background Job:**
```ruby
GenericWorker.perform_async('SearchService', 'reindex_everything')
```

### No Rake Task

There is **no dedicated rake task** for search reindexing. The `lib/tasks/loomio.rake` file contains various maintenance tasks but none for search.

### Partial Reindex Methods

The `SearchService` provides scoped reindex methods:

| Method | Purpose | Clears then rebuilds |
|--------|---------|---------------------|
| `reindex_by_author_id(author_id)` | Reindex all content by a specific author | All documents where `author_id` matches |
| `reindex_by_discussion_id(discussion_id)` | Reindex a discussion and its content | Discussion + comments + polls + stances + outcomes in that discussion |
| `reindex_by_poll_id(poll_id)` | Reindex a poll and its votes | Poll + stances + outcomes for that poll |
| `reindex_by_comment_id(comment_id)` | Reindex a single comment | Single comment document |

### Per-Model Rebuild

Each searchable model also has a class method for bulk rebuilding:

```ruby
# app/models/concerns/searchable.rb:10-12
def rebuild_pg_search_documents
  connection.execute pg_search_insert_statement
end
```

Usage example:
```ruby
Discussion.rebuild_pg_search_documents
Comment.rebuild_pg_search_documents
```

**Note:** These methods INSERT without first clearing existing documents, so they may create duplicates. Use `SearchService.reindex_everything` instead for a clean rebuild.

## 3. Searchable Models

**Confidence: HIGH**

Five models include the `Searchable` concern:

| Model | File | Content Indexed |
|-------|------|-----------------|
| Discussion | `app/models/discussion.rb:16` | title, description, author name |
| Comment | `app/models/comment.rb:10` | body, author name |
| Poll | `app/models/poll.rb:14` | title, details, author name |
| Stance | `app/models/stance.rb:8` | reason, participant name |
| Outcome | `app/models/outcome.rb:11` | statement, author name |

### Index Content Details

Each model defines `pg_search_insert_statement` which specifies:
- What text content to index (with HTML stripping)
- Associated metadata (group_id, discussion_id, poll_id, author_id, authored_at)
- Filter conditions (e.g., exclude discarded records, exclude anonymous/hidden stances)

Example from Discussion (`app/models/discussion.rb:18-48`):
```ruby
def self.pg_search_insert_statement(id: nil, author_id: nil)
  content_str = "regexp_replace(CONCAT_WS(' ', discussions.title, discussions.description, users.name), E'<[^>]+>', '', 'gi')"
  <<~SQL.squish
    INSERT INTO pg_search_documents (...)
    SELECT 'Discussion' AS searchable_type,
      discussions.id AS searchable_id,
      ...
      to_tsvector('simple', #{content_str}) as ts_content,
      ...
    FROM discussions
      LEFT JOIN users ON users.id = discussions.author_id
    WHERE discarded_at IS NULL
      #{id ? " AND discussions.id = #{id.to_i} LIMIT 1" : ''}
  SQL
end
```

### Special Filtering Rules

**Stances** (`app/models/stance.rb:42-45`):
- Only indexed if `cast_at IS NOT NULL` (vote was cast)
- Excluded if poll is anonymous and still open
- Excluded if poll has `hide_results = 2` (until_closed) and still open

## 4. Search Query Patterns

**Confidence: HIGH**

### Search Controller

Location: `app/controllers/api/v1/search_controller.rb`

### Query Execution

Uses pg_search's `multisearch` method with the query:
```ruby
PgSearch.multisearch(params[:query])
```

This searches against the `ts_content` tsvector column in `pg_search_documents`.

### pg_search Configuration

Location: `config/initializers/pg_search.rb`

```ruby
PgSearch.multisearch_options = {
  using: {
    tsearch: {
      prefix: true,         # Enables prefix matching (e.g., "meet" matches "meeting")
      negation: true,       # Enables NOT operator
      tsvector_column: 'ts_content',
      highlight: {
        StartSel: '<b>',
        StopSel: '</b>',
      }
    },
  }
}
```

### Text Configuration

Uses PostgreSQL's `'simple'` text search configuration (not language-specific):
```sql
to_tsvector('simple', content)
```

This means:
- No stemming (exact word matching)
- No stop word removal
- Case-insensitive
- Works for any language

### Access Control Filtering

The search controller applies access control before returning results:

1. **For guests** (group_id = 0): Only search discussions where user is a guest
2. **For group members**: Filter by `group_id IN (user's browseable groups)`
3. **Global search**: Combines both group and guest discussions

```ruby
# app/controllers/api/v1/search_controller.rb:3-13
if group_or_org_id.to_i == 0
  rel = PgSearch.multisearch(params[:query]).where("group_id is null and discussion_id IN (:discussion_ids)", discussion_ids: current_user.guest_discussion_ids)
end

if group_or_org_id.to_i > 0
  rel = PgSearch.multisearch(params[:query]).where("group_id IN (:group_ids)", group_ids: group_ids)
end
```

### Additional Filters

- **By type**: `params[:type]` filters by searchable_type (Discussion, Comment, Poll, Stance, Outcome)
- **By tag**: Filters to discussions/polls with matching tags
- **Ordering**: `authored_at_desc` or `authored_at_asc`
- **Highlighting**: Results include `pg_search_highlight` for matched terms

### Results Limit

Hard-coded to 20 results:
```ruby
results = rel.limit(20).with_pg_search_highlight.all
```

## 5. Database Schema

**Confidence: HIGH**

### pg_search_documents Table

Location: `db/schema.rb:659-679`

| Column | Type | Purpose |
|--------|------|---------|
| content | text | Plain text content (for display/debugging) |
| ts_content | tsvector | PostgreSQL full-text search vector |
| author_id | bigint | Author of the searchable content |
| group_id | bigint | Group containing the content |
| discussion_id | bigint | Discussion containing the content |
| poll_id | bigint | Poll (if applicable) |
| searchable_type | string | Polymorphic type (Discussion, Comment, etc.) |
| searchable_id | bigint | Polymorphic ID |
| authored_at | datetime | When content was authored |
| created_at/updated_at | datetime | Record timestamps |

### Indexes

- `ts_content`: GIN index for fast full-text search
- `authored_at`: Both ASC and DESC indexes for ordering
- `author_id`, `group_id`, `discussion_id`, `poll_id`: B-tree indexes for filtering
- `(searchable_type, searchable_id)`: Composite index for polymorphic lookups

## Summary

| Aspect | Finding | Confidence |
|--------|---------|------------|
| Trigger mechanism | Manual via services + pg_search callbacks | HIGH |
| Automatic indexing | On create/update via pg_search's after_save | HIGH |
| Full reindex | `SearchService.reindex_everything` | HIGH |
| Rake task | None exists | HIGH |
| Admin interface | None for search reindexing | HIGH |
| Searchable models | 5 (Discussion, Comment, Poll, Stance, Outcome) | HIGH |
| Search configuration | PostgreSQL 'simple' with prefix matching | HIGH |

## Recommendations

1. **Add a rake task** for full reindexing:
   ```ruby
   namespace :search do
     task reindex: :environment do
       SearchService.reindex_everything
     end
   end
   ```

2. **Consider adding reindex on create** for Discussion, Comment, Poll, Outcome to ensure immediate searchability.

3. **Add progress logging** to `reindex_everything` for large databases.

4. **Consider background job queue** for full reindex to avoid blocking during deployment.
