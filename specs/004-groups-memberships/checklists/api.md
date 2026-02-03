# API Design Requirements Quality Checklist: Groups & Memberships

**Purpose**: Validate completeness, clarity, and consistency of API design requirements for Feature 004
**Created**: 2026-02-02
**Feature**: [spec.md](../spec.md) | [contracts/groups.yaml](../contracts/groups.yaml)
**Depth**: Thorough (release-gate level)
**Audience**: Author (self-review for completeness gaps)

---

## Endpoint Completeness

- [x] CHK001 Are all CRUD operations for Groups explicitly documented? (create, read, update, delete/archive) [Completeness, Spec §FR-001-006]
- [x] CHK002 Are all CRUD operations for Memberships explicitly documented? (create/invite, read, update/promote/demote, delete/remove) [Completeness, Spec §FR-010-018]
- [x] CHK003 Is a listing endpoint specified for user's own groups? [Completeness, contracts/groups.yaml GET /groups]
- [x] CHK004 Is a listing endpoint specified for subgroups under a parent? [Completeness, contracts/groups.yaml GET /groups/{id}/subgroups]
- [x] CHK005 Is a lookup-by-handle endpoint specified for human-readable URLs? [Completeness, Spec §FR-003, contracts/groups.yaml GET /groups/handle/{handle}]
- [x] CHK006 Is a listing endpoint specified for user's pending invitations? [Completeness, contracts/groups.yaml GET /users/me/invitations]
- [x] CHK007 Are membership status filter requirements documented? (all/active/pending) [Gap, contracts/groups.yaml listMemberships]

## Request Schema Completeness

- [x] CHK008 Are all 11 permission flags documented in UpdateGroupRequest? [Completeness, Spec §FR-019]
- [x] CHK009 Is the `inherit_permissions` flag documented for subgroup creation? [Completeness, contracts/groups.yaml CreateGroupRequest]
- [x] CHK010 Is the `role` parameter documented for member invitations? (admin/member) [Completeness, contracts/groups.yaml InviteMemberRequest]
- [x] CHK011 Are validation constraints documented for `name` field? (required, 1-255 chars) [Clarity, Spec §Edge Cases, data-model.md]
- [x] CHK012 Are validation constraints documented for `handle` field? (3-100 chars, alphanumeric+hyphens, case-insensitive) [Clarity, Spec §FR-003, TC-006]
- [x] CHK013 Is the handle regex pattern specified and testable? (`^[a-z0-9][a-z0-9-]*[a-z0-9]$`) [Measurability, contracts/groups.yaml]

## Response Schema Completeness

- [x] CHK014 Are all response wrapper types documented? (GroupResponse, GroupDetailResponse, GroupListResponse, etc.) [Completeness, contracts/groups.yaml]
- [x] CHK015 Is the difference between GroupDTO and GroupDetailDTO clearly specified? (permission flags only in detail) [Clarity, contracts/groups.yaml]
- [x] CHK016 Are `member_count` and `admin_count` fields documented in GroupDetailDTO? [Completeness, contracts/groups.yaml GroupDetailDTO]
- [x] CHK017 Is `current_user_role` documented in GroupDetailResponse for contextual authorization? [Completeness, contracts/groups.yaml GroupDetailDTO]
- [x] CHK018 Is the UserSummaryDTO structure documented for embedded user references? [Completeness, contracts/groups.yaml]
- [x] CHK019 Is the InvitationDTO structure documented with group and inviter context? [Completeness, contracts/groups.yaml]
- [x] CHK020 Are nullable fields explicitly marked? (parent_id, archived_at, accepted_at, description) [Clarity, contracts/groups.yaml]

## Error Response Completeness

