-- pgTap tests for schema structure validation
-- Verifies tables, functions, and types are in their correct schemas
-- Run with: pg_prove -d loomio_test db/tests/006_schema_structure_test.sql

BEGIN;
SELECT plan(20);

-- =====================================================
-- Schema Existence Tests
-- =====================================================

SELECT has_schema('data', 'data schema should exist');
SELECT has_schema('private', 'private schema should exist');
SELECT has_schema('audit', 'audit schema should exist');

-- =====================================================
-- Tables in data schema
-- =====================================================

SELECT has_table('data', 'users', 'users table should be in data schema');
SELECT has_table('data', 'groups', 'groups table should be in data schema');
SELECT has_table('data', 'memberships', 'memberships table should be in data schema');

-- Verify tables are NOT in public schema
SELECT ok(
    NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users'),
    'users table should NOT be in public schema'
);
SELECT ok(
    NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'groups'),
    'groups table should NOT be in public schema'
);
SELECT ok(
    NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'memberships'),
    'memberships table should NOT be in public schema'
);

-- =====================================================
-- Functions in private schema
-- =====================================================

SELECT has_function('private', 'set_updated_at', 'set_updated_at should be in private schema');
SELECT has_function('private', 'protect_last_admin', 'protect_last_admin should be in private schema');

-- Verify functions are NOT in public schema
SELECT ok(
    NOT EXISTS (
        SELECT 1 FROM pg_proc p
        JOIN pg_namespace n ON p.pronamespace = n.oid
        WHERE n.nspname = 'public' AND p.proname = 'update_updated_at'
    ),
    'update_updated_at should NOT be in public schema'
);
SELECT ok(
    NOT EXISTS (
        SELECT 1 FROM pg_proc p
        JOIN pg_namespace n ON p.pronamespace = n.oid
        WHERE n.nspname = 'public' AND p.proname = 'prevent_last_admin_removal'
    ),
    'prevent_last_admin_removal should NOT be in public schema'
);

-- =====================================================
-- Audit objects in audit schema
-- =====================================================

SELECT has_schema('audit', 'audit schema should exist');
SELECT has_table('audit', 'record_version', 'record_version table should be in audit schema');
SELECT has_type('audit', 'operation', 'operation type should be in audit schema');
SELECT has_function('audit', 'insert_update_delete_trigger', 'audit trigger function should be in audit schema');

-- =====================================================
-- Extension in correct location
-- =====================================================

-- citext extension should be available
SELECT ok(
    EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'citext'),
    'citext extension should be installed'
);

-- =====================================================
-- No stale objects in public schema
-- =====================================================

-- Check that no application tables are in public schema
-- (only system tables and extensions should be there)
SELECT ok(
    (SELECT COUNT(*)
     FROM information_schema.tables
     WHERE table_schema = 'public'
       AND table_type = 'BASE TABLE'
       AND table_name NOT IN ('goose_db_version')  -- Goose migration tracking table is OK
    ) = 0,
    'No application tables should remain in public schema'
);

-- Check that no application functions remain in public schema
SELECT ok(
    (SELECT COUNT(*)
     FROM pg_proc p
     JOIN pg_namespace n ON p.pronamespace = n.oid
     WHERE n.nspname = 'public'
       AND p.proname IN ('update_updated_at', 'prevent_last_admin_removal', 'set_updated_at', 'protect_last_admin')
    ) = 0,
    'No application trigger functions should remain in public schema'
);

SELECT * FROM finish();
ROLLBACK;
