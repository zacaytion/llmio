//go:build integration

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
	"github.com/zacaytion/llmio/internal/db/testutil"
)

// auditRecord represents a row from audit.record_version for test assertions.
type auditRecord struct {
	ID          int64
	RecordID    *string
	OldRecordID *string
	Op          string
	TableName   string
	Record      map[string]any
	OldRecord   map[string]any
	ActorID     *int64
	XactID      int64
}

// testAuditSetup holds shared test infrastructure for audit tests.
type testAuditSetup struct {
	pool              *pgxpool.Pool
	queries           *db.Queries
	sessions          *auth.SessionStore
	groupHandler      *GroupHandler
	membershipHandler *MembershipHandler
	mux               *http.ServeMux
	cleanup           func()
}

// setupAuditTest creates a test environment for audit tests.
func setupAuditTest(t *testing.T) *testAuditSetup {
	t.Helper()
	ctx := context.Background()

	connStr, cleanup := testutil.SetupTestDB(ctx, t)

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		cleanup()
		t.Fatalf("failed to create pool: %v", err)
	}

	queries := db.New(pool)
	sessions := auth.NewSessionStore()

	groupHandler := NewGroupHandler(pool, queries, sessions)
	membershipHandler := NewMembershipHandler(pool, queries, sessions)

	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))
	groupHandler.RegisterRoutes(api)
	membershipHandler.RegisterRoutes(api)

	return &testAuditSetup{
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
func (s *testAuditSetup) createTestUser(t *testing.T, email, name string) *db.User {
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

	_, err = s.pool.Exec(ctx, "UPDATE users SET email_verified = true WHERE id = $1", user.ID)
	if err != nil {
		t.Fatalf("failed to verify user: %v", err)
	}

	return user
}

// createTestSession creates a session for the given user.
func (s *testAuditSetup) createTestSession(t *testing.T, userID int64) string {
	t.Helper()
	session, err := s.sessions.Create(userID, "", "")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	return session.Token
}

// makeRequest is a helper to make HTTP requests.
func (s *testAuditSetup) makeRequest(method, path string, body any, token string) *httptest.ResponseRecorder {
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
	s.mux.ServeHTTP(w, req)
	return w
}

// getAuditRecords retrieves audit records for a specific table, ordered by ID.
func (s *testAuditSetup) getAuditRecords(t *testing.T, tableName string) []auditRecord {
	t.Helper()
	ctx := context.Background()

	rows, err := s.pool.Query(ctx, `
		SELECT id, record_id, old_record_id, op::text, table_name, record, old_record, actor_id, xact_id
		FROM audit.record_version
		WHERE table_name = $1
		ORDER BY id ASC
	`, tableName)
	if err != nil {
		t.Fatalf("failed to query audit records: %v", err)
	}
	defer rows.Close()

	var records []auditRecord
	for rows.Next() {
		var r auditRecord
		var recordJSON, oldRecordJSON []byte
		err := rows.Scan(&r.ID, &r.RecordID, &r.OldRecordID, &r.Op, &r.TableName, &recordJSON, &oldRecordJSON, &r.ActorID, &r.XactID)
		if err != nil {
			t.Fatalf("failed to scan audit record: %v", err)
		}
		if recordJSON != nil {
			if err := json.Unmarshal(recordJSON, &r.Record); err != nil {
				t.Fatalf("failed to parse record JSON: %v", err)
			}
		}
		if oldRecordJSON != nil {
			if err := json.Unmarshal(oldRecordJSON, &r.OldRecord); err != nil {
				t.Fatalf("failed to parse old_record JSON: %v", err)
			}
		}
		records = append(records, r)
	}

	return records
}

// getAuditRecordsByXactID retrieves all audit records for a specific transaction.
func (s *testAuditSetup) getAuditRecordsByXactID(t *testing.T, xactID int64) []auditRecord {
	t.Helper()
	ctx := context.Background()

	rows, err := s.pool.Query(ctx, `
		SELECT id, record_id, old_record_id, op::text, table_name, record, old_record, actor_id, xact_id
		FROM audit.record_version
		WHERE xact_id = $1
		ORDER BY id ASC
	`, xactID)
	if err != nil {
		t.Fatalf("failed to query audit records by xact_id: %v", err)
	}
	defer rows.Close()

	var records []auditRecord
	for rows.Next() {
		var r auditRecord
		var recordJSON, oldRecordJSON []byte
		err := rows.Scan(&r.ID, &r.RecordID, &r.OldRecordID, &r.Op, &r.TableName, &recordJSON, &oldRecordJSON, &r.ActorID, &r.XactID)
		if err != nil {
			t.Fatalf("failed to scan audit record: %v", err)
		}
		if recordJSON != nil {
			if err := json.Unmarshal(recordJSON, &r.Record); err != nil {
				t.Fatalf("failed to parse record JSON: %v", err)
			}
		}
		if oldRecordJSON != nil {
			if err := json.Unmarshal(oldRecordJSON, &r.OldRecord); err != nil {
				t.Fatalf("failed to parse old_record JSON: %v", err)
			}
		}
		records = append(records, r)
	}

	return records
}

// ============================================================
// T110: Audit verification tests
// ============================================================

// TestAudit_GroupCreation verifies audit records are created for group creation.
// T110a: Test audit record created for group creation (INSERT)
func Test_Audit_GroupCreation(t *testing.T) {
	setup := setupAuditTest(t)
	defer setup.cleanup()

	user := setup.createTestUser(t, "alice@example.com", "Alice")
	token := setup.createTestSession(t, user.ID)

	// Create a group
	w := setup.makeRequest(http.MethodPost, "/api/v1/groups", map[string]any{
		"name":        "Audit Test Group",
		"description": "Testing audit logging",
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create group: %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	groupID := int64(resp["group"].(map[string]any)["id"].(float64))

	// Check audit records
	records := setup.getAuditRecords(t, "groups")
	if len(records) == 0 {
		t.Fatal("expected at least one audit record for groups table")
	}

	// Find the INSERT record for our group
	var groupInsert *auditRecord
	for i := range records {
		if records[i].Op == "INSERT" && records[i].RecordID != nil && *records[i].RecordID == fmt.Sprintf("%d", groupID) {
			groupInsert = &records[i]
			break
		}
	}

	if groupInsert == nil {
		t.Fatal("no INSERT audit record found for created group")
	}

	// T110a: Verify table_name and record JSONB
	if groupInsert.TableName != "groups" {
		t.Errorf("expected table_name='groups', got %s", groupInsert.TableName)
	}
	if groupInsert.Record == nil {
		t.Fatal("expected record JSONB to be present for INSERT")
	}
	if groupInsert.Record["name"] != "Audit Test Group" {
		t.Errorf("expected record.name='Audit Test Group', got %v", groupInsert.Record["name"])
	}
	if groupInsert.OldRecord != nil {
		t.Error("expected old_record to be nil for INSERT")
	}

	// T110g: Verify actor_id matches authenticated user
	if groupInsert.ActorID == nil {
		t.Error("expected actor_id to be set")
	} else if *groupInsert.ActorID != user.ID {
		t.Errorf("expected actor_id=%d, got %d", user.ID, *groupInsert.ActorID)
	}
}

// TestAudit_MembershipInvite verifies audit records are created for membership invitation.
// T110b: Test audit record created for membership invite (INSERT)
func Test_Audit_MembershipInvite(t *testing.T) {
	setup := setupAuditTest(t)
	defer setup.cleanup()

	admin := setup.createTestUser(t, "admin@example.com", "Admin")
	adminToken := setup.createTestSession(t, admin.ID)
	invitee := setup.createTestUser(t, "invitee@example.com", "Invitee")

	// Create a group first
	w := setup.makeRequest(http.MethodPost, "/api/v1/groups", map[string]any{
		"name": "Membership Test Group",
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create group: %d: %s", w.Code, w.Body.String())
	}
	var groupResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &groupResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	groupID := int64(groupResp["group"].(map[string]any)["id"].(float64))

	// Invite a user
	w = setup.makeRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), map[string]any{
		"user_id": invitee.ID,
		"role":    "member",
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("failed to invite user: %d: %s", w.Code, w.Body.String())
	}
	var inviteResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &inviteResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	membershipID := int64(inviteResp["membership"].(map[string]any)["id"].(float64))

	// Check audit records for memberships
	records := setup.getAuditRecords(t, "memberships")

	// Find the INSERT record for our invitation
	var inviteInsert *auditRecord
	for i := range records {
		if records[i].Op == "INSERT" && records[i].RecordID != nil && *records[i].RecordID == fmt.Sprintf("%d", membershipID) {
			inviteInsert = &records[i]
			break
		}
	}

	if inviteInsert == nil {
		t.Fatal("no INSERT audit record found for membership invitation")
	}

	// Verify record contains expected fields
	if inviteInsert.Record["role"] != "member" {
		t.Errorf("expected record.role='member', got %v", inviteInsert.Record["role"])
	}
	if int64(inviteInsert.Record["user_id"].(float64)) != invitee.ID {
		t.Errorf("expected record.user_id=%d, got %v", invitee.ID, inviteInsert.Record["user_id"])
	}
	if int64(inviteInsert.Record["inviter_id"].(float64)) != admin.ID {
		t.Errorf("expected record.inviter_id=%d, got %v", admin.ID, inviteInsert.Record["inviter_id"])
	}

	// T110g: Verify actor_id
	if inviteInsert.ActorID == nil || *inviteInsert.ActorID != admin.ID {
		t.Errorf("expected actor_id=%d, got %v", admin.ID, inviteInsert.ActorID)
	}
}

// TestAudit_MembershipAccept verifies audit records for accepting an invitation.
// T110c: Test audit record created for membership accept (UPDATE)
func Test_Audit_MembershipAccept(t *testing.T) {
	setup := setupAuditTest(t)
	defer setup.cleanup()

	admin := setup.createTestUser(t, "admin@example.com", "Admin")
	adminToken := setup.createTestSession(t, admin.ID)
	invitee := setup.createTestUser(t, "invitee@example.com", "Invitee")
	inviteeToken := setup.createTestSession(t, invitee.ID)

	// Create group
	w := setup.makeRequest(http.MethodPost, "/api/v1/groups", map[string]any{
		"name": "Accept Test Group",
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create group: %d: %s", w.Code, w.Body.String())
	}
	var groupResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &groupResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	groupID := int64(groupResp["group"].(map[string]any)["id"].(float64))

	// Invite user
	w = setup.makeRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), map[string]any{
		"user_id": invitee.ID,
		"role":    "member",
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("failed to invite: %d: %s", w.Code, w.Body.String())
	}
	var inviteResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &inviteResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	membershipID := int64(inviteResp["membership"].(map[string]any)["id"].(float64))

	// Accept invitation
	w = setup.makeRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/accept", membershipID), nil, inviteeToken)
	if w.Code != http.StatusOK {
		t.Fatalf("failed to accept: %d: %s", w.Code, w.Body.String())
	}

	// Check audit records
	records := setup.getAuditRecords(t, "memberships")

	// Find the UPDATE record
	var acceptUpdate *auditRecord
	for i := range records {
		if records[i].Op == "UPDATE" && records[i].RecordID != nil && *records[i].RecordID == fmt.Sprintf("%d", membershipID) {
			acceptUpdate = &records[i]
			break
		}
	}

	if acceptUpdate == nil {
		t.Fatal("no UPDATE audit record found for membership acceptance")
	}

	// T110c: Verify old_record shows null accepted_at
	if acceptUpdate.OldRecord == nil {
		t.Fatal("expected old_record to be present for UPDATE")
	}
	if acceptUpdate.OldRecord["accepted_at"] != nil {
		t.Errorf("expected old_record.accepted_at to be null, got %v", acceptUpdate.OldRecord["accepted_at"])
	}

	// Verify new record has accepted_at set
	if acceptUpdate.Record == nil {
		t.Fatal("expected record to be present for UPDATE")
	}
	if acceptUpdate.Record["accepted_at"] == nil {
		t.Error("expected record.accepted_at to be set after acceptance")
	}

	// T110g: Verify actor_id is the invitee (who accepted)
	if acceptUpdate.ActorID == nil || *acceptUpdate.ActorID != invitee.ID {
		t.Errorf("expected actor_id=%d (invitee), got %v", invitee.ID, acceptUpdate.ActorID)
	}
}

// TestAudit_MembershipPromote verifies audit records for promoting a member.
// T110d: Test audit record created for membership promote (UPDATE)
func Test_Audit_MembershipPromote(t *testing.T) {
	setup := setupAuditTest(t)
	defer setup.cleanup()

	admin := setup.createTestUser(t, "admin@example.com", "Admin")
	adminToken := setup.createTestSession(t, admin.ID)
	member := setup.createTestUser(t, "member@example.com", "Member")
	memberToken := setup.createTestSession(t, member.ID)

	// Create group
	w := setup.makeRequest(http.MethodPost, "/api/v1/groups", map[string]any{
		"name": "Promote Test Group",
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create group: %d: %s", w.Code, w.Body.String())
	}
	var groupResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &groupResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	groupID := int64(groupResp["group"].(map[string]any)["id"].(float64))

	// Invite and accept
	w = setup.makeRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), map[string]any{
		"user_id": member.ID,
		"role":    "member",
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("failed to invite: %d: %s", w.Code, w.Body.String())
	}
	var inviteResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &inviteResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	membershipID := int64(inviteResp["membership"].(map[string]any)["id"].(float64))

	w = setup.makeRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/accept", membershipID), nil, memberToken)
	if w.Code != http.StatusOK {
		t.Fatalf("failed to accept: %d: %s", w.Code, w.Body.String())
	}

	// Promote to admin
	w = setup.makeRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/promote", membershipID), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("failed to promote: %d: %s", w.Code, w.Body.String())
	}

	// Check audit records
	records := setup.getAuditRecords(t, "memberships")

	// Find the UPDATE record for promotion (role change)
	var promoteUpdate *auditRecord
	for i := range records {
		r := &records[i]
		if r.Op == "UPDATE" && r.RecordID != nil && *r.RecordID == fmt.Sprintf("%d", membershipID) {
			// Check if this is the promote update (role changed from member to admin)
			if r.OldRecord != nil && r.OldRecord["role"] == "member" && r.Record != nil && r.Record["role"] == "admin" {
				promoteUpdate = r
				break
			}
		}
	}

	if promoteUpdate == nil {
		t.Fatal("no UPDATE audit record found for membership promotion")
	}

	// T110d: Verify role change in record/old_record
	if promoteUpdate.OldRecord["role"] != "member" {
		t.Errorf("expected old_record.role='member', got %v", promoteUpdate.OldRecord["role"])
	}
	if promoteUpdate.Record["role"] != "admin" {
		t.Errorf("expected record.role='admin', got %v", promoteUpdate.Record["role"])
	}

	// T110g: Verify actor_id is the admin who promoted
	if promoteUpdate.ActorID == nil || *promoteUpdate.ActorID != admin.ID {
		t.Errorf("expected actor_id=%d, got %v", admin.ID, promoteUpdate.ActorID)
	}
}

// TestAudit_MembershipDemote verifies audit records for demoting an admin.
// T110e: Test audit record created for membership demote (UPDATE)
func Test_Audit_MembershipDemote(t *testing.T) {
	setup := setupAuditTest(t)
	defer setup.cleanup()

	admin1 := setup.createTestUser(t, "admin1@example.com", "Admin1")
	admin1Token := setup.createTestSession(t, admin1.ID)
	admin2 := setup.createTestUser(t, "admin2@example.com", "Admin2")
	admin2Token := setup.createTestSession(t, admin2.ID)

	// Create group (admin1 is admin)
	w := setup.makeRequest(http.MethodPost, "/api/v1/groups", map[string]any{
		"name": "Demote Test Group",
	}, admin1Token)
	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create group: %d: %s", w.Code, w.Body.String())
	}
	var groupResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &groupResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	groupID := int64(groupResp["group"].(map[string]any)["id"].(float64))

	// Invite admin2 as admin
	w = setup.makeRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), map[string]any{
		"user_id": admin2.ID,
		"role":    "admin",
	}, admin1Token)
	if w.Code != http.StatusCreated {
		t.Fatalf("failed to invite: %d: %s", w.Code, w.Body.String())
	}
	var inviteResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &inviteResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	admin2MembershipID := int64(inviteResp["membership"].(map[string]any)["id"].(float64))

	// Accept
	w = setup.makeRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/accept", admin2MembershipID), nil, admin2Token)
	if w.Code != http.StatusOK {
		t.Fatalf("failed to accept: %d: %s", w.Code, w.Body.String())
	}

	// Demote admin2 to member
	w = setup.makeRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/demote", admin2MembershipID), nil, admin1Token)
	if w.Code != http.StatusOK {
		t.Fatalf("failed to demote: %d: %s", w.Code, w.Body.String())
	}

	// Check audit records
	records := setup.getAuditRecords(t, "memberships")

	// Find the UPDATE record for demotion
	var demoteUpdate *auditRecord
	for i := range records {
		r := &records[i]
		if r.Op == "UPDATE" && r.RecordID != nil && *r.RecordID == fmt.Sprintf("%d", admin2MembershipID) {
			if r.OldRecord != nil && r.OldRecord["role"] == "admin" && r.Record != nil && r.Record["role"] == "member" {
				demoteUpdate = r
				break
			}
		}
	}

	if demoteUpdate == nil {
		t.Fatal("no UPDATE audit record found for membership demotion")
	}

	// T110e: Verify role change
	if demoteUpdate.OldRecord["role"] != "admin" {
		t.Errorf("expected old_record.role='admin', got %v", demoteUpdate.OldRecord["role"])
	}
	if demoteUpdate.Record["role"] != "member" {
		t.Errorf("expected record.role='member', got %v", demoteUpdate.Record["role"])
	}
}