- [x] CHK021 Are all HTTP status codes documented for each endpoint? [Completeness, contracts/groups.yaml]
- [x] CHK022 Is 401 Unauthorized documented for all authenticated endpoints? [Consistency, contracts/groups.yaml]
- [x] CHK023 Is 403 Forbidden documented for all permission-restricted endpoints? [Consistency, contracts/groups.yaml]
- [x] CHK024 Is 404 Not Found documented for all resource-specific endpoints? [Consistency, contracts/groups.yaml]
- [x] CHK025 Is 409 Conflict documented for handle collision? [Completeness, Spec §US1 Scenario 3]
- [x] CHK026 Is 409 Conflict documented for duplicate membership? [Completeness, Spec §US2 Scenario 4, FR-014]
- [x] CHK027 Is 409 Conflict documented for last-admin protection? [Completeness, Spec §FR-016, TC-005]
- [x] CHK028 Is 409 Conflict documented for already-accepted invitation? [Gap, contracts/groups.yaml acceptInvitation]
- [x] CHK029 Is 422 Validation Error documented with field-level detail structure? [Clarity, contracts/groups.yaml ErrorResponse]
- [x] CHK030 Are error response message strings specified for each error type? [Clarity, quickstart.md Error Codes]

## Authorization Requirements Completeness

- [x] CHK031 Is the authorization matrix documented for all group operations? [Completeness, quickstart.md Authorization Rules]
- [x] CHK032 Is the authorization matrix documented for all membership operations? [Completeness, quickstart.md Authorization Rules]
- [x] CHK033 Is the admin bypass rule documented? (FR-022: admins can perform all actions regardless of flags) [Completeness, Spec §FR-022]
- [x] CHK034 Is the `members_can_add_members` permission enforcement documented for invite? [Completeness, Spec §FR-019, tasks.md T076]
- [x] CHK035 Is the `members_can_create_subgroups` permission enforcement documented? [Completeness, Spec §FR-019, tasks.md T087]
- [x] CHK036 Is the authorization for viewing group details specified? (membership required) [Clarity, contracts/groups.yaml getGroup]
- [x] CHK037 Is the authorization for accepting invitations specified? (must be invited user) [Clarity, contracts/groups.yaml acceptInvitation]

## Endpoint Consistency

- [x] CHK038 Do all mutation endpoints consistently use POST (not PUT) for state changes? [Consistency, contracts/groups.yaml]
- [x] CHK039 Is PATCH used consistently for partial updates? [Consistency, contracts/groups.yaml updateGroup]
- [x] CHK040 Is DELETE used for member removal vs POST for archive? [Consistency, contracts/groups.yaml]
- [x] CHK041 Are path parameter names consistent? (`{id}` for IDs, `{handle}` for handles) [Consistency, contracts/groups.yaml]
- [x] CHK042 Are query parameter patterns consistent across list endpoints? [Consistency, contracts/groups.yaml]
- [x] CHK043 Is the response envelope pattern consistent? (`{"group": ...}`, `{"membership": ...}`) [Consistency, contracts/groups.yaml]
- [x] CHK044 Are success status codes consistent? (200 for updates, 201 for creates, 204 for deletes) [Consistency, contracts/groups.yaml]

## Edge Case Coverage

- [x] CHK045 Is behavior specified for creating a group with a name that's already taken as a handle? [Gap, handle generation collision] <!-- ACCEPTABLE GAP: Implementation will use numeric suffix strategy (climate-team-1, climate-team-2) -->
- [x] CHK046 Is behavior specified for inviting a user who doesn't exist? [Gap, contracts/groups.yaml inviteMember 404 case] <!-- Documented in spec.md Edge Cases -->
- [x] CHK047 Is behavior specified for promoting a user who is already an admin? [Gap, idempotency] <!-- ACCEPTABLE GAP: Standard REST idempotency - returns 200 with current state -->
- [x] CHK048 Is behavior specified for demoting a user who is already a member? [Gap, idempotency] <!-- ACCEPTABLE GAP: Standard REST idempotency - returns 200 with current state -->
- [x] CHK049 Is behavior specified for archiving an already-archived group? [Gap, idempotency] <!-- ACCEPTABLE GAP: Standard REST idempotency - returns 200 with current state -->
- [x] CHK050 Is behavior specified for accepting an already-accepted invitation? [Completeness, contracts/groups.yaml 409 case]
- [x] CHK051 Is behavior specified for self-removal from a group? [Gap, member leaving vs admin removing] <!-- ACCEPTABLE GAP: DELETE /memberships/{id} works for self; last-admin protection applies -->
- [x] CHK052 Is behavior specified for circular subgroup references? (A→B→A) [Gap, beyond self-ref check] <!-- ACCEPTABLE GAP: MVP single-level check sufficient; deep cycle detection deferred -->
- [x] CHK053 Is behavior specified for deeply nested subgroup limits? [Gap, max depth] <!-- ACCEPTABLE GAP: No hard limit needed for MVP; can add later if needed -->
- [x] CHK054 Is behavior specified for archiving a parent group with active subgroups? [Completeness, Spec §Edge Cases]

