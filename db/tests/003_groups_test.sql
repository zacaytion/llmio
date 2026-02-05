-- pgTap tests for groups table schema and constraints
-- Run with: pg_prove -d loomio_test db/tests/003_groups_test.sql

BEGIN;
SELECT plan(46);

-- =====================================================
-- Schema and Table Structure Tests
-- =====================================================

-- Test table exists in data schema
SELECT has_table('data', 'groups', 'groups table should exist in data schema');

-- Test columns exist
SELECT has_column('data', 'groups', 'id', 'groups should have id column');
SELECT has_column('data', 'groups', 'name', 'groups should have name column');
SELECT has_column('data', 'groups', 'handle', 'groups should have handle column');
SELECT has_column('data', 'groups', 'description', 'groups should have description column');
SELECT has_column('data', 'groups', 'parent_id', 'groups should have parent_id column');
SELECT has_column('data', 'groups', 'created_by_id', 'groups should have created_by_id column');
SELECT has_column('data', 'groups', 'archived_at', 'groups should have archived_at column');
SELECT has_column('data', 'groups', 'created_at', 'groups should have created_at column');
SELECT has_column('data', 'groups', 'updated_at', 'groups should have updated_at column');

-- Test permission flag columns exist (all 11 flags)
SELECT has_column('data', 'groups', 'members_can_add_members', 'groups should have members_can_add_members column');
SELECT has_column('data', 'groups', 'members_can_add_guests', 'groups should have members_can_add_guests column');
SELECT has_column('data', 'groups', 'members_can_start_discussions', 'groups should have members_can_start_discussions column');
SELECT has_column('data', 'groups', 'members_can_raise_motions', 'groups should have members_can_raise_motions column');
SELECT has_column('data', 'groups', 'members_can_edit_discussions', 'groups should have members_can_edit_discussions column');
SELECT has_column('data', 'groups', 'members_can_edit_comments', 'groups should have members_can_edit_comments column');
SELECT has_column('data', 'groups', 'members_can_delete_comments', 'groups should have members_can_delete_comments column');
SELECT has_column('data', 'groups', 'members_can_announce', 'groups should have members_can_announce column');
SELECT has_column('data', 'groups', 'members_can_create_subgroups', 'groups should have members_can_create_subgroups column');
SELECT has_column('data', 'groups', 'admins_can_edit_user_content', 'groups should have admins_can_edit_user_content column');
SELECT has_column('data', 'groups', 'parent_members_can_see_discussions', 'groups should have parent_members_can_see_discussions column');

-- Test column types
SELECT col_type_is('data', 'groups', 'id', 'bigint', 'id should be bigint');
SELECT col_type_is('data', 'groups', 'handle', 'citext', 'handle should be citext for case-insensitivity');

-- Test foreign key to users
SELECT col_is_fk('data', 'groups', 'created_by_id', 'created_by_id should be a foreign key');

-- Test self-referential FK for parent_id
SELECT col_is_fk('data', 'groups', 'parent_id', 'parent_id should be a foreign key');

-- Test unique index on handle (using unique index, not constraint, for CITEXT)
SELECT is(
    (SELECT indisunique FROM pg_index WHERE indexrelid = 'data.groups_handle_key'::regclass),
    true,
    'groups_handle_key should be unique'
);

-- =====================================================
-- Permission Flag Default Values Tests
-- =====================================================

-- Create test user for groups tests
INSERT INTO data.users (email, name, username, password_hash, key)
VALUES ('grouptest@example.com', 'Group Test User', 'grouptest', 'hash123', 'grouptest-key');

-- Create a minimal group to test defaults
INSERT INTO data.groups (name, handle, created_by_id)
VALUES ('Default Test Group', 'default-test-group', (SELECT id FROM data.users WHERE email = 'grouptest@example.com'));

