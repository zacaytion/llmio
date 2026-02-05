-- pgTap tests for users table schema
-- Run with: pg_prove -d loomio_test db/tests/001_create_users_test.sql

BEGIN;
SELECT plan(22);

-- =====================================================
-- Schema and Table Structure Tests
-- =====================================================

-- Test table exists in data schema
SELECT has_table('data', 'users', 'users table should exist in data schema');

-- Test columns exist with correct types
SELECT has_column('data', 'users', 'id', 'users table should have id column');
SELECT col_type_is('data', 'users', 'id', 'bigint', 'id should be bigint');

SELECT has_column('data', 'users', 'email', 'users table should have email column');
SELECT col_type_is('data', 'users', 'email', 'citext', 'email should be citext for case-insensitivity');

SELECT has_column('data', 'users', 'name', 'users table should have name column');
SELECT has_column('data', 'users', 'username', 'users table should have username column');
SELECT has_column('data', 'users', 'password_hash', 'users table should have password_hash column');
SELECT has_column('data', 'users', 'email_verified', 'users table should have email_verified column');
SELECT has_column('data', 'users', 'deactivated_at', 'users table should have deactivated_at column');
SELECT has_column('data', 'users', 'key', 'users table should have key column');
SELECT has_column('data', 'users', 'created_at', 'users table should have created_at column');
SELECT has_column('data', 'users', 'updated_at', 'users table should have updated_at column');

-- Test unique constraints
SELECT col_is_unique('data', 'users', 'email', 'email should be unique');
SELECT col_is_unique('data', 'users', 'username', 'username should be unique');
SELECT col_is_unique('data', 'users', 'key', 'key should be unique');

-- =====================================================
-- Constraint Tests
-- =====================================================

-- Test email format constraint
SELECT throws_ok(
    $$INSERT INTO data.users (email, name, username, password_hash, key)
      VALUES ('invalid-email', 'Test User', 'testuser', 'hash123', 'key123')$$,
    '23514',  -- check_violation
    NULL,
    'Invalid email format should be rejected'
);

-- Test username format constraint (must start with alphanumeric)
SELECT throws_ok(
    $$INSERT INTO data.users (email, name, username, password_hash, key)
      VALUES ('test@example.com', 'Test User', '-invalid', 'hash123', 'key123')$$,
    '23514',  -- check_violation
    NULL,
    'Username starting with hyphen should be rejected'
);

-- Test username format constraint (must end with alphanumeric)
SELECT throws_ok(
    $$INSERT INTO data.users (email, name, username, password_hash, key)
      VALUES ('test@example.com', 'Test User', 'invalid-', 'hash123', 'key123')$$,
    '23514',  -- check_violation
    NULL,
    'Username ending with hyphen should be rejected'
);

-- Test username minimum length constraint
SELECT throws_ok(
    $$INSERT INTO data.users (email, name, username, password_hash, key)
      VALUES ('test@example.com', 'Test User', 'a', 'hash123', 'key123')$$,
    '23514',  -- check_violation
    NULL,
    'Username with length < 2 should be rejected'
);

-- Test name not empty constraint
SELECT throws_ok(
    $$INSERT INTO data.users (email, name, username, password_hash, key)
      VALUES ('test@example.com', '   ', 'testuser', 'hash123', 'key123')$$,
    '23514',  -- check_violation
    NULL,
    'Empty or whitespace-only name should be rejected'
);

-- =====================================================
-- Trigger Tests
-- =====================================================

-- Test updated_at trigger exists
SELECT trigger_is(
    'data',
    'users',
    'users_updated_at',
    'private',
    'set_updated_at',
    'users_updated_at trigger should call private.set_updated_at function'
);

-- Note: Cannot test updated_at > created_at within a transaction because NOW() is constant.
-- The trigger attachment test above verifies the trigger is in place.
-- Actual timestamp behavior is tested via Go integration tests.

SELECT * FROM finish();
ROLLBACK;
