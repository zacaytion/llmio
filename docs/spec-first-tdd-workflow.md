# Spec-First TDD Workflow Guide

A comprehensive guide for using speckit commands and superpowers skills to manage context and develop features with test-driven discipline.

## Overview

This workflow combines **speckit** (specification management) with **superpowers** (development discipline) to create a structured, TDD-based development flow. Each stage produces artifacts that feed into the next, maintaining traceability from requirements to implementation.

## The Pipeline

| Stage | Command | Output | Purpose |
|-------|---------|--------|---------|
| 1. **Specify** | `/speckit.specify` | `specs/N-feature/spec.md` | WHAT users need (business requirements) |
| 2. **Clarify** | `/speckit.clarify` | Updated `spec.md` | Fill gaps via targeted questions |
| 3. **Plan** | `/speckit.plan` | `plan.md`, `research.md`, `data-model.md`, `contracts/` | HOW to build it (technical design) |
| 4. **Tasks** | `/speckit.tasks` | `tasks.md` | Ordered implementation tasks with TDD |
| 5. **Analyze** | `/speckit.analyze` | Consistency report | Cross-artifact validation |
| 6. **Implement** | `/speckit.implement` | Code | Execute tasks with TDD |

## Superpowers Integration Points

Use these superpowers skills at strategic moments to enforce discipline:

| Superpower Skill | When to Use | How It Helps |
|------------------|-------------|--------------|
| `/superpowers:brainstorming` | **Before** `/speckit.specify` | Explore requirements, edge cases, user needs before committing to spec |
| `/superpowers:writing-plans` | **After** spec is clear | Transition from spec to technical plan with structure |
| `/superpowers:test-driven-development` | **During** `/speckit.implement` | Enforces Red-Green-Refactor cycle rigorously |
| `/superpowers:systematic-debugging` | When tests fail unexpectedly | Root cause analysis before attempting fixes |
| `/superpowers:verification-before-completion` | **Before** claiming done | Ensures tests actually pass, evidence before assertions |
| `/superpowers:requesting-code-review` | After implementing a user story | Verify implementation against spec requirements |
| `/superpowers:dispatching-parallel-agents` | When tasks.md has `[P]` markers | Run independent tasks concurrently for speed |
| `/superpowers:executing-plans` | When plan.md is ready | Execute plan in separate session with review checkpoints |

## Detailed Workflow

### Phase 1: Requirements (Spec)

```bash
# Optional: Brainstorm before specifying
/superpowers:brainstorming "Feature description here"

# Create the specification
/speckit.specify Add user authentication with OAuth2 and email/password
```

**What happens:**
- Creates feature branch `N-feature-name`
- Generates `specs/N-feature-name/spec.md`
- Produces quality checklist at `specs/N-feature-name/checklists/requirements.md`

**Output artifacts:**
- `spec.md` - User stories, acceptance scenarios, requirements, success criteria

### Phase 2: Clarification (Optional)

```bash
/speckit.clarify
```

**When to use:**
- Spec contains `[NEEDS CLARIFICATION]` markers
- Requirements are ambiguous
- Edge cases need definition

**What happens:**
- Identifies underspecified areas
- Asks up to 5 targeted questions
- Updates spec with answers

### Phase 3: Planning

```bash
/speckit.plan
```

**What happens:**
- Validates against project constitution
- Researches codebase for existing patterns
- Generates technical design documents

**Output artifacts:**
- `plan.md` - Technical context, project structure, complexity tracking
- `research.md` - Codebase analysis, existing patterns
- `data-model.md` - Entity definitions, relationships
- `contracts/` - API contracts, interface definitions

### Phase 4: Task Generation

```bash
/speckit.tasks
```

**What happens:**
- Reads all design documents
- Creates ordered, dependency-aware task list
- Organizes tasks by user story for independent delivery
- Marks parallelizable tasks with `[P]`

**Output artifacts:**
- `tasks.md` - Phased task list with TDD structure

