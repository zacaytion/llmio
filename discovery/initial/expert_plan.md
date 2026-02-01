# Expert Plan: Domain Investigation Tasks

**Generated:** 2026-02-01
**Purpose:** Task breakdown for swarm agents investigating Loomio's 10 domains

---

## Domain: auth

### Priority
**High** - Foundation for all user interactions; other domains depend on knowing who the user is and how they authenticate.

### Key Questions to Answer
1. How does the sign-in flow work from frontend to session creation?
2. What authentication strategies are supported (password, OAuth, SAML, magic link)?
3. How are login tokens generated, validated, and expired?
4. How does OAuth identity linking work for existing users?
5. How is email verification handled for new users?
6. What password policies are enforced (length, pwned password check)?
7. How does "remember me" functionality work?
8. How are user sessions invalidated on password change?
9. What is the LoggedOutUser object and when is it used?
10. How does anonymous/restricted user access work?

### Files to Investigate
- `/app/models/user.rb` - User model with Devise modules
- `/app/models/login_token.rb` - Magic link token model
- `/app/models/logged_out_user.rb` - Null object for unauthenticated users
- `/app/models/anonymous_user.rb` - For truly anonymous contexts
- `/app/models/identity.rb` - OAuth identity storage
- `/app/services/user_service.rb` - User operations including verification
- `/app/services/login_token_service.rb` - Magic link token operations
- `/app/controllers/api/v1/sessions_controller.rb` - Login API
- `/app/controllers/api/v1/registrations_controller.rb` - Signup API
- `/app/controllers/login_tokens_controller.rb` - Magic link handling
- `/app/controllers/identities/base_controller.rb` - OAuth base
- `/app/controllers/identities/google_controller.rb` - Google OAuth
- `/app/controllers/identities/saml_controller.rb` - SAML SSO
- `/app/helpers/current_user_helper.rb` - Current user resolution
- `/config/initializers/devise.rb` - Devise configuration
- `/config/providers.yml` - OAuth provider configuration

### Expected Outputs
- Document all authentication methods with their flows
- Map the session lifecycle (creation, validation, expiration)
- Document OAuth provider integration patterns
- Document magic link token lifecycle
- Document user verification requirements

### Dependencies on Other Domains
- Depends on: None (foundation domain)
- Depended on by: All other domains (user identity is required everywhere)

---

## Domain: groups

### Priority
**High** - Groups are the organizational container for all collaborative features. Membership and permissions affect all other domains.

### Key Questions to Answer
1. What is the Group hierarchy structure (parent/subgroups)?
2. How do FormalGroup, GuestGroup, and NullGroup differ?
3. How are memberships created, accepted, and revoked?
4. What are all the group permission settings and how do they cascade?
5. How does invitation redemption work for new vs existing users?
6. What is the delegate role and how does it differ from admin?
7. How are membership requests handled (request to join flow)?
8. How does group privacy work (public, secret, closed)?
9. How are subgroup permissions inherited from parent?
10. How does subscription/billing integrate with groups?

### Files to Investigate
- `/app/models/group.rb` - Group model with all settings
- `/app/models/formal_group.rb` - STI subclass for formal groups
- `/app/models/guest_group.rb` - STI subclass for guest contexts
- `/app/models/membership.rb` - Membership model
- `/app/models/membership_request.rb` - Join request model
- `/app/models/subscription.rb` - Billing subscription
- `/app/models/ability/group.rb` - Group permissions
- `/app/models/ability/membership.rb` - Membership permissions
- `/app/models/ability/membership_request.rb` - Request permissions
- `/app/models/concerns/group_privacy.rb` - Privacy concern
- `/app/models/concerns/self_referencing.rb` - Parent/child hierarchy
- `/app/services/group_service.rb` - Group operations
- `/app/services/membership_service.rb` - Membership operations
- `/app/services/membership_request_service.rb` - Request operations
- `/app/queries/group_query.rb` - Group visibility queries
- `/app/queries/membership_query.rb` - Membership queries
- `/app/controllers/api/v1/groups_controller.rb` - Group API
- `/app/controllers/api/v1/memberships_controller.rb` - Membership API
- `/app/controllers/api/v1/membership_requests_controller.rb` - Request API