// TestAudit_MembershipRemove verifies audit records for removing a member.
// T110f: Test audit record created for membership remove (DELETE)
func Test_Audit_MembershipRemove(t *testing.T) {
	setup := setupAuditTest(t)
	defer setup.cleanup()

	admin := setup.createTestUser(t, "admin@example.com", "Admin")
	adminToken := setup.createTestSession(t, admin.ID)
	member := setup.createTestUser(t, "member@example.com", "Member")
	memberToken := setup.createTestSession(t, member.ID)

	// Create group
	w := setup.makeRequest(http.MethodPost, "/api/v1/groups", map[string]any{
		"name": "Remove Test Group",
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create group: %d: %s", w.Code, w.Body.String())
	}
	var groupResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &groupResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	groupID := int64(groupResp["group"].(map[string]any)["id"].(float64))

	// Invite and accept
	w = setup.makeRequest(http.MethodPost, fmt.Sprintf("/api/v1/groups/%d/memberships", groupID), map[string]any{
		"user_id": member.ID,
		"role":    "member",
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("failed to invite: %d: %s", w.Code, w.Body.String())
	}
	var inviteResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &inviteResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	membershipID := int64(inviteResp["membership"].(map[string]any)["id"].(float64))

	w = setup.makeRequest(http.MethodPost, fmt.Sprintf("/api/v1/memberships/%d/accept", membershipID), nil, memberToken)
	if w.Code != http.StatusOK {
		t.Fatalf("failed to accept: %d: %s", w.Code, w.Body.String())
	}

	// Remove member
	w = setup.makeRequest(http.MethodDelete, fmt.Sprintf("/api/v1/memberships/%d", membershipID), nil, adminToken)
	if w.Code != http.StatusNoContent {
		t.Fatalf("failed to remove: %d: %s", w.Code, w.Body.String())
	}

	// Check audit records
	records := setup.getAuditRecords(t, "memberships")

	// Find the DELETE record
	var removeDelete *auditRecord
	for i := range records {
		r := &records[i]
		if r.Op == "DELETE" && r.OldRecordID != nil && *r.OldRecordID == fmt.Sprintf("%d", membershipID) {
			removeDelete = r
			break
		}
	}

	if removeDelete == nil {
		t.Fatal("no DELETE audit record found for membership removal")
	}

	// T110f: Verify old_record contains the deleted membership
	if removeDelete.OldRecord == nil {
		t.Fatal("expected old_record to be present for DELETE")
	}
	if int64(removeDelete.OldRecord["user_id"].(float64)) != member.ID {
		t.Errorf("expected old_record.user_id=%d, got %v", member.ID, removeDelete.OldRecord["user_id"])
	}
	if removeDelete.Record != nil {
		t.Error("expected record to be nil for DELETE")
	}

	// T110g: Verify actor_id
	if removeDelete.ActorID == nil || *removeDelete.ActorID != admin.ID {
		t.Errorf("expected actor_id=%d, got %v", admin.ID, removeDelete.ActorID)
	}
}

// TestAudit_TransactionCorrelation verifies xact_id correlates operations in the same transaction.
// T110h: Test xact_id correlates createGroup + createMembership in same transaction
func Test_Audit_TransactionCorrelation(t *testing.T) {
	setup := setupAuditTest(t)
	defer setup.cleanup()

	admin := setup.createTestUser(t, "admin@example.com", "Admin")
	adminToken := setup.createTestSession(t, admin.ID)

	// Create group (this creates both group and admin membership in same transaction)
	w := setup.makeRequest(http.MethodPost, "/api/v1/groups", map[string]any{
		"name": "Transaction Test Group",
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create group: %d: %s", w.Code, w.Body.String())
	}
	var groupResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &groupResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	groupID := int64(groupResp["group"].(map[string]any)["id"].(float64))

	// Get group INSERT audit record to find xact_id
	groupRecords := setup.getAuditRecords(t, "groups")
	var groupInsert *auditRecord
	for i := range groupRecords {
		if groupRecords[i].Op == "INSERT" && groupRecords[i].RecordID != nil && *groupRecords[i].RecordID == fmt.Sprintf("%d", groupID) {
			groupInsert = &groupRecords[i]
			break
		}
	}
	if groupInsert == nil {
		t.Fatal("no INSERT audit record found for group")
	}

	// Get all records in the same transaction
	xactRecords := setup.getAuditRecordsByXactID(t, groupInsert.XactID)

	// Should have at least 2 records: group INSERT and membership INSERT
	if len(xactRecords) < 2 {
		t.Fatalf("expected at least 2 records in transaction, got %d", len(xactRecords))
	}

	// Verify we have both groups and memberships tables in the same transaction
	tables := make(map[string]bool)
	for _, r := range xactRecords {
		tables[r.TableName] = true
	}

	if !tables["groups"] {
		t.Error("expected groups table in transaction")
	}
	if !tables["memberships"] {
		t.Error("expected memberships table in transaction (admin membership created with group)")
	}

	t.Logf("Transaction %d contains %d audit records across tables: %v", groupInsert.XactID, len(xactRecords), tables)
}

// TestAudit_RecordJSONBContents verifies record/old_record JSONB contains expected field values.
// T110i: Test record/old_record JSONB contains expected field values
func Test_Audit_RecordJSONBContents(t *testing.T) {
	setup := setupAuditTest(t)
	defer setup.cleanup()

	admin := setup.createTestUser(t, "admin@example.com", "Admin User")
	adminToken := setup.createTestSession(t, admin.ID)

	// Create group with specific values
	w := setup.makeRequest(http.MethodPost, "/api/v1/groups", map[string]any{
		"name":        "JSONB Test Group",
		"description": "Testing JSONB contents",
		"handle":      "jsonb-test-handle",
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create group: %d: %s", w.Code, w.Body.String())
	}
	var groupResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &groupResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	groupID := int64(groupResp["group"].(map[string]any)["id"].(float64))

	// Get the INSERT audit record
	records := setup.getAuditRecords(t, "groups")
	var groupInsert *auditRecord
	for i := range records {
		if records[i].Op == "INSERT" && records[i].RecordID != nil && *records[i].RecordID == fmt.Sprintf("%d", groupID) {
			groupInsert = &records[i]
			break
		}
	}
	if groupInsert == nil {
		t.Fatal("no INSERT audit record found")
	}

	// Verify all expected fields are in the JSONB
	tests := []struct {
		field string
		want  any
	}{
		{"name", "JSONB Test Group"},
		{"description", "Testing JSONB contents"},
		{"handle", "jsonb-test-handle"},
		{"created_by_id", float64(admin.ID)},
		// Permission flags should be present with defaults
		{"members_can_add_members", true},
		{"members_can_add_guests", true},
		{"members_can_start_discussions", true},
		{"members_can_raise_motions", true},
		{"members_can_edit_discussions", false},
		{"members_can_edit_comments", true},
		{"members_can_delete_comments", true},
		{"members_can_announce", false},
		{"members_can_create_subgroups", false},
		{"admins_can_edit_user_content", false},
		{"parent_members_can_see_discussions", false},
	}

	for _, tt := range tests {
		got, ok := groupInsert.Record[tt.field]
		if !ok {
			t.Errorf("record missing field %q", tt.field)
			continue
		}
		if got != tt.want {
			t.Errorf("record[%q] = %v (%T), want %v (%T)", tt.field, got, got, tt.want, tt.want)
		}
	}

	// Verify timestamps are present (don't check exact values)
	for _, field := range []string{"created_at", "updated_at"} {
		if _, ok := groupInsert.Record[field]; !ok {
			t.Errorf("record missing timestamp field %q", field)
		}
	}

	// Verify nullable fields
	if groupInsert.Record["parent_id"] != nil {
		t.Errorf("expected parent_id to be nil, got %v", groupInsert.Record["parent_id"])
	}
	if groupInsert.Record["archived_at"] != nil {
		t.Errorf("expected archived_at to be nil, got %v", groupInsert.Record["archived_at"])
	}
}
