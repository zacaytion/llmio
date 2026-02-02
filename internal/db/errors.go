// Package db provides database access and query helpers.
package db

import (
	"errors"

	"github.com/jackc/pgx/v5"
)

// IsNotFound returns true if the error indicates no rows were found.
// This wraps pgx.ErrNoRows to provide a consistent API for error checking.
func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
