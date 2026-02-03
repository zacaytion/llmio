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

// testMembershipsSetup holds shared test infrastructure for membership tests.
type testMembershipsSetup struct {
	pool              *pgxpool.Pool
	queries           *db.Queries
	sessions          *auth.SessionStore
	groupHandler      *GroupHandler
	membershipHandler *MembershipHandler
	mux               *http.ServeMux
	cleanup           func()
}

// setupMembershipsTest creates a test environment with a real database container.
func setupMembershipsTest(t *testing.T) *testMembershipsSetup {
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

	return &testMembershipsSetup{
		pool:              pool,
		queries:           queries,
		sessions:          sessions,
		groupHandler:      groupHandler,
		membershipHandler: membershipHandler,
		mux:               mux,
		cleanup: func() {
			pool.Close()
			cleanup()
		},
	}
}

// createTestUser creates a test user directly in the database.
func (s *testMembershipsSetup) createTestUser(t *testing.T, email, name string) *db.User {
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
func (s *testMembershipsSetup) createTestSession(t *testing.T, userID int64) string {
	t.Helper()
	session, err := s.sessions.Create(userID, "", "")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	return session.Token
}

// createTestGroup creates a group via API and returns the group ID.
func (s *testMembershipsSetup) createTestGroup(t *testing.T, token, name string) int64 {
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

// ============================================================
// T031-T037: User Story 2 - Invite Members Tests (TDD)
// ============================================================

// TestInviteMember_TableDriven is the main table-driven test for inviteMember handler.
// T031: Write table-driven tests for inviteMember handler
func TestInviteMember_TableDriven(t *testing.T) {
	setup := setupMembershipsTest(t)
	defer setup.cleanup()

	// Create admin user (group creator)
	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	// Create member user (has membership but not admin)
	memberUser := setup.createTestUser(t, "member@example.com", "Member User")
	memberToken := setup.createTestSession(t, memberUser.ID)

	// Create user to invite
	inviteeUser := setup.createTestUser(t, "invitee@example.com", "Invitee User")

	// Create non-member user
	nonMemberUser := setup.createTestUser(t, "nonmember@example.com", "Non Member")
	nonMemberToken := setup.createTestSession(t, nonMemberUser.ID)

	// Create a group (adminUser becomes admin)
	groupID := setup.createTestGroup(t, adminToken, "Test Group for Invites")

	// Add memberUser as a member (not admin)
	// First invite, then accept
	inviteBody := map[string]any{"user_id": memberUser.ID, "role": "member"}
	inviteBytes, _ := json.Marshal(inviteBody)
	inviteReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), bytes.NewBuffer(inviteBytes))
	inviteReq.Header.Set("Content-Type", "application/json")
	inviteReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})
	inviteW := httptest.NewRecorder()
	setup.mux.ServeHTTP(inviteW, inviteReq)
	if inviteW.Code != http.StatusCreated {
		t.Fatalf("failed to invite member user: %d: %s", inviteW.Code, inviteW.Body.String())
	}
	// Parse membership ID from response
	var inviteResp map[string]any
	if err := json.Unmarshal(inviteW.Body.Bytes(), &inviteResp); err != nil {
		t.Fatalf("failed to parse invite response: %v", err)
	}
	memberMembershipID := int64(inviteResp["membership"].(map[string]any)["id"].(float64))

	// Accept the invitation as memberUser
	acceptReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/accept", memberMembershipID), nil)
	acceptReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: memberToken})
	acceptW := httptest.NewRecorder()
	setup.mux.ServeHTTP(acceptW, acceptReq)
	if acceptW.Code != http.StatusOK {
		t.Fatalf("failed to accept invitation: %d: %s", acceptW.Code, acceptW.Body.String())
	}

	tests := []struct {
		name           string
		cookie         string
		groupID        int64
		body           map[string]any
		wantStatus     int
		wantErrMessage string
	}{
		// T032: Test admin invites user → 201 + pending membership created
		{
			name:       "admin invites user successfully",
			cookie:     adminToken,
			groupID:    groupID,
			body:       map[string]any{"user_id": inviteeUser.ID, "role": "member"},
			wantStatus: http.StatusCreated,
		},
		// T033: Test non-admin without permission → 403 Forbidden
		// Note: Default is members_can_add_members=true, so member CAN invite
		// We need to test with members_can_add_members=false

		// T036a: Test invite non-existent user → 404 Not Found
		{
			name:           "invite non-existent user returns 404",
			cookie:         adminToken,
			groupID:        groupID,
			body:           map[string]any{"user_id": 99999, "role": "member"},
			wantStatus:     http.StatusNotFound,
			wantErrMessage: "User not found",
		},
		// Unauthenticated request
		{
			name:           "unauthenticated request returns 401",
			cookie:         "",
			groupID:        groupID,
			body:           map[string]any{"user_id": inviteeUser.ID, "role": "member"},
			wantStatus:     http.StatusUnauthorized,
			wantErrMessage: "Not authenticated",
		},
		// Non-member trying to invite
		{
			name:           "non-member cannot invite",
			cookie:         nonMemberToken,
			groupID:        groupID,
			body:           map[string]any{"user_id": inviteeUser.ID, "role": "member"},
			wantStatus:     http.StatusForbidden,
			wantErrMessage: "Not authorized to invite members",
		},
		// Invite to non-existent group
		{
			name:           "invite to non-existent group returns 404",
			cookie:         adminToken,
			groupID:        99999,
			body:           map[string]any{"user_id": inviteeUser.ID, "role": "member"},
			wantStatus:     http.StatusNotFound,
			wantErrMessage: "Group not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", tt.groupID), bytes.NewBuffer(bodyBytes))
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

				membership, ok := resp["membership"].(map[string]any)
				if !ok {
					t.Fatalf("response missing membership object: %v", resp)
				}

				// Verify membership is pending (accepted_at is null)
				if membership["accepted_at"] != nil {
					t.Error("new invitation should have null accepted_at")
				}

				// Verify role
				if role, ok := membership["role"].(string); ok {
					expectedRole := tt.body["role"].(string)
					if role != expectedRole {
						t.Errorf("expected role %q, got %q", expectedRole, role)
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

// TestInviteMember_AlreadyMember tests that inviting an existing member returns 409.
// T034: Test invite already-member → 409 Conflict
func TestInviteMember_AlreadyMember(t *testing.T) {
	setup := setupMembershipsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	memberUser := setup.createTestUser(t, "member@example.com", "Member User")

	groupID := setup.createTestGroup(t, adminToken, "Test Group")

	// First invitation (should succeed)
	body := map[string]any{"user_id": memberUser.ID, "role": "member"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("first invitation failed: %d: %s", w.Code, w.Body.String())
	}

	// Second invitation (should fail with 409)
	bodyBytes2, _ := json.Marshal(body)
	req2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), bytes.NewBuffer(bodyBytes2))
	req2.Header.Set("Content-Type", "application/json")
	req2.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w2 := httptest.NewRecorder()
	setup.mux.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict for duplicate membership, got %d: %s", w2.Code, w2.Body.String())
	}

	if !bytes.Contains(w2.Body.Bytes(), []byte("already a member")) {
		t.Errorf("expected error about existing membership, got: %s", w2.Body.String())
	}
}

