package api

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// handleRegex matches valid handle characters after initial slugification.
var handleRegex = regexp.MustCompile(`[^a-z0-9-]+`)

// multiHyphenRegex matches multiple consecutive hyphens.
var multiHyphenRegex = regexp.MustCompile(`-+`)

// GenerateHandle creates a URL-safe handle from a group name.
// The handle is:
//   - Lowercased
//   - Transliterated (accented chars → ASCII)
//   - Non-alphanumeric chars replaced with hyphens
//   - Multiple hyphens collapsed to single hyphen
//   - Leading/trailing hyphens removed
//   - Truncated to 100 characters max
//
// If the result is shorter than 3 characters, it returns an empty string
// (the caller should handle this by requiring an explicit handle).
func GenerateHandle(name string) string {
	if name == "" {
		return ""
	}

	// Step 1: Normalize Unicode (NFD decomposition)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, name)

	// Step 2: Lowercase
	result = strings.ToLower(result)

	// Step 3: Replace non-alphanumeric with hyphens
	result = handleRegex.ReplaceAllString(result, "-")

	// Step 4: Collapse multiple hyphens
	result = multiHyphenRegex.ReplaceAllString(result, "-")

	// Step 5: Trim leading/trailing hyphens
	result = strings.Trim(result, "-")

	// Step 6: Truncate to max length (100 chars)
	if len(result) > 100 {
		result = result[:100]
		// Re-trim if we cut mid-hyphen
		result = strings.Trim(result, "-")
	}

	// Step 7: Validate minimum length (3 chars)
	if len(result) < 3 {
		return ""
	}

	return result
}

// GenerateUniqueHandle generates a handle and appends a numeric suffix if needed
// to ensure uniqueness. The checkExists function should return true if the handle
// already exists in the database.
//
// Example:
//
//	handle := GenerateUniqueHandle("Climate Team", func(h string) bool {
//	    exists, _ := queries.HandleExists(ctx, h)
//	    return exists.Exists
//	})
//
// Returns: "climate-team", "climate-team-1", "climate-team-2", etc.
func GenerateUniqueHandle(name string, checkExists func(handle string) bool) string {
	base := GenerateHandle(name)
	if base == "" {
		return ""
	}

	// Try the base handle first
	if !checkExists(base) {
		return base
	}

	// Try numeric suffixes
	for i := 1; i <= 1000; i++ {
		candidate := base + "-" + itoa(i)
		// Ensure we don't exceed max length
		if len(candidate) > 100 {
			// Truncate base to make room for suffix
			maxBase := 100 - len("-") - len(itoa(i))
			if maxBase < 3 {
				return "" // Can't generate valid handle
			}
			candidate = base[:maxBase] + "-" + itoa(i)
		}
		if !checkExists(candidate) {
			return candidate
		}
	}

	// Extremely unlikely: 1000 collisions
	return ""
}

// itoa converts an int to string without importing strconv.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}
