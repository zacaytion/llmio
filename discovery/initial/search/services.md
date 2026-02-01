# Search Domain - Services

**Generated:** 2026-02-01
**Confidence:** 5/5

---

## SearchService

**Location:** `/app/services/search_service.rb`

SearchService manages search index maintenance operations. Unlike most Loomio services, it does not handle CRUD operations or authorization - it focuses exclusively on rebuilding search documents.

---

## Methods

### reindex_everything

Performs a complete rebuild of all search documents.

**Behavior:**
1. Executes insert statements for all five searchable models in sequence
2. Does not clear existing documents first (relies on pg_search_insert_statement behavior)

**Use Cases:**
- Initial population after database setup
- Recovery from corrupted search index
- Major schema changes

**Note:** This is called via GenericWorker for background execution during migrations.

---

### reindex_by_author_id(author_id)

Reindexes all content authored by a specific user.

**Behavior:**
1. Deletes all pg_search_documents where author_id matches
2. Executes filtered insert statements for all five models, scoped to that author

**Triggers:**
- User name changes (UserService.update) - updates author name in search content
- User redaction (RedactUserWorker) - removes user's searchable content
- User deactivation (DeactivateUserWorker) - updates searchable content

---

### reindex_by_discussion_id(discussion_id)

Reindexes a discussion and all its nested content.

**Behavior:**
1. Deletes all pg_search_documents where discussion_id matches
2. Executes filtered insert statements for:
   - The discussion itself
   - All comments in the discussion
   - All polls in the discussion
   - All stances on polls in the discussion
   - All outcomes on polls in the discussion

**Triggers:**
- Discussion update (DiscussionService.update)
- Discussion move to different group (DiscussionService.move)
- Event repair (EventService.repair_thread)
- Comment moves between discussions (MoveCommentsWorker)

---

### reindex_by_poll_id(poll_id)

Reindexes a poll and its voting content.

**Behavior:**
1. Deletes all pg_search_documents where poll_id matches
2. Executes filtered insert statements for:
   - The poll itself
   - All stances on the poll
   - All outcomes on the poll

**Triggers:**
- Poll update (PollService.update) - after poll details change
- Poll close (PollService.do_closing_work) - after closing, stances may become visible
- Poll move (PollService.move) - when moved to different discussion/group

---

### reindex_by_comment_id(comment_id)

Reindexes a single comment.

**Behavior:**
1. Deletes pg_search_document for the specific comment
2. Executes filtered insert statement for that comment

**Current Status:** Method exists but appears unused in current codebase. Comment reindexing occurs via discussion-level reindex.

---

## Reindex Trigger Points

### When Reindexing Happens

| Event | Reindex Method | Called From |
|-------|---------------|-------------|
| User name changes | reindex_by_author_id | UserService.update |
| User merged | reindex_by_author_id | UserService.merge |
| User deactivated | reindex_by_author_id | DeactivateUserWorker |
| User redacted | reindex_by_author_id | RedactUserWorker |
| Discussion updated | reindex_by_discussion_id | DiscussionService.update |
| Discussion moved | reindex_by_discussion_id | DiscussionService.move |
| Thread repaired | reindex_by_discussion_id | EventService.repair_thread |
| Comments moved | reindex_by_discussion_id | MoveCommentsWorker |
| Poll updated | reindex_by_poll_id | PollService.update |
| Poll closed | reindex_by_poll_id | PollService.do_closing_work |
| Poll moved | reindex_by_discussion_id | PollService.move |

### Automatic Document Updates

Beyond service-triggered reindexing, pg_search's multisearchable feature calls `update_pg_search_document` automatically when:
- A searchable record is created
- A searchable record is saved with changes

The Searchable concern overrides this to use the custom insert statement approach.

---

## Background Execution

Reindex operations are typically called via GenericWorker for asynchronous execution:

**Pattern:**
```pseudo
GenericWorker.perform_async('SearchService', 'reindex_by_discussion_id', discussion.id)
```

This queues a Sidekiq job that will call SearchService.reindex_by_discussion_id with the provided discussion ID.

---

## Performance Considerations

### Batch Insert vs Record-by-Record

The pg_search_insert_statement approach performs bulk inserts using raw SQL, which is significantly faster than individual record saves. A single INSERT statement can index hundreds of records atomically.

### Index Maintenance

When reindexing large discussions or authors:
1. Documents are deleted first (fast due to indexed foreign keys)
2. New documents are inserted in a single statement
3. GIN index on ts_content is updated atomically

### No Full-Table Locks

The delete/insert approach allows:
- Concurrent searches to continue working
- No blocking of write operations
- Graceful handling of missing documents during the brief reindex window
