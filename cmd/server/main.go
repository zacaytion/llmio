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

	// Database flags
	rootCmd.Flags().String("db-host", "localhost", "database host")
	rootCmd.Flags().Int("db-port", 5432, "database port")
	rootCmd.Flags().String("db-user", "postgres", "database user")
	rootCmd.Flags().String("db-password", "", "database password")
	rootCmd.Flags().String("db-name", "loomio_development", "database name")
	rootCmd.Flags().String("db-sslmode", "disable", "database SSL mode")
	rootCmd.Flags().Int32("db-max-conns", 25, "max database connections")
	rootCmd.Flags().Int32("db-min-conns", 2, "min database connections")
	rootCmd.Flags().Duration("db-max-conn-lifetime", time.Hour, "max connection lifetime")
	rootCmd.Flags().Duration("db-max-conn-idle-time", 30*time.Minute, "max connection idle time")
	rootCmd.Flags().Duration("db-health-check-period", time.Minute, "health check period")

	// Session flags
	rootCmd.Flags().Duration("session-duration", 168*time.Hour, "session duration")
	rootCmd.Flags().Duration("session-cleanup-interval", 10*time.Minute, "session cleanup interval")

	// Logging flags
	rootCmd.Flags().String("log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.Flags().String("log-format", "json", "log format (json, text)")
	rootCmd.Flags().String("log-output", "stdout", "log output (stdout, stderr, or file path)")

	// Bind flags to viper for priority override
	bindFlags(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func bindFlags(cmd *cobra.Command) {
	// Bind server flags to our viper instance
	_ = v.BindPFlag("server.port", cmd.Flags().Lookup("port"))
	_ = v.BindPFlag("server.read_timeout", cmd.Flags().Lookup("http-read-timeout"))
	_ = v.BindPFlag("server.write_timeout", cmd.Flags().Lookup("http-write-timeout"))
	_ = v.BindPFlag("server.idle_timeout", cmd.Flags().Lookup("http-idle-timeout"))

	// Bind database flags
	_ = v.BindPFlag("database.host", cmd.Flags().Lookup("db-host"))
	_ = v.BindPFlag("database.port", cmd.Flags().Lookup("db-port"))
	_ = v.BindPFlag("database.user", cmd.Flags().Lookup("db-user"))
	_ = v.BindPFlag("database.password", cmd.Flags().Lookup("db-password"))
	_ = v.BindPFlag("database.name", cmd.Flags().Lookup("db-name"))
	_ = v.BindPFlag("database.sslmode", cmd.Flags().Lookup("db-sslmode"))
	_ = v.BindPFlag("database.max_conns", cmd.Flags().Lookup("db-max-conns"))
	_ = v.BindPFlag("database.min_conns", cmd.Flags().Lookup("db-min-conns"))
	_ = v.BindPFlag("database.max_conn_lifetime", cmd.Flags().Lookup("db-max-conn-lifetime"))
	_ = v.BindPFlag("database.max_conn_idle_time", cmd.Flags().Lookup("db-max-conn-idle-time"))
	_ = v.BindPFlag("database.health_check_period", cmd.Flags().Lookup("db-health-check-period"))

	// Bind session flags
	_ = v.BindPFlag("session.duration", cmd.Flags().Lookup("session-duration"))
	_ = v.BindPFlag("session.cleanup_interval", cmd.Flags().Lookup("session-cleanup-interval"))

	// Bind logging flags
	_ = v.BindPFlag("logging.level", cmd.Flags().Lookup("log-level"))
	_ = v.BindPFlag("logging.format", cmd.Flags().Lookup("log-format"))
	_ = v.BindPFlag("logging.output", cmd.Flags().Lookup("log-output"))
}

func runServer(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Setup logging from config
	logging.SetupDefault(cfg.Logging)

	// Connect to database using config
	pool, err := db.NewPoolFromConfig(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	// Create session store with configured duration
	sessionStore := auth.NewSessionStoreWithConfig(cfg.Session.Duration)

	// Start session cleanup goroutine
	go startSessionCleanup(sessionStore, cfg.Session.CleanupInterval)

	// Create router using stdlib ServeMux
	mux := http.NewServeMux()

	// Create Huma API with stdlib adapter
	humaAPI := humago.New(mux, huma.DefaultConfig("Loomio API", "1.0.0"))

	// Create queries instance
	queries := db.New(pool)

	// Create app with dependencies
	app := &App{
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

	// Start server in goroutine
	go func() {
		slog.Info("server starting", "port", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	slog.Info("server exited")
	return nil
}

func startSessionCleanup(store *auth.SessionStore, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		cleaned := store.CleanupExpired()
		if cleaned > 0 {
			slog.Info("cleaned expired sessions", "count", cleaned)
		}
	}
}

// App holds application dependencies for handler registration.
type App struct {
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

	slog.Debug("routes registered")
}
