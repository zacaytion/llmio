// Package main provides database migration commands using goose.
//
// Note: This migrate tool intentionally provides a simplified subset of goose commands:
//   - up: Run all pending migrations
//   - down: Roll back the most recent migration
//   - status: Show migration status
//   - version: Show current database version
//   - create: Create a new migration file
//
// Advanced goose commands (up-by-one, up-to, down-to, redo, reset) are omitted
// for simplicity. For advanced use cases, use goose CLI directly:
//
//	go install github.com/pressly/goose/v3/cmd/goose@latest
//	goose -dir migrations postgres "user=z dbname=loomio_development" up-by-one
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/zacaytion/llmio/internal/config"
)

var (
	cfgFile string
	cfg     *config.Config
	v       *viper.Viper
)

func main() {
	// Create viper instance for CLI flag binding
	v = config.NewViper()

	rootCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration tool",
		Long:  "Database migration tool using goose - manages PostgreSQL schema changes",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			cfg, err = config.LoadWithViper(v, cfgFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			return nil
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")

	// PostgreSQL flags (shared across all subcommands)
	// These match the LLMIO_PG_* environment variable names
	rootCmd.PersistentFlags().String("pg-host", "localhost", "PostgreSQL host")
	rootCmd.PersistentFlags().Int("pg-port", 5432, "PostgreSQL port")
	rootCmd.PersistentFlags().String("pg-database", "loomio_development", "PostgreSQL database name")
	rootCmd.PersistentFlags().String("pg-sslmode", "disable", "PostgreSQL SSL mode")

	// Admin credentials (superuser for role creation)
	rootCmd.PersistentFlags().String("pg-user-admin", "postgres", "PostgreSQL admin user")
	rootCmd.PersistentFlags().String("pg-pass-admin", "", "PostgreSQL admin password")

	// Migration role credentials (for connecting as migration user)
	rootCmd.PersistentFlags().String("pg-user-migration", "", "PostgreSQL migration user")
	rootCmd.PersistentFlags().String("pg-pass-migration", "", "PostgreSQL migration password")

	// Bind flags to viper
	if err := bindFlags(rootCmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Subcommands
	rootCmd.AddCommand(upCmd())
	rootCmd.AddCommand(downCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(createCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func upCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Migrate database to most recent version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigration("up")
		},
	}
}

func downCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "down",
		Short: "Roll back the most recent migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigration("down")
		},
	}
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigration("status")
		},
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show current database version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigration("version")
		},
	}
}

func createCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new migration file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			migrationsDir, err := findMigrationsDir()
			if err != nil {
				return fmt.Errorf("cannot locate migrations: %w", err)
			}
			if err := goose.Create(nil, migrationsDir, args[0], "sql"); err != nil {
				return fmt.Errorf("failed to create migration: %w", err)
			}
			return nil
		},
	}
}

func runMigration(command string) error {
	// Find migrations directory
	migrationsDir, err := findMigrationsDir()
	if err != nil {
		return fmt.Errorf("cannot locate migrations: %w", err)
	}

	// Bridge config values to goose envsub environment variables.
	// Goose reads directly from OS env vars, not from Viper config,
	// so we need to set these before running migrations.
	setGooseEnvVars()

	// Build DSN from config - migrations run as admin user
	dsn := cfg.PG.AdminDSN()

	slog.Info("connecting to database",
		"host", cfg.PG.Host,
		"port", cfg.PG.Port,
		"database", cfg.PG.Database,
		"user", cfg.PG.UserAdmin,
	)

	// Open database connection using pgx stdlib driver
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { _ = conn.Close() }()

	// Verify connection
	if err := conn.PingContext(context.Background()); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	// Run the command
	if err := goose.RunContext(context.Background(), command, conn, migrationsDir); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}

// setGooseEnvVars bridges Viper config to OS environment variables for goose envsub.
// Goose's ENVSUB directive reads directly from OS env vars, not from Viper,
// so this function must be called before running any migrations.
//
// The env var names match LLMIO_PG_* pattern used by the app config.
func setGooseEnvVars() {
	if cfg.PG.PassMigration != "" {
		_ = os.Setenv("LLMIO_PG_PASS_MIGRATION", cfg.PG.PassMigration)
	}
	if cfg.PG.PassApp != "" {
		_ = os.Setenv("LLMIO_PG_PASS_APP", cfg.PG.PassApp)
	}
}

// findMigrationsDir looks for the migrations directory.
// Returns an error if no migrations directory can be found.
func findMigrationsDir() (string, error) {
	// Try common locations
	candidates := []string{
		"db/migrations",
		"./db/migrations",
		"../../db/migrations", // When running from cmd/migrate
	}

	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			absPath, err := filepath.Abs(dir)
			if err != nil {
				return "", fmt.Errorf("failed to get absolute path for migrations directory: %w", err)
			}
			return absPath, nil
		}
	}

	return "", fmt.Errorf("migrations directory not found (tried: %v)", candidates)
}

func bindFlags(cmd *cobra.Command) error {
	// flagBinder collects BindPFlag errors to catch flag name typos at startup
	b := &flagBinder{v: v, cmd: cmd}

	b.bind("pg.host", "pg-host")
	b.bind("pg.port", "pg-port")
	b.bind("pg.database", "pg-database")
	b.bind("pg.sslmode", "pg-sslmode")
	b.bind("pg.user_admin", "pg-user-admin")
	b.bind("pg.pass_admin", "pg-pass-admin")
	b.bind("pg.user_migration", "pg-user-migration")
	b.bind("pg.pass_migration", "pg-pass-migration")

	return b.err()
}

// flagBinder collects errors from BindPFlag calls.
type flagBinder struct {
	v      *viper.Viper
	cmd    *cobra.Command
	errors []string
}

func (b *flagBinder) bind(key, flagName string) {
	flag := b.cmd.PersistentFlags().Lookup(flagName)
	if flag == nil {
		b.errors = append(b.errors, fmt.Sprintf("flag %q not found", flagName))
		return
	}
	if err := b.v.BindPFlag(key, flag); err != nil {
		b.errors = append(b.errors, fmt.Sprintf("failed to bind %q to %q: %v", flagName, key, err))
	}
}

func (b *flagBinder) err() error {
	if len(b.errors) == 0 {
		return nil
	}
	return fmt.Errorf("flag binding errors: %v", b.errors)
}