### Expected Outputs
- Document complete Group model structure and settings
- Document membership states (pending, accepted, revoked)
- Map invitation flow from send to acceptance
- Document all permission settings and their effects
- Document subscription integration points

### Dependencies on Other Domains
- Depends on: auth (users must exist)
- Depended on by: discussions, polls, events, documents, integrations, templates

---

## Domain: discussions

### Priority
**Medium** - Primary collaboration feature, but builds on groups and auth.

### Key Questions to Answer
1. How is a discussion created and associated with a group?
2. How do DiscussionReader records track read state?
3. What is the comment threading model (parent_id, position)?
4. How does max_depth affect reply nesting?
5. How are discussions moved between groups?
6. How does discussion forking work?
7. What is the volume system for notification preferences?
8. How does pinning work?
9. How are discussion guests (non-members) handled?
10. How does the closed/reopened state work?

### Files to Investigate
- `/app/models/discussion.rb` - Discussion model
- `/app/models/discussion_reader.rb` - Read state tracking
- `/app/models/comment.rb` - Comment model
- `/app/models/null_discussion.rb` - Null object pattern
- `/app/models/ability/discussion.rb` - Discussion permissions
- `/app/models/ability/discussion_reader.rb` - Reader permissions
- `/app/models/ability/comment.rb` - Comment permissions
- `/app/models/concerns/has_rich_text.rb` - Rich text handling
- `/app/models/concerns/has_mentions.rb` - @mention parsing
- `/app/services/discussion_service.rb` - Discussion operations
- `/app/services/discussion_reader_service.rb` - Read state operations
- `/app/services/comment_service.rb` - Comment operations
- `/app/queries/discussion_query.rb` - Discussion visibility
- `/app/controllers/api/v1/discussions_controller.rb` - Discussion API
- `/app/controllers/api/v1/comments_controller.rb` - Comment API
- `/app/controllers/api/v1/discussion_readers_controller.rb` - Reader API

### Expected Outputs
- Document discussion lifecycle (create, edit, close, move, fork)
- Document comment threading model and sequence handling
- Document read state tracking via DiscussionReader
- Map guest access patterns for discussions
- Document volume preferences and notification behavior

### Dependencies on Other Domains
- Depends on: auth, groups
- Depended on by: polls (polls can be in discussions), events, search

---

## Domain: polls

### Priority
**Medium** - Core decision-making feature, complex voting mechanics.

### Key Questions to Answer
1. What are all the poll types and how do they differ in voting mechanics?
2. How does the Stance model represent votes?
3. How do StanceChoices work for multi-choice polls?
4. What is the poll lifecycle (active, closing_soon, closed, reopened)?
5. How does anonymous voting work?
6. How is hide_results (until voted, until closed) implemented?
7. How are poll options scored and results calculated?
8. How do outcomes work after a poll closes?
9. How does specified_voters_only vs open voting work?
10. How are reminders sent for polls?

### Files to Investigate
- `/app/models/poll.rb` - Poll model (largest model file)
- `/app/models/poll_option.rb` - Poll option model
- `/app/models/stance.rb` - Vote/response model
- `/app/models/stance_choice.rb` - Individual choice in vote
- `/app/models/outcome.rb` - Poll outcome model
- `/app/models/null_poll.rb` - Null object pattern
- `/app/models/ability/poll.rb` - Poll permissions
- `/app/models/ability/stance.rb` - Voting permissions
- `/app/models/ability/outcome.rb` - Outcome permissions
- `/app/services/poll_service.rb` - Poll operations (largest service)
- `/app/services/stance_service.rb` - Voting operations
- `/app/services/outcome_service.rb` - Outcome operations
- `/app/queries/poll_query.rb` - Poll visibility
- `/app/controllers/api/v1/polls_controller.rb` - Poll API
- `/app/controllers/api/v1/stances_controller.rb` - Voting API
- `/app/controllers/api/v1/outcomes_controller.rb` - Outcome API
- `/config/poll_types.yml` - Poll type configuration