// TestAcceptInvitation tests accepting a pending invitation.
// T035: Test acceptInvitation → membership.accepted_at set
func TestAcceptInvitation(t *testing.T) {
	setup := setupMembershipsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	inviteeUser := setup.createTestUser(t, "invitee@example.com", "Invitee User")
	inviteeToken := setup.createTestSession(t, inviteeUser.ID)

	groupID := setup.createTestGroup(t, adminToken, "Test Group")

	// Create invitation
	body := map[string]any{"user_id": inviteeUser.ID, "role": "member"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("invitation failed: %d: %s", w.Code, w.Body.String())
	}

	var inviteResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &inviteResp); err != nil {
		t.Fatalf("failed to parse invite response: %v", err)
	}
	membershipID := int64(inviteResp["membership"].(map[string]any)["id"].(float64))

	// Accept invitation
	acceptReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/accept", membershipID), nil)
	acceptReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: inviteeToken})

	acceptW := httptest.NewRecorder()
	setup.mux.ServeHTTP(acceptW, acceptReq)

	if acceptW.Code != http.StatusOK {
		t.Fatalf("accept invitation failed: %d: %s", acceptW.Code, acceptW.Body.String())
	}

	var acceptResp map[string]any
	if err := json.Unmarshal(acceptW.Body.Bytes(), &acceptResp); err != nil {
		t.Fatalf("failed to parse accept response: %v", err)
	}

	membership := acceptResp["membership"].(map[string]any)
	if membership["accepted_at"] == nil {
		t.Error("accepted membership should have accepted_at set")
	}
}

