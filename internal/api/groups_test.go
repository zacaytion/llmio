package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zacaytion/llmio/internal/auth"
	"github.com/zacaytion/llmio/internal/db"
	"github.com/zacaytion/llmio/internal/testutil"
)

// testGroupsSetup holds shared test infrastructure for group tests.
type testGroupsSetup struct {
	pool         *pgxpool.Pool
	queries      *db.Queries
	sessions     *auth.SessionStore
	groupHandler *GroupHandler
	mux          *http.ServeMux
	cleanup      func()
}

// setupGroupsTest creates a test environment with a real database container.
func setupGroupsTest(t *testing.T) *testGroupsSetup {
	t.Helper()
	ctx := context.Background()

	// Create postgres container with migrations
	connStr, cleanup := testutil.SetupTestDB(ctx, t)

	// Create connection pool
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		cleanup()
		t.Fatalf("failed to create pool: %v", err)
	}

	queries := db.New(pool)
	sessions := auth.NewSessionStore()

	// Create handlers
	groupHandler := NewGroupHandler(pool, queries, sessions)
	membershipHandler := NewMembershipHandler(pool, queries, sessions)

	// Create Huma API
	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))
	groupHandler.RegisterRoutes(api)
	membershipHandler.RegisterRoutes(api)

	return &testGroupsSetup{
		pool:         pool,
		queries:      queries,
		sessions:     sessions,
		groupHandler: groupHandler,
		mux:          mux,
		cleanup: func() {
			pool.Close()
			cleanup()
		},
	}
}

// createTestUser creates a test user directly in the database.
func (s *testGroupsSetup) createTestUser(t *testing.T, email, name string) *db.User {
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

// createTestSession creates a session for the given user and returns the token.
func (s *testGroupsSetup) createTestSession(t *testing.T, userID int64) string {
	t.Helper()
	session, err := s.sessions.Create(userID, "", "")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	return session.Token
}

// ============================================================
// T016-T020d: User Story 1 - Create Group Tests (TDD)
// ============================================================

// TestCreateGroup_TableDriven is the main table-driven test for createGroup handler.
// T016: Write table-driven tests for createGroup handler
func TestCreateGroup_TableDriven(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	// Create test users
	user1 := setup.createTestUser(t, "user1@example.com", "User One")
	user1Token := setup.createTestSession(t, user1.ID)

	tests := []struct {
		name           string
		cookie         string
		body           map[string]any
		wantStatus     int
		wantHandle     string // expected handle in response (if applicable)
		wantErrMessage string // expected error message substring
	}{
		// T017: Test authenticated user creates group → 201 + group returned with auto-generated handle
		{
			name:       "authenticated user creates group with auto-generated handle",
			cookie:     user1Token,
			body:       map[string]any{"name": "My First Group", "description": "Test group"},
			wantStatus: http.StatusCreated,
			wantHandle: "my-first-group",
		},
		// T018: Test authenticated user creates group with custom handle → 201 + handle preserved
		{
			name:       "authenticated user creates group with custom handle",
			cookie:     user1Token,
			body:       map[string]any{"name": "Second Group", "handle": "custom-handle"},
			wantStatus: http.StatusCreated,
			wantHandle: "custom-handle",
		},
		// T020: Test unauthenticated → 401 Unauthorized
		{
			name:           "unauthenticated request returns 401",
			cookie:         "",
			body:           map[string]any{"name": "No Auth Group"},
			wantStatus:     http.StatusUnauthorized,
			wantErrMessage: "Not authenticated",
		},
		{
			name:           "invalid session returns 401",
			cookie:         "invalid-token",
			body:           map[string]any{"name": "Bad Token Group"},
			wantStatus:     http.StatusUnauthorized,
			wantErrMessage: "Not authenticated",
		},
		// T020d: Test empty name rejected with 422 (handle generation requires name)
		// Note: Huma validation returns "expected length >= 1" before our custom validation
		{
			name:           "empty name returns 422",
			cookie:         user1Token,
			body:           map[string]any{"name": ""},
			wantStatus:     http.StatusUnprocessableEntity,
			wantErrMessage: "", // Empty - Huma provides its own validation message
		},
		{
			name:           "whitespace-only name returns 422",
			cookie:         user1Token,
			body:           map[string]any{"name": "   "},
			wantStatus:     http.StatusUnprocessableEntity,
			wantErrMessage: "Name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			if tt.cookie != "" {
				req.AddCookie(&http.Cookie{Name: "loomio_session", Value: tt.cookie})
			}

			w := httptest.NewRecorder()
			setup.mux.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d: %s", tt.wantStatus, w.Code, w.Body.String())
				return
			}

			if tt.wantStatus == http.StatusCreated {
				var resp map[string]any
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}

				group, ok := resp["group"].(map[string]any)
				if !ok {
					t.Fatalf("response missing group object: %v", resp)
				}

				if handle, ok := group["handle"].(string); ok {
					if tt.wantHandle != "" && handle != tt.wantHandle {
						t.Errorf("expected handle %q, got %q", tt.wantHandle, handle)
					}
				} else if tt.wantHandle != "" {
					t.Errorf("expected handle in response, got none")
				}

				// Verify creator became admin
				if id, ok := group["id"].(float64); ok {
					membership, err := setup.queries.GetMembershipByGroupAndUser(context.Background(), db.GetMembershipByGroupAndUserParams{
						GroupID: int64(id),
						UserID:  user1.ID,
					})
					if err != nil {
						t.Errorf("creator should have membership: %v", err)
					} else if membership.Role != "admin" {
						t.Errorf("creator should be admin, got role %q", membership.Role)
					}
				}
			}

			if tt.wantErrMessage != "" {
				if !bytes.Contains(w.Body.Bytes(), []byte(tt.wantErrMessage)) {
					t.Errorf("expected error containing %q, got: %s", tt.wantErrMessage, w.Body.String())
				}
			}
		})
	}
}

