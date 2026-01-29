# Glossary and FAQ

Quick reference guide for terms, concepts, and frequently asked questions about the Loomio Go rewrite project.

---

## Glossary of Terms

### Architecture & Patterns

**ADR (Architecture Decision Record)**
A document that captures an important architectural decision made along with its context and consequences.

**Big Bang Rewrite**
A migration strategy where the entire application is rewritten before deployment, with a single cutover event.

**Strangler Fig Pattern**
An incremental migration strategy where new functionality gradually replaces old code, eventually "strangling" the legacy system.

**Vertical Slice**
A complete feature implementation that cuts through all layers of the application (UI, API, business logic, database).

**Horizontal Scaling**
Adding more servers/instances to handle increased load, rather than making individual servers more powerful.

**Blue-Green Deployment**
A deployment strategy where two identical environments (blue and green) are maintained, allowing instant rollback.

**Canary Deployment**
Gradually rolling out changes to a small subset of users before full deployment.

**Technical Debt**
Code that works but needs improvement; shortcuts taken that will require future work.

---

### Go-Specific Terms

**Goroutine**
A lightweight thread managed by the Go runtime; Go's approach to concurrent programming.

**Channel**
A typed conduit through which you can send and receive values between goroutines.

**Interface**
A type that specifies a method set; Go's approach to polymorphism and abstraction.

**Struct**
A composite data type that groups together variables under a single name.

**Pointer**
A variable that stores the memory address of another variable; used for efficiency and mutability.

**Package**
Go's way of organizing and reusing code; similar to modules or libraries in other languages.

**Module**
A collection of related Go packages versioned together (since Go 1.11).

**Context**
A standard way to carry deadlines, cancellation signals, and request-scoped values across API boundaries.

---

### Database Terms

**ORM (Object-Relational Mapping)**
A technique to query and manipulate data using an object-oriented paradigm.

**Migration**
A versioned change to the database schema; allows tracking and applying database changes.

**N+1 Query Problem**
A performance anti-pattern where one query is executed followed by N additional queries in a loop.

**JSONB**
PostgreSQL's binary JSON data type; more efficient than plain JSON.

**Index**
A database structure that improves query performance at the cost of write performance and storage.

**Foreign Key**
A field that references the primary key in another table, enforcing referential integrity.

**CTE (Common Table Expression)**
A temporary named result set (WITH clause in SQL); useful for complex queries.

---

### Rails-Specific Terms

**ActiveRecord**
Rails' ORM framework for database interaction.

**ActionCable**
Rails' framework for handling WebSocket connections and real-time features.

**ActionMailer**
Rails' framework for sending emails.

**Devise**
Popular Rails authentication gem.

**Pundit**
Popular Rails authorization gem.

**Sidekiq**
Popular Rails background job processing library using Redis.

**Rake Task**
Command-line tasks in Rails (similar to Make or npm scripts).

---

### Project Management Terms

**Sprint**
A fixed time period (typically 2 weeks) for completing a set of work items.

**Story Points**
A unit of measure for expressing the effort required to implement a user story.

**Velocity**
The amount of work a team completes in a sprint, measured in story points.

**MVP (Minimum Viable Product)**
The simplest version of a product that can be released to validate assumptions.

**WBS (Work Breakdown Structure)**
A hierarchical decomposition of project work into smaller, manageable components.

**Critical Path**
The sequence of dependent tasks that determines the minimum project duration.

**Burn Down Chart**
A graph showing remaining work over time.

---

### Testing Terms

**Unit Test**
Tests a single unit of code (function, method) in isolation.

**Integration Test**
Tests how multiple units work together.

**End-to-End (E2E) Test**
Tests the entire application flow from user's perspective.

**Test Coverage**
The percentage of code executed during testing.

**Mock**
A simulated object that mimics real object behavior for testing.

**Fixture**
Test data used to populate database for testing.

