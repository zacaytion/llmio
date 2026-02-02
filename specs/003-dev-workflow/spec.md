# Feature Specification: Local Development Workflow

**Feature Branch**: `003-dev-workflow`
**Created**: 2026-02-02
**Status**: Draft
**Input**: User description: "Local Development Workflow Improvement - compose.yml and Makefile for local development infrastructure"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Start Development Services (Priority: P1)

A developer wants to quickly spin up all required infrastructure services (database, cache, email testing) with a single command so they can begin working on the Go backend immediately.

**Why this priority**: Without running services, no development or testing can occur. This is the foundational capability that enables all other work.

**Independent Test**: Can be fully tested by running a single command and verifying all services respond on their expected ports.

**Acceptance Scenarios**:

1. **Given** a fresh clone of the repository with Podman installed, **When** the developer runs `make up`, **Then** PostgreSQL, Redis, PgAdmin, and Mailpit containers start and become healthy within 60 seconds
2. **Given** running services, **When** the developer runs `make down`, **Then** all containers stop gracefully
3. **Given** running services, **When** the developer runs `make logs`, **Then** aggregated logs from all services stream to the terminal

---

### User Story 2 - Build and Run Go Application (Priority: P1)

A developer wants to build and run the Go server and migration tools using simple, memorable commands.

**Why this priority**: Core development loop - developers need to compile and run their code frequently throughout the day.

**Independent Test**: Can be fully tested by running build commands and verifying binaries are created, then running the server and hitting an endpoint.

**Acceptance Scenarios**:

1. **Given** the repository with Go installed, **When** the developer runs `make build`, **Then** both `server` and `migrate` binaries are created in the `bin/` directory
2. **Given** a built codebase, **When** the developer runs `make run-server`, **Then** the API server starts and accepts connections
3. **Given** a running database, **When** the developer runs `make run-migrate`, **Then** database migrations execute

---

### User Story 3 - Run Tests with Coverage (Priority: P2)

A developer wants to run tests and view coverage reports to ensure code quality before committing.

**Why this priority**: Critical for maintaining code quality, but secondary to being able to run the application.

**Independent Test**: Can be fully tested by running test command and verifying coverage file is generated with valid data.

**Acceptance Scenarios**:

1. **Given** a codebase with tests, **When** the developer runs `make test`, **Then** all tests run and a coverage file is generated at `.var/coverage/coverage.out`
2. **Given** a coverage file exists, **When** the developer runs `make coverage-view`, **Then** an HTML coverage report opens in the default browser

---

### User Story 4 - Code Quality Checks (Priority: P2)

A developer wants to lint and format code to maintain consistent style and catch issues early.

**Why this priority**: Important for code quality, but developers can work without it initially.

**Independent Test**: Can be fully tested by introducing a style violation and verifying the linter catches it.

**Acceptance Scenarios**:

1. **Given** Go code in the repository, **When** the developer runs `make lint`, **Then** linter output displays in the terminal AND is saved to `.var/log/golangci-lint.log`
2. **Given** code with auto-fixable issues, **When** the developer runs `make lint-fix`, **Then** issues are automatically corrected
3. **Given** unformatted code, **When** the developer runs `make fmt`, **Then** code is formatted according to project standards

---

### User Story 5 - Database Management via PgAdmin (Priority: P3)

A developer wants to visually inspect and manage the database through a web interface without manual connection configuration.

**Why this priority**: Convenience feature - developers can use command-line tools as an alternative.

**Independent Test**: Can be fully tested by accessing PgAdmin web UI and verifying the database connection works without manual setup.

**Acceptance Scenarios**:

1. **Given** services are running, **When** the developer opens `http://localhost:5050`, **Then** PgAdmin loads with the development database pre-configured
2. **Given** PgAdmin is open, **When** the developer clicks the pre-configured server, **Then** they can browse tables without entering connection details

---

### User Story 6 - Email Testing (Priority: P3)

A developer wants to test email functionality without sending real emails, capturing all outbound mail locally.

**Why this priority**: Only needed when working on email-related features.

**Independent Test**: Can be fully tested by sending a test email from the application and viewing it in Mailpit.

**Acceptance Scenarios**:

1. **Given** services are running, **When** the application sends an email to SMTP port 1025, **Then** the email appears in Mailpit's web UI at `http://localhost:8025`

---

### Edge Cases

- What happens when Podman is not installed? The `make up` command should fail with a clear error message.
- What happens when ports are already in use? Container startup should fail with port conflict error.
- What happens when `.env` file is missing? Commands should fail with guidance to copy from `.env.example`.
- What happens when running `make coverage-view` without a coverage file? The command should fail with a helpful error.

## Requirements *(mandatory)*

### Functional Requirements

