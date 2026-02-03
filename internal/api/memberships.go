// Package api provides HTTP handlers and DTOs for the authentication API.
package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zacaytion/llmio/internal/auth"
	"github.com/zacaytion/llmio/internal/db"
)

// MembershipHandler handles membership-related HTTP requests.
type MembershipHandler struct {
	pool     *pgxpool.Pool
	queries  *db.Queries
	sessions *auth.SessionStore
}

// NewMembershipHandler creates a new membership handler.
func NewMembershipHandler(pool *pgxpool.Pool, queries *db.Queries, sessions *auth.SessionStore) *MembershipHandler {
	return &MembershipHandler{
		pool:     pool,
		queries:  queries,
		sessions: sessions,
	}
}

// RegisterRoutes registers all membership routes.
func (h *MembershipHandler) RegisterRoutes(api huma.API) {
	// List memberships in a group
	huma.Register(api, huma.Operation{
		OperationID: "listMemberships",
		Method:      http.MethodGet,
		Path:        "/api/v1/groups/{groupId}/memberships",
		Summary:     "List group members",
		Description: "Returns all memberships in a group. Requires membership in the group.",
		Tags:        []string{"Memberships"},
	}, h.handleListMemberships)

	// Invite a user to a group
	huma.Register(api, huma.Operation{
		OperationID:   "inviteMember",
		Method:        http.MethodPost,
		Path:          "/api/v1/groups/{groupId}/memberships",
		Summary:       "Invite user to group",
		Description:   "Invites a user to join a group. Requires admin role or members_can_add_members permission.",
		Tags:          []string{"Memberships"},
		DefaultStatus: http.StatusCreated,
	}, h.handleInviteMember)

	// Get single membership by ID
	huma.Register(api, huma.Operation{
		OperationID: "getMembership",
		Method:      http.MethodGet,
		Path:        "/api/v1/memberships/{id}",
		Summary:     "Get membership details",
		Description: "Returns a single membership by ID. Requires membership in the associated group.",
		Tags:        []string{"Memberships"},
	}, h.handleGetMembership)

	// Accept an invitation
	huma.Register(api, huma.Operation{
		OperationID: "acceptInvitation",
		Method:      http.MethodPost,
		Path:        "/api/v1/memberships/{id}/accept",
		Summary:     "Accept invitation",
		Description: "Accepts a pending invitation. Must be the invited user.",
		Tags:        []string{"Memberships"},
	}, h.handleAcceptInvitation)

	// List current user's pending invitations
	huma.Register(api, huma.Operation{
		OperationID: "listMyInvitations",
		Method:      http.MethodGet,
		Path:        "/api/v1/users/me/invitations",
		Summary:     "List my pending invitations",
		Description: "Returns all pending group invitations for the current user.",
		Tags:        []string{"Memberships"},
	}, h.handleListMyInvitations)

	// Promote member to admin
	huma.Register(api, huma.Operation{
		OperationID: "promoteMember",
		Method:      http.MethodPost,
		Path:        "/api/v1/memberships/{id}/promote",
		Summary:     "Promote member to admin",
		Description: "Promotes a member to admin role. Requires admin permission.",
		Tags:        []string{"Memberships"},
	}, h.handlePromoteMember)

	// Demote admin to member
	huma.Register(api, huma.Operation{
		OperationID: "demoteMember",
		Method:      http.MethodPost,
		Path:        "/api/v1/memberships/{id}/demote",
		Summary:     "Demote admin to member",
		Description: "Demotes an admin to member role. Cannot demote the last admin. Requires admin permission.",
		Tags:        []string{"Memberships"},
	}, h.handleDemoteMember)

	// Remove member from group
	huma.Register(api, huma.Operation{
		OperationID:   "removeMember",
		Method:        http.MethodDelete,
		Path:          "/api/v1/memberships/{id}",
		Summary:       "Remove member from group",
		Description:   "Removes a membership from a group. Cannot remove the last admin. Requires admin permission.",
		Tags:          []string{"Memberships"},
		DefaultStatus: http.StatusNoContent,
	}, h.handleRemoveMember)
}

// ListMembershipsInput is the request for listing memberships.
type ListMembershipsInput struct {
	Cookie  string `cookie:"loomio_session"`
	GroupID int64  `path:"groupId" doc:"Group ID"`
	Status  string `query:"status" enum:"all,active,pending" default:"all" doc:"Filter by membership status"`
}

