# Loomio Rewrite: Technology Stack Design

## Context

Rewriting Loomio - a collaborative decision-making platform - from Rails 8 + Vue 3 to a Go + SvelteKit stack. Primary goal: **better developer experience** through faster iteration, clearer codebase, and easier onboarding.

Key constraints:
- Full-stack team comfortable with both backend and frontend
- Prefer simpler tooling over heavy frameworks
- Moderately interactive UI (real-time nice-to-have, not critical for most flows)
- Collaborative editing deferred to later phase
- Strong preference for code generation and type safety across the stack

## Technology Stack

### Backend

| Component | Technology | Notes |
|-----------|------------|-------|
| Language | Go 1.25.6 | |
| Web framework | [Huma](https://huma.rocks/) | OpenAPI-native, pragmatic conventions, sits on net/http |
| Database | PostgreSQL 18 | |
| DB access | [sqlc](https://sqlc.dev/) + pgx/v5 | Type-safe SQL, generates Go code |
| Migrations | [goose](https://github.com/pressly/goose) | Embedded in binary, programmatic execution |
| Background jobs | [River](https://riverqueue.com/) | Postgres-backed, native Go |
| Cache/pubsub | Redis | Carried over from original architecture |
| Email | [go-mail](https://github.com/wneessen/go-mail) | Direct SMTP sending |
| Sessions | Custom implementation | Cookie-based sessions with Redis/Postgres store |
| OAuth/SAML/LDAP | Custom via Huma | Implement as Huma middleware/operations |

### Frontend

| Component | Technology | Notes |
|-----------|------------|-------|
| Framework | [SvelteKit](https://kit.svelte.dev/) | Hybrid SSR, compiles to minimal JS |
| File uploads | [Uppy](https://uppy.io/) | Presigned URLs direct to cloud storage |
| Unit testing | [Vitest](https://vitest.dev/) | |
| Component testing | [Storybook](https://storybook.js.org/) | |
| E2E testing | [Playwright](https://playwright.dev/) | |

### Real-time

| Use case | Transport |
|----------|-----------|
| Notifications, vote updates | Server-Sent Events (SSE) |
| Chat, bidirectional features | WebSockets |

### API Workflow

```
OpenAPI specs (discovery/openapi/)
        ↓
   oapi-codegen
        ↓
Go types (request/response structs)
        ↓
   Huma handlers
        ↓
   Huma-generated OpenAPI (live spec)
        ↓
   TypeScript generator
        ↓
SvelteKit types
```

Spec-first workflow: edit OpenAPI specs, regenerate Go types, implement handlers.

### Testing Strategy

| Layer | Approach |
|-------|----------|
| Go backend | stdlib `testing`, table-driven tests, TDD workflow |
| Database logic | [pgTap](https://pgtap.org/) |
| Frontend units | Vitest |
| Components | Storybook |
| E2E | Playwright |

### File Storage

Uppy on frontend → presigned URLs from Go backend → direct upload to S3/GCS/DO Spaces

Go backend never handles file bytes, only generates presigned URLs and stores metadata.

## Key Decisions Rationale

### Why Huma over Echo/Chi/net-http?

Echo is solid but doesn't generate OpenAPI from code. With 204 endpoints already specified, we want the framework to enforce the spec. Huma validates requests against OpenAPI definitions and generates live documentation. Pragmatic conventions without excessive magic.

### Why Svelte over htmx/Astro?

- **htmx**: No type safety across the stack. For moderately interactive UIs with real-time elements, we'd fight it.
- **Astro**: Content-site focused. Adds indirection we don't need for an app.
- **Svelte**: Compiles to minimal JS, first-class TypeScript, hybrid SSR matches our preference, excellent DX.

### Why goose over golang-migrate/tern/dbmate?

- **golang-migrate**: Embeds well but quirky driver handling
- **tern**: pgx author's tool but no Go embedding
- **dbmate**: Shell-based, doesn't embed

goose: embeds via `goose.SetBaseFS()`, runs programmatically, supports plain SQL files, works with pgx/v5.

### Why spec-first over code-first?

We already have comprehensive OpenAPI specs in `discovery/openapi/`. Spec-first preserves that investment as the source of truth. Edit YAML → generate Go types → implement handlers. Single source of truth, full type chain to frontend.

## Deferred Decisions

- Collaborative editing (Hocuspocus/Yjs) - later phase
- Mobile strategy - responsive SvelteKit for now
- Matrix chatbot integration - needs investigation

## Implementation Notes

### Project Structure (proposed)

```
cmd/
  server/           # Main application entrypoint
internal/
  api/              # Huma operations, middleware
  auth/             # Sessions, OAuth, SAML, LDAP
  db/               # sqlc generated code, migrations
  jobs/             # River job definitions
  mail/             # Email templates and sending
  realtime/         # SSE and WebSocket handlers
pkg/                # Shared utilities (if needed)
web/                # SvelteKit application
migrations/         # SQL migration files (embedded)
openapi/            # Source OpenAPI specs
generated/          # oapi-codegen output
```

### Migration Path

1. Set up Go project skeleton with Huma, sqlc, goose, River
2. Generate Go types from existing OpenAPI specs
3. Port database schema via goose migrations
4. Implement core auth (sessions, login/logout)
5. Port API endpoints incrementally (start with read-only, then mutations)
6. Build SvelteKit frontend consuming the API
7. Add real-time features (SSE, then WebSockets for chat)
8. Port background jobs to River

## Verification

To validate this stack works together:

1. Scaffold minimal Go server with Huma + sqlc + goose
2. Create one migration, one query, one API endpoint
3. Generate TypeScript types from Huma's OpenAPI output
4. Build minimal SvelteKit page consuming the endpoint
5. Verify type errors surface at compile time on both ends
