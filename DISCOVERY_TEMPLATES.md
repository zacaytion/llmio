# Discovery Phase Templates & Checklists

This document provides ready-to-use templates for the discovery phase of the Loomio → Go rewrite project.

---

## Template 1: Feature Inventory Spreadsheet

Use this format in Google Sheets, Excel, or CSV:

| Feature Name | Category | Description | Complexity (L/M/H) | Criticality (Must/Should/Could) | Current Usage (H/M/L) | Rails LOC | Dependencies | Notes | Owner | Status |
|--------------|----------|-------------|-------------------|--------------------------------|---------------------|-----------|--------------|-------|-------|--------|
| User Registration | Auth | User signup with email verification | M | Must | H | ~500 | Devise, ActionMailer | Email verification required | | Not Started |
| OAuth Login | Auth | Login via Google/Facebook | M | Must | M | ~300 | OmniAuth | Multiple providers | | Not Started |
| Create Discussion | Core | Start new discussion thread | H | Must | H | ~800 | ActionCable, ActiveStorage | Real-time updates | | Not Started |
| Voting/Polls | Core | Create and vote on proposals | H | Must | H | ~1200 | Complex state machine | Multiple poll types | | Not Started |
| Email Notifications | Notifications | Send email notifications | M | Must | H | ~600 | ActionMailer, Sidekiq | Template system | | Not Started |
| File Attachments | Collaboration | Attach files to discussions | M | Should | M | ~400 | ActiveStorage, S3 | Image processing | | Not Started |
| Search | Secondary | Full-text search | M | Should | M | ~300 | PostgreSQL FTS | Consider Elasticsearch | | Not Started |
| Admin Dashboard | Admin | Admin management interface | L | Could | L | ~200 | ActiveAdmin | Low priority | | Not Started |

**Instructions:**
1. Copy this to a spreadsheet
2. Add one row per feature
3. Fill in all columns based on code analysis
4. Sort by Criticality, then Complexity
5. Use for prioritization discussions

---

## Template 2: Database Schema Analysis

### Table Inventory

| Table Name | Purpose | Row Count (Est.) | Key Columns | Indexes | Foreign Keys | Special Features | Migration Complexity |
|------------|---------|------------------|-------------|---------|--------------|------------------|---------------------|
| users | Store user accounts | 10,000 | id, email, encrypted_password | email_idx | - | bcrypt hash | Low |
| groups | Organizations/teams | 2,000 | id, name, parent_id | parent_id_idx | parent_id → groups | Self-referential | Medium |
| discussions | Discussion threads | 50,000 | id, group_id, author_id | group_id_idx | group_id → groups | - | Low |
| polls | Voting polls | 20,000 | id, discussion_id, poll_type | discussion_id_idx | discussion_id → discussions | JSONB data | Medium |
| votes | Individual votes | 100,000 | id, poll_id, user_id | poll_user_idx | poll_id → polls | Composite unique | Low |

### Relationships Map

```
users (1) ──< (many) discussions
users (1) ──< (many) votes
users (1) ──< (many) memberships
groups (1) ──< (many) discussions
groups (1) ──< (many) memberships
groups (1) ──< (many) groups (self-join via parent_id)
discussions (1) ──< (many) comments
discussions (1) ──< (many) polls
polls (1) ──< (many) votes
polls (1) ──< (many) poll_options
```

### PostgreSQL-Specific Features Used

- [ ] JSONB columns (list tables/columns)
- [ ] Full-text search (tsvector)
- [ ] Triggers
- [ ] Stored procedures
- [ ] Views
- [ ] Materialized views
- [ ] Custom types/enums
- [ ] Arrays
- [ ] Partitioning
- [ ] Extensions (pg_trgm, uuid-ossp, etc.)

**Action Items:**
- Export schema: `pg_dump --schema-only > schema.sql`
- Generate ER diagram: Use rails-erd or SchemaSpy
- Document all constraints and validations
- Identify potential N+1 query issues

---

## Template 3: API Endpoint Inventory

### REST Endpoints

