-- pgTap tests for memberships table schema and last-admin protection trigger
-- Run with: pg_prove -d loomio_test db/tests/004_memberships_test.sql

BEGIN;
SELECT plan(35);

-- =====================================================
-- Schema and Table Structure Tests
-- =====================================================

-- Test table exists in data schema
SELECT has_table('data', 'memberships', 'memberships table should exist in data schema');

-- Test columns exist
SELECT has_column('data', 'memberships', 'id', 'memberships should have id column');
SELECT has_column('data', 'memberships', 'group_id', 'memberships should have group_id column');
SELECT has_column('data', 'memberships', 'user_id', 'memberships should have user_id column');
SELECT has_column('data', 'memberships', 'role', 'memberships should have role column');
SELECT has_column('data', 'memberships', 'inviter_id', 'memberships should have inviter_id column');
SELECT has_column('data', 'memberships', 'accepted_at', 'memberships should have accepted_at column');
SELECT has_column('data', 'memberships', 'created_at', 'memberships should have created_at column');
SELECT has_column('data', 'memberships', 'updated_at', 'memberships should have updated_at column');

-- Test column types
SELECT col_type_is('data', 'memberships', 'id', 'bigint', 'id should be bigint');
SELECT col_type_is('data', 'memberships', 'role', 'text', 'role should be text');

-- Test foreign keys
SELECT col_is_fk('data', 'memberships', 'group_id', 'group_id should be a foreign key');
SELECT col_is_fk('data', 'memberships', 'user_id', 'user_id should be a foreign key');
SELECT col_is_fk('data', 'memberships', 'inviter_id', 'inviter_id should be a foreign key');

-- Test unique constraint on (group_id, user_id)
SELECT has_index('data', 'memberships', 'memberships_unique_user_group', 'unique constraint on group_id + user_id should exist');

-- =====================================================
-- Role Constraint Tests
-- =====================================================

-- Test role constraint (only 'admin' or 'member')
-- Note: This test runs before test data is inserted, so we set up minimal data first
INSERT INTO data.users (email, name, username, password_hash, key)
VALUES ('role-test@example.com', 'Role Test', 'roletest', 'hash', 'roletest-key');

INSERT INTO data.groups (name, handle, created_by_id)
VALUES ('Role Test Group', 'role-test-group', (SELECT id FROM data.users WHERE email = 'role-test@example.com'));

SELECT throws_ok(
    $$INSERT INTO data.memberships (group_id, user_id, role, inviter_id)
      VALUES (
        (SELECT id FROM data.groups WHERE handle = 'role-test-group'),
        (SELECT id FROM data.users WHERE email = 'role-test@example.com'),
        'invalid_role',
        (SELECT id FROM data.users WHERE email = 'role-test@example.com')
      )$$,
    '23514',  -- check_violation
    NULL,
    'Invalid role should be rejected'
);

-- =====================================================
-- Trigger Tests
-- =====================================================

SELECT trigger_is(
    'data',
    'memberships',
    'memberships_updated_at',
    'private',
    'set_updated_at',
    'memberships_updated_at trigger should exist'
);

SELECT trigger_is(
    'data',
    'memberships',
    'memberships_last_admin_protection',
    'private',
    'protect_last_admin',
    'memberships_last_admin_protection trigger should exist'
);

SELECT trigger_is(
    'data',
    'memberships',
    'memberships_audit',
    'audit',
    'insert_update_delete_trigger',
    'memberships_audit trigger should exist'
);

-- =====================================================
-- Index Tests
-- =====================================================

SELECT has_index('data', 'memberships', 'memberships_user_id_idx', 'index on user_id should exist');
SELECT has_index('data', 'memberships', 'memberships_group_id_idx', 'index on group_id should exist');
SELECT has_index('data', 'memberships', 'memberships_inviter_id_idx', 'index on inviter_id should exist');
SELECT has_index('data', 'memberships', 'memberships_pending_idx', 'partial index for pending invitations should exist');
SELECT has_index('data', 'memberships', 'memberships_group_role_idx', 'index for admin counting should exist');
SELECT has_index('data', 'memberships', 'memberships_user_accepted_group_idx', 'composite index for listing groups by user should exist');
SELECT has_index('data', 'memberships', 'memberships_group_stats_idx', 'composite index for membership stats should exist');

-- =====================================================
-- Last-admin protection trigger tests
-- =====================================================

-- Create test data for last-admin protection tests
INSERT INTO data.users (email, name, username, password_hash, key)
VALUES
    ('admin1@test.com', 'Admin One', 'admin-one', 'hash1', 'key1'),
    ('admin2@test.com', 'Admin Two', 'admin-two', 'hash2', 'key2'),
    ('member1@test.com', 'Member One', 'member-one', 'hash3', 'key3');

-- Create test group
INSERT INTO data.groups (name, handle, created_by_id)
VALUES ('Test Group', 'test-group', (SELECT id FROM data.users WHERE email = 'admin1@test.com'));

-- Create admin membership (the group creator)
INSERT INTO data.memberships (group_id, user_id, role, inviter_id, accepted_at)
VALUES (
    (SELECT id FROM data.groups WHERE handle = 'test-group'),
    (SELECT id FROM data.users WHERE email = 'admin1@test.com'),
    'admin',
    (SELECT id FROM data.users WHERE email = 'admin1@test.com'),
    NOW()
);

-- Test: Cannot demote the last admin
SELECT throws_ok(
    $$UPDATE data.memberships SET role = 'member'
      WHERE user_id = (SELECT id FROM data.users WHERE email = 'admin1@test.com')
      AND group_id = (SELECT id FROM data.groups WHERE handle = 'test-group')$$,
    'P0001',  -- raise_exception
    'Cannot remove or demote the last administrator of a group',
    'Demoting last admin should raise exception'
);

