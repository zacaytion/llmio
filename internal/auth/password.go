// Package auth provides authentication utilities including password hashing and session management.
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters following OWASP recommendations.
// See: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
const (
	argonTime    = 3     // Number of iterations.
	argonMemory  = 65536 // 64 MB.
	argonThreads = 4     // Parallelism factor.
	argonKeyLen  = 32    // Length of the derived key.
	argonSaltLen = 16    // Length of the salt.
)

var (
	// ErrEmptyPassword is returned when an empty password is provided.
	ErrEmptyPassword = errors.New("password cannot be empty")
)

// HashPassword creates an Argon2id hash of the password.
// Returns the hash in the format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	// Generate a random salt
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate the hash
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	// Encode to base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads, b64Salt, b64Hash)

	return encoded, nil
}

// VerifyPassword checks if the password matches the hash.
// Returns true if the password is correct, false otherwise.
// Uses constant-time comparison to prevent timing attacks.
func VerifyPassword(password, encodedHash string) bool {
	if password == "" {
		return false
	}

	// Parse the encoded hash
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false
	}

	if parts[1] != "argon2id" {
		return false
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false
	}

	var memory, time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	// Compute the hash of the provided password.
	// G115: The len() of a decoded base64 hash is bounded (typically 32 bytes), so overflow is not possible.
	computedHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(expectedHash))) //nolint:gosec

	// Constant-time comparison
	return subtle.ConstantTimeCompare(computedHash, expectedHash) == 1
}
