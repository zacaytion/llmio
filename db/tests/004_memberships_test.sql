-- pgTap tests for memberships table schema and last-admin protection trigger
-- Run with: pg_prove -d loomio_test tests/pgtap/004_memberships_test.sql

BEGIN;
SELECT plan(27);

-- Test table exists
SELECT has_table('memberships', 'memberships table should exist');

-- Test columns exist
SELECT has_column('memberships', 'id', 'memberships should have id column');
SELECT has_column('memberships', 'group_id', 'memberships should have group_id column');
SELECT has_column('memberships', 'user_id', 'memberships should have user_id column');
SELECT has_column('memberships', 'role', 'memberships should have role column');
SELECT has_column('memberships', 'inviter_id', 'memberships should have inviter_id column');
SELECT has_column('memberships', 'accepted_at', 'memberships should have accepted_at column');
SELECT has_column('memberships', 'created_at', 'memberships should have created_at column');
SELECT has_column('memberships', 'updated_at', 'memberships should have updated_at column');

-- Test column types
SELECT col_type_is('memberships', 'id', 'bigint', 'id should be bigint');
SELECT col_type_is('memberships', 'role', 'text', 'role should be text');

-- Test foreign keys
SELECT col_is_fk('memberships', 'group_id', 'group_id should be a foreign key');
SELECT col_is_fk('memberships', 'user_id', 'user_id should be a foreign key');
SELECT col_is_fk('memberships', 'inviter_id', 'inviter_id should be a foreign key');

-- Test unique constraint on (group_id, user_id)
-- UNIQUE constraint creates an implicit index with the constraint name
SELECT has_index('memberships', 'memberships_unique_user_group', 'unique constraint on group_id + user_id should exist');

-- Test role constraint (only 'admin' or 'member')
SELECT throws_ok(
    $$INSERT INTO memberships (group_id, user_id, role, inviter_id) VALUES (1, 1, 'invalid_role', 1)$$,
    '23514',  -- check_violation
    NULL,
    'Invalid role should be rejected'
);

-- Test triggers exist
SELECT trigger_is(
    'memberships',
    'memberships_updated_at',
    'update_updated_at',
    'memberships_updated_at trigger should exist'
);

SELECT trigger_is(
    'memberships',
    'memberships_last_admin_protection',
    'prevent_last_admin_removal',
    'memberships_last_admin_protection trigger should exist'
);

-- Note: pgTap reports function name without schema prefix
SELECT trigger_is(
    'memberships',
    'memberships_audit',
    'insert_update_delete_trigger',
    'memberships_audit trigger should exist'
);

-- Test indexes exist
SELECT has_index('memberships', 'memberships_user_id_idx', 'index on user_id should exist');
SELECT has_index('memberships', 'memberships_group_id_idx', 'index on group_id should exist');
SELECT has_index('memberships', 'memberships_inviter_id_idx', 'index on inviter_id should exist');

-- =====================================================
-- Last-admin protection trigger tests
-- =====================================================

-- Create test data for last-admin protection tests
-- Note: We need real users and groups for FK constraints

-- Create test users
INSERT INTO users (email, name, username, password_hash, key)
VALUES
    ('admin1@test.com', 'Admin One', 'admin-one', 'hash1', 'key1'),
    ('admin2@test.com', 'Admin Two', 'admin-two', 'hash2', 'key2'),
    ('member1@test.com', 'Member One', 'member-one', 'hash3', 'key3');

-- Create test group
INSERT INTO groups (name, handle, created_by_id)
VALUES ('Test Group', 'test-group', (SELECT id FROM users WHERE email = 'admin1@test.com'));

-- Create admin membership (the group creator)
INSERT INTO memberships (group_id, user_id, role, inviter_id, accepted_at)
VALUES (
    (SELECT id FROM groups WHERE handle = 'test-group'),
    (SELECT id FROM users WHERE email = 'admin1@test.com'),
    'admin',
    (SELECT id FROM users WHERE email = 'admin1@test.com'),
    NOW()
);

-- Test: Cannot demote the last admin
SELECT throws_ok(
    $$UPDATE memberships SET role = 'member'
      WHERE user_id = (SELECT id FROM users WHERE email = 'admin1@test.com')
      AND group_id = (SELECT id FROM groups WHERE handle = 'test-group')$$,
    'P0001',  -- raise_exception
    'Cannot remove or demote the last administrator of a group',
    'Demoting last admin should raise exception'
);

-- Test: Cannot delete the last admin
SELECT throws_ok(
    $$DELETE FROM memberships
      WHERE user_id = (SELECT id FROM users WHERE email = 'admin1@test.com')
      AND group_id = (SELECT id FROM groups WHERE handle = 'test-group')$$,
    'P0001',  -- raise_exception
    'Cannot remove or demote the last administrator of a group',
    'Deleting last admin should raise exception'
);

-- Add second admin to test that we CAN demote when there are multiple admins
INSERT INTO memberships (group_id, user_id, role, inviter_id, accepted_at)
VALUES (
    (SELECT id FROM groups WHERE handle = 'test-group'),
    (SELECT id FROM users WHERE email = 'admin2@test.com'),
    'admin',
    (SELECT id FROM users WHERE email = 'admin1@test.com'),
    NOW()
);

-- Test: CAN demote an admin when there are multiple admins
SELECT lives_ok(
    $$UPDATE memberships SET role = 'member'
      WHERE user_id = (SELECT id FROM users WHERE email = 'admin2@test.com')
      AND group_id = (SELECT id FROM groups WHERE handle = 'test-group')$$,
    'Demoting admin should succeed when multiple admins exist'
);

-- Test: CAN delete a member (not an admin)
SELECT lives_ok(
    $$DELETE FROM memberships
      WHERE user_id = (SELECT id FROM users WHERE email = 'admin2@test.com')
      AND group_id = (SELECT id FROM groups WHERE handle = 'test-group')$$,
    'Deleting member should succeed'
);

-- Test: Pending admin (not yet accepted) can be demoted even if only admin
INSERT INTO memberships (group_id, user_id, role, inviter_id, accepted_at)
VALUES (
    (SELECT id FROM groups WHERE handle = 'test-group'),
    (SELECT id FROM users WHERE email = 'admin2@test.com'),
    'admin',
    (SELECT id FROM users WHERE email = 'admin1@test.com'),
    NULL  -- Not yet accepted
);

SELECT lives_ok(
    $$UPDATE memberships SET role = 'member'
      WHERE user_id = (SELECT id FROM users WHERE email = 'admin2@test.com')
      AND group_id = (SELECT id FROM groups WHERE handle = 'test-group')$$,
    'Demoting pending admin should succeed (not yet accepted)'
);

SELECT * FROM finish();
ROLLBACK;
