package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/zacaytion/llmio/internal/auth"
	"github.com/zacaytion/llmio/internal/db"
)

// GroupHandler handles group-related HTTP requests.
type GroupHandler struct {
	pool     *pgxpool.Pool
	queries  *db.Queries
	sessions *auth.SessionStore
}

// NewGroupHandler creates a new group handler.
func NewGroupHandler(pool *pgxpool.Pool, queries *db.Queries, sessions *auth.SessionStore) *GroupHandler {
	return &GroupHandler{
		pool:     pool,
		queries:  queries,
		sessions: sessions,
	}
}

// RegisterRoutes registers all group routes.
func (h *GroupHandler) RegisterRoutes(api huma.API) {
	// Create group
	huma.Register(api, huma.Operation{
		OperationID:   "createGroup",
		Method:        http.MethodPost,
		Path:          "/api/v1/groups",
		Summary:       "Create a new group",
		Description:   "Creates a new group with the authenticated user as the first admin.",
		Tags:          []string{"Groups"},
		DefaultStatus: http.StatusCreated,
	}, h.handleCreateGroup)

	// Get group
	huma.Register(api, huma.Operation{
		OperationID: "getGroup",
		Method:      http.MethodGet,
		Path:        "/api/v1/groups/{id}",
		Summary:     "Get group details",
		Description: "Returns detailed group information including permission flags and counts.",
		Tags:        []string{"Groups"},
	}, h.handleGetGroup)

	// Update group
	huma.Register(api, huma.Operation{
		OperationID: "updateGroup",
		Method:      http.MethodPatch,
		Path:        "/api/v1/groups/{id}",
		Summary:     "Update group",
		Description: "Updates group settings and permission flags. Requires admin role.",
		Tags:        []string{"Groups"},
	}, h.handleUpdateGroup)

	// Create subgroup
	huma.Register(api, huma.Operation{
		OperationID:   "createSubgroup",
		Method:        http.MethodPost,
		Path:          "/api/v1/groups/{id}/subgroups",
		Summary:       "Create subgroup",
		Description:   "Creates a subgroup under the specified parent group. Requires admin role or members_can_create_subgroups permission.",
		Tags:          []string{"Groups"},
		DefaultStatus: http.StatusCreated,
	}, h.handleCreateSubgroup)

	// List subgroups
	huma.Register(api, huma.Operation{
		OperationID: "listSubgroups",
		Method:      http.MethodGet,
		Path:        "/api/v1/groups/{id}/subgroups",
		Summary:     "List subgroups",
		Description: "Returns all subgroups under the specified parent group.",
		Tags:        []string{"Groups"},
	}, h.handleListSubgroups)

	// Archive group
	huma.Register(api, huma.Operation{
		OperationID: "archiveGroup",
		Method:      http.MethodPost,
		Path:        "/api/v1/groups/{id}/archive",
		Summary:     "Archive group",
		Description: "Archives a group. Requires admin role.",
		Tags:        []string{"Groups"},
	}, h.handleArchiveGroup)

	// Unarchive group
	huma.Register(api, huma.Operation{
		OperationID: "unarchiveGroup",
		Method:      http.MethodPost,
		Path:        "/api/v1/groups/{id}/unarchive",
		Summary:     "Unarchive group",
		Description: "Restores an archived group. Requires admin role.",
		Tags:        []string{"Groups"},
	}, h.handleUnarchiveGroup)

	// List groups (user's memberships)
	huma.Register(api, huma.Operation{
		OperationID: "listGroups",
		Method:      http.MethodGet,
		Path:        "/api/v1/groups",
		Summary:     "List groups",
		Description: "Returns all groups the current user is a member of.",
		Tags:        []string{"Groups"},
	}, h.handleListGroups)

	// Get group by handle - using separate path to avoid conflict with /groups/{id} pattern
	huma.Register(api, huma.Operation{
		OperationID: "getGroupByHandle",
		Method:      http.MethodGet,
		Path:        "/api/v1/group-by-handle/{handle}",
		Summary:     "Get group by handle",
		Description: "Returns detailed group information by handle.",
		Tags:        []string{"Groups"},
	}, h.handleGetGroupByHandle)
}

// CreateGroupInput is the request body for creating a group.
type CreateGroupInput struct {
	Cookie string `cookie:"loomio_session"`
	Body   struct {
		Name        string  `json:"name" required:"true" minLength:"1" maxLength:"255" doc:"Group name (1-255 chars)"`
		Handle      string  `json:"handle,omitempty" minLength:"3" maxLength:"100" doc:"URL-safe handle (auto-generated if not provided)"`
		Description *string `json:"description,omitempty" doc:"Optional group description"`
	}
}

