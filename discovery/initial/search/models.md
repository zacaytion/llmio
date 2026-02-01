# Search Domain - Models

**Generated:** 2026-02-01
**Confidence:** 5/5

---

## Overview

Loomio implements full-text search using the pg_search gem with PostgreSQL's built-in full-text search capabilities. Search operates across five content types and uses a centralized search documents table for efficient querying.

---

## SearchResult Model

**Location:** `/app/models/search_result.rb`

SearchResult is a plain Ruby object (not ActiveRecord) that uses ActiveModel::Model and ActiveModel::Serialization. It serves as a data transfer object for presenting search results to the frontend.

### Attributes

- `id` - The pg_search_document id
- `searchable_type` - Type of the matched record (Discussion, Comment, Poll, Stance, Outcome)
- `searchable_id` - ID of the matched record
- `poll_title` - Title of related poll (if applicable)
- `discussion_title` - Title of related discussion
- `discussion_key` - URL-safe key for the discussion
- `highlight` - Search result snippet with matched terms bolded
- `poll_key` - URL-safe key for the poll (if applicable)
- `poll_id` - ID of related poll
- `sequence_id` - Position in discussion timeline (for navigation)
- `group_handle` - URL handle of the group
- `group_key` - URL-safe key for the group
- `group_id` - ID of the group
- `group_name` - Display name of the group
- `author_name` - Name of content author
- `author_id` - ID of the author
- `authored_at` - When the content was created
- `tags` - Combined tags from poll and discussion

### Accessor Methods

- `poll` - Lazy-loads the Poll record by poll_id
- `author` - Lazy-loads the User record by author_id

---

## Searchable Concern

**Location:** `/app/models/concerns/searchable.rb`

The Searchable concern is included in all models that participate in full-text search. It integrates with pg_search's multisearch feature.

### Behavior

When included, the concern:
1. Includes PgSearch::Model to provide search capabilities
2. Calls `multisearchable` to register the model with pg_search's global search

### Class Methods

- `rebuild_pg_search_documents` - Executes the model's pg_search_insert_statement to repopulate search documents
- `pg_search_insert_statement(id:, author_id:, discussion_id:)` - Abstract method that must be implemented by each model; generates SQL INSERT statement for search documents

### Document Update Override

The concern overrides `PgSearch::Multisearchable.update_pg_search_document` to:
1. Delete existing documents for the record
2. Execute the model's pg_search_insert_statement for the specific record ID

---

## Searchable Models

Five models include the Searchable concern:

### Discussion

**Fields Indexed:** title, description, author name

**Insert Statement Logic:**
- Concatenates title, description, and author name
- Strips HTML tags using regexp_replace
- Excludes discarded discussions
- Links to group_id and discussion_id

### Comment

**Fields Indexed:** body, author name

**Insert Statement Logic:**
- Concatenates body and author name
- Strips HTML tags
- Excludes discarded comments and comments on discarded discussions
- Inherits group_id from parent discussion
- Includes discussion_id for visibility filtering

### Poll

**Fields Indexed:** title, details, author name

**Insert Statement Logic:**
- Concatenates title, details (poll description), and author name
- Strips HTML tags
- Excludes discarded polls
- Includes poll_id, group_id, and discussion_id references

### Stance (Vote)

**Fields Indexed:** reason, voter name

**Insert Statement Logic:**
- Concatenates reason (voter's explanation) and voter name
- Strips HTML tags
- Excludes stances on discarded polls
- Excludes stances that have not been cast (cast_at is null)
- **Privacy filtering:** Excludes stances from anonymous polls that are still open
- **Privacy filtering:** Excludes stances from polls with hidden results (until_closed) that are still open
- Includes poll_id, group_id, and discussion_id references

### Outcome

**Fields Indexed:** statement, author name

**Insert Statement Logic:**
- Concatenates statement (outcome summary) and author name
- Strips HTML tags
- Excludes outcomes on discarded polls
- Includes poll_id, group_id, and discussion_id references

---

## pg_search_documents Table

**Location:** `db/migrate/20230809101642_create_pg_search_documents.rb`

### Schema

| Column | Type | Purpose |
|--------|------|---------|
| id | bigint | Primary key |
| content | text | Plain text content for highlighting |
| ts_content | tsvector | Pre-computed search vector |
| author_id | bigint | Author for author-based reindexing |
| group_id | bigint | Group for visibility filtering |
| discussion_id | bigint | Discussion for visibility filtering |
| poll_id | bigint | Poll for poll-based reindexing |
| searchable_type | string | Polymorphic type |
| searchable_id | bigint | Polymorphic ID |
| authored_at | datetime | Content creation time for ordering |
| created_at | datetime | Index creation time |
| updated_at | datetime | Last index update |

### Indexes

- `index_pg_search_documents_on_author_id` - For author-based reindexing
- `index_pg_search_documents_on_group_id` - For group-scoped searches
- `index_pg_search_documents_on_discussion_id` - For discussion-scoped visibility
- `index_pg_search_documents_on_poll_id` - For poll-based reindexing
- `index_pg_search_documents_on_searchable` - Polymorphic lookup (type + id)
- `pg_search_documents_searchable_index` - GIN index on ts_content for fast full-text search
- `pg_search_documents_authored_at_asc_index` - For chronological ordering
- `pg_search_documents_authored_at_desc_index` - For reverse chronological ordering

---

## Text Search Configuration

**Location:** `/config/initializers/pg_search.rb`

pg_search is configured with the following options:

- **Dictionary:** Uses 'simple' dictionary (no stemming, no stop words)
- **Prefix matching:** Enabled - allows partial word matching
- **Negation:** Enabled - allows excluding terms with minus prefix
- **Custom tsvector column:** Uses ts_content instead of generating vectors at query time
- **Highlight configuration:** Wraps matched terms in `<b>` tags

### Why 'simple' Dictionary

The 'simple' dictionary is used instead of language-specific dictionaries because:
1. Loomio is multilingual - content may be in any language
2. Simple dictionary provides consistent behavior across all languages
3. Author names are included in content and should not be stemmed