| Method | Path | Controller#Action | Auth Required | Description | Request Body | Response | Used By | Notes |
|--------|------|-------------------|---------------|-------------|--------------|----------|---------|-------|
| POST | /api/v1/users | users#create | No | Create new user | `{email, password}` | User object | Frontend, Mobile | Rate limited |
| GET | /api/v1/users/:id | users#show | Yes | Get user details | - | User object | Frontend | Include groups |
| POST | /api/v1/discussions | discussions#create | Yes | Create discussion | `{title, body, group_id}` | Discussion object | Frontend | Broadcasts via WS |
| GET | /api/v1/discussions/:id | discussions#show | Yes | Get discussion | - | Discussion + comments | Frontend | Paginated |
| POST | /api/v1/polls | polls#create | Yes | Create poll | `{type, options}` | Poll object | Frontend | Complex validation |

**Generate with:** `rails routes > routes.txt`

### WebSocket Events

| Event Name | Direction | Payload | Triggered By | Purpose |
|------------|-----------|---------|--------------|---------|
| discussion.new | Server → Client | Discussion object | POST /discussions | Real-time discussion creation |
| comment.added | Server → Client | Comment object | POST /comments | Real-time comment updates |
| poll.closed | Server → Client | Poll ID | Poll deadline reached | Notify poll closed |
| user.online | Bidirectional | User ID | WebSocket connect | Presence system |

---

## Template 4: Gem → Go Package Mapping

| Rails Gem | Purpose | Go Equivalent Options | Recommendation | Notes |
|-----------|---------|----------------------|----------------|-------|
| devise | Authentication | JWT + bcrypt, golang-jwt | JWT + bcrypt | More flexibility |
| pundit | Authorization | casbin, go-guardian | casbin | Policy-based |
| active_model_serializers | JSON serialization | encoding/json, jsonapi | encoding/json | Stdlib sufficient |
| sidekiq | Background jobs | asynq, machinery | asynq | Redis-backed |
| actioncable | WebSockets | gorilla/websocket, nhooyr.io/websocket | nhooyr.io/websocket | Better API |
| paperclip/carrierwave | File uploads | minio-go, aws-sdk-go | minio-go | S3-compatible |
| kaminari/will_paginate | Pagination | Custom middleware | Custom | Simple to implement |
| rspec | Testing | testify, ginkgo | testify | Popular choice |
| faker | Test data | gofakeit, faker | gofakeit | Good coverage |
| bullet | N+1 detection | Not needed | - | Fix at design time |
| ransack | Search/filtering | Custom | Custom | Build specific to needs |
| ancestry | Hierarchical data | Custom queries | Custom | Use recursive CTEs |

**Action:** For each gem in Gemfile, research and document Go equivalent

---

## Template 5: Risk Register

| Risk ID | Risk Description | Category | Probability (1-5) | Impact (1-5) | Score | Mitigation Strategy | Owner | Status |
|---------|------------------|----------|------------------|--------------|-------|---------------------|-------|--------|
| R001 | Data loss during migration | Data | 2 | 5 | 10 | Multiple backups, dry runs, validation scripts | DBA | Open |
| R002 | Performance regression | Technical | 3 | 4 | 12 | Early benchmarking, load testing, profiling | Backend | Open |
| R003 | Key team member leaves | People | 2 | 4 | 8 | Documentation, pair programming, knowledge sharing | Manager | Open |
| R004 | Underestimated complexity | Planning | 4 | 3 | 12 | 50% buffer, regular re-estimation, MVP approach | PM | Open |
| R005 | Go learning curve | Skills | 3 | 3 | 9 | Training, code reviews, mentoring | Tech Lead | Open |
| R006 | API breaking changes | Integration | 3 | 4 | 12 | Versioning, deprecation notices, backward compatibility | API Owner | Open |
| R007 | Security vulnerabilities | Security | 2 | 5 | 10 | Security audit, penetration testing, static analysis | Security | Open |
| R008 | Budget overrun | Financial | 3 | 3 | 9 | Weekly budget reviews, scope management | PM | Open |
| R009 | Timeline delays | Schedule | 4 | 3 | 12 | Agile approach, frequent releases, scope flexibility | PM | Open |
| R010 | User resistance to changes | Adoption | 2 | 3 | 6 | Beta testing, gradual rollout, user communication | Product | Open |

**Risk Scoring:** Probability × Impact = Risk Score (prioritize 9+)

---

## Template 6: Architecture Decision Record (ADR)

