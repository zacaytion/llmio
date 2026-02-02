# Feature Specification: Configuration System

**Feature Branch**: `002-config-system`
**Created**: 2026-02-02
**Status**: Draft
**Input**: User description: "Add centralized configuration system using Viper + Cobra + log/slog to eliminate hardcoded values, reduce dev friction, and enable separate test database setup"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Developer Configures Local Environment (Priority: P1)

A developer clones the repository and needs to run the application with their local database credentials and preferred port. They create a configuration file with their settings and start the server, which uses those values instead of defaults.

**Why this priority**: This is the core value proposition - eliminating hardcoded values that cause friction when developers have different local setups (different ports, credentials, database names).

**Independent Test**: Can be fully tested by creating a config file with custom values and verifying the application uses them. Delivers immediate value by allowing any developer to run the app with their setup.

**Acceptance Scenarios**:

1. **Given** a developer has a config file with custom database credentials, **When** they start the server, **Then** the server connects to the database using those credentials
2. **Given** a developer has a config file specifying port 9000, **When** they start the server, **Then** the server listens on port 9000
3. **Given** no config file exists, **When** they start the server, **Then** the server uses sensible defaults and starts successfully

---

### User Story 2 - Developer Runs Tests Against Test Database (Priority: P1)

A developer wants to run integration tests against an isolated test database (separate from development). They specify a test configuration that points to `loomio_test` database and run tests without affecting their development data.

**Why this priority**: Equal priority with Story 1 because test isolation is critical for TDD workflow. Developers need confidence that tests won't corrupt development data.

**Independent Test**: Can be fully tested by running tests with test config and verifying they use the test database. Delivers immediate value by enabling safe, isolated test runs.

**Acceptance Scenarios**:

1. **Given** a test config file pointing to `loomio_test`, **When** tests run with that config, **Then** all database operations occur in `loomio_test`
2. **Given** development data exists in `loomio_development`, **When** tests run with test config, **Then** development data remains unchanged
3. **Given** a test database doesn't exist, **When** running migrations with test config, **Then** migrations run against the test database

---

### User Story 3 - Developer Overrides Config via Command Line (Priority: P2)

A developer needs to quickly test with a different port or log level without modifying their config file. They pass command-line flags that override the config file values for that single run.

**Why this priority**: Convenience feature that builds on P1. Not essential for basic operation but significantly improves developer experience for ad-hoc testing.

**Independent Test**: Can be fully tested by starting server with CLI flags and verifying they take precedence over config file values.

**Acceptance Scenarios**:

1. **Given** config file specifies port 8080, **When** starting server with `--port 9000` flag, **Then** server listens on port 9000
2. **Given** config file specifies log level "info", **When** starting with `--log-level debug`, **Then** debug logs appear
3. **Given** multiple override sources exist (CLI, env, file), **When** starting server, **Then** CLI flags take highest precedence

---

### User Story 4 - Developer Uses Environment Variables (Priority: P2)

A developer or CI system needs to configure the application via environment variables (common in containers, CI pipelines, and production deployments). Environment variables override config file values but are overridden by CLI flags.

**Why this priority**: Important for CI/CD and container deployments, but secondary to local file-based config for development workflow.

**Independent Test**: Can be fully tested by setting environment variables and verifying application uses them.

**Acceptance Scenarios**:

1. **Given** environment variable `LOOMIO_DATABASE_NAME=loomio_ci`, **When** starting server, **Then** server connects to `loomio_ci` database
2. **Given** env var and config file both set database name, **When** starting server, **Then** env var value is used

---

### User Story 5 - Developer Views Structured Logs (Priority: P3)

A developer or operations team needs structured JSON logs for debugging and log aggregation. The logging format, level, and output destination are configurable.

**Why this priority**: Production readiness feature. Local development works fine with basic logging; structured logs become important for debugging and monitoring in deployed environments.

**Independent Test**: Can be fully tested by configuring log format as JSON and verifying output is valid JSON with expected fields.

**Acceptance Scenarios**:

1. **Given** log format configured as "json", **When** server handles a request, **Then** log output is valid JSON with structured fields
2. **Given** log level configured as "warn", **When** info-level events occur, **Then** they are not logged
3. **Given** log output configured as a file path, **When** server logs events, **Then** logs appear in the specified file

---

### Edge Cases

- What happens when config file contains invalid YAML syntax? → Clear error message with file location and line number
- What happens when config file specifies an unknown setting? → Ignored (forward compatibility) with optional warning
- What happens when required database connection fails? → Server exits with clear error message, does not start in degraded mode
- What happens when specified log file path is not writable? → Falls back to stdout with warning; application continues running
- What happens when CLI flag type is wrong (e.g., `--port abc`)? → Clear error message showing expected type
- How are credentials in config files protected? → Config files containing credentials should be added to `.gitignore`; example config files (e.g., `config.example.yaml`) with placeholder values can be committed

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST load configuration from YAML files
- **FR-002**: System MUST support configuration via environment variables
- **FR-003**: System MUST support configuration via command-line flags
- **FR-004**: System MUST apply configuration in priority order: CLI flags > environment variables > config file > defaults
- **FR-005**: System MUST provide sensible defaults for all settings so the application runs without any configuration
- **FR-006**: System MUST support namespaced environment variables (`LOOMIO_DATABASE_HOST`, `LOOMIO_SERVER_PORT`, etc.)
- **FR-008**: System MUST allow all database connection settings to be configured (host, port, user, password, database name, SSL mode, connection pool settings, health check period)
- **FR-009**: System MUST allow all server settings to be configured (port, HTTP read/write/idle timeouts)
- **FR-010**: System MUST allow session settings to be configured (duration, cleanup interval)
- **FR-011**: System MUST allow logging settings to be configured (level, format, output destination)
- **FR-012**: System MUST output logs in JSON format by default
- **FR-013**: System MUST provide clear error messages when configuration is invalid
- **FR-014**: Server command MUST support all configuration as CLI flags
- **FR-015**: Migrate command MUST support database configuration via same mechanisms as server

### Key Entities

- **Configuration**: The complete set of application settings, loaded from multiple sources and merged by priority
- **Database Settings**: Connection parameters (host, port, credentials, pool settings) for the database
- **Server Settings**: HTTP server parameters (port, timeouts)
- **Session Settings**: User session parameters (duration, cleanup frequency)
- **Logging Settings**: Log output parameters (level, format, destination)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can start the application with custom database credentials in under 1 minute by creating a config file
- **SC-002**: Developers can run tests against an isolated test database without affecting development data
- **SC-003**: All previously hardcoded values (database credentials, ports, timeouts, session duration) are configurable
- **SC-004**: Application starts successfully with zero configuration (all defaults work)
- **SC-005**: Configuration errors produce clear, actionable error messages that identify the problem and location

## Clarifications

### Session 2026-02-02

- Q: How should sensitive credentials in config files be handled? → A: Config files may store credentials but should not be committed; document in `.gitignore` guidance
- Q: What happens when log file path is not writable? → A: Fall back to stdout with warning; application continues running

## Assumptions

- YAML is the preferred config file format (confirmed during brainstorming)
- JSON is the default log format (confirmed during brainstorming)
- Test database will be named `loomio_test` by convention
- Default development database remains `loomio_development`
- Config file search path: `./config.yaml`, then `./config/config.yaml`
- Health check period should be configurable (confirmed during brainstorming)
- Logging belongs in its own package, separate from config (confirmed during brainstorming)
