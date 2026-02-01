# Infrastructure - Confirmed Architecture

## Summary

Both Discovery and Research documentation **fully agree** on the deployment infrastructure. Loomio runs as a Docker Compose stack with 10 services.

## Key Details

### Docker Services (10 Confirmed)

| Service | Image/Technology | Port | Purpose |
|---------|------------------|------|---------|
| `nginx-proxy` | jwilder/nginx-proxy | 80, 443 | Reverse proxy, SSL termination |
| `nginx-proxy-acme` | nginxproxy/acme-companion | - | Let's Encrypt certificate automation |
| `app` | loomio/loomio | 3000 | Rails application server |
| `worker` | loomio/loomio | - | Sidekiq background jobs |
| `db` | postgres:17 | 5432 | PostgreSQL database |
| `redis` | redis:8.4 | 6379 | Cache, sessions, pub/sub, jobs |
| `haraka` | loomio/haraka | 25 | SMTP server for inbound email |
| `channels` | loomio/loomio_channel_server | 5000 | Socket.io real-time updates |
| `hocuspocus` | loomio/loomio_channel_server | 5000 | Y.js collaborative editing |
| `pgbackups` | prodrigestivill/postgres-backup-local | - | Automated daily database backups |

### Storage Backends (5 Confirmed)

| Backend | Environment Variables | Use Case |
|---------|----------------------|----------|
| Local Disk (test) | `tmp/storage` | Development/testing |
| Local Disk (production) | `storage/` volume | Self-hosted, no cloud |
| Amazon S3 | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_BUCKET`, `AWS_REGION` | Production cloud |
| DigitalOcean Spaces | `DO_ENDPOINT`, `DO_ACCESS_KEY_ID`, `DO_SECRET_ACCESS_KEY`, `DO_BUCKET` | DO hosting |
| Google Cloud Storage | `GCS_CREDENTIALS`, `GCS_PROJECT`, `GCS_BUCKET` | GCP hosting |

### PostgreSQL Extensions

Both sources confirm required extensions:

| Extension | Purpose |
|-----------|---------|
| `citext` | Case-insensitive text (email, handles, tags) |
| `hstore` | Key-value pairs (translations, headers) |
| `pgcrypto` | UUID generation, encryption |
| `pg_stat_statements` | Query performance monitoring |

### Background Jobs (Sidekiq)

Both sources confirm **38 workers** with queue priorities:

| Priority | Value | Queues | Example Workers |
|----------|-------|--------|-----------------|
| Critical | 10 | `critical` | - |
| High | 6 | `high` | `PublishEventWorker` |
| Default | 3 | `default` | `GenericWorker`, `WebhookWorker` |
| Low | 1 | `low`, `mailers` | `DeliverAnnouncementWorker` |

**Sidekiq Configuration**:
- Production: 20 threads
- Development: 5 threads

### Full-Text Search (pg_search)

Both sources confirm pg_search with:

| Aspect | Value |
|--------|-------|
| Table | `pg_search_documents` |
| Column | `ts_content` (tsvector) |
| Searchable models | Discussion, Comment, Poll, Stance, Outcome |

### Environment Variables (60+ Confirmed)

**Core:**
- `CANONICAL_HOST` - Primary domain
- `SECRET_COOKIE_SECRET` - Session encryption
- `DEVISE_SECRET_KEY` - Devise encryption
- `RAILS_MASTER_KEY` - Rails credentials
- `DATABASE_URL` - PostgreSQL connection
- `REDIS_URL` - Redis connection

**Real-time:**
- `CHANNELS_URI` - Socket.io WebSocket URL
- `HOCUSPOCUS_URI` - Y.js WebSocket URL
- `PRIVATE_APP_URL` - Internal Rails URL for Hocuspocus auth

**Email:**
- `SMTP_DOMAIN`, `SMTP_USERNAME`, `SMTP_PASSWORD` - Outbound
- `REPLY_HOSTNAME` - Inbound email domain

**Features:**
- `FEATURES_DISABLE_*` - Feature flag toggles
- `MAX_ATTACHMENT_BYTES` - Upload limit

## Source Alignment

| Aspect | Discovery | Research | Status |
|--------|-----------|----------|--------|
| Docker services count | 10 | 10 | ✅ Confirmed |
| Service names | Identical | Identical | ✅ Confirmed |
| Storage backends | 5 | 5 | ✅ Confirmed |
| PostgreSQL extensions | 4 | 4 | ✅ Confirmed |
| Sidekiq workers | 38 | 38 | ✅ Confirmed |
| Environment variables | 60+ | 60+ | ✅ Confirmed |
| pg_search configuration | Basic | Basic | ✅ Confirmed |

## Implementation Notes

### Go Deployment Considerations

**Replace Rails containers:**
```yaml
# docker-compose.yml changes
services:
  app:
    image: llmio/llmio:latest  # Go binary
    # Same env vars, volumes, ports

  worker:
    image: llmio/llmio:latest
    command: ["./llmio", "worker"]  # River job processor
