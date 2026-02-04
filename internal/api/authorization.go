package api

import (
	"context"

	"github.com/zacaytion/llmio/internal/db"
)

// Role represents a membership role in a group.
// T205: Create proper Role type with Valid() method.
// Valid roles are "admin" and "member".
type Role string

// Role constants to avoid magic strings scattered across the codebase.
// T127: Create Role type with constants (RoleAdmin, RoleMember)
const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

// Valid returns true if the role is one of the known valid roles.
// T205: Role.Valid() method for validation.
func (r Role) Valid() bool {
	return r == RoleAdmin || r == RoleMember
}

// String returns the string representation of the role.
func (r Role) String() string {
	return string(r)
}

// ParseRole converts a string to a Role if valid, or returns an empty Role if invalid.
// T206: ParseRole function for safe role parsing.
// Returns RoleMember if role is empty or invalid (safe default).
// Returns the parsed Role if valid.
// Use ParseRoleStrict if you need to detect invalid roles.
func ParseRole(s string) Role {
	r := Role(s)
	if r.Valid() {
		return r
	}
	return RoleMember // Safe default
}

// ParseRoleStrict converts a string to a Role.
// T206: Strict version that returns the role as-is for validation.
// Use with r.Valid() to check if parsing succeeded.
func ParseRoleStrict(s string) Role {
	return Role(s)
}

// AuthorizationContext holds authorization-related data for a request.
// Note: The fields are exported for read access in handlers.
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
	// T167: Simplified - both NotFound and other errors return the error
	if err != nil {
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
		// T207: Compare using Role type for type safety
		authCtx.IsAdmin = Role(membership.Role) == RoleAdmin
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
