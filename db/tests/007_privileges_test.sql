-- pgTap tests for role-based privilege separation
-- Verifies loomio_app and loomio_migration roles have correct privileges
-- Run with: pg_prove -d loomio_test db/tests/007_privileges_test.sql

BEGIN;
SELECT plan(20);

-- =====================================================
-- Role Existence Tests
-- =====================================================

SELECT ok(
    EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'loomio_app'),
    'loomio_app role should exist'
);

SELECT ok(
    EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'loomio_migration'),
    'loomio_migration role should exist'
);

-- =====================================================
-- loomio_app Schema Usage Tests
-- =====================================================

-- Test loomio_app can use data schema
SELECT ok(
    has_schema_privilege('loomio_app', 'data', 'USAGE'),
    'loomio_app should have USAGE on data schema'
);

-- Test loomio_app can use private schema
SELECT ok(
    has_schema_privilege('loomio_app', 'private', 'USAGE'),
    'loomio_app should have USAGE on private schema'
);

-- Test loomio_app can use audit schema
SELECT ok(
    has_schema_privilege('loomio_app', 'audit', 'USAGE'),
    'loomio_app should have USAGE on audit schema'
);

-- Test loomio_app CANNOT create in data schema
SELECT ok(
    NOT has_schema_privilege('loomio_app', 'data', 'CREATE'),
    'loomio_app should NOT have CREATE on data schema'
);

-- =====================================================
-- loomio_app DML Tests (on data schema)
-- =====================================================

-- Test loomio_app has SELECT on data.users
SELECT ok(
    has_table_privilege('loomio_app', 'data.users', 'SELECT'),
    'loomio_app should have SELECT on data.users'
);

-- Test loomio_app has INSERT on data.users
SELECT ok(
    has_table_privilege('loomio_app', 'data.users', 'INSERT'),
    'loomio_app should have INSERT on data.users'
);

-- Test loomio_app has UPDATE on data.users
SELECT ok(
    has_table_privilege('loomio_app', 'data.users', 'UPDATE'),
    'loomio_app should have UPDATE on data.users'
);

-- Test loomio_app has DELETE on data.users
SELECT ok(
    has_table_privilege('loomio_app', 'data.users', 'DELETE'),
    'loomio_app should have DELETE on data.users'
);

-- =====================================================
-- loomio_app Audit Read-Only Tests
-- =====================================================

-- Test loomio_app has SELECT on audit.record_version
SELECT ok(
    has_table_privilege('loomio_app', 'audit.record_version', 'SELECT'),
    'loomio_app should have SELECT on audit.record_version'
);

-- Test loomio_app CANNOT INSERT to audit
SELECT ok(
    NOT has_table_privilege('loomio_app', 'audit.record_version', 'INSERT'),
    'loomio_app should NOT have INSERT on audit.record_version'
);

-- =====================================================
-- loomio_migration Schema Privilege Tests
-- =====================================================

-- Test loomio_migration has CREATE on data schema
SELECT ok(
    has_schema_privilege('loomio_migration', 'data', 'CREATE'),
    'loomio_migration should have CREATE on data schema'
);

-- Test loomio_migration has CREATE on private schema
SELECT ok(
    has_schema_privilege('loomio_migration', 'private', 'CREATE'),
    'loomio_migration should have CREATE on private schema'
);

-- Test loomio_migration has CREATE on audit schema
SELECT ok(
    has_schema_privilege('loomio_migration', 'audit', 'CREATE'),
    'loomio_migration should have CREATE on audit schema'
);

-- =====================================================
-- Function Execution Privileges
-- =====================================================

-- Test loomio_app can execute private.set_updated_at (triggers need this)
SELECT ok(
    has_function_privilege('loomio_app', 'private.set_updated_at()', 'EXECUTE'),
    'loomio_app should have EXECUTE on private.set_updated_at'
);

-- Test loomio_app can execute private.protect_last_admin (triggers need this)
SELECT ok(
    has_function_privilege('loomio_app', 'private.protect_last_admin()', 'EXECUTE'),
    'loomio_app should have EXECUTE on private.protect_last_admin'
);

-- Test loomio_app can execute audit trigger (triggers need this)
SELECT ok(
    has_function_privilege('loomio_app', 'audit.insert_update_delete_trigger()', 'EXECUTE'),
    'loomio_app should have EXECUTE on audit.insert_update_delete_trigger'
);

-- =====================================================
-- Sequence Privileges (needed for SERIAL columns)
-- =====================================================

-- Test loomio_app has USAGE on users_id_seq
SELECT ok(
    has_sequence_privilege('loomio_app', 'data.users_id_seq', 'USAGE'),
    'loomio_app should have USAGE on data.users_id_seq'
);

-- Test loomio_migration has ALL on users_id_seq
SELECT ok(
    has_sequence_privilege('loomio_migration', 'data.users_id_seq', 'USAGE'),
    'loomio_migration should have USAGE on data.users_id_seq'
);

SELECT * FROM finish();
ROLLBACK;
