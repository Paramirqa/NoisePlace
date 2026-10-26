package auth_test

import (
	"strings"
	"testing"

	"music-platform/internal/auth"
)

func TestHashPassword(t *testing.T) {
	password := "correct-horse-battery-staple"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() returned an error: %v", err)
	}

	if hash == "" {
		t.Fatal("HashPassword() returned an empty hash")
	}

	if !strings.HasPrefix(hash, "$argon2id$v=19$") {
		t.Fatalf("unexpected hash format: %q", hash)
	}

	if hash == password {
		t.Fatal("password must not be stored as plaintext")
	}

	valid, err := auth.VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword() returned an error: %v", err)
	}

	if !valid {
		t.Fatal("VerifyPassword() rejected the correct password")
	}
}

func TestHashPasswordUsesDifferentSalts(t *testing.T) {
	password := "correct-horse-battery-staple"

	firstHash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("first HashPassword() returned an error: %v", err)
	}

	secondHash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("second HashPassword() returned an error: %v", err)
	}

	if firstHash == secondHash {
		t.Fatal("hashes of the same password should differ because of random salts")
	}
}

func TestVerifyPasswordRejectsWrongPassword(t *testing.T) {
	hash, err := auth.HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("HashPassword() returned an error: %v", err)
	}

	valid, err := auth.VerifyPassword("wrong-password", hash)
	if err != nil {
		t.Fatalf("VerifyPassword() returned an error: %v", err)
	}

	if valid {
		t.Fatal("VerifyPassword() accepted an incorrect password")
	}
}

func TestVerifyPasswordRejectsMalformedHash(t *testing.T) {
	tests := []struct {
		name string
		hash string
	}{
		{
			name: "empty hash",
			hash: "",
		},
		{
			name: "invalid format",
			hash: "not-a-password-hash",
		},
		{
			name: "missing parameters",
			hash: "$argon2id$v=19$",
		},
		{
			name: "unsupported algorithm",
			hash: "$bcrypt$v=19$m=65536,t=3,p=2$c2FsdA$aGFzaA",
		},
		{
			name: "parameters out of range",
			hash: "$argon2id$v=19$m=999999999,t=3,p=2$c2FsdA$aGFzaA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := auth.VerifyPassword("some-password", tt.hash)

			if err == nil {
				t.Fatal("VerifyPassword() should return an error for malformed hashes")
			}

			if valid {
				t.Fatal("VerifyPassword() accepted a malformed hash")
			}
		})
	}
}
