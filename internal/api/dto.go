// Package api provides HTTP handlers and DTOs for the groups, memberships, and authentication APIs.
package api

import (
	"time"

	"github.com/zacaytion/llmio/internal/db"
)

// UserDTO represents a user in API responses.
// Excludes sensitive fields like password_hash and deactivated_at.
type UserDTO struct {
	ID            int64     `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	Username      string    `json:"username"`
	EmailVerified bool      `json:"email_verified"`
	Key           string    `json:"key"`
	CreatedAt     time.Time `json:"created_at"`
}

// UserDTOFromUser converts a db.User to a UserDTO for API responses.
func UserDTOFromUser(u *db.User) UserDTO {
	return UserDTO{
		ID:            u.ID,
		Email:         u.Email,
		Name:          u.Name,
		Username:      u.Username,
		EmailVerified: u.EmailVerified,
		Key:           u.Key,
		CreatedAt:     u.CreatedAt.Time,
	}
}

// UserResponse wraps a UserDTO for consistent API responses.
type UserResponse struct {
	Body struct {
		User UserDTO `json:"user"`
	}
}

// ValidationError represents a field validation error.
type ValidationError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// AuthError represents an authentication error.
type AuthError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// SuccessResponse represents a simple success response.
type SuccessResponse struct {
	Body struct {
		Success bool `json:"success"`
	}
}

// ============================================
// Group DTOs (Feature 004)
// ============================================

// GroupDTO represents a group in API responses (basic view).
// T171: Handle field has format constraints:
//   - Length: 3-100 characters
//   - Pattern: ^[a-z0-9][a-z0-9-]*[a-z0-9]$ (lowercase alphanumeric with hyphens, must start/end with alphanumeric)
//   - Case: Always lowercase (normalized via CITEXT in database)
//   - Uniqueness: Globally unique across all groups
type GroupDTO struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Handle      string     `json:"handle"`
	Description *string    `json:"description,omitempty"`
	ParentID    *int64     `json:"parent_id,omitempty"`
	ArchivedAt  *time.Time `json:"archived_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// GroupDTOFromGroup converts a db.Group to GroupDTO.
func GroupDTOFromGroup(g *db.Group) GroupDTO {
	dto := GroupDTO{
		ID:        g.ID,
		Name:      g.Name,
		Handle:    g.Handle,
		CreatedAt: g.CreatedAt.Time,
	}
	if g.Description.Valid {
		dto.Description = &g.Description.String
	}
	if g.ParentID.Valid {
		dto.ParentID = &g.ParentID.Int64
	}
	if g.ArchivedAt.Valid {
		dto.ArchivedAt = &g.ArchivedAt.Time
	}
	return dto
}

// GroupDetailDTO extends GroupDTO with permission flags and member counts.
// Used for getGroup responses where full detail is needed.
// T173: CurrentUserRole indicates the requesting user's role in this group:
//   - "admin": User is an administrator of the group
//   - "member": User is a regular member of the group
//   - "": (empty string) User is not a member (should not normally occur in getGroup responses, as non-members get 403)
type GroupDetailDTO struct {
	GroupDTO

	// Permission flags
	MembersCanAddMembers           bool `json:"members_can_add_members"`
	MembersCanAddGuests            bool `json:"members_can_add_guests"`
	MembersCanStartDiscussions     bool `json:"members_can_start_discussions"`
	MembersCanRaiseMotions         bool `json:"members_can_raise_motions"`
	MembersCanEditDiscussions      bool `json:"members_can_edit_discussions"`
	MembersCanEditComments         bool `json:"members_can_edit_comments"`
	MembersCanDeleteComments       bool `json:"members_can_delete_comments"`
	MembersCanAnnounce             bool `json:"members_can_announce"`
	MembersCanCreateSubgroups      bool `json:"members_can_create_subgroups"`
	AdminsCanEditUserContent       bool `json:"admins_can_edit_user_content"`
	ParentMembersCanSeeDiscussions bool `json:"parent_members_can_see_discussions"`
	// Counts
	MemberCount     int64  `json:"member_count"`
	AdminCount      int64  `json:"admin_count"`
	CurrentUserRole string `json:"current_user_role"` // T173: "admin", "member", or "" (see type docs)
	// Parent status (for subgroups)
	ParentArchived *bool `json:"parent_archived,omitempty"`
}

// GroupDetailDTOFromGroup converts a db.Group to GroupDetailDTO with counts.
func GroupDetailDTOFromGroup(g *db.Group, memberCount, adminCount int64, currentUserRole string) GroupDetailDTO {
	return GroupDetailDTO{
		GroupDTO:                       GroupDTOFromGroup(g),
		MembersCanAddMembers:           g.MembersCanAddMembers,
		MembersCanAddGuests:            g.MembersCanAddGuests,
		MembersCanStartDiscussions:     g.MembersCanStartDiscussions,
		MembersCanRaiseMotions:         g.MembersCanRaiseMotions,
		MembersCanEditDiscussions:      g.MembersCanEditDiscussions,
		MembersCanEditComments:         g.MembersCanEditComments,
		MembersCanDeleteComments:       g.MembersCanDeleteComments,
		MembersCanAnnounce:             g.MembersCanAnnounce,
		MembersCanCreateSubgroups:      g.MembersCanCreateSubgroups,
		AdminsCanEditUserContent:       g.AdminsCanEditUserContent,
		ParentMembersCanSeeDiscussions: g.ParentMembersCanSeeDiscussions,
		MemberCount:                    memberCount,
		AdminCount:                     adminCount,
		CurrentUserRole:                currentUserRole,
	}
}

// ============================================
// Membership DTOs (Feature 004)
// ============================================

// UserSummaryDTO represents minimal user info for embedding in other DTOs.
type UserSummaryDTO struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

// UserSummaryDTOFromUser converts a db.User to UserSummaryDTO.
func UserSummaryDTOFromUser(u *db.User) UserSummaryDTO {
	return UserSummaryDTO{
		ID:       u.ID,
		Name:     u.Name,
		Username: u.Username,
	}
}

// MembershipDTO represents a membership in API responses.
// T172: AcceptedAt semantics:
//   - nil (omitted in JSON): Pending invitation that has not been accepted yet
//   - non-nil: Active membership with acceptance timestamp
//   - Only accepted members (AcceptedAt != nil) are counted as active members
//   - Pending members have limited permissions (cannot view group, cannot invite others)
type MembershipDTO struct {
	ID         int64           `json:"id"`
	GroupID    int64           `json:"group_id"`
	UserID     int64           `json:"user_id"`
	Role       string          `json:"role"`
	AcceptedAt *time.Time      `json:"accepted_at,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	User       *UserSummaryDTO `json:"user,omitempty"`
	Inviter    *UserSummaryDTO `json:"inviter,omitempty"`
}

// MembershipDTOFromMembership converts a db.Membership to MembershipDTO.
func MembershipDTOFromMembership(m *db.Membership) MembershipDTO {
	dto := MembershipDTO{
		ID:        m.ID,
		GroupID:   m.GroupID,
		UserID:    m.UserID,
		Role:      m.Role,
		CreatedAt: m.CreatedAt.Time,
	}
	if m.AcceptedAt.Valid {
		dto.AcceptedAt = &m.AcceptedAt.Time
	}
	return dto
}

// MembershipDTOFromMembershipWithUsers adds user/inviter info to MembershipDTO.
func MembershipDTOFromMembershipWithUsers(m *db.ListMembershipsWithUsersRow) MembershipDTO {
	dto := MembershipDTO{
		ID:        m.ID,
		GroupID:   m.GroupID,
		UserID:    m.UserID,
		Role:      m.Role,
		CreatedAt: m.CreatedAt.Time,
		User: &UserSummaryDTO{
			ID:       m.UserID,
			Name:     m.UserName,
			Username: m.UserUsername,
		},
		Inviter: &UserSummaryDTO{
			ID:       m.InviterID,
			Name:     m.InviterName,
			Username: m.InviterUsername,
		},
	}
	if m.AcceptedAt.Valid {
		dto.AcceptedAt = &m.AcceptedAt.Time
	}
	return dto
}

// InvitationDTO represents a pending invitation in API responses.
// Includes group and inviter context for user-facing display.
type InvitationDTO struct {
	ID        int64          `json:"id"`
	Group     GroupDTO       `json:"group"`
	Inviter   UserSummaryDTO `json:"inviter"`
	Role      string         `json:"role"`
	CreatedAt time.Time      `json:"created_at"`
}

// InvitationDTOFromRow converts a ListInvitationsWithGroupsRow to InvitationDTO.
func InvitationDTOFromRow(row *db.ListInvitationsWithGroupsRow) InvitationDTO {
	dto := InvitationDTO{
		ID:   row.ID,
		Role: row.Role,
		Group: GroupDTO{
			ID:        row.GroupID,
			Name:      row.GroupName,
			Handle:    row.GroupHandle,
			CreatedAt: row.CreatedAt.Time, // Membership created_at
		},
		Inviter: UserSummaryDTO{
			ID:       row.InviterID,
			Name:     row.InviterName,
			Username: row.InviterUsername,
		},
		CreatedAt: row.CreatedAt.Time,
	}
	if row.GroupDescription.Valid {
		dto.Group.Description = &row.GroupDescription.String
	}
	return dto
}

// ============================================
// API Response Wrappers (Feature 004)
// ============================================

// GroupResponse wraps a GroupDTO for API responses.
type GroupResponse struct {
	Body struct {
		Group GroupDTO `json:"group"`
	}
}

// GroupDetailResponse wraps a GroupDetailDTO for API responses.
type GroupDetailResponse struct {
	Body struct {
		Group GroupDetailDTO `json:"group"`
	}
}

// GroupListResponse wraps a list of GroupDTOs for API responses.
type GroupListResponse struct {
	Body struct {
		Groups []GroupDTO `json:"groups"`
	}
}

// MembershipResponse wraps a MembershipDTO for API responses.
type MembershipResponse struct {
	Body struct {
		Membership MembershipDTO `json:"membership"`
	}
}

// MembershipDetailResponse wraps a MembershipDTO with group context.
type MembershipDetailResponse struct {
	Body struct {
		Membership MembershipDTO `json:"membership"`
		Group      GroupDTO      `json:"group"`
	}
}

// MembershipListResponse wraps a list of MembershipDTOs for API responses.
type MembershipListResponse struct {
	Body struct {
		Memberships []MembershipDTO `json:"memberships"`
	}
}

// InvitationListResponse wraps a list of InvitationDTOs for API responses.
type InvitationListResponse struct {
	Body struct {
		Invitations []InvitationDTO `json:"invitations"`
	}
}