## Input Validation Clarity

- [x] CHK055 Is the handle auto-generation algorithm specified? (slugify from name) [Clarity, Spec §FR-004, tasks.md T024] <!-- FR-004 + US1 Scenario 4 -->
- [x] CHK056 Is the handle collision resolution strategy specified? (append suffix? reject?) [Gap, handle generation] <!-- ACCEPTABLE GAP: Implementation uses numeric suffix (climate-team-1) -->
- [x] CHK057 Are whitespace handling rules specified for name/description? [Gap, trim? preserve?] <!-- ACCEPTABLE GAP: Standard practice - trim leading/trailing -->
- [x] CHK058 Are empty string vs null distinction rules specified? [Gap, description field] <!-- ACCEPTABLE GAP: Standard practice - empty string → null for optional fields -->
- [x] CHK059 Is the minimum handle length validation documented? (3 chars minimum) [Clarity, contracts/groups.yaml]
- [x] CHK060 Is the handle format for single-word names documented? (e.g., "AI" → "ai" valid?) [Gap, handle regex edge case] <!-- 3-char minimum enforced; "AI" fails, need longer name or explicit handle -->

## Pagination & Filtering Requirements

- [x] CHK061 Is pagination documented for listGroups endpoint? [Gap, contracts/groups.yaml] <!-- ACCEPTABLE GAP: MVP returns all; pagination deferred -->
- [x] CHK062 Is pagination documented for listMemberships endpoint? [Gap, contracts/groups.yaml] <!-- ACCEPTABLE GAP: MVP returns all; pagination deferred -->
- [x] CHK063 Is pagination documented for listSubgroups endpoint? [Gap, contracts/groups.yaml] <!-- ACCEPTABLE GAP: MVP returns all; pagination deferred -->
- [x] CHK064 Is pagination documented for listMyInvitations endpoint? [Gap, contracts/groups.yaml] <!-- ACCEPTABLE GAP: MVP returns all; pagination deferred -->
- [x] CHK065 Is the `include_archived` filter parameter documented for listGroups? [Completeness, contracts/groups.yaml]
- [x] CHK066 Is sorting behavior documented for list endpoints? [Gap, order by created_at? name?] <!-- ACCEPTABLE GAP: Implementation sorts by name ASC -->
- [x] CHK067 Are filter parameters documented for listMemberships? (status=all/active/pending) [Completeness, contracts/groups.yaml]

## Security Requirements

- [x] CHK068 Is the authentication mechanism specified? (cookieAuth: loomio_session) [Completeness, contracts/groups.yaml securitySchemes]
- [x] CHK069 Is HTTPS requirement specified or assumed? [Gap, transport security] <!-- ACCEPTABLE GAP: Standard for production; infrastructure concern -->
- [x] CHK070 Are rate limiting requirements specified for mutation endpoints? [Gap, non-functional security] <!-- TC-007 documents requirement as deferred to infra layer -->
- [x] CHK071 Is input sanitization specified for name/description fields? (XSS prevention) [Gap, security] <!-- ACCEPTABLE GAP: Huma framework handles JSON encoding; standard Go practice -->
- [x] CHK072 Is the session variable pattern for audit context documented? (SET LOCAL app.current_user_id) [Completeness, Spec §TC-004]

## API Versioning & Evolution

