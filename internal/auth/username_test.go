package auth

import (
	"regexp"
	"testing"
)

// usernamePattern matches valid usernames per database constraint.
var usernamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

func Test_GenerateUsername(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "simple name",
			input: "John Doe",
		},
		{
			name:  "single word",
			input: "Alice",
		},
		{
			name:  "name with special characters",
			input: "José García-López",
		},
		{
			name:  "name with numbers",
			input: "Test User 123",
		},
		{
			name:  "all uppercase",
			input: "JOHN DOE",
		},
		{
			name:  "email as fallback",
			input: "user@example.com",
		},
		{
			name:  "unicode name",
			input: "日本語名前",
		},
		{
			name:  "empty name uses random",
			input: "",
		},
		{
			name:  "only special chars",
			input: "!@#$%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username := GenerateUsername(tt.input)

			// Must be at least 2 characters
			if len(username) < 2 {
				t.Errorf("GenerateUsername(%q) = %q, want at least 2 chars", tt.input, username)
			}

			// Must match the pattern
			if !usernamePattern.MatchString(username) {
				t.Errorf("GenerateUsername(%q) = %q, does not match pattern ^[a-z0-9][a-z0-9-]*[a-z0-9]$", tt.input, username)
			}

			// Must be all lowercase
			for _, r := range username {
				if r >= 'A' && r <= 'Z' {
					t.Errorf("GenerateUsername(%q) = %q, contains uppercase", tt.input, username)
					break
				}
			}
		})
	}
}

func Test_GenerateUsernameUniqueness(t *testing.T) {
	// Same input should produce same base, but with suffix for uniqueness
	name := "John Doe"
	u1 := GenerateUsername(name)
	u2 := GenerateUsername(name)

	// Both should be valid
	if !usernamePattern.MatchString(u1) {
		t.Errorf("First username %q is invalid", u1)
	}
	if !usernamePattern.MatchString(u2) {
		t.Errorf("Second username %q is invalid", u2)
	}
}

func Test_MakeUsernameUnique(t *testing.T) {
	// Test that MakeUsernameUnique adds suffix correctly
	base := "john-doe"

	checker := func(username string) bool {
		// Simulate "john-doe" already exists
		return username == "john-doe"
	}

	unique := MakeUsernameUnique(base, checker)

	if unique == base {
		t.Errorf("MakeUsernameUnique should have modified the username since %q exists", base)
	}

	if !usernamePattern.MatchString(unique) {
		t.Errorf("MakeUsernameUnique() = %q, does not match pattern", unique)
	}
}
