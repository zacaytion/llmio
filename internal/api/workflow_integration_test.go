//go:build integration

package api_test

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

	"github.com/zacaytion/llmio/internal/api"
	"github.com/zacaytion/llmio/internal/auth"
	"github.com/zacaytion/llmio/internal/db"
	"github.com/zacaytion/llmio/internal/testutil"
)

// T109: Integration test - full group creation → invite → accept → promote workflow
// This test verifies the complete user journey through the groups and memberships API.
func Test_Integration_FullGroupWorkflow(t *testing.T) {
	t.Cleanup(func() { testutil.Restore(t) })

	ctx := context.Background()
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

	// Helper to make requests
	makeRequest := func(method, path string, body any, token string) *httptest.ResponseRecorder {
		var reqBody *bytes.Buffer
		if body != nil {
			bodyBytes, _ := json.Marshal(body)
			reqBody = bytes.NewBuffer(bodyBytes)
		} else {
			reqBody = bytes.NewBuffer(nil)
		}
		req := httptest.NewRequest(method, path, reqBody)
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.AddCookie(&http.Cookie{Name: "loomio_session", Value: token})
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		return w
	}

	// Create test users
	createUser := func(email, name string) (*db.User, string) {
		hash, _ := auth.HashPassword("password123")
		user, err := queries.CreateUser(ctx, db.CreateUserParams{
			Email:        email,
			Name:         name,
			Username:     auth.GenerateUsername(name),
			PasswordHash: hash,
			Key:          auth.GeneratePublicKey(),
		})
		if err != nil {
			t.Fatalf("failed to create user %s: %v", email, err)
		}
		_, err = pool.Exec(ctx, "UPDATE users SET email_verified = true WHERE id = $1", user.ID)
		if err != nil {
			t.Fatalf("failed to verify user: %v", err)
		}
		session, _ := sessions.Create(user.ID, "", "")
		return user, session.Token
	}

	// Step 1: Create users
	t.Log("Step 1: Creating test users...")
	alice, aliceToken := createUser("alice@example.com", "Alice Admin")
	bob, bobToken := createUser("bob@example.com", "Bob Member")
	charlie, charlieToken := createUser("charlie@example.com", "Charlie Invited")

	// Step 2: Alice creates a group
	t.Log("Step 2: Alice creates a group...")
	w := makeRequest(http.MethodPost, "/api/v1/groups", map[string]any{
		"name":        "Climate Action Team",
		"description": "Working on climate initiatives",
	}, aliceToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("Step 2 failed: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var createGroupResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &createGroupResp); err != nil {
		t.Fatalf("failed to parse create group response: %v", err)
	}
	group := createGroupResp["group"].(map[string]any)
	groupID := int64(group["id"].(float64))
	t.Logf("  Created group ID=%d, handle=%s", groupID, group["handle"])

	// Verify Alice is an admin by calling getGroup (which returns GroupDetailDTO with current_user_role)
	w = makeRequest(http.MethodGet, fmt.Sprintf("/api/v1/groups/%d", groupID), nil, aliceToken)
	if w.Code != http.StatusOK {
		t.Fatalf("Step 2 verification failed: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var getGroupResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &getGroupResp); err != nil {
		t.Fatalf("failed to parse get group response: %v", err)
	}
	groupDetail := getGroupResp["group"].(map[string]any)
	if groupDetail["current_user_role"] != "admin" {
		t.Errorf("Step 2 verification failed: expected creator to be admin, got %v", groupDetail["current_user_role"])
	}

	// Step 3: Alice invites Bob as a member
	t.Log("Step 3: Alice invites Bob as a member...")
	w = makeRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), map[string]any{
		"user_id": bob.ID,
		"role":    "member",
	}, aliceToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("Step 3 failed: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var inviteBobResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &inviteBobResp); err != nil {
		t.Fatalf("failed to parse invite response: %v", err)
	}
	bobMembership := inviteBobResp["membership"].(map[string]any)
	bobMembershipID := int64(bobMembership["id"].(float64))
	t.Logf("  Created membership ID=%d for Bob (pending)", bobMembershipID)

	// Verify invitation is pending (accepted_at is null)
	if bobMembership["accepted_at"] != nil {
		t.Errorf("Step 3 verification failed: expected pending invitation (null accepted_at)")
	}

	// Step 4: Bob sees his pending invitation
	t.Log("Step 4: Bob checks his pending invitations...")
	w = makeRequest(http.MethodGet, "/api/v1/users/me/invitations", nil, bobToken)
	if w.Code != http.StatusOK {
		t.Fatalf("Step 4 failed: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var invitationsResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &invitationsResp); err != nil {
		t.Fatalf("failed to parse invitations response: %v", err)
	}
	invitations := invitationsResp["invitations"].([]any)
	if len(invitations) != 1 {
		t.Errorf("Step 4 verification failed: expected 1 invitation, got %d", len(invitations))
	}

	// Step 5: Bob accepts the invitation
	t.Log("Step 5: Bob accepts the invitation...")
	w = makeRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/accept", bobMembershipID), nil, bobToken)
	if w.Code != http.StatusOK {
		t.Fatalf("Step 5 failed: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var acceptResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &acceptResp); err != nil {
		t.Fatalf("failed to parse accept response: %v", err)
	}
	acceptedMembership := acceptResp["membership"].(map[string]any)
	if acceptedMembership["accepted_at"] == nil {
		t.Errorf("Step 5 verification failed: expected accepted_at to be set")
	}
	t.Log("  Bob is now an active member")

	// Step 6: Alice invites Charlie
	t.Log("Step 6: Alice invites Charlie...")
	w = makeRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), map[string]any{
		"user_id": charlie.ID,
		"role":    "member",
	}, aliceToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("Step 6 failed: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var inviteCharlieResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &inviteCharlieResp); err != nil {
		t.Fatalf("failed to parse invite response: %v", err)
	}
	charlieMembership := inviteCharlieResp["membership"].(map[string]any)
	charlieMembershipID := int64(charlieMembership["id"].(float64))

	// Step 7: Charlie accepts
	t.Log("Step 7: Charlie accepts the invitation...")
	w = makeRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/accept", charlieMembershipID), nil, charlieToken)
	if w.Code != http.StatusOK {
		t.Fatalf("Step 7 failed: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Step 8: Alice promotes Bob to admin
	t.Log("Step 8: Alice promotes Bob to admin...")
	w = makeRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/promote", bobMembershipID), nil, aliceToken)
	if w.Code != http.StatusOK {
		t.Fatalf("Step 8 failed: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var promoteResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &promoteResp); err != nil {
		t.Fatalf("failed to parse promote response: %v", err)
	}
	promotedMembership := promoteResp["membership"].(map[string]any)
	if promotedMembership["role"] != "admin" {
		t.Errorf("Step 8 verification failed: expected role=admin, got %v", promotedMembership["role"])
	}
	t.Log("  Bob is now an admin")

	// Step 9: Verify both admins can manage the group
	t.Log("Step 9: Verifying Bob (now admin) can invite users...")
	// Bob should now be able to invite
	// (Note: members_can_add_members is true by default, so this tests admin capability)
	newUser, _ := createUser("dave@example.com", "Dave Newcomer")
	w = makeRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), map[string]any{
		"user_id": newUser.ID,
		"role":    "member",
	}, bobToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("Step 9 failed: expected Bob as admin to invite, got %d: %s", w.Code, w.Body.String())
	}
	t.Log("  Bob successfully invited Dave")

	// Step 10: Verify group details show correct counts
	t.Log("Step 10: Verifying group details...")
	w = makeRequest(http.MethodGet, fmt.Sprintf("/api/v1/groups/%d", groupID), nil, aliceToken)
	if w.Code != http.StatusOK {
		t.Fatalf("Step 10 failed: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var groupDetailResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &groupDetailResp); err != nil {
		t.Fatalf("failed to parse group detail response: %v", err)
	}
	groupDetail = groupDetailResp["group"].(map[string]any)
	// Active members: Alice, Bob, Charlie (Dave's invitation is pending)
	memberCount := int(groupDetail["member_count"].(float64))
	adminCount := int(groupDetail["admin_count"].(float64))
	t.Logf("  Group has %d members, %d admins", memberCount, adminCount)
	if memberCount != 3 {
		t.Errorf("Step 10 verification failed: expected 3 active members, got %d", memberCount)
	}
	if adminCount != 2 {
		t.Errorf("Step 10 verification failed: expected 2 admins (Alice, Bob), got %d", adminCount)
	}

	// Step 11: Test last-admin protection
	t.Log("Step 11: Testing last-admin protection...")
	// First, get Alice's membership ID
	w = makeRequest(http.MethodGet, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), nil, aliceToken)
	if w.Code != http.StatusOK {
		t.Fatalf("failed to list memberships: %v", w.Body.String())
	}
	var membershipsResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &membershipsResp); err != nil {
		t.Fatalf("failed to parse memberships response: %v", err)
	}
	memberships := membershipsResp["memberships"].([]any)
	var aliceMembershipID int64
	for _, m := range memberships {
		mem := m.(map[string]any)
		if int64(mem["user_id"].(float64)) == alice.ID {
			aliceMembershipID = int64(mem["id"].(float64))
			break
		}
	}

	// Demote Alice - should succeed since Bob is also admin
	w = makeRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/demote", aliceMembershipID), nil, bobToken)
	if w.Code != http.StatusOK {
		t.Errorf("Step 11a failed: expected to demote Alice since Bob is also admin, got %d: %s", w.Code, w.Body.String())
	} else {
		t.Log("  Successfully demoted Alice (Bob is still admin)")
	}

	// Try to demote Bob (last admin) - should fail
	w = makeRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/demote", bobMembershipID), nil, bobToken)
	if w.Code != http.StatusConflict {
		t.Errorf("Step 11b failed: expected 409 Conflict when demoting last admin, got %d: %s", w.Code, w.Body.String())
	} else {
		t.Log("  Last-admin protection working: cannot demote Bob")
	}

	t.Log("Integration test completed successfully!")
}
