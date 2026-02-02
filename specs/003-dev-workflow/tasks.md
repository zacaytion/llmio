# Tasks: Local Development Workflow

**Input**: Design documents from `/specs/003-dev-workflow/`
**Prerequisites**: plan.md, spec.md, research.md, quickstart.md

**Tests**: Not applicable - this feature creates configuration files verified via manual testing of Makefile targets.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Files created at repository root:
- `compose.yml` - Podman Compose services
- `Makefile` - Development commands
- `.env.example` - Environment template
- `bin/.gitkeep` - Binary directory placeholder
- `docker/pgadmin/servers.json` - PgAdmin pre-configuration

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create environment configuration and directory structure

- [x] T001 [P] Create `.env.example` with default development credentials at `.env.example`
- [x] T002 [P] Create `bin/` directory with `.gitkeep` at `bin/.gitkeep`
- [x] T003 [P] Create PgAdmin servers config at `docker/pgadmin/servers.json`
- [x] T004 Update `.gitignore` to ignore `.env` and `bin/*` (keep `bin/.gitkeep`)

**Checkpoint**: Environment and directory structure ready

---

## Phase 2: Foundational (Compose Services)

**Purpose**: Create compose.yml with all services - MUST be complete before Makefile targets can be verified

**⚠️ CRITICAL**: Container targets in Makefile depend on compose.yml existing

- [x] T005 Create `compose.yml` with PostgreSQL 18 service (port 5432, health check, named volume `postgres_data`)
- [x] T006 Add Redis 8 service to `compose.yml` (port 6379, health check, no volume)
- [x] T007 Add PgAdmin4 service to `compose.yml` (port 5050, servers.json mount, named volume `pgadmin_data`, depends_on postgres healthy)
- [x] T008 Add Mailpit service to `compose.yml` (ports 8025/1025, no volume)
- [x] T009 Define named volumes (`postgres_data`, `pgadmin_data`) in `compose.yml`

**Checkpoint**: `podman compose up -d` should start all services successfully

---

## Phase 3: User Story 1 - Start Development Services (Priority: P1) 🎯 MVP

**Goal**: Developer can start/stop all services with simple commands

**Independent Test**: Run `make up`, verify all services respond on expected ports, run `make down`

### Implementation for User Story 1

- [x] T010 [US1] Create `Makefile` with `.DEFAULT_GOAL := help` and `.PHONY` declarations
- [x] T011 [US1] Add `help` target with awk-based self-documentation to `Makefile`
- [x] T012 [US1] Add `##@ Containers` section header to `Makefile`
- [x] T013 [US1] Add `up` target (`podman compose up -d`) to `Makefile`
- [x] T014 [US1] Add `down` target (`podman compose down`) to `Makefile`
- [x] T015 [US1] Add `logs` target (`podman compose logs -f`) to `Makefile`
- [x] T016 [P] [US1] Add `logs-postgres` target to `Makefile`
- [x] T017 [P] [US1] Add `logs-redis` target to `Makefile`
- [x] T018 [P] [US1] Add `logs-pgadmin` target to `Makefile`
- [x] T019 [P] [US1] Add `logs-mailpit` target to `Makefile`
- [x] T020 [US1] Add `clean-volumes` target (stop + remove postgres_data and pgadmin_data) to `Makefile`

**Checkpoint**: `make up`, `make down`, `make logs`, `make clean-volumes` all work

---

## Phase 4: User Story 2 - Build and Run Go Application (Priority: P1)

**Goal**: Developer can build binaries and run server/migrations

**Independent Test**: Run `make build`, verify binaries in `bin/`, run `make run-server`

### Implementation for User Story 2

- [x] T021 [US2] Add `##@ Build` section header to `Makefile`
- [x] T022 [US2] Add `build-server` target (`go build -o bin/server ./cmd/server`) to `Makefile`
- [x] T023 [US2] Add `build-migrate` target (`go build -o bin/migrate ./cmd/migrate`) to `Makefile`
- [x] T024 [US2] Add `build` target (depends on build-server, build-migrate) to `Makefile`
- [x] T025 [US2] Add `##@ Run` section header to `Makefile`
- [x] T026 [US2] Add `run-server` target (`go run ./cmd/server`) to `Makefile`
- [x] T027 [US2] Add `run-migrate` target (`go run ./cmd/migrate up`) to `Makefile`
- [x] T028 [US2] Add `##@ Dependencies` section header to `Makefile`
- [x] T029 [US2] Add `install` target (`go mod download`) to `Makefile`
- [x] T030 [US2] Add `tidy` target (`go mod tidy`) to `Makefile`

**Checkpoint**: `make build`, `make run-server`, `make run-migrate` all work

---

## Phase 5: User Story 3 - Run Tests with Coverage (Priority: P2)

**Goal**: Developer can run tests and view coverage reports

**Independent Test**: Run `make test`, verify `.var/coverage/coverage.out` exists, run `make coverage-view`

### Implementation for User Story 3

