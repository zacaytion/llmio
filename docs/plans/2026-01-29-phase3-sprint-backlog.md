# Loomio Go Rewrite: Sprint Backlog

**Date:** 2026-01-29
**Status:** Proposed
**Phase:** 3 - Detailed Plan Creation

## Document Purpose

This document breaks down the WBS into 2-week sprints with:
- Sprint goals
- User stories with acceptance criteria
- Technical tasks
- Testing requirements
- Definition of done

**Total sprints:** 20 sprints (~40 weeks)
**Team capacity:** 52 SP/sprint (6.5 FTE × 4 SP/week × 2 weeks)

---

## Sprint Calendar Overview

| Sprint | Weeks | Phase | Focus Area | SP |
|--------|-------|-------|------------|------|
| 1-2 | 1-4 | 0 | Foundation & Infrastructure | 168 |
| 3 | 5-6 | 1 | Channel Server (WebSocket) | 56 |
| 4 | 7-8 | 1 | Channel Server (Bot + Testing) | 56 |
| 5-6 | 9-12 | 2 | Auth & Users Read API | 91 |
| 7-8 | 13-16 | 2 | Groups & Discussions Read API | 105 |
| 9 | 17-18 | 2 | Polls Read API | 56 |
| 10-11 | 19-22 | 3 | Permission System | 84 |
| 12-13 | 23-26 | 3 | Users/Groups Write | 112 |
| 14-15 | 27-30 | 3 | Discussions Write | 84 |
| 16-17 | 31-34 | 3 | Polls Write + Notifications | 168 |
| 18 | 35-36 | 3 | Email + Files + Search | 168 |
| 19 | 37-38 | 4 | Integration & Security | 154 |
| 20 | 39-40 | 5 | Beta & Launch | 168 |

---

## Sprint 1: Foundation Part 1

**Dates:** Weeks 1-2
**Sprint Goal:** Establish project structure, CI pipeline, and database connectivity

### User Stories

#### US-1.1: Project Initialization
**As a** developer
**I want** a working Go project with standard structure
**So that** I can start implementing features

**Acceptance Criteria:**
- [ ] Go module initialized with proper module path
- [ ] Directory structure follows golang-standards layout
- [ ] Makefile with build, test, lint commands
- [ ] README with setup instructions

**Tasks:**
- 0.1.1 Go module initialization (3 SP)
- 0.1.5 Makefile and dev scripts (4 SP)

#### US-1.2: Configuration Management
**As a** operator
**I want** environment-based configuration
**So that** I can deploy to different environments

**Acceptance Criteria:**
- [ ] All config via environment variables
- [ ] Sensible defaults for development
- [ ] Validation on startup for required vars
- [ ] Documentation of all config options

**Tasks:**
- 0.1.2 Configuration management (5 SP)

#### US-1.3: Observability Foundation
**As a** operator
**I want** structured logging
**So that** I can debug issues in production

**Acceptance Criteria:**
- [ ] slog-based structured logging
- [ ] Request IDs propagated through handlers
- [ ] Log levels configurable via env
- [ ] JSON output for production

**Tasks:**
- 0.1.3 Logging setup (3 SP)
- 0.1.4 Error handling patterns (5 SP)

#### US-1.4: CI Pipeline
**As a** developer
**I want** automated testing and linting
**So that** code quality is maintained

**Acceptance Criteria:**
- [ ] GitHub Actions runs on every PR
- [ ] golangci-lint with strict config
- [ ] Tests run with coverage reporting
- [ ] Build fails on lint errors or test failures

**Tasks:**
- 0.2.1 Lint workflow (3 SP)
- 0.2.2 Unit test workflow (5 SP)
- 0.2.4 Docker image build (5 SP)

#### US-1.5: Database Connectivity
**As a** developer
**I want** PostgreSQL connection with sqlc
**So that** I can write type-safe queries

**Acceptance Criteria:**
- [ ] pgx connection pool configured
- [ ] sqlc generates code from schema
- [ ] JSONB, arrays, citext handled correctly
- [ ] Transaction helper functions work