**Task structure:**
```
Phase 1: Setup (shared infrastructure)
Phase 2: Foundational (blocking prerequisites)
Phase 3+: User Stories (independently implementable)
Phase N: Polish (cross-cutting concerns)
```

### Phase 5: Validation

```bash
/speckit.analyze
```

**What happens:**
- Cross-validates spec ↔ plan ↔ tasks
- Checks for inconsistencies
- Identifies gaps or contradictions

**Use before implementation to catch issues early.**

### Phase 6: Implementation

```bash
# Full automated implementation
/speckit.implement

# Or with manual TDD control
/superpowers:test-driven-development
```

**TDD Cycle (per task):**
1. Write failing test
2. Run test - confirm RED
3. Implement minimum code to pass
4. Run test - confirm GREEN
5. Refactor if needed
6. Commit

### Phase 7: Verification

```bash
/superpowers:verification-before-completion
```

**Before claiming any task complete:**
- Run all tests
- Confirm output shows passing
- Evidence before assertions

## Context Management

### Feature Branches Isolate Context

Each `/speckit.specify` creates:
```
specs/N-feature-name/
├── spec.md              # Requirements (Phase 1)
├── checklists/          # Quality validation
│   └── requirements.md
├── plan.md              # Technical design (Phase 3)
├── research.md          # Codebase analysis (Phase 3)
├── data-model.md        # Entity design (Phase 3)
├── contracts/           # API definitions (Phase 3)
└── tasks.md             # Implementation tasks (Phase 4)
```

### Constitution Enforces Consistency

The project constitution (`.specify/memory/constitution.md`) defines:
- Mandatory TDD (tests before implementation)
- Spec-first API design (OpenAPI before code)
- Technology constraints (Go, SvelteKit, PostgreSQL)
- Code review requirements

**All plans are validated against the constitution.**

### Artifacts Reference Each Other

- `tasks.md` references `plan.md`, `spec.md`, `data-model.md`
- `plan.md` references `spec.md` and constitution
- Implementation references `tasks.md` for order and dependencies

**When context is needed, read the referenced artifacts.**

## Parallel Execution

When `tasks.md` contains `[P]` markers, use:

```bash
/superpowers:dispatching-parallel-agents
```

This launches independent tasks concurrently:
- Different files, no dependencies
- Models can be created in parallel
- Tests can be written in parallel

## Checklists

Generate custom quality checklists:

```bash
/speckit.checklist
```

Creates validation checklists based on:
- Feature requirements
- Project standards
- Constitution principles

## GitHub Integration

Convert tasks to GitHub issues:

```bash
/speckit.taskstoissues
```

Creates dependency-ordered issues for team collaboration.

## Quick Reference

### Starting a New Feature
```bash
/superpowers:brainstorming "feature idea"     # Optional: explore first
/speckit.specify Feature description here     # Create spec
/speckit.clarify                              # If gaps exist
/speckit.plan                                 # Technical design
/speckit.analyze                              # Validate consistency
/speckit.tasks                                # Generate task list
```

### Implementing a Feature
```bash
/speckit.implement                            # Automated TDD implementation
# OR
/superpowers:test-driven-development          # Manual TDD control
```

### Before Marking Complete
```bash
/superpowers:verification-before-completion   # Evidence before assertions
/superpowers:requesting-code-review           # Verify against requirements
```

### Debugging Failures
```bash
/superpowers:systematic-debugging             # Root cause before fixes
```

## Best Practices

1. **Never skip specification** - All features start with `/speckit.specify`
2. **Constitution is law** - Violations require documented justification
3. **TDD is non-negotiable** - Tests fail before implementation exists
4. **Incremental delivery** - Each user story is independently deployable
5. **Verify before claiming** - Run tests, show output, then claim success
6. **Parallel when possible** - Use `[P]` markers and dispatching agents
7. **Review against spec** - Implementation must satisfy acceptance scenarios
