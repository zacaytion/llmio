//go:build integration

package api_test

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	dbtestutil "github.com/zacaytion/llmio/internal/db/testutil"
	"github.com/zacaytion/llmio/internal/testutil"
)

// Test_AppRole_PrivilegeEnforcement verifies that the loomio_app role has correct
// privileges: can perform DML but cannot perform DDL or write to audit schema.
// This is a critical security test that validates the three-role model.
func Test_AppRole_PrivilegeEnforcement(t *testing.T) {
	t.Cleanup(func() { testutil.Restore(t) })

	ctx := context.Background()
	container := testutil.GetContainer()
	if container == nil {
		t.Fatal("container not initialized - TestMain may have failed")
	}

	// First, ensure roles have test passwords
	if err := container.SetupTestRoles(ctx); err != nil {
		t.Fatalf("failed to setup test roles: %v", err)
	}

	// Connect as loomio_app role
	connStr, err := container.ConnectionStringForRole(
		ctx,
		dbtestutil.TestRoleCredentials.AppUser,
		dbtestutil.TestRoleCredentials.AppPassword,
		"sslmode=disable",
	)
	if err != nil {
		t.Fatalf("failed to get connection string for app role: %v", err)
	}

	appPool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to connect as loomio_app: %v", err)
	}
	defer appPool.Close()

	// Verify connection works
	var currentUser string
	err = appPool.QueryRow(ctx, "SELECT current_user").Scan(&currentUser)
	if err != nil {
		t.Fatalf("failed to query current_user: %v", err)
	}
	if currentUser != "loomio_app" {
		t.Errorf("expected current_user to be loomio_app, got %s", currentUser)
	}

	t.Run("can_perform_dml_on_data_schema", func(t *testing.T) {
		// loomio_app should be able to INSERT into data.users
		_, err := appPool.Exec(ctx, `
			INSERT INTO data.users (email, name, username, password_hash, key)
			VALUES ('apptest@example.com', 'App Test', 'apptest', 'hash', 'key123')
		`)
		if err != nil {
			t.Errorf("loomio_app should be able to INSERT into data.users: %v", err)
		}

		// Should be able to SELECT
		var count int
		err = appPool.QueryRow(ctx, "SELECT COUNT(*) FROM data.users WHERE email = 'apptest@example.com'").Scan(&count)
		if err != nil {
			t.Errorf("loomio_app should be able to SELECT from data.users: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1 user, got %d", count)
		}

		// Should be able to UPDATE
		_, err = appPool.Exec(ctx, "UPDATE data.users SET name = 'Updated' WHERE email = 'apptest@example.com'")
		if err != nil {
			t.Errorf("loomio_app should be able to UPDATE data.users: %v", err)
		}

		// Should be able to DELETE
		_, err = appPool.Exec(ctx, "DELETE FROM data.users WHERE email = 'apptest@example.com'")
		if err != nil {
			t.Errorf("loomio_app should be able to DELETE from data.users: %v", err)
		}
	})

	t.Run("cannot_perform_ddl", func(t *testing.T) {
		// loomio_app should NOT be able to CREATE TABLE
		_, err := appPool.Exec(ctx, "CREATE TABLE data.test_forbidden (id SERIAL PRIMARY KEY)")
		if err == nil {
			t.Error("loomio_app should NOT be able to CREATE TABLE in data schema")
			// Clean up if it somehow succeeded
			_, _ = appPool.Exec(ctx, "DROP TABLE data.test_forbidden")
		} else if !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("expected 'permission denied' error, got: %v", err)
		}

		// Should NOT be able to DROP TABLE
		_, err = appPool.Exec(ctx, "DROP TABLE data.users")
		if err == nil {
			t.Error("loomio_app should NOT be able to DROP TABLE")
		} else if !strings.Contains(err.Error(), "permission denied") && !strings.Contains(err.Error(), "must be owner") {
			t.Errorf("expected permission error, got: %v", err)
		}

		// Should NOT be able to ALTER TABLE
		_, err = appPool.Exec(ctx, "ALTER TABLE data.users ADD COLUMN forbidden TEXT")
		if err == nil {
			t.Error("loomio_app should NOT be able to ALTER TABLE")
			// Clean up if it somehow succeeded
			_, _ = appPool.Exec(ctx, "ALTER TABLE data.users DROP COLUMN forbidden")
		} else if !strings.Contains(err.Error(), "permission denied") && !strings.Contains(err.Error(), "must be owner") {
			t.Errorf("expected permission error, got: %v", err)
		}

		// Should NOT be able to TRUNCATE
		_, err = appPool.Exec(ctx, "TRUNCATE data.users")
		if err == nil {
			t.Error("loomio_app should NOT be able to TRUNCATE")
		} else if !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("expected 'permission denied' error, got: %v", err)
		}
	})

	t.Run("cannot_write_to_audit_schema", func(t *testing.T) {
		// loomio_app should NOT be able to INSERT into audit schema
		_, err := appPool.Exec(ctx, `
			INSERT INTO audit.record_version (record_id, old_record_id, op, ts, table_oid, table_schema, table_name, record)
			VALUES ('fake', 'fake', 'I', now(), 0, 'audit', 'fake', '{}')
		`)
		if err == nil {
			t.Error("loomio_app should NOT be able to INSERT into audit.record_version")
		} else if !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("expected 'permission denied' error, got: %v", err)
		}

		// Should NOT be able to DELETE from audit schema
		_, err = appPool.Exec(ctx, "DELETE FROM audit.record_version")
		if err == nil {
			t.Error("loomio_app should NOT be able to DELETE from audit.record_version")
		} else if !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("expected 'permission denied' error, got: %v", err)
		}

		// Should NOT be able to TRUNCATE audit schema
		_, err = appPool.Exec(ctx, "TRUNCATE audit.record_version")
		if err == nil {
			t.Error("loomio_app should NOT be able to TRUNCATE audit.record_version")
		} else if !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("expected 'permission denied' error, got: %v", err)
		}
	})

	t.Run("can_read_audit_schema", func(t *testing.T) {
		// loomio_app SHOULD be able to SELECT from audit schema (read-only access)
		var count int
		err := appPool.QueryRow(ctx, "SELECT COUNT(*) FROM audit.record_version").Scan(&count)
		if err != nil {
			t.Errorf("loomio_app should be able to SELECT from audit.record_version: %v", err)
		}
		// count can be 0, we just care that the query succeeded
	})

	t.Run("cannot_access_private_schema", func(t *testing.T) {
		// loomio_app should NOT be able to call private functions directly
		// Note: private schema contains trigger functions that are SECURITY DEFINER,
		// so they run as the owner. But direct access should be denied.
		_, err := appPool.Exec(ctx, "SELECT private.protect_last_admin()")
		if err == nil {
			t.Error("loomio_app should NOT be able to call private.protect_last_admin() directly")
		}
		// Error could be "permission denied" or "function does not exist" (no USAGE on schema)
	})
}
