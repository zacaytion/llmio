package auth

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

const (
	// sessionTokenBytes is the number of random bytes for session tokens (256 bits).
	sessionTokenBytes = 32
	// SessionDuration is how long sessions remain valid.
	SessionDuration = 7 * 24 * time.Hour
)

// Session represents an authenticated user's active login state.
type Session struct {
	Token     string    // Primary key (32 bytes, base64url)
	UserID    int64     // Foreign key to users.id
	CreatedAt time.Time // Session creation time
	ExpiresAt time.Time // Session expiration (CreatedAt + 7 days)
	UserAgent string    // Request User-Agent header
	IPAddress string    // Request IP address
}

// IsExpired returns true if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// SessionStore manages in-memory sessions.
type SessionStore struct {
	sessions sync.Map      // map[token]Session
	duration time.Duration // configurable session duration
}

// NewSessionStore creates a new in-memory session store with default duration.
func NewSessionStore() *SessionStore {
	return &SessionStore{
		duration: SessionDuration,
	}
}

// NewSessionStoreWithConfig creates a session store with custom duration.
func NewSessionStoreWithConfig(duration time.Duration) *SessionStore {
	return &SessionStore{
		duration: duration,
	}
}

// Create generates a new session for the given user.
func (s *SessionStore) Create(userID int64, userAgent, ipAddress string) (*Session, error) {
	token, err := generateSessionToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &Session{
		Token:     token,
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(s.duration),
		UserAgent: userAgent,
		IPAddress: ipAddress,
	}

	s.sessions.Store(token, session)
	return session, nil
}

// Get retrieves a session by token.
// Returns nil, false if the session doesn't exist or is expired.
func (s *SessionStore) Get(token string) (*Session, bool) {
	if token == "" {
		return nil, false
	}

	value, ok := s.sessions.Load(token)
	if !ok {
		return nil, false
	}

	session := value.(*Session)
	if session.IsExpired() {
		// Clean up expired session
		s.sessions.Delete(token)
		return nil, false
	}

	return session, true
}

// Delete removes a session by token.
func (s *SessionStore) Delete(token string) {
	s.sessions.Delete(token)
}

// GetByUserID returns all active sessions for a user.
func (s *SessionStore) GetByUserID(userID int64) []*Session {
	var sessions []*Session

	s.sessions.Range(func(key, value any) bool {
		session := value.(*Session)
		if session.UserID == userID && !session.IsExpired() {
			sessions = append(sessions, session)
		}
		return true
	})

	return sessions
}

// DeleteByUserID removes all sessions for a user.
func (s *SessionStore) DeleteByUserID(userID int64) {
	var toDelete []string

	s.sessions.Range(func(key, value any) bool {
		session := value.(*Session)
		if session.UserID == userID {
			toDelete = append(toDelete, session.Token)
		}
		return true
	})

	for _, token := range toDelete {
		s.sessions.Delete(token)
	}
}

// CleanupExpired removes all expired sessions.
// This should be called periodically (e.g., every 10 minutes).
func (s *SessionStore) CleanupExpired() int {
	var cleaned int
	var toDelete []string

	s.sessions.Range(func(key, value any) bool {
		session := value.(*Session)
		if session.IsExpired() {
			toDelete = append(toDelete, session.Token)
		}
		return true
	})

	for _, token := range toDelete {
		s.sessions.Delete(token)
		cleaned++
	}

	return cleaned
}

// generateSessionToken creates a cryptographically random session token.
func generateSessionToken() (string, error) {
	bytes := make([]byte, sessionTokenBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
