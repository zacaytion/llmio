// Package api provides HTTP handlers and middleware.
package api

import (
	"context"
	"log"
	"net/http"
)

// AuthFailureReason describes why an authentication attempt failed.
type AuthFailureReason string

const (
	ReasonUserNotFound       AuthFailureReason = "user_not_found"
	ReasonInvalidPassword    AuthFailureReason = "invalid_password"
	ReasonEmailNotVerified   AuthFailureReason = "email_not_verified"
	ReasonAccountDeactivated AuthFailureReason = "account_deactivated"
)

// LogAuthFailure logs an authentication failure for security auditing.
// The response to the client should remain generic to prevent enumeration.
func LogAuthFailure(ctx context.Context, email string, reason AuthFailureReason) {
	log.Printf("AUTH_FAILURE: email=%q reason=%s", email, reason)
}

// LogAuthFailureWithRequest logs an authentication failure with request details.
// Use when http.Request is available.
func LogAuthFailureWithRequest(ctx context.Context, email string, reason AuthFailureReason, r *http.Request) {
	ip := getClientIP(r)
	userAgent := r.UserAgent()
	log.Printf("AUTH_FAILURE: email=%q reason=%s ip=%s user_agent=%q",
		email, reason, ip, userAgent)
}

// LogDBError logs a database error for debugging and alerting.
func LogDBError(ctx context.Context, operation string, err error) {
	log.Printf("DB_ERROR: operation=%s error=%v", operation, err)
}

// LogRegistrationSuccess logs a successful user registration.
func LogRegistrationSuccess(ctx context.Context, email string, userID int64, r *http.Request) {
	ip := getClientIP(r)
	log.Printf("AUTH_REGISTER: email=%q user_id=%d ip=%s", email, userID, ip)
}

// LogLoginSuccess logs a successful login.
func LogLoginSuccess(ctx context.Context, email string, userID int64, r *http.Request) {
	ip := getClientIP(r)
	log.Printf("AUTH_LOGIN: email=%q user_id=%d ip=%s", email, userID, ip)
}

// LogLogout logs a logout event.
func LogLogout(ctx context.Context, userID int64, r *http.Request) {
	ip := getClientIP(r)
	log.Printf("AUTH_LOGOUT: user_id=%d ip=%s", userID, ip)
}

// getClientIP extracts the client IP from the request.
// Checks X-Forwarded-For first (for proxied requests), then falls back to RemoteAddr.
func getClientIP(r *http.Request) string {
	if r == nil {
		return "unknown"
	}

	// Check X-Forwarded-For header (set by proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs; take the first (original client)
		// Format: "client, proxy1, proxy2"
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	// Check X-Real-IP header (nginx convention)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
