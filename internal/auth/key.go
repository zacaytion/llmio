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
// Returns an error after 100 failed attempts (indicates a bug in exists function).
func MakePublicKeyUnique(exists func(string) bool) (string, error) {
	for range 100 {
		key := GeneratePublicKey()
		if !exists(key) {
			return key, nil
		}
	}

	// 100 collisions with 128-bit keys is virtually impossible
	// This likely indicates the exists function is broken (always returns true)
	return "", ErrKeyGenerationFailed
}

// ErrKeyGenerationFailed indicates key generation failed after max attempts.
var ErrKeyGenerationFailed = errKeyGenerationFailed{}

type errKeyGenerationFailed struct{}

func (errKeyGenerationFailed) Error() string {
	return "failed to generate unique public key after 100 attempts"
}
