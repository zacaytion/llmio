-- pgTap tests for private schema functions
-- Run with: pg_prove -d loomio_test db/tests/005_private_functions_test.sql

BEGIN;
SELECT plan(11);

-- =====================================================
-- Schema Existence Tests
-- =====================================================

SELECT has_schema('private', 'private schema should exist');

-- =====================================================
-- Function Existence Tests
-- =====================================================

SELECT has_function(
    'private',
    'set_updated_at',
    'private.set_updated_at function should exist'
);

SELECT has_function(
    'private',
    'protect_last_admin',
    'private.protect_last_admin function should exist'
);

-- =====================================================
-- set_updated_at Function Tests
-- =====================================================

-- Test function returns trigger type
SELECT is(
    (SELECT pg_get_function_result(p.oid)
     FROM pg_proc p
     JOIN pg_namespace n ON p.pronamespace = n.oid
     WHERE n.nspname = 'private' AND p.proname = 'set_updated_at'),
    'trigger',
    'set_updated_at should return trigger type'
);

-- Test function language
SELECT is(
    (SELECT l.lanname
     FROM pg_proc p
     JOIN pg_namespace n ON p.pronamespace = n.oid
     JOIN pg_language l ON p.prolang = l.oid
     WHERE n.nspname = 'private' AND p.proname = 'set_updated_at'),
    'plpgsql',
    'set_updated_at should be written in plpgsql'
);

-- Create test data for protect_last_admin tests (also needed for function tests below)
INSERT INTO data.users (email, name, username, password_hash, key)
VALUES ('private-func-test@example.com', 'Function Test User', 'functest', 'hash123', 'functest-key');

-- Note: Cannot test updated_at > created_at within a transaction because NOW() is constant.
-- The trigger attachment test via trigger_is() verifies the trigger is in place.
-- Actual timestamp behavior is tested via Go integration tests.

-- =====================================================
-- protect_last_admin Function Tests
-- =====================================================

-- Test function returns trigger type
SELECT is(
    (SELECT pg_get_function_result(p.oid)
     FROM pg_proc p
     JOIN pg_namespace n ON p.pronamespace = n.oid
     WHERE n.nspname = 'private' AND p.proname = 'protect_last_admin'),
    'trigger',
    'protect_last_admin should return trigger type'
);

-- Test function language
SELECT is(
    (SELECT l.lanname
     FROM pg_proc p
     JOIN pg_namespace n ON p.pronamespace = n.oid
     JOIN pg_language l ON p.prolang = l.oid
     WHERE n.nspname = 'private' AND p.proname = 'protect_last_admin'),
    'plpgsql',
    'protect_last_admin should be written in plpgsql'
);

-- Create test group and membership for protect_last_admin tests
INSERT INTO data.groups (name, handle, created_by_id)
VALUES ('Protect Admin Test', 'protect-admin-test', (SELECT id FROM data.users WHERE email = 'private-func-test@example.com'));

INSERT INTO data.memberships (group_id, user_id, role, inviter_id, accepted_at)
VALUES (
    (SELECT id FROM data.groups WHERE handle = 'protect-admin-test'),
    (SELECT id FROM data.users WHERE email = 'private-func-test@example.com'),
    'admin',
    (SELECT id FROM data.users WHERE email = 'private-func-test@example.com'),
    NOW()
);

-- Test protect_last_admin blocks demotion
SELECT throws_ok(
    $$UPDATE data.memberships SET role = 'member'
      WHERE group_id = (SELECT id FROM data.groups WHERE handle = 'protect-admin-test')$$,
    'P0001',
    'Cannot remove or demote the last administrator of a group',
    'protect_last_admin should prevent demoting last admin'
);

-- Test protect_last_admin blocks deletion
SELECT throws_ok(
    $$DELETE FROM data.memberships
      WHERE group_id = (SELECT id FROM data.groups WHERE handle = 'protect-admin-test')$$,
    'P0001',
    'Cannot remove or demote the last administrator of a group',
    'protect_last_admin should prevent deleting last admin'
);

-- Test protect_last_admin allows demotion when multiple admins exist
INSERT INTO data.users (email, name, username, password_hash, key)
VALUES ('admin2-func@example.com', 'Admin 2', 'admin2func', 'hash456', 'admin2func-key');

INSERT INTO data.memberships (group_id, user_id, role, inviter_id, accepted_at)
VALUES (
    (SELECT id FROM data.groups WHERE handle = 'protect-admin-test'),
    (SELECT id FROM data.users WHERE email = 'admin2-func@example.com'),
    'admin',
    (SELECT id FROM data.users WHERE email = 'private-func-test@example.com'),
    NOW()
);

SELECT lives_ok(
    $$UPDATE data.memberships SET role = 'member'
      WHERE user_id = (SELECT id FROM data.users WHERE email = 'admin2-func@example.com')
      AND group_id = (SELECT id FROM data.groups WHERE handle = 'protect-admin-test')$$,
    'protect_last_admin should allow demotion when multiple admins exist'
);

-- Test protect_last_admin ignores pending admins
INSERT INTO data.users (email, name, username, password_hash, key)
VALUES ('pending-admin@example.com', 'Pending Admin', 'pendingadmin', 'hash789', 'pending-key');

INSERT INTO data.memberships (group_id, user_id, role, inviter_id, accepted_at)
VALUES (
    (SELECT id FROM data.groups WHERE handle = 'protect-admin-test'),
    (SELECT id FROM data.users WHERE email = 'pending-admin@example.com'),
    'admin',
    (SELECT id FROM data.users WHERE email = 'private-func-test@example.com'),
    NULL  -- Pending, not accepted
);

SELECT lives_ok(
    $$UPDATE data.memberships SET role = 'member'
      WHERE user_id = (SELECT id FROM data.users WHERE email = 'pending-admin@example.com')
      AND group_id = (SELECT id FROM data.groups WHERE handle = 'protect-admin-test')$$,
    'protect_last_admin should ignore pending (unaccepted) admins'
);

SELECT * FROM finish();
ROLLBACK;
