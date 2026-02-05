# Research: Discussions & Comments

**Feature**: 005-discussions | **Date**: 2026-02-02

## Overview

This document captures technical decisions and patterns for implementing threaded discussions with comments.

---

## 1. Comment Threading Strategy

**Decision**: Adjacency list with depth tracking

**Rationale**: Each comment stores `parent_id` (nullable FK to comments) and computed `depth` (integer). The database enforces `depth <= max_depth` via CHECK constraint. Queries retrieve comments in creation order; the API returns flat lists with parent references for client-side tree construction.

**Alternatives considered**:
- **Nested sets**: Complex updates on insert; overkill for shallow trees (max depth 3-10)
- **Materialized path**: String concatenation adds complexity; adjacency list suffices
- **Recursive CTE only**: Selected for reads, but storing `depth` avoids runtime computation

---

## 2. Direct Discussion Participant Storage

**Decision**: Reuse `discussion_readers` table with `participant` boolean flag

**Rationale**: Direct discussions (no group) need explicit participant lists. Rather than a separate `discussion_participants` table, extend `discussion_readers` with a `participant` boolean. Participants are readers who can also write; non-participants cannot access the discussion at all.

**Alternatives considered**:
- **Separate `discussion_participants` table**: Adds complexity; participant and reader concerns overlap significantly
- **JSON array on discussion**: Loses FK integrity; harder to query

---

## 3. Permission Check Pattern

**Decision**: Centralized `permissions.go` with group flag evaluation

**Rationale**: All permission checks flow through `CanUserAccessDiscussion`, `CanUserCreateComment`, etc. These functions:
1. Check if discussion is direct → verify user is participant
2. Check if discussion is in group → verify membership + relevant `members_can_*` flag
3. Admins bypass member-level flags

**Pattern** (from Feature 004):
```go
func CanUserCreateDiscussion(ctx context.Context, userID, groupID int64) (bool, error) {
    membership, err := db.GetMembership(ctx, userID, groupID)
    if err != nil { return false, err }
    if membership.Admin { return true, nil }
    group, err := db.GetGroup(ctx, groupID)
    if err != nil { return false, err }
    return group.MembersCanStartDiscussions, nil
}
```

---

## 4. Soft Delete Display

**Decision**: API returns `body: "[deleted]"` for discarded comments; `discarded_at` exposed in metadata

**Rationale**: Preserves reply structure. Children of deleted comments remain visible. The `[deleted]` placeholder matches Loomio's existing behavior. Actual body content is not returned—database retains it for audit but API hides it.

---

## 5. Read State Updates

**Decision**: Upsert on discussion access; background job not needed for MVP

**Rationale**: When a user opens a discussion, the API upserts their `discussion_readers` row with `last_read_at = NOW()`. This synchronous approach meets the 500ms target (SC-004). Background aggregation deferred until scale demands it.

**Query pattern**:
```sql
INSERT INTO discussion_readers (discussion_id, user_id, last_read_at, volume)
VALUES ($1, $2, NOW(), 'normal')
ON CONFLICT (discussion_id, user_id) DO UPDATE SET last_read_at = NOW();
```

---

## 6. Volume Enum

**Decision**: PostgreSQL CHECK constraint with enum-like values

**Rationale**: Store `volume` as `varchar(20)` with CHECK constraint `volume IN ('mute', 'normal', 'loud')`. Avoids PostgreSQL ENUM type's migration complexity while maintaining data integrity.

---

## 7. Title and Body Constraints

**Decision**: Title max 255 chars (required); body unlimited text (optional)

**Rationale**: Matches Loomio's existing schema. Title length enforced via Huma validation and database VARCHAR(255). Body stored as TEXT with no application-level limit for MVP.

---

## 8. Closed Discussion Enforcement

**Decision**: Check `closed_at IS NULL` in comment creation, not via database trigger

**Rationale**: Application-level check in `CommentService.Create()` returns clear error message. Database trigger would obscure the reason. Closed state is a business rule, not a data integrity constraint.

---

## Summary

| Topic | Decision |
|-------|----------|
| Threading | Adjacency list with `parent_id` + `depth` |
| Direct participants | `discussion_readers.participant` boolean |
| Permissions | Centralized `permissions.go` checking group flags |
| Soft delete | Return `"[deleted]"` body; keep children visible |
| Read tracking | Synchronous upsert on access |
| Volume | VARCHAR(20) with CHECK constraint |
| Title/body | VARCHAR(255) required / TEXT optional |
| Closed enforcement | Application-level check |