**Test Double**
Generic term for mock, stub, fake, spy, or dummy objects used in testing.

---

## Frequently Asked Questions

### Planning & Strategy

**Q: How long will this rewrite take?**
A: Based on similar projects, expect 12-18 months for a complete rewrite. This includes:
- 2 months: Planning and discovery
- 8-12 months: Development
- 2-4 months: Testing, beta, and rollout

**Q: Should we rewrite everything at once or incrementally?**
A: For Loomio's complexity, a hybrid approach is recommended. Rewrite backend modules independently while maintaining the Rails app, then gradually cut over using feature flags.

**Q: How much will this cost?**
A: Costs vary based on team size and location. Budget for:
- Development team (2-5 engineers × 12-18 months)
- Infrastructure (staging/production environments)
- Tools and services (CI/CD, monitoring, etc.)
- Buffer for unexpected issues (20-30%)

**Q: What if we discover it's taking too long?**
A: Build in milestone reviews every 2-3 months. If falling behind:
- Reduce scope (cut non-essential features)
- Add resources (carefully - adding people can slow things down initially)
- Extend timeline (be realistic with stakeholders)

**Q: Can we launch with partial feature parity?**
A: Yes, if you prioritize correctly. Launch with core features (80%) used by 80% of users. Add remaining features post-launch.

---

### Technical

**Q: Why Go? Why not [other language]?**
A: Go offers:
- Better performance than Rails
- Excellent concurrency support
- Fast compilation
- Strong standard library
- Easy deployment (single binary)
- Good for web services and APIs

Alternatives like Rust (steeper learning curve) or Node.js (similar to Rails performance) have tradeoffs.

**Q: Will Go be faster than Rails?**
A: Generally yes, but it depends on:
- How well you write Go code
- Where bottlenecks actually are (often database)
- Concurrency patterns used

Expect 2-5x improvement in CPU-bound operations, less improvement in I/O-bound operations.

**Q: How do we handle Rails "magic"?**
A: Go is more explicit. Rails magic must be replaced with:
- Explicit code for associations
- Clear dependency injection
- Explicit error handling
- Manual routing setup

This is actually a benefit - code is more maintainable.

**Q: Can we keep using PostgreSQL?**
A: Absolutely! Go works excellently with PostgreSQL. You can even share the same database during migration.

**Q: What about the Vue.js frontend?**
A: Keep it mostly unchanged. Just update API endpoints to point to Go backend. The frontend doesn't care what language serves the API.

**Q: How do we handle authentication migration?**
A: Options:
1. Keep same session mechanism initially
2. Gradually migrate to JWT tokens
3. Support both during transition
4. Use a token exchange mechanism

**Q: What about background jobs?**
A: Replace Sidekiq with Asynq (Redis-backed) or River (PostgreSQL-backed). Both offer similar functionality with better performance.

**Q: How do we replicate ActionCable functionality?**
A: Use Go's native WebSocket libraries (gorilla/websocket or nhooyr.io/websocket) with Redis Pub/Sub for multi-instance support. Go handles this better than Rails.

---

### Process

**Q: Should we freeze new features during rewrite?**
A: Recommended approach:
- Freeze major features
- Allow critical bug fixes
- Evaluate new feature requests carefully
- Communicate timeline to stakeholders

**Q: How do we maintain the Rails app during rewrite?**
A: Minimal maintenance mode:
- Security updates only
- Critical bug fixes
- No new features
- Document workarounds for known issues

**Q: Can we involve the community?**
A: Yes! Loomio is open source (AGPL). Consider:
- Sharing progress publicly
- Accepting contributions on isolated features
- Beta testing with community
- Documentation contributions

**Q: How do we test the rewrite?**
A: Multi-layered approach:
- Unit tests for all new code (>80% coverage)
- Integration tests for feature parity
- Run existing E2E tests against Go backend
- Beta testing with real users
- Load testing before production

