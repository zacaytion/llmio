-- +goose Up
-- +goose StatementBegin

-- Audit schema for tracking all changes to groups and memberships
-- Based on supa_audit pattern with xact_id for transaction correlation

CREATE SCHEMA IF NOT EXISTS audit;

-- Operation types for audit records
CREATE TYPE audit.operation AS ENUM ('INSERT', 'UPDATE', 'DELETE', 'TRUNCATE', 'SNAPSHOT');

-- Audit log table
-- Stores JSONB snapshots of old/new record state for all tracked tables
CREATE TABLE audit.record_version (
    id              BIGSERIAL PRIMARY KEY,
    record_id       TEXT,                    -- Primary key of the record (stored as text for flexibility with BIGSERIAL PKs)
    old_record_id   TEXT,                    -- Previous record_id (for UPDATE/DELETE)
    op              audit.operation NOT NULL,
    ts              TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    xact_id         BIGINT NOT NULL DEFAULT txid_current(), -- Transaction correlation
    table_oid       OID NOT NULL,
    table_schema    NAME NOT NULL,
    table_name      NAME NOT NULL,
    record          JSONB,                   -- New state (INSERT/UPDATE)
    old_record      JSONB,                   -- Old state (UPDATE/DELETE)
    actor_id        BIGINT,                  -- User who made the change (from session var)

    -- Constraints to ensure data integrity
    CONSTRAINT audit_record_version_record_id_check
        CHECK (COALESCE(record_id, old_record_id) IS NOT NULL OR op IN ('TRUNCATE', 'SNAPSHOT')),
    CONSTRAINT audit_record_version_record_check
        CHECK (op IN ('INSERT', 'UPDATE', 'SNAPSHOT') = (record IS NOT NULL)),
    CONSTRAINT audit_record_version_old_record_check
        CHECK (op IN ('UPDATE', 'DELETE') = (old_record IS NOT NULL))
);

-- Indexes for efficient querying
-- BRIN index on timestamp for time-range queries (99% smaller than B-tree for append-only tables)
CREATE INDEX record_version_ts_brin ON audit.record_version USING BRIN(ts);
-- B-tree indexes for point lookups
CREATE INDEX record_version_record_id ON audit.record_version(record_id) WHERE record_id IS NOT NULL;
CREATE INDEX record_version_table_oid ON audit.record_version(table_oid);
CREATE INDEX record_version_xact_id ON audit.record_version(xact_id);
CREATE INDEX record_version_actor_id ON audit.record_version(actor_id) WHERE actor_id IS NOT NULL;

-- Generic trigger function for INSERT/UPDATE/DELETE auditing
-- Captures actor_id from session variable app.current_user_id
CREATE OR REPLACE FUNCTION audit.insert_update_delete_trigger()
RETURNS TRIGGER
SECURITY DEFINER
LANGUAGE plpgsql
AS $$
DECLARE
    record_jsonb JSONB;
    old_record_jsonb JSONB;
    v_actor_id BIGINT;
    v_record_id TEXT;
    v_old_record_id TEXT;
BEGIN
    -- Get actor from session variable (NULL if not set)
    -- The 'true' parameter makes current_setting return NULL instead of error if not set
    v_actor_id := NULLIF(current_setting('app.current_user_id', true), '')::BIGINT;

    -- Convert records to JSONB
    IF TG_OP != 'DELETE' THEN
        record_jsonb := to_jsonb(NEW);
        v_record_id := NEW.id::TEXT;
    END IF;

    IF TG_OP != 'INSERT' THEN
        old_record_jsonb := to_jsonb(OLD);
        v_old_record_id := OLD.id::TEXT;
    END IF;

    INSERT INTO audit.record_version (
        record_id,
        old_record_id,
        op,
        table_oid,
        table_schema,
        table_name,
        record,
        old_record,
        actor_id
    ) VALUES (
        v_record_id,
        v_old_record_id,
        TG_OP::audit.operation,
        TG_RELID,
        TG_TABLE_SCHEMA,
        TG_TABLE_NAME,
        record_jsonb,
        old_record_jsonb,
        v_actor_id
    );

    RETURN COALESCE(NEW, OLD);
END;
$$;

COMMENT ON SCHEMA audit IS 'Audit logging schema for tracking changes to groups and memberships';
COMMENT ON TABLE audit.record_version IS 'Immutable audit log storing JSONB snapshots of record changes';
COMMENT ON COLUMN audit.record_version.xact_id IS 'Transaction ID for correlating changes in the same transaction';
COMMENT ON COLUMN audit.record_version.actor_id IS 'User ID from app.current_user_id session variable';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP FUNCTION IF EXISTS audit.insert_update_delete_trigger() CASCADE;
DROP TABLE IF EXISTS audit.record_version;
DROP TYPE IF EXISTS audit.operation;
DROP SCHEMA IF EXISTS audit CASCADE;

-- +goose StatementEnd
