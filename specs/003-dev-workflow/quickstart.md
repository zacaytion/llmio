# Quickstart: Local Development Workflow

## Prerequisites

- Podman and Podman Compose installed
- Go 1.25+ (managed via mise)
- golangci-lint (managed via mise)
- goimports (managed via mise)

## First-Time Setup

```bash
# 1. Copy environment template
cp .env.example .env

# 2. Start infrastructure services
make up

# 3. Wait for services to be healthy (check logs)
make logs

# 4. Install Go dependencies
make install

# 5. Run database migrations
make run-migrate

# 6. Verify setup
make test
```

## Daily Development

```bash
# Start services (if not running)
make up

# Run the server
make run-server

# In another terminal, watch for changes and run tests
make test

# Check code quality
make lint
make fmt
```

## Available Commands

Run `make` or `make help` for full list:

| Command | Description |
|---------|-------------|
| `make up` | Start all services (Postgres, Redis, PgAdmin, Mailpit) |
| `make down` | Stop all services |
| `make clean-volumes` | Stop services and delete Postgres/PgAdmin data for fresh start |
| `make logs` | Tail logs for all services |
| `make build` | Build all Go binaries |
| `make run-server` | Run the API server |
| `make run-migrate` | Run database migrations |
| `make test` | Run tests with coverage |
| `make coverage-view` | Open coverage report in browser |
| `make lint` | Run linter |
| `make lint-fix` | Run linter with auto-fix |
| `make fmt` | Format code |

## Service URLs

| Service | URL | Credentials |
|---------|-----|-------------|
| PostgreSQL | `localhost:5432` | See `.env` |
| Redis | `localhost:6379` | No auth |
| PgAdmin | http://localhost:5050 | See `.env` (PGADMIN_*) |
| Mailpit Web | http://localhost:8025 | No auth |
| Mailpit SMTP | `localhost:1025` | No auth |

## Troubleshooting

### Port already in use

Check if another process is using the port:
```bash
lsof -i :5432  # PostgreSQL
lsof -i :6379  # Redis
```

Stop conflicting services or change ports in `.env` (requires updating compose.yml).

### Services won't start

Check Podman is running:
```bash
podman machine list
podman machine start  # if needed
```

### PgAdmin can't connect to database

1. Ensure PostgreSQL is healthy: `make logs-postgres`
2. Check credentials match between `.env` and PgAdmin login
3. Database host inside containers is `postgres` (not `localhost`)