### Expected Outputs
- Document all poll types and their voting mechanics
- Document stance creation and update flow
- Document result calculation for each poll type
- Map poll lifecycle states and transitions
- Document anonymous voting and result hiding

### Dependencies on Other Domains
- Depends on: auth, groups, discussions (optional)
- Depended on by: events, search, templates

---

## Domain: events

### Priority
**High** - Central to activity tracking and notifications across all features.

### Key Questions to Answer
1. How are events created and what triggers publishing?
2. How does the STI system work for event types?
3. How do event concerns compose behavior (InApp, ByEmail, LiveUpdate)?
4. What is the trigger chain when an event is published?
5. How are event recipients determined?
6. How do sequence_id and position_key work for threading?
7. How does the EventBus work for side effects?
8. How are events used for timeline display?
9. How are events pinned in discussions?
10. How are parent events and child counts managed?

### Files to Investigate
- `/app/models/event.rb` - Base event model
- `/app/models/events/*.rb` - All 42 event subclasses
- `/app/models/notification.rb` - Notification model
- `/app/models/ability/event.rb` - Event permissions
- `/app/models/concerns/events/live_update.rb` - Real-time concern
- `/app/models/concerns/events/notify/in_app.rb` - In-app notifications
- `/app/models/concerns/events/notify/by_email.rb` - Email notifications
- `/app/models/concerns/events/notify/mentions.rb` - Mention handling
- `/app/models/concerns/events/notify/chatbots.rb` - Chatbot notifications
- `/app/models/concerns/events/notify/subscribers.rb` - Volume-based notifications
- `/app/services/event_service.rb` - Event operations
- `/app/services/notification_service.rb` - Notification operations
- `/app/services/sequence_service.rb` - Sequence generation
- `/app/workers/publish_event_worker.rb` - Event publishing worker
- `/lib/event_bus.rb` - Pub/sub system
- `/config/initializers/event_bus.rb` - EventBus listeners
- `/app/controllers/api/v1/events_controller.rb` - Event API
- `/app/controllers/api/v1/notifications_controller.rb` - Notification API
- `/app/extras/queries/users_by_volume_query.rb` - Recipient filtering

### Expected Outputs
- Document complete event publishing flow
- Catalog all 42 event types and their behaviors
- Document notification recipient determination
- Map the trigger chain for each notification type
- Document event threading and sequencing

### Dependencies on Other Domains
- Depends on: auth (for recipients), all other domains (events reference all models)
- Depended on by: All domains (all features generate events)

---

## Domain: documents

### Priority
**Low** - Supporting feature for file attachments.

### Key Questions to Answer
1. How are files attached via Active Storage?
2. What is the Document model vs has_many_attached?
3. How are file uploads handled (direct upload)?
4. What file types are supported?
5. How are images processed/previewed?
6. What is the storage backend configuration?
7. How are documents associated with groups/discussions/polls?
8. How is document authorization handled?
9. How are attachments serialized for the frontend?
10. How are orphaned attachments cleaned up?

### Files to Investigate
- `/app/models/document.rb` - Document model
- `/app/models/attachment.rb` - Attachment wrapper
- `/app/models/concerns/has_rich_text.rb` - File attachment handling
- `/app/models/ability/document.rb` - Document permissions
- `/app/models/ability/attachment.rb` - Attachment permissions
- `/app/services/document_service.rb` - Document operations
- `/app/controllers/api/v1/documents_controller.rb` - Document API
- `/app/controllers/api/v1/attachments_controller.rb` - Attachment API
- `/app/controllers/direct_uploads_controller.rb` - Direct upload handling
- `/app/workers/download_attachment_worker.rb` - Attachment processing
- `/app/workers/attach_document_worker.rb` - Document attachment
- `/app/serializers/document_serializer.rb` - Document serialization
- `/app/serializers/attachment_serializer.rb` - Attachment serialization
- `/config/storage.yml` - Storage configuration
- `/config/doctypes.yml` - Document type icons

