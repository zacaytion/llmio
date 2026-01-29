# Execution Guide: How to Use the Meta-Plan

This guide explains how to systematically work through the meta-plan to create a comprehensive rewrite plan for Loomio in Go.

---

## Overview

The meta-plan is structured in 7 phases. Think of it as a checklist and framework—not a rigid prescription. Adapt it to your team's needs and constraints.

**Time allocation suggestion:**
- Phase 1 (Discovery): 3-4 weeks
- Phase 2 (Planning Framework): 1-2 weeks
- Phase 3 (Detailed Plan Creation): 2-3 weeks
- Phases 4-7: Ongoing throughout execution

---

## Week-by-Week Breakdown

### Week 1: Repository Deep Dive

**Goal:** Understand the Rails codebase structure

**Tasks:**
1. Clone the Loomio repository locally
2. Set up a local development environment following DEVSETUP.md
3. Run the application and explore features as a user
4. Document the directory structure:
   ```
   app/
     controllers/
     models/
     views/
     services/
     jobs/
   config/
   db/
     migrate/
     schema.rb
   lib/
   spec/
   vue/
   ```

**Deliverables:**
- Architecture diagram (high-level)
- List of all models and their relationships
- List of all API endpoints
- Notes on any surprising/complex patterns

**Tools to use:**
- `rails routes` - list all routes
- `rails db:schema:dump` - export schema
- `tree -L 3` - visualize directory structure
- IDE navigation to explore code

---

### Week 2: Database & Dependencies

**Goal:** Map data layer and external dependencies

**Tasks:**
1. Export full database schema to a document
2. Identify all foreign keys and indexes
3. List all gems in Gemfile with purpose
4. Research Go equivalents for each gem
5. Document PostgreSQL-specific features used
6. Identify all external API integrations

**Deliverables:**
- Complete ER diagram
- Gem → Go package mapping spreadsheet
- Database migration complexity assessment
- External dependency risk matrix

**Tools to use:**
- `rails-erd` gem for ER diagrams
- SchemaSpy for database documentation
- Bundle graph for dependency visualization

---

### Week 3: Feature Inventory & Testing

**Goal:** Catalog all features and understand test coverage

**Tasks:**
1. Go through the UI and create a feature list
2. Cross-reference with user documentation
3. Rate each feature (complexity, criticality, usage)
4. Analyze test suite structure
5. Measure current test coverage
6. Document test patterns used

**Deliverables:**
- Feature inventory spreadsheet with ratings
- Test coverage report
- List of untested/poorly tested areas
- Testing strategy recommendations

**Tools to use:**
- SimpleCov for coverage reports
- Manual UI exploration
- User documentation review

---

### Week 4: Technical Stack Research

**Goal:** Choose Go ecosystem components

**Tasks:**
1. Create proof-of-concept (POC) projects for each decision:
   - Web framework comparison
   - ORM/query builder comparison
   - Background job system comparison
   - WebSocket solution comparison
2. Benchmark performance of key components
3. Evaluate documentation quality and community support
4. Consider team learning curve

**Deliverables:**
- Stack selection decision matrix
- POC repositories with benchmarks
- Recommendation document with rationale
- Learning resources list for chosen stack

**Sample POC structure:**
```
poc/
  gin-vs-echo/
  gorm-vs-sqlc/
  asynq-demo/
  websocket-demo/
```

---

### Week 5: Architecture Decisions

**Goal:** Make and document key architectural decisions

**Tasks:**
1. Create ADRs (Architecture Decision Records) for:
   - Migration strategy (strangler vs big bang)
   - API design (REST vs GraphQL)
   - Monolith vs microservices
   - Database migration approach
   - Error handling patterns
   - Logging strategy
   - Configuration management
2. Review ADRs with team
3. Get stakeholder approval on critical decisions

**Deliverables:**
- 7-10 ADR documents using standard format
- Architecture diagram (detailed)
- API design specification
- Migration strategy document

**ADR Template:**
```markdown
# ADR-XXX: [Title]

## Status
[Proposed | Accepted | Deprecated | Superseded]

## Context
What is the issue we're facing?

## Decision
What decision did we make?

## Consequences
What becomes easier or more difficult?

## Alternatives Considered
What other options did we evaluate?
```

---

### Week 6: Risk Assessment & Mitigation

**Goal:** Identify and plan for risks

