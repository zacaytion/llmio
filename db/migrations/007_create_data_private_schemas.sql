-- +goose Up
-- +goose StatementBegin

-- Create data schema for application data tables
-- This separates application data from PostgreSQL system schemas (pg_*, information_schema)
-- and from our internal infrastructure (audit, private)
CREATE SCHEMA IF NOT EXISTS data;
COMMENT ON SCHEMA data IS 'Application data tables (users, groups, memberships, discussions, etc.)';

-- Create private schema for internal functions and triggers
-- These are implementation details that should not be directly called by application code
CREATE SCHEMA IF NOT EXISTS private;
COMMENT ON SCHEMA private IS 'Internal trigger functions and utilities (not for direct application use)';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Note: CASCADE will drop any objects in these schemas
-- This is intentional for clean rollback but requires careful ordering
DROP SCHEMA IF EXISTS private CASCADE;
DROP SCHEMA IF EXISTS data CASCADE;

-- +goose StatementEnd
