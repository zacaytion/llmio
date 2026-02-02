package auth

import (
	"strings"
	"testing"
	"time"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "secretpassword123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
		},
		{
			name:     "unicode password",
			password: "пароль密码🔐",
			wantErr:  false,
		},
		{
			name:     "very long password",
			password: strings.Repeat("a", 1000),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Hash should start with argon2id identifier
				if !strings.HasPrefix(hash, "$argon2id$") {
					t.Errorf("HashPassword() hash should start with $argon2id$, got %s", hash[:20])
				}
			}
		})
	}
}

func TestHashPasswordUniqueness(t *testing.T) {
	password := "samepassword"
	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash1 == hash2 {
		t.Error("HashPassword() should produce different hashes for same password (unique salt)")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "secretpassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			want:     true,
		},
		{
			name:     "wrong password",
			password: "wrongpassword",
			hash:     hash,
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			want:     false,
		},
		{
			name:     "invalid hash format",
			password: password,
			hash:     "notahash",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VerifyPassword(tt.password, tt.hash)
			if got != tt.want {
				t.Errorf("VerifyPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifyPasswordTiming(t *testing.T) {
	// Verify that verification takes consistent time regardless of result
	// This prevents timing attacks for account enumeration
	password := "secretpassword123"
	hash, _ := HashPassword(password)

	// Measure time for correct password
	start := time.Now()
	for i := 0; i < 10; i++ {
		VerifyPassword(password, hash)
	}
	correctTime := time.Since(start)

	// Measure time for wrong password
	start = time.Now()
	for i := 0; i < 10; i++ {
		VerifyPassword("wrongpassword", hash)
	}
	wrongTime := time.Since(start)

	// Times should be within 50% of each other
	// (argon2 inherently provides constant-time comparison)
	ratio := float64(correctTime) / float64(wrongTime)
	if ratio < 0.5 || ratio > 2.0 {
		t.Errorf("VerifyPassword timing ratio = %v, want between 0.5 and 2.0", ratio)
	}
}
