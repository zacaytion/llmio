# Meta-Plan: Loomio Rewrite from Ruby on Rails to Go

## Executive Summary

This document outlines a strategic meta-plan for rewriting Loomio, a collaborative decision-making tool currently built with Ruby on Rails, into Go. This is not the implementation plan itself, but rather a framework for creating a comprehensive rewrite plan.

**Current State:** Ruby on Rails monolith (~55.7% Ruby, ~22.2% Vue.js frontend)  
**Target State:** Go-based backend with retained/modernized Vue.js frontend  
**Estimated Timeline:** 12-18 months (to be refined in detailed plan)

---

## Phase 1: Discovery & Analysis (Meta-Planning Phase)

### 1.1 Repository Analysis Tasks

**Objective:** Understand the full scope of what needs to be rewritten

- [ ] Clone and audit the Loomio repository structure
- [ ] Document the Rails application architecture
  - Controllers, models, views, services
  - Database schema and migrations
  - Background jobs and workers
  - API endpoints (REST/GraphQL)
  - WebSocket/real-time features (via loomio_channel_server)
- [ ] Identify external dependencies and integrations
  - Third-party gems and their Go equivalents
  - Email systems (ActionMailbox/Haraka)
  - File storage solutions
  - Authentication/OAuth providers
- [ ] Map the Vue.js frontend structure
  - API contracts and data flows
  - State management patterns
  - Component dependencies
- [ ] Document the test suite
  - Unit tests (RSpec)
  - Integration tests
  - E2E tests
  - Test coverage metrics

### 1.2 Database & Data Model Analysis

**Objective:** Plan data layer migration strategy

- [ ] Export complete database schema
- [ ] Identify database-specific features (PostgreSQL functions, triggers, etc.)
- [ ] Document data migration patterns used
- [ ] Analyze query patterns and ORM usage
- [ ] Identify N+1 query issues (opportunities for improvement)
- [ ] Document data validation rules and constraints

### 1.3 Feature Inventory & Prioritization

**Objective:** Create a comprehensive feature list with complexity ratings

- [ ] Extract feature list from codebase
- [ ] Cross-reference with documentation and user guides
- [ ] Categorize features:
  - Core features (MVP)
  - Secondary features
  - Nice-to-have features
  - Deprecated/rarely-used features
- [ ] Rate each feature by:
  - Complexity (Low/Medium/High)
  - Business criticality (Must-have/Should-have/Could-have)
  - Usage frequency (High/Medium/Low)
  - Technical debt in current implementation

### 1.4 Technical Stack Selection

**Objective:** Choose the Go ecosystem components

Research and evaluate options for:

- [ ] **Web Framework**
  - Options: Gin, Echo, Fiber, Chi, net/http (stdlib)
  - Criteria: Performance, middleware ecosystem, learning curve
- [ ] **Database ORM/Query Builder**
  - Options: GORM, sqlx, sqlc, ent
  - Criteria: Type safety, migrations, relation handling
- [ ] **Authentication**
  - Options: JWT libraries, OAuth2 libraries
  - Criteria: Security, standards compliance
- [ ] **Background Jobs**
  - Options: Asynq, Machinery, River
  - Criteria: Reliability, Redis/PostgreSQL backend
- [ ] **WebSocket/Real-time**
  - Options: gorilla/websocket, nhooyr.io/websocket, Centrifugo
  - Criteria: Scalability, reconnection handling
- [ ] **Email Handling**
  - Options: gomail, Mailgun SDK, native SMTP
  - Criteria: Template support, delivery reliability
- [ ] **File Storage**
  - Options: Minio SDK, AWS SDK, local filesystem
  - Criteria: S3 compatibility, streaming support
- [ ] **Testing Framework**
  - Options: testify, ginkgo/gomega
  - Criteria: Assertion library, mocking support
- [ ] **API Documentation**
  - Options: Swagger/OpenAPI, protobuf
  - Criteria: Auto-generation, frontend compatibility

### 1.5 Architecture Decision Records (ADRs)

**Objective:** Document key architectural decisions

Create ADRs for:

- [ ] Monolith vs. Microservices decision
- [ ] Database migration strategy (big bang vs. gradual)
- [ ] API versioning approach
- [ ] Error handling patterns
- [ ] Logging and observability strategy
- [ ] Configuration management
- [ ] Deployment strategy
- [ ] Frontend-backend contract (REST vs. GraphQL)

---

## Phase 2: Planning Framework

### 2.1 Migration Strategy Selection

**Options to evaluate:**

1. **Big Bang Rewrite**
   - Pros: Clean slate, no dual maintenance
   - Cons: High risk, long time to production
   