// TestCreateGroup_HandleConflict tests that handle conflicts return 409.
// T019: Test handle conflict → 409 Conflict
func TestCreateGroup_HandleConflict(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	user := setup.createTestUser(t, "conflict-user@example.com", "Conflict User")
	token := setup.createTestSession(t, user.ID)

	// Create first group with explicit handle
	body1 := map[string]any{"name": "First Group", "handle": "taken-handle"}
	bodyBytes1, _ := json.Marshal(body1)
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewBuffer(bodyBytes1))
	req1.Header.Set("Content-Type", "application/json")
	req1.AddCookie(&http.Cookie{Name: "loomio_session", Value: token})

	w1 := httptest.NewRecorder()
	setup.mux.ServeHTTP(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("first group creation failed: %d: %s", w1.Code, w1.Body.String())
	}

	// Try to create second group with same handle → should fail with 409
	body2 := map[string]any{"name": "Second Group", "handle": "taken-handle"}
	bodyBytes2, _ := json.Marshal(body2)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewBuffer(bodyBytes2))
	req2.Header.Set("Content-Type", "application/json")
	req2.AddCookie(&http.Cookie{Name: "loomio_session", Value: token})

	w2 := httptest.NewRecorder()
	setup.mux.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict for duplicate handle, got %d: %s", w2.Code, w2.Body.String())
	}

	if !bytes.Contains(w2.Body.Bytes(), []byte("Handle already taken")) {
		t.Errorf("expected error message about handle conflict, got: %s", w2.Body.String())
	}
}

// TestCreateGroup_HandleAutoGeneration tests handle auto-generation from name.
// T020a: Test handle auto-generated from name with spaces → "my group" becomes "my-group"
// T020b: Test handle auto-generated from name with special chars → "Team @#$% 2026" becomes "team-2026"
func TestCreateGroup_HandleAutoGeneration(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	user := setup.createTestUser(t, "autogen-user@example.com", "AutoGen User")
	token := setup.createTestSession(t, user.ID)

	tests := []struct {
		name       string
		groupName  string
		wantHandle string
	}{
		// T020a
		{
			name:       "spaces become hyphens",
			groupName:  "my group",
			wantHandle: "my-group",
		},
		// T020b
		{
			name:       "special chars removed",
			groupName:  "Team @#$% 2026",
			wantHandle: "team-2026",
		},
		{
			name:       "accented characters transliterated",
			groupName:  "Café München",
			wantHandle: "cafe-munchen",
		},
		{
			name:       "uppercase becomes lowercase",
			groupName:  "CLIMATE ACTION",
			wantHandle: "climate-action",
		},
		{
			name:       "multiple spaces collapse",
			groupName:  "Too   Many   Spaces",
			wantHandle: "too-many-spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the group name directly - the test names are unique enough
			body := map[string]any{"name": tt.groupName}
			bodyBytes, _ := json.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(&http.Cookie{Name: "loomio_session", Value: token})

			w := httptest.NewRecorder()
			setup.mux.ServeHTTP(w, req)

			if w.Code != http.StatusCreated {
				t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
			}

			var resp map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			group := resp["group"].(map[string]any)
			handle := group["handle"].(string)

			// Handle should match expected pattern
			if handle != tt.wantHandle {
				t.Errorf("expected handle %q, got %q", tt.wantHandle, handle)
			}
		})
	}
}

