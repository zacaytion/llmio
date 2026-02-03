# Comprehensive Requirements Quality Checklist: Discussions & Comments

**Purpose**: Author self-review of all functional requirements (FR-001–FR-016) plus edge cases and error paths before PR submission
**Created**: 2026-02-03
**Reviewed**: 2026-02-03
**Feature**: [spec.md](../spec.md)
**Depth**: Comprehensive (~81 items)
**Audience**: Author (pre-PR self-review)

---

## Requirement Completeness

- [x] CHK001 - Are all 16 functional requirements (FR-001 through FR-016) traceable to at least one task? ✓ tasks.md maps all FRs: FR-001→T017-T025, FR-002→T019, FR-003→T049-T058, etc. Coverage: 100%
- [x] CHK002 - Are discussion creation requirements complete for both group and direct discussion paths? ✓ Group: FR-001, US1; Direct: FR-003, US4. Both in contracts/discussions.yaml
- [x] CHK003 - Are comment CRUD operations fully specified (create, read, update, soft-delete)? ✓ Create: FR-005/FR-006, Read: GET /discussions/{id}, Update: FR-008, Delete: FR-009. All in contracts/comments.yaml
- [x] CHK004 - Are all permission check scenarios documented (member, admin, non-member, author)? ✓ Member: FR-001, Admin: FR-002, Non-member: 403 response, Author: FR-008/FR-009/FR-010
- [x] CHK005 - Are read tracking requirements complete for both timestamp and volume preferences? ✓ Timestamp: FR-012, Volume: FR-013. Both in contracts/discussions.yaml PUT /read-state
- [x] CHK006 - Is the participant management flow fully specified for direct discussions? ✓ Access: FR-014, Add: FR-015. contracts/discussions.yaml POST /participants

**Score: 6/6 (100%)** ✓ Pass

---

## Requirement Clarity

- [x] CHK007 - Is `max_depth` behavior clearly defined when depth equals vs exceeds the limit? ✓ spec.md US2.3: "reply appears at the same level (flattened)"; research.md:15 confirms parent preserved
- [x] CHK008 - Is "soft delete" clearly distinguished from hard delete with explicit behavior? ✓ Soft delete: FR-009 (comments, sets discarded_at, shows "[deleted]"); Hard delete: discussion CASCADE
- [x] CHK009 - Are the three volume levels (mute/normal/loud) clearly differentiated in behavior? ✓ FR-013, US5.3 (mute = no notifications). Future 009-notifications will define loud behavior
- [x] CHK010 - Is "private=true default" clearly specified as the behavior when no value provided? ✓ FR-016, data-model.md:38, contracts/discussions.yaml:250 all specify default true
- [x] CHK011 - Is "group admin" role clearly defined in relation to discussion/comment permissions? ✓ FR-002 (create bypass), FR-009 (delete any), FR-010/011 (close/reopen). Admin = membership.admin flag
- [x] CHK012 - Are "open" vs "closed" discussion states clearly defined with allowed operations? ✓ Open: closed_at IS NULL, comments allowed. Closed: closed_at set, comments blocked (FR-010)
- [x] CHK013 - Is "participant" clearly distinguished from "reader" in direct discussions? ✓ research.md:22-27: participant = reader who can write; data-model.md:85 participant boolean flag

**Score: 7/7 (100%)** ✓ Pass

---

## Requirement Consistency

