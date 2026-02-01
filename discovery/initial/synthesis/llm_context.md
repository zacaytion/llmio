# Loomio LLM Context Document

## Project Summary

Loomio is a collaborative decision-making tool for organizations. Users create groups, start discussions (threads), and run polls to reach decisions together. It's a Rails 8 API backend with a Vue 3 SPA frontend.

## Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Ruby 3.4, Rails 8, PostgreSQL |
| Frontend | Vue 3, Vuetify, Vite |
| Background Jobs | Sidekiq 7, Redis |
| Search | pg_search (PostgreSQL full-text) |
| Real-time | Hocuspocus (Yjs), Redis pub/sub |
| Rich Text | Tiptap with collaborative editing |
| Testing | RSpec (backend), Nightwatch (E2E) |

## Architecture Patterns

### Service Layer
All mutations go through service classes (`app/services/*_service.rb`). Services handle authorization, business logic, and event publishing. Controllers never modify models directly.

```
Controller -> Service.action(model:, actor:) -> authorize! -> save -> EventBus.broadcast -> Event.publish!
```

### Event Sourcing
Actions create Event records (42 STI subclasses) that drive notifications, activity feeds, and timelines. Events trigger side effects via concern composition:
- `Events::LiveUpdate` - Real-time UI updates
- `Events::Notify::InApp` - In-app notifications
- `Events::Notify::ByEmail` - Email notifications
- `Events::Notify::Chatbots` - Webhook delivery

### Authorization
CanCanCan with modular abilities in `app/models/ability/`. Group settings cascade permissions:
- `members_can_start_discussions`
- `members_can_raise_motions`
- `members_can_announce`
- etc.

### Query Objects
Complex queries in `app/queries/` with `visible_to(user:)` methods for authorization scoping.

### Frontend Store
LokiJS in-memory document database. Records interface in `vue/src/shared/interfaces/` mirror Rails models.

## Domain Model

### Core Entities

**User** - Account with email, name, avatar. Has memberships, stances, notifications.

**Group** - Organization/team. Has subgroups (self-referencing), memberships, discussions, polls. Settings control member permissions.

**Membership** - Links User to Group. Has admin/delegate roles, volume preference, invitation state.

**Discussion** - Threaded conversation. Belongs to Group (optional for direct discussions). Has comments, polls, events.

**DiscussionReader** - Per-user read state. Tracks last_read_at, read_ranges (compressed), volume, dismissed state.

**Comment** - Threaded within Discussion. Has parent_id for nesting, rich text body with mentions.

**Poll** - Decision/vote. Types: proposal, count, check, question, ranked_choice, score, meeting, dot_vote. Has options, stances, outcome.

**PollOption** - Voting choice. Has name, meaning (agree/disagree/block), icon.

**Stance** - User's vote. Links to Poll and participant. Has option_scores, reason, cast_at. Latest flag tracks current vote.

**Outcome** - Poll result summary. Created when poll closes.

**Event** - Activity record (STI). Links eventable (Poll, Comment, etc.) to Discussion timeline. Has sequence_id for ordering, position_key for threading.

**Notification** - User notification. Links Event to User via notifications table.

### Key Relationships

```
Group -> has_many :memberships -> User
Group -> has_many :discussions
Group -> has_many :polls
Discussion -> has_many :comments
Discussion -> has_many :polls
Discussion -> has_many :events
Poll -> has_many :poll_options
Poll -> has_many :stances -> User (as participant)
Event -> belongs_to :eventable (polymorphic)
Event -> has_many :notifications -> User
```

## API Structure

### Primary API: `/api/v1/`

| Resource | Key Endpoints |
|----------|--------------|
| sessions | POST (login), DELETE (logout) |
| registrations | POST (signup) |
| groups | CRUD, subgroups, export, token |
| memberships | CRUD, join_group, make_admin, set_volume |
| discussions | CRUD, dashboard, inbox, mark_as_read, close, move |
| comments | CRUD, discard |
| polls | CRUD, close, reopen, remind, receipts |
| stances | CRUD, uncast, make_admin, revoke |
| events | index, timeline, pin, unpin |
| search | index (full-text search) |
| documents | CRUD, for_group, for_discussion |

### Bot API: `/api/b2/`
External integrations with api_key auth. Create discussions, polls, comments; sync memberships.

### Admin API: `/api/b3/`
Server-side key auth. User deactivate/reactivate.

## Critical Business Rules

1. **Volume Cascade**: Notification volume resolves as Stance > DiscussionReader > Membership > User default

2. **Guest Access**: Users can access specific content without group membership via DiscussionReader or Stance with guest flag

3. **Poll Anonymity**: Anonymous polls scrub participant data on closing; cannot be reopened

4. **Event Threading**: Events have sequence_id (linear) and position_key (hierarchical) for timeline rendering

5. **Soft Delete**: Most models use `discarded_at` timestamp. `.kept` scope excludes discarded.

6. **Paper Trail**: Discussion, Poll, Comment, Outcome track versions for edit history

7. **Templates**: Groups have discussion and poll templates. System templates loaded from YAML, custom templates stored in DB.

## Key Files

| Purpose | Location |
|---------|----------|
| Routes | `config/routes.rb` |
| Services | `app/services/*_service.rb` |
| Abilities | `app/models/ability/*.rb` |
| Event Types | `app/models/events/*.rb` |
| Serializers | `app/serializers/*.rb` |
| Query Objects | `app/queries/*.rb` |
| Frontend Routes | `vue/src/routes.js` |
| Record Store | `vue/src/shared/services/records.js` |
| Model Interfaces | `vue/src/shared/interfaces/*.js` |

## Development Commands

```bash
# Backend
rails s                           # Start server (port 3000)
bundle exec rspec                 # Run tests
bundle exec rspec spec/path:line  # Single test

# Frontend
cd vue && npm run serve           # Dev server (port 8080)
cd vue && npm run test            # E2E tests
cd vue && npm run build           # Production build

# Database
rake db:setup                     # Create and seed
rake db:migrate                   # Run migrations
```

## External Services

| Service | Purpose |
|---------|---------|
| Redis | Caching, sessions, pub/sub, background jobs |
| S3/GCS | File storage |
| Sentry | Error tracking |
| Hocuspocus | Collaborative editing |
| SMTP | Outbound email |
| Google Translate | Content translation |
| OAuth/SAML | SSO authentication |

## Common Patterns

**Creating content**: Call service with actor, handle authorization failure (403), return event on success.

**Querying**: Use Query.visible_to(user:) for authorization, .filter(params:) for filtering.

**Real-time**: MessageChannelService.publish_models broadcasts to Redis, external service pushes to clients.

**Notifications**: Events include notification concerns, UsersByVolumeQuery filters recipients by preferences.