// TestAcceptInvitation_NotFound tests accepting a non-existent invitation.
// T036: Test accept non-existent invitation → 404 Not Found
func TestAcceptInvitation_NotFound(t *testing.T) {
	setup := setupMembershipsTest(t)
	defer setup.cleanup()

	user := setup.createTestUser(t, "user@example.com", "Test User")
	token := setup.createTestSession(t, user.ID)

	// Try to accept non-existent invitation
	req := httptest.NewRequest(http.MethodPost, "/api/v1/memberships/99999/accept", nil)
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: token})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d: %s", w.Code, w.Body.String())
	}
}

// TestAcceptInvitation_WrongUser tests that a user cannot accept another user's invitation.
// T037: Test accept someone else's invitation → 403 Forbidden
func TestAcceptInvitation_WrongUser(t *testing.T) {
	setup := setupMembershipsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	inviteeUser := setup.createTestUser(t, "invitee@example.com", "Invitee User")

	wrongUser := setup.createTestUser(t, "wrong@example.com", "Wrong User")
	wrongToken := setup.createTestSession(t, wrongUser.ID)

	groupID := setup.createTestGroup(t, adminToken, "Test Group")

	// Create invitation for inviteeUser
	body := map[string]any{"user_id": inviteeUser.ID, "role": "member"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("invitation failed: %d: %s", w.Code, w.Body.String())
	}

	var inviteResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &inviteResp); err != nil {
		t.Fatalf("failed to parse invite response: %v", err)
	}
	membershipID := int64(inviteResp["membership"].(map[string]any)["id"].(float64))

	// wrongUser tries to accept inviteeUser's invitation
	acceptReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/accept", membershipID), nil)
	acceptReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: wrongToken})

	acceptW := httptest.NewRecorder()
	setup.mux.ServeHTTP(acceptW, acceptReq)

	if acceptW.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got %d: %s", acceptW.Code, acceptW.Body.String())
	}
}

// TestListMemberships_StatusFilter tests filtering memberships by status.
// T036b: Test listMemberships with status=pending → returns only pending invitations
// T036c: Test listMemberships with status=active → returns only accepted memberships
func TestListMemberships_StatusFilter(t *testing.T) {
	setup := setupMembershipsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	memberUser := setup.createTestUser(t, "member@example.com", "Member User")
	memberToken := setup.createTestSession(t, memberUser.ID)

	pendingUser := setup.createTestUser(t, "pending@example.com", "Pending User")

	groupID := setup.createTestGroup(t, adminToken, "Test Group")

	// Invite memberUser and have them accept
	body := map[string]any{"user_id": memberUser.ID, "role": "member"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})
	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("invite member failed: %d: %s", w.Code, w.Body.String())
	}
	var memberInviteResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &memberInviteResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	memberMembershipID := int64(memberInviteResp["membership"].(map[string]any)["id"].(float64))

	// Accept membership
	acceptReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/accept", memberMembershipID), nil)
	acceptReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: memberToken})
	acceptW := httptest.NewRecorder()
	setup.mux.ServeHTTP(acceptW, acceptReq)
	if acceptW.Code != http.StatusOK {
		t.Fatalf("accept membership failed: %d: %s", acceptW.Code, acceptW.Body.String())
	}

	// Invite pendingUser but don't accept
	body2 := map[string]any{"user_id": pendingUser.ID, "role": "member"}
	bodyBytes2, _ := json.Marshal(body2)
	req2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), bytes.NewBuffer(bodyBytes2))
	req2.Header.Set("Content-Type", "application/json")
	req2.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})
	w2 := httptest.NewRecorder()
	setup.mux.ServeHTTP(w2, req2)
	if w2.Code != http.StatusCreated {
		t.Fatalf("invite pending user failed: %d: %s", w2.Code, w2.Body.String())
	}

	// Test status=active (should have admin + member = 2)
	listActiveReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/groups/%d/memberships?status=active", groupID), nil)
	listActiveReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})
	listActiveW := httptest.NewRecorder()
	setup.mux.ServeHTTP(listActiveW, listActiveReq)

	if listActiveW.Code != http.StatusOK {
		t.Fatalf("list active memberships failed: %d: %s", listActiveW.Code, listActiveW.Body.String())
	}

	var activeResp map[string]any
	if err := json.Unmarshal(listActiveW.Body.Bytes(), &activeResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	activeMemberships := activeResp["memberships"].([]any)
	if len(activeMemberships) != 2 {
		t.Errorf("expected 2 active memberships, got %d", len(activeMemberships))
	}

	// Test status=pending (should have 1)
	listPendingReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/groups/%d/memberships?status=pending", groupID), nil)
	listPendingReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})
	listPendingW := httptest.NewRecorder()
	setup.mux.ServeHTTP(listPendingW, listPendingReq)

	if listPendingW.Code != http.StatusOK {
		t.Fatalf("list pending memberships failed: %d: %s", listPendingW.Code, listPendingW.Body.String())
	}

	var pendingResp map[string]any
	if err := json.Unmarshal(listPendingW.Body.Bytes(), &pendingResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	pendingMemberships := pendingResp["memberships"].([]any)
	if len(pendingMemberships) != 1 {
		t.Errorf("expected 1 pending membership, got %d", len(pendingMemberships))
	}
}

