# Confirmed Architecture Summary

This document serves as an entry point to the `synthesis/` directory, which contains 4 documents consolidating architecture findings **confirmed by both research teams**. These represent the "ground truth" we've established through cross-validation.

## How to Use This Document

1. **Reference the Key Numbers** below for quick facts
2. **Read the relevant synthesis doc** in `synthesis/` for detailed confirmation with source alignment tables
3. **Use for implementation planning** - these findings can be trusted for Go rewrite decisions
4. Each document includes Go implementation notes and code examples

---

## Document Index

| File | Topics Covered | Key Confirmations |
|------|---------------|-------------------|
| `core_models.md` | Polls, Events, Stances, Soft Delete | 9 poll types, 42 event kinds, voting mechanics, volume levels |
| `authorization.md` | Permissions, Roles, Access Control | CanCanCan framework, 25 ability modules, 4 membership roles |
| `realtime_architecture.md` | WebSocket, Redis, Collaborative Editing | Redis channel patterns, Socket.io rooms, Hocuspocus auth |
| `infrastructure.md` | Deployment, Background Jobs, Storage | 10 Docker services, 38 workers, 5 storage backends |

---

## Key Numbers (Confirmed by Both Teams)

### Domain Model
- **9 poll types**: proposal, poll, count, score, ranked_choice, meeting, dot_vote, check, question
- **42 event kinds**: STI pattern with kind discriminator
- **14 webhook-eligible events**: (enumeration pending - see `follow_up/webhook_eligible_events.md`)
- **4 volume levels**: mute(0), quiet(1), normal(2), loud(3)
- **17 counter caches** on Group model

### Authorization
- **25 CanCanCan ability modules**: prepend-based composition
- **4 membership roles**: guest, member, delegate, admin
- **9 core permission flags**: `members_can_*` booleans on Group

### Infrastructure
- **10 Docker services**: nginx-proxy, acme, app, worker, db, redis, haraka, channels, hocuspocus, pgbackups
- **38 Sidekiq workers**: with queue priorities (critical=10, high=6, default=3, low=1)
- **5 storage backends**: local disk, S3, DigitalOcean Spaces, GCS, S3-compatible
- **60+ environment variables**: documented in deployment

### Real-time
- **3 Redis pub/sub channels**: `/records`, `/system_notice`, `chatbot/*`
- **3 Socket.io room patterns**: `notice`, `user-{id}`, `group-{id}`
- **9 Hocuspocus record types**: comment, discussion, poll, stance, outcome, pollTemplate, discussionTemplate, group, user

---

## Quick Reference by Topic

### If You Need To Understand...

**Poll Types and Voting Mechanics**
→ Read `synthesis/core_models.md`
- All 9 poll types with voting semantics
- Stance model with option_scores JSONB
- StanceChoice scoring per poll type

**Event System and Notifications**
→ Read `synthesis/core_models.md`
- 42 event kinds listed with categories
- STI pattern explanation
- Webhook-eligible subset

**Permission Model**
→ Read `synthesis/authorization.md`
- CanCanCan architecture
- Ability module composition
- Group permission flags
- Membership roles and access levels

**WebSocket and Real-time Updates**
→ Read `synthesis/realtime_architecture.md`
- Redis pub/sub channel patterns
- Socket.io room structure
- Connection authentication flow

**Collaborative Editing (Hocuspocus)**
→ Read `synthesis/realtime_architecture.md`
- Auth endpoint and token format
- Document naming convention
- Ephemeral storage design (intentional)

**Deployment and Infrastructure**
→ Read `synthesis/infrastructure.md`
- Docker Compose service map
- Storage backend configuration
- Environment variables

**Background Jobs**
→ Read `synthesis/infrastructure.md`
- 38 worker inventory
- Queue priorities
- Sidekiq configuration

---

## Synthesis Document Format

Each synthesis document follows this structure:

```
# [Topic] - Confirmed Architecture

## Summary
What both sources agree on

## Key Details
Specific technical details with tables

## Source Alignment
| Aspect | Discovery | Research | Status |
|--------|-----------|----------|--------|
| ... | ... | ... | ✅ Confirmed |

## Implementation Notes
Go code examples and migration guidance
```

---

## Confidence Levels

These documents represent our **highest confidence findings**:

| Document | Confidence | Notes |
|----------|------------|-------|
| `core_models.md` | Very High | Both teams independently enumerated same poll types, event kinds |
| `authorization.md` | High | Agreement on framework and patterns; minor flag count question in follow_up |
| `realtime_architecture.md` | Very High | Research had deeper detail, discovery confirmed key patterns |
| `infrastructure.md` | Very High | Deployment configs are explicit and verifiable |

---

## Relationship to Follow-up Documents

The `synthesis/` docs contain **confirmed** findings. The `follow_up/` docs contain **discrepancies** that need codebase verification.

Example: Both teams agree on "14 webhook-eligible events" (synthesis), but neither enumerated them (follow-up investigation needed).

When a follow-up investigation resolves a discrepancy, the finding should be added to the appropriate synthesis document.
