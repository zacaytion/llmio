//go:build integration

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zacaytion/llmio/internal/api"
	"github.com/zacaytion/llmio/internal/auth"
	"github.com/zacaytion/llmio/internal/db"
	"github.com/zacaytion/llmio/internal/testutil"
)

// testAPISetup holds shared test infrastructure for API integration tests.
type testAPISetup struct {
	pool     *pgxpool.Pool
	queries  *db.Queries
	sessions *auth.SessionStore
	mux      *http.ServeMux
}

// setupAPITest creates a test environment using the shared container.
func setupAPITest(t *testing.T) *testAPISetup {
	t.Helper()

	pool := testutil.GetPool()
	if pool == nil {
		t.Fatal("pool not initialized - TestMain may have failed")
	}

	queries := db.New(pool)
	sessions := auth.NewSessionStore()

	// Create handlers
	groupHandler := api.NewGroupHandler(pool, queries, sessions)
	membershipHandler := api.NewMembershipHandler(pool, queries, sessions)

	// Create Huma API
	mux := http.NewServeMux()
	humaAPI := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))
	groupHandler.RegisterRoutes(humaAPI)
	membershipHandler.RegisterRoutes(humaAPI)

	return &testAPISetup{
		pool:     pool,
		queries:  queries,
		sessions: sessions,
		mux:      mux,
	}
}

// createTestUser creates a test user directly in the database.
func (s *testAPISetup) createTestUser(t *testing.T, email, name string) *db.User {
	t.Helper()
	ctx := context.Background()

	hash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		Name:         name,
		Username:     auth.GenerateUsername(name),
		PasswordHash: hash,
		Key:          auth.GeneratePublicKey(),
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Mark user as verified
	_, err = s.pool.Exec(ctx, "UPDATE users SET email_verified = true WHERE id = $1", user.ID)
	if err != nil {
		t.Fatalf("failed to verify user: %v", err)
	}

	return user
}

// createTestSession creates a session for a user.
func (s *testAPISetup) createTestSession(t *testing.T, userID int64) string {
	t.Helper()
	session, err := s.sessions.Create(userID, "", "")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	return session.Token
}

// Test_CreateGroup_ValidInput tests successful group creation.
func Test_CreateGroup_ValidInput(t *testing.T) {
	t.Cleanup(func() { testutil.Restore(t) })

	setup := setupAPITest(t)
	user := setup.createTestUser(t, "creator@example.com", "Creator")
	token := setup.createTestSession(t, user.ID)

	body := map[string]any{
		"name":        "Test Group",
		"handle":      "test-group",
		"description": "A test group",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: token})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var resp struct {
		Group struct {
			Handle string `json:"handle"`
		} `json:"group"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Group.Handle != "test-group" {
		t.Errorf("expected handle 'test-group', got %v", resp.Group.Handle)
	}
}

// Test_CreateGroup_DuplicateHandle tests conflict on duplicate handle.
func Test_CreateGroup_DuplicateHandle(t *testing.T) {
	t.Cleanup(func() { testutil.Restore(t) })

	setup := setupAPITest(t)
	user := setup.createTestUser(t, "creator@example.com", "Creator")
	token := setup.createTestSession(t, user.ID)

	// Create first group
	body := map[string]any{
		"name":   "First Group",
		"handle": "duplicate-handle",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: token})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("first group creation failed: %d - %s", w.Code, w.Body.String())
	}

	// Try to create second group with same handle
	body["name"] = "Second Group"
	bodyBytes, _ = json.Marshal(body)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: token})

	w = httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d for duplicate handle, got %d: %s",
			http.StatusConflict, w.Code, w.Body.String())
	}
}

// Test_CreateGroup_Unauthenticated tests rejection without session.
func Test_CreateGroup_Unauthenticated(t *testing.T) {
	t.Cleanup(func() { testutil.Restore(t) })

	setup := setupAPITest(t)

	body := map[string]any{
		"name":   "Test Group",
		"handle": "test-group",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	// No session cookie

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d: %s", http.StatusUnauthorized, w.Code, w.Body.String())
	}
}