// TestInviteMember_InviterID tests that inviter_id is correctly recorded.
// T036d: Test inviteMember records correct inviter_id
func TestInviteMember_InviterID(t *testing.T) {
	setup := setupMembershipsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	inviteeUser := setup.createTestUser(t, "invitee@example.com", "Invitee User")

	groupID := setup.createTestGroup(t, adminToken, "Test Group")

	// Invite user
	body := map[string]any{"user_id": inviteeUser.ID, "role": "member"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("invitation failed: %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	membership := resp["membership"].(map[string]any)

	// Verify inviter info is present
	inviter, ok := membership["inviter"].(map[string]any)
	if !ok {
		t.Fatal("response should include inviter info")
	}

	inviterID := int64(inviter["id"].(float64))
	if inviterID != adminUser.ID {
		t.Errorf("expected inviter_id %d, got %d", adminUser.ID, inviterID)
	}
}

// ============================================================
// T051-T057: User Story 3 - Manage Group Members Tests (TDD)
// ============================================================

// TestPromoteMember_TableDriven is the main table-driven test for promote/demote handlers.
// T051: Write table-driven tests for promoteMember handler
func TestPromoteMember_TableDriven(t *testing.T) {
	setup := setupMembershipsTest(t)
	defer setup.cleanup()

	// Create admin user (group creator)
	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	// Create member user
	memberUser := setup.createTestUser(t, "member@example.com", "Member User")
	memberToken := setup.createTestSession(t, memberUser.ID)

	// Create a group
	groupID := setup.createTestGroup(t, adminToken, "Test Group")

	// Invite member and accept
	memberMembershipID := setup.inviteAndAccept(t, adminToken, memberToken, groupID, memberUser.ID)

	tests := []struct {
		name           string
		cookie         string
		membershipID   int64
		wantStatus     int
		wantRole       string
		wantErrMessage string
	}{
		// T052: Test admin promotes member → role becomes admin
		{
			name:         "admin promotes member to admin",
			cookie:       adminToken,
			membershipID: memberMembershipID,
			wantStatus:   http.StatusOK,
			wantRole:     "admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/promote", tt.membershipID), nil)
			if tt.cookie != "" {
				req.AddCookie(&http.Cookie{Name: "loomio_session", Value: tt.cookie})
			}

			w := httptest.NewRecorder()
			setup.mux.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d: %s", tt.wantStatus, w.Code, w.Body.String())
				return
			}

			if tt.wantStatus == http.StatusOK && tt.wantRole != "" {
				var resp map[string]any
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				membership := resp["membership"].(map[string]any)
				if role := membership["role"].(string); role != tt.wantRole {
					t.Errorf("expected role %q, got %q", tt.wantRole, role)
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

// TestPromoteMember_NonAdmin tests that non-admins cannot promote.
// T053: Test non-admin tries to promote → 403 Forbidden
func TestPromoteMember_NonAdmin(t *testing.T) {
	setup := setupMembershipsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	memberUser := setup.createTestUser(t, "member@example.com", "Member User")
	memberToken := setup.createTestSession(t, memberUser.ID)

	targetUser := setup.createTestUser(t, "target@example.com", "Target User")

	groupID := setup.createTestGroup(t, adminToken, "Test Group")

	// Invite member and accept
	_ = setup.inviteAndAccept(t, adminToken, memberToken, groupID, memberUser.ID)

	// Invite target and accept
	targetToken := setup.createTestSession(t, targetUser.ID)
	targetMembershipID := setup.inviteAndAccept(t, adminToken, targetToken, groupID, targetUser.ID)

	// Member tries to promote target (should fail)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/promote", targetMembershipID), nil)
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: memberToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got %d: %s", w.Code, w.Body.String())
	}
}

// TestDemoteMember tests demoting an admin to member.
// T054: Test admin demotes other admin → role becomes member
func TestDemoteMember(t *testing.T) {
	setup := setupMembershipsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	otherAdminUser := setup.createTestUser(t, "other-admin@example.com", "Other Admin")
	otherAdminToken := setup.createTestSession(t, otherAdminUser.ID)

	groupID := setup.createTestGroup(t, adminToken, "Test Group")

	// Invite other admin and accept, then promote to admin
	otherAdminMembershipID := setup.inviteAndAccept(t, adminToken, otherAdminToken, groupID, otherAdminUser.ID)

	// Promote to admin first
	promoteReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/promote", otherAdminMembershipID), nil)
	promoteReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})
	promoteW := httptest.NewRecorder()
	setup.mux.ServeHTTP(promoteW, promoteReq)
	if promoteW.Code != http.StatusOK {
		t.Fatalf("failed to promote: %d: %s", promoteW.Code, promoteW.Body.String())
	}

	// Now demote back to member
	demoteReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/demote", otherAdminMembershipID), nil)
	demoteReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})
	demoteW := httptest.NewRecorder()
	setup.mux.ServeHTTP(demoteW, demoteReq)

	if demoteW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", demoteW.Code, demoteW.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(demoteW.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	membership := resp["membership"].(map[string]any)
	if role := membership["role"].(string); role != "member" {
		t.Errorf("expected role 'member', got %q", role)
	}
}