2. **Strangler Fig Pattern**
   - Pros: Incremental migration, reduced risk
   - Cons: Complex infrastructure, dual maintenance period
   
3. **Hybrid Approach**
   - Pros: Balance of speed and risk
   - Cons: Requires careful boundary definition

**Deliverable:** Decision matrix with recommendation

### 2.2 Team & Resource Planning Template

- [ ] Required skill sets (Go, Rails, Vue.js, DevOps)
- [ ] Team size and composition
- [ ] Training needs
- [ ] External consultants/contractors
- [ ] Time allocation (full-time vs. part-time)

### 2.3 Risk Assessment Framework

**Risk Categories:**

| Risk Type | Questions to Answer |
|-----------|-------------------|
| Technical | What are the hardest problems? Which Rails magic is hardest to replicate? |
| Business | How do we maintain velocity during rewrite? |
| Data | How do we ensure zero data loss? |
| User Experience | How do we avoid disrupting users? |
| Performance | Will Go provide measurable improvements? |
| Team | Do we have Go expertise? What's the learning curve? |

### 2.4 Testing Strategy Framework

- [ ] Test coverage requirements (minimum %)
- [ ] Test migration approach (rewrite vs. translate)
- [ ] Integration testing strategy
- [ ] Performance testing benchmarks
- [ ] Load testing scenarios
- [ ] Data migration validation tests

### 2.5 Deployment & Rollout Planning

- [ ] Deployment infrastructure (containers, K8s, etc.)
- [ ] Blue-green deployment strategy
- [ ] Feature flags for gradual rollout
- [ ] Rollback procedures
- [ ] Monitoring and alerting setup
- [ ] Performance baseline establishment

---

## Phase 3: Detailed Plan Creation Structure

### 3.1 Work Breakdown Structure (WBS)

**Template for breaking down work:**

```
1. Foundation (Weeks 1-4)
   1.1 Project setup
   1.2 CI/CD pipeline
   1.3 Database layer
   1.4 Authentication system
   
2. Core Features (Weeks 5-20)
   2.1 User management
   2.2 Group/organization management
   2.3 Discussion threads
   2.4 Polling/voting
   2.5 Decision-making workflows
   
3. Secondary Features (Weeks 21-32)
   3.1 Notifications
   3.2 Email integration
   3.3 File attachments
   3.4 Search functionality
   
4. Integration & Polish (Weeks 33-40)
   4.1 Frontend integration
   4.2 Data migration
   4.3 Performance optimization
   4.4 Security hardening
   
5. Testing & Launch (Weeks 41-48)
   5.1 Full test suite
   5.2 Load testing
   5.3 Beta deployment
   5.4 Production migration
```

### 3.2 Sprint Planning Template

**For each 2-week sprint:**

- Sprint goals
- User stories with acceptance criteria
- Technical tasks
- Testing requirements
- Documentation needs
- Definition of done

### 3.3 Milestone Definition

**Key milestones to define:**

| Milestone | Criteria | Target Date |
|-----------|----------|-------------|
| M1: Foundation Complete | Auth + DB + Basic API working | TBD |
| M2: Feature Parity 50% | Core features implemented | TBD |
| M3: Feature Parity 100% | All features implemented | TBD |
| M4: Beta Ready | Tested, documented, deployable | TBD |
| M5: Production Launch | Live for all users | TBD |

---

## Phase 4: Quality Assurance Framework

### 4.1 Code Quality Standards

- [ ] Define Go coding standards (golangci-lint config)
- [ ] Code review process and checklist
- [ ] Documentation requirements
- [ ] Test coverage thresholds
- [ ] Performance benchmarks

### 4.2 Compatibility Matrix

**Ensure compatibility with:**

- [ ] Existing user data
- [ ] Current API clients (if any)
- [ ] Browser support matrix
- [ ] Mobile app compatibility (if applicable)
- [ ] Integration webhooks and APIs

### 4.3 Performance Benchmarks

**Metrics to track:**

| Metric | Rails Baseline | Go Target | Measurement Method |
|--------|---------------|-----------|-------------------|
| API response time (p95) | TBD | < X ms | Load testing |
| Concurrent users | TBD | X users | Stress testing |
| Memory usage | TBD | < X MB | Profiling |
| Database query time | TBD | < X ms | Query logging |
| Build time | TBD | < X min | CI metrics |

---

## Phase 5: Documentation Strategy

### 5.1 Technical Documentation Needs

- [ ] Architecture diagrams (before/after)
- [ ] API documentation (OpenAPI/Swagger)
- [ ] Database schema documentation
- [ ] Deployment runbooks
- [ ] Development setup guide
- [ ] Testing guide
- [ ] Troubleshooting guide

