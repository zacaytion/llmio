# Loomio Codebase Broad Overview

**Generated:** 2026-02-01
**Purpose:** Structural overview for reverse-engineering documentation project

---

## Table of Contents

1. [Project Summary](#1-project-summary)
2. [File Structure & Counts](#2-file-structure--counts)
3. [Database Overview](#3-database-overview)
4. [Key Dependencies](#4-key-dependencies)
5. [API Structure](#5-api-structure)
6. [Architectural Patterns Observed](#6-architectural-patterns-observed)
7. [External Services Detected](#7-external-services-detected)
8. [Domain Boundaries](#8-domain-boundaries)
9. [Complexity Indicators](#9-complexity-indicators)
10. [Open Questions](#10-open-questions)

---

## 1. Project Summary

### What is Loomio?

Loomio is a **collaborative decision-making tool** for organizations. From the README:

> "Loomio is a decision-making tool for collaborative organizations."

The application enables users to:
- Create and manage **groups** (organizations, teams)
- Start **discussions** (threads) for deliberation
- Run **polls** with various voting mechanisms to reach decisions
- Track activity through an **event-based system**
- Collaborate in real-time on documents

### Tech Stack

| Component | Technology | Version |
|-----------|------------|---------|
| Ruby | Ruby | 3.4.7 |
| Backend Framework | Rails | ~> 8.0.0 |
| Frontend Framework | Vue.js | 3.5.27 |
| UI Framework | Vuetify | latest |
| Database | PostgreSQL | (via `pg` gem) |
| Background Jobs | Sidekiq | ~> 7.0 |
| Search | pg_search | PostgreSQL full-text search |
| Real-time Collaboration | Tiptap + Yjs + Hocuspocus | 3.16.0+ |
| Build Tool (Frontend) | Vite | 7.3.1 |
| E2E Testing | Nightwatch | 3.15.0 |
| Backend Testing | RSpec | 7.1.1 |

### License

GNU Affero General Public License (AGPL)

---

## 2. File Structure & Counts

### Backend Ruby Files (app/ directory)

| Directory | File Count | Description |
|-----------|------------|-------------|
| **Total Ruby Files** | 442 | All .rb files in app/ |
| app/models/ | 152 | Domain models and concerns |
| app/models/ (top-level only) | 49 | Core model classes |
| app/models/concerns/ | 36 | Shared model behaviors |
| app/models/events/ | 42 | Event STI subclasses |
| app/models/ability/ | 23 | CanCanCan ability classes |
| app/controllers/ | 89 | All controller files |
| app/services/ | 44 | Service objects |
| app/serializers/ | 65 | JSON serializers |
| app/workers/ | 38 | Sidekiq background jobs |
| app/mailers/ | 7 | Email mailers |
| app/queries/ | 8 | Query objects |
| app/admin/ | 5 | ActiveAdmin resources |
| app/extras/ | ~10 | Utility classes and clients |

### Controller Breakdown by API Version

| Namespace | File Count | Purpose |
|-----------|------------|---------|
| api/v1/ | 39 | Primary internal API for Vue SPA |
| api/b2/ | 5 | External bot/integration API v2 |
| api/b3/ | 1 | External bot/integration API v3 (users) |
| identities/ | 5 | OAuth provider controllers |
| dev/ | 5 | Development/testing controllers |
| (root controllers) | ~20 | Non-API controllers |

### Frontend Vue Files (vue/src/)

| Directory | File Count | Description |
|-----------|------------|-------------|
| components/ | 217 | Vue single-file components |
| shared/ | 117 | Shared JS services/utilities |
| shared/services/ | 35 | API clients, session, utilities |
| shared/interfaces/ | 30 | LokiJS record interfaces |
| shared/models/ | ~30 | Client-side model definitions |

### Vue Component Organization

Components are organized by feature domain:
- `auth/` - Authentication UI
- `group/` - Group management
- `discussion/` - Discussion/thread views
- `poll/` - Poll creation and voting
- `poll_template/` - Poll template management
- `thread/` - Thread navigation
- `strand/` - Event timeline display
- `common/` - Shared UI components
- `lmo_textarea/` - Rich text editor (Tiptap)
- `tags/` - Tagging system
- `tasks/` - Task management
- `inbox/` - Notification inbox
- `dashboard/` - Dashboard views
- `sidebar/` - Navigation sidebar
- `search/` - Search interface
- `profile/` - User profile
- `email_settings/` - Email preferences
- `report/` - Reporting

### Test Files

| Directory | File Count | Description |
|-----------|------------|-------------|
| spec/ (total) | 116 | RSpec test files |
| spec/models/ | 25 | Model specs |
| spec/controllers/ | 46 | Controller specs |
| spec/services/ | 20 | Service specs |
| spec/factories.rb | 317 lines | FactoryBot definitions |
| vue/tests/e2e/specs/ | 14 | Nightwatch E2E tests |

### Internationalization

- **Locale files:** 49 files in config/locales/
- Languages supported include: English, Spanish, French, German, Italian, Japanese, Portuguese, Russian, and many others

---

## 3. Database Overview

### Table Count

**Total Tables: 56**

### Table Listing with Column Complexity

Tables are sorted by complexity (column count indicates data richness):

| Table | Approximate Columns | Purpose |
|-------|---------------------|---------|
| **users** | ~50 | User accounts and preferences |
| **groups** | ~55 | Organizations/teams |
| **polls** | ~45 | Decision-making polls |
| **discussions** | ~35 | Discussion threads |
| **stances** | ~25 | User votes on polls |
| **memberships** | ~20 | Group membership records |
| **events** | ~18 | Activity event log |
| **comments** | ~18 | Discussion comments |
| **poll_templates** | ~35 | Reusable poll configurations |
| **discussion_templates** | ~22 | Reusable discussion configurations |
| **outcomes** | ~15 | Poll outcome records |
| **poll_options** | ~14 | Options within polls |
| **discussion_readers** | ~15 | Per-user read state tracking |
| **notifications** | ~8 | User notifications |
| **documents** | ~12 | Attached documents |
| **reactions** | ~6 | Emoji reactions |
| **tags** | ~8 | Tagging system |
| **taggings** | ~5 | Tag associations |
| **tasks** | ~12 | Task items |
| **subscriptions** | ~18 | Billing/subscription info |
| **chatbots** | ~10 | Webhook/chatbot integrations |
| **webhooks** | ~12 | Outbound webhooks |
| **versions** | ~6 | Paper Trail audit history |
| **translations** | ~5 | Translated content |
| **login_tokens** | ~7 | Magic link tokens |
| **membership_requests** | ~10 | Pending join requests |
| **omniauth_identities** | ~10 | OAuth identity records |
| **oauth_applications** | ~12 | OAuth 2.0 applications |
| **oauth_access_tokens** | ~8 | OAuth tokens |
| **oauth_access_grants** | ~7 | OAuth grants |
| **received_emails** | ~8 | Inbound email records |
| **pg_search_documents** | ~10 | Full-text search index |
| **demos** | ~8 | Demo group data |
| **group_surveys** | ~12 | Group onboarding surveys |
| **blocked_domains** | ~2 | Spam domain blocklist |
| **forward_email_rules** | ~4 | Email forwarding rules |
| **group_identities** | ~5 | SSO identity links |
| **member_email_aliases** | ~7 | Email alias config |
| **stance_choices** | ~5 | Individual vote choices |
| **stance_receipts** | ~6 | Vote receipt tracking |

### Infrastructure Tables

- `active_storage_blobs`, `active_storage_attachments`, `active_storage_variant_records` - Rails Active Storage
- `active_admin_comments` - ActiveAdmin
- `action_mailbox_inbound_emails` - Action Mailbox
- `blazer_*` (5 tables) - Blazer analytics dashboard
- `cohorts` - User cohort tracking
- `default_group_covers` - Default group cover images
- `partition_sequences` - Sequence generation

### PostgreSQL Extensions Used

- `citext` - Case-insensitive text
- `hstore` - Key-value storage
- `pg_stat_statements` - Query statistics
- `pgcrypto` - Cryptographic functions
- `plpgsql` - Procedural language

### Key Foreign Key Relationships

The schema shows these primary relationship patterns:

**Core Hierarchy:**
- Group -> parent Group (self-referencing for subgroups)
- Discussion -> Group
- Poll -> Discussion (optional)
- Poll -> Group
- Comment -> Discussion
- Event -> Discussion

**User Relationships:**
- Membership -> User, Group
- DiscussionReader -> User, Discussion
- Stance -> Poll, User (as participant)
- Notification -> User, Event

**Template Relationships:**
- Discussion -> DiscussionTemplate
- Poll -> PollTemplate
- DiscussionTemplate -> Group
- PollTemplate -> Group

---

## 4. Key Dependencies

### Backend (Gemfile)

#### Authentication & Authorization

| Gem | Purpose |
|-----|---------|
| `devise` (~> 4.9.4) | User authentication |
| `devise-i18n` | Devise translations |
| `devise-pwned_password` | Password breach checking |
| `cancancan` | Authorization/abilities |
| `ruby-saml` | SAML SSO support |

#### Background Processing

| Gem | Purpose |
|-----|---------|
| `sidekiq` (~> 7.0) | Background job processing |

#### Data & Storage

| Gem | Purpose |
|-----|---------|
| `pg` | PostgreSQL adapter |
| `active_record_extended` | Extended AR features |
| `pg_search` | Full-text search |
| `paper_trail` (~> 17.0.0) | Model versioning/audit |
| `aws-sdk-s3` | AWS S3 storage |
| `google-cloud-storage` | Google Cloud Storage |
| `image_processing` | Image manipulation |
| `ruby-vips` | Image processing library |

#### API & Serialization

| Gem | Purpose |
|-----|---------|
| `active_model_serializers` (~> 0.8.1) | JSON serialization |
| `friendly_id` (~> 5.6.0) | Slug generation |

#### Real-time & External Services

| Gem | Purpose |
|-----|---------|
| `httparty` | HTTP client |
| `google-cloud-translate` | Translation API |
| `ruby-openai` | OpenAI integration |
| `maxminddb` | GeoIP lookup |

#### Content Processing

| Gem | Purpose |
|-----|---------|
| `redcarpet` | Markdown rendering |
| `nokogiri` | HTML parsing |
| `reverse_markdown` | HTML to Markdown |
| `premailer-rails` | Email CSS inlining |
| `icalendar` | Calendar event generation |
| `video_info` | Video metadata extraction |

#### Monitoring & Admin

| Gem | Purpose |
|-----|---------|
| `sentry-ruby`, `sentry-rails`, `sentry-sidekiq` | Error tracking |
| `activeadmin` (~> 3.4.0) | Admin interface |
| `blazer` | SQL analytics dashboard |
| `lograge` | Structured logging |

#### Security & Rate Limiting

| Gem | Purpose |
|-----|---------|
| `rack-attack` | Rate limiting |

#### Other Notable Gems

| Gem | Purpose |
|-----|---------|
| `discard` | Soft delete |
| `discriminator` | STI helpers |
| `custom_counter_cache` | Counter caches |
| `redis-objects` | Redis data structures |
| `twitter-text` | @mention parsing |
| `cld` | Language detection |

### Frontend (package.json)

#### Core Framework

| Package | Version | Purpose |
|---------|---------|---------|
| `vue` | ^3.5.27 | Frontend framework |
| `vue-router` | ^4.2.5 | Client-side routing |
| `vue-i18n` | 9.11.0 | Internationalization |
| `vuetify` | latest | Material Design UI |

#### State Management

| Package | Purpose |
|---------|---------|
| `lokijs` | 1.5.12 | In-memory document database |

#### Rich Text Editing (Tiptap Suite)

| Package | Purpose |
|---------|---------|
| `@tiptap/core` | Rich text editor core |
| `@tiptap/vue-3` | Vue 3 integration |
| `@tiptap/extension-*` (25+) | Editor extensions |
| `@tiptap/extension-collaboration` | Real-time collaboration |
| `@tiptap/extension-collaboration-caret` | Collaborative cursors |
| `@tiptap/suggestion` | Mention suggestions |

#### Real-time Collaboration

| Package | Purpose |
|---------|---------|
| `@hocuspocus/provider` | Hocuspocus client for Yjs |
| `yjs` | CRDT implementation |
| `y-indexeddb` | Yjs IndexedDB persistence |
| `socket.io-client` | WebSocket client |

#### Data Visualization

| Package | Purpose |
|---------|---------|
| `chart.js` | Charting library |
| `vue-chartjs` | Vue Chart.js wrapper |

#### Utilities

| Package | Purpose |
|---------|---------|
| `date-fns`, `date-fns-tz` | Date manipulation |
| `lodash-es` | Utility functions |
| `marked` | Markdown parsing |
| `md5` | Hash generation |
| `deepmerge` | Object merging |
| `turndown` | HTML to Markdown |

#### Monitoring

| Package | Purpose |
|---------|---------|
| `@sentry/vue`, `@sentry/browser` | Error tracking |
| `plausible-tracker` | Privacy-focused analytics |

#### File Handling

| Package | Purpose |
|---------|---------|
| `activestorage` | Rails Active Storage client |
| `pretty-bytes` | File size formatting |

---

## 5. API Structure

### API Namespaces

The application exposes multiple API namespaces with distinct purposes:

#### `/api/v1/` - Primary Internal API

Used by the Vue SPA. This is the main API with 39 controller files.

**Resource Endpoints:**

| Resource | Actions | Special Endpoints |
|----------|---------|-------------------|
| `groups` | index, show, create, update, destroy | token, reset_token, subgroups, export, export_csv, upload_photo, count_explore_results, suggest_handle |
| `discussions` | show, index, create, update | mark_as_seen, dismiss, recall, set_volume, pin/unpin, move, mark_as_read, close, reopen, discard, move_comments, history, search, dashboard, inbox, direct |
| `polls` | show, index, create, update | receipts, remind, discard, close, reopen, add_to_thread, voters, closed |
| `stances` | index, create, update | uncast, invite, users, my_stances, make_admin, remove_admin, revoke |
| `comments` | create, update, destroy | discard, undiscard |
| `memberships` | index, create, update, destroy | user_name, join_group, for_user, autocomplete, my_memberships, invitables, undecided, make_admin, remove_admin, make_delegate, remove_delegate, save_experience, resend, set_volume |
| `events` | index | count, pin, unpin, comment, position_keys, timeline, remove_from_thread |
| `notifications` | index | viewed |
| `profile` | show, index | time_zones, me, groups, email_status, email_exists, update_profile, upload_avatar, save_experience, deactivate, remind |
| `documents` | create, update, destroy, index | for_group, for_discussion |
| `reactions` | create, update, index, destroy | - |
| `tags` | create, update, destroy | priority |
| `search` | index | - |
| `outcomes` | create, update | - |
| `announcements` | create | audience, count, new_member_count, search, history, users_notified_count |
| `discussion_templates` | create, index, show, update, destroy | browse_tags, browse, hide, unhide, discard, undiscard, positions |
| `poll_templates` | index, create, update, show, destroy | hide, unhide, discard, undiscard, positions, settings |
| `discussion_readers` | index | remove_admin, make_admin, resend, revoke |
| `webhooks` | create, destroy, index, update | - |
| `chatbots` | create, destroy, index, update | test |
| `tasks` | index | update_done, mark_as_done, mark_as_not_done |
| `received_emails` | index | aliases, destroy_alias, allow, block |
| `membership_requests` | create | my_pending, pending, previous, approve, ignore |
| `login_tokens` | create | - |
| `contact_messages` | create | - |
| `versions` | - | show (collection) |
| `translations` | - | inline |
| `reports` | index | - |
| `trials` | create | - |
| `attachments` | index, destroy | - |
| `boot` | - | site, user, version |
| `demos` | index | clone |
| `link_previews` | create | - |
| `mentions` | index | - |
| `sessions` | create, destroy | unauthorized |
| `registrations` | create | oauth |
| `identities` | - | command (dynamic) |

#### `/api/b1/` - Bot API v1 (deprecated, routes to b2)

Limited external API for integrations:
- `discussions` (create, show)
- `polls` (create, show)
- `memberships` (index, create)

#### `/api/b2/` - Bot API v2

External integration API with 5 controllers:
- `discussions` (create, show)
- `polls` (create, show)
- `memberships` (index, create)
- `comments` (create)

#### `/api/b3/` - Bot API v3

User management API with 1 controller:
- `users` (deactivate, reactivate)

#### `/api/hocuspocus` - Real-time Collaboration

Single endpoint for Hocuspocus collaborative editing authentication.

### Non-API Routes

**Authentication:**
- OAuth provider routes (Google, Nextcloud, generic OAuth, SAML)
- Devise user authentication routes
- Login token routes
- Membership invitation routes

**Email Actions:**
- Unsubscribe
- Set volume (group, discussion, poll)
- Mark as read

**Content Pages:**
- Group pages (`g/:key`)
- Discussion pages (`d/:key`)
- Poll pages (`p/:key`)
- User profiles (`u/:username`)
- Dashboard, inbox, explore, profile

**Admin:**
- ActiveAdmin at `/admin`
- Sidekiq at `/admin/sidekiq`
- Blazer at `/admin/blazer`

**Utilities:**
- Direct uploads
- Manifests
- Robots.txt
- Sitemap
- Help/API documentation

---

## 6. Architectural Patterns Observed

### 6.1 Service Object Pattern

All mutations go through service classes in `app/services/`. Services follow a consistent pattern:

**Structure:**
- Named as `{Model}Service` (e.g., `PollService`, `DiscussionService`)
- Class methods for actions: `create`, `update`, `destroy`, `close`, etc.
- Accept keyword arguments including `actor:` for authorization
- Return Event objects for successful operations
- Handle authorization via CanCanCan abilities

**Service Files (44 total):**
- `poll_service.rb` (454 lines) - Most complex
- `report_service.rb` (472 lines) - Analytics/reporting
- `discussion_service.rb` (280 lines)
- `group_export_service.rb` (333 lines)
- `membership_service.rb` (208 lines)
- `received_email_service.rb` (234 lines)
- And 38 more...

### 6.2 Event Sourcing / Activity Tracking

The application uses an event-sourcing-inspired pattern where all significant actions create Event records.

**Event Model:**
- Base `Event` class in `app/models/event.rb` (214 lines)
- STI subclasses in `app/models/events/` (42 event types)
- Events drive notifications, activity feeds, and timeline displays

**Event Types (42):**
- Discussion events: new_discussion, discussion_edited, discussion_closed, discussion_reopened, discussion_moved, discussion_forked, discussion_announced
- Comment events: new_comment, comment_edited, comment_replied_to
- Poll events: poll_created, poll_edited, poll_closed_by_user, poll_expired, poll_reopened, poll_closing_soon, poll_announced, poll_option_added, poll_reminder
- Stance events: stance_created, stance_updated
- Outcome events: outcome_created, outcome_updated, outcome_announced, outcome_review_due
- Membership events: membership_created, membership_resent, invitation_accepted, user_added_to_group, user_joined_group, membership_request_approved, membership_requested
- User events: user_mentioned, group_mentioned, user_reactivated
- Admin events: new_coordinator, new_delegate
- Other: announcement_resend, reaction_created, unknown_sender

**Event Publishing:**
- `Event.publish!` creates event and enqueues `PublishEventWorker`
- `EventBus` broadcasts events for side effects (reader updates, real-time sync)
- Events store `eventable_version_id` for audit trail linking to Paper Trail

### 6.3 Single Table Inheritance (STI)

STI is used extensively for polymorphic behavior:

**Events:**
- Base `Event` class with `kind` column
- Subclasses in `app/models/events/` override behavior via concerns

**Groups:**
- `FormalGroup` inherits from `Group`
- `GuestGroup` for guest access scenarios

**Users:**
- `AnonymousUser` for unauthenticated scenarios
- `LoggedOutUser` for session-less operations

### 6.4 Concerns/Mixins

Model concerns in `app/models/concerns/` (36 files) provide shared behavior:

**Rich Text & Content:**
- `HasRichText` - Rich text formatting support
- `HasMentions` - @mention parsing
- `HasTags` - Tagging support
- `Translatable` - Multi-language translation
- `Searchable` - Full-text search indexing

**Events & Activity:**
- `HasEvents` - Event association
- `HasCreatedEvent` - Auto-create event on record creation
- `MessageChannel` - Real-time messaging

**User & Access:**
- `HasAvatar`, `AvatarInitials` - Avatar handling
- `HasVolume` - Notification volume preferences
- `HasExperiences` - Feature flag/experience tracking
- `HasTokens` - Token generation

**Data & State:**
- `HasCustomFields` - JSONB custom fields
- `HasDefaults` - Default value setting
- `HasTimeframe` - Date range scoping
- `SelfReferencing` - Tree structures
- `ReadableUnguessableUrls` - Slug/key generation

**Privacy:**
- `GroupPrivacy` - Group visibility settings

**Export:**
- `DiscussionExportRelations` - Discussion export
- `GroupExportRelations` - Group export

**Spam Prevention:**
- `NoSpam` - Anti-spam measures
- `NoForbiddenEmails` - Email validation

### 6.5 Authorization (CanCanCan)

Authorization uses CanCanCan with per-model ability classes:

**Ability Classes (23):**
Located in `app/models/ability/`:
- `Ability::Base` - Base class
- `Ability::User`, `Ability::Group`, `Ability::Discussion`, `Ability::Poll`
- `Ability::Comment`, `Ability::Stance`, `Ability::Outcome`
- `Ability::Membership`, `Ability::MembershipRequest`
- `Ability::Document`, `Ability::Reaction`, `Ability::Event`
- `Ability::Tag`, `Ability::Task`
- `Ability::DiscussionTemplate`, `Ability::PollTemplate`
- `Ability::DiscussionReader`, `Ability::ReceivedEmail`
- `Ability::Chatbot`, `Ability::Attachment`
- `Ability::PollOption`, `Ability::Identity`

### 6.6 API Base Controller Pattern

API controllers inherit from `SnorlaxBase` (280 lines) which provides:
- Standard CRUD actions (show, index, create, update, destroy)
- Automatic service delegation
- Pagination and filtering
- Serialization with RecordCache
- Error handling with standard responses
- Timeframe scoping

### 6.7 Paper Trail Versioning

Models with audit requirements include Paper Trail:
- `Discussion`, `Poll`, `Comment`, `Outcome`
- Stores changes in `versions` table with JSONB `object_changes`
- `eventable_version_id` links events to specific versions

### 6.8 Query Objects

Query objects in `app/queries/` (8 files) encapsulate complex queries:
- `AttachmentQuery`
- `ContactableQuery`
- `DiscussionQuery`
- `GroupQuery`
- `MembershipQuery`
- `PollQuery`
- `ReactionQuery`
- `UserQuery`

Additional queries in `app/extras/queries/`:
- `UsersByVolumeQuery` - Notification recipient filtering

### 6.9 Frontend Record Store (LokiJS)

The Vue frontend uses LokiJS as an in-memory document database:

**RecordStore Pattern:**
- `records.js` initializes LokiJS and registers interfaces
- 28 record interfaces mirror backend models
- Interfaces define computed properties, relationships, and API methods

**Interfaces:**
- Comment, Chatbot, Discussion, DiscussionTemplate, DiscussionReader
- Event, Group, Membership, MembershipRequest, Notification
- User, Version, Translation, Session, Registration
- Poll, PollTemplate, PollOption, Stance, Outcome
- ContactMessage, Reaction, Document, Attachment
- LoginToken, MessageChannel, Tag, Task, Webhook, ReceivedEmail

---

## 7. External Services Detected

### 7.1 OAuth Providers

Configured in `config/providers.yml`:
- **oauth** - Generic OAuth 2.0
- **saml** - SAML 2.0 SSO
- **google** - Google OAuth
- **nextcloud** - Nextcloud OAuth

Controller support in `app/controllers/identities/`:
- `base_controller.rb`
- `google_controller.rb`
- `nextcloud_controller.rb`
- `oauth_controller.rb`
- `saml_controller.rb`

### 7.2 Cloud Storage

- **AWS S3** (`aws-sdk-s3`)
- **Google Cloud Storage** (`google-cloud-storage`)
- Active Storage for file handling

### 7.3 Error Tracking

- **Sentry** (`sentry-ruby`, `sentry-rails`, `sentry-sidekiq`)
- Configured in `config/initializers/sentry.rb`
- Sample rate configurable via `SENTRY_SAMPLE_RATE`

### 7.4 Analytics

- **Blazer** - Internal SQL analytics dashboard
- **Plausible** (`plausible-tracker`) - Privacy-focused web analytics

### 7.5 Translation Service

- **Google Cloud Translate** (`google-cloud-translate`)
- `TranslationService` handles automatic translation

### 7.6 AI/ML Services

- **OpenAI** (`ruby-openai`) - Likely for transcription or content generation
- `TranscriptionService` - Audio/video transcription

### 7.7 GeoIP

- **MaxMind** (`maxminddb`) - Geographic location lookup
- `GeoLocationWorker` for async processing

### 7.8 Real-time Collaboration

- **Hocuspocus** - Yjs WebSocket server for collaborative editing
- Channels service for real-time notifications
- URLs configured via `CHANNELS_URL` and `HOCUSPOCUS_URL`

### 7.9 Email Services

- **Premailer** - Email CSS inlining
- **Action Mailbox** - Inbound email processing
- Reply handling via `REPLY_HOSTNAME`

### 7.10 Webhook/Chatbot Integrations

Clients in `app/extras/clients/`:
- `webhook.rb` - Generic webhook client
- `google.rb` - Google integration
- `nextcloud.rb` - Nextcloud integration
- `oauth.rb` - OAuth client base
- `request.rb` - HTTP request base

Chatbot model supports:
- Webhook-based notifications
- Multiple webhook formats (markdown, etc.)
- Event-driven notifications

### 7.11 Billing/Subscriptions

- **Chargify** - Subscription management (when `CHARGIFY_API_KEY` is set)
- Optional LoomioSubs engine (`engines/loomio_subs`)

---

## 8. Domain Boundaries

Based on the codebase structure, here are the 10 identified domain boundaries with their primary files:

### 8.1 Auth Domain

**Purpose:** User authentication, sessions, OAuth, SAML

**Backend Files:**
- `app/models/user.rb`
- `app/models/identity.rb`
- `app/models/login_token.rb`
- `app/models/logged_out_user.rb`
- `app/models/anonymous_user.rb`
- `app/services/user_service.rb`
- `app/services/login_token_service.rb`
- `app/controllers/identities/*.rb`
- `app/controllers/api/v1/sessions_controller.rb`
- `app/controllers/api/v1/registrations_controller.rb`
- `app/controllers/login_tokens_controller.rb`
- `config/initializers/devise.rb`

**Frontend Files:**
- `vue/src/components/auth/`
- `vue/src/shared/services/auth_service.js`
- `vue/src/shared/services/session.js`
- `vue/src/shared/interfaces/session_records_interface.js`
- `vue/src/shared/interfaces/registration_records_interface.js`
- `vue/src/shared/interfaces/login_token_records_interface.js`

### 8.2 Groups Domain

**Purpose:** Organization/group management, memberships, subgroups

**Backend Files:**
- `app/models/group.rb`
- `app/models/formal_group.rb`
- `app/models/guest_group.rb`
- `app/models/membership.rb`
- `app/models/membership_request.rb`
- `app/models/subscription.rb`
- `app/services/group_service.rb`
- `app/services/group_service/` (directory)
- `app/services/membership_service.rb`
- `app/services/membership_request_service.rb`
- `app/controllers/api/v1/groups_controller.rb`
- `app/controllers/api/v1/memberships_controller.rb`
- `app/controllers/api/v1/membership_requests_controller.rb`
- `app/models/ability/group.rb`
- `app/models/ability/membership.rb`
- `app/models/ability/membership_request.rb`
- `app/queries/group_query.rb`
- `app/queries/membership_query.rb`
- `app/serializers/group_serializer.rb`
- `app/serializers/membership_serializer.rb`

**Frontend Files:**
- `vue/src/components/group/`
- `vue/src/components/start_group/`
- `vue/src/shared/services/group_service.js`
- `vue/src/shared/interfaces/group_records_interface.js`
- `vue/src/shared/interfaces/membership_records_interface.js`
- `vue/src/shared/interfaces/membership_request_records_interface.js`

### 8.3 Discussions Domain

**Purpose:** Discussion threads, comments, thread navigation

**Backend Files:**
- `app/models/discussion.rb`
- `app/models/discussion_reader.rb`
- `app/models/comment.rb`
- `app/models/null_discussion.rb`
- `app/services/discussion_service.rb`
- `app/services/discussion_reader_service.rb`
- `app/services/comment_service.rb`
- `app/controllers/api/v1/discussions_controller.rb`
- `app/controllers/api/v1/discussion_readers_controller.rb`
- `app/controllers/api/v1/comments_controller.rb`
- `app/models/ability/discussion.rb`
- `app/models/ability/discussion_reader.rb`
- `app/models/ability/comment.rb`
- `app/queries/discussion_query.rb`
- `app/serializers/discussion_serializer.rb`
- `app/serializers/comment_serializer.rb`
- `app/mailers/group_mailer.rb`

**Frontend Files:**
- `vue/src/components/discussion/`
- `vue/src/components/thread/`
- `vue/src/components/strand/`
- `vue/src/components/lmo_textarea/`
- `vue/src/shared/services/discussion_service.js`
- `vue/src/shared/services/discussion_reader_service.js`
- `vue/src/shared/services/comment_service.js`
- `vue/src/shared/interfaces/discussion_records_interface.js`
- `vue/src/shared/interfaces/discussion_reader_records_interface.js`
- `vue/src/shared/interfaces/comment_records_interface.js`

### 8.4 Polls Domain

**Purpose:** Polls, voting, stances, outcomes

**Backend Files:**
- `app/models/poll.rb`
- `app/models/poll_option.rb`
- `app/models/stance.rb`
- `app/models/stance_choice.rb`
- `app/models/outcome.rb`
- `app/models/null_poll.rb`
- `app/services/poll_service.rb`
- `app/services/stance_service.rb`
- `app/services/outcome_service.rb`
- `app/controllers/api/v1/polls_controller.rb`
- `app/controllers/api/v1/stances_controller.rb`
- `app/controllers/api/v1/outcomes_controller.rb`
- `app/models/ability/poll.rb`
- `app/models/ability/poll_option.rb`
- `app/models/ability/stance.rb`
- `app/models/ability/outcome.rb`
- `app/queries/poll_query.rb`
- `app/serializers/poll_serializer.rb`
- `app/serializers/stance_serializer.rb`
- `app/serializers/outcome_serializer.rb`
- `config/poll_types.yml`

**Frontend Files:**
- `vue/src/components/poll/`
- `vue/src/shared/services/poll_service.js`
- `vue/src/shared/services/stance_service.js`
- `vue/src/shared/services/outcome_service.js`
- `vue/src/shared/interfaces/poll_records_interface.js`
- `vue/src/shared/interfaces/poll_option_records_interface.js`
- `vue/src/shared/interfaces/stance_records_interface.js`
- `vue/src/shared/interfaces/outcome_records_interface.js`

### 8.5 Events Domain

**Purpose:** Activity tracking, notifications, timeline

**Backend Files:**
- `app/models/event.rb`
- `app/models/events/*.rb` (42 event classes)
- `app/models/notification.rb`
- `app/models/concerns/events/`
- `app/services/event_service.rb`
- `app/services/notification_service.rb`
- `app/controllers/api/v1/events_controller.rb`
- `app/controllers/api/v1/notifications_controller.rb`
- `app/models/ability/event.rb`
- `app/serializers/event_serializer.rb`
- `app/serializers/notification_serializer.rb`
- `app/workers/publish_event_worker.rb`
- `config/initializers/event_bus.rb`
- `lib/event_bus.rb`

**Frontend Files:**
- `vue/src/components/strand/` (event timeline)
- `vue/src/shared/services/event_service.js`
- `vue/src/shared/services/event_bus.js`
- `vue/src/shared/interfaces/event_records_interface.js`
- `vue/src/shared/interfaces/notification_records_interface.js`

### 8.6 Documents Domain

**Purpose:** File attachments, documents, Active Storage

**Backend Files:**
- `app/models/document.rb`
- `app/models/attachment.rb`
- `app/services/document_service.rb`
- `app/controllers/api/v1/documents_controller.rb`
- `app/controllers/api/v1/attachments_controller.rb`
- `app/controllers/direct_uploads_controller.rb`
- `app/models/ability/document.rb`
- `app/models/ability/attachment.rb`
- `app/serializers/document_serializer.rb`
- `app/serializers/attachment_serializer.rb`
- `app/workers/download_attachment_worker.rb`
- `app/workers/attach_document_worker.rb`
- `config/storage.yml`

**Frontend Files:**
- `vue/src/components/document/`
- `vue/src/shared/services/attachment_service.js`
- `vue/src/shared/services/file_uploader.js`
- `vue/src/shared/interfaces/document_records_interface.js`
- `vue/src/shared/interfaces/attachment_records_interface.js`

### 8.7 Search Domain

**Purpose:** Full-text search across content

**Backend Files:**
- `app/models/search_result.rb`
- `app/models/pg_search_document.rb` (implicit from pg_search)
- `app/models/concerns/searchable.rb`
- `app/services/search_service.rb`
- `app/controllers/api/v1/search_controller.rb`
- `app/serializers/search_result_serializer.rb`
- `config/initializers/pg_search.rb`

**Frontend Files:**
- `vue/src/components/search/`

### 8.8 Export Domain

**Purpose:** Data export (group, discussion, poll)

**Backend Files:**
- `app/services/group_export_service.rb`
- `app/extras/group_exporter.rb`
- `app/extras/poll_exporter.rb`
- `app/models/concerns/group_export_relations.rb`
- `app/models/concerns/discussion_export_relations.rb`
- `app/workers/group_export_worker.rb`
- `app/workers/group_export_csv_worker.rb`
- `app/controllers/groups_controller.rb` (export action)
- `app/controllers/polls_controller.rb` (export action)
- `app/controllers/discussions_controller.rb` (export action)

### 8.9 Integrations Domain

**Purpose:** Webhooks, chatbots, external services

**Backend Files:**
- `app/models/chatbot.rb`
- `app/models/webhook.rb` (table exists)
- `app/models/received_email.rb`
- `app/services/chatbot_service.rb`
- `app/services/received_email_service.rb`
- `app/controllers/api/v1/webhooks_controller.rb`
- `app/controllers/api/v1/chatbots_controller.rb`
- `app/controllers/api/v1/received_emails_controller.rb`
- `app/controllers/api/b2/*.rb` (bot API)
- `app/controllers/api/b3/*.rb` (bot API)
- `app/controllers/received_emails_controller.rb`
- `app/models/ability/chatbot.rb`
- `app/extras/clients/*.rb`
- `app/serializers/chatbot_serializer.rb`
- `app/serializers/webhook_serializer.rb`
- `app/serializers/received_email_serializer.rb`
- `config/webhook_event_kinds.yml`

**Frontend Files:**
- `vue/src/components/chatbot/`
- `vue/src/shared/services/chatbot_service.js`
- `vue/src/shared/interfaces/chatbot_records_interface.js`
- `vue/src/shared/interfaces/webhook_records_interface.js`
- `vue/src/shared/interfaces/received_email_records_interface.js`

### 8.10 Templates Domain

**Purpose:** Reusable discussion and poll templates

**Backend Files:**
- `app/models/discussion_template.rb`
- `app/models/poll_template.rb`
- `app/services/discussion_template_service.rb`
- `app/services/poll_template_service.rb`
- `app/controllers/api/v1/discussion_templates_controller.rb`
- `app/controllers/api/v1/poll_templates_controller.rb`
- `app/controllers/poll_templates_controller.rb`
- `app/controllers/thread_templates_controller.rb`
- `app/models/ability/discussion_template.rb`
- `app/models/ability/poll_template.rb`
- `app/serializers/discussion_template_serializer.rb`
- `app/serializers/poll_template_serializer.rb`
- `app/workers/convert_discussion_templates_worker.rb`
- `app/workers/migrate_poll_templates_worker.rb`
- `config/poll_templates.yml`
- `config/discussion_templates.yml`

**Frontend Files:**
- `vue/src/components/poll_template/`
- `vue/src/components/thread_template/`
- `vue/src/shared/services/discussion_template_service.js`
- `vue/src/shared/services/poll_template_service.js`
- `vue/src/shared/interfaces/discussion_template_records_interface.js`
- `vue/src/shared/interfaces/poll_template_records_interface.js`

### Cross-Cutting Concerns

**Tags:**
- `app/models/tag.rb`, `app/models/tagging.rb`
- `app/services/tag_service.rb`
- `vue/src/components/tags/`

**Tasks:**
- `app/models/task.rb`, `app/models/tasks_user.rb`
- `app/services/task_service.rb`
- `vue/src/components/tasks/`

**Reactions:**
- `app/models/reaction.rb`
- `app/services/reaction_service.rb`
- `vue/src/components/reaction/`

**Translations:**
- `app/models/translation.rb`
- `app/services/translation_service.rb`

**Announcements:**
- `app/services/announcement_service.rb`
- `app/controllers/api/v1/announcements_controller.rb`

---

## 9. Complexity Indicators

### 9.1 Large Model Files (>200 lines)

| File | Lines | Notes |
|------|-------|-------|
| `app/models/poll.rb` | 546 | Most complex model, many poll types and states |
| `app/models/group.rb` | 476 | Complex permissions and hierarchy |
| `app/models/user.rb` | 377 | Many user-related features |
| `app/models/stance.rb` | 321 | Voting logic complexity |
| `app/models/discussion.rb` | 287 | Thread management |
| `app/models/permitted_params.rb` | 269 | Parameter whitelisting for all models |
| `app/models/event.rb` | 214 | Event system core |

### 9.2 Large Service Files (>200 lines)

| File | Lines | Notes |
|------|-------|-------|
| `app/services/report_service.rb` | 472 | Analytics and reporting |
| `app/services/poll_service.rb` | 454 | Poll lifecycle management |
| `app/services/record_cloner.rb` | 451 | Record duplication logic |
| `app/services/record_cache.rb` | 359 | Serialization caching |
| `app/services/group_export_service.rb` | 333 | Export generation |
| `app/services/discussion_service.rb` | 280 | Discussion lifecycle |
| `app/services/received_email_service.rb` | 234 | Email processing |
| `app/services/translation_service.rb` | 221 | Translation handling |
| `app/services/membership_service.rb` | 208 | Membership management |

### 9.3 Large Controller Files (>100 lines)

| File | Lines | Notes |
|------|-------|-------|
| `app/controllers/api/v1/snorlax_base.rb` | 280 | Base controller with all CRUD logic |
| `app/controllers/api/v1/announcements_controller.rb` | 188 | Complex announcement logic |
| `app/controllers/api/v1/discussions_controller.rb` | 187 | Many discussion actions |
| `app/controllers/api/v1/profile_controller.rb` | 167 | User profile management |
| `app/controllers/api/v1/stances_controller.rb` | 156 | Voting endpoints |
| `app/controllers/api/v1/discussion_templates_controller.rb` | 142 | Template management |
| `app/controllers/api/v1/search_controller.rb` | 119 | Search logic |
| `app/controllers/api/v1/events_controller.rb` | 117 | Event timeline |
| `app/controllers/api/v1/memberships_controller.rb` | 114 | Membership actions |
| `app/controllers/api/v1/poll_templates_controller.rb` | 105 | Poll template management |

### 9.4 Large Vue Components (>300 lines)

| File | Lines | Notes |
|------|-------|-------|
| `vue/src/components/lmo_textarea/collab_editor.vue` | 831 | Collaborative rich text editor |
| `vue/src/components/poll/common/form.vue` | 590 | Poll creation form |
| `vue/src/components/poll_template/form.vue` | 521 | Poll template form |
| `vue/src/components/common/recipients_autocomplete.vue` | 419 | User/email autocomplete |
| `vue/src/components/group/form.vue` | 401 | Group settings form |
| `vue/src/components/common/icon.vue` | 393 | Icon component (large icon set) |
| `vue/src/components/common/formatted_text.vue` | 353 | Rich text display |
| `vue/src/components/report/page.vue` | 350 | Reporting page |
| `vue/src/components/group/members_panel.vue` | 325 | Member management |

### 9.5 Heavy Concern Usage

Models with many concerns (indicating complexity):

- **Poll** - Uses HasRichText, HasMentions, HasTags, HasEvents, HasCustomFields, HasTimeframe, Searchable, Translatable, ReadableUnguessableUrls
- **Discussion** - Uses HasRichText, HasMentions, HasTags, HasEvents, HasCustomFields, Searchable, Translatable, ReadableUnguessableUrls
- **Group** - Uses HasRichText, GroupPrivacy, SelfReferencing, HasAvatar, HasExperiences, HasEvents, Searchable

### 9.6 STI Hierarchy Depth

The Event class has 42 STI subclasses, all with concern-based behavior customization. This is a significant complexity hotspot.

### 9.7 Deep Nesting in Frontend

Component nesting patterns observed:
- `strand/` components for timeline rendering are deeply nested
- `poll/common/` has complex form components with many sub-components
- `lmo_textarea/` has the collaborative editor with Tiptap extensions

### 9.8 Migration Count

High migration activity suggests significant schema evolution over time.

---

## 10. Open Questions

### Architecture Questions

1. **How does the real-time collaboration work?**
   - How does Hocuspocus integrate with the Rails backend?
   - What is the relationship between the Channels service and Hocuspocus?
   - How is Y.js state synchronized with the database?

2. **What is the full Event lifecycle?**
   - How do EventBus listeners work in detail?
   - What triggers notifications vs emails vs real-time updates?
   - How are events replayed or queried for timelines?

3. **How does the template system work?**
   - How are templates instantiated into discussions/polls?
   - Are templates versioned?
   - How do template changes affect existing content?

4. **How does the subscription/billing system work?**
   - What is the LoomioSubs engine?
   - How does Chargify integration function?
   - What are the subscription tiers and limits?

### Domain Questions

5. **What are all the poll types?**
   - The config files mention various poll types
   - How do they differ in voting mechanics?
   - What is the stance calculation logic for each type?

6. **How does the permission system work in detail?**
   - How do group permissions cascade to discussions and polls?
   - What is the delegate role?
   - How do guest permissions work?

7. **How does the email processing work?**
   - How are inbound emails parsed and routed?
   - What is the email forwarding system?
   - How do email aliases work?

8. **What is the demo/trial system?**
   - How are demo groups created?
   - What is the trial flow?
   - How are trials converted to paid subscriptions?

### Technical Questions

9. **What is the RecordCache system?**
   - How does serialization caching work?
   - What is the cache invalidation strategy?
   - How does it interact with real-time updates?

10. **How is search implemented?**
    - What is indexed for search?
    - How is pg_search configured?
    - Are there any search performance considerations?

11. **What is the SequenceService?**
    - How are sequence IDs generated?
    - What is the partition_sequences table for?
    - How does this relate to event ordering?

12. **How does the frontend record store sync?**
    - When does LokiJS sync with the backend?
    - How are conflicts handled?
    - What is the offline capability?

### Maintenance Questions

13. **What are all the background workers doing?**
    - Some workers seem to be one-time migrations
    - Which workers run regularly?
    - What is the job queue configuration?

14. **What is the deployment architecture?**
    - How is the application deployed?
    - What is the relationship between loomio-deploy repo and this repo?
    - What environment variables are required?

15. **What is the testing coverage?**
    - Which areas have good test coverage?
    - Which areas need more testing?
    - How are E2E tests structured?

---

## Appendix A: File Count Summary

| Category | Count |
|----------|-------|
| Total Ruby files (app/) | 442 |
| Models | 152 |
| Controllers | 89 |
| Services | 44 |
| Serializers | 65 |
| Workers | 38 |
| Mailers | 7 |
| Queries | 8 |
| Vue Components | 217 |
| Vue Services/Interfaces | 117 |
| RSpec Tests | 116 |
| E2E Tests | 14 |
| Database Tables | 56 |
| Locale Files | 49 |
| Event Types | 42 |
| Model Concerns | 36 |
| Ability Classes | 23 |
| Record Interfaces | 28 |

---

## Appendix B: Key Configuration Files

| File | Purpose |
|------|---------|
| `config/routes.rb` | All application routes |
| `config/providers.yml` | OAuth provider configuration |
| `config/poll_types.yml` | Poll type definitions |
| `config/poll_templates.yml` | Default poll templates |
| `config/discussion_templates.yml` | Default discussion templates |
| `config/webhook_event_kinds.yml` | Webhook event types |
| `config/doctypes.yml` | Document type definitions |
| `config/colors.yml` | Color palette |
| `config/emojis.yml` | Emoji definitions |
| `config/locales/*.yml` | Translation files |
| `vue/src/routes.js` | Frontend routing |
| `vue/vite.config.js` | Frontend build configuration |

---

## Appendix C: Database Schema Version

Schema version: `2025_12_03_031449`

This indicates the database schema was last modified on December 3, 2025.

---

*End of Broad Overview*