```markdown
# ADR-001: Use Gin Web Framework for HTTP Routing

## Status
Accepted

## Context
We need to choose a web framework for the Go rewrite of Loomio. The framework must:
- Handle HTTP routing efficiently
- Support middleware for auth, logging, etc.
- Have good documentation and community support
- Provide adequate performance for our scale (10K concurrent users)
- Be actively maintained

We evaluated: Gin, Echo, Fiber, Chi, and stdlib net/http.

## Decision
We will use Gin (github.com/gin-gonic/gin) as our web framework.

## Rationale
- **Performance:** Gin is built on httprouter, one of the fastest routers
- **Middleware:** Rich middleware ecosystem
- **Documentation:** Excellent docs and examples
- **Community:** 75K+ GitHub stars, active maintenance
- **Learning Curve:** Similar to Express.js, familiar to team
- **Validation:** Built-in validation using struct tags
- **Testing:** Good testing support

Benchmarks (req/sec):
- Gin: 45,000
- Echo: 43,000
- Chi: 38,000
- Stdlib: 35,000

## Consequences

### Positive
- Fast development with familiar patterns
- Good performance out of the box
- Large ecosystem of middleware
- Easy to find solutions to common problems

### Negative
- Slightly less idiomatic than stdlib approach
- Adds external dependency
- Some magic with context binding
- Opinionated routing style

### Neutral
- Need to establish patterns for error handling
- Will need custom middleware for some Rails patterns

## Alternatives Considered

### Echo
Similar to Gin, slightly different API. Team preferred Gin syntax.

### Chi
More idiomatic Go, but smaller ecosystem. Good option but less familiar.

### Stdlib net/http
Most idiomatic, no dependencies, but more boilerplate. Too low-level for our timeline.

### Fiber
Fastest benchmarks but Express.js-like API might confuse. Less mature.

## Implementation Notes
- Use Gin's grouping for API versioning
- Implement standard middleware: auth, logging, error handling, CORS
- Use gin.Context for request/response handling
- Follow Gin best practices for testing

## References
- Gin Documentation: https://gin-gonic.com/docs/
- Benchmark source: [link to our benchmarks]
- Team discussion: [link to discussion thread]

## Review Date
2024-06-01 (6 months after implementation)

---
**Author:** [Name]  
**Date:** 2024-01-15  
**Reviewed By:** [Names]  
**Approved By:** [Technical Lead]
```

---

## Template 7: Sprint Planning Template

### Sprint N: [Name/Theme]

**Duration:** 2 weeks  
**Start Date:** YYYY-MM-DD  
**End Date:** YYYY-MM-DD  
**Team Capacity:** X story points / Y hours

### Sprint Goal
One sentence describing what we aim to achieve this sprint.

### User Stories

#### Story 1: [Title]
**As a** [user role]  
**I want** [functionality]  
**So that** [benefit]

**Acceptance Criteria:**
- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

**Technical Tasks:**
- [ ] Task 1 (Estimated: 4h)
- [ ] Task 2 (Estimated: 6h)
- [ ] Task 3 (Estimated: 2h)

**Testing Requirements:**
- Unit tests with >80% coverage
- Integration test for happy path
- Error case handling

**Story Points:** 5  
**Priority:** High  
**Assignee:** [Name]  
**Dependencies:** None

---

### Technical Debt Items
- [ ] Refactor X module (2h)
- [ ] Update documentation for Y (1h)

### Sprint Risks
- Risk 1: Description and mitigation
- Risk 2: Description and mitigation

### Definition of Done
- [ ] Code written and reviewed
- [ ] Tests written and passing
- [ ] Documentation updated
- [ ] Deployed to staging
- [ ] Demo prepared
- [ ] Acceptance criteria met

---

## Template 8: Weekly Progress Report

