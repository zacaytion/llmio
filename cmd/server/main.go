// Package main provides the server entrypoint.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"

	"github.com/zacaytion/llmio/internal/api"
	"github.com/zacaytion/llmio/internal/auth"
	"github.com/zacaytion/llmio/internal/db"
)

func main() {
	ctx := context.Background()

	// Load configuration
	dbCfg := db.DefaultConfig()
	port := getEnv("PORT", "8080")

	// Connect to database
	pool, err := db.NewPool(ctx, dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Create session store
	sessionStore := auth.NewSessionStore()

	// Start session cleanup goroutine
	go startSessionCleanup(sessionStore)

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

	// Create server with logging middleware
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      api.LoggingMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func startSessionCleanup(store *auth.SessionStore) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cleaned := store.CleanupExpired()
		if cleaned > 0 {
			log.Printf("Cleaned up %d expired sessions", cleaned)
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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

	fmt.Println("Routes registered")
}