// CreateGroupOutput is the response for creating a group.
type CreateGroupOutput struct {
	Body struct {
		Group GroupDTO `json:"group"`
	}
}

func (h *GroupHandler) handleCreateGroup(ctx context.Context, input *CreateGroupInput) (*CreateGroupOutput, error) {
	// Authenticate
	if input.Cookie == "" {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	session, found := h.sessions.Get(input.Cookie)
	if !found {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// Validate name
	name := strings.TrimSpace(input.Body.Name)
	if name == "" {
		return nil, huma.Error422UnprocessableEntity("Name is required",
			&huma.ErrorDetail{
				Location: "body.name",
				Message:  "Name is required",
			})
	}

	// Generate or validate handle
	handle := strings.TrimSpace(input.Body.Handle)
	if handle == "" {
		// Auto-generate handle from name
		var handleErr error
		queries := h.queries // Capture for closure
		handle = GenerateUniqueHandle(name, func(candidate string) bool {
			exists, err := queries.HandleExists(ctx, candidate)
			if err != nil {
				handleErr = err
				return true // Conservatively treat as "exists" on error
			}
			return exists
		})
		if handleErr != nil {
			LogDBError(ctx, "HandleExists", handleErr)
			return nil, huma.Error500InternalServerError("Database error")
		}
		if handle == "" {
			return nil, huma.Error422UnprocessableEntity("Could not generate handle from name",
				&huma.ErrorDetail{
					Location: "body.name",
					Message:  "Name is too short to generate a handle (need at least 3 alphanumeric characters)",
				})
		}
	} else {
		// T139: Validate user-provided handle format
		handle = strings.ToLower(handle) // Normalize to lowercase
		if !isValidHandle(handle) {
			return nil, huma.Error422UnprocessableEntity("Invalid handle format",
				&huma.ErrorDetail{
					Location: "body.handle",
					Message:  "Handle must be 3-100 characters, start and end with alphanumeric, contain only lowercase letters, numbers, and hyphens",
					Value:    handle,
				})
		}
	}

	// Build description as pgtype.Text
	var description pgtype.Text
	if input.Body.Description != nil && *input.Body.Description != "" {
		description = pgtype.Text{String: *input.Body.Description, Valid: true}
	}

	// Execute in transaction with audit context
	var group *db.Group
	err := pgx.BeginTxFunc(ctx, h.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		// Set audit context for triggers
		if auditErr := db.SetAuditContext(ctx, tx, session.UserID); auditErr != nil {
			return fmt.Errorf("SetAuditContext: %w", auditErr)
		}

		txQueries := h.queries.WithTx(tx)

		// Create group
		var createErr error
		group, createErr = txQueries.CreateGroup(ctx, db.CreateGroupParams{
			Name:        name,
			Handle:      handle,
			Description: description,
			CreatedByID: session.UserID,
			// Permission flags use database defaults (COALESCE in query)
		})
		if createErr != nil {
			// Check for unique constraint violation on handle
			if isUniqueViolation(createErr, "groups_handle_key") {
				return huma.Error409Conflict("Handle already taken",
					&huma.ErrorDetail{
						Location: "body.handle",
						Message:  "Handle already taken",
						Value:    handle,
					})
			}
			return fmt.Errorf("CreateGroup: %w", createErr)
		}

		// Create admin membership for creator (auto-accepted)
		_, membershipErr := txQueries.CreateMembership(ctx, db.CreateMembershipParams{
			GroupID:   group.ID,
			UserID:    session.UserID,
			Role:      RoleAdmin,
			InviterID: session.UserID, // Self-invited
			AcceptedAt: pgtype.Timestamptz{
				Time:  group.CreatedAt.Time, // Same timestamp as group creation
				Valid: true,
			},
		})
		if membershipErr != nil {
			return fmt.Errorf("CreateMembership: %w", membershipErr)
		}
		return nil
	})

	if err != nil {
		// Check if it's already a Huma error (from conflict handling)
		if humaErr, ok := err.(huma.StatusError); ok {
			return nil, humaErr
		}
		LogDBError(ctx, "CreateGroup", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Build response
	output := &CreateGroupOutput{}
	output.Body.Group = GroupDTOFromGroup(group)
	return output, nil
}

// isUniqueViolation checks if the error is a PostgreSQL unique constraint violation.
// T156: Uses pgconn.PgError type assertion for reliable error detection, even with wrapped errors.
func isUniqueViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// 23505 is the PostgreSQL error code for unique_violation
		return pgErr.Code == "23505" && pgErr.ConstraintName == constraintName
	}
	return false
}

// handleRegex matches valid handle characters after initial slugification.
var handleRegex = regexp.MustCompile(`[^a-z0-9-]+`)

// multiHyphenRegex matches multiple consecutive hyphens.
var multiHyphenRegex = regexp.MustCompile(`-+`)

// validHandleRegex validates the complete handle format:
// - 3-100 characters
// - Starts with alphanumeric, ends with alphanumeric
// - Only lowercase alphanumeric and hyphens in between
// T139: Handle format validation regex
var validHandleRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

// isValidHandle validates a handle against format requirements.
// Handles must be 3-100 chars, start/end with alphanumeric, only contain lowercase alphanumeric and hyphens.
func isValidHandle(handle string) bool {
	if len(handle) < 3 || len(handle) > 100 {
		return false
	}
	return validHandleRegex.MatchString(handle)
}

// GenerateHandle creates a URL-safe handle from a group name.
// The handle is:
//   - Lowercased
//   - Transliterated (accented chars → ASCII)
//   - Non-alphanumeric chars replaced with hyphens
//   - Multiple hyphens collapsed to single hyphen
//   - Leading/trailing hyphens removed
//   - Truncated to 100 characters max
//
// If the result is shorter than 3 characters, it returns an empty string
// (the caller should handle this by requiring an explicit handle).
func GenerateHandle(name string) string {
	if name == "" {
		return ""
	}

	// Step 1: Normalize Unicode (NFD decomposition)
	// T137: Handle transform.String errors with fallback to original name
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, err := transform.String(t, name)
	if err != nil {
		// Fallback to original name on transform error (e.g., invalid UTF-8)
		result = name
	}

	// Step 2: Lowercase
	result = strings.ToLower(result)

	// Step 3: Replace non-alphanumeric with hyphens
	result = handleRegex.ReplaceAllString(result, "-")

	// Step 4: Collapse multiple hyphens
	result = multiHyphenRegex.ReplaceAllString(result, "-")

	// Step 5: Trim leading/trailing hyphens
	result = strings.Trim(result, "-")

	// Step 6: Truncate to max length (100 chars)
	if len(result) > 100 {
		result = result[:100]
		// Re-trim if we cut mid-hyphen
		result = strings.Trim(result, "-")
	}

	// Step 7: Validate minimum length (3 chars)
	if len(result) < 3 {
		return ""
	}

	return result
}

// GenerateUniqueHandle generates a handle and appends a numeric suffix if needed
// to ensure uniqueness. The checkExists function should return true if the handle
// already exists in the database.
//
// Example:
//
//	handle := GenerateUniqueHandle("Climate Team", func(h string) bool {
//	    exists, _ := queries.HandleExists(ctx, h)
//	    return exists.Exists
//	})
//
// Returns: "climate-team", "climate-team-1", "climate-team-2", etc.
func GenerateUniqueHandle(name string, checkExists func(handle string) bool) string {
	base := GenerateHandle(name)
	if base == "" {
		return ""
	}

	// Try the base handle first
	if !checkExists(base) {
		return base
	}

	// Try numeric suffixes
	for i := 1; i <= 1000; i++ {
		candidate := base + "-" + itoa(i)
		// Ensure we don't exceed max length
		if len(candidate) > 100 {
			// Truncate base to make room for suffix
			maxBase := 100 - len("-") - len(itoa(i))
			if maxBase < 3 {
				return "" // Can't generate valid handle
			}
			candidate = base[:maxBase] + "-" + itoa(i)
		}
		if !checkExists(candidate) {
			return candidate
		}
	}

	// Extremely unlikely: 1000 collisions
	return ""
}

// itoa converts an int to string without importing strconv.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}

// ============================================================
// GET /api/v1/groups/{id} - Get group details
// ============================================================

// GetGroupInput is the request for getting a group.
type GetGroupInput struct {
	Cookie string `cookie:"loomio_session"`
	ID     int64  `path:"id" doc:"Group ID"`
}

// GetGroupOutput is the response for getting a group.
type GetGroupOutput struct {
	Body struct {
		Group GroupDetailDTO `json:"group"`
	}
}

func (h *GroupHandler) handleGetGroup(ctx context.Context, input *GetGroupInput) (*GetGroupOutput, error) {
	// Authenticate
	if input.Cookie == "" {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	session, found := h.sessions.Get(input.Cookie)
	if !found {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// Authorize: user must be a member of the group
	authCtx, err := NewAuthorizationContext(ctx, h.queries, session.UserID, input.ID)
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

	// T170: Get member stats with single query (more efficient than two separate queries)
	stats, err := h.queries.CountGroupMembershipStats(ctx, input.ID)
	if err != nil {
		LogDBError(ctx, "CountGroupMembershipStats", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Build response with full details
	output := &GetGroupOutput{}
	output.Body.Group = GroupDetailDTOFromGroup(authCtx.Group, stats.MemberCount, stats.AdminCount, authCtx.GetRole())

	// Check if parent is archived (for subgroups - T103a)
	// T121: Log error when parent fetch fails, don't silently suppress
	if authCtx.Group.ParentID.Valid {
		parentGroup, err := h.queries.GetGroupByID(ctx, authCtx.Group.ParentID.Int64)
		if err != nil {
			// Log the error but don't fail the request - parent info is supplementary
			LogDBError(ctx, "GetParentGroup", err)
		} else if parentGroup.ArchivedAt.Valid {
			t := true
			output.Body.Group.ParentArchived = &t
		}
	}

	return output, nil
}

// ============================================================
// PATCH /api/v1/groups/{id} - Update group settings
// ============================================================

// UpdateGroupInput is the request for updating a group.
type UpdateGroupInput struct {
	Cookie string `cookie:"loomio_session"`
	ID     int64  `path:"id" doc:"Group ID"`
	Body   struct {
		Name                           *string `json:"name,omitempty" minLength:"1" maxLength:"255" doc:"Group name"`
		Description                    *string `json:"description,omitempty" doc:"Group description"`
		MembersCanAddMembers           *bool   `json:"members_can_add_members,omitempty" doc:"Members can invite others"`
		MembersCanAddGuests            *bool   `json:"members_can_add_guests,omitempty" doc:"Members can add discussion guests"`
		MembersCanStartDiscussions     *bool   `json:"members_can_start_discussions,omitempty" doc:"Members can create discussions"`
		MembersCanRaiseMotions         *bool   `json:"members_can_raise_motions,omitempty" doc:"Members can create polls"`
		MembersCanEditDiscussions      *bool   `json:"members_can_edit_discussions,omitempty" doc:"Members can edit discussion titles"`
		MembersCanEditComments         *bool   `json:"members_can_edit_comments,omitempty" doc:"Members can edit own comments"`
		MembersCanDeleteComments       *bool   `json:"members_can_delete_comments,omitempty" doc:"Members can delete own comments"`
		MembersCanAnnounce             *bool   `json:"members_can_announce,omitempty" doc:"Members can send announcements"`
		MembersCanCreateSubgroups      *bool   `json:"members_can_create_subgroups,omitempty" doc:"Members can create subgroups"`
		AdminsCanEditUserContent       *bool   `json:"admins_can_edit_user_content,omitempty" doc:"Admins can edit any content"`
		ParentMembersCanSeeDiscussions *bool   `json:"parent_members_can_see_discussions,omitempty" doc:"Parent members see subgroup content"`
	}
}

// UpdateGroupOutput is the response for updating a group.
type UpdateGroupOutput struct {
	Body struct {
		Group GroupDetailDTO `json:"group"`
	}
}

func (h *GroupHandler) handleUpdateGroup(ctx context.Context, input *UpdateGroupInput) (*UpdateGroupOutput, error) {
	// Authenticate
	if input.Cookie == "" {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	session, found := h.sessions.Get(input.Cookie)
	if !found {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// Authorize: user must be an admin of the group
	authCtx, err := NewAuthorizationContext(ctx, h.queries, session.UserID, input.ID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, huma.Error404NotFound("Group not found")
		}
		LogDBError(ctx, "NewAuthorizationContext", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	if !authCtx.CanUpdateGroup() {
		return nil, huma.Error403Forbidden("Admin role required to update group")
	}

	// T142: Check if group is archived before allowing updates
	if authCtx.Group.ArchivedAt.Valid {
		return nil, huma.Error409Conflict("Cannot update an archived group")
	}

	// Build update params - use pgtype for nullable fields
	updateParams := db.UpdateGroupParams{
		ID: input.ID,
	}

	// Name update
	if input.Body.Name != nil {
		name := strings.TrimSpace(*input.Body.Name)
		if name == "" {
			return nil, huma.Error422UnprocessableEntity("Name cannot be empty",
				&huma.ErrorDetail{
					Location: "body.name",
					Message:  "Name cannot be empty",
				})
		}
		updateParams.Name = pgtype.Text{String: name, Valid: true}
	}

	// Description update
	if input.Body.Description != nil {
		updateParams.Description = pgtype.Text{String: *input.Body.Description, Valid: true}
	}

	// Permission flag updates - convert *bool to pgtype.Bool
	if input.Body.MembersCanAddMembers != nil {
		updateParams.MembersCanAddMembers = pgtype.Bool{Bool: *input.Body.MembersCanAddMembers, Valid: true}
	}
	if input.Body.MembersCanAddGuests != nil {
		updateParams.MembersCanAddGuests = pgtype.Bool{Bool: *input.Body.MembersCanAddGuests, Valid: true}
	}
	if input.Body.MembersCanStartDiscussions != nil {
		updateParams.MembersCanStartDiscussions = pgtype.Bool{Bool: *input.Body.MembersCanStartDiscussions, Valid: true}
	}
	if input.Body.MembersCanRaiseMotions != nil {
		updateParams.MembersCanRaiseMotions = pgtype.Bool{Bool: *input.Body.MembersCanRaiseMotions, Valid: true}
	}
	if input.Body.MembersCanEditDiscussions != nil {
		updateParams.MembersCanEditDiscussions = pgtype.Bool{Bool: *input.Body.MembersCanEditDiscussions, Valid: true}
	}
	if input.Body.MembersCanEditComments != nil {
		updateParams.MembersCanEditComments = pgtype.Bool{Bool: *input.Body.MembersCanEditComments, Valid: true}
	}
	if input.Body.MembersCanDeleteComments != nil {
		updateParams.MembersCanDeleteComments = pgtype.Bool{Bool: *input.Body.MembersCanDeleteComments, Valid: true}
	}
	if input.Body.MembersCanAnnounce != nil {
		updateParams.MembersCanAnnounce = pgtype.Bool{Bool: *input.Body.MembersCanAnnounce, Valid: true}
	}
	if input.Body.MembersCanCreateSubgroups != nil {
		updateParams.MembersCanCreateSubgroups = pgtype.Bool{Bool: *input.Body.MembersCanCreateSubgroups, Valid: true}
	}
	if input.Body.AdminsCanEditUserContent != nil {
		updateParams.AdminsCanEditUserContent = pgtype.Bool{Bool: *input.Body.AdminsCanEditUserContent, Valid: true}
	}
	if input.Body.ParentMembersCanSeeDiscussions != nil {
		updateParams.ParentMembersCanSeeDiscussions = pgtype.Bool{Bool: *input.Body.ParentMembersCanSeeDiscussions, Valid: true}
	}

	// Execute update in transaction with audit context
	var group *db.Group
	err = pgx.BeginTxFunc(ctx, h.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		// Set audit context for triggers
		if auditErr := db.SetAuditContext(ctx, tx, session.UserID); auditErr != nil {
			return fmt.Errorf("SetAuditContext: %w", auditErr)
		}

		txQueries := h.queries.WithTx(tx)
		var updateErr error
		group, updateErr = txQueries.UpdateGroup(ctx, updateParams)
		return updateErr
	})

	if err != nil {
		LogDBError(ctx, "UpdateGroup", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Get updated counts
	memberCount, err := h.queries.CountGroupMembers(ctx, input.ID)
	if err != nil {
		LogDBError(ctx, "CountGroupMembers", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	adminCount, err := h.queries.CountGroupAdmins(ctx, input.ID)
	if err != nil {
		LogDBError(ctx, "CountGroupAdmins", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Build response
	output := &UpdateGroupOutput{}
	output.Body.Group = GroupDetailDTOFromGroup(group, memberCount, adminCount, authCtx.GetRole())
	return output, nil
}

// ============================================================
// Subgroup handlers
// ============================================================

// CreateSubgroupInput is the request for creating a subgroup.
type CreateSubgroupInput struct {
	Cookie   string `cookie:"loomio_session"`
	ParentID int64  `path:"id" doc:"Parent group ID"`
	Body     struct {
		Name               string `json:"name" required:"true" minLength:"1" maxLength:"255" doc:"Subgroup name"`
		Handle             string `json:"handle,omitempty" minLength:"3" maxLength:"100" doc:"URL-safe handle (auto-generated if not provided)"`
		Description        *string `json:"description,omitempty" doc:"Optional description"`
		InheritPermissions *bool  `json:"inherit_permissions,omitempty" doc:"Copy parent's permission flags (defaults to false)"`
	}
}

// CreateSubgroupOutput is the response for creating a subgroup.
type CreateSubgroupOutput struct {
	Body struct {
		Group GroupDetailDTO `json:"group"`
	}
}

func (h *GroupHandler) handleCreateSubgroup(ctx context.Context, input *CreateSubgroupInput) (*CreateSubgroupOutput, error) {
	// Authenticate
	if input.Cookie == "" {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	session, found := h.sessions.Get(input.Cookie)
	if !found {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// Authorize: user must be able to create subgroups in the parent group
	authCtx, err := NewAuthorizationContext(ctx, h.queries, session.UserID, input.ParentID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, huma.Error404NotFound("Parent group not found")
		}
		LogDBError(ctx, "NewAuthorizationContext", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	if !authCtx.CanCreateSubgroups() {
		return nil, huma.Error403Forbidden("Permission to create subgroups required")
	}

	// Validate and prepare name
	name := strings.TrimSpace(input.Body.Name)
	if name == "" {
		return nil, huma.Error422UnprocessableEntity("Name is required",
			&huma.ErrorDetail{
				Location: "body.name",
				Message:  "Name is required",
			})
	}

	// Generate or validate handle
	handle := strings.TrimSpace(input.Body.Handle)
	if handle == "" {
		var handleErr error
		queries := h.queries
		handle = GenerateUniqueHandle(name, func(candidate string) bool {
			exists, err := queries.HandleExists(ctx, candidate)
			if err != nil {
				handleErr = err
				return true
			}
			return exists
		})
		if handleErr != nil {
			LogDBError(ctx, "HandleExists", handleErr)
			return nil, huma.Error500InternalServerError("Database error")
		}
		if handle == "" {
			return nil, huma.Error422UnprocessableEntity("Could not generate handle from name",
				&huma.ErrorDetail{
					Location: "body.name",
					Message:  "Name is too short to generate a handle",
				})
		}
	} else {
		// T140: Validate user-provided handle format for subgroups
		handle = strings.ToLower(handle) // Normalize to lowercase
		if !isValidHandle(handle) {
			return nil, huma.Error422UnprocessableEntity("Invalid handle format",
				&huma.ErrorDetail{
					Location: "body.handle",
					Message:  "Handle must be 3-100 characters, start and end with alphanumeric, contain only lowercase letters, numbers, and hyphens",
					Value:    handle,
				})
		}
	}

	// Build description
	var description pgtype.Text
	if input.Body.Description != nil && *input.Body.Description != "" {
		description = pgtype.Text{String: *input.Body.Description, Valid: true}
	}

	// Prepare create params
	createParams := db.CreateGroupParams{
		Name:        name,
		Handle:      handle,
		Description: description,
		ParentID:    pgtype.Int8{Int64: input.ParentID, Valid: true},
		CreatedByID: session.UserID,
	}

	// If inheriting permissions, copy from parent
	if input.Body.InheritPermissions != nil && *input.Body.InheritPermissions {
		parent := authCtx.Group
		createParams.MembersCanAddMembers = pgtype.Bool{Bool: parent.MembersCanAddMembers, Valid: true}
		createParams.MembersCanAddGuests = pgtype.Bool{Bool: parent.MembersCanAddGuests, Valid: true}
		createParams.MembersCanStartDiscussions = pgtype.Bool{Bool: parent.MembersCanStartDiscussions, Valid: true}
		createParams.MembersCanRaiseMotions = pgtype.Bool{Bool: parent.MembersCanRaiseMotions, Valid: true}
		createParams.MembersCanEditDiscussions = pgtype.Bool{Bool: parent.MembersCanEditDiscussions, Valid: true}
		createParams.MembersCanEditComments = pgtype.Bool{Bool: parent.MembersCanEditComments, Valid: true}
		createParams.MembersCanDeleteComments = pgtype.Bool{Bool: parent.MembersCanDeleteComments, Valid: true}
		createParams.MembersCanAnnounce = pgtype.Bool{Bool: parent.MembersCanAnnounce, Valid: true}
		createParams.MembersCanCreateSubgroups = pgtype.Bool{Bool: parent.MembersCanCreateSubgroups, Valid: true}
		createParams.AdminsCanEditUserContent = pgtype.Bool{Bool: parent.AdminsCanEditUserContent, Valid: true}
		createParams.ParentMembersCanSeeDiscussions = pgtype.Bool{Bool: parent.ParentMembersCanSeeDiscussions, Valid: true}
	}
	// If not inheriting, the COALESCE in the SQL query will use database defaults

	// Execute in transaction
	var group *db.Group
	err = pgx.BeginTxFunc(ctx, h.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		// Set audit context
		if auditErr := db.SetAuditContext(ctx, tx, session.UserID); auditErr != nil {
			return fmt.Errorf("SetAuditContext: %w", auditErr)
		}

		txQueries := h.queries.WithTx(tx)

		// Create group
		var createErr error
		group, createErr = txQueries.CreateGroup(ctx, createParams)
		if createErr != nil {
			if isUniqueViolation(createErr, "groups_handle_key") {
				return huma.Error409Conflict("Handle already taken",
					&huma.ErrorDetail{
						Location: "body.handle",
						Message:  "Handle already taken",
						Value:    handle,
					})
			}
			return fmt.Errorf("CreateGroup: %w", createErr)
		}

		// Create admin membership for creator
		_, membershipErr := txQueries.CreateMembership(ctx, db.CreateMembershipParams{
			GroupID:   group.ID,
			UserID:    session.UserID,
			Role:      RoleAdmin,
			InviterID: session.UserID,
			AcceptedAt: pgtype.Timestamptz{
				Time:  group.CreatedAt.Time,
				Valid: true,
			},
		})
		if membershipErr != nil {
			return fmt.Errorf("CreateMembership: %w", membershipErr)
		}
		return nil
	})

	if err != nil {
		if humaErr, ok := err.(huma.StatusError); ok {
			return nil, humaErr
		}
		LogDBError(ctx, "CreateSubgroup", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Build response - new subgroup has 1 member (creator) who is admin
	output := &CreateSubgroupOutput{}
	output.Body.Group = GroupDetailDTOFromGroup(group, 1, 1, RoleAdmin)
	return output, nil
}

// ListSubgroupsInput is the request for listing subgroups.
type ListSubgroupsInput struct {
	Cookie          string `cookie:"loomio_session"`
	ParentID        int64  `path:"id" doc:"Parent group ID"`
	IncludeArchived bool   `query:"include_archived" default:"false" doc:"Include archived subgroups"`
}

// ListSubgroupsOutput is the response for listing subgroups.
type ListSubgroupsOutput struct {
	Body struct {
		Groups []GroupDTO `json:"groups"`
	}
}

func (h *GroupHandler) handleListSubgroups(ctx context.Context, input *ListSubgroupsInput) (*ListSubgroupsOutput, error) {
	// Authenticate
	if input.Cookie == "" {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	session, found := h.sessions.Get(input.Cookie)
	if !found {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// Authorize: user must be a member of the parent group
	authCtx, err := NewAuthorizationContext(ctx, h.queries, session.UserID, input.ParentID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, huma.Error404NotFound("Parent group not found")
		}
		LogDBError(ctx, "NewAuthorizationContext", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	if !authCtx.CanViewGroup() {
		return nil, huma.Error403Forbidden("Not a member of this group")
	}

	// List subgroups
	subgroups, err := h.queries.ListSubgroupsByParent(ctx, db.ListSubgroupsByParentParams{
		ParentID:        pgtype.Int8{Int64: input.ParentID, Valid: true},
		IncludeArchived: input.IncludeArchived,
	})
	if err != nil {
		LogDBError(ctx, "ListSubgroupsByParent", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Build response
	output := &ListSubgroupsOutput{}
	output.Body.Groups = make([]GroupDTO, len(subgroups))
	for i, g := range subgroups {
		output.Body.Groups[i] = GroupDTOFromGroup(g)
	}
	return output, nil
}

// ============================================================
// Archive/unarchive and list handlers
// ============================================================

// ArchiveGroupInput is the request for archiving a group.
type ArchiveGroupInput struct {
	Cookie string `cookie:"loomio_session"`
	ID     int64  `path:"id" doc:"Group ID"`
}

// ArchiveGroupOutput is the response for archiving a group.
type ArchiveGroupOutput struct {
	Body struct {
		Group GroupDTO `json:"group"`
	}
}

func (h *GroupHandler) handleArchiveGroup(ctx context.Context, input *ArchiveGroupInput) (*ArchiveGroupOutput, error) {
	// Authenticate
	if input.Cookie == "" {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	session, found := h.sessions.Get(input.Cookie)
	if !found {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// Authorize: user must be an admin of the group
	authCtx, err := NewAuthorizationContext(ctx, h.queries, session.UserID, input.ID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, huma.Error404NotFound("Group not found")
		}
		LogDBError(ctx, "NewAuthorizationContext", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	if !authCtx.CanArchiveGroup() {
		return nil, huma.Error403Forbidden("Admin role required to archive group")
	}

	// Execute archive in transaction
	var group *db.Group
	err = pgx.BeginTxFunc(ctx, h.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if auditErr := db.SetAuditContext(ctx, tx, session.UserID); auditErr != nil {
			return fmt.Errorf("SetAuditContext: %w", auditErr)
		}

		txQueries := h.queries.WithTx(tx)
		var archiveErr error
		group, archiveErr = txQueries.ArchiveGroup(ctx, input.ID)
		return archiveErr
	})

	if err != nil {
		LogDBError(ctx, "ArchiveGroup", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	output := &ArchiveGroupOutput{}
	output.Body.Group = GroupDTOFromGroup(group)
	return output, nil
}

// UnarchiveGroupInput is the request for unarchiving a group.
type UnarchiveGroupInput struct {
	Cookie string `cookie:"loomio_session"`
	ID     int64  `path:"id" doc:"Group ID"`
}

// UnarchiveGroupOutput is the response for unarchiving a group.
type UnarchiveGroupOutput struct {
	Body struct {
		Group GroupDTO `json:"group"`
	}
}

func (h *GroupHandler) handleUnarchiveGroup(ctx context.Context, input *UnarchiveGroupInput) (*UnarchiveGroupOutput, error) {
	// Authenticate
	if input.Cookie == "" {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	session, found := h.sessions.Get(input.Cookie)
	if !found {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// Authorize: user must be an admin of the group
	authCtx, err := NewAuthorizationContext(ctx, h.queries, session.UserID, input.ID)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, huma.Error404NotFound("Group not found")
		}
		LogDBError(ctx, "NewAuthorizationContext", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	if !authCtx.CanArchiveGroup() {
		return nil, huma.Error403Forbidden("Admin role required to unarchive group")
	}

	// Execute unarchive in transaction
	var group *db.Group
	err = pgx.BeginTxFunc(ctx, h.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if auditErr := db.SetAuditContext(ctx, tx, session.UserID); auditErr != nil {
			return fmt.Errorf("SetAuditContext: %w", auditErr)
		}

		txQueries := h.queries.WithTx(tx)
		var unarchiveErr error
		group, unarchiveErr = txQueries.UnarchiveGroup(ctx, input.ID)
		return unarchiveErr
	})

	if err != nil {
		LogDBError(ctx, "UnarchiveGroup", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	output := &UnarchiveGroupOutput{}
	output.Body.Group = GroupDTOFromGroup(group)
	return output, nil
}

// ListGroupsInput is the request for listing groups.
type ListGroupsInput struct {
	Cookie          string `cookie:"loomio_session"`
	IncludeArchived bool   `query:"include_archived" default:"false" doc:"Include archived groups"`
}

// ListGroupsOutput is the response for listing groups.
type ListGroupsOutput struct {
	Body struct {
		Groups []GroupDTO `json:"groups"`
	}
}

func (h *GroupHandler) handleListGroups(ctx context.Context, input *ListGroupsInput) (*ListGroupsOutput, error) {
	// Authenticate
	if input.Cookie == "" {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	session, found := h.sessions.Get(input.Cookie)
	if !found {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// List groups the user is a member of
	groups, err := h.queries.ListGroupsByUser(ctx, db.ListGroupsByUserParams{
		UserID:          session.UserID,
		IncludeArchived: input.IncludeArchived,
	})
	if err != nil {
		LogDBError(ctx, "ListGroupsByUser", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Build response
	output := &ListGroupsOutput{}
	output.Body.Groups = make([]GroupDTO, len(groups))
	for i, g := range groups {
		output.Body.Groups[i] = GroupDTOFromGroup(g)
	}
	return output, nil
}

// GetGroupByHandleInput is the request for getting a group by handle.
type GetGroupByHandleInput struct {
	Cookie string `cookie:"loomio_session"`
	Handle string `path:"handle" doc:"Group handle (URL-safe)"`
}

// GetGroupByHandleOutput is the response for getting a group by handle.
type GetGroupByHandleOutput struct {
	Body struct {
		Group GroupDetailDTO `json:"group"`
	}
}

func (h *GroupHandler) handleGetGroupByHandle(ctx context.Context, input *GetGroupByHandleInput) (*GetGroupByHandleOutput, error) {
	// Authenticate
	if input.Cookie == "" {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	session, found := h.sessions.Get(input.Cookie)
	if !found {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// Get group by handle
	group, err := h.queries.GetGroupByHandle(ctx, input.Handle)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, huma.Error404NotFound("Group not found")
		}
		LogDBError(ctx, "GetGroupByHandle", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Authorize: user must be a member of the group
	authCtx, err := NewAuthorizationContext(ctx, h.queries, session.UserID, group.ID)
	if err != nil {
		LogDBError(ctx, "NewAuthorizationContext", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	if !authCtx.CanViewGroup() {
		return nil, huma.Error403Forbidden("Not a member of this group")
	}

	// Get member counts
	memberCount, err := h.queries.CountGroupMembers(ctx, group.ID)
	if err != nil {
		LogDBError(ctx, "CountGroupMembers", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	adminCount, err := h.queries.CountGroupAdmins(ctx, group.ID)
	if err != nil {
		LogDBError(ctx, "CountGroupAdmins", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Build response with full details
	output := &GetGroupByHandleOutput{}
	output.Body.Group = GroupDetailDTOFromGroup(group, memberCount, adminCount, authCtx.GetRole())

	// Check if parent is archived (for subgroups)
	// T122: Log error when parent fetch fails, don't silently suppress
	if group.ParentID.Valid {
		parentGroup, err := h.queries.GetGroupByID(ctx, group.ParentID.Int64)
		if err != nil {
			// Log the error but don't fail the request - parent info is supplementary
			LogDBError(ctx, "GetParentGroup", err)
		} else if parentGroup.ArchivedAt.Valid {
			t := true
			output.Body.Group.ParentArchived = &t
		}
	}

	return output, nil
}
