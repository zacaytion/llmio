package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SetAuditContext sets the app.current_user_id session variable for audit logging.
// This must be called at the start of a transaction before any mutations.
// The actor_id will be captured by audit triggers on groups and memberships tables.
//
// Usage:
//
//	err := pgx.BeginTxFunc(ctx, pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
//	    if err := db.SetAuditContext(ctx, tx, userID); err != nil {
//	        return err
//	    }
//	    // ... perform mutations ...
//	    return nil
//	})
func SetAuditContext(ctx context.Context, tx pgx.Tx, userID int64) error {
	_, err := tx.Exec(ctx, "SET LOCAL app.current_user_id = $1", fmt.Sprintf("%d", userID))
	return err
}

// WithAuditContext executes a function within a transaction with audit context set.
// This is a convenience wrapper that combines transaction creation, audit context setup,
// and automatic commit/rollback.
//
// Usage:
//
//	result, err := db.WithAuditContext(ctx, pool, userID, func(tx pgx.Tx) (*db.Group, error) {
//	    return queries.WithTx(tx).CreateGroup(ctx, params)
//	})
func WithAuditContext[T any](ctx context.Context, pool *pgxpool.Pool, userID int64, fn func(tx pgx.Tx) (T, error)) (T, error) {
	var result T
	err := pgx.BeginTxFunc(ctx, pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if err := SetAuditContext(ctx, tx, userID); err != nil {
			return err
		}
		var fnErr error
		result, fnErr = fn(tx)
		return fnErr
	})
	return result, err
}

// WithAuditContextExec executes a function within a transaction with audit context set.
// Similar to WithAuditContext but for functions that don't return a value.
func WithAuditContextExec(ctx context.Context, pool *pgxpool.Pool, userID int64, fn func(tx pgx.Tx) error) error {
	return pgx.BeginTxFunc(ctx, pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if err := SetAuditContext(ctx, tx, userID); err != nil {
			return err
		}
		return fn(tx)
	})
}
