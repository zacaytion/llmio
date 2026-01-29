# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Current Progress

- [x] Phase 1: Discovery & Analysis — Complete (see `docs/plans/2026-01-29-loomio-discovery-report.md`)
- [x] Phase 2: Planning Framework — Complete
  - ADR-001: Migration Strategy (`docs/plans/2026-01-29-adr-001-migration-strategy.md`)
  - ADR-002: Channel Server (`docs/plans/2026-01-29-adr-002-channel-server-migration.md`)
  - ADR-003: Go Stack (`docs/plans/2026-01-29-adr-003-go-stack-selection.md`)
  - Risk Register (`docs/plans/2026-01-29-risk-register.md`)
  - Testing Strategy (`docs/plans/2026-01-29-testing-strategy.md`)
  - Team Resource Plan (`docs/plans/2026-01-29-team-resource-plan.md`)
  - Deployment Plan (`docs/plans/2026-01-29-deployment-rollout-plan.md`)
- [ ] Phase 3: Detailed Plan Creation — Next
- [ ] Phase 4-7: Execution phases

## Project Overview

This repository contains **meta-planning documentation** for rewriting Loomio (a collaborative decision-making tool) from Ruby on Rails to Go. This is NOT an implementation repository—it's a strategic planning framework.

**Important:** There is no code to build, test, or run. This is purely a documentation project containing planning templates, decision frameworks, and execution guides.

## Repository Structure

The repository follows a document hierarchy with interdependencies:

```
llmio/
├── README.md                    # Entry point - navigation and overview
├── META_PLAN.md                 # Strategic framework (7-phase approach)
├── EXECUTION_GUIDE.md           # Tactical week-by-week breakdown
├── GETTING_STARTED.md           # Day-by-day onboarding guide
├── DECISION_TREE.md             # 10 key decision frameworks with pros/cons
├── DISCOVERY_TEMPLATES.md       # Ready-to-use planning templates
└── GLOSSARY_AND_FAQ.md          # Terms, FAQs, Rails→Go comparisons
```

## Source Analysis Setup

The `orig/` directory (gitignored) contains shallow clones of Loomio repositories for analysis:

```
orig/
├── loomio/                 # Main Rails app (github.com/loomio/loomio)
├── loomio_channel_server/  # Real-time Node.js server (github.com/loomio/loomio_channel_server)
└── loomio-deploy/          # Docker Compose deployment (github.com/loomio/loomio-deploy)
```

Clone commands (shallow to save space):
- `git clone --depth 1 https://github.com/loomio/loomio.git orig/loomio`
- `git clone --depth 1 https://github.com/loomio/loomio_channel_server.git orig/loomio_channel_server`
- `git clone --depth 1 https://github.com/loomio/loomio-deploy.git orig/loomio-deploy`

## Analysis Outputs

Discovery reports and plans go in `docs/plans/` with date-prefixed filenames:
- `docs/plans/YYYY-MM-DD-<topic>-design.md`

Example: `docs/plans/2026-01-29-loomio-discovery-report.md`

## Document Hierarchy & Reading Order

**For new readers:**
1. Start with `README.md` for orientation
2. Read `GETTING_STARTED.md` for immediate next steps
3. Review `META_PLAN.md` sections 1-3 for strategic framework
4. Consult `DECISION_TREE.md` when making architectural choices
5. Use `DISCOVERY_TEMPLATES.md` as working documents
6. Reference `GLOSSARY_AND_FAQ.md` for definitions

**Document relationships:**
- `META_PLAN.md` defines the 7-phase strategic approach
- `EXECUTION_GUIDE.md` breaks down those phases into 8 weeks of tasks
- `GETTING_STARTED.md` provides day-by-day guidance for Week 1
- `DECISION_TREE.md` provides decision frameworks for choices mentioned in META_PLAN
- `DISCOVERY_TEMPLATES.md` provides templates referenced throughout other docs

## Key Concepts

### The 7-Phase Framework

All documents reference this core structure from `META_PLAN.md`:

1. **Phase 1: Discovery & Analysis** (3-4 weeks) - Repository analysis, database mapping, feature inventory
2. **Phase 2: Planning Framework** (1-2 weeks) - Migration strategy, team planning, risk assessment
3. **Phase 3: Detailed Plan Creation** (2-3 weeks) - Work breakdown, sprint planning, timeline
   - **Inputs:** Use ADRs (001-003), risk register, and testing strategy from Phase 2
   - **Outputs:** WBS with effort estimates, sprint backlog, milestone definitions