// TestDemoteLastAdmin tests that demoting the last admin is blocked.
// T055: Test demote last admin → 409 Conflict
func TestDemoteLastAdmin(t *testing.T) {
	setup := setupMembershipsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	groupID := setup.createTestGroup(t, adminToken, "Test Group")

	// Get admin's membership ID
	membership, err := setup.queries.GetMembershipByGroupAndUser(context.Background(), db.GetMembershipByGroupAndUserParams{
		GroupID: groupID,
		UserID:  adminUser.ID,
	})
	if err != nil {
		t.Fatalf("failed to get admin membership: %v", err)
	}

	// Try to demote the only admin (should fail)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/demote", membership.ID), nil)
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict, got %d: %s", w.Code, w.Body.String())
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("last admin")) {
		t.Errorf("expected error about last admin, got: %s", w.Body.String())
	}
}

// TestRemoveMember tests removing a member from a group.
// T056: Test remove member → membership deleted
func TestRemoveMember(t *testing.T) {
	setup := setupMembershipsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	memberUser := setup.createTestUser(t, "member@example.com", "Member User")
	memberToken := setup.createTestSession(t, memberUser.ID)

	groupID := setup.createTestGroup(t, adminToken, "Test Group")

	// Invite member and accept
	membershipID := setup.inviteAndAccept(t, adminToken, memberToken, groupID, memberUser.ID)

	// Remove member
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/memberships/%d", membershipID), nil)
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d: %s", w.Code, w.Body.String())
	}

	// Verify membership is deleted
	_, err := setup.queries.GetMembershipByID(context.Background(), membershipID)
	if !db.IsNotFound(err) {
		t.Errorf("membership should be deleted, but query returned: %v", err)
	}
}