// TestCreateGroup_HandleCollisionRetry tests automatic suffix on collision.
// T020c: Test handle auto-generated collision retry
func TestCreateGroup_HandleCollisionRetry(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	user := setup.createTestUser(t, "collision-user@example.com", "Collision User")
	token := setup.createTestSession(t, user.ID)

	// Create first group named "Climate Team" → expect handle "climate-team"
	body1 := map[string]any{"name": "Climate Team"}
	bodyBytes1, _ := json.Marshal(body1)
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewBuffer(bodyBytes1))
	req1.Header.Set("Content-Type", "application/json")
	req1.AddCookie(&http.Cookie{Name: "loomio_session", Value: token})

	w1 := httptest.NewRecorder()
	setup.mux.ServeHTTP(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("first group creation failed: %d: %s", w1.Code, w1.Body.String())
	}

	var resp1 map[string]any
	if err := json.Unmarshal(w1.Body.Bytes(), &resp1); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	group1 := resp1["group"].(map[string]any)
	handle1 := group1["handle"].(string)

	if handle1 != "climate-team" {
		t.Errorf("first group should have handle 'climate-team', got %q", handle1)
	}

	// Create second group with same name → expect handle "climate-team-1"
	body2 := map[string]any{"name": "Climate Team"}
	bodyBytes2, _ := json.Marshal(body2)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewBuffer(bodyBytes2))
	req2.Header.Set("Content-Type", "application/json")
	req2.AddCookie(&http.Cookie{Name: "loomio_session", Value: token})

	w2 := httptest.NewRecorder()
	setup.mux.ServeHTTP(w2, req2)

	if w2.Code != http.StatusCreated {
		t.Fatalf("second group creation failed: %d: %s", w2.Code, w2.Body.String())
	}

	var resp2 map[string]any
	if err := json.Unmarshal(w2.Body.Bytes(), &resp2); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	group2 := resp2["group"].(map[string]any)
	handle2 := group2["handle"].(string)

	if handle2 != "climate-team-1" {
		t.Errorf("second group should have handle 'climate-team-1', got %q", handle2)
	}

	// Create third group with same name → expect handle "climate-team-2"
	body3 := map[string]any{"name": "Climate Team"}
	bodyBytes3, _ := json.Marshal(body3)
	req3 := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewBuffer(bodyBytes3))
	req3.Header.Set("Content-Type", "application/json")
	req3.AddCookie(&http.Cookie{Name: "loomio_session", Value: token})

	w3 := httptest.NewRecorder()
	setup.mux.ServeHTTP(w3, req3)

	if w3.Code != http.StatusCreated {
		t.Fatalf("third group creation failed: %d: %s", w3.Code, w3.Body.String())
	}

	var resp3 map[string]any
	if err := json.Unmarshal(w3.Body.Bytes(), &resp3); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	group3 := resp3["group"].(map[string]any)
	handle3 := group3["handle"].(string)

	if handle3 != "climate-team-2" {
		t.Errorf("third group should have handle 'climate-team-2', got %q", handle3)
	}
}

// TestCreateGroup_CreatorBecomesAdmin verifies that the creator automatically becomes an admin.
// This is part of T017 but explicitly tests the membership creation.
func TestCreateGroup_CreatorBecomesAdmin(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	user := setup.createTestUser(t, "admin-test@example.com", "Admin Test User")
	token := setup.createTestSession(t, user.ID)

	body := map[string]any{"name": "Admin Test Group"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: token})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	group := resp["group"].(map[string]any)
	groupID := int64(group["id"].(float64))

	// Verify membership
	membership, err := setup.queries.GetMembershipByGroupAndUser(context.Background(), db.GetMembershipByGroupAndUserParams{
		GroupID: groupID,
		UserID:  user.ID,
	})
	if err != nil {
		t.Fatalf("membership not found: %v", err)
	}

	if membership.Role != "admin" {
		t.Errorf("expected role 'admin', got %q", membership.Role)
	}

	// Membership should be auto-accepted (creator doesn't need to accept their own invitation)
	if !membership.AcceptedAt.Valid {
		t.Error("creator's membership should be auto-accepted")
	}

	// Inviter should be self
	if membership.InviterID != user.ID {
		t.Errorf("expected inviter_id %d, got %d", user.ID, membership.InviterID)
	}
}

// ============================================================
// T066-T071: User Story 4 - Configure Group Permissions Tests (TDD)
// ============================================================

