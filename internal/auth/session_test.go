package auth

import (
	"testing"
	"time"
)

func TestSessionStore_Create(t *testing.T) {
	store := NewSessionStore()

	session, err := store.Create(123, "Mozilla/5.0", "192.168.1.1")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Token should be 43 chars (32 bytes base64url encoded)
	if len(session.Token) != 43 {
		t.Errorf("Create() token length = %d, want 43", len(session.Token))
	}

	// UserID should match
	if session.UserID != 123 {
		t.Errorf("Create() UserID = %d, want 123", session.UserID)
	}

	// Expiry should be ~7 days from now
	expectedExpiry := time.Now().Add(7 * 24 * time.Hour)
	if session.ExpiresAt.Before(expectedExpiry.Add(-time.Minute)) ||
		session.ExpiresAt.After(expectedExpiry.Add(time.Minute)) {
		t.Errorf("Create() ExpiresAt = %v, want ~%v", session.ExpiresAt, expectedExpiry)
	}
}

func TestSessionStore_Get(t *testing.T) {
	store := NewSessionStore()

	// Create a session
	created, _ := store.Create(123, "Mozilla/5.0", "192.168.1.1")

	tests := []struct {
		name      string
		token     string
		wantFound bool
	}{
		{
			name:      "existing session",
			token:     created.Token,
			wantFound: true,
		},
		{
			name:      "non-existent session",
			token:     "nonexistenttoken123456789012345678901234",
			wantFound: false,
		},
		{
			name:      "empty token",
			token:     "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, found := store.Get(tt.token)
			if found != tt.wantFound {
				t.Errorf("Get() found = %v, want %v", found, tt.wantFound)
			}
			if tt.wantFound && session.UserID != 123 {
				t.Errorf("Get() UserID = %d, want 123", session.UserID)
			}
		})
	}
}

func TestSessionStore_Delete(t *testing.T) {
	store := NewSessionStore()

	// Create a session
	created, _ := store.Create(123, "Mozilla/5.0", "192.168.1.1")

	// Delete it
	store.Delete(created.Token)

	// Should not be found anymore
	_, found := store.Get(created.Token)
	if found {
		t.Error("Get() should not find deleted session")
	}
}

func TestSessionStore_Expiry(t *testing.T) {
	store := NewSessionStore()

	// Create a session with a very short expiry for testing
	session := &Session{
		Token:     "testtoken12345678901234567890123456789012",
		UserID:    123,
		CreatedAt: time.Now().Add(-8 * 24 * time.Hour), // Created 8 days ago
		ExpiresAt: time.Now().Add(-24 * time.Hour),     // Expired 1 day ago
		UserAgent: "Mozilla/5.0",
		IPAddress: "192.168.1.1",
	}
	store.sessions.Store(session.Token, session)

	// Should not be found (expired)
	_, found := store.Get(session.Token)
	if found {
		t.Error("Get() should not return expired session")
	}
}

func TestSessionStore_GetByUserID(t *testing.T) {
	store := NewSessionStore()

	// Create multiple sessions for the same user
	_, _ = store.Create(123, "Chrome", "192.168.1.1")
	_, _ = store.Create(123, "Firefox", "192.168.1.2")
	_, _ = store.Create(456, "Safari", "192.168.1.3")

	// Get sessions for user 123
	sessions := store.GetByUserID(123)
	if len(sessions) != 2 {
		t.Errorf("GetByUserID() returned %d sessions, want 2", len(sessions))
	}

	// Get sessions for user 456
	sessions = store.GetByUserID(456)
	if len(sessions) != 1 {
		t.Errorf("GetByUserID() returned %d sessions, want 1", len(sessions))
	}
}

func TestSessionStore_DeleteByUserID(t *testing.T) {
	store := NewSessionStore()

	// Create sessions for multiple users
	s1, _ := store.Create(123, "Chrome", "192.168.1.1")
	s2, _ := store.Create(123, "Firefox", "192.168.1.2")
	s3, _ := store.Create(456, "Safari", "192.168.1.3")

	// Delete all sessions for user 123
	store.DeleteByUserID(123)

	// User 123's sessions should be gone
	if _, found := store.Get(s1.Token); found {
		t.Error("Session s1 should be deleted")
	}
	if _, found := store.Get(s2.Token); found {
		t.Error("Session s2 should be deleted")
	}

	// User 456's session should remain
	if _, found := store.Get(s3.Token); !found {
		t.Error("Session s3 should still exist")
	}
}

func TestSessionStore_Cleanup(t *testing.T) {
	store := NewSessionStore()

	// Create a valid session
	validSession, _ := store.Create(123, "Chrome", "192.168.1.1")

	// Create expired sessions directly in the store
	expiredSession1 := &Session{
		Token:     "expiredtoken1234567890123456789012345678901",
		UserID:    456,
		CreatedAt: time.Now().Add(-8 * 24 * time.Hour),
		ExpiresAt: time.Now().Add(-24 * time.Hour), // Expired 1 day ago
		UserAgent: "Firefox",
		IPAddress: "192.168.1.2",
	}
	expiredSession2 := &Session{
		Token:     "expiredtoken2345678901234567890123456789012",
		UserID:    789,
		CreatedAt: time.Now().Add(-10 * 24 * time.Hour),
		ExpiresAt: time.Now().Add(-3 * 24 * time.Hour), // Expired 3 days ago
		UserAgent: "Safari",
		IPAddress: "192.168.1.3",
	}
	store.sessions.Store(expiredSession1.Token, expiredSession1)
	store.sessions.Store(expiredSession2.Token, expiredSession2)

	// Run cleanup
	cleaned := store.CleanupExpired()

	// Should have cleaned 2 expired sessions
	if cleaned != 2 {
		t.Errorf("CleanupExpired() = %d, want 2", cleaned)
	}

	// Expired sessions should be gone (they wouldn't be returned by Get anyway,
	// but they should also be removed from the internal store)
	count := 0
	store.sessions.Range(func(key, value any) bool {
		count++
		return true
	})
	if count != 1 {
		t.Errorf("Store should have 1 session after cleanup, got %d", count)
	}

	// Valid session should still exist
	if _, found := store.Get(validSession.Token); !found {
		t.Error("Valid session should still exist after cleanup")
	}
}
