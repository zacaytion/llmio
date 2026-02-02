# Loomio OpenAPI 3.0 Specification

**Generated:** 2026-02-01
**Phase:** 2 of Delivery Plan
**Status:** ✅ COMPLETE
**Confidence:** HIGH for structure, MEDIUM for edge cases

## Overview

This directory contains the OpenAPI 3.0 specification for Loomio's REST API, extracted from the Rails codebase as part of the blackbox rewrite contract.

## Directory Structure

```
openapi/
├── openapi.yaml                 # Root specification file
├── README.md                    # This file
├── paths/
│   ├── auth.yaml               # Authentication endpoints (6)
│   ├── boot.yaml               # Application bootstrap (3)
│   ├── groups.yaml             # Group management (12)
│   ├── discussions.yaml        # Discussion threads (23)
│   ├── discussion_readers.yaml # Reader management (5)
│   ├── polls.yaml              # Polls and voting (12)
│   ├── stances.yaml            # Vote/stance management (10)
│   ├── comments.yaml           # Comments and reactions (7)
│   ├── events.yaml             # Timeline events (8)
│   ├── memberships.yaml        # Group memberships (18)
│   ├── membership_requests.yaml # Join requests (6)
│   ├── notifications.yaml      # User notifications (2)
│   ├── search.yaml             # Full-text search (1)
│   ├── users.yaml              # User profiles (18)
│   ├── templates.yaml          # Discussion & poll templates (22)
│   ├── tags.yaml               # Tag management (4)
│   ├── documents.yaml          # Document attachments (5)
│   ├── tasks.yaml              # Task tracking (4)
│   ├── announcements.yaml      # Invitation management (7)
│   ├── chatbots.yaml           # Chatbot integrations (5)
│   ├── misc.yaml               # Attachments, mentions, versions, etc. (16)
│   ├── oauth.yaml              # OAuth and SAML flows (7)
│   ├── bot_b2.yaml             # External bot API (6)
│   └── bot_b3.yaml             # Admin bot API (2)
├── components/
│   ├── schemas/
│   │   ├── _index.yaml         # Schema index + additional types
│   │   ├── user.yaml           # User, Author, CurrentUser
│   │   ├── group.yaml          # Group, Subscription, Attachment
│   │   ├── discussion.yaml     # Discussion, DiscussionReader, DiscussionTemplate
│   │   ├── poll.yaml           # Poll, PollOption, PollTemplate, Outcome
│   │   ├── comment.yaml        # Comment, Reaction
│   │   ├── stance.yaml         # Stance, StanceChoice
│   │   ├── event.yaml          # Event (42 types)
│   │   ├── membership.yaml     # Membership, MembershipRequest
│   │   ├── notification.yaml   # Notification
│   │   └── error.yaml          # Error schemas
│   ├── parameters/
│   │   ├── common.yaml         # Common query parameters
│   │   └── pagination.yaml     # Pagination patterns documentation
│   ├── responses/
│   │   └── errors.yaml         # Error response definitions
│   └── securitySchemes/
│       └── auth.yaml           # Authentication methods
└── examples/
    └── (empty - add as needed)
```

## API Namespaces

| Namespace | Base Path | Purpose | Auth Method |
|-----------|-----------|---------|-------------|
| V1 | `/api/v1/` | Primary internal API | Session/Token |
| B2 | `/api/b2/` | External bot integration | User API key |
| B3 | `/api/b3/` | Admin user management | Env API key |

## Key Patterns

### Sideloading

Loomio uses ActiveModelSerializers 0.8 with sideloading. Related records are returned in separate arrays rather than nested:

```json
{
  "discussions": [
    { "id": 1, "author_id": 10 }
  ],
  "users": [
    { "id": 10, "name": "Alice" }
  ]
}
```

Use `exclude_types` to reduce payload: `?exclude_types=poll reaction`

### Pagination

Cursor-based pagination with `from` (cursor) and `per` (limit) parameters. The `from` value is typically the last ID from the previous page, not an offset.

### Batch Fetching

Use `xids` parameter for efficient multi-record fetching: `?xids=1x5x23x42`

## Coverage Summary

### Complete API Coverage: ~204 Endpoints

