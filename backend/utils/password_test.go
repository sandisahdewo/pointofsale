package utils

import (
	"strings"
	"testing"
)

func TestHashPassword_ValidPassword_ReturnsHash(t *testing.T) {
	password := "SecurePassword123!"
	hash, err := HashPassword(password)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hash == "" {
		t.Fatal("expected hash to be non-empty")
	}
	// Verify hash format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	if !strings.HasPrefix(hash, "$argon2id$v=19$m=65536,t=3,p=4$") {
		t.Errorf("hash format incorrect, got: %s", hash)
	}
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		t.Errorf("expected 6 parts in hash, got %d", len(parts))
	}
}

func TestHashPassword_DifferentSalts_DifferentHashes(t *testing.T) {
	password := "SecurePassword123!"
	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	if err1 != nil || err2 != nil {
		t.Fatalf("expected no errors, got %v, %v", err1, err2)
	}
	if hash1 == hash2 {
		t.Error("expected different hashes due to different salts, but got identical hashes")
	}
}

func TestVerifyPassword_CorrectPassword_ReturnsTrue(t *testing.T) {
	password := "SecurePassword123!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	match, err := VerifyPassword(hash, password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !match {
		t.Error("expected password to match, but it didn't")
	}
}

func TestVerifyPassword_WrongPassword_ReturnsFalse(t *testing.T) {
	password := "SecurePassword123!"
	wrongPassword := "WrongPassword456!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	match, err := VerifyPassword(hash, wrongPassword)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if match {
		t.Error("expected password to not match, but it did")
	}
}

func TestVerifyPassword_InvalidHash_ReturnsError(t *testing.T) {
	tests := []struct {
		name        string
		invalidHash string
	}{
		{
			name:        "malformed hash",
			invalidHash: "not-a-valid-hash",
		},
		{
			name:        "too few parts",
			invalidHash: "$argon2id$v=19$m=65536",
		},
		{
			name:        "invalid base64 salt",
			invalidHash: "$argon2id$v=19$m=65536,t=3,p=4$!!!invalid!!!$validhash",
		},
		{
			name:        "invalid parameters",
			invalidHash: "$argon2id$v=19$invalidparams$salt$hash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := VerifyPassword(tt.invalidHash, "password")
			if err == nil {
				t.Error("expected error for invalid hash, but got none")
			}
		})
	}
}
