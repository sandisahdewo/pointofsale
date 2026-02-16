package testutil

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pointofsale/backend/utils"
	"github.com/stretchr/testify/require"
)

// Test JWT secrets for consistent token generation
const (
	TestJWTAccessSecret  = "test-access-secret"
	TestJWTRefreshSecret = "test-refresh-secret"
)

// GenerateTestAccessToken creates a valid JWT access token for testing.
func GenerateTestAccessToken(t *testing.T, userID uint, isSuperAdmin bool) string {
	t.Helper()

	token, err := utils.GenerateAccessToken(userID, isSuperAdmin, TestJWTAccessSecret, 15*time.Minute)
	require.NoError(t, err, "failed to generate test access token")

	return token
}

// GenerateTestRefreshToken creates a valid JWT refresh token for testing.
func GenerateTestRefreshToken(t *testing.T, userID uint, isSuperAdmin bool, secret string, expiry time.Duration) (string, error) {
	t.Helper()
	return utils.GenerateRefreshToken(userID, isSuperAdmin, secret, expiry)
}

// ValidateToken validates a token for testing.
func ValidateToken(token string, secret string) (*utils.Claims, error) {
	return utils.ValidateToken(token, secret)
}

// Context returns a background context for testing.
func Context() context.Context {
	return context.Background()
}

// AuthenticatedRequest creates an HTTP request with Authorization header.
func AuthenticatedRequest(t *testing.T, method, url string, body io.Reader, token string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, url, body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	return req
}