### Expected Outputs
- Document file upload flow (direct vs server-side)
- Document Active Storage configuration
- Map document associations to other models
- Document file processing and preview generation

### Dependencies on Other Domains
- Depends on: auth, groups (for authorization context)
- Depended on by: discussions, polls (they have attachments)

---

## Domain: search

### Priority
**Low** - Supporting feature, but important for discoverability.

### Key Questions to Answer
1. How is pg_search configured for full-text search?
2. What models are searchable and what fields are indexed?
3. How is the pg_search_documents table maintained?
4. How does search visibility work (can only search what you can see)?
5. How are search results ranked?
6. How does the SearchResult model work?
7. When is the search index rebuilt?
8. How does content locale affect search?
9. What is the search API contract?
10. Are there any search performance considerations?

### Files to Investigate
- `/app/models/search_result.rb` - Search result model
- `/app/models/concerns/searchable.rb` - Searchable concern
- `/app/services/search_service.rb` - Search operations
- `/app/controllers/api/v1/search_controller.rb` - Search API
- `/app/serializers/search_result_serializer.rb` - Result serialization
- `/config/initializers/pg_search.rb` - pg_search configuration
- Check models that include Searchable (Discussion, Poll, Group, Comment)

### Expected Outputs
- Document search indexing configuration
- Document search query API
- Document visibility filtering in search
- Map which models are searchable and what fields

### Dependencies on Other Domains
- Depends on: auth (for visibility), groups, discussions, polls (searchable content)
- Depended on by: None directly

---

## Domain: export

### Priority
**Low** - Supporting feature for data portability.

### Key Questions to Answer
1. What export formats are supported (JSON, CSV)?
2. What data is included in a group export?
3. How are large exports handled (background jobs)?
4. How are export files delivered to users?
5. What are the export permissions?
6. How is PII handled in exports?
7. How does discussion export differ from group export?
8. How does poll export work?
9. What is the export file structure?
10. Are there size limits on exports?

### Files to Investigate
- `/app/services/group_export_service.rb` - Export logic
- `/app/extras/group_exporter.rb` - Export generation
- `/app/extras/poll_exporter.rb` - Poll export
- `/app/models/concerns/group_export_relations.rb` - Export associations
- `/app/models/concerns/discussion_export_relations.rb` - Discussion export
- `/app/workers/group_export_worker.rb` - JSON export worker
- `/app/workers/group_export_csv_worker.rb` - CSV export worker
- `/app/controllers/groups_controller.rb` - Group export endpoint
- `/app/controllers/polls_controller.rb` - Poll export endpoint
- `/app/controllers/discussions_controller.rb` - Discussion export endpoint

### Expected Outputs
- Document export request flow
- Document export file formats and structure
- Map which data is included in each export type
- Document export delivery mechanism

### Dependencies on Other Domains
- Depends on: groups, discussions, polls (content being exported)
- Depended on by: None

---

## Domain: integrations

### Priority
**Medium** - Important for connecting Loomio to external systems.

### Key Questions to Answer
1. How do webhooks work for outbound notifications?
2. How do chatbots receive notifications?
3. What events can trigger webhook calls?
4. How is the bot API (b2/b3) structured?
5. How are inbound emails processed?
6. How does email forwarding work?
7. What chatbot platforms are supported?
8. How are integration credentials stored?
9. How are OAuth applications (doorkeeper) configured?
10. What is the webhook payload format?

