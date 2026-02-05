-- +goose Up
-- +goose StatementBegin
-- +goose ENVSUB ON

-- Three-role privilege separation for security
--
-- postgres (superuser) - Initial setup and emergency operations only
-- loomio_migration    - DDL privileges for schema changes (used by cmd/migrate)
-- loomio_app          - DML privileges for runtime queries (used by cmd/server)
--
-- This separation follows the principle of least privilege:
-- - The application cannot modify schema, reducing attack surface
-- - Migration tool has limited scope, reducing blast radius of bugs

-- ============================================================
-- Migration Role: Schema changes only
-- ============================================================

-- Role passwords use goose environment variable substitution.
-- Set via environment variables for production:
--   LLMIO_PG_PASS_MIGRATION - Password for loomio_migration role
--   LLMIO_PG_PASS_APP       - Password for loomio_app role
-- Defaults are placeholder values for development only.

DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'loomio_migration') THEN
        CREATE ROLE loomio_migration WITH LOGIN PASSWORD '${LLMIO_PG_PASS_MIGRATION:-change_me_migration}';
    END IF;
END $$;

-- Grant schema usage and creation
GRANT USAGE ON SCHEMA data TO loomio_migration;
GRANT USAGE ON SCHEMA private TO loomio_migration;
GRANT USAGE ON SCHEMA audit TO loomio_migration;
GRANT CREATE ON SCHEMA data TO loomio_migration;
GRANT CREATE ON SCHEMA private TO loomio_migration;
GRANT CREATE ON SCHEMA audit TO loomio_migration;

-- Grant full DDL on existing tables
GRANT ALL ON ALL TABLES IN SCHEMA data TO loomio_migration;
GRANT ALL ON ALL TABLES IN SCHEMA private TO loomio_migration;
GRANT ALL ON ALL TABLES IN SCHEMA audit TO loomio_migration;

-- Grant sequence access (needed for SERIAL columns during migrations)
GRANT ALL ON ALL SEQUENCES IN SCHEMA data TO loomio_migration;
GRANT ALL ON ALL SEQUENCES IN SCHEMA audit TO loomio_migration;

-- Grant function execution (needed for creating/modifying functions)
GRANT ALL ON ALL FUNCTIONS IN SCHEMA private TO loomio_migration;
GRANT ALL ON ALL FUNCTIONS IN SCHEMA audit TO loomio_migration;

-- Future objects get same privileges
ALTER DEFAULT PRIVILEGES IN SCHEMA data
    GRANT ALL ON TABLES TO loomio_migration;
ALTER DEFAULT PRIVILEGES IN SCHEMA data
    GRANT ALL ON SEQUENCES TO loomio_migration;
ALTER DEFAULT PRIVILEGES IN SCHEMA private
    GRANT ALL ON FUNCTIONS TO loomio_migration;
ALTER DEFAULT PRIVILEGES IN SCHEMA audit
    GRANT ALL ON TABLES TO loomio_migration;

-- ============================================================
-- Application Role: Runtime DML only
-- ============================================================

DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'loomio_app') THEN
        CREATE ROLE loomio_app WITH LOGIN PASSWORD '${LLMIO_PG_PASS_APP:-change_me_app}';
    END IF;
END $$;

-- +goose ENVSUB OFF

-- Grant schema usage (not CREATE - cannot add new objects)
GRANT USAGE ON SCHEMA data TO loomio_app;
GRANT USAGE ON SCHEMA private TO loomio_app;
GRANT USAGE ON SCHEMA audit TO loomio_app;

-- Grant DML on data tables (SELECT, INSERT, UPDATE, DELETE)
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA data TO loomio_app;

-- Grant sequence usage (needed for SERIAL columns on INSERT)
GRANT USAGE ON ALL SEQUENCES IN SCHEMA data TO loomio_app;

-- Grant function execution in private schema (triggers call these)
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA private TO loomio_app;

-- Grant read-only on audit schema (can read audit trail, not modify)
GRANT SELECT ON ALL TABLES IN SCHEMA audit TO loomio_app;

-- Grant execute on audit trigger function (triggers call this)
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA audit TO loomio_app;

-- Future tables/sequences get same privileges
ALTER DEFAULT PRIVILEGES IN SCHEMA data
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO loomio_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA data
    GRANT USAGE ON SEQUENCES TO loomio_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA private
    GRANT EXECUTE ON FUNCTIONS TO loomio_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA audit
    GRANT SELECT ON TABLES TO loomio_app;

-- Comments for documentation
COMMENT ON ROLE loomio_migration IS 'Migration role with DDL privileges for schema changes';
COMMENT ON ROLE loomio_app IS 'Application role with DML privileges for runtime queries';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Revoke privileges before dropping roles
REVOKE ALL ON ALL TABLES IN SCHEMA data FROM loomio_app;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA data FROM loomio_app;
REVOKE ALL ON ALL FUNCTIONS IN SCHEMA private FROM loomio_app;
REVOKE ALL ON ALL TABLES IN SCHEMA audit FROM loomio_app;
REVOKE ALL ON ALL FUNCTIONS IN SCHEMA audit FROM loomio_app;
REVOKE USAGE ON SCHEMA data, private, audit FROM loomio_app;

REVOKE ALL ON ALL TABLES IN SCHEMA data FROM loomio_migration;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA data FROM loomio_migration;
REVOKE ALL ON ALL TABLES IN SCHEMA private FROM loomio_migration;
REVOKE ALL ON ALL FUNCTIONS IN SCHEMA private FROM loomio_migration;
REVOKE ALL ON ALL TABLES IN SCHEMA audit FROM loomio_migration;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA audit FROM loomio_migration;
REVOKE ALL ON ALL FUNCTIONS IN SCHEMA audit FROM loomio_migration;
REVOKE CREATE, USAGE ON SCHEMA data, private, audit FROM loomio_migration;

-- Reset default privileges
ALTER DEFAULT PRIVILEGES IN SCHEMA data
    REVOKE ALL ON TABLES FROM loomio_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA data
    REVOKE ALL ON SEQUENCES FROM loomio_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA private
    REVOKE ALL ON FUNCTIONS FROM loomio_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA audit
    REVOKE ALL ON TABLES FROM loomio_app;

ALTER DEFAULT PRIVILEGES IN SCHEMA data
    REVOKE ALL ON TABLES FROM loomio_migration;
ALTER DEFAULT PRIVILEGES IN SCHEMA data
    REVOKE ALL ON SEQUENCES FROM loomio_migration;
ALTER DEFAULT PRIVILEGES IN SCHEMA private
    REVOKE ALL ON FUNCTIONS FROM loomio_migration;
ALTER DEFAULT PRIVILEGES IN SCHEMA audit
    REVOKE ALL ON TABLES FROM loomio_migration;

-- Drop roles
DROP ROLE IF EXISTS loomio_app;
DROP ROLE IF EXISTS loomio_migration;

-- +goose StatementEnd