// TestUpdateGroup_PermissionFlags tests updating group permission flags.
// T066-T067: Write table-driven tests for updateGroup handler
func TestUpdateGroup_PermissionFlags(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	groupID := setup.createTestGroupAndGetID(t, adminToken, "Permission Test Group")

	// Update permission flags
	body := map[string]any{
		"members_can_add_members":      false,
		"members_can_create_subgroups": false,
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/groups/%d", groupID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	group := resp["group"].(map[string]any)
	if group["members_can_add_members"].(bool) != false {
		t.Error("members_can_add_members should be false")
	}
	if group["members_can_create_subgroups"].(bool) != false {
		t.Error("members_can_create_subgroups should be false")
	}
}

// TestUpdateGroup_NonAdmin tests that non-admins cannot update groups.
// T068: Test non-admin tries to update → 403 Forbidden
func TestUpdateGroup_NonAdmin(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	memberUser := setup.createTestUser(t, "member@example.com", "Member User")
	memberToken := setup.createTestSession(t, memberUser.ID)

	groupID := setup.createTestGroupAndGetID(t, adminToken, "Test Group")

	// Invite member and accept
	setup.inviteMember(t, adminToken, groupID, memberUser.ID)
	setup.acceptInvitation(t, memberToken, memberUser.ID, groupID)

	// Member tries to update (should fail)
	body := map[string]any{"name": "Hacked Name"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/groups/%d", groupID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: memberToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got %d: %s", w.Code, w.Body.String())
	}
}

// TestInviteMember_PermissionFlag tests that members_can_add_members flag is enforced.
// T069-T070: Test members_can_add_members enforcement
func TestInviteMember_PermissionFlag(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	memberUser := setup.createTestUser(t, "member@example.com", "Member User")
	memberToken := setup.createTestSession(t, memberUser.ID)

	inviteeUser := setup.createTestUser(t, "invitee@example.com", "Invitee User")

	groupID := setup.createTestGroupAndGetID(t, adminToken, "Test Group")

	// Invite member and accept
	setup.inviteMember(t, adminToken, groupID, memberUser.ID)
	setup.acceptInvitation(t, memberToken, memberUser.ID, groupID)

	// By default, members_can_add_members is true, so member can invite
	body := map[string]any{"user_id": inviteeUser.ID, "role": "member"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: memberToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("member should be able to invite when members_can_add_members=true, got %d: %s", w.Code, w.Body.String())
	}

	// Now disable members_can_add_members
	updateBody := map[string]any{"members_can_add_members": false}
	updateBytes, _ := json.Marshal(updateBody)
	updateReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/groups/%d", groupID), bytes.NewBuffer(updateBytes))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})
	updateW := httptest.NewRecorder()
	setup.mux.ServeHTTP(updateW, updateReq)
	if updateW.Code != http.StatusOK {
		t.Fatalf("failed to update group: %d: %s", updateW.Code, updateW.Body.String())
	}

	// Create another invitee
	invitee2User := setup.createTestUser(t, "invitee2@example.com", "Invitee 2")

	// Now member cannot invite
	body2 := map[string]any{"user_id": invitee2User.ID, "role": "member"}
	bodyBytes2, _ := json.Marshal(body2)
	req2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), bytes.NewBuffer(bodyBytes2))
	req2.Header.Set("Content-Type", "application/json")
	req2.AddCookie(&http.Cookie{Name: "loomio_session", Value: memberToken})

	w2 := httptest.NewRecorder()
	setup.mux.ServeHTTP(w2, req2)

	if w2.Code != http.StatusForbidden {
		t.Errorf("member should NOT be able to invite when members_can_add_members=false, got %d: %s", w2.Code, w2.Body.String())
	}
}

// TestInviteMember_AdminBypass tests that admins can invite even when members_can_add_members=false.
// T070a: Test admin bypass
func TestInviteMember_AdminBypass(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	inviteeUser := setup.createTestUser(t, "invitee@example.com", "Invitee User")

	groupID := setup.createTestGroupAndGetID(t, adminToken, "Test Group")

	// Disable members_can_add_members
	updateBody := map[string]any{"members_can_add_members": false}
	updateBytes, _ := json.Marshal(updateBody)
	updateReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/groups/%d", groupID), bytes.NewBuffer(updateBytes))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})
	updateW := httptest.NewRecorder()
	setup.mux.ServeHTTP(updateW, updateReq)
	if updateW.Code != http.StatusOK {
		t.Fatalf("failed to update group: %d: %s", updateW.Code, updateW.Body.String())
	}

	// Admin can still invite (FR-022 admin bypass)
	body := map[string]any{"user_id": inviteeUser.ID, "role": "member"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("admin should always be able to invite (FR-022), got %d: %s", w.Code, w.Body.String())
	}
}