4. **Phase 4: Quality Assurance** (ongoing) - Testing strategy and monitoring
5. **Phase 5: Documentation** (ongoing) - ADRs, API docs, runbooks
6. **Phase 6: Progress Tracking** (ongoing) - Metrics and reporting
7. **Phase 7: Success Metrics** (ongoing) - KPI monitoring and validation

### Migration Strategies

Three approaches are discussed throughout the documents:
- **Big Bang Rewrite:** Complete rewrite before deployment (high risk)
- **Strangler Fig Pattern:** Gradual migration with complex routing (lower risk, higher complexity)
- **Hybrid Approach:** Independent module rewrites with parallel systems (recommended for Loomio)

### Architecture Decision Records (ADRs)

The framework heavily emphasizes documenting decisions using ADRs. Template structure:
```
# ADR-NNN: [Decision Title]
Date: YYYY-MM-DD
Status: [Proposed|Accepted|Deprecated|Superseded]

## Context
## Decision
## Consequences
## Alternatives Considered
```

## Working With This Repository

### When Making Edits

**Cross-reference consistency:** When updating one document, check for related content in others:
- Timeline changes in `EXECUTION_GUIDE.md` should reflect in `README.md` timeline summary
- New decisions in `DECISION_TREE.md` may need references in `META_PLAN.md`
- Template additions in `DISCOVERY_TEMPLATES.md` should be mentioned in relevant phase descriptions

**Maintain the planning stance:** This is a meta-plan (plan for creating a plan), not an implementation guide. Avoid:
- Specific code examples or Go implementations
- Detailed technical specifications
- Implementation-level decisions

Keep focus on:
- Planning processes and frameworks
- Decision-making approaches
- Template structures
- Checkpoint questions

### Document Formatting Conventions

- Use `[ ]` for task checklists throughout planning documents
- Reference other documents with markdown links: `[META_PLAN.md](META_PLAN.md)`
- Use code blocks with `bash` for command examples
- Use decision tree diagrams with text-based flowcharts
- Maintain consistent heading hierarchy (no skipping levels)

### Common Edit Patterns

**Adding a new phase task:**
1. Add to appropriate phase section in `META_PLAN.md`
2. Add corresponding week breakdown in `EXECUTION_GUIDE.md`
3. If it's a Phase 1 task, consider adding day-level detail to `GETTING_STARTED.md`
4. Add any templates needed to `DISCOVERY_TEMPLATES.md`

**Adding a decision point:**
1. Create decision tree in `DECISION_TREE.md` with pros/cons
2. Reference it in the relevant `META_PLAN.md` phase
3. Add any context to `GLOSSARY_AND_FAQ.md` if terms need definition

**Adding templates:**
1. Add template to `DISCOVERY_TEMPLATES.md`
2. Reference in the phase that uses it (`META_PLAN.md`)
3. Note in the deliverables for the corresponding week (`EXECUTION_GUIDE.md`)

## Document Maintenance

### Consistency Checks

When updating content, verify:
- Timeline estimates are consistent across all documents
- Phase numbers (1-7) are used consistently
- Tool recommendations in decision trees match those in execution guide
- Template names in references match actual section headers

### Red Flags

Watch for these issues when editing:
- Implementation details creeping into planning documents
- Inconsistent phase durations across documents
- Broken internal document references
- Templates referenced but not included
- Decision trees without clear recommendations

## Target Audience Context

These documents are written for:
- Technical leaders planning a Rails→Go migration
- Engineering teams evaluating rewrite strategies
- Stakeholders assessing timeline and scope
- Future implementers who will create the detailed plan

The tone should be:
- Pragmatic and action-oriented
- Comprehensive but not overwhelming
- Framework-focused rather than prescriptive
- Honest about trade-offs and challenges

## Document Purpose Summary

| Document | Purpose | Update When |
|----------|---------|-------------|
| README.md | Navigation and quick start | Structure changes, new documents added |
| META_PLAN.md | Strategic framework | Phase approach changes, new analysis areas |
| EXECUTION_GUIDE.md | Week-by-week tactics | Timeline changes, new tasks discovered |
| GETTING_STARTED.md | Day 1 onboarding | Week 1 priorities change |
| DECISION_TREE.md | Choice frameworks | New architectural decisions needed |
| DISCOVERY_TEMPLATES.md | Working templates | New deliverables required |
| GLOSSARY_AND_FAQ.md | Reference guide | New terms introduced, questions arise |

## No Build/Test Commands

Since this is a documentation-only repository:
- There is no code to build or compile
- There are no tests to run
- There is no deployment process
- There are no linters or formatters configured

The only "validation" is ensuring document consistency and completeness.
