package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"github.com/zacaytion/llmio/internal/auth"
	"github.com/zacaytion/llmio/internal/db"
)

// dummyPasswordHash is a valid Argon2id hash used for timing-safe comparisons
// when the user doesn't exist. Generated at init to ensure consistent timing.
var dummyPasswordHash string

func init() {
	// Generate a real Argon2id hash for timing consistency
	hash, err := auth.HashPassword("dummy-timing-placeholder-password")
	if err != nil {
		panic("failed to generate dummy hash: " + err.Error())
	}
	dummyPasswordHash = hash
}

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	queries  *db.Queries
	sessions *auth.SessionStore
}

// NewAuthHandler creates a new authentication handler.
func NewAuthHandler(queries *db.Queries, sessions *auth.SessionStore) *AuthHandler {
	return &AuthHandler{
		queries:  queries,
		sessions: sessions,
	}
}

// RegisterRoutes registers all authentication routes.
func (h *AuthHandler) RegisterRoutes(api huma.API) {
	// Registration
	huma.Register(api, huma.Operation{
		OperationID:   "createRegistration",
		Method:        http.MethodPost,
		Path:          "/api/v1/registrations",
		Summary:       "Register a new user",
		Description:   "Creates a new user account with email and password.",
		Tags:          []string{"Authentication"},
		DefaultStatus: http.StatusCreated,
	}, h.handleRegistration)

	// Login
	huma.Register(api, huma.Operation{
		OperationID: "createSession",
		Method:      http.MethodPost,
		Path:        "/api/v1/sessions",
		Summary:     "Log in (create session)",
		Description: "Authenticates a user with email and password. Sets session cookie on success.",
		Tags:        []string{"Authentication"},
	}, h.handleLogin)

	// Logout
	huma.Register(api, huma.Operation{
		OperationID: "destroySession",
		Method:      http.MethodDelete,
		Path:        "/api/v1/sessions",
		Summary:     "Log out (destroy session)",
		Description: "Logs out the current user by invalidating their session.",
		Tags:        []string{"Authentication"},
	}, h.handleLogout)

	// Get current user
	huma.Register(api, huma.Operation{
		OperationID: "getCurrentSession",
		Method:      http.MethodGet,
		Path:        "/api/v1/sessions/me",
		Summary:     "Get current user",
		Description: "Returns the currently authenticated user's information.",
		Tags:        []string{"Authentication"},
	}, h.handleGetCurrentSession)
}

// RegistrationInput is the request body for user registration.
type RegistrationInput struct {
	Body struct {
		Email                string `json:"email" required:"true" format:"email" doc:"User's email address (case-insensitive)"`
		Name                 string `json:"name" required:"true" minLength:"1" doc:"User's display name"`
		Password             string `json:"password" required:"true" minLength:"8" doc:"Password (minimum 8 characters)"`
		PasswordConfirmation string `json:"password_confirmation" required:"true" doc:"Must match password"`
	}
}

// RegistrationOutput is the response body for successful registration.
type RegistrationOutput struct {
	Body struct {
		User UserDTO `json:"user"`
	}
}

func (h *AuthHandler) handleRegistration(ctx context.Context, input *RegistrationInput) (*RegistrationOutput, error) {
	// Normalize email
	email := strings.ToLower(strings.TrimSpace(input.Body.Email))
	name := strings.TrimSpace(input.Body.Name)

	// Validate name is not empty
	if name == "" {
		return nil, huma.Error422UnprocessableEntity("Name is required",
			&huma.ErrorDetail{
				Location: "body.name",
				Message:  "Name is required",
				Value:    input.Body.Name,
			})
	}

	// Validate password length (Huma minLength should catch this, but belt-and-suspenders)
	if len(input.Body.Password) < 8 {
		return nil, huma.Error422UnprocessableEntity("Password must be at least 8 characters",
			&huma.ErrorDetail{
				Location: "body.password",
				Message:  "Password must be at least 8 characters",
			})
	}

	// Validate password confirmation
	if input.Body.Password != input.Body.PasswordConfirmation {
		return nil, huma.Error422UnprocessableEntity("Passwords do not match",
			&huma.ErrorDetail{
				Location: "body.password_confirmation",
				Message:  "Passwords do not match",
			})
	}

	// Check if email already exists
	emailExists, err := h.queries.EmailExists(ctx, email)
	if err != nil {
		LogDBError(ctx, "EmailExists", err)
		return nil, huma.Error500InternalServerError("Database error")
	}
	if emailExists {
		return nil, huma.Error409Conflict("Email already taken",
			&huma.ErrorDetail{
				Location: "body.email",
				Message:  "Email already taken",
				Value:    email,
			})
	}

	// Hash password
	passwordHash, err := auth.HashPassword(input.Body.Password)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to process password")
	}

	// Generate unique username
	// Track if we hit a DB error during uniqueness checks
	var usernameDBError error
	baseUsername := auth.GenerateUsername(name)
	username := auth.MakeUsernameUnique(baseUsername, func(u string) bool {
		exists, err := h.queries.UsernameExists(ctx, u)
		if err != nil {
			usernameDBError = err
			return true // Conservatively treat as "exists" to retry with different username
		}
		return exists
	})
	if usernameDBError != nil {
		LogDBError(ctx, "UsernameExists", usernameDBError)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Generate unique public key
	var keyDBError error
	key, err := auth.MakePublicKeyUnique(func(k string) bool {
		_, err := h.queries.GetUserByKey(ctx, k)
		if err != nil {
			if db.IsNotFound(err) {
				return false // Key doesn't exist, it's available
			}
			keyDBError = err
			return true // Conservatively treat as "exists" to retry
		}
		return true // Key exists
	})
	if keyDBError != nil {
		LogDBError(ctx, "GetUserByKey", keyDBError)
		return nil, huma.Error500InternalServerError("Database error")
	}
	if err != nil {
		LogDBError(ctx, "MakePublicKeyUnique", err)
		return nil, huma.Error500InternalServerError("Failed to generate unique key")
	}

	// Create user in database
	user, err := h.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		Name:         name,
		Username:     username,
		PasswordHash: passwordHash,
		Key:          key,
	})
	if err != nil {
		LogDBError(ctx, "CreateUser", err)
		return nil, huma.Error500InternalServerError("Failed to create user")
	}

	// Build response
	output := &RegistrationOutput{}
	output.Body.User = UserDTOFromUser(user)

	return output, nil
}