**Tasks:**
1. Brainstorm all possible risks with team
2. Rate each risk (probability × impact)
3. Create mitigation strategies for high risks
4. Establish early warning indicators
5. Define escalation procedures
6. Create contingency plans

**Deliverables:**
- Risk register with 20-30 identified risks
- Mitigation plan for top 10 risks
- Risk monitoring dashboard/spreadsheet
- Escalation playbook

**Risk Categories to Consider:**
- Technical complexity
- Team skill gaps
- Data migration issues
- Performance regressions
- Security vulnerabilities
- Budget overruns
- Timeline delays
- User adoption problems
- Integration failures

---

### Week 7: Work Breakdown & Estimation

**Goal:** Break down work into estimable chunks

**Tasks:**
1. Create hierarchical work breakdown structure (WBS)
2. Estimate each work item (story points or time)
3. Identify dependencies between work items
4. Determine critical path
5. Allocate work to team members
6. Build initial timeline/Gantt chart

**Deliverables:**
- Complete WBS (3-4 levels deep)
- Effort estimates for all work items
- Dependency graph
- Initial project timeline (12-18 months)
- Resource allocation plan

**Estimation Tips:**
- Use planning poker for team estimates
- Add 30-50% buffer for unknowns
- Break down epics into 2-week sprints
- Consider team velocity from past projects

---

### Week 8: Plan Assembly & Review

**Goal:** Compile everything into final plan

**Tasks:**
1. Write executive summary
2. Compile all sections into master plan document
3. Create visual timeline and milestones
4. Write success criteria clearly
5. Define metrics and KPIs
6. Schedule plan review meetings
7. Incorporate feedback
8. Get formal sign-off

**Deliverables:**
- Complete project plan (50-100 pages)
- Executive presentation (10-15 slides)
- Approved and signed plan
- Kickoff meeting scheduled

**Plan Document Structure:**
```
1. Executive Summary (2 pages)
2. Project Scope & Goals (5 pages)
3. Current State Analysis (10 pages)
4. Technical Architecture (15 pages)
5. Migration Strategy (10 pages)
6. Work Breakdown & Timeline (20 pages)
7. Resource Plan (5 pages)
8. Risk Management (10 pages)
9. Testing Strategy (8 pages)
10. Success Metrics (5 pages)
11. Appendices (research, ADRs, etc.)
```

---

## Practical Tips

### Daily Practices

1. **Document as you go**: Don't wait until the end to write things down
2. **Share findings regularly**: Quick Slack updates, daily standups
3. **Use version control**: Keep all documents in Git
4. **Timebox research**: Don't get stuck in analysis paralysis (2-4 hours per topic max)
5. **Build small POCs**: Code speaks louder than words

### Tools & Templates

**Project Management:**
- Use GitHub Projects, Jira, or Linear for task tracking
- Miro or Figma for visual diagrams
- Google Sheets for matrices and inventories
- Notion or Confluence for documentation

**Code Analysis:**
- `tokei` - count lines of code
- `gocloc` - compare with Go LOC estimates
- `rails stats` - Rails codebase statistics
- GitHub dependency graph

**Communication:**
- Weekly written updates
- Bi-weekly demo sessions (show POCs)
- Dedicated Slack channel
- Recorded walkthrough videos

### Decision-Making Framework

When stuck on a decision:

1. **Define the decision clearly**: What exactly are we choosing?
2. **List criteria**: What matters? (performance, maintainability, team skill, etc.)
3. **Weight criteria**: Not everything is equally important
4. **Score options**: Rate each option against criteria
5. **Calculate**: Multiply scores by weights
6. **Gut check**: Does the math match intuition?
7. **Document**: Write it down as an ADR
8. **Time-box**: If still unclear after 2-4 hours, pick one and move on

### Red Flags to Watch For

🚩 **Analysis paralysis**: Spending >1 week on a single decision  
🚩 **Scope creep**: "Let's also rewrite the frontend in React..."  
🚩 **Underestimating complexity**: "It's just a CRUD app"  
🚩 **Ignoring team capacity**: Planning work without team input  
🚩 **No stakeholder engagement**: Building plan in isolation  
🚩 **Premature optimization**: Worrying about scale before basics work  
🚩 **Missing rollback plan**: "We'll figure it out if things go wrong"  

### Success Patterns

