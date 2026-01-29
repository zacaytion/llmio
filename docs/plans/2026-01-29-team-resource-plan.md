# Loomio Rewrite Team & Resource Plan

**Date:** 2026-01-29
**Status:** Template (to be filled with actual team data)

## Required Skill Sets

| Skill | Required Level | Purpose |
|-------|----------------|---------|
| Go | Intermediate+ | Backend development |
| PostgreSQL | Intermediate | Database work, migrations |
| Ruby/Rails | Basic | Understanding existing code |
| Vue.js | Basic | Frontend debugging |
| WebSocket | Basic | Real-time features |
| DevOps | Intermediate | CI/CD, deployment |

## Team Composition Options

### Option A: Dedicated Rewrite Team (Recommended)

| Role | Count | Responsibility |
|------|-------|----------------|
| Tech Lead | 1 | Architecture, decisions, code review |
| Senior Go Dev | 2 | Core implementation |
| Mid Go Dev | 2 | Feature implementation |
| QA Engineer | 1 | Testing strategy, automation |
| DevOps | 0.5 | CI/CD, deployment |

**Total:** 6.5 FTE
**Duration:** 12-18 months

### Option B: Part-Time Migration

| Role | Allocation | Responsibility |
|------|------------|----------------|
| Existing team | 50% | Rewrite alongside maintenance |

**Total:** Varies by team size
**Duration:** 18-24 months (longer due to context switching)

### Option C: External Augmentation

| Role | Source | Responsibility |
|------|--------|----------------|
| Go consultants | Contract | Initial architecture, training |
| Existing team | Internal | Implementation after training |

**Total:** 2 contractors + internal team
**Duration:** 14-18 months

## Training Plan

### Phase 1: Go Fundamentals (Week 1-2)
- Tour of Go (all team members)
- Effective Go reading
- Go by Example exercises
- Internal code review sessions

### Phase 2: Stack Training (Week 3-4)
- Chi routing patterns
- sqlc + pgx database patterns
- River job processing
- Testing with testify

### Phase 3: Loomio Domain (Week 5-6)
- Rails codebase walkthrough
- Discovery Report review
- Domain model deep-dive
- Permission system review

## Resource Requirements

### Development Environment
- [ ] Go 1.22+ installed
- [ ] PostgreSQL 17 local instance
- [ ] Redis local instance
- [ ] Docker for testcontainers
- [ ] IDE with Go support (GoLand, VS Code)

### Infrastructure
- [ ] CI/CD pipeline (GitHub Actions)
- [ ] Staging environment (mirrors production)
- [ ] Load testing environment
- [ ] Anonymized production data copy

### Budget Considerations

| Category | Estimate | Notes |
|----------|----------|-------|
| Personnel | $X/month | Based on team composition |
| Infrastructure | $X/month | Staging, CI minutes |
| Training | $X one-time | Courses, materials |
| Tools | $X/month | IDE licenses, monitoring |

## Communication Plan

| Meeting | Frequency | Participants | Purpose |
|---------|-----------|--------------|---------|
| Daily standup | Daily | Dev team | Progress, blockers |
| Tech review | Weekly | Tech lead + seniors | Architecture decisions |
| Stakeholder update | Bi-weekly | All + stakeholders | Progress report |
| Retrospective | Per sprint | Dev team | Process improvement |

## Decision Authority

| Decision Type | Authority | Escalation |
|---------------|-----------|------------|
| Code style | Any developer | Tech lead |
| Package selection | Tech lead | Team consensus |
| Architecture | Tech lead + seniors | ADR process |
| Timeline changes | Tech lead | Stakeholders |
| Scope changes | Stakeholders | - |
