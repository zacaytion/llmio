# Deployment

> Docker Compose services, configuration, and operations.

## Overview

Single-host Docker Compose deployment with 10 containers.

**Source:** `orig/loomio-deploy/docker-compose.yml`

## Services

| Service | Image | Port | Purpose |
|---------|-------|------|---------|
| nginx-proxy | nginxproxy/nginx-proxy:alpine | 80, 443 | Reverse proxy, SSL termination |
| nginx-proxy-acme | nginxproxy/acme-companion | - | Let's Encrypt certificates |
| app | loomio/loomio | 3000 (internal) | Rails application |
| worker | loomio/loomio | - | Sidekiq background jobs |
| db | postgres:17 | 5432 (internal) | PostgreSQL database |
| redis | redis:8.4 | 6379 (internal) | Cache, queue, pub/sub |
| haraka | loomio/haraka | 25 | SMTP for reply-by-email |
| channels | loomio/channels | 5000 (internal) | Socket.io real-time |
| hocuspocus | loomio/channels | 5000 (internal) | Yjs collaborative editing |
| pgbackups | prodrigestivill/postgres-backup-local | - | Automated backups |

## Container Dependencies

```
nginx-proxy ◀── nginx-proxy-acme
     ▲
     │
app ─┼── worker
 │   │
 │   ├── channels
 │   │
 │   └── hocuspocus
 │
 ├── db ◀── pgbackups
 │
 └── redis
```

## Volumes

| Path | Purpose |
|------|---------|
| ./uploads | User file uploads |
| ./storage | ActiveStorage files |
| ./files | Static files |
| ./plugins | Custom plugins |
| ./pgdata | PostgreSQL data |
| ./pgdumps | Database backups |
| certs | SSL certificates |
| acme | ACME challenge data |

## Environment Variables

**Source:** `orig/loomio-deploy/env_template` (230 lines)

### Core

| Variable | Purpose |
|----------|---------|
| CANONICAL_HOST | Primary domain |
| SECRET_COOKIE_SECRET | Session encryption |
| DEVISE_SECRET_KEY | Auth secrets |
| RAILS_MASTER_KEY | Credentials encryption |
| DATABASE_URL | PostgreSQL connection |
| REDIS_URL | Redis connection |

### Real-time

| Variable | Purpose |
|----------|---------|
| CHANNELS_URI | Socket.io URL (wss://...) |
| HOCUSPOCUS_URI | Hocuspocus URL (wss://...) |
| PRIVATE_APP_URL | Internal Rails URL for auth |

### Email

| Variable | Purpose |
|----------|---------|
| SMTP_DOMAIN | Outbound email domain |
| SMTP_USERNAME | SMTP auth |
| SMTP_PASSWORD | SMTP auth |
| REPLY_HOSTNAME | Reply-by-email domain |

### Storage

**5 Storage Backends** (source: `orig/loomio/config/storage.yml`)

| Backend | Service | Environment Variables |
|---------|---------|----------------------|
| Disk (test) | Disk | Root: `tmp/storage` |
| Disk (local) | Disk | Root: `storage/` |
| Amazon S3 | S3 | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_BUCKET`, `AWS_REGION` |
| DigitalOcean Spaces | S3 | `DO_ENDPOINT`, `DO_ACCESS_KEY_ID`, `DO_SECRET_ACCESS_KEY`, `DO_BUCKET` |
| Google Cloud Storage | GCS | `GCS_CREDENTIALS`, `GCS_PROJECT`, `GCS_BUCKET` |

### Features

| Variable | Purpose |
|----------|---------|
| FEATURES_DISABLE_DISCUSSIONS | Disable discussions |
| FEATURES_DISABLE_GROUPS | Disable groups |
| FEATURES_DISABLE_POLLS | Disable polls |
| MAX_ATTACHMENT_BYTES | Upload size limit |

## Shell Scripts

### create_env.sh

Generates `.env` from `env_template` with random secrets.

```bash
#!/bin/bash
cp env_template .env
SECRET=$(openssl rand -hex 32)
sed -i "s/SECRET_COOKIE_SECRET=/SECRET_COOKIE_SECRET=$SECRET/" .env
# ... generate other secrets
```

### create_backup.sh

Manual database backup with timestamp.

```bash
docker compose exec db pg_dump -U postgres loomio > ./pgdumps/backup_$(date +%Y%m%d%H%M%S).sql
```

### update.sh

Pull latest images and restart.

```bash
docker compose pull
docker compose up -d
```

### create_swapfile.sh

For low-RAM servers.

```bash
fallocate -l 4G /swapfile
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile
```

## DNS Requirements

| Record | Value | Purpose |
|--------|-------|---------|
| A | Server IP | Main app |
| MX | mail.domain.com | Reply-by-email |
| CNAME channels | main domain | Socket.io |
| CNAME hocuspocus | main domain | Collaborative editing |

## Health Checks

```yaml
app:
  healthcheck:
    test: ['CMD-SHELL', 'curl --fail http://localhost:3000/ || exit 1']

db:
  healthcheck:
    test: ['CMD', 'pg_isready', '-U', 'postgres']

redis:
  healthcheck:
    test: ['CMD', 'redis-cli', 'ping']
```

## Backup Schedule

**pgbackups service:**
- Daily at 3 AM
- 7-day retention
- Stored in `./pgdumps/`

## Initial Setup

1. Clone loomio-deploy
2. Run `./create_env.sh`
3. Edit `.env` with domain, email, etc.
4. Run `docker compose up -d`
5. Run `docker compose exec app rails db:setup`
6. Access at `https://your-domain.com`

## Upgrade Procedure

1. Backup database: `./create_backup.sh`
2. Pull latest: `docker compose pull`
3. Restart: `docker compose up -d`
4. Run migrations: `docker compose exec app rails db:migrate`

---