**Q: What's the rollback plan?**
A: Always have a rollback plan:
- Keep Rails app running and ready
- Use feature flags for gradual cutover
- Monitor metrics closely after launch
- Keep database backups current
- Practice rollback procedures

---

### Team & Skills

**Q: Do we need Go experts on the team?**
A: Not necessarily. Go is relatively easy to learn for experienced developers. Ideal team:
- 1 Go expert (lead/mentor)
- 2-3 intermediate developers willing to learn
- Good engineering practices matter more than Go expertise

**Q: How long does it take to learn Go?**
A: For experienced developers:
- Basic proficiency: 1-2 weeks
- Comfortable writing production code: 1-2 months
- Deep expertise: 6-12 months

**Q: Should we hire or train?**
A: Hybrid approach works best:
- Train existing team (they know the domain)
- Hire 1 experienced Go developer as mentor
- Use pair programming for knowledge transfer

**Q: What if key people leave?**
A: Mitigate with:
- Comprehensive documentation
- Pair programming (spread knowledge)
- Code reviews (multiple people understand each part)
- Video walkthroughs of key systems
- Onboarding documentation

---

### Data & Migration

**Q: Will we lose any data during migration?**
A: Not if done carefully:
- Multiple backups before migration
- Dry runs in staging
- Validation scripts to compare data
- Checksums to verify integrity
- Rollback plan ready

**Q: Can we migrate data gradually?**
A: Yes, use one of these strategies:
- Dual write (write to both systems)
- Shared database (both apps use same DB)
- CDC (Change Data Capture) for streaming changes
- Periodic synchronization

**Q: What if the data doesn't fit Go's data model?**
A: You may need:
- Data transformation layer
- Migration scripts to reshape data
- Temporary compatibility shims
- Gradual schema evolution

---

### Post-Launch

**Q: When can we retire the Rails app?**
A: After:
- Go app stable for 1-3 months in production
- All features migrated and tested
- All users successfully using new system
- No critical issues discovered
- Team confident in Go codebase

**Q: What about ongoing maintenance?**
A: Go typically requires less maintenance:
- Simpler deployment (single binary)
- Fewer dependencies to update
- More explicit code (fewer surprises)
- Better performance (less scaling issues)

**Q: How do we measure success?**
A: Track these metrics:
- Response time improvements
- Server resource usage (CPU, memory)
- Deployment time
- Bug rate (should decrease)
- Development velocity (should increase over time)
- Team satisfaction
- User satisfaction

---

## Common Pitfalls

### ❌ Don't Do This

**Underestimating complexity**
- "It's just CRUD" - there's always hidden complexity
- Add 30-50% buffer to estimates

**Ignoring the database**
- Database is often the bottleneck, not the application
- Profile and optimize queries in both systems

**Rewriting without understanding**
- Don't blindly translate code
- Understand the business logic first

**Premature optimization**
- Make it work, then make it fast
- Profile before optimizing

**Skipping tests**
- "We'll add tests later" - you won't
- Test-driven development pays off

**No rollback plan**
- Always have a way back
- Practice rollback procedures

### ✅ Do This Instead

**Start with a vertical slice**
- Pick one complete feature
- Build it end-to-end
- Learn from it before scaling

**Automate everything**
- CI/CD from day one
- Automated testing
- Automated deployments

**Communicate constantly**
- Over-communication is better than under
- Regular updates to all stakeholders
- Celebrate small wins

**Document decisions**
- ADRs for architectural choices
- Comments for complex logic
- README for setup

**Plan for failure**
- Things will go wrong
- Have contingencies
- Stay flexible

---

## Useful Commands

### Go Commands
```bash
go mod init              # Initialize new module
go mod tidy              # Clean up dependencies
go test ./...            # Run all tests
go test -cover ./...     # Test with coverage
go build                 # Compile application
go run main.go           # Run directly
go fmt ./...             # Format code
go vet ./...             # Static analysis
golangci-lint run        # Comprehensive linting
```

