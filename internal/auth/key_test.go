package auth

import (
	"errors"
	"regexp"
	"testing"
)

// keyPattern matches valid public keys (22 chars, base64url safe).
var keyPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{22}$`)

func TestGeneratePublicKey(t *testing.T) {
	key := GeneratePublicKey()

	// Must be exactly 22 characters (128 bits encoded in base64url)
	if len(key) != 22 {
		t.Errorf("GeneratePublicKey() = %q, want 22 chars, got %d", key, len(key))
	}

	// Must match base64url pattern
	if !keyPattern.MatchString(key) {
		t.Errorf("GeneratePublicKey() = %q, does not match pattern ^[a-zA-Z0-9_-]{22}$", key)
	}
}

func TestGeneratePublicKeyUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	collisions := 0

	// Generate 1000 keys and check for collisions
	for i := 0; i < 1000; i++ {
		key := GeneratePublicKey()
		if seen[key] {
			collisions++
		}
		seen[key] = true
	}

	if collisions > 0 {
		t.Errorf("GeneratePublicKey() had %d collisions in 1000 generations", collisions)
	}
}

func TestMakePublicKeyUnique(t *testing.T) {
	// First key is always "taken"
	takenKey := GeneratePublicKey()

	checker := func(key string) bool {
		return key == takenKey
	}

	unique, err := MakePublicKeyUnique(checker)
	if err != nil {
		t.Fatalf("MakePublicKeyUnique() error = %v, want nil", err)
	}

	// Should be different from taken key
	if unique == takenKey {
		t.Errorf("MakePublicKeyUnique should generate different key when first is taken")
	}

	// Should still be valid format
	if !keyPattern.MatchString(unique) {
		t.Errorf("MakePublicKeyUnique() = %q, does not match pattern", unique)
	}
}

func TestMakePublicKeyUnique_ErrorAfterMaxAttempts(t *testing.T) {
	// Checker always returns true (key always "taken")
	checker := func(key string) bool {
		return true
	}

	_, err := MakePublicKeyUnique(checker)
	if err == nil {
		t.Error("MakePublicKeyUnique() expected error after 100 attempts, got nil")
	}
	if !errors.Is(err, ErrKeyGenerationFailed) {
		t.Errorf("MakePublicKeyUnique() error = %v, want ErrKeyGenerationFailed", err)
	}
}
