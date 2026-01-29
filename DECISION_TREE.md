# Decision Tree: Key Choices for Loomio Go Rewrite

This document provides decision trees for the most critical choices you'll face during the planning phase.

---

## Decision 1: Migration Strategy

```
START: How should we approach the rewrite?
│
├─ Do you need to maintain the Rails app during rewrite?
│  │
│  ├─ NO → Consider: BIG BANG REWRITE
│  │      ├─ Pros: Clean slate, no dual maintenance, simpler
│  │      ├─ Cons: High risk, 12+ months to production, no revenue during rewrite
│  │      └─ Best for: Small teams, low user base, or new product version
│  │
│  └─ YES → Do you have complex infrastructure capacity?
│     │
│     ├─ YES → Consider: STRANGLER FIG PATTERN
│     │        ├─ Pros: Gradual migration, reduced risk, continuous delivery
│     │        ├─ Cons: Complex routing, dual maintenance, longer overall timeline
│     │        └─ Best for: Large apps, high traffic, zero-downtime requirement
│     │
│     └─ NO → Consider: HYBRID APPROACH
│              ├─ Pros: Balance of speed and risk, manageable complexity
│              ├─ Cons: Requires careful boundary definition, some dual maintenance
│              └─ Best for: Medium complexity, moderate traffic, practical teams
│
└─ RECOMMENDATION: For Loomio (medium app, active users, moderate traffic)
   → Start with HYBRID APPROACH
   → Rewrite backend modules independently
   → Keep Rails running alongside Go during transition
   → Use feature flags for gradual cutover
```

---

## Decision 2: Web Framework Selection

```
START: Which Go web framework should we use?
│
├─ What's your team's Go experience level?
│  │
│  ├─ BEGINNER → Do you want familiar patterns?
│  │              │
│  │              ├─ YES (Express.js-like) → GIN or FIBER
│  │              │    └─ Choose GIN (more mature, larger community)
│  │              │
│  │              └─ NO (learn idiomatic Go) → CHI or STDLIB
│  │                   └─ Choose CHI (good balance of features + idioms)
│  │
│  ├─ INTERMEDIATE → What's your priority?
│  │                 │
│  │                 ├─ SPEED OF DEVELOPMENT → GIN
│  │                 │    ├─ Rich middleware ecosystem
│  │                 │    └─ Built-in validation
│  │                 │
│  │                 ├─ PERFORMANCE → FIBER or GIN
│  │                 │    └─ Both are fast, choose GIN for maturity
│  │                 │
│  │                 └─ IDIOMATIC GO → CHI or ECHO
│  │                      └─ Choose CHI (more idiomatic)
│  │
│  └─ ADVANCED → What's your philosophy?
│                │
│                ├─ MINIMAL DEPENDENCIES → STDLIB (net/http)
│                │    ├─ Most control, no magic
│                │    └─ More boilerplate
│                │
│                └─ PRAGMATIC → CHI or GIN
│                     └─ Choose CHI for clean code, GIN for features
│
└─ RECOMMENDATION FOR LOOMIO:
   → GIN (github.com/gin-gonic/gin)
   → Reasons: Active community, good docs, fast development, proven at scale
```

---

## Decision 3: ORM/Database Layer

```
START: How should we handle database access?
│
├─ What's your priority?
│  │
│  ├─ TYPE SAFETY & COMPILE-TIME CHECKS
│  │  │
│  │  └─ Choose: SQLC
│  │       ├─ Generates Go code from SQL
│  │       ├─ Full type safety
│  │       ├─ No runtime overhead
│  │       └─ Best for: Teams comfortable writing SQL
│  │
│  ├─ RAPID DEVELOPMENT & FAMILIAR ORM PATTERNS
│  │  │
│  │  └─ Choose: GORM
│  │       ├─ ActiveRecord-like API (familiar for Rails devs)
│  │       ├─ Auto-migrations
│  │       ├─ Rich associations
│  │       └─ Best for: Quick prototyping, Rails-like patterns
│  │
│  ├─ FLEXIBILITY & CONTROL
│  │  │
│  │  └─ Choose: SQLX
│  │       ├─ Minimal abstraction over database/sql
│  │       ├─ Write your own SQL
│  │       ├─ Struct scanning helpers
│  │       └─ Best for: Complex queries, performance optimization
│  │
│  └─ ADVANCED FEATURES & SCHEMA-AS-CODE
│     │
│     └─ Choose: ENT
│          ├─ Graph-based schema definition
│          ├─ Code generation
│          ├─ Advanced querying
│          └─ Best for: Complex data models, graph relationships
│
└─ RECOMMENDATION FOR LOOMIO:
   → Start with GORM (familiar patterns, quick migration)
   → Consider SQLC for performance-critical paths
   → Migrate to SQLC gradually if needed
```

