# Pre-Merge API Requirements Quality Checklist: Groups & Memberships

**Purpose**: Final pre-merge review validating requirements quality for critical gaps identified in PR #4 review
**Created**: 2026-02-03
**Updated**: 2026-02-03 (specification gaps addressed)
**Feature**: [spec.md](../spec.md) | [contracts/groups.yaml](../contracts/groups.yaml)
**Scope**: Includes deferred functionality (9 permission flags for Features 005/006)
**Context**: Phase 11 PR review findings from code-reviewer, silent-failure-hunter, pr-test-analyzer agents

---

## Critical Requirements Gaps (From PR Review)

> Issues identified in tasks.md Phase 11 that indicate specification gaps

- [x] CHK001 - Is the error detection mechanism for unique violations specified (type assertion vs string matching)? [Fixed: Spec §TC-008 specifies pgconn.PgError type assertion]
- [x] CHK002 - Is the PostgreSQL error code mapping specified for DB trigger errors (P0001 → 409)? [Fixed: Spec §TC-009 specifies P0001 mapping to 409]
- [x] CHK003 - Are mutation restrictions for archived groups explicitly documented for ALL membership operations? [Fixed: Spec §Edge Cases "Archived Group Behavior" section]
- [x] CHK004 - Is the authorization boundary for GET /memberships/{id} explicitly specified (who can view)? [Fixed: Contract getMembership 403 description + Spec §Edge Cases "Authorization Boundaries"]

## Authorization Boundary Completeness

> Requirements for pending member access (discovered via tasks T123-T124)

- [x] CHK005 - Are requirements specified for pending member attempting to view group details via GET /groups/{id}? [Fixed: Spec §Edge Cases "Membership Invitations" - returns 403]
- [x] CHK006 - Are requirements specified for pending member attempting to invite others via POST /groups/{id}/memberships? [Fixed: Spec §Edge Cases + Contract inviteMember 403 description]
- [x] CHK007 - Are requirements specified for non-member attempting to remove member via DELETE /memberships/{id}? [Fixed: Spec §Edge Cases "Authorization Boundaries" + Contract removeMember 403]
- [x] CHK008 - Is the authorization check order documented (authentication → authorization → resource validation)? [Fixed: Spec §TC-011 specifies authorization order]

## Error Response Specification Gaps

> Missing error response specifications identified during implementation

- [x] CHK009 - Is the error response for non-admin attempting to invite with admin role specified? [Fixed: Spec §Edge Cases + Contract inviteMember description]
- [x] CHK010 - Is the error response for mutations on archived groups specified with HTTP status code? [Fixed: Spec §Edge Cases + Contract endpoints with 409 archived group]
- [x] CHK011 - Is the behavior for accepting an already-accepted invitation specified with clear 409 message? [Fixed: Spec §Edge Cases + Contract ErrorResponse 409 messages]
- [x] CHK012 - Is the behavior for promoting an already-admin member specified? [Fixed: Spec §Edge Cases "Role Changes" + Contract promoteMember 409]
- [x] CHK013 - Is the behavior for demoting an already-member (non-admin) specified? [Fixed: Spec §Edge Cases "Role Changes" + Contract demoteMember 409]

## Handle Validation Requirements

> Validation gaps discovered during implementation (tasks T138-T140)

- [x] CHK014 - Is the error response for invalid handle format (422 vs 500) explicitly specified? [Fixed: Spec §Edge Cases "Group Lifecycle" - returns 422]
- [x] CHK015 - Is handle validation specified to occur BEFORE database insert (application layer vs DB constraint)? [Fixed: Spec §TC-010 specifies application-layer validation]
- [x] CHK016 - Are leading/trailing hyphen restrictions documented in the handle format specification? [Fixed: Spec §Edge Cases + Contract GroupDTO.handle description with full regex]

## Deferred Functionality Requirements

> Quality of placeholder specs for Features 005/006

