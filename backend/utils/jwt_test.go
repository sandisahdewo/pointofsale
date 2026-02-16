package utils

import (
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAccessToken_ValidUser_ReturnsToken(t *testing.T) {
	userID := uint(123)
	isSuperAdmin := false
	secret := "test-access-secret"
	expiry := 15 * time.Minute

	token, err := GenerateAccessToken(userID, isSuperAdmin, secret, expiry)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected token to be non-empty")
	}

	// Verify token can be parsed and contains correct claims
	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("generated token should be valid, got error: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("expected UserID %d, got %d", userID, claims.UserID)
	}
	if claims.IsSuperAdmin != isSuperAdmin {
		t.Errorf("expected IsSuperAdmin %v, got %v", isSuperAdmin, claims.IsSuperAdmin)
	}
}

func TestGenerateRefreshToken_ValidUser_ReturnsToken(t *testing.T) {
	userID := uint(456)
	isSuperAdmin := true
	secret := "test-refresh-secret"
	expiry := 7 * 24 * time.Hour

	token, err := GenerateRefreshToken(userID, isSuperAdmin, secret, expiry)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected token to be non-empty")
	}

	// Verify token can be parsed and contains correct claims
	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("generated token should be valid, got error: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("expected UserID %d, got %d", userID, claims.UserID)
	}
	if claims.IsSuperAdmin != isSuperAdmin {
		t.Errorf("expected IsSuperAdmin %v, got %v", isSuperAdmin, claims.IsSuperAdmin)
	}
}

func TestValidateToken_ValidToken_ReturnsClaims(t *testing.T) {
	userID := uint(789)
	isSuperAdmin := false
	secret := "test-secret"
	expiry := 1 * time.Hour

	token, err := GenerateAccessToken(userID, isSuperAdmin, secret, expiry)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := ValidateToken(token, secret)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if claims == nil {
		t.Fatal("expected claims to be non-nil")
	}
	if claims.UserID != userID {
		t.Errorf("expected UserID %d, got %d", userID, claims.UserID)
	}
	if claims.IsSuperAdmin != isSuperAdmin {
		t.Errorf("expected IsSuperAdmin %v, got %v", isSuperAdmin, claims.IsSuperAdmin)
	}
	if claims.ID == "" {
		t.Error("expected jti (ID) to be set")
	}
}

func TestValidateToken_ExpiredToken_ReturnsError(t *testing.T) {
	userID := uint(100)
	isSuperAdmin := false
	secret := "test-secret"
	expiry := -1 * time.Hour // Already expired

	token, err := GenerateAccessToken(userID, isSuperAdmin, secret, expiry)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = ValidateToken(token, secret)

	if err == nil {
		t.Fatal("expected error for expired token, got none")
	}
	// Check that error is related to expiration
	if !strings.Contains(err.Error(), "expired") && !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected expiration-related error, got: %v", err)
	}
}

func TestValidateToken_InvalidSignature_ReturnsError(t *testing.T) {
	userID := uint(200)
	isSuperAdmin := false
	correctSecret := "correct-secret"
	wrongSecret := "wrong-secret"
	expiry := 1 * time.Hour

	token, err := GenerateAccessToken(userID, isSuperAdmin, correctSecret, expiry)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = ValidateToken(token, wrongSecret)

	if err == nil {
		t.Fatal("expected error for invalid signature, got none")
	}
}

func TestValidateToken_MalformedToken_ReturnsError(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "random string",
			token: "not-a-jwt-token",
		},
		{
			name:  "incomplete token",
			token: "header.payload",
		},
		{
			name:  "invalid base64",
			token: "invalid!!!.payload!!!.signature!!!",
		},
	}

	secret := "test-secret"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateToken(tt.token, secret)
			if err == nil {
				t.Error("expected error for malformed token, got none")
			}
		})
	}
}

func TestGenerateResetToken_ReturnsHexString(t *testing.T) {
	token, err := GenerateResetToken()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected token to be non-empty")
	}

	// Verify it's a hex string
	_, err = hex.DecodeString(token)
	if err != nil {
		t.Errorf("expected token to be valid hex, got error: %v", err)
	}

	// Verify length (32 bytes = 64 hex characters)
	if len(token) != 64 {
		t.Errorf("expected token length 64, got %d", len(token))
	}
}

func TestGenerateResetToken_UniquenessCheck(t *testing.T) {
	// Generate multiple tokens and verify they're unique
	tokens := make(map[string]bool)
	for i := 0; i < 10; i++ {
		token, err := GenerateResetToken()
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}
		if tokens[token] {
			t.Errorf("generated duplicate token: %s", token)
		}
		tokens[token] = true
	}
}

func TestClaims_StructureMatches(t *testing.T) {
	// Verify the Claims struct has the expected fields
	claims := Claims{
		UserID:       123,
		IsSuperAdmin: true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        "test-jti",
		},
	}

	if claims.UserID != 123 {
		t.Errorf("expected UserID 123, got %d", claims.UserID)
	}
	if !claims.IsSuperAdmin {
		t.Error("expected IsSuperAdmin to be true")
	}
	if claims.ID != "test-jti" {
		t.Errorf("expected ID 'test-jti', got %s", claims.ID)
	}
}