// ListMembershipsOutput is the response for listing memberships.
type ListMembershipsOutput struct {
	Body struct {
		Memberships []MembershipDTO `json:"memberships"`
	}
}

func (h *MembershipHandler) handleListMemberships(ctx context.Context, input *ListMembershipsInput) (*ListMembershipsOutput, error) {
	// Authenticate
	if input.Cookie == "" {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	session, found := h.sessions.Get(input.Cookie)
	if !found {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// Authorize: user must be a member of the group
	authCtx, err := NewAuthorizationContext(ctx, h.queries, session.UserID, input.GroupID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, huma.Error404NotFound("Group not found")
		}
		LogDBError(ctx, "NewAuthorizationContext", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	if !authCtx.CanViewGroup() {
		return nil, huma.Error403Forbidden("Not a member of this group")
	}

	// List memberships with user info
	rows, err := h.queries.ListMembershipsWithUsers(ctx, db.ListMembershipsWithUsersParams{
		GroupID: input.GroupID,
		Status:  input.Status,
	})
	if err != nil {
		LogDBError(ctx, "ListMembershipsWithUsers", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Convert to DTOs
	memberships := make([]MembershipDTO, len(rows))
	for i, row := range rows {
		memberships[i] = MembershipDTOFromMembershipWithUsers(row)
	}

	output := &ListMembershipsOutput{}
	output.Body.Memberships = memberships
	return output, nil
}

// InviteMemberInput is the request for inviting a user to a group.
type InviteMemberInput struct {
	Cookie  string `cookie:"loomio_session"`
	GroupID int64  `path:"groupId" doc:"Group ID"`
	Body    struct {
		UserID int64  `json:"user_id" required:"true" doc:"ID of the user to invite"`
		Role   string `json:"role" enum:"admin,member" default:"member" doc:"Role to assign when invitation is accepted"`
	}
}

// InviteMemberOutput is the response for inviting a user.
type InviteMemberOutput struct {
	Body struct {
		Membership MembershipDTO `json:"membership"`
	}
}

func (h *MembershipHandler) handleInviteMember(ctx context.Context, input *InviteMemberInput) (*InviteMemberOutput, error) {
	// Authenticate
	if input.Cookie == "" {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	session, found := h.sessions.Get(input.Cookie)
	if !found {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// Authorize: user must be able to invite members
	authCtx, err := NewAuthorizationContext(ctx, h.queries, session.UserID, input.GroupID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, huma.Error404NotFound("Group not found")
		}
		LogDBError(ctx, "NewAuthorizationContext", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	if !authCtx.CanInviteMembers() {
		return nil, huma.Error403Forbidden("Not authorized to invite members")
	}

	// Verify the user to invite exists
	invitee, err := h.queries.GetUserByID(ctx, input.Body.UserID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, huma.Error404NotFound("User not found")
		}
		LogDBError(ctx, "GetUserByID", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Default role to "member" if not specified
	role := input.Body.Role
	if role == "" {
		role = "member"
	}

	// Execute in transaction with audit context
	var membership *db.Membership
	err = pgx.BeginTxFunc(ctx, h.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		// Set audit context for triggers
		if auditErr := db.SetAuditContext(ctx, tx, session.UserID); auditErr != nil {
			return fmt.Errorf("SetAuditContext: %w", auditErr)
		}

		txQueries := h.queries.WithTx(tx)

		// Create pending membership (no accepted_at)
		var createErr error
		membership, createErr = txQueries.CreateMembership(ctx, db.CreateMembershipParams{
			GroupID:   input.GroupID,
			UserID:    input.Body.UserID,
			Role:      role,
			InviterID: session.UserID,
			// AcceptedAt is nil for pending invitations
		})
		if createErr != nil {
			// Check for unique constraint violation
			if isUniqueViolation(createErr, "memberships_unique_user_group") {
				return huma.Error409Conflict("User is already a member of this group")
			}
			return fmt.Errorf("CreateMembership: %w", createErr)
		}
		return nil
	})

	if err != nil {
		// Check if it's already a Huma error
		if humaErr, ok := err.(huma.StatusError); ok {
			return nil, humaErr
		}
		LogDBError(ctx, "InviteMember", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Build response with inviter info
	output := &InviteMemberOutput{}
	output.Body.Membership = MembershipDTO{
		ID:        membership.ID,
		GroupID:   membership.GroupID,
		UserID:    membership.UserID,
		Role:      membership.Role,
		CreatedAt: membership.CreatedAt.Time,
		User: &UserSummaryDTO{
			ID:       invitee.ID,
			Name:     invitee.Name,
			Username: invitee.Username,
		},
		Inviter: &UserSummaryDTO{
			ID:       session.UserID,
			Name:     "", // We don't have the inviter's name handy, but ID is sufficient
			Username: "",
		},
	}

	// Get inviter info for complete response
	inviter, err := h.queries.GetUserByID(ctx, session.UserID)
	if err == nil {
		output.Body.Membership.Inviter.Name = inviter.Name
		output.Body.Membership.Inviter.Username = inviter.Username
	}

	return output, nil
}

// GetMembershipInput is the request for getting a single membership.
type GetMembershipInput struct {
	Cookie       string `cookie:"loomio_session"`
	MembershipID int64  `path:"id" doc:"Membership ID"`
}

// GetMembershipOutput is the response for getting a single membership.
type GetMembershipOutput struct {
	Body struct {
		Membership MembershipDTO `json:"membership"`
		Group      GroupDTO      `json:"group"`
	}
}

func (h *MembershipHandler) handleGetMembership(ctx context.Context, input *GetMembershipInput) (*GetMembershipOutput, error) {
	// Authenticate
	if input.Cookie == "" {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	session, found := h.sessions.Get(input.Cookie)
	if !found {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// Get membership with user info
	membershipRow, err := h.queries.GetMembershipWithUser(ctx, input.MembershipID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, huma.Error404NotFound("Membership not found")
		}
		LogDBError(ctx, "GetMembershipWithUser", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Authorize: user must be a member of the group
	authCtx, err := NewAuthorizationContext(ctx, h.queries, session.UserID, membershipRow.GroupID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, huma.Error404NotFound("Group not found")
		}
		LogDBError(ctx, "NewAuthorizationContext", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	if !authCtx.CanViewGroup() {
		return nil, huma.Error403Forbidden("Not a member of this group")
	}

	// Get group for response
	group, err := h.queries.GetGroupByID(ctx, membershipRow.GroupID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, huma.Error404NotFound("Group not found")
		}
		LogDBError(ctx, "GetGroupByID", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Build response
	output := &GetMembershipOutput{}
	output.Body.Membership = MembershipDTO{
		ID:        membershipRow.ID,
		GroupID:   membershipRow.GroupID,
		UserID:    membershipRow.UserID,
		Role:      membershipRow.Role,
		CreatedAt: membershipRow.CreatedAt.Time,
		User: &UserSummaryDTO{
			ID:       membershipRow.UserID,
			Name:     membershipRow.UserName,
			Username: membershipRow.UserUsername,
		},
		Inviter: &UserSummaryDTO{
			ID:       membershipRow.InviterID,
			Name:     membershipRow.InviterName,
			Username: membershipRow.InviterUsername,
		},
	}
	if membershipRow.AcceptedAt.Valid {
		output.Body.Membership.AcceptedAt = &membershipRow.AcceptedAt.Time
	}
	output.Body.Group = GroupDTOFromGroup(group)

	return output, nil
}

// AcceptInvitationInput is the request for accepting an invitation.
type AcceptInvitationInput struct {
	Cookie       string `cookie:"loomio_session"`
	MembershipID int64  `path:"id" doc:"Membership ID"`
}

// AcceptInvitationOutput is the response for accepting an invitation.
type AcceptInvitationOutput struct {
	Body struct {
		Membership MembershipDTO `json:"membership"`
	}
}

func (h *MembershipHandler) handleAcceptInvitation(ctx context.Context, input *AcceptInvitationInput) (*AcceptInvitationOutput, error) {
	// Authenticate
	if input.Cookie == "" {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	session, found := h.sessions.Get(input.Cookie)
	if !found {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// Get the membership
	membership, err := h.queries.GetMembershipByID(ctx, input.MembershipID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, huma.Error404NotFound("Invitation not found")
		}
		LogDBError(ctx, "GetMembershipByID", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Verify the current user is the invited user
	if membership.UserID != session.UserID {
		return nil, huma.Error403Forbidden("You can only accept your own invitations")
	}

	// Check if already accepted
	if membership.AcceptedAt.Valid {
		return nil, huma.Error409Conflict("Invitation has already been accepted")
	}

	// Execute in transaction with audit context
	var updatedMembership *db.Membership
	err = pgx.BeginTxFunc(ctx, h.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		// Set audit context for triggers
		if auditErr := db.SetAuditContext(ctx, tx, session.UserID); auditErr != nil {
			return fmt.Errorf("SetAuditContext: %w", auditErr)
		}

		txQueries := h.queries.WithTx(tx)

		var acceptErr error
		updatedMembership, acceptErr = txQueries.AcceptMembership(ctx, input.MembershipID)
		if acceptErr != nil {
			return fmt.Errorf("AcceptMembership: %w", acceptErr)
		}
		return nil
	})

	if err != nil {
		LogDBError(ctx, "AcceptInvitation", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Build response
	output := &AcceptInvitationOutput{}
	output.Body.Membership = MembershipDTOFromMembership(updatedMembership)
	return output, nil
}

// ListMyInvitationsInput is the request for listing current user's invitations.
type ListMyInvitationsInput struct {
	Cookie string `cookie:"loomio_session"`
}

// ListMyInvitationsOutput is the response for listing invitations.
type ListMyInvitationsOutput struct {
	Body struct {
		Invitations []InvitationDTO `json:"invitations"`
	}
}

func (h *MembershipHandler) handleListMyInvitations(ctx context.Context, input *ListMyInvitationsInput) (*ListMyInvitationsOutput, error) {
	// Authenticate
	if input.Cookie == "" {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	session, found := h.sessions.Get(input.Cookie)
	if !found {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// List invitations with group and inviter info
	rows, err := h.queries.ListInvitationsWithGroups(ctx, session.UserID)
	if err != nil {
		LogDBError(ctx, "ListInvitationsWithGroups", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Convert to DTOs
	invitations := make([]InvitationDTO, len(rows))
	for i, row := range rows {
		invitations[i] = InvitationDTOFromRow(row)
	}

	output := &ListMyInvitationsOutput{}
	output.Body.Invitations = invitations
	return output, nil
}

// PromoteMemberInput is the request for promoting a member to admin.
type PromoteMemberInput struct {
	Cookie       string `cookie:"loomio_session"`
	MembershipID int64  `path:"id" doc:"Membership ID"`
}

// PromoteMemberOutput is the response for promoting a member.
type PromoteMemberOutput struct {
	Body struct {
		Membership MembershipDTO `json:"membership"`
	}
}

func (h *MembershipHandler) handlePromoteMember(ctx context.Context, input *PromoteMemberInput) (*PromoteMemberOutput, error) {
	// Authenticate
	if input.Cookie == "" {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	session, found := h.sessions.Get(input.Cookie)
	if !found {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// Get the membership to promote
	membership, err := h.queries.GetMembershipByID(ctx, input.MembershipID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, huma.Error404NotFound("Membership not found")
		}
		LogDBError(ctx, "GetMembershipByID", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Authorize: current user must be admin of the group
	authCtx, err := NewAuthorizationContext(ctx, h.queries, session.UserID, membership.GroupID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, huma.Error404NotFound("Group not found")
		}
		LogDBError(ctx, "NewAuthorizationContext", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	if !authCtx.CanManageMembers() {
		return nil, huma.Error403Forbidden("Only admins can promote members")
	}

	// Check if already admin
	if membership.Role == "admin" {
		return nil, huma.Error409Conflict("Member is already an admin")
	}

	// Execute promotion in transaction with audit context
	var updatedMembership *db.Membership
	err = pgx.BeginTxFunc(ctx, h.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if auditErr := db.SetAuditContext(ctx, tx, session.UserID); auditErr != nil {
			return fmt.Errorf("SetAuditContext: %w", auditErr)
		}

		txQueries := h.queries.WithTx(tx)
		var updateErr error
		updatedMembership, updateErr = txQueries.UpdateMembershipRole(ctx, db.UpdateMembershipRoleParams{
			ID:   input.MembershipID,
			Role: "admin",
		})
		if updateErr != nil {
			return fmt.Errorf("UpdateMembershipRole: %w", updateErr)
		}
		return nil
	})

	if err != nil {
		LogDBError(ctx, "PromoteMember", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	output := &PromoteMemberOutput{}
	output.Body.Membership = MembershipDTOFromMembership(updatedMembership)
	return output, nil
}

// DemoteMemberInput is the request for demoting an admin to member.
type DemoteMemberInput struct {
	Cookie       string `cookie:"loomio_session"`
	MembershipID int64  `path:"id" doc:"Membership ID"`
}

// DemoteMemberOutput is the response for demoting a member.
type DemoteMemberOutput struct {
	Body struct {
		Membership MembershipDTO `json:"membership"`
	}
}

func (h *MembershipHandler) handleDemoteMember(ctx context.Context, input *DemoteMemberInput) (*DemoteMemberOutput, error) {
	// Authenticate
	if input.Cookie == "" {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	session, found := h.sessions.Get(input.Cookie)
	if !found {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// Get the membership to demote
	membership, err := h.queries.GetMembershipByID(ctx, input.MembershipID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, huma.Error404NotFound("Membership not found")
		}
		LogDBError(ctx, "GetMembershipByID", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Authorize: current user must be admin of the group
	authCtx, err := NewAuthorizationContext(ctx, h.queries, session.UserID, membership.GroupID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, huma.Error404NotFound("Group not found")
		}
		LogDBError(ctx, "NewAuthorizationContext", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	if !authCtx.CanManageMembers() {
		return nil, huma.Error403Forbidden("Only admins can demote members")
	}

	// Check if already member
	if membership.Role == "member" {
		return nil, huma.Error409Conflict("Member is already a regular member")
	}

	// Check if this is the last admin
	adminCount, err := h.queries.CountAdminsByGroup(ctx, membership.GroupID)
	if err != nil {
		LogDBError(ctx, "CountAdminsByGroup", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	if adminCount <= 1 {
		return nil, huma.Error409Conflict("Cannot demote the last admin of a group")
	}

	// Execute demotion in transaction with audit context
	var updatedMembership *db.Membership
	err = pgx.BeginTxFunc(ctx, h.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if auditErr := db.SetAuditContext(ctx, tx, session.UserID); auditErr != nil {
			return fmt.Errorf("SetAuditContext: %w", auditErr)
		}

		txQueries := h.queries.WithTx(tx)
		var updateErr error
		updatedMembership, updateErr = txQueries.UpdateMembershipRole(ctx, db.UpdateMembershipRoleParams{
			ID:   input.MembershipID,
			Role: "member",
		})
		if updateErr != nil {
			return fmt.Errorf("UpdateMembershipRole: %w", updateErr)
		}
		return nil
	})

	if err != nil {
		LogDBError(ctx, "DemoteMember", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	output := &DemoteMemberOutput{}
	output.Body.Membership = MembershipDTOFromMembership(updatedMembership)
	return output, nil
}

// RemoveMemberInput is the request for removing a member from a group.
type RemoveMemberInput struct {
	Cookie       string `cookie:"loomio_session"`
	MembershipID int64  `path:"id" doc:"Membership ID"`
}

// RemoveMemberOutput is an empty response for member removal.
type RemoveMemberOutput struct{}

func (h *MembershipHandler) handleRemoveMember(ctx context.Context, input *RemoveMemberInput) (*RemoveMemberOutput, error) {
	// Authenticate
	if input.Cookie == "" {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	session, found := h.sessions.Get(input.Cookie)
	if !found {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// Get the membership to remove
	membership, err := h.queries.GetMembershipByID(ctx, input.MembershipID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, huma.Error404NotFound("Membership not found")
		}
		LogDBError(ctx, "GetMembershipByID", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Authorize: current user must be admin of the group
	authCtx, err := NewAuthorizationContext(ctx, h.queries, session.UserID, membership.GroupID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, huma.Error404NotFound("Group not found")
		}
		LogDBError(ctx, "NewAuthorizationContext", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	if !authCtx.CanManageMembers() {
		return nil, huma.Error403Forbidden("Only admins can remove members")
	}

	// Check if this is the last admin (can't remove last admin)
	if membership.Role == "admin" {
		adminCount, countErr := h.queries.CountAdminsByGroup(ctx, membership.GroupID)
		if countErr != nil {
			LogDBError(ctx, "CountAdminsByGroup", countErr)
			return nil, huma.Error500InternalServerError("Database error")
		}

		if adminCount <= 1 {
			return nil, huma.Error409Conflict("Cannot remove the last admin of a group")
		}
	}

	// Execute removal in transaction with audit context
	err = pgx.BeginTxFunc(ctx, h.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if auditErr := db.SetAuditContext(ctx, tx, session.UserID); auditErr != nil {
			return fmt.Errorf("SetAuditContext: %w", auditErr)
		}

		txQueries := h.queries.WithTx(tx)
		if deleteErr := txQueries.DeleteMembership(ctx, input.MembershipID); deleteErr != nil {
			return fmt.Errorf("DeleteMembership: %w", deleteErr)
		}
		return nil
	})

	if err != nil {
		LogDBError(ctx, "RemoveMember", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	return &RemoveMemberOutput{}, nil
}