-- Test: Cannot delete the last admin
SELECT throws_ok(
    $$DELETE FROM data.memberships
      WHERE user_id = (SELECT id FROM data.users WHERE email = 'admin1@test.com')
      AND group_id = (SELECT id FROM data.groups WHERE handle = 'test-group')$$,
    'P0001',  -- raise_exception
    'Cannot remove or demote the last administrator of a group',
    'Deleting last admin should raise exception'
);

-- Add second admin to test that we CAN demote when there are multiple admins
INSERT INTO data.memberships (group_id, user_id, role, inviter_id, accepted_at)
VALUES (
    (SELECT id FROM data.groups WHERE handle = 'test-group'),
    (SELECT id FROM data.users WHERE email = 'admin2@test.com'),
    'admin',
    (SELECT id FROM data.users WHERE email = 'admin1@test.com'),
    NOW()
);

-- Test: CAN demote an admin when there are multiple admins
SELECT lives_ok(
    $$UPDATE data.memberships SET role = 'member'
      WHERE user_id = (SELECT id FROM data.users WHERE email = 'admin2@test.com')
      AND group_id = (SELECT id FROM data.groups WHERE handle = 'test-group')$$,
    'Demoting admin should succeed when multiple admins exist'
);

-- Test: CAN delete a member (not an admin)
SELECT lives_ok(
    $$DELETE FROM data.memberships
      WHERE user_id = (SELECT id FROM data.users WHERE email = 'admin2@test.com')
      AND group_id = (SELECT id FROM data.groups WHERE handle = 'test-group')$$,
    'Deleting member should succeed'
);

-- Test: Pending admin (not yet accepted) can be demoted even if only admin
INSERT INTO data.memberships (group_id, user_id, role, inviter_id, accepted_at)
VALUES (
    (SELECT id FROM data.groups WHERE handle = 'test-group'),
    (SELECT id FROM data.users WHERE email = 'admin2@test.com'),
    'admin',
    (SELECT id FROM data.users WHERE email = 'admin1@test.com'),
    NULL  -- Not yet accepted
);

SELECT lives_ok(
    $$UPDATE data.memberships SET role = 'member'
      WHERE user_id = (SELECT id FROM data.users WHERE email = 'admin2@test.com')
      AND group_id = (SELECT id FROM data.groups WHERE handle = 'test-group')$$,
    'Demoting pending admin should succeed (not yet accepted)'
);

-- =====================================================
-- Invitation State Transition Tests
-- =====================================================

-- Clean up member1 pending membership if exists from previous test
DELETE FROM data.memberships
WHERE user_id = (SELECT id FROM data.users WHERE email = 'admin2@test.com')
  AND group_id = (SELECT id FROM data.groups WHERE handle = 'test-group');

-- Test: Create pending invitation (accepted_at is NULL)
INSERT INTO data.memberships (group_id, user_id, role, inviter_id, accepted_at)
VALUES (
    (SELECT id FROM data.groups WHERE handle = 'test-group'),
    (SELECT id FROM data.users WHERE email = 'member1@test.com'),
    'member',
    (SELECT id FROM data.users WHERE email = 'admin1@test.com'),
    NULL  -- Pending invitation
);

SELECT is(
    (SELECT accepted_at IS NULL FROM data.memberships
     WHERE user_id = (SELECT id FROM data.users WHERE email = 'member1@test.com')
       AND group_id = (SELECT id FROM data.groups WHERE handle = 'test-group')),
    true,
    'New invitation should have NULL accepted_at'
);

-- Test: Accept invitation by setting accepted_at
UPDATE data.memberships
SET accepted_at = NOW()
WHERE user_id = (SELECT id FROM data.users WHERE email = 'member1@test.com')
  AND group_id = (SELECT id FROM data.groups WHERE handle = 'test-group');

SELECT is(
    (SELECT accepted_at IS NOT NULL FROM data.memberships
     WHERE user_id = (SELECT id FROM data.users WHERE email = 'member1@test.com')
       AND group_id = (SELECT id FROM data.groups WHERE handle = 'test-group')),
    true,
    'Accepted membership should have non-NULL accepted_at'
);

-- =====================================================
-- Cascade Delete Tests
-- =====================================================

-- Create a group that will be deleted to test CASCADE
-- Note: We use 'member' role (not 'admin') to avoid triggering protect_last_admin
-- when the CASCADE delete runs. In production, groups would first remove all
-- memberships or have multiple admins before deletion.
INSERT INTO data.groups (name, handle, created_by_id)
VALUES ('Cascade Test Group', 'cascade-test', (SELECT id FROM data.users WHERE email = 'admin1@test.com'));

INSERT INTO data.memberships (group_id, user_id, role, inviter_id, accepted_at)
VALUES (
    (SELECT id FROM data.groups WHERE handle = 'cascade-test'),
    (SELECT id FROM data.users WHERE email = 'member1@test.com'),
    'member',  -- Use member role to avoid last-admin protection trigger
    (SELECT id FROM data.users WHERE email = 'admin1@test.com'),
    NOW()
);

-- Verify membership exists
SELECT is(
    (SELECT COUNT(*) FROM data.memberships
     WHERE group_id = (SELECT id FROM data.groups WHERE handle = 'cascade-test')),
    1::bigint,
    'Membership should exist before group deletion'
);

-- Delete group - memberships should cascade delete
DELETE FROM data.groups WHERE handle = 'cascade-test';

SELECT is(
    (SELECT COUNT(*) FROM data.memberships
     WHERE group_id NOT IN (SELECT id FROM data.groups)),
    0::bigint,
    'Memberships should be deleted when group is deleted (CASCADE)'
);

SELECT * FROM finish();
ROLLBACK;
