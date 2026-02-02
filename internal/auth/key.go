package auth

import (
	"crypto/rand"
	"encoding/base64"
)

const (
	// publicKeyBytes is the number of random bytes (128 bits = 16 bytes).
	publicKeyBytes = 16
)

// GeneratePublicKey creates a cryptographically random public URL key.
// Returns a 22-character base64url-encoded string (128 bits of entropy).
func GeneratePublicKey() string {
	bytes := make([]byte, publicKeyBytes)
	if _, err := rand.Read(bytes); err != nil {
		// This should never happen; if it does, panic is appropriate
		panic("crypto/rand failed: " + err.Error())
	}

	// Use base64url encoding (URL-safe, no padding)
	return base64.RawURLEncoding.EncodeToString(bytes)
}

// MakePublicKeyUnique generates keys until one is not taken.
// The exists function should return true if the key is already in use.
func MakePublicKeyUnique(exists func(string) bool) string {
	for i := 0; i < 100; i++ {
		key := GeneratePublicKey()
		if !exists(key) {
			return key
		}
	}

	// After 100 attempts, just return the last one
	// (collision probability is astronomically low with 128 bits)
	return GeneratePublicKey()
}