### Files to Investigate
- `/app/models/chatbot.rb` - Chatbot model
- `/app/models/webhook.rb` - Webhook model
- `/app/models/received_email.rb` - Inbound email model
- `/app/models/ability/chatbot.rb` - Chatbot permissions
- `/app/services/chatbot_service.rb` - Chatbot operations
- `/app/services/received_email_service.rb` - Email processing
- `/app/controllers/api/v1/webhooks_controller.rb` - Webhook management
- `/app/controllers/api/v1/chatbots_controller.rb` - Chatbot management
- `/app/controllers/api/v1/received_emails_controller.rb` - Email management
- `/app/controllers/api/b2/*.rb` - Bot API v2 controllers
- `/app/controllers/api/b3/*.rb` - Bot API v3 controllers
- `/app/controllers/received_emails_controller.rb` - Inbound email endpoint
- `/app/models/concerns/events/notify/chatbots.rb` - Chatbot notification
- `/app/extras/clients/*.rb` - HTTP clients for integrations
- `/config/webhook_event_kinds.yml` - Webhook event types

### Expected Outputs
- Document webhook configuration and payload format
- Document chatbot setup for different platforms
- Document inbound email processing flow
- Document bot API authentication and endpoints

### Dependencies on Other Domains
- Depends on: auth, groups, events (content being sent)
- Depended on by: None directly

---

## Domain: templates

### Priority
**Low** - Supporting feature for repeatability.

### Key Questions to Answer
1. How are discussion templates structured?
2. How are poll templates structured?
3. How is a template instantiated into a real discussion/poll?
4. Can templates be shared across groups?
5. How are template positions/ordering managed?
6. How do hidden/discarded templates work?
7. What fields are templatable vs fixed?
8. How do templates relate to poll types?
9. Can templates include attached poll templates?
10. How are default system templates managed?

### Files to Investigate
- `/app/models/discussion_template.rb` - Discussion template model
- `/app/models/poll_template.rb` - Poll template model
- `/app/models/ability/discussion_template.rb` - Template permissions
- `/app/models/ability/poll_template.rb` - Template permissions
- `/app/services/discussion_template_service.rb` - Template operations
- `/app/services/poll_template_service.rb` - Template operations
- `/app/controllers/api/v1/discussion_templates_controller.rb` - Template API
- `/app/controllers/api/v1/poll_templates_controller.rb` - Template API
- `/app/controllers/poll_templates_controller.rb` - Public template page
- `/app/controllers/thread_templates_controller.rb` - Public template page
- `/app/serializers/discussion_template_serializer.rb` - Template serialization
- `/app/serializers/poll_template_serializer.rb` - Template serialization
- `/config/poll_templates.yml` - Default poll templates
- `/config/discussion_templates.yml` - Default discussion templates

### Expected Outputs
- Document template model structure
- Document template instantiation flow
- Map template sharing and visibility rules
- Document default template configuration

### Dependencies on Other Domains
- Depends on: auth, groups, discussions, polls (what templates create)
- Depended on by: None

---

## Investigation Order Recommendation

Based on dependencies and priority:

1. **auth** (High, no dependencies) - Start here to understand user identity
2. **groups** (High, depends on auth) - Foundation for all content
3. **events** (High, depends on auth) - Central to understanding data flow
4. **discussions** (Medium, depends on groups) - Primary collaboration feature
5. **polls** (Medium, depends on groups/discussions) - Decision-making feature
6. **integrations** (Medium, depends on events) - External connectivity
7. **search** (Low, depends on multiple) - Discoverability
8. **documents** (Low, depends on groups) - File attachments
9. **export** (Low, depends on multiple) - Data portability
10. **templates** (Low, depends on multiple) - Process repeatability

---

## General Investigation Guidelines

For each domain:

1. **Read the model first** - Understand the data structure, associations, and concerns
2. **Read the service** - Understand the business operations
3. **Read the ability** - Understand authorization rules
4. **Read the query object** - Understand visibility scoping
5. **Read the controller** - Understand API endpoints
6. **Read the serializer** - Understand API output
7. **Check specs** - Tests often document edge cases
8. **Check events** - Understand what activities are tracked
9. **Check workers** - Understand background processing

Document findings in:
- Plain English descriptions (no code snippets)
- File paths for reference
- Data flow diagrams where helpful
- Edge cases and gotchas
