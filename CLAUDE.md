# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **monorepo for rewriting Loomio** - a collaborative decision-making platform. The `orig/` directory contains the existing Rails 8 + Vue 3 codebase being analyzed; `discovery/` contains comprehensive specifications for the rewrite.

**Primary LLM Reference:** `discovery/loomio_rewrite_context.md` (~25K tokens) - read this first for complete context.

## Development Commands

### Version Management (mise)

This project uses [mise](https://mise.jdx.dev/) for tool version management with experimental monorepo mode.

```bash
mise install              # Install all tools (Go, Ruby 3.4.7, Node 22, pnpm)
```

### Backend (Rails API) - from `orig/loomio/`

```bash
bundle install            # Install Ruby dependencies
rails s                   # Start Rails server (port 3000)
rails c                   # Rails console
bundle exec rspec         # Run all tests
bundle exec rspec spec/path/to/file_spec.rb      # Run single test file
bundle exec rspec spec/path/to/file_spec.rb:42   # Run specific line
```

### Frontend (Vue 3 SPA) - from `orig/loomio/`

```bash
mise run frontend-serve   # Hot-reload dev server (port 5173 → proxies to 3000)
mise run frontend-build   # Production build → public/client3
mise run frontend-test    # E2E tests (Nightwatch)
mise run pnpm-install     # Install frontend dependencies
```

Or directly from `orig/loomio/vue/`:
```bash
pnpm install && pnpm run serve
```

### WebSocket Server - from `orig/loomio_channel_server/`

```bash
mise run serve-ws         # Start Socket.io server
mise run serve-hocuspocus # Start collaborative editing server
```

### Database Setup

```bash
createdb loomio_development
cd orig/loomio && rake db:setup
```

## Git Workflow

Conventional commits enforced via commitlint. Valid types:
`build`, `chore`, `ci`, `docs`, `feat`, `fix`, `perf`, `plan`, `refactor`, `revert`, `style`, `test`

Pre-commit hooks run `golangci-lint` on Go files.

## Architecture

### Core Domain (see `discovery/specifications/` for details)

| Concept | Description |
|---------|-------------|
| **Group** | Organization with members and permission flags |
| **Discussion** | Conversation thread (can belong to group or be "direct") |
| **Poll** | Decision tool (proposal, ranked choice, dot vote, etc.) |
| **Stance** | User's vote/position on a poll |
| **Event** | Activity record driving timelines, notifications, webhooks |

### Key Patterns

1. **Service Layer** - All mutations flow through `*Service` classes (`PollService.create`, `DiscussionService.update`)
2. **Event Sourcing** - Actions create Event records that publish to Redis → Socket.io → Vue clients
3. **Permission Flags** - Groups have 12 `members_can_*` boolean flags controlling capabilities
4. **Client-side ORM** - LokiJS mirrors Rails models with 28 record interfaces

### Request Flow

```
Vue SPA → REST /api/v1/* → Controller → authorize!(CanCanCan) → *Service.action() → Event.publish!
                                                                                        ↓
Vue SPA ← Socket.io (records) ← Redis pub/sub ← PublishEventWorker
```

### Directory Structure

```
discovery/                 # Rewrite specifications (read first!)
  ├── loomio_rewrite_context.md  # Executive summary (25K tokens)
  ├── specifications/            # 26 detailed spec files
  ├── openapi/                   # API documentation (~204 endpoints)
  └── schemas/                   # Database and request/response schemas

orig/loomio/               # Rails 8 API + Vue 3 frontend
  ├── app/
  │   ├── controllers/api/v1/    # REST endpoints (~30 controllers)
  │   ├── models/                # ActiveRecord + concerns
  │   ├── services/              # Business logic (*Service classes)
  │   └── workers/               # Sidekiq jobs
  └── vue/src/
      ├── components/            # 217 Vue components
      └── shared/
          ├── services/          # 35 services (records.js, session.js)
          ├── models/            # 31 client-side models
          └── interfaces/        # 28 LokiJS record interfaces

orig/loomio_channel_server/  # Node.js WebSocket server
```

## Technology Stack

| Layer | Technology |
|-------|------------|
| Backend | Rails 8 API-only, Ruby 3.4.7 |
| Frontend | Vue 3, Vite, Vuetify |
| Database | PostgreSQL with pg_search |
| Queue | Sidekiq + Redis |
| Real-time | Socket.io, Hocuspocus + Yjs |
| Client State | LokiJS in-memory DB |
| Testing | RSpec (backend), Nightwatch (E2E) |