✅ **Start small**: Build one complete vertical slice first  
✅ **Automate early**: CI/CD from day one  
✅ **Test constantly**: Don't accumulate testing debt  
✅ **Communicate proactively**: Over-communicate rather than under  
✅ **Celebrate milestones**: Keep team motivated  
✅ **Learn in public**: Share progress with community  
✅ **Stay pragmatic**: Perfect is the enemy of done  

---

## Sample Execution Schedule

Here's a realistic 16-month timeline:

```
Month 1-2: Discovery & Planning (you are here)
Month 3: Foundation setup (project structure, CI/CD, basic auth)
Month 4-5: Core data models and database layer
Month 6-7: User and group management features
Month 8-9: Discussion and threading features
Month 10-11: Polling and voting features
Month 12: Integration work (email, notifications, websockets)
Month 13: Frontend integration and API stabilization
Month 14: Performance optimization and security hardening
Month 15: Beta testing and bug fixes
Month 16: Production migration and monitoring
```

---

## Checkpoint Questions

Ask these at the end of each week:

### Week 1-2 Checkpoint
- [ ] Can we describe the Rails app architecture in 5 minutes?
- [ ] Do we know all the models and their relationships?
- [ ] Have we identified the 10 most complex features?
- [ ] Do we understand the data flow?

### Week 3-4 Checkpoint
- [ ] Have we built POCs for critical decisions?
- [ ] Do we have confidence in our stack choices?
- [ ] Have we measured test coverage?
- [ ] Do we know what features are most important?

### Week 5-6 Checkpoint
- [ ] Have we made all critical architecture decisions?
- [ ] Are ADRs written and approved?
- [ ] Have we identified top 10 risks?
- [ ] Do we have mitigation plans?

### Week 7-8 Checkpoint
- [ ] Is the work broken down into 2-week chunks?
- [ ] Are estimates realistic with buffers?
- [ ] Do we have stakeholder buy-in?
- [ ] Is the plan document complete?
- [ ] Are we ready to start coding?

---

## Common Questions

**Q: Should we rewrite tests or start fresh?**  
A: Hybrid approach. Translate business logic tests, rewrite implementation tests. Aim for better coverage than original.

**Q: How do we handle the frontend?**  
A: Keep Vue.js frontend mostly unchanged. Focus backend rewrite first. Update API contracts carefully with versioning.

**Q: What if we discover something is impossible in Go?**  
A: Very rare. Document it, find workarounds, or keep that piece in Ruby temporarily (microservice approach).

**Q: How do we maintain the Rails app during rewrite?**  
A: Minimize new features. Bug fixes only. Communicate freeze period to stakeholders.

**Q: What's the biggest risk?**  
A: Usually data migration. Plan this carefully, test extensively, have rollback plan ready.

**Q: Should we go open source with the Go version?**  
A: Loomio is AGPL, so yes. Consider community involvement early.

---

## Resources

### Learning Go (if team needs it)
- Tour of Go: https://go.dev/tour/
- Effective Go: https://go.dev/doc/effective_go
- Go by Example: https://gobyexample.com/
- Ardan Labs courses: https://www.ardanlabs.com/

### Go Project Structure
- golang-standards/project-layout (controversial but helpful)
- How to structure Go apps (various blog posts)

### Rails to Go Migration Stories
- Search for "Rails to Go migration" case studies
- Read about Shopify, GitHub (partial), others

### Tools
- GitHub Copilot for code translation
- AI assistants for documentation
- Draw.io for diagrams
- Linear/Jira for task management

---

## Final Checklist Before Starting Implementation

- [ ] Complete plan document written and approved
- [ ] All stakeholders aligned on approach
- [ ] Team trained on Go basics
- [ ] Development environment set up
- [ ] CI/CD pipeline ready
- [ ] First sprint planned in detail
- [ ] Success metrics defined and agreed
- [ ] Risk mitigation strategies in place
- [ ] Communication channels established
- [ ] Kickoff meeting completed

---

## Summary

This meta-plan gives you a framework to:
1. **Systematically discover** what needs to be done
2. **Make informed decisions** about architecture and approach
3. **Create realistic estimates** based on actual complexity
4. **Manage risks** proactively
5. **Communicate effectively** with all stakeholders
6. **Execute confidently** with a solid foundation

Remember: **The plan will change as you learn more. That's expected and healthy.** The goal isn't a perfect plan—it's a plan that gets you started with confidence and adapts as you go.

Good luck with your rewrite! 🚀

---

**Next Action:** Start Week 1 tasks tomorrow. Set up a daily standup. Share progress in a visible place.