---

## Decision 4: Background Job Processing

```
START: How should we handle async/background jobs?
│
├─ What's your infrastructure?
│  │
│  ├─ ALREADY USING REDIS
│  │  │
│  │  └─ Choose: ASYNQ
│  │       ├─ Redis-backed
│  │       ├─ Web UI for monitoring
│  │       ├─ Retry logic built-in
│  │       ├─ Similar to Sidekiq
│  │       └─ Recommended: YES for Loomio
│  │
│  ├─ PREFER POSTGRESQL (single DB)
│  │  │
│  │  └─ Choose: RIVER
│  │       ├─ PostgreSQL-backed
│  │       ├─ No additional infrastructure
│  │       ├─ ACID guarantees
│  │       └─ Best for: Simplicity, single database
│  │
│  └─ NEED DISTRIBUTED WORKFLOWS
│     │
│     └─ Choose: TEMPORAL or MACHINERY
│          ├─ Complex workflow orchestration
│          ├─ Multi-step processes
│          └─ Best for: Enterprise needs
│
└─ RECOMMENDATION FOR LOOMIO:
   → ASYNQ (github.com/hibiken/asynq)
   → Reasons: Redis likely already in stack, Sidekiq-like, proven reliability
```

---

## Decision 5: Authentication Strategy

```
START: How should we handle authentication?
│
├─ What auth methods does Loomio currently use?
│  │
│  ├─ SESSION-BASED (Rails cookies)
│  │  │
│  │  └─ Do you want to maintain sessions?
│  │     │
│  │     ├─ YES → Use session middleware + secure cookies
│  │     │        └─ Library: gorilla/sessions
│  │     │
│  │     └─ NO → Migrate to TOKEN-BASED
│  │              └─ See below ↓
│  │
│  └─ TOKEN-BASED (API clients)
│     │
│     └─ Choose: JWT (JSON Web Tokens)
│          ├─ Library: golang-jwt/jwt
│          ├─ Stateless authentication
│          ├─ Works across services
│          └─ Mobile/SPA friendly
│
├─ Does Loomio support OAuth?
│  │
│  └─ YES → Implement OAuth2
│           ├─ Library: golang.org/x/oauth2
│           ├─ Support Google, GitHub, etc.
│           └─ OIDC for user info
│
└─ RECOMMENDATION FOR LOOMIO:
   → JWT for API authentication
   → OAuth2 for social login
   → bcrypt for password hashing (golang.org/x/crypto/bcrypt)
   → Consider: Maintain backward compatibility during transition
```

---

## Decision 6: WebSocket/Real-time Strategy

```
START: How should we handle real-time features?
│
├─ What's your scale requirement?
│  │
│  ├─ < 1,000 CONCURRENT USERS
│  │  │
│  │  └─ Choose: BUILT-IN WEBSOCKET
│  │       ├─ Library: nhooyr.io/websocket (recommended)
│  │       ├─ Or: gorilla/websocket (more popular)
│  │       ├─ Simple, sufficient for most cases
│  │       └─ Implement in application
│  │
│  ├─ 1,000 - 10,000 CONCURRENT USERS
│  │  │
│  │  └─ Choose: BUILT-IN + REDIS PUB/SUB
│  │       ├─ WebSocket library + Redis for distribution
│  │       ├─ Scale horizontally
│  │       └─ Implement connection pooling
│  │
│  └─ > 10,000 CONCURRENT USERS
│     │
│     └─ Choose: CENTRIFUGO or CUSTOM SOLUTION
│          ├─ Dedicated real-time messaging server
│          ├─ Handles scaling complexity
│          └─ Production-grade features
│
└─ RECOMMENDATION FOR LOOMIO:
   → nhooyr.io/websocket (better API than gorilla)
   → Redis Pub/Sub for multi-instance
   → Connection management middleware
   → Graceful reconnection handling
```

