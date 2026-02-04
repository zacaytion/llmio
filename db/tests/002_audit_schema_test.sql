-- pgTap tests for audit schema
-- Run with: pg_prove -d loomio_test tests/pgtap/002_audit_schema_test.sql

BEGIN;
SELECT plan(20);

-- Test schema exists
SELECT has_schema('audit', 'audit schema should exist');

-- Test audit.operation enum exists
SELECT has_type('audit', 'operation', 'audit.operation enum should exist');

-- Test record_version table exists
SELECT has_table('audit', 'record_version', 'audit.record_version table should exist');

-- Test columns exist with correct types
SELECT has_column('audit', 'record_version', 'id', 'record_version should have id column');
SELECT col_type_is('audit', 'record_version', 'id', 'bigint', 'id should be bigint');

SELECT has_column('audit', 'record_version', 'record_id', 'record_version should have record_id column');
SELECT col_type_is('audit', 'record_version', 'record_id', 'text', 'record_id should be text');

SELECT has_column('audit', 'record_version', 'old_record_id', 'record_version should have old_record_id column');

SELECT has_column('audit', 'record_version', 'op', 'record_version should have op column');
SELECT col_type_is('audit', 'record_version', 'op', 'audit.operation', 'op should be audit.operation enum');

SELECT has_column('audit', 'record_version', 'ts', 'record_version should have ts column');
SELECT has_column('audit', 'record_version', 'xact_id', 'record_version should have xact_id column');
SELECT has_column('audit', 'record_version', 'table_oid', 'record_version should have table_oid column');
SELECT has_column('audit', 'record_version', 'table_schema', 'record_version should have table_schema column');
SELECT has_column('audit', 'record_version', 'table_name', 'record_version should have table_name column');
SELECT has_column('audit', 'record_version', 'record', 'record_version should have record column');
SELECT has_column('audit', 'record_version', 'old_record', 'record_version should have old_record column');
SELECT has_column('audit', 'record_version', 'actor_id', 'record_version should have actor_id column');

-- Test trigger function exists
SELECT has_function(
    'audit',
    'insert_update_delete_trigger',
    'audit.insert_update_delete_trigger function should exist'
);

-- Test BRIN index exists on ts column
SELECT has_index(
    'audit',
    'record_version',
    'record_version_ts_brin',
    'BRIN index on ts should exist'
);

SELECT * FROM finish();
ROLLBACK;
