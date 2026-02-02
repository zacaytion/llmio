-- pgTap tests for users table schema
-- Run with: pg_prove -d loomio_development migrations/001_create_users_test.sql

BEGIN;
SELECT plan(15);

-- Test table exists
SELECT has_table('users', 'users table should exist');

-- Test columns exist with correct types
SELECT has_column('users', 'id', 'users table should have id column');
SELECT col_type_is('users', 'id', 'bigint', 'id should be bigint');

SELECT has_column('users', 'email', 'users table should have email column');
SELECT col_type_is('users', 'email', 'citext', 'email should be citext for case-insensitivity');

SELECT has_column('users', 'name', 'users table should have name column');
SELECT has_column('users', 'username', 'users table should have username column');
SELECT has_column('users', 'password_hash', 'users table should have password_hash column');
SELECT has_column('users', 'email_verified', 'users table should have email_verified column');
SELECT has_column('users', 'deactivated_at', 'users table should have deactivated_at column');
SELECT has_column('users', 'key', 'users table should have key column');
SELECT has_column('users', 'created_at', 'users table should have created_at column');
SELECT has_column('users', 'updated_at', 'users table should have updated_at column');

-- Test unique constraints
SELECT col_is_unique('users', 'email', 'email should be unique');
SELECT col_is_unique('users', 'username', 'username should be unique');

SELECT * FROM finish();
ROLLBACK;