// TestGetGroup_PermissionFlags tests that getGroup returns all permission flags.
// T071: Test getGroup returns all 11 permission flags
func TestGetGroup_PermissionFlags(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	groupID := setup.createTestGroupAndGetID(t, adminToken, "Test Group")

	// Get group
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/groups/%d", groupID), nil)
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	group := resp["group"].(map[string]any)

	// Check all 11 permission flags exist
	permissionFlags := []string{
		"members_can_add_members",
		"members_can_add_guests",
		"members_can_start_discussions",
		"members_can_raise_motions",
		"members_can_edit_discussions",
		"members_can_edit_comments",
		"members_can_delete_comments",
		"members_can_announce",
		"members_can_create_subgroups",
		"admins_can_edit_user_content",
		"parent_members_can_see_discussions",
	}

	for _, flag := range permissionFlags {
		if _, ok := group[flag]; !ok {
			t.Errorf("missing permission flag: %s", flag)
		}
	}

	// Check counts exist
	if _, ok := group["member_count"]; !ok {
		t.Error("missing member_count")
	}
	if _, ok := group["admin_count"]; !ok {
		t.Error("missing admin_count")
	}
	if _, ok := group["current_user_role"]; !ok {
		t.Error("missing current_user_role")
	}
}

// Helper functions for tests

// createTestGroupAndGetID creates a group and returns its ID.
func (s *testGroupsSetup) createTestGroupAndGetID(t *testing.T, token, name string) int64 {
	t.Helper()

	body := map[string]any{"name": name}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: token})

	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create group: %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	group := resp["group"].(map[string]any)
	return int64(group["id"].(float64))
}

// inviteMember invites a user to a group.
func (s *testGroupsSetup) inviteMember(t *testing.T, adminToken string, groupID, userID int64) int64 {
	t.Helper()

	body := map[string]any{"user_id": userID, "role": "member"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("failed to invite: %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	return int64(resp["membership"].(map[string]any)["id"].(float64))
}

// acceptInvitation accepts a pending invitation.
func (s *testGroupsSetup) acceptInvitation(t *testing.T, token string, userID, groupID int64) {
	t.Helper()

	// Get membership ID
	membership, err := s.queries.GetMembershipByGroupAndUser(context.Background(), db.GetMembershipByGroupAndUserParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		t.Fatalf("failed to get membership: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/accept", membership.ID), nil)
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: token})

	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("failed to accept: %d: %s", w.Code, w.Body.String())
	}
}

// ============================================================
// T080-T085b: User Story 5 - Create Subgroups Tests (TDD)
// ============================================================

// TestCreateSubgroup_AdminCreates tests that an admin can create a subgroup.
// T080-T081: Write table-driven tests for createSubgroup handler
func TestCreateSubgroup_AdminCreates(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	// Create parent group
	parentGroupID := setup.createTestGroupAndGetID(t, adminToken, "Parent Group")

	// Create subgroup
	body := map[string]any{"name": "Child Group", "description": "A subgroup"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/subgroups", parentGroupID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	group := resp["group"].(map[string]any)

	// Verify parent_id is set correctly
	if parentID, ok := group["parent_id"].(float64); !ok || int64(parentID) != parentGroupID {
		t.Errorf("expected parent_id %d, got %v", parentGroupID, group["parent_id"])
	}

	// Verify creator became admin of subgroup
	subgroupID := int64(group["id"].(float64))
	membership, err := setup.queries.GetMembershipByGroupAndUser(context.Background(), db.GetMembershipByGroupAndUserParams{
		GroupID: subgroupID,
		UserID:  adminUser.ID,
	})
	if err != nil {
		t.Fatalf("creator should have membership: %v", err)
	}
	if membership.Role != "admin" {
		t.Errorf("creator should be admin, got role %q", membership.Role)
	}
}

// TestCreateSubgroup_MemberWithPermission tests that a member with permission can create subgroups.
// T082: Test member with permission creates subgroup → allowed
func TestCreateSubgroup_MemberWithPermission(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	memberUser := setup.createTestUser(t, "member@example.com", "Member User")
	memberToken := setup.createTestSession(t, memberUser.ID)

	// Create parent group
	parentGroupID := setup.createTestGroupAndGetID(t, adminToken, "Parent Group")

	// Invite member and accept
	setup.inviteMember(t, adminToken, parentGroupID, memberUser.ID)
	setup.acceptInvitation(t, memberToken, memberUser.ID, parentGroupID)

	// Enable members_can_create_subgroups
	updateBody := map[string]any{"members_can_create_subgroups": true}
	updateBytes, _ := json.Marshal(updateBody)
	updateReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/groups/%d", parentGroupID), bytes.NewBuffer(updateBytes))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})
	updateW := httptest.NewRecorder()
	setup.mux.ServeHTTP(updateW, updateReq)
	if updateW.Code != http.StatusOK {
		t.Fatalf("failed to update group: %d: %s", updateW.Code, updateW.Body.String())
	}

	// Member creates subgroup (should succeed)
	body := map[string]any{"name": "Member's Subgroup"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/subgroups", parentGroupID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: memberToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("member with permission should create subgroup, got %d: %s", w.Code, w.Body.String())
	}
}

// TestCreateSubgroup_MemberWithoutPermission tests that a member without permission cannot create subgroups.
// T083: Test member without permission → 403 Forbidden
func TestCreateSubgroup_MemberWithoutPermission(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	memberUser := setup.createTestUser(t, "member@example.com", "Member User")
	memberToken := setup.createTestSession(t, memberUser.ID)

	// Create parent group (members_can_create_subgroups defaults to false)
	parentGroupID := setup.createTestGroupAndGetID(t, adminToken, "Parent Group")

	// Invite member and accept
	setup.inviteMember(t, adminToken, parentGroupID, memberUser.ID)
	setup.acceptInvitation(t, memberToken, memberUser.ID, parentGroupID)

	// Member tries to create subgroup (should fail)
	body := map[string]any{"name": "Unauthorized Subgroup"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/subgroups", parentGroupID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: memberToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got %d: %s", w.Code, w.Body.String())
	}
}