-- Test permission flag defaults (TRUE defaults)
SELECT is(
    (SELECT members_can_add_members FROM data.groups WHERE handle = 'default-test-group'),
    true, 'members_can_add_members should default to TRUE'
);
SELECT is(
    (SELECT members_can_add_guests FROM data.groups WHERE handle = 'default-test-group'),
    true, 'members_can_add_guests should default to TRUE'
);
SELECT is(
    (SELECT members_can_start_discussions FROM data.groups WHERE handle = 'default-test-group'),
    true, 'members_can_start_discussions should default to TRUE'
);
SELECT is(
    (SELECT members_can_raise_motions FROM data.groups WHERE handle = 'default-test-group'),
    true, 'members_can_raise_motions should default to TRUE'
);
SELECT is(
    (SELECT members_can_edit_comments FROM data.groups WHERE handle = 'default-test-group'),
    true, 'members_can_edit_comments should default to TRUE'
);
SELECT is(
    (SELECT members_can_delete_comments FROM data.groups WHERE handle = 'default-test-group'),
    true, 'members_can_delete_comments should default to TRUE'
);

-- Test permission flag defaults (FALSE defaults)
SELECT is(
    (SELECT members_can_edit_discussions FROM data.groups WHERE handle = 'default-test-group'),
    false, 'members_can_edit_discussions should default to FALSE'
);
SELECT is(
    (SELECT members_can_announce FROM data.groups WHERE handle = 'default-test-group'),
    false, 'members_can_announce should default to FALSE'
);
SELECT is(
    (SELECT members_can_create_subgroups FROM data.groups WHERE handle = 'default-test-group'),
    false, 'members_can_create_subgroups should default to FALSE'
);
SELECT is(
    (SELECT admins_can_edit_user_content FROM data.groups WHERE handle = 'default-test-group'),
    false, 'admins_can_edit_user_content should default to FALSE'
);
SELECT is(
    (SELECT parent_members_can_see_discussions FROM data.groups WHERE handle = 'default-test-group'),
    false, 'parent_members_can_see_discussions should default to FALSE'
);

-- =====================================================
-- Constraint Tests
-- =====================================================

-- Test name length constraint (empty name)
SELECT throws_ok(
    $$INSERT INTO data.groups (name, handle, created_by_id)
      VALUES ('', 'test-handle', (SELECT id FROM data.users WHERE email = 'grouptest@example.com'))$$,
    '23514',  -- check_violation
    NULL,
    'Empty name should be rejected by constraint'
);

-- Test handle format constraint (must match pattern)
SELECT throws_ok(
    $$INSERT INTO data.groups (name, handle, created_by_id)
      VALUES ('Test', 'ab', (SELECT id FROM data.users WHERE email = 'grouptest@example.com'))$$,
    '23514',  -- check_violation
    NULL,
    'Handle shorter than 3 chars should be rejected'
);

SELECT throws_ok(
    $$INSERT INTO data.groups (name, handle, created_by_id)
      VALUES ('Test', '-invalid', (SELECT id FROM data.users WHERE email = 'grouptest@example.com'))$$,
    '23514',  -- check_violation
    NULL,
    'Handle starting with hyphen should be rejected'
);

SELECT throws_ok(
    $$INSERT INTO data.groups (name, handle, created_by_id)
      VALUES ('Test', 'invalid-', (SELECT id FROM data.users WHERE email = 'grouptest@example.com'))$$,
    '23514',  -- check_violation
    NULL,
    'Handle ending with hyphen should be rejected'
);

-- =====================================================
-- Trigger Tests
-- =====================================================

-- Test updated_at trigger exists
SELECT trigger_is(
    'data',
    'groups',
    'groups_updated_at',
    'private',
    'set_updated_at',
    'groups_updated_at trigger should call private.set_updated_at function'
);

-- Test audit trigger exists
SELECT trigger_is(
    'data',
    'groups',
    'groups_audit',
    'audit',
    'insert_update_delete_trigger',
    'groups_audit trigger should call audit.insert_update_delete_trigger function'
);

-- =====================================================
-- Index Tests
-- =====================================================

SELECT has_index('data', 'groups', 'groups_handle_key', 'unique index on handle should exist');
SELECT has_index('data', 'groups', 'groups_parent_id_idx', 'index on parent_id should exist');
SELECT has_index('data', 'groups', 'groups_created_by_id_idx', 'index on created_by_id should exist');

SELECT * FROM finish();
ROLLBACK;