### Rails Commands (for comparison)
```bash
rails routes             # List all routes
rails db:schema:dump     # Export schema
rails console            # Interactive console
rails stats              # Code statistics
bundle list              # List gems
rake -T                  # List rake tasks
```

### Docker Commands
```bash
docker build -t app .                # Build image
docker run -p 8080:8080 app         # Run container
docker-compose up                    # Start services
docker-compose down                  # Stop services
docker logs -f <container>          # View logs
docker exec -it <container> sh      # Shell access
```

---

## Resources by Topic

### Learning Go
- Tour of Go: https://go.dev/tour/
- Go by Example: https://gobyexample.com/
- Effective Go: https://go.dev/doc/effective_go
- Go Standard Library: https://pkg.go.dev/std

### Go Project Structure
- golang-standards/project-layout
- How I Write HTTP Services in Go (Mat Ryer)
- Practical Go (Dave Cheney)

### Rails to Go Migration
- Search: "rails to go migration case study"
- Search: "golang for rails developers"

### Loomio-Specific
- Loomio GitHub: https://github.com/loomio/loomio
- Loomio Docs: https://help.loomio.com/
- Loomio Handbook: https://github.com/loomio/loomio-coop-handbook

### Tools
- golangci-lint: https://golangci-lint.run/
- testify: https://github.com/stretchr/testify
- Delve (debugger): https://github.com/go-delve/delve
- Air (live reload): https://github.com/cosmtrek/air

---

## Quick Reference: Rails → Go

| Rails | Go Equivalent | Notes |
|-------|---------------|-------|
| `rails new` | `go mod init` | Initialize project |
| `rails server` | `go run main.go` | Start server |
| `rails console` | Custom REPL or debugger | No built-in console |
| `rails generate` | Manual or code generation tools | More explicit |
| `bundle install` | `go mod download` | Install dependencies |
| `rake db:migrate` | Custom migrations or tools | More manual |
| `puts` | `fmt.Println()` | Print to console |
| `nil` | `nil` | Null value |
| `def method_name` | `func MethodName()` | Function definition |
| `@instance_var` | `self.InstanceVar` | Instance variable |
| `.each` | `for range` | Iteration |
| `rescue` | `if err != nil` | Error handling |

---

## Acronyms & Abbreviations

- **ADR**: Architecture Decision Record
- **API**: Application Programming Interface
- **CRUD**: Create, Read, Update, Delete
- **CI/CD**: Continuous Integration/Continuous Deployment
- **DB**: Database
- **DNS**: Domain Name System
- **E2E**: End-to-End
- **ER**: Entity-Relationship
- **HTTP**: Hypertext Transfer Protocol
- **JSON**: JavaScript Object Notation
- **JWT**: JSON Web Token
- **K8s**: Kubernetes
- **LOC**: Lines of Code
- **MVP**: Minimum Viable Product
- **N+1**: N+1 Query Problem
- **ORM**: Object-Relational Mapping
- **POC**: Proof of Concept
- **REST**: Representational State Transfer
- **SQL**: Structured Query Language
- **TDD**: Test-Driven Development
- **UI**: User Interface
- **VPS**: Virtual Private Server
- **WBS**: Work Breakdown Structure
- **WS**: WebSocket

---

## Getting Help

### During Planning Phase
- Review all meta-plan documents
- Ask in team discussions
- Research similar migrations
- Consult with Go experts

### During Development
- Go documentation (go.dev/doc)
- Stack Overflow (golang tag)
- Go Forum (forum.golangbridge.org)
- Reddit (r/golang)
- Gopher Slack (invite.slack.golangbridge.org)

### For Loomio-Specific
- Loomio GitHub issues
- Loomio community forums
- Existing contributor documentation

---

**Remember:** There are no stupid questions. If you're confused about something, chances are others are too. Document the answers for the next person!

---

*Last Updated: [Date]*
*Next Review: Monthly during planning phase*