- [x] CHK017 - Is `members_can_add_guests` behavior sufficiently specified for Feature 005 (what is a "guest"?)? [Fixed: Spec §Deferred Functionality table defines "guest"]
- [x] CHK018 - Is `members_can_start_discussions` enforcement context specified (where does check occur)? [Fixed: Spec §Deferred Functionality table specifies createDiscussion handler]
- [x] CHK019 - Is `members_can_raise_motions` "motion" vs "poll" terminology clarified? [Fixed: Spec §Deferred Functionality clarifies synonymous terminology]
- [x] CHK020 - Is `members_can_edit_discussions` scope clear (which fields: title, description, context)? [Fixed: Spec §Deferred Functionality table specifies fields]
- [x] CHK021 - Is `members_can_edit_comments` scope clear (own comments only, or any member's comments)? [Fixed: Spec §Deferred Functionality table clarifies own comments only]
- [x] CHK022 - Is `members_can_delete_comments` scope clear (own comments only, or any member's comments)? [Fixed: Spec §Deferred Functionality table clarifies own comments only]
- [x] CHK023 - Is `members_can_announce` "announcement" behavior defined (what is an announcement)? [Fixed: Spec §Deferred Functionality table defines announcement]
- [x] CHK024 - Is `admins_can_edit_user_content` scope clear (discussions + comments, or more)? [Fixed: Spec §Deferred Functionality table specifies scope]
- [x] CHK025 - Is `parent_members_can_see_discussions` interaction with subgroup membership documented? [Fixed: Spec §Deferred Functionality table specifies enforcement]

## Edge Case Specifications

> Edge cases discovered during PR review (not covered in existing checklist)

- [x] CHK026 - Is behavior specified for archiving a group with pending invitations? [Fixed: Spec §Edge Cases "Archived Group Behavior" - invitations remain pending]
- [x] CHK027 - Is behavior specified for unarchiving a group when its parent is still archived? [Fixed: Spec §Edge Cases - subgroups can unarchive independently]
- [x] CHK028 - Is behavior specified for creating a subgroup under an archived parent? [Fixed: Spec §Edge Cases + Contract createSubgroup 409]
- [x] CHK029 - Is behavior for concurrent promote/demote on same membership specified? [Fixed: Spec §Edge Cases "Concurrency" - DB trigger atomic enforcement]
- [x] CHK030 - Is the maximum subgroup nesting depth specified (or explicitly unlimited)? [Fixed: Spec §Edge Cases "Subgroup Hierarchy" - explicitly unlimited]

## Audit Trail Requirements

> Audit requirements that lack specification detail

- [x] CHK031 - Is the audit record JSONB field mapping specified (which fields captured per table)? [Fixed: Spec §Audit Trail Specification table and JSONB Field Mapping section]
- [x] CHK032 - Is the actor_id source specified for system-initiated actions (cron jobs, migrations)? [Fixed: Spec §Audit Trail Specification "System-Initiated Actions"]
- [x] CHK033 - Is audit record retention policy specified? [Fixed: Spec §Audit Trail Specification "Retention Policy" - indefinite with future TTL note]
- [x] CHK034 - Is audit access API specified (or explicitly excluded from Feature 004 scope)? [Fixed: Spec §Audit Trail Specification "Audit Access API" - explicitly excluded]

## Schema Quality (API Contract)

> Type and schema specification gaps identified by type-design-analyzer

- [x] CHK035 - Is GroupDTO.handle format constraint documented in schema description? [Fixed: Contract GroupDTO.handle description with full constraints]
- [x] CHK036 - Is MembershipDTO.accepted_at null semantics documented (null = pending invitation)? [Fixed: Contract MembershipDTO description and accepted_at field]
- [x] CHK037 - Is GroupDetailDTO.current_user_role enum exhaustively documented (admin|member only, or future roles)? [Fixed: Contract GroupDetailDTO.current_user_role description - "exhaustive, no additional roles planned"]
- [x] CHK038 - Is GroupDetailDTO.member_count definition clear (includes pending, or active only)? [Fixed: Contract GroupDetailDTO.member_count - "active only, does not include pending"]
- [x] CHK039 - Is GroupDetailDTO.admin_count definition clear (includes pending admin invites)? [Fixed: Contract GroupDetailDTO.admin_count - "does not include pending"]

## API Design Consistency

> Consistency issues not caught by original checklist

- [x] CHK040 - Is the handle lookup path consistent (spec: /groups/handle/{handle} vs tasks: /group-by-handle/{handle})? [Fixed: Contract updated to /group-by-handle/{handle} with explanation note]
- [x] CHK041 - Is 204 vs 200 consistent for successful mutations (archive returns 200 with body, remove returns 204)? [Accepted: Intentional - archive returns body for client convenience, remove has no useful body]
- [x] CHK042 - Are error message strings specified for all 409 Conflict scenarios? [Fixed: Contract ErrorResponse description lists all 409 messages]

## Cross-Feature Integration

> Requirements for integration with Features 001 (auth) and 005/006

- [x] CHK043 - Is the session cookie contract with Feature 001 explicitly documented? [Fixed: Spec §Cross-Feature Integration "Feature 001"]
- [x] CHK044 - Is the permission flag contract with Feature 005 (Discussions) explicitly documented? [Fixed: Spec §Cross-Feature Integration "Feature 005 Contract"]
- [x] CHK045 - Is the permission flag contract with Feature 006 (Polls) explicitly documented? [Fixed: Spec §Cross-Feature Integration "Feature 006 Contract"]
- [x] CHK046 - Is API versioning strategy documented for adding new permission flags? [Fixed: Spec §Cross-Feature Integration "API Versioning"]

## Performance & Non-Functional Requirements

> NFR gaps identified during implementation

- [x] CHK047 - Is "normal load" quantified in SC-001 ("< 500ms p95 under normal load")? [Fixed: Spec §SC-001 defines normal load parameters]
- [x] CHK048 - Are response time requirements specified for each endpoint type (create/read/update/delete)? [Fixed: Spec §SC-001 includes per-operation targets]
- [x] CHK049 - Is the rate limit for invitations quantified in TC-007 (currently "e.g., max 10 invites")? [Accepted: TC-007 notes this is deferred to infrastructure layer; example is guidance not requirement]
- [x] CHK050 - Are concurrent request handling requirements specified (pessimistic vs optimistic locking)? [Fixed: Spec §Edge Cases "Concurrency" + TC-005 describes DB trigger atomic enforcement]

---

## Summary

| Quality Dimension | Item Count | Completed |
|-------------------|------------|-----------|
| Critical Requirements Gaps (PR Review) | CHK001-CHK004 | 4/4 ✓ |
| Authorization Boundary Completeness | CHK005-CHK008 | 4/4 ✓ |
| Error Response Specification Gaps | CHK009-CHK013 | 5/5 ✓ |
| Handle Validation Requirements | CHK014-CHK016 | 3/3 ✓ |
| Deferred Functionality Requirements | CHK017-CHK025 | 9/9 ✓ |
| Edge Case Specifications | CHK026-CHK030 | 5/5 ✓ |
| Audit Trail Requirements | CHK031-CHK034 | 4/4 ✓ |
| Schema Quality (API Contract) | CHK035-CHK039 | 5/5 ✓ |
| API Design Consistency | CHK040-CHK042 | 3/3 ✓ |
| Cross-Feature Integration | CHK043-CHK046 | 4/4 ✓ |
| Performance & Non-Functional | CHK047-CHK050 | 4/4 ✓ |

**Total**: 50/50 items complete ✓

---

## Usage Notes

- Check items off as completed: `[x]`
- Add findings inline with resolution or deferral decision
- Items reference tasks.md task IDs where applicable (T### format)
- Traceability: `[Spec §X]` = spec.md section, `[Contract LN]` = groups.yaml line, `[Gap]` = missing requirement
- For deferred items: document in which feature the requirement will be addressed
- This checklist complements existing `api.md` (2026-02-02) which covers initial design requirements
