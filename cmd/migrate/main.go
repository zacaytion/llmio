// Package main provides database migration commands using goose.
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

	// Database flags (shared across all subcommands)
	rootCmd.PersistentFlags().String("db-host", "localhost", "database host")
	rootCmd.PersistentFlags().Int("db-port", 5432, "database port")
	rootCmd.PersistentFlags().String("db-user", "postgres", "database user")
	rootCmd.PersistentFlags().String("db-password", "", "database password")
	rootCmd.PersistentFlags().String("db-name", "loomio_development", "database name")
	rootCmd.PersistentFlags().String("db-sslmode", "disable", "database SSL mode")

	// Bind database flags to viper
	_ = v.BindPFlag("database.host", rootCmd.PersistentFlags().Lookup("db-host"))
	_ = v.BindPFlag("database.port", rootCmd.PersistentFlags().Lookup("db-port"))
	_ = v.BindPFlag("database.user", rootCmd.PersistentFlags().Lookup("db-user"))
	_ = v.BindPFlag("database.password", rootCmd.PersistentFlags().Lookup("db-password"))
	_ = v.BindPFlag("database.name", rootCmd.PersistentFlags().Lookup("db-name"))
	_ = v.BindPFlag("database.sslmode", rootCmd.PersistentFlags().Lookup("db-sslmode"))

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
			migrationsDir := findMigrationsDir()
			return goose.Create(nil, migrationsDir, args[0], "sql")
		},
	}
}

func runMigration(command string) error {
	// Find migrations directory
	migrationsDir := findMigrationsDir()

	// Build DSN from config
	dsn := cfg.Database.DSN()

	slog.Info("connecting to database",
		"host", cfg.Database.Host,
		"port", cfg.Database.Port,
		"name", cfg.Database.Name,
		"user", cfg.Database.User,
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