#### Container Services

- **FR-001**: System MUST provide a `compose.yml` file defining PostgreSQL 18, Redis 8, PgAdmin4, and Mailpit services
- **FR-002**: PostgreSQL MUST be accessible on port 5432 with credentials from environment variables
- **FR-003**: Redis MUST be accessible on port 6379
- **FR-004**: PgAdmin MUST be accessible on port 5050 with pre-configured connection to PostgreSQL
- **FR-005**: Mailpit MUST provide web UI on port 8025 and SMTP on port 1025
- **FR-006**: All services MUST include health checks for dependency ordering
- **FR-007**: PostgreSQL and PgAdmin data MUST persist across container restarts via named volumes (`postgres_data`, `pgadmin_data`); Redis and Mailpit data is ephemeral (no volumes)

#### Environment Configuration

- **FR-008**: System MUST use `.env` file for service credentials and configuration
- **FR-009**: System MUST provide `.env.example` template with default development values
- **FR-010**: `.env` file MUST be gitignored; `.env.example` MUST be committed

#### Build Targets

- **FR-011**: `make build` MUST compile all Go binaries to `bin/` directory
- **FR-012**: `make build-server` MUST compile only the server binary
- **FR-013**: `make build-migrate` MUST compile only the migrate binary
- **FR-014**: `bin/` directory MUST exist with `.gitkeep` (binaries gitignored)

#### Run Targets

- **FR-015**: `make run-server` MUST start the API server via `go run`
- **FR-016**: `make run-migrate` MUST execute database migrations

#### Dependency Targets

- **FR-017**: `make install` MUST download Go dependencies
- **FR-018**: `make tidy` MUST tidy and verify Go module dependencies

#### Quality Targets

- **FR-019**: `make lint` MUST run golangci-lint with output to terminal AND `.var/log/golangci-lint.log`
- **FR-020**: `make lint-fix` MUST run golangci-lint with auto-fix enabled
- **FR-021**: `make fmt` MUST format code using gofmt and goimports

#### Testing Targets

- **FR-022**: `make test` MUST run all tests and generate coverage at `.var/coverage/coverage.out`
- **FR-023**: `make coverage-view` MUST generate HTML report and open in browser
- **FR-024**: `coverage-view` MUST have `.var/coverage/coverage.out` as a prerequisite

#### Container Targets

- **FR-025**: `make up` MUST start all services via `podman compose up -d`
- **FR-026**: `make down` MUST stop all services via `podman compose down`
- **FR-027**: `make logs` MUST tail logs for all services
- **FR-028**: Individual log targets (`logs-postgres`, `logs-redis`, `logs-pgadmin`, `logs-mailpit`) MUST tail specific service logs
- **FR-029**: `make clean-volumes` MUST stop services and delete `postgres_data` and `pgadmin_data` volumes for fresh starts

#### Makefile Structure

- **FR-030**: Makefile MUST be self-documenting with `make help` as default target
- **FR-031**: Targets MUST be grouped by category (Build, Run, Dependencies, Quality, Testing, Containers)
- **FR-032**: Makefile MUST declare proper prerequisites for directory creation

### Key Entities

- **Environment Variables**: Configuration values for service credentials (database user, password, database name, PgAdmin login)
- **Named Volumes**: Persistent storage for PostgreSQL data (`postgres_data`) and PgAdmin settings (`pgadmin_data`)
- **Build Artifacts**: Compiled Go binaries (`bin/server`, `bin/migrate`)
- **Coverage Reports**: Test coverage data (`.var/coverage/coverage.out`, `.var/coverage/coverage.html`)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developer can start all services and begin coding within 2 minutes of cloning the repository
- **SC-002**: All Makefile targets are discoverable via `make` or `make help` without reading documentation
- **SC-003**: Build, test, and lint commands complete successfully on a fresh clone after `make install`
- **SC-004**: PgAdmin connects to database automatically without manual configuration steps
- **SC-005**: Test coverage reports are generated and viewable with two commands (`make test` then `make coverage-view`)
- **SC-006**: Linter output is immediately visible in terminal while also being logged for later reference

## Clarifications

### Session 2026-02-02

- Q: Should Redis and Mailpit data persist across container restarts? → A: No persistence (ephemeral) - data cleared on `make down`. Added `make clean-volumes` target for deleting PostgreSQL and PgAdmin volumes when fresh starts are needed.

## Assumptions

- Podman and Podman Compose are installed on the developer's machine
- Go 1.25+ is installed (managed via mise)
- Standard ports (5432, 6379, 5050, 8025, 1025) are available on the developer's machine
- macOS `open` command is available for opening HTML files in browser (coverage-view target)
- golangci-lint is installed (managed via mise)
- goimports is installed for formatting