// TestCreateSubgroup_AdminBypass tests that admin can create subgroups even when permission is disabled.
// T083a: Test admin bypass
func TestCreateSubgroup_AdminBypass(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	// Create parent group (members_can_create_subgroups defaults to false)
	parentGroupID := setup.createTestGroupAndGetID(t, adminToken, "Parent Group")

	// Admin creates subgroup (should succeed even with flag=false)
	body := map[string]any{"name": "Admin Subgroup"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/subgroups", parentGroupID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("admin should always create subgroup (FR-022), got %d: %s", w.Code, w.Body.String())
	}
}

// TestListSubgroups tests that listSubgroups returns child groups.
// T084: Test listSubgroups returns child groups
func TestListSubgroups(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	// Create parent group
	parentGroupID := setup.createTestGroupAndGetID(t, adminToken, "Parent Group")

	// Create two subgroups
	setup.createSubgroup(t, adminToken, parentGroupID, "Subgroup One")
	setup.createSubgroup(t, adminToken, parentGroupID, "Subgroup Two")

	// List subgroups
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/groups/%d/subgroups", parentGroupID), nil)
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	groups := resp["groups"].([]any)
	if len(groups) != 2 {
		t.Errorf("expected 2 subgroups, got %d", len(groups))
	}
}

// TestCreateSubgroup_SelfRefBlocked tests that a subgroup cannot be its own parent.
// T085: Test subgroup cannot be its own parent (self-ref blocked)
func TestCreateSubgroup_SelfRefBlocked(t *testing.T) {
	// This is actually blocked at the database level by CONSTRAINT groups_parent_not_self
	// The API doesn't allow setting parent_id to self because you create subgroups via POST to parent
	// So this test verifies the database constraint by checking that if somehow parent_id=id, it fails
	// In practice, this constraint is tested in pgTap tests (003_groups_test.sql)
	t.Skip("Self-reference is blocked by DB constraint; tested in pgTap tests")
}

// TestCreateSubgroup_InheritPermissions tests that subgroups can inherit parent permissions.
// T085a: Test subgroup with inherit_permissions=true copies parent permission flags
func TestCreateSubgroup_InheritPermissions(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	// Create parent group with custom permissions
	parentGroupID := setup.createTestGroupAndGetID(t, adminToken, "Parent Group")

	// Set custom permissions on parent
	updateBody := map[string]any{
		"members_can_add_members":      false,
		"members_can_create_subgroups": true,
		"members_can_announce":         true,
	}
	updateBytes, _ := json.Marshal(updateBody)
	updateReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/groups/%d", parentGroupID), bytes.NewBuffer(updateBytes))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})
	updateW := httptest.NewRecorder()
	setup.mux.ServeHTTP(updateW, updateReq)
	if updateW.Code != http.StatusOK {
		t.Fatalf("failed to update group: %d: %s", updateW.Code, updateW.Body.String())
	}

	// Create subgroup with inherit_permissions=true
	body := map[string]any{"name": "Inherited Subgroup", "inherit_permissions": true}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/subgroups", parentGroupID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	group := resp["group"].(map[string]any)

	// Verify inherited permissions
	if group["members_can_add_members"].(bool) != false {
		t.Error("members_can_add_members should be inherited as false")
	}
	if group["members_can_create_subgroups"].(bool) != true {
		t.Error("members_can_create_subgroups should be inherited as true")
	}
	if group["members_can_announce"].(bool) != true {
		t.Error("members_can_announce should be inherited as true")
	}
}

