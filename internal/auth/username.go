package auth

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var (
	// nonAlphanumeric matches anything that's not a letter or number.
	nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)
	// multiDash matches multiple consecutive dashes.
	multiDash = regexp.MustCompile(`-+`)
)

// GenerateUsername creates a URL-safe username from a display name.
// The result matches the pattern ^[a-z0-9][a-z0-9-]*[a-z0-9]$ with min length 2.
func GenerateUsername(name string) string {
	// Normalize unicode and convert to lowercase
	name = strings.ToLower(norm.NFKD.String(name))

	// Remove accents by keeping only ASCII
	var ascii strings.Builder
	for _, r := range name {
		if r < 128 {
			ascii.WriteRune(r)
		}
	}
	slug := ascii.String()

	// Replace non-alphanumeric with dashes
	slug = nonAlphanumeric.ReplaceAllString(slug, "-")

	// Collapse multiple dashes
	slug = multiDash.ReplaceAllString(slug, "-")

	// Trim leading/trailing dashes
	slug = strings.Trim(slug, "-")

	// If slug is too short or empty, generate a random one
	if len(slug) < 2 {
		slug = randomSlug(8)
	}

	// Ensure it starts and ends with alphanumeric
	slug = ensureAlphanumericEnds(slug)

	return slug
}

// MakeUsernameUnique ensures the username is unique by appending a suffix if needed.
// The exists function should return true if the username is already taken.
func MakeUsernameUnique(base string, exists func(string) bool) string {
	if !exists(base) {
		return base
	}

	// Try with numeric suffixes
	for i := 1; i < 100; i++ {
		candidate := base + randomSlug(4)
		if !exists(candidate) {
			return candidate
		}
	}

	// Fallback to fully random
	return randomSlug(12)
}

// randomSlug generates a random alphanumeric string of the given length.
func randomSlug(length int) string {
	bytes := make([]byte, (length+1)/2)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a fixed string if random fails (should never happen)
		return strings.Repeat("x", length)
	}
	return hex.EncodeToString(bytes)[:length]
}

// ensureAlphanumericEnds makes sure the string starts and ends with alphanumeric.
func ensureAlphanumericEnds(s string) string {
	if len(s) == 0 {
		return randomSlug(8)
	}

	runes := []rune(s)

	// Trim non-alphanumeric from start
	start := 0
	for start < len(runes) && !isAlphanumeric(runes[start]) {
		start++
	}

	// Trim non-alphanumeric from end
	end := len(runes)
	for end > start && !isAlphanumeric(runes[end-1]) {
		end--
	}

	if end-start < 2 {
		return randomSlug(8)
	}

	return string(runes[start:end])
}

func isAlphanumeric(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}