// TestRemoveLastAdmin tests that removing the last admin is blocked.
// T057: Test remove last admin → 409 Conflict
func TestRemoveLastAdmin(t *testing.T) {
	setup := setupMembershipsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	groupID := setup.createTestGroup(t, adminToken, "Test Group")

	// Get admin's membership ID
	membership, err := setup.queries.GetMembershipByGroupAndUser(context.Background(), db.GetMembershipByGroupAndUserParams{
		GroupID: groupID,
		UserID:  adminUser.ID,
	})
	if err != nil {
		t.Fatalf("failed to get admin membership: %v", err)
	}

	// Try to remove the only admin (should fail)
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/memberships/%d", membership.ID), nil)
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w := httptest.NewRecorder()
	setup.mux.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict, got %d: %s", w.Code, w.Body.String())
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("last admin")) {
		t.Errorf("expected error about last admin, got: %s", w.Body.String())
	}
}

// inviteAndAccept is a helper that invites a user and accepts on their behalf.
func (s *testMembershipsSetup) inviteAndAccept(t *testing.T, adminToken, memberToken string, groupID, userID int64) int64 {
	t.Helper()

	// Invite user
	body := map[string]any{"user_id": userID, "role": "member"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})

	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("invite failed: %d: %s", w.Code, w.Body.String())
	}

	var inviteResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &inviteResp); err != nil {
		t.Fatalf("failed to parse invite response: %v", err)
	}
	membershipID := int64(inviteResp["membership"].(map[string]any)["id"].(float64))

	// Accept invitation
	acceptReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/accept", membershipID), nil)
	acceptReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: memberToken})
	acceptW := httptest.NewRecorder()
	s.mux.ServeHTTP(acceptW, acceptReq)

	if acceptW.Code != http.StatusOK {
		t.Fatalf("accept failed: %d: %s", acceptW.Code, acceptW.Body.String())
	}

	return membershipID
}

// TestListMyInvitations tests listing pending invitations for current user.
func TestListMyInvitations(t *testing.T) {
	setup := setupMembershipsTest(t)
	defer setup.cleanup()

	adminUser := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, adminUser.ID)

	inviteeUser := setup.createTestUser(t, "invitee@example.com", "Invitee User")
	inviteeToken := setup.createTestSession(t, inviteeUser.ID)

	// Create two groups and invite inviteeUser to both
	group1ID := setup.createTestGroup(t, adminToken, "Group One")
	group2ID := setup.createTestGroup(t, adminToken, "Group Two")

	// Invite to group 1
	body1 := map[string]any{"user_id": inviteeUser.ID, "role": "member"}
	bodyBytes1, _ := json.Marshal(body1)
	req1 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", group1ID), bytes.NewBuffer(bodyBytes1))
	req1.Header.Set("Content-Type", "application/json")
	req1.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})
	w1 := httptest.NewRecorder()
	setup.mux.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("invite to group 1 failed: %d: %s", w1.Code, w1.Body.String())
	}

	// Invite to group 2
	body2 := map[string]any{"user_id": inviteeUser.ID, "role": "admin"}
	bodyBytes2, _ := json.Marshal(body2)
	req2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", group2ID), bytes.NewBuffer(bodyBytes2))
	req2.Header.Set("Content-Type", "application/json")
	req2.AddCookie(&http.Cookie{Name: "loomio_session", Value: adminToken})
	w2 := httptest.NewRecorder()
	setup.mux.ServeHTTP(w2, req2)
	if w2.Code != http.StatusCreated {
		t.Fatalf("invite to group 2 failed: %d: %s", w2.Code, w2.Body.String())
	}

	// List invitations for inviteeUser
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/invitations", nil)
	listReq.AddCookie(&http.Cookie{Name: "loomio_session", Value: inviteeToken})
	listW := httptest.NewRecorder()
	setup.mux.ServeHTTP(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("list invitations failed: %d: %s", listW.Code, listW.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(listW.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	invitations := resp["invitations"].([]any)
	if len(invitations) != 2 {
		t.Errorf("expected 2 invitations, got %d", len(invitations))
	}

	// Verify each invitation has group and inviter info
	for i, inv := range invitations {
		invitation := inv.(map[string]any)
		if _, ok := invitation["group"]; !ok {
			t.Errorf("invitation %d missing group info", i)
		}
		if _, ok := invitation["inviter"]; !ok {
			t.Errorf("invitation %d missing inviter info", i)
		}
	}
}
