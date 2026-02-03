# Data Model: Discussions & Comments

**Feature**: 005-discussions | **Date**: 2026-02-02

## Entity Relationship

```
┌─────────────┐     ┌─────────────────┐     ┌─────────────┐
│   groups    │────<│   discussions   │>────│    users    │
└─────────────┘     └─────────────────┘     └─────────────┘
      1:N                   │                     1:N
   (optional)               │                  (author_id)
                           1:N
                            │
               ┌────────────┴────────────┐
               │                         │
        ┌──────▼──────┐         ┌────────▼────────┐
        │  comments   │         │discussion_readers│
        └─────────────┘         └─────────────────┘
              │
           self-ref
         (parent_id)
```

---

## Tables

### discussions

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | BIGSERIAL | PK | |
| `title` | VARCHAR(255) | NOT NULL | Required |
| `description` | TEXT | | Optional body |
| `group_id` | BIGINT | FK groups(id) ON DELETE CASCADE | NULL = direct discussion |
| `author_id` | BIGINT | FK users(id) ON DELETE SET NULL | Creator (nullable if user deleted) |
| `private` | BOOLEAN | NOT NULL DEFAULT true | Visibility flag |
| `max_depth` | INTEGER | NOT NULL DEFAULT 3, CHECK (max_depth >= 0) | Comment nesting limit |
| `closed_at` | TIMESTAMPTZ | | NULL = open |
| `created_at` | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | |
| `updated_at` | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | |

**Indexes**:
- `idx_discussions_group_id` on `group_id`
- `idx_discussions_author_id` on `author_id`
- `idx_discussions_created_at` on `created_at DESC`

---

### comments

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | BIGSERIAL | PK | |
| `discussion_id` | BIGINT | FK discussions(id) ON DELETE CASCADE, NOT NULL | Parent discussion |
| `author_id` | BIGINT | FK users(id) ON DELETE SET NULL | Creator (nullable if user deleted) |
| `parent_id` | BIGINT | FK comments(id) ON DELETE CASCADE | NULL = top-level |
| `body` | TEXT | NOT NULL | Comment content |
| `depth` | INTEGER | NOT NULL DEFAULT 0, CHECK (depth >= 0) | Nesting level (0 = root) |
| `edited_at` | TIMESTAMPTZ | | NULL = never edited |
| `discarded_at` | TIMESTAMPTZ | | Soft delete timestamp |
| `created_at` | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | |

**Indexes**:
- `idx_comments_discussion_id` on `discussion_id`
- `idx_comments_parent_id` on `parent_id`
- `idx_comments_author_id` on `author_id`
- `idx_comments_created_at` on `discussion_id, created_at`

**Constraints**:
- Trigger or application logic enforces `depth <= discussion.max_depth`

---

### discussion_readers

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | BIGSERIAL | PK | |
| `discussion_id` | BIGINT | FK discussions(id) ON DELETE CASCADE, NOT NULL | |
| `user_id` | BIGINT | FK users(id) ON DELETE CASCADE, NOT NULL | |
| `last_read_at` | TIMESTAMPTZ | | NULL = never read |
| `volume` | VARCHAR(20) | NOT NULL DEFAULT 'normal', CHECK (volume IN ('mute', 'normal', 'loud')) | Notification preference |
| `participant` | BOOLEAN | NOT NULL DEFAULT false | True = explicit participant (for direct discussions) |
| `created_at` | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | |

**Indexes**:
- UNIQUE `idx_discussion_readers_unique` on `(discussion_id, user_id)`
- `idx_discussion_readers_user_id` on `user_id`

---

## State Transitions

### Discussion Lifecycle

```
┌─────────┐   close    ┌────────┐
│  Open   │───────────>│ Closed │
└─────────┘            └────────┘
     ^                      │
     │       reopen         │
     └──────────────────────┘
```

- **Open**: `closed_at IS NULL` — comments allowed
- **Closed**: `closed_at IS NOT NULL` — comments blocked

### Comment Lifecycle

```
┌─────────┐   edit     ┌─────────┐
│ Active  │───────────>│ Edited  │
└─────────┘            └─────────┘
     │                      │
     │ soft-delete          │ soft-delete
     v                      v
┌──────────┐           ┌──────────┐
│ Deleted  │           │ Deleted  │
└──────────┘           └──────────┘
```

- **Active**: `discarded_at IS NULL`, `edited_at IS NULL`
- **Edited**: `discarded_at IS NULL`, `edited_at IS NOT NULL`
- **Deleted**: `discarded_at IS NOT NULL` — body hidden, children visible

---

## Validation Rules

| Entity | Field | Rule |
|--------|-------|------|
| Discussion | title | Required, 1-255 characters |
| Discussion | author_id | Required on creation; becomes NULL if user deleted |
| Discussion | max_depth | Integer ≥ 0, default 3 |
| Comment | body | Required, non-empty |
| Comment | author_id | Required on creation; becomes NULL if user deleted |
| Comment | depth | Must be ≤ parent discussion's max_depth |
| DiscussionReader | volume | One of: mute, normal, loud |

---

## Foreign Key Cascade Behavior

| Parent | Child | ON DELETE |
|--------|-------|-----------|
| groups | discussions | CASCADE (archive group archives discussions) |
| discussions | comments | CASCADE |
| discussions | discussion_readers | CASCADE |
| comments | comments (parent) | CASCADE (delete parent deletes subtree) |
| users | discussions (author) | SET NULL |
| users | comments (author) | SET NULL |
| users | discussion_readers | CASCADE |