```markdown
# Weekly Progress Report - Week of [Date]

## Summary
One paragraph summary of the week's progress.

## Completed This Week
- ✅ Item 1 - [Link to PR]
- ✅ Item 2 - [Link to PR]
- ✅ Item 3 - [Link to PR]

## In Progress
- 🔄 Item 1 - 60% complete - [Link]
- 🔄 Item 2 - 30% complete - [Link]

## Blocked
- 🚫 Item 1 - Blocked by: [reason] - Owner: [name]

## Metrics
- **Story points completed:** X / Y planned
- **PRs merged:** N
- **Test coverage:** X%
- **Bugs found:** N (P0: n, P1: n, P2: n)
- **Bugs fixed:** N

## Risks & Issues
1. **[Risk/Issue Title]** - [Red/Yellow/Green]
   - Impact: [description]
   - Mitigation: [action taken]
   - Owner: [name]

## Learnings & Insights
- Learning 1
- Learning 2

## Next Week's Focus
- Priority 1
- Priority 2
- Priority 3

## Help Needed
- Request 1
- Request 2

## Links
- Sprint board: [link]
- Meeting notes: [link]
```

---

## Template 9: Technical Debt Log

| ID | Description | Area | Severity | Estimated Effort | Impact if Not Fixed | Created | Status |
|----|-------------|------|----------|-----------------|-------------------|---------|--------|
| TD001 | No integration tests for auth | Testing | High | 2 days | Hard to refactor safely | 2024-01-15 | Open |
| TD002 | Hardcoded config values | Config | Medium | 4 hours | Difficult deployments | 2024-01-16 | Open |
| TD003 | Missing database indexes | Performance | High | 1 day | Slow queries at scale | 2024-01-17 | In Progress |
| TD004 | Inconsistent error handling | Code Quality | Medium | 3 days | Poor debugging experience | 2024-01-18 | Open |
| TD005 | Copy-pasted validation logic | Code Quality | Low | 1 day | Maintenance burden | 2024-01-19 | Backlog |

**Review Frequency:** Every sprint planning session  
**Allocation:** 20% of sprint capacity for tech debt

---

## Template 10: Code Review Checklist

### Functionality
- [ ] Code does what it's supposed to do
- [ ] Edge cases are handled
- [ ] Error cases are handled gracefully
- [ ] No obvious bugs

### Code Quality
- [ ] Code is readable and well-organized
- [ ] Functions are small and focused
- [ ] No code duplication (DRY principle)
- [ ] Follows Go conventions and idioms
- [ ] Comments explain "why" not "what"

### Testing
- [ ] Unit tests are present and pass
- [ ] Test coverage is adequate (>80%)
- [ ] Tests are readable and maintainable
- [ ] Integration tests for key flows
- [ ] Edge cases are tested

### Security
- [ ] No hardcoded secrets or credentials
- [ ] Input validation is present
- [ ] SQL injection prevention (parameterized queries)
- [ ] XSS prevention in templates
- [ ] Authentication/authorization checks

### Performance
- [ ] No obvious performance issues
- [ ] Database queries are optimized
- [ ] No N+1 query problems
- [ ] Appropriate use of caching
- [ ] No memory leaks

### Documentation
- [ ] Public functions have godoc comments
- [ ] Complex logic is explained
- [ ] API changes are documented
- [ ] README is updated if needed

### Dependencies
- [ ] No unnecessary dependencies added
- [ ] Dependencies are up to date
- [ ] License compatibility checked

### Migration
- [ ] Database migrations are reversible
- [ ] Migrations tested on copy of production data
- [ ] Backward compatibility maintained

---

## Usage Instructions

1. **Copy templates** to your project management tool (Notion, Confluence, etc.)
2. **Customize** based on your specific needs
3. **Fill out** as you complete discovery work
4. **Review** regularly with the team
5. **Update** as you learn more
6. **Reference** when making decisions

These templates are starting points—adapt them to your team's workflow and culture.

---

## Quick Start Checklist

Week 1:
- [ ] Set up discovery folder/workspace
- [ ] Copy all templates
- [ ] Assign owners for each template
- [ ] Schedule daily syncs
- [ ] Start feature inventory
- [ ] Begin database analysis

Week 2:
- [ ] Complete gem mapping
- [ ] Start API inventory
- [ ] Begin risk register
- [ ] Document current architecture

Week 3:
- [ ] Build POCs for stack decisions
- [ ] Write first ADRs
- [ ] Update risk register

Week 4:
- [ ] Complete all templates
- [ ] Review with team
- [ ] Prepare final plan document

**Remember:** Templates are tools, not rules. Use what helps, skip what doesn't. The goal is clarity and confidence, not documentation for its own sake.