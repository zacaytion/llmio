# Feature Specification: Discussions & Comments

**Feature Branch**: `005-discussions`
**Created**: 2026-02-02
**Status**: Draft
**Input**: User description: "Create next feature (discussions) following core domain foundation plan, using git worktree for parallel development"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Start a Group Discussion (Priority: P1)

As a group member, I want to start a discussion within my group so that I can propose topics for collaborative conversation.

**Why this priority**: Core functionality - without the ability to create discussions, the entire feature has no value. This is the entry point for all collaborative conversation.

**Independent Test**: Can be fully tested by creating a group, adding a member, and verifying they can create a discussion with title and description. Delivers the foundational value of structured group conversations.

**Acceptance Scenarios**:

1. **Given** a user is a member of a group with `members_can_start_discussions` enabled, **When** they create a discussion with title "Q1 Planning" and a description, **Then** the discussion is created, visible to group members, and the user is recorded as author
2. **Given** a user is a member of a group with `members_can_start_discussions` disabled, **When** they attempt to create a discussion, **Then** the system denies the action with a clear error message
3. **Given** a user is a group admin, **When** they create a discussion regardless of the `members_can_start_discussions` flag, **Then** the discussion is created successfully (admins always have permission)

---

### User Story 2 - Reply with Comments (Priority: P2)

As a discussion participant, I want to reply to a discussion or to specific comments so that I can contribute to the conversation thread.

**Why this priority**: Comments are the primary interaction mechanism after discussion creation. Without comments, discussions are static documents rather than collaborative conversations.

**Independent Test**: Can be tested by creating a discussion, adding a comment, and verifying threaded reply capability up to max depth. Delivers the core value of asynchronous conversation.

**Acceptance Scenarios**:

1. **Given** a user has access to a discussion, **When** they post a comment with text content, **Then** the comment appears in the discussion timeline
2. **Given** a comment exists, **When** a user replies to that comment, **Then** the reply is nested under the parent comment
3. **Given** comments are nested at the maximum depth (configurable via `max_depth`), **When** a user replies to a comment at max depth, **Then** the reply appears at the same level (flattened) rather than nesting deeper
4. **Given** a user is the author of a comment, **When** they edit the comment, **Then** the comment body is updated and an `edited_at` timestamp is recorded

---

### User Story 3 - Close and Reopen Discussion (Priority: P3)

As a discussion author or group admin, I want to close a discussion when it has reached conclusion and optionally reopen it if further conversation is needed.

**Why this priority**: Provides discussion lifecycle management. Important for maintaining organized group spaces but not critical for basic functionality.

**Independent Test**: Can be tested by creating a discussion, closing it, verifying no new comments can be added, then reopening and confirming comments are enabled again.

**Acceptance Scenarios**:

1. **Given** a user is the discussion author, **When** they close the discussion, **Then** the discussion is marked as closed with a timestamp
2. **Given** a discussion is closed, **When** a user attempts to add a comment, **Then** the system prevents the comment and indicates the discussion is closed
3. **Given** a discussion is closed, **When** the author or admin reopens it, **Then** the closed timestamp is cleared and comments can be added again
4. **Given** a user is not the author or an admin, **When** they attempt to close a discussion, **Then** the action is denied

---

### User Story 4 - Direct Discussions (No Group) (Priority: P4)

As a user, I want to start a private discussion with specific people without requiring a group, so that I can have focused conversations.

**Why this priority**: Extends discussions beyond group boundaries. Useful for private coordination but not essential for group-based collaboration.

**Independent Test**: Can be tested by creating a discussion without a group, inviting specific participants, and verifying only those participants can access it.

**Acceptance Scenarios**:

1. **Given** a user wants to start a private conversation, **When** they create a discussion without selecting a group and specify participant emails/usernames, **Then** only the author and specified participants can view and participate
2. **Given** a direct discussion exists, **When** someone not in the participant list attempts to access it, **Then** access is denied
3. **Given** a direct discussion exists, **When** the author adds a new participant, **Then** that participant gains access to view and comment

---

### User Story 5 - Read Tracking (Priority: P5)

As a discussion participant, I want the system to track what I have read so that I can easily identify new content.

**Why this priority**: Enhances user experience but not critical for core functionality. Can be added after basic discussion/comment features work.

**Independent Test**: Can be tested by opening a discussion, verifying last-read timestamp is recorded, adding new comments, and confirming unread indicators appear.

**Acceptance Scenarios**:

1. **Given** a user opens a discussion they have access to, **When** the page loads, **Then** the system records the current timestamp as their `last_read_at` for that discussion
2. **Given** a user has read a discussion, **When** new comments are added after their `last_read_at`, **Then** the API response includes an unread comment count (frontend visual display deferred)
3. **Given** a user sets their notification volume for a discussion to "mute", **When** new activity occurs, **Then** the user does not receive notifications but the discussion is still accessible