```

**Keep unchanged:**
- nginx-proxy, nginx-proxy-acme (SSL)
- db (PostgreSQL - same schema)
- redis (same pub/sub channels)
- haraka (SMTP - same email format)
- channels, hocuspocus (same Redis protocol)
- pgbackups (same backup format)

### Go Storage Configuration

```go
// Storage backend selection
type StorageConfig struct {
    Backend string // "disk", "s3", "gcs", "do_spaces"

    // Disk
    Path string

    // S3/DO Spaces
    AccessKeyID     string
    SecretAccessKey string
    Bucket          string
    Region          string
    Endpoint        string // For S3-compatible

    // GCS
    Credentials string
    Project     string
}

func NewStorage(cfg StorageConfig) (Storage, error) {
    switch cfg.Backend {
    case "disk":
        return NewDiskStorage(cfg.Path)
    case "s3", "do_spaces":
        return NewS3Storage(cfg)
    case "gcs":
        return NewGCSStorage(cfg)
    default:
        return nil, fmt.Errorf("unknown storage backend: %s", cfg.Backend)
    }
}
```

### Go Background Jobs with River

```go
// River job processor (approved in CLAUDE.md)
import "github.com/riverqueue/river"

type PublishEventArgs struct {
    EventID int64 `json:"event_id"`
}

func (PublishEventArgs) Kind() string { return "publish_event" }

type PublishEventWorker struct {
    river.WorkerDefaults[PublishEventArgs]
    eventService *EventService
}

func (w *PublishEventWorker) Work(ctx context.Context, job *river.Job[PublishEventArgs]) error {
    return w.eventService.Publish(ctx, job.Args.EventID)
}

// Queue configuration
riverClient, _ := river.NewClient(riverpgxv5.New(pgxPool), &river.Config{
    Queues: map[string]river.QueueConfig{
        "critical": {MaxWorkers: 10},
        "high":     {MaxWorkers: 6},
        "default":  {MaxWorkers: 3},
        "low":      {MaxWorkers: 1},
    },
})
```

### Health Checks

Same patterns for Go:

```go
// Readiness probe
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    // Check DB
    if err := db.PingContext(r.Context()); err != nil {
        http.Error(w, "db unhealthy", 503)
        return
    }
    // Check Redis
    if err := redis.Ping(r.Context()).Err(); err != nil {
        http.Error(w, "redis unhealthy", 503)
        return
    }
    w.WriteHeader(200)
})
```

### Backup Compatibility

pgbackups container works unchanged - same pg_dump format:

```bash
# Backup command (from pgbackups)
pg_dump -Fc -h db -U loomio loomio > /backups/loomio_$(date +%Y%m%d).dump

# Restore works for Go app
pg_restore -h db -U loomio -d loomio /backups/loomio_20250201.dump
```