**Tasks:**
- 0.3.1 pgx connection pool (5 SP)
- 0.3.2 sqlc configuration (8 SP)

### Sprint 1 Totals
**Story Points:** 46 SP
**Testing Requirements:** Unit tests for config, logging, error handling
**Definition of Done:** All tasks complete, CI green, code reviewed

---

## Sprint 2: Foundation Part 2

**Dates:** Weeks 3-4
**Sprint Goal:** Complete database layer, HTTP server, and background jobs

### User Stories

#### US-2.1: Core Database Queries
**As a** developer
**I want** sqlc queries for core tables
**So that** I can implement API endpoints

**Acceptance Criteria:**
- [ ] Users table queries (CRUD operations)
- [ ] Groups table queries (with hierarchy)
- [ ] Factory pattern for test data
- [ ] All queries compile without errors

**Tasks:**
- 0.3.3 Core table queries (10 SP)
- 0.3.6 Database test fixtures (4 SP)

#### US-2.2: Migration Tooling
**As a** developer
**I want** database migration management
**So that** I can evolve the schema safely

**Acceptance Criteria:**
- [ ] goose or golang-migrate configured
- [ ] Can run up/down migrations
- [ ] Migrations tested in CI
- [ ] Transaction wrapping for safety

**Tasks:**
- 0.3.4 Database migrations tooling (8 SP)
- 0.3.5 Transaction management (5 SP)

#### US-2.3: HTTP Server Foundation
**As a** developer
**I want** Chi router with middleware
**So that** I can implement API endpoints

**Acceptance Criteria:**
- [ ] Chi router configured
- [ ] Logging, recovery, CORS middleware
- [ ] Health check endpoints working
- [ ] Request ID in all responses

**Tasks:**
- 0.4.1 Chi router setup (3 SP)
- 0.4.2 Middleware stack (8 SP)
- 0.4.3 Health check endpoints (2 SP)
- 0.4.4 Request/response helpers (4 SP)
- 0.4.5 OpenAPI spec generation (3 SP)

#### US-2.4: Background Jobs
**As a** developer
**I want** River job queue
**So that** I can process async work

**Acceptance Criteria:**
- [ ] River configured with PostgreSQL
- [ ] Sample job can be enqueued and processed
- [ ] Job monitoring available
- [ ] Failed jobs retry correctly

**Tasks:**
- 0.5.1 River setup (5 SP)
- 0.5.2 Job registration patterns (5 SP)
- 0.5.3 Job monitoring (5 SP)

#### US-2.5: Integration Test Infrastructure
**As a** developer
**I want** testcontainers integration tests
**So that** I can test against real databases

**Acceptance Criteria:**
- [ ] testcontainers-go configured
- [ ] PostgreSQL container starts in tests
- [ ] Redis container available
- [ ] CI runs integration tests

**Tasks:**
- 0.2.3 Integration test workflow (8 SP)
- 0.2.5 Contract test integration (4 SP)

### Sprint 2 Totals
**Story Points:** 74 SP
**Testing Requirements:** Integration tests for database, HTTP handlers
**Definition of Done:** All foundation complete, ready for feature work

---

## Sprint 3: Channel Server WebSocket

**Dates:** Weeks 5-6
**Sprint Goal:** Implement WebSocket server for live updates (records.js replacement)

### User Stories

#### US-3.1: WebSocket Server
**As a** user
**I want** real-time updates when content changes
**So that** I see new comments/votes immediately

**Acceptance Criteria:**
- [ ] WebSocket server listens on configured port
- [ ] Clients can connect with auth token
- [ ] Room-based broadcasting (user-{id}, group-{id})
- [ ] Connection handling (join, leave, reconnect)

**Tasks:**
- 1.1.1 WebSocket server setup (8 SP)
- 1.1.2 Room-based broadcasting (10 SP)
- 1.1.5 Connection management (8 SP)

#### US-3.2: Redis Pub/Sub Integration
**As a** Go server
**I want** to receive events from Rails via Redis
**So that** I can broadcast to connected clients