// TestCreateSubgroup_DefaultPermissions tests that subgroups use defaults without inheritance.
// T085b: Test subgroup with inherit_permissions=false uses default permission flags
func TestCreateSubgroup_DefaultPermissions(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	// Create parent group with custom permissions
	parentGroupID := setup.createTestGroupAndGetID(t, adminToken, "Parent Group")

	// Set custom permissions on parent (non-default values)
	updateBody := map[string]any{
		"members_can_add_members":      false,
		"members_can_create_subgroups": true,
	}
	updateBytes, _ := json.Marshal(updateBody)
	updateReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/groups/%d", parentGroupID), bytes.NewBuffer(updateBytes))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})
	updateW := httptest.NewRecorder()
	setup.mux.ServeHTTP(updateW, updateReq)
	if updateW.Code != http.StatusOK {
		t.Fatalf("failed to update group: %d: %s", updateW.Code, updateW.Body.String())
	}

	// Create subgroup WITHOUT inherit_permissions (defaults to false)
	body := map[string]any{"name": "Default Subgroup"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/subgroups", parentGroupID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	group := resp["group"].(map[string]any)

	// Verify DEFAULT permissions (not inherited)
	// Default: members_can_add_members=true, members_can_create_subgroups=false
	if group["members_can_add_members"].(bool) != true {
		t.Error("members_can_add_members should be default (true), not inherited")
	}
	if group["members_can_create_subgroups"].(bool) != false {
		t.Error("members_can_create_subgroups should be default (false), not inherited")
	}
}

// Helper: createSubgroup creates a subgroup and returns its ID.
func (s *testGroupsSetup) createSubgroup(t *testing.T, token string, parentID int64, name string) int64 {
	t.Helper()

	body := map[string]any{"name": name}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/subgroups", parentID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: token})

	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create subgroup: %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	group := resp["group"].(map[string]any)
	return int64(group["id"].(float64))
}

// ============================================================
// T093-T099a: User Story 6 - Archive Group Tests (TDD)
// ============================================================

// TestArchiveGroup_AdminArchives tests that an admin can archive a group.
// T093-T094: Write table-driven tests for archiveGroup handler
func TestArchiveGroup_AdminArchives(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	groupID := setup.createTestGroupAndGetID(t, adminToken, "Archive Test Group")

	// Archive the group
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/archive", groupID), nil)
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	group := resp["group"].(map[string]any)

	// Verify archived_at is set
	if group["archived_at"] == nil {
		t.Error("archived_at should be set")
	}
}

// TestArchiveGroup_NonAdmin tests that non-admins cannot archive groups.
// T095: Test non-admin tries to archive → 403 Forbidden
func TestArchiveGroup_NonAdmin(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	memberUser := setup.createTestUser(t, "member@example.com", "Member User")
	memberToken := setup.createTestSession(t, memberUser.ID)

	groupID := setup.createTestGroupAndGetID(t, adminToken, "Archive Test Group")

	// Invite member and accept
	setup.inviteMember(t, adminToken, groupID, memberUser.ID)
	setup.acceptInvitation(t, memberToken, memberUser.ID, groupID)

	// Member tries to archive (should fail)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/archive", groupID), nil)
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: memberToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got %d: %s", w.Code, w.Body.String())
	}
}

// TestListGroups_ExcludesArchived tests that archived groups are excluded by default.
// T096: Test archived group excluded from listGroups (default)
func TestListGroups_ExcludesArchived(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	// Create two groups
	setup.createTestGroupAndGetID(t, adminToken, "Active Group")
	archivedGroupID := setup.createTestGroupAndGetID(t, adminToken, "Archived Group")

	// Archive one group
	archiveReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/archive", archivedGroupID), nil)
	archiveReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})
	archiveW := httptest.NewRecorder()
	setup.mux.ServeHTTP(archiveW, archiveReq)
	if archiveW.Code != http.StatusOK {
		t.Fatalf("failed to archive group: %d: %s", archiveW.Code, archiveW.Body.String())
	}

	// List groups (default - exclude archived)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups", nil)
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	groups := resp["groups"].([]any)
	if len(groups) != 1 {
		t.Errorf("expected 1 active group, got %d", len(groups))
	}

	// Verify it's the active group
	firstGroup := groups[0].(map[string]any)
	if firstGroup["name"].(string) != "Active Group" {
		t.Errorf("expected 'Active Group', got %q", firstGroup["name"])
	}
}

