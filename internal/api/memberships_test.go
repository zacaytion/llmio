package api

import (
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

// ============================================================
// T157: Phase 12 - Last Admin Trigger Error Detection
// ============================================================

// Test_IsLastAdminTriggerError_ErrorDetection tests that isLastAdminTriggerError
// correctly identifies the PostgreSQL P0001 error from the last-admin protection trigger.
// T157: Verify DB trigger error code P0001 detection
func Test_IsLastAdminTriggerError_ErrorDetection(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "exact trigger error",
			err: &pgconn.PgError{
				Code:    "P0001",
				Message: "Cannot remove or demote the last administrator of a group",
			},
			want: true,
		},
		{
			name: "wrapped trigger error",
			err: fmt.Errorf("transaction failed: %w", &pgconn.PgError{
				Code:    "P0001",
				Message: "Cannot remove or demote the last administrator of a group",
			}),
			want: true,
		},
		{
			name: "different P0001 message",
			err: &pgconn.PgError{
				Code:    "P0001",
				Message: "Some other raise exception",
			},
			want: false, // Not our specific trigger
		},
		{
			name: "wrong error code with matching message",
			err: &pgconn.PgError{
				Code:    "23505",
				Message: "Cannot remove or demote the last administrator of a group",
			},
			want: false, // Wrong code
		},
		{
			name: "plain error not pgconn.PgError",
			err:  fmt.Errorf("P0001 Cannot remove or demote the last administrator"),
			want: false, // Not a PgError
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isLastAdminTriggerError(tt.err)
			if got != tt.want {
				t.Errorf("isLastAdminTriggerError() = %v, want %v", got, tt.want)
			}
		})
	}
}