---

### Edge Cases

- What happens when a comment's parent is soft-deleted? → The reply remains visible as a top-level reply in its subtree
- How does system handle concurrent edits to the same comment? → Last write wins; consider adding optimistic locking in future
- What happens when a group is archived? → Discussions remain accessible but read-only; no new discussions can be created
- How are @mentions in discussion/comment bodies handled? → Captured as part of rich text; mention events are a future feature (006-events)
- What if max_depth is set to 0? → All comments appear at root level (flat discussion)
- What if parent_id references a comment in a different discussion? → Returns 404 (parent not found in this discussion)
- Can a soft-deleted comment be edited? → No, returns 404 (comment effectively no longer exists for editing)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow users to create discussions within groups they are members of, subject to the group's `members_can_start_discussions` permission flag
- **FR-002**: System MUST allow group admins to create discussions regardless of the `members_can_start_discussions` flag
- **FR-003**: System MUST support discussions without a group ("direct discussions") where access is limited to explicitly invited participants
- **FR-004**: System MUST require a title for all discussions; description is optional
- **FR-005**: System MUST allow users with access to add comments to open discussions
- **FR-006**: System MUST support comment threading with replies nested under parent comments
- **FR-007**: System MUST enforce a configurable maximum nesting depth (`max_depth`, default 3); replies beyond this depth MUST appear flattened at the max depth level
- **FR-008**: System MUST allow comment authors to edit their own comments and record the edit timestamp
- **FR-009**: System MUST allow comment deletion via soft delete (by comment author or group admin), displaying "[deleted]" placeholder while preserving reply structure
- **FR-010**: System MUST allow discussion authors and group admins to close discussions, preventing new comments
- **FR-011**: System MUST allow discussion authors and group admins to reopen closed discussions
- **FR-012**: System MUST track per-user, per-discussion read state including `last_read_at` timestamp
- **FR-013**: System MUST allow users to set their notification volume for specific discussions (mute/normal/loud)
- **FR-014**: System MUST enforce that only participants of a direct discussion can view or interact with it
- **FR-015**: System MUST support adding participants to direct discussions by the discussion author
- **FR-016**: System MUST default discussions to private (`private = true`) when created

### Key Entities

- **Discussion**: A conversation thread with a title, optional description, optional group association, author, and configurable reply depth. May be private to a group or direct (participant-only).
- **Comment**: A reply within a discussion, authored by a user, optionally nested under a parent comment. Supports editing and soft deletion.
- **DiscussionReader**: Tracks a user's read state for a specific discussion, including last-read timestamp and notification volume preference.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can create a discussion and add their first comment within 60 seconds *(UX metric; applies when frontend exists)*
- **SC-002**: System correctly enforces group permission flags in 100% of access attempts (verified via permission matrix tests)
- **SC-003**: Comment threading displays correctly with up to 10 levels of nesting (respecting max_depth configuration) *(backend enforces depth; frontend display deferred)*
- **SC-004**: p95 of read-state updates complete within 500ms of user opening a discussion (measured server-side from request receipt to response sent)
- **SC-005**: Direct discussions are never visible to non-participants (verified via security tests)
- **SC-006**: Soft-deleted comments retain their children (reply structure preserved) and display "[deleted]" placeholder
- **SC-007**: Closed discussions prevent 100% of comment creation attempts (verified via integration tests)

## Clarifications

### Session 2026-02-02

- Q: Who can delete comments (soft delete)? → A: Authors can delete their own comments; group admins can delete any comment in their group's discussions
- Q: What is the default value for max_depth? → A: Default 3 (moderate threading, matches Loomio)

## Assumptions

- Feature 004 (Groups & Memberships) is complete and provides the permission flag infrastructure
- Rich text content (discussion descriptions, comment bodies) will be stored as plain text or markdown for MVP; advanced rich text (Yjs/Hocuspocus) is a future enhancement
- Email notifications for new comments/discussions are out of scope (future 009-notifications feature)
- Real-time updates (live comment appearance) are out of scope (future real-time feature)
- @mention parsing and events are out of scope (will be handled by 006-events)

## Dependencies

- **Feature 004**: Groups & Memberships - provides group entities and permission flags
- **Feature 001**: User Authentication - provides authenticated user context

## Out of Scope

- Real-time comment streaming (WebSocket/Socket.io integration)
- Email notifications for new activity
- Polls attached to discussions (future feature)
- @mention events and notifications
- Rich text collaboration (Yjs/Hocuspocus)
- Search within discussions