- [x] T031 [US3] Add `.var/coverage` directory prerequisite target to `Makefile`
- [x] T032 [US3] Add `##@ Testing` section header to `Makefile`
- [x] T033 [US3] Add `test` target (go test with coverprofile, depends on .var/coverage) to `Makefile`
- [x] T034 [US3] Add `.var/coverage/coverage.out` file prerequisite target to `Makefile`
- [x] T035 [US3] Add `coverage-view` target (go tool cover -html, open browser, depends on coverage.out) to `Makefile`

**Checkpoint**: `make test` generates coverage, `make coverage-view` opens HTML report

---

## Phase 6: User Story 4 - Code Quality Checks (Priority: P2)

**Goal**: Developer can lint and format code

**Independent Test**: Run `make lint`, verify output in terminal and `.var/log/golangci-lint.log`

### Implementation for User Story 4

- [x] T036 [US4] Add `.var/log` directory prerequisite target to `Makefile`
- [x] T037 [US4] Add `##@ Quality` section header to `Makefile`
- [x] T038 [US4] Add `lint` target (golangci-lint with tee to terminal and log, depends on .var/log) to `Makefile`
- [x] T039 [US4] Add `lint-fix` target (golangci-lint --fix) to `Makefile`
- [x] T040 [US4] Add `fmt` target (gofmt -w . && goimports -w -local github.com/zacaytion/llmio .) to `Makefile`

**Checkpoint**: `make lint`, `make lint-fix`, `make fmt` all work

---

## Phase 7: User Story 5 & 6 - PgAdmin and Mailpit (Priority: P3)

**Goal**: PgAdmin connects automatically; Mailpit captures test emails

**Independent Test**: Access http://localhost:5050, verify pre-configured connection; send test email to port 1025, verify in http://localhost:8025

### Implementation for User Stories 5 & 6

No additional Makefile tasks needed - these are fulfilled by:
- T003: PgAdmin servers.json configuration
- T007: PgAdmin service with volume mount
- T008: Mailpit service

**Checkpoint**: PgAdmin shows pre-configured server; Mailpit UI accessible

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final verification and documentation

- [x] T041 Copy `.env.example` to `.env` for local testing
- [x] T042 Verify `make help` displays all targets with descriptions
- [x] T043 Run full workflow: `make up` → `make install` → `make run-migrate` → `make test` → `make down`
- [x] T044 Verify `make clean-volumes` removes persistent data
- [x] T045 Verify `make coverage-view` fails gracefully when no coverage.out exists
- [x] T046 Verify `make up` fails with clear error when port 5432 is already in use (N/A: Podman VM abstracts ports)
- [x] T047 Add `server` and `migrate` targets that run pre-built binaries with ARGS passthrough
- [x] T048 Add `clean` target that removes volumes, binaries, and test/lint artifacts
- [x] T049 Add `clean-go-{build,test,mod,fuzz,all}` targets for Go cache management

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup (T003 for servers.json) - BLOCKS container Makefile targets
- **User Story 1 (Phase 3)**: Depends on Foundational (compose.yml must exist)
- **User Story 2 (Phase 4)**: Depends only on Makefile existing (T010)
- **User Stories 3-4 (Phase 5-6)**: Depend only on Makefile existing (T010)
- **User Stories 5-6 (Phase 7)**: Fulfilled by Phase 2 tasks
- **Polish (Phase 8)**: Depends on all phases complete

### User Story Dependencies

- **User Story 1 (Containers)**: Requires compose.yml (Phase 2) before verification
- **User Story 2 (Build/Run)**: Independent - can implement after T010 creates Makefile
- **User Story 3 (Testing)**: Independent - can implement after T010 creates Makefile
- **User Story 4 (Quality)**: Independent - can implement after T010 creates Makefile
- **User Stories 5-6 (PgAdmin/Mailpit)**: Fulfilled by compose.yml in Phase 2

### Within Makefile Development

- T010 (create Makefile) must come first
- Section headers (T012, T021, T025, T028, T032, T037) before their targets
- Directory prerequisites (T031, T036) before targets that depend on them
- Individual log targets (T016-T019) can run in parallel

### Parallel Opportunities

```bash
# Phase 1 - all parallel:
T001, T002, T003, T004

# Phase 2 - sequential (compose.yml is single file):
T005 → T006 → T007 → T008 → T009

# User Story 1 - individual log targets parallel:
T016, T017, T018, T019

# User Stories 2-4 can be developed in parallel after T010
```

---

## Parallel Example: Phase 1 Setup

```bash
# Launch all setup tasks together:
Task: "Create .env.example"
Task: "Create bin/.gitkeep"
Task: "Create docker/pgadmin/servers.json"
Task: "Update .gitignore"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (compose.yml)
3. Complete Phase 3: User Story 1 (container targets)
4. **STOP and VALIDATE**: `make up`, verify services, `make down`
5. Developer can now use containers for development

### Incremental Delivery

1. Complete Setup + Foundational → compose.yml ready
2. Add User Story 1 → `make up/down/logs` work (MVP!)
3. Add User Story 2 → `make build/run-*` work
4. Add User Stories 3-4 → `make test/lint/fmt` work
5. Polish → Full workflow validated

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- compose.yml is a single file, so Phase 2 tasks are sequential
- Makefile is also a single file, but organized by section
- Commit after each phase checkpoint
- Test each user story independently before moving to next
