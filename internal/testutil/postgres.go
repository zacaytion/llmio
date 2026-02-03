// Package testutil provides test utilities for database testing with isolated containers.
//
// PostgresContainer provides isolated PostgreSQL containers for testing using testcontainers-go.
//
// Usage:
//
//	ctx := context.Background()
//	pg, err := testutil.NewPostgresContainer(ctx)
//	if err != nil { ... }
//	defer pg.Terminate(ctx)
//
//	// Run migrations
//	if err := pg.RunMigrations(ctx); err != nil { ... }
//
//	// Get connection for tests
//	connStr, err := pg.ConnectionString(ctx)
//
//	// Snapshot/restore for fast test reset
//	if err := pg.Snapshot(ctx); err != nil { ... }
//	t.Cleanup(func() { pg.Restore(ctx) })
package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver for database/sql
)

// PostgresContainer wraps a testcontainers PostgreSQL instance with migration support.
type PostgresContainer struct {
	*postgres.PostgresContainer

	migrationsDir string
}

// NewPostgresContainer creates a new PostgreSQL container with CITEXT extension enabled.
// The container uses PostgreSQL 18 (alpine) to match production.
func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	// Locate project root from this file's location
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get current file path")
	}
	projectRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")
	migrationsDir := filepath.Join(projectRoot, "migrations")

	// Create init script that enables required extensions
	initScript := `
		CREATE EXTENSION IF NOT EXISTS citext;
		CREATE EXTENSION IF NOT EXISTS pgtap;
	`

	container, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("loomio_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.WithSQLDriver("pgx"),
		testcontainers.WithStartupCommand(
			testcontainers.NewRawCommand([]string{
				"sh", "-c",
				fmt.Sprintf("echo '%s' > /docker-entrypoint-initdb.d/00-extensions.sql", initScript),
			}),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	return &PostgresContainer{
		PostgresContainer: container,
		migrationsDir:     migrationsDir,
	}, nil
}

// RunMigrations executes all goose migrations against the container database.
func (p *PostgresContainer) RunMigrations(ctx context.Context) error {
	connStr, err := p.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to get connection string: %w", err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, p.migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Snapshot creates a database snapshot for fast restore between tests.
// Call this after migrations and initial data setup, then use Restore() in test cleanup.
func (p *PostgresContainer) Snapshot(ctx context.Context) error {
	return p.PostgresContainer.Snapshot(ctx)
}

// Restore restores the database to the last snapshot.
// Use in t.Cleanup() to reset state between tests.
func (p *PostgresContainer) Restore(ctx context.Context) error {
	return p.PostgresContainer.Restore(ctx)
}

// Terminate stops and removes the container.
// Always defer this after creating a container.
func (p *PostgresContainer) Terminate(ctx context.Context) error {
	return testcontainers.TerminateContainer(p.PostgresContainer)
}

// DB returns a new database connection to the container.
// Caller is responsible for closing the connection.
func (p *PostgresContainer) DB(ctx context.Context) (*sql.DB, error) {
	connStr, err := p.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return db, nil
}
