# Loomio Go Deployment & Rollout Plan

**Date:** 2026-01-29
**Status:** Proposed

## Current Deployment Architecture

From Discovery Report, Loomio uses Docker Compose on single host:
- nginx-proxy (SSL termination)
- app (Rails/Puma)
- worker (Sidekiq)
- channels (Socket.io)
- hocuspocus (Y.js)
- db (PostgreSQL 17)
- redis

## Target Architecture (Go)

### Phase 1: Parallel Systems

```
┌─────────────────────────────────────────────────┐
│                  nginx-proxy                     │
│         (routes based on path/feature flag)      │
└─────────────────┬───────────────────────────────┘
                  │
      ┌───────────┼───────────┐
      │           │           │
      ▼           ▼           ▼
┌─────────┐ ┌─────────┐ ┌─────────────┐
│  Rails  │ │   Go    │ │  channels   │
│  (old)  │ │  (new)  │ │  (Go/Node)  │
└────┬────┘ └────┬────┘ └─────────────┘
     │           │
     └─────┬─────┘
           ▼
     ┌──────────┐
     │ PostgreSQL│
     │  (shared) │
     └──────────┘
```

### Phase 2: Go Primary

```
┌─────────────────────────────────────────────────┐
│                  nginx-proxy                     │
└─────────────────┬───────────────────────────────┘
                  │
      ┌───────────┼───────────┐
      │           │           │
      ▼           ▼           ▼
┌─────────┐ ┌─────────────┐ ┌────────────┐
│   Go    │ │  channels   │ │ hocuspocus │
│   API   │ │    (Go)     │ │  (Node.js) │
└────┬────┘ └──────┬──────┘ └────────────┘
     │             │
     └──────┬──────┘
            ▼
     ┌──────────┐
     │ PostgreSQL│
     └──────────┘
```

## Rollout Strategy

### Stage 1: Shadow Mode (2 weeks)
- Deploy Go alongside Rails
- Mirror traffic to Go (write to logs only)
- Compare responses for discrepancies
- No user impact

### Stage 2: Canary (2 weeks)
- Route 1% of read-only traffic to Go
- Monitor error rates and latency
- Automatic rollback on anomalies

### Stage 3: Gradual Rollout (4 weeks)
- Increase Go traffic: 1% → 10% → 50% → 100%
- Each increase after 1 week of stability
- Maintain Rails as hot standby

### Stage 4: Rails Sunset (2 weeks)
- Go handles 100% of traffic
- Rails kept running but not receiving traffic
- Final data verification
- Rails containers removed

## Feature Flag Strategy

Use environment-based routing in nginx:

```nginx
# Example: Route /api/v1/groups to Go
location /api/v1/groups {
    if ($go_enabled = "true") {
        proxy_pass http://go-api:8080;
    }
    proxy_pass http://rails-app:3000;
}
```

Feature flags in database for fine-grained control:
- Per-endpoint toggles
- Per-user toggles (for beta testers)
- Percentage-based rollout

## Monitoring Requirements

### Metrics to Track
- Request latency (p50, p95, p99)
- Error rate (5xx, 4xx)
- Database query time
- WebSocket connection count
- Memory and CPU usage

### Alerting Thresholds
- Error rate > 1%: Page on-call
- p95 latency > 500ms: Warning
- p99 latency > 2s: Page on-call
- Memory > 80%: Warning

### Dashboards
- Side-by-side Rails vs Go metrics
- User-facing error tracking (Sentry)
- Database performance

## Rollback Procedures

### Automatic Rollback Triggers
- Error rate > 5% for 5 minutes
- p99 latency > 5s for 5 minutes
- Health check failures

### Manual Rollback Steps
1. Set feature flag to route all traffic to Rails
2. Verify Rails health
3. Investigate Go issues
4. Do not remove Go containers (for debugging)

### Rollback Time Target
- Automatic: < 1 minute
- Manual: < 5 minutes

## Infrastructure Changes

### New Containers
- `loomio/loomio-go:latest` - Go API server
- `loomio/channels-go:latest` - Go channel server (records + bots)

### Retained Containers
- `loomio/loomio_channel_server` - Hocuspocus only

### Removed Containers (after full migration)
- `loomio/loomio` - Rails app
- Sidekiq worker (replaced by River in Go)

## Database Migration

### Schema Changes
- Minimal schema changes needed
- Go uses same PostgreSQL database
- sqlc generates types from existing schema

### Data Migration
- No data migration needed (same schema)
- Application-level compatibility only

## Self-Hosted User Communication

1. **Announcement:** 3 months before release
2. **Beta period:** Opt-in for testing
3. **Release notes:** Detailed upgrade guide
4. **Docker image tags:** Maintain Rails image for 6 months
5. **Support:** Dedicated channel for migration issues