// LoginInput is the request body for login.
type LoginInput struct {
	Body struct {
		Email    string `json:"email" required:"true" format:"email" doc:"Registered email address"`
		Password string `json:"password" required:"true" doc:"Account password"`
	}
}

// LoginOutput is the response body for successful login.
type LoginOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
	Body      struct {
		User UserDTO `json:"user"`
	}
}

func (h *AuthHandler) handleLogin(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
	email := strings.ToLower(strings.TrimSpace(input.Body.Email))

	// Look up user by email
	user, err := h.queries.GetUserByEmail(ctx, email)

	// Check for database errors (not just "not found")
	var userExists bool
	if err != nil {
		if !db.IsNotFound(err) {
			// Actual database error - log and return 500
			LogDBError(ctx, "GetUserByEmail", err)
			return nil, huma.Error500InternalServerError("Database error")
		}
		userExists = false
	} else {
		userExists = true
	}

	// Always do a password check to prevent timing attacks (even if user not found)
	// Use the pre-generated valid Argon2id hash to ensure consistent timing
	hashToCheck := dummyPasswordHash
	if userExists {
		hashToCheck = user.PasswordHash
	}

	passwordValid := auth.VerifyPassword(input.Body.Password, hashToCheck)

	// Check all conditions with same error message (no enumeration)
	// Log failures for security auditing
	if !userExists {
		LogAuthFailure(ctx, email, ReasonUserNotFound)
		return nil, huma.Error401Unauthorized("Invalid credentials")
	}
	if !passwordValid {
		LogAuthFailure(ctx, email, ReasonInvalidPassword)
		return nil, huma.Error401Unauthorized("Invalid credentials")
	}
	if !user.EmailVerified {
		LogAuthFailure(ctx, email, ReasonEmailNotVerified)
		return nil, huma.Error401Unauthorized("Invalid credentials")
	}
	if user.DeactivatedAt.Valid {
		LogAuthFailure(ctx, email, ReasonAccountDeactivated)
		return nil, huma.Error401Unauthorized("Invalid credentials")
	}

	// Create session
	session, err := h.sessions.Create(user.ID, "", "") // TODO: extract user agent and IP
	if err != nil {
		LogDBError(ctx, "sessions.Create", err)
		return nil, huma.Error500InternalServerError("Failed to create session")
	}

	// Build response with cookie
	output := &LoginOutput{
		SetCookie: http.Cookie{
			Name:     "loomio_session",
			Value:    session.Token,
			Path:     "/",
			MaxAge:   int(auth.SessionDuration.Seconds()),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		},
	}
	output.Body.User = UserDTOFromUser(user)

	return output, nil
}

// LogoutInput is the request for logout (requires session cookie).
type LogoutInput struct {
	Cookie string `cookie:"loomio_session"`
}

// LogoutOutput is the response for logout.
type LogoutOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
	Body      struct {
		Success bool `json:"success"`
	}
}

func (h *AuthHandler) handleLogout(ctx context.Context, input *LogoutInput) (*LogoutOutput, error) {
	if input.Cookie == "" {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// Validate session exists
	_, found := h.sessions.Get(input.Cookie)
	if !found {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// Delete session
	h.sessions.Delete(input.Cookie)

	// Return success with cleared cookie
	output := &LogoutOutput{
		SetCookie: http.Cookie{
			Name:     "loomio_session",
			Value:    "",
			Path:     "/",
			MaxAge:   -1, // Instructs browser to delete cookie
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		},
	}
	output.Body.Success = true
	return output, nil
}

// GetCurrentSessionInput is the request for getting current user.
type GetCurrentSessionInput struct {
	Cookie string `cookie:"loomio_session"`
}

// GetCurrentSessionOutput is the response for get current user.
type GetCurrentSessionOutput struct {
	Body struct {
		User UserDTO `json:"user"`
	}
}

func (h *AuthHandler) handleGetCurrentSession(ctx context.Context, input *GetCurrentSessionInput) (*GetCurrentSessionOutput, error) {
	if input.Cookie == "" {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// Validate session exists and not expired
	session, found := h.sessions.Get(input.Cookie)
	if !found {
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	// Look up user by ID
	user, err := h.queries.GetUserByID(ctx, session.UserID)
	if err != nil {
		if db.IsNotFound(err) {
			// User was deleted but session still exists - clean up
			h.sessions.Delete(input.Cookie)
			return nil, huma.Error401Unauthorized("Not authenticated")
		}
		// Actual database error
		LogDBError(ctx, "GetUserByID", err)
		return nil, huma.Error500InternalServerError("Database error")
	}

	// Check if user is deactivated
	if user.DeactivatedAt.Valid {
		h.sessions.Delete(input.Cookie)
		return nil, huma.Error401Unauthorized("Not authenticated")
	}

	output := &GetCurrentSessionOutput{}
	output.Body.User = UserDTOFromUser(user)
	return output, nil
}