// TestListGroups_IncludesArchived tests that archived groups can be included.
// T097: Test archived group included with include_archived=true
func TestListGroups_IncludesArchived(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	// Create two groups
	setup.createTestGroupAndGetID(t, adminToken, "Active Group")
	archivedGroupID := setup.createTestGroupAndGetID(t, adminToken, "Archived Group")

	// Archive one group
	archiveReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/archive", archivedGroupID), nil)
	archiveReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})
	archiveW := httptest.NewRecorder()
	setup.mux.ServeHTTP(archiveW, archiveReq)
	if archiveW.Code != http.StatusOK {
		t.Fatalf("failed to archive group: %d: %s", archiveW.Code, archiveW.Body.String())
	}

	// List groups with include_archived=true
	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups?include_archived=true", nil)
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	groups := resp["groups"].([]any)
	if len(groups) != 2 {
		t.Errorf("expected 2 groups (including archived), got %d", len(groups))
	}
}

// TestUnarchiveGroup tests that unarchive sets archived_at to NULL.
// T098: Test unarchive sets archived_at to NULL
func TestUnarchiveGroup(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	groupID := setup.createTestGroupAndGetID(t, adminToken, "Unarchive Test Group")

	// Archive the group
	archiveReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/archive", groupID), nil)
	archiveReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})
	archiveW := httptest.NewRecorder()
	setup.mux.ServeHTTP(archiveW, archiveReq)
	if archiveW.Code != http.StatusOK {
		t.Fatalf("failed to archive group: %d: %s", archiveW.Code, archiveW.Body.String())
	}

	// Unarchive the group
	unarchiveReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/unarchive", groupID), nil)
	unarchiveReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	unarchiveW := httptest.NewRecorder()
	setup.mux.ServeHTTP(unarchiveW, unarchiveReq)

	if unarchiveW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", unarchiveW.Code, unarchiveW.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(unarchiveW.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	group := resp["group"].(map[string]any)

	// Verify archived_at is NULL (not present in JSON)
	if group["archived_at"] != nil {
		t.Errorf("archived_at should be NULL after unarchive, got %v", group["archived_at"])
	}
}

// TestGetGroup_ArchivedAccessible tests that archived groups are still accessible by ID.
// T099: Test archived group accessible via getGroup/getGroupByHandle
func TestGetGroup_ArchivedAccessible(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	groupID := setup.createTestGroupAndGetID(t, adminToken, "Archived Accessible Group")

	// Archive the group
	archiveReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/archive", groupID), nil)
	archiveReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})
	archiveW := httptest.NewRecorder()
	setup.mux.ServeHTTP(archiveW, archiveReq)
	if archiveW.Code != http.StatusOK {
		t.Fatalf("failed to archive group: %d: %s", archiveW.Code, archiveW.Body.String())
	}

	// Get group by ID (should still work)
	getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/groups/%d", groupID), nil)
	getReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	getW := httptest.NewRecorder()
	setup.mux.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Errorf("expected 200 OK for archived group by ID, got %d: %s", getW.Code, getW.Body.String())
	}

	// Get group by handle (should still work)
	handleReq := httptest.NewRequest(http.MethodGet, "/api/v1/group-by-handle/archived-accessible-group", nil)
	handleReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	handleW := httptest.NewRecorder()
	setup.mux.ServeHTTP(handleW, handleReq)

	if handleW.Code != http.StatusOK {
		t.Errorf("expected 200 OK for archived group by handle, got %d: %s", handleW.Code, handleW.Body.String())
	}
}

// TestSubgroup_ArchivedParent tests that subgroups show parent_archived indicator.
// T099a: Test subgroup with archived parent shows parent relationship as archived
func TestSubgroup_ArchivedParent(t *testing.T) {
	setup := setupGroupsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	// Create parent group
	parentGroupID := setup.createTestGroupAndGetID(t, adminToken, "Parent Group")

	// Create subgroup
	subgroupID := setup.createSubgroup(t, adminToken, parentGroupID, "Child Group")

	// Archive parent group
	archiveReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/archive", parentGroupID), nil)
	archiveReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})
	archiveW := httptest.NewRecorder()
	setup.mux.ServeHTTP(archiveW, archiveReq)
	if archiveW.Code != http.StatusOK {
		t.Fatalf("failed to archive parent group: %d: %s", archiveW.Code, archiveW.Body.String())
	}

	// Get subgroup and check for parent_archived indicator
	getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/groups/%d", subgroupID), nil)
	getReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	getW := httptest.NewRecorder()
	setup.mux.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", getW.Code, getW.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(getW.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	group := resp["group"].(map[string]any)

	// Check for parent_archived indicator
	if parentArchived, ok := group["parent_archived"].(bool); !ok || !parentArchived {
		t.Error("parent_archived should be true when parent is archived")
	}
}
