package api

import (
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

// ============================================================
// T155-T156: Phase 12 - Unique Violation Detection Tests
// ============================================================

// TestIsUniqueViolation_WrappedErrors tests that isUniqueViolation works with wrapped errors.
// T155: Test for unique violation detection with wrapped error
func Test_IsUniqueViolation_WrappedErrors(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		constraintName string
		want           bool
	}{
		{
			name: "direct pgconn.PgError unique violation",
			err: &pgconn.PgError{
				Code:           "23505",
				ConstraintName: "groups_handle_key",
			},
			constraintName: "groups_handle_key",
			want:           true,
		},
		{
			name: "wrapped pgconn.PgError unique violation",
			err: fmt.Errorf("transaction failed: %w", &pgconn.PgError{
				Code:           "23505",
				ConstraintName: "groups_handle_key",
			}),
			constraintName: "groups_handle_key",
			want:           true,
		},
		{
			name: "deeply wrapped pgconn.PgError",
			err: fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", &pgconn.PgError{
				Code:           "23505",
				ConstraintName: "memberships_unique_user_group",
			})),
			constraintName: "memberships_unique_user_group",
			want:           true,
		},
		{
			name: "wrong constraint name",
			err: &pgconn.PgError{
				Code:           "23505",
				ConstraintName: "groups_handle_key",
			},
			constraintName: "wrong_constraint",
			want:           false,
		},
		{
			name: "wrong error code (23503 = foreign key)",
			err: &pgconn.PgError{
				Code:           "23503",
				ConstraintName: "groups_handle_key",
			},
			constraintName: "groups_handle_key",
			want:           false,
		},
		{
			name:           "plain error (not pgconn.PgError)",
			err:            fmt.Errorf("some random error 23505 groups_handle_key"),
			constraintName: "groups_handle_key",
			want:           false, // T155: String matching should NOT match
		},
		{
			name:           "nil error",
			err:            nil,
			constraintName: "groups_handle_key",
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUniqueViolation(tt.err, tt.constraintName)
			if got != tt.want {
				t.Errorf("isUniqueViolation() = %v, want %v", got, tt.want)
			}
		})
	}
}