---

## Decision 7: API Design

```
START: What API style should we use?
│
├─ What does the Vue.js frontend expect?
│  │
│  ├─ RESTFUL JSON API
│  │  │
│  │  └─ Stick with REST
│  │       ├─ Minimal frontend changes
│  │       ├─ Easier migration
│  │       └─ Document with OpenAPI/Swagger
│  │
│  └─ COULD BE CHANGED
│     │
│     └─ Consider GraphQL?
│        │
│        ├─ Benefits:
│        │  ├─ Flexible queries
│        │  ├─ Single endpoint
│        │  └─ Strong typing
│        │
│        └─ Drawbacks:
│           ├─ Frontend rewrite needed
│           ├─ Caching complexity
│           └─ Learning curve
│
├─ DECISION CRITERIA:
│  │
│  ├─ Minimize frontend changes? → REST
│  ├─ Complex nested data? → GraphQL
│  ├─ Mobile app with varying needs? → GraphQL
│  └─ Simple, predictable patterns? → REST
│
└─ RECOMMENDATION FOR LOOMIO:
   → Stick with RESTful JSON API
   → Version it: /api/v1/, /api/v2/
   → Use JSON:API or custom standard
   → Add GraphQL later if needed
   → Document with OpenAPI spec
```

---

## Decision 8: Testing Strategy

```
START: How should we approach testing?
│
├─ What types of tests do we need?
│  │
│  ├─ UNIT TESTS
│  │  │
│  │  └─ Framework choice:
│  │     ├─ TESTIFY/ASSERT (most popular)
│  │     │  └─ github.com/stretchr/testify
│  │     │
│  │     └─ GINKGO/GOMEGA (BDD style)
│  │        └─ More verbose, RSpec-like
│  │
│  ├─ INTEGRATION TESTS
│  │  │
│  │  └─ Strategy:
│  │     ├─ Use testcontainers-go for DB
│  │     ├─ Test with real PostgreSQL
│  │     └─ Isolated test databases
│  │
│  └─ E2E TESTS
│     │
│     └─ Keep existing tests?
│        ├─ YES → Run against Go backend
│        └─ NO → Rewrite with Playwright/Cypress
│
├─ What's the coverage goal?
│  │
│  ├─ Minimum: 70% for new code
│  ├─ Target: 80% overall
│  └─ Critical paths: 100%
│
└─ RECOMMENDATION FOR LOOMIO:
   → testify/assert for unit tests
   → testcontainers for integration
   → Keep E2E tests, point at Go
   → golangci-lint for static analysis
   → Continuous coverage tracking
```

---

## Decision 9: Deployment Strategy

```
START: How should we deploy the Go application?
│
├─ What's your current infrastructure?
│  │
│  ├─ HEROKU / PLATFORM-AS-A-SERVICE
│  │  │
│  │  └─ Can you containerize?
│  │     │
│  │     ├─ YES → Docker + Heroku
│  │     │        └─ Dockerfile provided
│  │     │
│  │     └─ NO → Buildpack
│  │              └─ Use Heroku Go buildpack
│  │
│  ├─ KUBERNETES
│  │  │
│  │  └─ Deployment strategy:
│  │     ├─ Blue-Green deployment
│  │     ├─ Canary releases
│  │     └─ Feature flags for gradual rollout
│  │
│  ├─ VPS / BARE METAL
│  │  │
│  │  └─ Process management:
│  │     ├─ systemd service
│  │     ├─ Docker Compose
│  │     └─ Supervisor
│  │
│  └─ CLOUD (AWS/GCP/Azure)
│     │
│     └─ Choose:
│        ├─ Containers: ECS/Cloud Run/Container Apps
│        ├─ Functions: Lambda/Cloud Functions (if microservices)
│        └─ VMs: EC2/Compute Engine (traditional)
│
└─ RECOMMENDATION FOR LOOMIO:
   → Docker containers (platform-agnostic)
   → Multi-stage builds for small images
   → Docker Compose for development
   → K8s or ECS for production
   → Blue-green deployment
   → Feature flags: Use Unleash or LaunchDarkly
```

