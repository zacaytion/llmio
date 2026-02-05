// Package main provides the server entrypoint.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/zacaytion/llmio/internal/api"
	"github.com/zacaytion/llmio/internal/auth"
	"github.com/zacaytion/llmio/internal/config"
	"github.com/zacaytion/llmio/internal/db"
	"github.com/zacaytion/llmio/internal/logging"
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
		Use:   "server",
		Short: "Loomio API server",
		Long:  "Loomio API server - a collaborative decision-making platform",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			cfg, err = config.LoadWithViper(v, cfgFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			return nil
		},
		RunE: runServer,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")

	// Server flags
	rootCmd.Flags().Int("port", 8080, "server port")
	rootCmd.Flags().Duration("http-read-timeout", 15*time.Second, "HTTP read timeout")
	rootCmd.Flags().Duration("http-write-timeout", 15*time.Second, "HTTP write timeout")
	rootCmd.Flags().Duration("http-idle-timeout", 60*time.Second, "HTTP idle timeout")

	// PostgreSQL flags (match LLMIO_PG_* env vars)
	rootCmd.Flags().String("pg-host", "localhost", "PostgreSQL host")
	rootCmd.Flags().Int("pg-port", 5432, "PostgreSQL port")
	rootCmd.Flags().String("pg-database", "loomio_development", "PostgreSQL database name")
	rootCmd.Flags().String("pg-sslmode", "disable", "PostgreSQL SSL mode")
	rootCmd.Flags().Int32("pg-max-conns", 25, "max database connections")
	rootCmd.Flags().Int32("pg-min-conns", 2, "min database connections")
	rootCmd.Flags().Duration("pg-max-conn-lifetime", time.Hour, "max connection lifetime")
	rootCmd.Flags().Duration("pg-max-conn-idle-time", 30*time.Minute, "max connection idle time")
	rootCmd.Flags().Duration("pg-health-check-period", time.Minute, "health check period")

	// App credentials (for runtime queries)
	rootCmd.Flags().String("pg-user-app", "", "PostgreSQL app user")
	rootCmd.Flags().String("pg-pass-app", "", "PostgreSQL app password")

	// Session flags
	rootCmd.Flags().Duration("session-duration", 168*time.Hour, "session duration")
	rootCmd.Flags().Duration("session-cleanup-interval", 10*time.Minute, "session cleanup interval")

	// Logging flags
	rootCmd.Flags().String("log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.Flags().String("log-format", "json", "log format (json, text)")
	rootCmd.Flags().String("log-output", "stdout", "log output (stdout, stderr, or file path)")

	// Bind flags to viper for priority override.
	// Note: We use fmt.Fprintln to stderr for errors here because logging
	// isn't configured yet - config must be loaded first to know log settings.
	if err := bindFlags(rootCmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func bindFlags(cmd *cobra.Command) error {
	// flagBinder collects BindPFlag errors to catch flag name typos at startup
	b := &flagBinder{v: v, cmd: cmd}

	// Bind server flags to our viper instance
	b.bind("server.port", "port")
	b.bind("server.read_timeout", "http-read-timeout")
	b.bind("server.write_timeout", "http-write-timeout")
	b.bind("server.idle_timeout", "http-idle-timeout")

	// Bind PostgreSQL flags
	b.bind("pg.host", "pg-host")
	b.bind("pg.port", "pg-port")
	b.bind("pg.database", "pg-database")
	b.bind("pg.sslmode", "pg-sslmode")
	b.bind("pg.max_conns", "pg-max-conns")
	b.bind("pg.min_conns", "pg-min-conns")
	b.bind("pg.max_conn_lifetime", "pg-max-conn-lifetime")
	b.bind("pg.max_conn_idle_time", "pg-max-conn-idle-time")
	b.bind("pg.health_check_period", "pg-health-check-period")
	b.bind("pg.user_app", "pg-user-app")
	b.bind("pg.pass_app", "pg-pass-app")

	// Bind session flags
	b.bind("session.duration", "session-duration")
	b.bind("session.cleanup_interval", "session-cleanup-interval")

	// Bind logging flags
	b.bind("logging.level", "log-level")
	b.bind("logging.format", "log-format")
	b.bind("logging.output", "log-output")

	return b.err()
}

// flagBinder collects errors from BindPFlag calls.
type flagBinder struct {
	v      *viper.Viper
	cmd    *cobra.Command
	errors []string
}

func (b *flagBinder) bind(key, flagName string) {
	flag := b.cmd.Flags().Lookup(flagName)
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

func runServer(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Setup logging from config with cleanup support
	closeLogger, err := logging.SetupDefaultWithCleanup(cfg.Logging)
	if err != nil {
		return fmt.Errorf("failed to setup logging: %w", err)
	}
	defer func() {
		if err := closeLogger(); err != nil {
			// Can't use slog here as we're closing it
			fmt.Fprintf(os.Stderr, "error closing log file: %v\n", err)
		}
	}()

	// Connect to database using config
	pool, err := db.NewPoolFromConfig(ctx, cfg.PG)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	// Create session store with configured duration
	sessionStore := auth.NewSessionStoreWithConfig(cfg.Session.Duration)

	// Start session cleanup goroutine with cancellation context
	cleanupCtx, cancelCleanup := context.WithCancel(context.Background())
	defer cancelCleanup()
	go startSessionCleanup(cleanupCtx, sessionStore, cfg.Session.CleanupInterval)

	// Create router using stdlib ServeMux
	mux := http.NewServeMux()

	// Create Huma API with stdlib adapter
	humaAPI := humago.New(mux, huma.DefaultConfig("Loomio API", "1.0.0"))

	// Create queries instance
	queries := db.New(pool)

	// Create app with dependencies
	app := &App{
		Pool:         pool,
		Queries:      queries,
		SessionStore: sessionStore,
	}

	// Register routes
	app.RegisterRoutes(humaAPI)

	// Create server with config values
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      api.LoggingMiddleware(mux),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Error channel for server goroutine
	serverErr := make(chan error, 1)

	// Start server in goroutine
	go func() {
		slog.Info("server starting", "port", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for interrupt signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return fmt.Errorf("server failed: %w", err)
	case <-quit:
		slog.Info("shutting down server")
	}

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	slog.Info("server exited")
	return nil
}

func startSessionCleanup(ctx context.Context, store *auth.SessionStore, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.DebugContext(ctx, "session cleanup goroutine stopped")
			return
		case <-ticker.C:
			cleaned := store.CleanupExpired()
			if cleaned > 0 {
				slog.InfoContext(ctx, "cleaned expired sessions", "count", cleaned)
			}
		}
	}
}

// App holds application dependencies for handler registration.
type App struct {
	Pool         *pgxpool.Pool
	Queries      *db.Queries
	SessionStore *auth.SessionStore
}

// RegisterRoutes registers all API routes.
func (a *App) RegisterRoutes(humaAPI huma.API) {
	// Health check
	huma.Get(humaAPI, "/health", func(ctx context.Context, input *struct{}) (*struct {
		Body struct {
			Status string `json:"status"`
		}
	}, error) {
		return &struct {
			Body struct {
				Status string `json:"status"`
			}
		}{Body: struct {
			Status string `json:"status"`
		}{Status: "ok"}}, nil
	})

	// Auth routes
	authHandler := api.NewAuthHandler(a.Queries, a.SessionStore)
	authHandler.RegisterRoutes(humaAPI)

	// Group routes (Feature 004)
	groupHandler := api.NewGroupHandler(a.Pool, a.Queries, a.SessionStore)
	groupHandler.RegisterRoutes(humaAPI)

	// Membership routes (Feature 004)
	membershipHandler := api.NewMembershipHandler(a.Pool, a.Queries, a.SessionStore)
	membershipHandler.RegisterRoutes(humaAPI)

	slog.Debug("routes registered")
}