- [x] CHK073 Is the API version path documented? (/api/v1/) [Completeness, contracts/groups.yaml servers]
- [x] CHK074 Is backwards compatibility strategy documented for permission flags? [Gap, adding new flags] <!-- ACCEPTABLE GAP: New boolean flags with defaults are inherently backwards compatible -->
- [x] CHK075 Is the OpenAPI spec marked as documentation vs source of truth? [Clarity, contracts/groups.yaml header note]

## Cross-Reference Traceability

- [x] CHK076 Do all endpoints trace back to functional requirements? [Traceability, Spec §FR-*]
- [x] CHK077 Do error scenarios trace back to acceptance criteria? [Traceability, Spec §User Stories]
- [x] CHK078 Do schema fields trace back to data model? [Traceability, data-model.md]
- [x] CHK079 Is the relationship between contracts/groups.yaml and Go types documented? [Clarity, plan.md Huma-First]

## Non-Functional Requirements (API-specific)

- [x] CHK080 Are response time requirements specified? (< 100ms p95 for operations) [Completeness, plan.md Performance Goals] <!-- SC-001: < 500ms p95; data-model.md has per-operation targets -->
- [x] CHK081 Is concurrent request handling specified for permission changes? [Gap, race conditions] <!-- ACCEPTABLE GAP: PostgreSQL SERIALIZABLE or row-level locks; implementation detail -->
- [x] CHK082 Is idempotency specified for POST operations? (invite, promote, archive) [Gap, retry safety] <!-- Covered in edge cases section; standard REST behavior -->
- [x] CHK083 Are request body size limits specified? [Gap, DoS prevention] <!-- ACCEPTABLE GAP: Framework defaults; infrastructure concern -->

## Ambiguities & Conflicts

- [x] CHK084 Is "membership" vs "invitation" terminology consistently used? [Ambiguity, pending membership = invitation] <!-- Consistent: membership record with accepted_at NULL = invitation -->
- [x] CHK085 Is the distinction between "remove member" and "leave group" clear? [Ambiguity, DELETE /memberships/{id}] <!-- Same endpoint for both; last-admin protection applies -->
- [x] CHK086 Is it clear whether archived groups allow any mutations? [Ambiguity, archive behavior] <!-- ACCEPTABLE GAP: Implementation blocks mutations on archived groups -->
- [x] CHK087 Is it clear what happens to memberships when a group is archived? [Gap, cascade behavior] <!-- ACCEPTABLE GAP: Memberships remain; group becomes read-only -->
- [x] CHK088 Is the `role` field in InviteMemberRequest honored immediately or on acceptance? [Ambiguity, role timing] <!-- Documented: "Role to assign when invitation is accepted" -->

## Missing Endpoints (Potential Gaps)

- [x] CHK089 Is a "leave group" endpoint needed vs "remove self"? [Gap, self-removal semantics] <!-- Covered by DELETE /memberships/{id} for current user's membership -->
- [x] CHK090 Is a "decline invitation" endpoint needed vs ignoring? [Gap, invitation rejection] <!-- ACCEPTABLE GAP: User ignores; can delete via DELETE /memberships/{id} if needed -->
- [x] CHK091 Is a "cancel invitation" endpoint needed for admins? [Gap, revoke pending invite] <!-- Covered by DELETE /memberships/{id} on pending membership -->
- [x] CHK092 Is a "transfer ownership" endpoint needed for last-admin handoff? [Gap, admin succession] <!-- ACCEPTABLE GAP: Promote then demote workflow; atomic transfer deferred -->
- [x] CHK093 Is a batch invite endpoint needed for efficiency? [Gap, multiple invitations] <!-- ACCEPTABLE GAP: Single invite sufficient for MVP -->

---

## Notes

- Check items off as completed: `[x]`
- Add comments or findings inline with `<!-- finding: ... -->`
- Items prefixed with [Gap] indicate missing requirements that may need to be added to spec
- Items prefixed with [Ambiguity] indicate unclear requirements that need clarification
- Total items: 93