| Category | File | Endpoints | Status |
|----------|------|-----------|--------|
| Authentication | auth.yaml | 6 | ✅ Complete |
| Boot | boot.yaml | 3 | ✅ Complete |
| Groups | groups.yaml | 12 | ✅ Complete |
| Discussions | discussions.yaml | 23 | ✅ Complete |
| Discussion Readers | discussion_readers.yaml | 5 | ✅ Complete |
| Polls | polls.yaml | 12 | ✅ Complete |
| Stances | stances.yaml | 10 | ✅ Complete |
| Comments | comments.yaml | 7 | ✅ Complete |
| Events | events.yaml | 8 | ✅ Complete |
| Memberships | memberships.yaml | 18 | ✅ Complete |
| Membership Requests | membership_requests.yaml | 6 | ✅ Complete |
| Notifications | notifications.yaml | 2 | ✅ Complete |
| Search | search.yaml | 1 | ✅ Complete |
| Profile/Users | users.yaml | 18 | ✅ Complete |
| Templates | templates.yaml | 22 | ✅ Complete |
| Tags | tags.yaml | 4 | ✅ Complete |
| Documents | documents.yaml | 5 | ✅ Complete |
| Tasks | tasks.yaml | 4 | ✅ Complete |
| Announcements | announcements.yaml | 7 | ✅ Complete |
| Chatbots | chatbots.yaml | 5 | ✅ Complete |
| Misc (Attachments, Mentions, etc.) | misc.yaml | 16 | ✅ Complete |
| OAuth/SAML | oauth.yaml | 7 | ✅ Complete |
| Bot API (B2) | bot_b2.yaml | 6 | ✅ Complete |
| Admin Bot API (B3) | bot_b3.yaml | 2 | ✅ Complete |

### Schema Coverage

| Schema | File | Types |
|--------|------|-------|
| Users | user.yaml | User, Author, CurrentUser |
| Groups | group.yaml | Group, Subscription, Attachment, LinkPreview |
| Discussions | discussion.yaml | Discussion, DiscussionReader, DiscussionTemplate |
| Polls | poll.yaml | Poll, PollOption, PollTemplate, Outcome |
| Comments | comment.yaml | Comment, Reaction |
| Stances | stance.yaml | Stance, StanceChoice |
| Events | event.yaml | Event (42 STI types) |
| Memberships | membership.yaml | Membership, MembershipRequest |
| Notifications | notification.yaml | Notification |
| Errors | error.yaml | Error, ValidationError, ModelError |
| Index | _index.yaml | Tag, Document, Version, Translation, SearchResult, Chatbot, Webhook, Task, ReceivedEmail, Identity |

## Security Concerns Noted

1. **Bot APIs (B2, B3) have NO rate limiting** - Documented in each endpoint
2. **ThrottleService returns 500 instead of 429** - Noted in error docs
3. **OAuth state parameter handling** - See Phase 1 security report

## Validation

This specification has NOT been validated against a running instance. Some edge cases may be inaccurate:

- Query parameter combinations
- Exact error response formats
- Conditional field inclusion
- Permission boundary cases

## Usage

To use this specification:

1. **View**: Open `openapi.yaml` in Swagger Editor or Stoplight
2. **Generate**: Use openapi-generator for client SDKs
3. **Test**: Use Postman/Insomnia to import the spec
4. **Validate**: Compare against actual API responses

### Quick Start

```bash
# View in Swagger Editor (requires Docker)
docker run -p 8080:8080 -e SWAGGER_JSON=/spec/openapi.yaml -v $(pwd):/spec swaggerapi/swagger-editor

# Generate TypeScript client
npx @openapitools/openapi-generator-cli generate -i openapi.yaml -g typescript-fetch -o ./client

# Validate specification
npx @openapitools/openapi-generator-cli validate -i openapi.yaml
```

## Data Sources

This specification was extracted from:

- `config/routes.rb` - Route definitions
- `app/controllers/api/**/*.rb` - Controller actions
- `app/serializers/*.rb` - Response schemas
- `app/models/permitted_params.rb` - Request parameters
- `discovery/schemas/` - Phase 1 extracted schemas

## Confidence Levels

- **HIGH**: Basic CRUD operations, standard patterns
- **MEDIUM**: Complex query parameters, conditional responses
- **LOW**: Edge cases, undocumented behaviors

Items marked with specific confidence levels in the YAML files.

## File Statistics

- **Total YAML files**: 40
- **Path definition files**: 24
- **Schema files**: 11
- **Parameter/Response/Security files**: 5
- **Total documented endpoints**: ~204
- **Total schema types**: ~35

## Next Steps

This specification provides the foundation for:

1. **Phase 3**: Business Logic Specification - Document services, state machines, and business rules
2. **Phase 4**: Event Catalog - Document all event types and their payloads
3. **Phase 5**: Data Model Specification - Entity relationships and constraints

## Changelog

### 2026-02-01 - Initial Release
- Created complete OpenAPI 3.0 specification
- Documented all 204 API endpoints across V1, B2, and B3 namespaces
- Defined 35+ schema types with full property documentation
- Documented security schemes, pagination patterns, and sideloading
- Identified and documented 3 security concerns for rewrite consideration