**Acceptance Criteria:**
- [ ] Subscribed to /records channel
- [ ] Messages parsed and routed to correct rooms
- [ ] User session lookup from Redis works
- [ ] Handles malformed messages gracefully

**Tasks:**
- 1.1.3 Redis pub/sub subscription (8 SP)
- 1.1.4 User authentication via Redis (6 SP)

### Sprint 3 Totals
**Story Points:** 40 SP
**Testing Requirements:** WebSocket integration tests, connection tests
**Definition of Done:** WebSocket server functional, can receive from Rails

---

## Sprint 4: Channel Server Bot + Testing

**Dates:** Weeks 7-8
**Sprint Goal:** Implement Matrix bot and complete channel server testing

### User Stories

#### US-4.1: Matrix Bot Integration
**As a** team
**I want** notifications sent to Matrix rooms
**So that** I stay informed in my chat tool

**Acceptance Criteria:**
- [ ] mautrix/go SDK configured
- [ ] Bot connects to Matrix homeserver
- [ ] Messages sent to configured rooms
- [ ] Multiple bot configurations supported

**Tasks:**
- 1.2.1 mautrix/go integration (8 SP)
- 1.2.2 Redis pub/sub for chatbot/* (5 SP)
- 1.2.3 Room resolution and messaging (8 SP)
- 1.2.4 Bot configuration management (4 SP)

#### US-4.2: Channel Server Verification
**As a** operator
**I want** confidence that Go channel server matches Node.js behavior
**So that** I can migrate without user impact

**Acceptance Criteria:**
- [ ] Integration tests pass with real Redis
- [ ] Load test shows equal or better performance
- [ ] Matrix bot tests pass with mocked API
- [ ] Shadow testing plan documented

**Tasks:**
- 1.3.1 WebSocket integration tests (8 SP)
- 1.3.2 Load testing (5 SP)
- 1.3.3 Matrix bot mock tests (2 SP)

### Sprint 4 Totals
**Story Points:** 40 SP
**Testing Requirements:** Full channel server test suite
**Definition of Done:** Channel server ready for shadow deployment

---

## Sprint 5: Authentication Foundation

**Dates:** Weeks 9-10
**Sprint Goal:** Implement core authentication (login, JWT, sessions)

### User Stories

#### US-5.1: JWT Authentication
**As a** user
**I want** to log in and receive a token
**So that** I can access protected resources

**Acceptance Criteria:**
- [ ] JWT tokens generated with correct claims
- [ ] Token validation middleware works
- [ ] Refresh token flow implemented
- [ ] Token expiration handled correctly

**Tasks:**
- 2.1.1 JWT generation/validation (8 SP)
- 2.1.2 Session cookie handling (8 SP)

#### US-5.2: Password Authentication
**As a** user
**I want** to log in with email and password
**So that** I can access my account

**Acceptance Criteria:**
- [ ] bcrypt password verification
- [ ] Login endpoint returns JWT
- [ ] Failed login attempts tracked
- [ ] Existing password hashes compatible

**Tasks:**
- 2.1.3 Password hashing (3 SP)
- 2.1.4 Login endpoint (8 SP)

#### US-5.3: Magic Link Login
**As a** user
**I want** to log in via email link
**So that** I don't need to remember a password

**Acceptance Criteria:**
- [ ] Magic link generation endpoint
- [ ] Token verification endpoint
- [ ] Links expire appropriately
- [ ] One-time use enforced

**Tasks:**
- 2.1.5 Magic link authentication (8 SP)

### Sprint 5 Totals
**Story Points:** 43 SP
**Testing Requirements:** Auth flow integration tests
**Definition of Done:** Users can log in via password or magic link

---

## Sprint 6: Users Read API

**Dates:** Weeks 11-12
**Sprint Goal:** Implement user read endpoints with API contract parity

### User Stories

#### US-6.1: User Profile Endpoints
**As a** user
**I want** to view my profile and other users' profiles
**So that** I can manage my information

**Acceptance Criteria:**
- [ ] GET /api/v1/profile returns current user
- [ ] GET /api/v1/users/:id returns user detail
- [ ] Response format matches Rails exactly
- [ ] Contract tests pass

**Tasks:**
- 2.2.1 User model and queries (5 SP)
- 2.2.2 GET /api/v1/profile (5 SP)
- 2.2.3 GET /api/v1/users/:id (5 SP)
- 2.2.5 User serializer (5 SP)

#### US-6.2: User Search
**As a** user
**I want** to search for users to mention
**So that** I can notify them in discussions

**Acceptance Criteria:**
- [ ] Search by name/email prefix
- [ ] Results scoped to groups user can see
- [ ] Fast response (<100ms)

**Tasks:**
- 2.2.4 GET /api/v1/users/search (5 SP)

#### US-6.3: OAuth Provider Reading
**As a** user
**I want** my linked OAuth accounts available
**So that** I can manage my login methods

**Acceptance Criteria:**
- [ ] Identity records retrieved correctly
- [ ] Multiple providers supported
- [ ] Included in profile response

**Tasks:**
- 2.1.6 OAuth provider integration (5 SP)

### Sprint 6 Totals
**Story Points:** 30 SP
**Testing Requirements:** Contract tests for all user endpoints
**Definition of Done:** All user read endpoints match Rails API exactly

---

## Sprint 7: Groups Read API

**Dates:** Weeks 13-14
**Sprint Goal:** Implement groups and memberships read endpoints

### User Stories

#### US-7.1: Groups Listing
**As a** user
**I want** to see my groups
**So that** I can navigate to discussions

**Acceptance Criteria:**
- [ ] GET /api/v1/groups returns user's groups
- [ ] Permission filtering works
- [ ] Includes membership info
- [ ] Hierarchical groups handled

**Tasks:**
- 2.3.1 Group model and queries (8 SP)
- 2.3.2 Membership model and queries (5 SP)
- 2.3.3 GET /api/v1/groups (8 SP)

#### US-7.2: Group Details
**As a** user
**I want** to view group details
**So that** I can see settings and members

**Acceptance Criteria:**
- [ ] GET /api/v1/groups/:id returns detail
- [ ] Subgroups endpoint works
- [ ] Response matches Rails format
- [ ] Permissions respected

**Tasks:**
- 2.3.4 GET /api/v1/groups/:id (5 SP)
- 2.3.5 GET /api/v1/groups/:id/subgroups (4 SP)
- 2.3.6 Group serializer (5 SP)

### Sprint 7 Totals
**Story Points:** 35 SP
**Testing Requirements:** Contract tests, permission tests
**Definition of Done:** Groups API matches Rails with correct permissions

---

## Sprint 8: Discussions Read API

**Dates:** Weeks 15-16
**Sprint Goal:** Implement discussions and events read endpoints

### User Stories

#### US-8.1: Discussions Listing
**As a** user
**I want** to see discussions in my groups
**So that** I can participate in conversations

**Acceptance Criteria:**
- [ ] GET /api/v1/discussions returns filtered list
- [ ] Permission-based filtering works
- [ ] Pagination implemented
- [ ] Response matches Rails format

**Tasks:**
- 2.4.1 Discussion model and queries (8 SP)
- 2.4.3 GET /api/v1/discussions (8 SP)

#### US-8.2: Discussion Detail with Timeline
**As a** user
**I want** to see discussion with threaded comments
**So that** I can follow the conversation

**Acceptance Criteria:**
- [ ] GET /api/v1/discussions/:id returns discussion
- [ ] Events (comments) loaded with correct threading
- [ ] position_key ordering correct
- [ ] Response matches Rails format

**Tasks:**
- 2.4.2 Event model and queries (10 SP) **[HIGH COMPLEXITY]**
- 2.4.4 GET /api/v1/discussions/:id (5 SP)
- 2.4.5 GET /api/v1/events (5 SP)
- 2.4.6 Discussion/Event serializers (4 SP)

### Sprint 8 Totals
**Story Points:** 40 SP
**Testing Requirements:** Event threading tests (property-based)
**Definition of Done:** Discussion timeline matches Rails exactly

---

## Sprint 9: Polls Read API

**Dates:** Weeks 17-18
**Sprint Goal:** Implement polls, stances, and outcomes read endpoints

### User Stories

#### US-9.1: Polls Listing
**As a** user
**I want** to see polls in my groups
**So that** I can participate in decisions

**Acceptance Criteria:**
- [ ] GET /api/v1/polls returns filtered list
- [ ] All poll types supported
- [ ] Permission filtering works
- [ ] Response matches Rails format

**Tasks:**
- 2.5.1 Poll model and queries (10 SP)
- 2.5.4 GET /api/v1/polls (8 SP)

#### US-9.2: Poll Detail with Stances
**As a** user
**I want** to see poll with current votes
**So that** I can make an informed decision

**Acceptance Criteria:**
- [ ] GET /api/v1/polls/:id returns poll
- [ ] Stances and choices included
- [ ] Outcome included if set
- [ ] All poll types render correctly

**Tasks:**
- 2.5.2 Stance/StanceChoice models (8 SP)
- 2.5.3 Outcome model (4 SP)
- 2.5.5 GET /api/v1/polls/:id (5 SP)
- 2.5.6 Poll serializers (5 SP)

#### US-9.3: Supporting Endpoints
**As a** user
**I want** notifications and memberships
**So that** I stay informed

**Acceptance Criteria:**
- [ ] Notifications endpoint works
- [ ] Memberships endpoint works
- [ ] Tags and translations endpoints work

**Tasks:**
- 2.6.1-2.6.4 Supporting endpoints (20 SP)

### Sprint 9 Totals
**Story Points:** 60 SP
**Testing Requirements:** Contract tests for all poll types
**Definition of Done:** All read APIs complete and tested

---

## Sprints 10-11: Permission System

**Dates:** Weeks 19-22
**Sprint Goal:** Implement comprehensive permission system

### User Stories

#### US-10.1: Permission Matrix Extraction
**As a** developer
**I want** all Rails abilities documented
**So that** I can implement them in Go

**Acceptance Criteria:**
- [ ] All CanCanCan rules extracted
- [ ] Matrix document created
- [ ] Edge cases identified
- [ ] Test cases defined

**Tasks:**
- 3.1.1 Extract CanCanCan rules (15 SP)

#### US-10.2: Permission Implementation
**As a** system
**I want** permission checks on all endpoints
**So that** users only see what they're allowed

**Acceptance Criteria:**
- [ ] Permission middleware works
- [ ] All endpoints protected
- [ ] Group inheritance works
- [ ] 403 responses for unauthorized

**Tasks:**
- 3.1.2 Implement permission checking (20 SP)
- 3.1.3 Group permission inheritance (15 SP)
- 3.1.4 Permission test suite (10 SP)

### Sprints 10-11 Totals
**Story Points:** 60 SP
**Testing Requirements:** Permission tests for all endpoints
**Definition of Done:** All endpoints permission-protected

---

## Sprints 12-13: Users/Groups Write Operations

**Dates:** Weeks 23-26
**Sprint Goal:** Implement write operations for users and groups

### User Stories

#### US-12.1: User Registration and Management
**As a** user
**I want** to register and manage my account
**So that** I can participate in Loomio

**Acceptance Criteria:**
- [ ] Registration endpoint works
- [ ] Profile update works
- [ ] Password change/reset works
- [ ] Account deletion works

**Tasks:**
- 3.2.1-3.2.4 Users write operations (30 SP)

#### US-12.2: Group Management
**As a** admin
**I want** to create and manage groups
**So that** my team can collaborate

**Acceptance Criteria:**
- [ ] Group creation works
- [ ] Settings updates work
- [ ] Subgroup creation works
- [ ] Archive/delete works

**Tasks:**
- 3.3.1-3.3.6 Groups write operations (50 SP)

### Sprints 12-13 Totals
**Story Points:** 80 SP
**Testing Requirements:** Write operation integration tests
**Definition of Done:** User and group management fully functional

---

## Sprints 14-15: Discussions Write Operations

**Dates:** Weeks 27-30
**Sprint Goal:** Implement discussion and comment write operations

### User Stories

#### US-14.1: Discussion Management
**As a** user
**I want** to create and edit discussions
**So that** I can start conversations

**Acceptance Criteria:**
- [ ] Discussion creation works
- [ ] Discussion editing works
- [ ] Templates can be used
- [ ] Permissions enforced

**Tasks:**
- 3.4.1 POST /api/v1/discussions (10 SP)
- 3.4.2 PUT /api/v1/discussions/:id (8 SP)
- 3.4.6 Discussion templates (8 SP)

#### US-14.2: Comment Threading
**As a** user
**I want** to post comments with proper threading
**So that** conversations stay organized

**Acceptance Criteria:**
- [ ] Comments create correctly
- [ ] position_key generated correctly
- [ ] Thread depth handled
- [ ] Edit and delete work

**Tasks:**
- 3.4.3 POST /api/v1/comments (12 SP) **[HIGH COMPLEXITY]**
- 3.4.4 PUT /api/v1/comments/:id (6 SP)
- 3.4.5 DELETE /api/v1/comments/:id (5 SP)
- 3.4.7 Event position_key generation (11 SP) **[HIGH COMPLEXITY]**

### Sprints 14-15 Totals
**Story Points:** 60 SP
**Testing Requirements:** Threading property tests
**Definition of Done:** Discussion threading works identically to Rails

---

## Sprints 16-17: Polls Write + Notifications

**Dates:** Weeks 31-34
**Sprint Goal:** Implement poll voting and notification system

### User Stories

#### US-16.1: Poll Creation and Voting
**As a** user
**I want** to create polls and cast votes
**So that** my group can make decisions

**Acceptance Criteria:**
- [ ] All poll types can be created
- [ ] Voting works for all poll types
- [ ] Vote tallying correct
- [ ] Outcomes can be recorded

**Tasks:**
- 3.5.1-3.5.7 Polls write operations (80 SP)

#### US-16.2: Notification System
**As a** user
**I want** to receive notifications
**So that** I stay informed of activity

**Acceptance Criteria:**
- [ ] Notifications generated on events
- [ ] Pushed via WebSocket
- [ ] Preferences respected
- [ ] Mark read/unread works

**Tasks:**
- 3.6.1-3.6.5 Notifications (40 SP)

### Sprints 16-17 Totals
**Story Points:** 120 SP
**Testing Requirements:** Voting algorithm property tests
**Definition of Done:** All poll types functional, notifications working

---

## Sprint 18: Email + Files + Search

**Dates:** Weeks 35-36
**Sprint Goal:** Implement email, file handling, and search

### User Stories

#### US-18.1: Email System
**As a** user
**I want** to receive email notifications
**So that** I stay informed outside the app

**Acceptance Criteria:**
- [ ] Transactional emails sent
- [ ] Notification emails work
- [ ] Digests compile and send
- [ ] Inbound email parsing works

**Tasks:**
- 3.7.1-3.7.5 Email system (50 SP)

#### US-18.2: File Handling
**As a** user
**I want** to attach files to discussions
**So that** I can share documents

**Acceptance Criteria:**
- [ ] File uploads work (S3/GCS)
- [ ] Images processed (thumbnails)
- [ ] Attachments link to discussions
- [ ] Download works

**Tasks:**
- 3.8.1-3.8.4 File handling (40 SP)

#### US-18.3: Search
**As a** user
**I want** to search across content
**So that** I can find past discussions

**Acceptance Criteria:**
- [ ] Full-text search works
- [ ] Scoped search works
- [ ] Permission filtering applied
- [ ] Reasonable performance

**Tasks:**
- 3.9.1-3.9.3 Search (30 SP)

#### US-18.4: SAML/SSO
**As an** enterprise admin
**I want** SAML SSO
**So that** my organization can use central auth

**Acceptance Criteria:**
- [ ] SAML SP works
- [ ] OAuth providers work
- [ ] Tested with common IdPs

**Tasks:**
- 3.10.1-3.10.3 SAML/SSO (40 SP)

### Sprint 18 Totals
**Story Points:** 160 SP (overflow to previous sprints)
**Testing Requirements:** Email tests, file upload tests, search tests
**Definition of Done:** All Phase 3 features complete

---

## Sprint 19: Integration & Security

**Dates:** Weeks 37-38
**Sprint Goal:** Frontend integration, security hardening

### User Stories

#### US-19.1: API Contract Verification
**As a** operator
**I want** confidence Go API matches Rails
**So that** the frontend works unchanged

**Acceptance Criteria:**
- [ ] All endpoints contract tested
- [ ] Any discrepancies fixed
- [ ] Shadow traffic comparison done

**Tasks:**
- 4.1.1-4.1.4 Frontend integration (40 SP)

#### US-19.2: Data Migration
**As a** operator
**I want** to migrate data safely
**So that** users don't lose anything

**Acceptance Criteria:**
- [ ] Migration scripts work
- [ ] Data validated after migration
- [ ] Rollback tested
- [ ] Dry run on prod copy complete

**Tasks:**
- 4.2.1-4.2.4 Data migration (50 SP)

#### US-19.3: Performance & Security
**As a** operator
**I want** performance and security verified
**So that** the system is production-ready

**Acceptance Criteria:**
- [ ] Performance meets baselines
- [ ] Security audit complete
- [ ] Rate limiting in place
- [ ] Input validation reviewed

**Tasks:**
- 4.3.1-4.3.4 Performance (40 SP)
- 4.4.1-4.4.3 Security (30 SP)

### Sprint 19 Totals
**Story Points:** 160 SP
**Testing Requirements:** E2E tests, load tests, security tests
**Definition of Done:** System ready for beta

---

## Sprint 20: Beta & Launch

**Dates:** Weeks 39-40
**Sprint Goal:** Deploy to production, migrate traffic

### User Stories

#### US-20.1: Beta Deployment
**As a** operator
**I want** to deploy Go version alongside Rails
**So that** I can verify it works

**Acceptance Criteria:**
- [ ] Shadow mode deployed
- [ ] Canary traffic working
- [ ] Monitoring dashboards ready
- [ ] Beta feedback collected

**Tasks:**
- 5.1.1-5.1.4 Beta deployment (40 SP)

#### US-20.2: Production Migration
**As a** operator
**I want** to migrate all traffic to Go
**So that** we complete the rewrite

**Acceptance Criteria:**
- [ ] Gradual rollout complete
- [ ] Rails sunset verified
- [ ] Documentation complete
- [ ] Team can operate Go version

**Tasks:**
- 5.2.1-5.2.5 Production migration (50 SP)
- 5.3.1-5.3.4 Documentation (30 SP)

### Sprint 20 Totals
**Story Points:** 120 SP
**Testing Requirements:** Production verification
**Definition of Done:** Go version serving all traffic

---

## Definition of Done (Global)

Every sprint must meet these criteria:

- [ ] All acceptance criteria for stories met
- [ ] Unit test coverage > 80% for new code
- [ ] Integration tests pass
- [ ] Contract tests pass (where applicable)
- [ ] Code reviewed by at least one other developer
- [ ] No golangci-lint warnings
- [ ] Documentation updated
- [ ] Sprint demo completed
- [ ] Retrospective held

---

## Velocity Tracking Template

| Sprint | Planned SP | Completed SP | Velocity | Notes |
|--------|------------|--------------|----------|-------|
| 1 | 46 | - | - | |
| 2 | 74 | - | - | |
| ... | ... | ... | ... | |

---

## Risk Monitoring per Sprint

Each sprint review should check:

1. Are we on track with the timeline? (Risk #9)
2. Any API contract issues found? (Risk #2)
3. Permission edge cases discovered? (Risk #4)
4. Event threading working correctly? (Risk #3)
5. Team Go proficiency improving? (Risk #6)
