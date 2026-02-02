// Package api provides HTTP handlers and DTOs for the authentication API.
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
