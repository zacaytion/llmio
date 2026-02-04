-- pgTap tests for groups table schema and constraints
-- Run with: pg_prove -d loomio_test tests/pgtap/003_groups_test.sql

BEGIN;
SELECT plan(35);

-- Test table exists
SELECT has_table('groups', 'groups table should exist');

-- Test columns exist
SELECT has_column('groups', 'id', 'groups should have id column');
SELECT has_column('groups', 'name', 'groups should have name column');
SELECT has_column('groups', 'handle', 'groups should have handle column');
SELECT has_column('groups', 'description', 'groups should have description column');
SELECT has_column('groups', 'parent_id', 'groups should have parent_id column');
SELECT has_column('groups', 'created_by_id', 'groups should have created_by_id column');
SELECT has_column('groups', 'archived_at', 'groups should have archived_at column');
SELECT has_column('groups', 'created_at', 'groups should have created_at column');
SELECT has_column('groups', 'updated_at', 'groups should have updated_at column');

-- Test permission flag columns exist
SELECT has_column('groups', 'members_can_add_members', 'groups should have members_can_add_members column');
SELECT has_column('groups', 'members_can_add_guests', 'groups should have members_can_add_guests column');
SELECT has_column('groups', 'members_can_start_discussions', 'groups should have members_can_start_discussions column');
SELECT has_column('groups', 'members_can_raise_motions', 'groups should have members_can_raise_motions column');
SELECT has_column('groups', 'members_can_edit_discussions', 'groups should have members_can_edit_discussions column');
SELECT has_column('groups', 'members_can_edit_comments', 'groups should have members_can_edit_comments column');
SELECT has_column('groups', 'members_can_delete_comments', 'groups should have members_can_delete_comments column');
SELECT has_column('groups', 'members_can_announce', 'groups should have members_can_announce column');
SELECT has_column('groups', 'members_can_create_subgroups', 'groups should have members_can_create_subgroups column');
SELECT has_column('groups', 'admins_can_edit_user_content', 'groups should have admins_can_edit_user_content column');
SELECT has_column('groups', 'parent_members_can_see_discussions', 'groups should have parent_members_can_see_discussions column');

-- Test column types
SELECT col_type_is('groups', 'id', 'bigint', 'id should be bigint');
SELECT col_type_is('groups', 'handle', 'citext', 'handle should be citext for case-insensitivity');

-- Test foreign key to users
SELECT col_is_fk('groups', 'created_by_id', 'created_by_id should be a foreign key');

-- Test self-referential FK for parent_id
SELECT col_is_fk('groups', 'parent_id', 'parent_id should be a foreign key');

-- Test unique constraint on handle
SELECT col_is_unique('groups', 'handle', 'handle should be unique');

-- Test name length constraint
SELECT throws_ok(
    $$INSERT INTO groups (name, handle, created_by_id) VALUES ('', 'test-handle', 1)$$,
    '23514',  -- check_violation
    NULL,
    'Empty name should be rejected by constraint'
);

-- Test handle format constraint (must match pattern)
SELECT throws_ok(
    $$INSERT INTO groups (name, handle, created_by_id) VALUES ('Test', 'ab', 1)$$,
    '23514',  -- check_violation
    NULL,
    'Handle shorter than 3 chars should be rejected'
);

SELECT throws_ok(
    $$INSERT INTO groups (name, handle, created_by_id) VALUES ('Test', '-invalid', 1)$$,
    '23514',  -- check_violation
    NULL,
    'Handle starting with hyphen should be rejected'
);

SELECT throws_ok(
    $$INSERT INTO groups (name, handle, created_by_id) VALUES ('Test', 'invalid-', 1)$$,
    '23514',  -- check_violation
    NULL,
    'Handle ending with hyphen should be rejected'
);

-- Test updated_at trigger exists
SELECT trigger_is(
    'groups',
    'groups_updated_at',
    'update_updated_at',
    'groups_updated_at trigger should call update_updated_at function'
);

-- Test audit trigger exists
SELECT trigger_is(
    'groups',
    'groups_audit',
    'audit.insert_update_delete_trigger',
    'groups_audit trigger should call audit.insert_update_delete_trigger function'
);

-- Test indexes exist
SELECT has_index('groups', 'groups_handle_key', 'unique index on handle should exist');
SELECT has_index('groups', 'groups_parent_id_idx', 'index on parent_id should exist');
SELECT has_index('groups', 'groups_created_by_id_idx', 'index on created_by_id should exist');

SELECT * FROM finish();
ROLLBACK;
