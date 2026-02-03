package api

import (
	"context"

	"github.com/zacaytion/llmio/internal/db"
)

// AuthorizationContext holds authorization-related data for a request.
type AuthorizationContext struct {
	UserID     int64
	Membership *db.Membership
	Group      *db.Group
	IsAdmin    bool
	IsMember   bool
}

// NewAuthorizationContext creates an AuthorizationContext by loading the user's
// membership in the specified group. Returns nil membership if not a member.
func NewAuthorizationContext(ctx context.Context, queries *db.Queries, userID, groupID int64) (*AuthorizationContext, error) {
	// Load group first
	group, err := queries.GetGroupByID(ctx, groupID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, err
		}
		return nil, err
	}

	// Load membership (may not exist)
	membership, err := queries.GetMembershipByGroupAndUser(ctx, db.GetMembershipByGroupAndUserParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil && !db.IsNotFound(err) {
		return nil, err
	}

	authCtx := &AuthorizationContext{
		UserID: userID,
		Group:  group,
	}

	if membership != nil && membership.AcceptedAt.Valid {
		authCtx.Membership = membership
		authCtx.IsMember = true
		authCtx.IsAdmin = membership.Role == "admin"
	}

	return authCtx, nil
}

// CanViewGroup checks if the user can view the group.
// Currently requires membership.
func (ac *AuthorizationContext) CanViewGroup() bool {
	return ac.IsMember
}

// CanUpdateGroup checks if the user can update group settings.
// Requires admin role.
func (ac *AuthorizationContext) CanUpdateGroup() bool {
	return ac.IsAdmin
}

// CanArchiveGroup checks if the user can archive/unarchive the group.
// Requires admin role.
func (ac *AuthorizationContext) CanArchiveGroup() bool {
	return ac.IsAdmin
}

// CanInviteMembers checks if the user can invite new members.
// Requires admin role OR (member role AND members_can_add_members flag).
// Per FR-022, admins bypass permission flags.
func (ac *AuthorizationContext) CanInviteMembers() bool {
	if ac.IsAdmin {
		return true
	}
	if ac.IsMember && ac.Group.MembersCanAddMembers {
		return true
	}
	return false
}

// CanManageMembers checks if the user can promote/demote/remove members.
// Requires admin role.
func (ac *AuthorizationContext) CanManageMembers() bool {
	return ac.IsAdmin
}

// CanCreateSubgroups checks if the user can create subgroups.
// Requires admin role OR (member role AND members_can_create_subgroups flag).
// Per FR-022, admins bypass permission flags.
func (ac *AuthorizationContext) CanCreateSubgroups() bool {
	if ac.IsAdmin {
		return true
	}
	if ac.IsMember && ac.Group.MembersCanCreateSubgroups {
		return true
	}
	return false
}

// GetRole returns the user's role string ("admin", "member", or empty).
func (ac *AuthorizationContext) GetRole() string {
	if ac.Membership == nil {
		return ""
	}
	return ac.Membership.Role
}