- [x] CHK014 - Are permission requirements consistent between spec.md and contracts/*.yaml? ✓ All 401/403 responses in contracts match FR permission requirements
- [x] CHK015 - Are entity field names consistent across data-model.md, contracts/, and spec.md? ✓ discussion_id, author_id, parent_id, closed_at, discarded_at - all consistent
- [x] CHK016 - Are close/reopen permission rules consistent between FR-010 and FR-011? ✓ Both: "discussion authors and group admins" - identical pattern
- [x] CHK017 - Is the "author or admin" permission pattern consistently applied across all operations? ✓ Edit discussion, close/reopen, delete comment - all use same pattern
- [x] CHK018 - Are depth values consistent (spec says default 3, plan says up to 10 levels)? ✓ No conflict: default=3 (configurable), SC-003 "up to 10 levels respecting max_depth configuration"
- [x] CHK019 - Are the table/column names in data-model.md consistent with migration file references in plan.md? ✓ discussions, comments, discussion_readers match plan.md:65-69 migration files

**Score: 6/6 (100%)** ✓ Pass

---

## Acceptance Criteria Quality

- [x] CHK020 - Can SC-002 (100% permission enforcement) be objectively verified via tests? ✓ Permission matrix tests in tasks T016, T017, T018 cover all combinations
- [x] CHK021 - Is SC-004 (500ms read-state update) measurable with specific instrumentation? ✓ **REMEDIATED** spec.md:134 now specifies "p95" with "measured server-side from request receipt to response sent"
- [x] CHK022 - Can SC-005 (direct discussion visibility) be verified with security tests? ✓ T052 updates CanUserAccessDiscussion; security tests verify non-participant denial
- [x] CHK023 - Can SC-006 (soft-delete preserves children) be verified via query tests? ✓ T074 GetCommentsByDiscussion returns "[deleted]" body; T082 verification task
- [x] CHK024 - Can SC-007 (closed discussion blocks comments) be verified via integration tests? ✓ T081 explicit verification task; US3.2 acceptance scenario
- [x] CHK025 - Are acceptance scenarios in each user story testable without ambiguity? ✓ All 5 user stories have Given/When/Then scenarios with concrete assertions

**Score: 6/6 (100%)** ✓ Pass

---

## Scenario Coverage - Primary Flows

- [x] CHK026 - Are requirements defined for creating a discussion in a group with permission enabled? ✓ US1.1, FR-001, T017-T025
- [x] CHK027 - Are requirements defined for creating a discussion as group admin (bypassing flag)? ✓ US1.3, FR-002, T019 CanUserCreateDiscussion
- [x] CHK028 - Are requirements defined for adding a top-level comment to an open discussion? ✓ US2.1, FR-005, T026-T038
- [x] CHK029 - Are requirements defined for adding a nested reply within max_depth? ✓ US2.2, FR-006, T030 depth calculation query
- [x] CHK030 - Are requirements defined for editing own comment and recording edited_at? ✓ US2.4, FR-008, T027/T033 UpdateCommentBody sets edited_at
- [x] CHK031 - Are requirements defined for closing a discussion as author? ✓ US3.1, FR-010, T044 DiscussionService.Close
- [x] CHK032 - Are requirements defined for reopening a closed discussion as admin? ✓ US3.3, FR-011, T045 DiscussionService.Reopen
- [x] CHK033 - Are requirements defined for creating a direct discussion with participants? ✓ US4.1, FR-003, T055 handles participant_ids
- [x] CHK034 - Are requirements defined for recording last_read_at on discussion access? ✓ US5.1, FR-012, T061 UpsertReadState

**Score: 9/9 (100%)** ✓ Pass

---

## Scenario Coverage - Alternate Flows

- [x] CHK035 - Are requirements defined for reply at max_depth (flattening behavior)? ✓ US2.3, FR-007, T032 depth validation against max_depth
- [x] CHK036 - Are requirements defined for creating discussion without description (optional field)? ✓ FR-004 "description is optional", contracts CreateDiscussionRequest
- [x] CHK037 - Are requirements defined for setting notification volume to mute? ✓ US5.3, FR-013, contracts UpdateReadStateRequest volume enum
- [x] CHK038 - Are requirements defined for adding participants to existing direct discussion? ✓ US4.3, FR-015, T053/T057 POST /discussions/{id}/participants

**Score: 4/4 (100%)** ✓ Pass

---

## Scenario Coverage - Exception/Error Flows

- [x] CHK039 - Are requirements defined for permission denied when member lacks `members_can_start_discussions`? ✓ US1.2, contracts 403 response
- [x] CHK040 - Are requirements defined for attempting to comment on a closed discussion? ✓ US3.2, contracts/comments.yaml:35 "Access denied or discussion closed"
- [x] CHK041 - Are requirements defined for non-participant accessing direct discussion? ✓ US4.2, FR-014, contracts 403 response
- [x] CHK042 - Are requirements defined for non-author/non-admin attempting to close discussion? ✓ US3.4, contracts/discussions.yaml:135 403 "Permission denied"
- [x] CHK043 - Are requirements defined for attempting to delete already-deleted comment? ✓ contracts/comments.yaml:100 409 "Already deleted"
- [x] CHK044 - Are requirements defined for invalid parent_id when creating nested comment? ✓ **REMEDIATED** contracts/comments.yaml:37 404 "parent not found in this discussion"
- [x] CHK045 - Are requirements defined for discussion creation with empty title? ✓ contracts/discussions.yaml minLength: 1; 400 response

**Score: 7/7 (100%)** ✓ Pass

---

## Edge Case Coverage

- [x] CHK046 - Is behavior specified when comment's parent is soft-deleted? ✓ spec.md:94 Edge Cases "reply remains visible as a top-level reply in its subtree"
- [x] CHK047 - Is behavior specified when max_depth is set to 0 (flat discussion)? ✓ spec.md:98 Edge Cases "All comments appear at root level"
- [x] CHK048 - Is behavior specified when group is archived (read-only discussions)? ✓ spec.md:96 Edge Cases "read-only; no new discussions can be created"
- [x] CHK049 - Is behavior specified for concurrent edits to same comment (last-write-wins)? ✓ spec.md:95 Edge Cases "Last write wins"
- [x] CHK050 - Is behavior specified when discussion author's account is deleted (SET NULL)? ✓ **REMEDIATED** data-model.md:37,135 author_id nullable with ON DELETE SET NULL
- [x] CHK051 - Is behavior specified when comment author's account is deleted (SET NULL)? ✓ **REMEDIATED** data-model.md:57,138 author_id nullable with ON DELETE SET NULL
- [x] CHK052 - Is behavior specified for direct discussion with zero additional participants (author-only)? ✓ contracts participant_ids array can be empty; author implicitly included
- [x] CHK053 - Is cascade delete behavior documented for discussion deletion? ✓ data-model.md:147-152 FK CASCADE table

**Score: 8/8 (100%)** ✓ Pass

---

## Non-Functional Requirements

- [x] CHK054 - Are performance requirements quantified for read-state updates (SC-004: 500ms)? ✓ **REMEDIATED** spec.md:134 "p95 of read-state updates complete within 500ms"
- [x] CHK055 - Are there index requirements specified for query performance? ✓ data-model.md:44-47, 66-69, 88-90 specify all indexes
- [x] CHK056 - Are security requirements for permission checking explicitly stated? ✓ SC-002 "100% of access attempts", SC-005 "never visible to non-participants"
- [x] CHK057 - Are audit trail requirements defined for comment edits/deletes? ✓ edited_at timestamp for edits; discarded_at for deletes. Full audit deferred to future feature.
- [x] CHK058 - Are rate limiting requirements defined for API endpoints? ✓ Constitution III.4 mandates rate limiting; infra layer responsibility
- [x] CHK059 - Are input validation requirements defined (title length, body content)? ✓ contracts: title 1-255 chars, body minLength: 1. T079/T080 validation tasks

**Score: 6/6 (100%)** ✓ Pass

---

## Dependencies & Assumptions

- [x] CHK060 - Is Feature 004 (Groups & Memberships) dependency explicitly documented? ✓ spec.md:156 Dependencies section
- [x] CHK061 - Is Feature 001 (User Authentication) dependency explicitly documented? ✓ spec.md:157 Dependencies section
- [x] CHK062 - Is the assumption about plain text/markdown content (no rich text) documented? ✓ spec.md:149 Assumptions
- [x] CHK063 - Is the assumption about deferred real-time updates documented? ✓ spec.md:151 Assumptions "Real-time updates... are out of scope"
- [x] CHK064 - Is the assumption about deferred email notifications documented? ✓ spec.md:150 Assumptions "Email notifications... are out of scope"
- [x] CHK065 - Are migration number dependencies (002, 003 from Feature 004) documented? ✓ plan.md:65-66, tasks.md T000 prerequisite check

**Score: 6/6 (100%)** ✓ Pass

---

## Ambiguities & Conflicts

- [x] CHK066 - Is there potential conflict between "default 3" max_depth and "up to 10 levels" in SC-003? ✓ No conflict - SC-003 says "respecting max_depth configuration"; 10 is system capability, 3 is default
- [x] CHK067 - Is "visual marker" in US5 sufficiently backend-agnostic after remediation? ✓ US5.2 remediated: "API response includes an unread comment count" - backend provides data
- [x] CHK068 - Is the distinction between "hard delete" (discussion) and "soft delete" (comment) clear? ✓ Comments: soft delete (discarded_at, body→"[deleted]"). Discussions: hard delete (CASCADE)
- [x] CHK069 - Is behavior defined when reopening an already-open discussion? ✓ contracts/discussions.yaml:164-165 409 Conflict "Already open"
- [x] CHK070 - Is behavior defined when closing an already-closed discussion? ✓ contracts/discussions.yaml:137-138 409 Conflict "Already closed"

**Score: 5/5 (100%)** ✓ Pass

---

## API Contract Completeness

- [x] CHK071 - Are all CRUD endpoints specified in contracts/*.yaml for discussions? ✓ **REMEDIATED** GET list, POST create, GET/{id}, PATCH/{id}, DELETE/{id}, plus close/reopen/participants/read-state
- [x] CHK072 - Are all CRUD endpoints specified in contracts/*.yaml for comments? ✓ POST create, PATCH/{id}, DELETE/{id}. Read via GET /discussions/{id}
- [x] CHK073 - Are all error response codes documented (400, 401, 403, 404, 409)? ✓ All endpoints have complete error responses in contracts
- [x] CHK074 - Are request body schemas complete with required/optional field markers? ✓ CreateDiscussionRequest, UpdateDiscussionRequest, CreateCommentRequest all specify required[]
- [x] CHK075 - Are response schemas complete with all entity fields? ✓ Discussion, Comment, DiscussionReader schemas match data-model.md
- [x] CHK076 - Is the DiscussionWithComments response schema fully specified? ✓ contracts/discussions.yaml:349-360 allOf Discussion + comments array + reader

**Score: 6/6 (100%)** ✓ Pass

---

## Data Model Completeness

- [x] CHK077 - Are all foreign key relationships documented with ON DELETE behavior? ✓ data-model.md:144-152 FK Cascade Behavior table
- [x] CHK078 - Are all indexes specified for query optimization? ✓ data-model.md:44-47 (discussions), 66-69 (comments), 88-90 (readers)
- [x] CHK079 - Are all CHECK constraints documented (volume values, depth >= 0, max_depth >= 0)? ✓ data-model.md:39 (max_depth), 60 (depth), 84 (volume IN clause)
- [x] CHK080 - Is the UNIQUE constraint on (discussion_id, user_id) for readers documented? ✓ data-model.md:89 idx_discussion_readers_unique
- [x] CHK081 - Are timestamp fields (created_at, updated_at, edited_at, discarded_at, closed_at) consistently specified? ✓ All timestamp fields documented with DEFAULT NOW() where applicable

**Score: 5/5 (100%)** ✓ Pass

---

## Summary

| Category | Items | Score | Status |
|----------|-------|-------|--------|
| Requirement Completeness | CHK001-CHK006 | 6/6 | ✓ Pass |
| Requirement Clarity | CHK007-CHK013 | 7/7 | ✓ Pass |
| Requirement Consistency | CHK014-CHK019 | 6/6 | ✓ Pass |
| Acceptance Criteria Quality | CHK020-CHK025 | 6/6 | ✓ Pass |
| Scenario Coverage - Primary | CHK026-CHK034 | 9/9 | ✓ Pass |
| Scenario Coverage - Alternate | CHK035-CHK038 | 4/4 | ✓ Pass |
| Scenario Coverage - Exception | CHK039-CHK045 | 7/7 | ✓ Pass |
| Edge Case Coverage | CHK046-CHK053 | 8/8 | ✓ Pass |
| Non-Functional Requirements | CHK054-CHK059 | 6/6 | ✓ Pass |
| Dependencies & Assumptions | CHK060-CHK065 | 6/6 | ✓ Pass |
| Ambiguities & Conflicts | CHK066-CHK070 | 5/5 | ✓ Pass |
| API Contract Completeness | CHK071-CHK076 | 6/6 | ✓ Pass |
| Data Model Completeness | CHK077-CHK081 | 5/5 | ✓ Pass |

**Total Items**: 81
**Verified**: 81/81 (100%)
**Remediations Applied**: 5 (referenced from prior analysis: C1, C2, C3, CHK008, CHK027)
**Status**: ✅ **PASS - Ready for PR Submission**

---

## Notes

- All items verified against spec.md, plan.md, data-model.md, contracts/*.yaml, tasks.md, and research.md
- PostgreSQL best practices confirmed: nullable FKs with SET NULL, proper CASCADE behavior, UNIQUE constraints for upsert patterns
- Constitution alignment verified: TDD workflow in tasks, Huma-first API design, security-first permission checks