---

## Decision 10: Data Migration Strategy

```
START: How should we migrate data from Rails to Go?
│
├─ Can you afford downtime?
│  │
│  ├─ YES (Maintenance Window Allowed)
│  │  │
│  │  └─ BIG BANG MIGRATION
│  │     ├─ 1. Freeze Rails app (read-only)
│  │     ├─ 2. Export all data
│  │     ├─ 3. Transform if needed
│  │     ├─ 4. Import to Go app
│  │     ├─ 5. Validate thoroughly
│  │     ├─ 6. Switch DNS/routing
│  │     └─ Rollback plan: Keep Rails backup ready
│  │
│  └─ NO (Zero Downtime Required)
│     │
│     └─ GRADUAL MIGRATION
│        │
│        ├─ Option A: DUAL WRITE
│        │  ├─ Write to both Rails and Go
│        │  ├─ Gradually migrate reads
│        │  └─ Eventually deprecate Rails
│        │
│        └─ Option B: CDC (Change Data Capture)
│           ├─ Use Debezium or similar
│           ├─ Stream changes to Go
│           └─ Eventual consistency
│
├─ Schema changes needed?
│  │
│  ├─ YES → Create migration scripts
│  │        ├─ Data transformation layer
│  │        └─ Validation checksums
│  │
│  └─ NO → Direct copy possible
│           └─ pg_dump → pg_restore
│
└─ RECOMMENDATION FOR LOOMIO:
   → Start with shared database
   → Both apps read/write same PostgreSQL
   → Gradually move endpoints to Go
   → Use feature flags to route traffic
   → Final cutover when 100% migrated
   → Keep Rails as backup for 1 month
```

---

## Quick Decision Matrix

| Decision | Low Complexity | Medium Complexity | High Complexity |
|----------|---------------|-------------------|-----------------|
| **Migration** | Big Bang | Hybrid | Strangler Fig |
| **Framework** | Gin | Gin or Chi | Chi or Stdlib |
| **Database** | GORM | GORM + SQLC | SQLC or Ent |
| **Jobs** | River (PG) | Asynq (Redis) | Temporal |
| **WebSocket** | nhooyr.io/websocket | nhooyr + Redis | Centrifugo |
| **Testing** | testify | testify + testcontainers | Full suite + E2E |
| **Deployment** | Docker Compose | Docker + CI/CD | Kubernetes |

---

## Final Recommendations for Loomio

Based on typical complexity and requirements:

```
✅ Migration Strategy: HYBRID APPROACH
   - Gradual feature migration
   - Shared database initially
   - Feature flags for routing

✅ Web Framework: GIN
   - Fast development
   - Good documentation
   - Large community

✅ Database: GORM
   - Familiar to Rails developers
   - Quick migration
   - Consider SQLC for hot paths

✅ Background Jobs: ASYNQ
   - Redis-backed
   - Sidekiq-like
   - Web UI included

✅ Authentication: JWT + OAuth2
   - Stateless
   - Mobile-friendly
   - Social login support

✅ WebSocket: nhooyr.io/websocket + Redis
   - Clean API
   - Scalable with Redis
   - Good documentation

✅ API Design: RESTful JSON API
   - Minimal frontend changes
   - OpenAPI documentation
   - Versioned endpoints

✅ Testing: testify + testcontainers
   - Popular and mature
   - Real database testing
   - Good CI integration

✅ Deployment: Docker + Kubernetes
   - Platform-agnostic
   - Scalable
   - Standard practices
```

---

## How to Use This Document

1. **Start at the top** of each decision tree
2. **Answer questions honestly** based on your constraints
3. **Follow the branches** to recommendations
4. **Document your choice** in an ADR (Architecture Decision Record)
5. **Review with team** before committing

**Remember:** These are guidelines, not rules. Your specific context may justify different choices.

---

*This decision tree should be revisited after Phase 1 discovery to validate assumptions.*