### 5.2 Knowledge Transfer Plan

- [ ] Rails → Go pattern translation guide
- [ ] Onboarding documentation for new team members
- [ ] Video walkthroughs of key systems
- [ ] Decision rationale documentation
- [ ] Post-mortem template

### 5.3 User-Facing Documentation

- [ ] Migration guide for self-hosted users
- [ ] API changelog (breaking changes)
- [ ] Feature parity documentation
- [ ] Upgrade instructions

---

## Phase 6: Execution Monitoring

### 6.1 Progress Tracking Metrics

**KPIs to monitor:**

- [ ] Features completed vs. planned
- [ ] Test coverage %
- [ ] Bug count (open/closed)
- [ ] Performance vs. benchmarks
- [ ] Technical debt accumulation
- [ ] Team velocity
- [ ] Budget vs. actual spend

### 6.2 Risk Monitoring

**Weekly risk assessment:**

- Identify new risks
- Update risk probability/impact
- Review mitigation effectiveness
- Escalation triggers

### 6.3 Stakeholder Communication Plan

**Regular updates to:**

- Development team (daily standups)
- Technical leadership (weekly)
- Product/business stakeholders (bi-weekly)
- Community/users (monthly/milestone-based)

---

## Phase 7: Success Criteria

### 7.1 Technical Success Metrics

- [ ] All automated tests pass
- [ ] Code coverage > X%
- [ ] No critical security vulnerabilities
- [ ] Performance meets/exceeds benchmarks
- [ ] Zero data loss during migration
- [ ] 99.9% uptime in first month post-launch

### 7.2 Business Success Metrics

- [ ] Feature parity achieved
- [ ] User adoption rate > X%
- [ ] User satisfaction maintained/improved
- [ ] Support ticket volume stable/decreased
- [ ] Hosting costs reduced by X%
- [ ] Development velocity increased by X%

### 7.3 Exit Criteria (When is Rails version deprecated?)

- [ ] Go version stable for X months
- [ ] All users migrated
- [ ] All data verified
- [ ] Documentation complete
- [ ] Team trained on Go codebase
- [ ] Monitoring and alerting proven effective

---

## Appendices

### A. Research Questions to Answer

Before creating the detailed plan, answer:

1. What is the current Rails version and upgrade path?
2. What is the total user base and usage patterns?
3. What are the most computationally expensive operations?
4. What are the current pain points in the Rails implementation?
5. What features are most requested by users?
6. What is the current hosting infrastructure?
7. What is the current team's Go experience level?
8. What is the business timeline/pressure for this rewrite?
9. Are there compliance requirements (GDPR, AGPL, etc.)?
10. What is the budget for this project?

### B. Key Loomio Features to Analyze

Based on preliminary research, key features include:

- **Core Decision-Making:**
  - Proposals and voting
  - Multiple poll types
  - Time-bound decisions
  - Consensus tracking
  
- **Collaboration:**
  - Discussion threads
  - Comments and replies
  - Mentions and notifications
  - Real-time updates
  
- **Organization:**
  - Groups and subgroups
  - Permissions and roles
  - Member management
  
- **Integrations:**
  - Email (inbound/outbound)
  - OAuth providers
  - Webhooks
  - File attachments

### C. Go Ecosystem Resources

- Go project structure best practices
- Go testing patterns
- Go security best practices
- Go performance optimization guides
- Go deployment patterns

### D. Rails → Go Translation Guide Template

| Rails Concept | Go Equivalent | Notes |
|---------------|---------------|-------|
| ActiveRecord | GORM/sqlc | Consider migration strategy |
| ActionCable | gorilla/websocket | WebSocket implementation |
| ActiveJob | Asynq/Machinery | Background job processing |
| ActionMailer | gomail | Email templates |
| Devise | JWT + bcrypt | Authentication |
| Pundit | casbin | Authorization |
| RSpec | testify | Testing framework |
| Rake tasks | cobra | CLI commands |

---

## Next Steps

1. **Review this meta-plan** with all stakeholders
2. **Assign owners** for each Phase 1 discovery task
3. **Set timeline** for completing Phase 1 (suggested: 2-4 weeks)
4. **Schedule regular sync meetings** to track discovery progress
5. **Create detailed plan document** based on discovery findings
6. **Get sign-off** on detailed plan before starting implementation

---

## Document Control

- **Version:** 1.0
- **Created:** [Date]
- **Author:** [Name]
- **Last Updated:** [Date]
- **Status:** Draft
- **Next Review:** After Phase 1 completion

---

**Note:** This is a living document. As we progress through discovery, we'll update this meta-plan to reflect new learnings and adjust our approach accordingly.