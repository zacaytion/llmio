// Package main provides database migration commands using goose.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/zacaytion/llmio/internal/db"
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]

	// Find migrations directory (relative to working directory)
	migrationsDir := findMigrationsDir()

	// Get database configuration
	cfg := db.DefaultConfig()
	dsn := cfg.DSN()

	// Open database connection using pgx stdlib driver
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = conn.Close() }()

	// Verify connection
	if err := conn.PingContext(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Failed to set dialect: %v", err)
	}

	// Run the command
	if err := goose.RunContext(context.Background(), command, conn, migrationsDir, args[1:]...); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}

// findMigrationsDir looks for the migrations directory.
func findMigrationsDir() string {
	// Try common locations
	candidates := []string{
		"migrations",
		"./migrations",
		"../../migrations", // When running from cmd/migrate
	}

	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			absPath, _ := filepath.Abs(dir)
			return absPath
		}
	}

	// Default to migrations in current directory
	return "migrations"
}

func printUsage() {
	fmt.Println("Usage: migrate <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up                   Migrate the database to the most recent version")
	fmt.Println("  up-by-one            Migrate the database up by one version")
	fmt.Println("  up-to VERSION        Migrate the database to a specific version")
	fmt.Println("  down                 Roll back the most recent migration")
	fmt.Println("  down-to VERSION      Roll back to a specific version")
	fmt.Println("  redo                 Roll back and re-run the most recent migration")
	fmt.Println("  reset                Roll back all migrations")
	fmt.Println("  status               Show migration status")
	fmt.Println("  version              Show current database version")
	fmt.Println("  create NAME sql      Create a new migration file")
	fmt.Println()
	fmt.Println("Environment variables:")
	fmt.Println("  DB_HOST     Database host (default: localhost)")
	fmt.Println("  DB_USER     Database user (default: postgres)")
	fmt.Println("  DB_PASSWORD Database password (default: empty)")
	fmt.Println("  DB_NAME     Database name (default: loomio_development)")
	fmt.Println("  DB_SSLMODE  SSL mode (default: disable)")
}
