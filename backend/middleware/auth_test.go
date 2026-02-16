package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/utils"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock user repository for middleware tests
type mockUserRepo struct {
	findByIDFn func(id uint) (*models.User, error)
}

func (m *mockUserRepo) FindByID(id uint) (*models.User, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(id)
	}
	return nil, nil
}

// setupTestMiddleware creates a test auth middleware with miniredis
func setupTestMiddleware(t *testing.T, userRepo UserRepository) (*AuthMiddleware, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	middleware := NewAuthMiddleware("test-secret", rdb, userRepo)
	return middleware, mr
}

// testHandler is a simple handler that returns the user ID from context
func testHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r.Context())
		isSuperAdmin := GetIsSuperAdmin(r.Context())
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("authenticated"))
		// Store in header for test assertions
		w.Header().Set("X-User-ID", string(rune(userID)))
		if isSuperAdmin {
			w.Header().Set("X-Super-Admin", "true")
		}
	})
}

func TestAuthMiddleware_ValidToken_SetsUserContext(t *testing.T) {
	// Setup mock user repo
	mockRepo := &mockUserRepo{
		findByIDFn: func(id uint) (*models.User, error) {
			return &models.User{
				ID:           1,
				Email:        "test@example.com",
				Status:       "active",
				IsSuperAdmin: false,
			}, nil
		},
	}

	authMiddleware, _ := setupTestMiddleware(t, mockRepo)

	// Generate valid token
	token, err := utils.GenerateAccessToken(1, false, "test-secret", 15*time.Minute)
	require.NoError(t, err)

	// Create request with valid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute middleware
	handler := authMiddleware.Authenticate(testHandler())
	handler.ServeHTTP(rr, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "authenticated", rr.Body.String())
}

func TestAuthMiddleware_NoAuthHeader_Returns401(t *testing.T) {
	mockRepo := &mockUserRepo{}
	authMiddleware, _ := setupTestMiddleware(t, mockRepo)

	// Create request without Authorization header
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	// Execute middleware
	handler := authMiddleware.Authenticate(testHandler())
	handler.ServeHTTP(rr, req)

	// Assert 401 response
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Missing authorization header")
}

func TestAuthMiddleware_InvalidAuthHeaderFormat_Returns401(t *testing.T) {
	mockRepo := &mockUserRepo{}
	authMiddleware, _ := setupTestMiddleware(t, mockRepo)

	testCases := []struct {
		name   string
		header string
	}{
		{"missing Bearer prefix", "some-token"},
		{"wrong prefix", "Basic some-token"},
		{"no token after Bearer", "Bearer "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tc.header)
			rr := httptest.NewRecorder()

			handler := authMiddleware.Authenticate(testHandler())
			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusUnauthorized, rr.Code)
		})
	}
}

func TestAuthMiddleware_InvalidToken_Returns401(t *testing.T) {
	mockRepo := &mockUserRepo{}
	authMiddleware, _ := setupTestMiddleware(t, mockRepo)

	// Create request with invalid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	// Execute middleware
	handler := authMiddleware.Authenticate(testHandler())
	handler.ServeHTTP(rr, req)

	// Assert 401 response
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid or expired token")
}

func TestAuthMiddleware_ExpiredToken_Returns401(t *testing.T) {
	mockRepo := &mockUserRepo{}
	authMiddleware, _ := setupTestMiddleware(t, mockRepo)

	// Generate expired token (negative duration)
	token, err := utils.GenerateAccessToken(1, false, "test-secret", -1*time.Hour)
	require.NoError(t, err)

	// Create request with expired token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	// Execute middleware
	handler := authMiddleware.Authenticate(testHandler())
	handler.ServeHTTP(rr, req)

	// Assert 401 response
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid or expired token")
}

func TestAuthMiddleware_BlacklistedToken_Returns401(t *testing.T) {
	mockRepo := &mockUserRepo{}
	authMiddleware, mr := setupTestMiddleware(t, mockRepo)

	// Generate valid token
	token, err := utils.GenerateAccessToken(1, false, "test-secret", 15*time.Minute)
	require.NoError(t, err)

	// Validate token to get claims
	claims, err := utils.ValidateToken(token, "test-secret")
	require.NoError(t, err)

	// Blacklist the token in Redis
	mr.Set("blacklist:"+claims.ID, "1")

	// Create request with blacklisted token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	// Execute middleware
	handler := authMiddleware.Authenticate(testHandler())
	handler.ServeHTTP(rr, req)

	// Assert 401 response
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Token has been revoked")
}

func TestAuthMiddleware_UserNotFound_Returns401(t *testing.T) {
	// Mock repo that returns error when finding user
	mockRepo := &mockUserRepo{
		findByIDFn: func(id uint) (*models.User, error) {
			return nil, assert.AnError
		},
	}
	authMiddleware, _ := setupTestMiddleware(t, mockRepo)

	// Generate valid token
	token, err := utils.GenerateAccessToken(1, false, "test-secret", 15*time.Minute)
	require.NoError(t, err)

	// Create request with valid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	// Execute middleware
	handler := authMiddleware.Authenticate(testHandler())
	handler.ServeHTTP(rr, req)

	// Assert 401 response
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "User not found")
}

func TestAuthMiddleware_InactiveUser_Returns401(t *testing.T) {
	// Mock repo that returns inactive user
	mockRepo := &mockUserRepo{
		findByIDFn: func(id uint) (*models.User, error) {
			return &models.User{
				ID:           1,
				Email:        "test@example.com",
				Status:       "inactive",
				IsSuperAdmin: false,
			}, nil
		},
	}
	authMiddleware, _ := setupTestMiddleware(t, mockRepo)

	// Generate valid token
	token, err := utils.GenerateAccessToken(1, false, "test-secret", 15*time.Minute)
	require.NoError(t, err)

	// Create request with valid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	// Execute middleware
	handler := authMiddleware.Authenticate(testHandler())
	handler.ServeHTTP(rr, req)

	// Assert 401 response
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Account has been deactivated")
}

func TestAuthMiddleware_SuperAdmin_SetsContextCorrectly(t *testing.T) {
	// Mock repo that returns super admin user
	mockRepo := &mockUserRepo{
		findByIDFn: func(id uint) (*models.User, error) {
			return &models.User{
				ID:           1,
				Email:        "admin@example.com",
				Status:       "active",
				IsSuperAdmin: true,
			}, nil
		},
	}
	authMiddleware, _ := setupTestMiddleware(t, mockRepo)

	// Generate valid super admin token
	token, err := utils.GenerateAccessToken(1, true, "test-secret", 15*time.Minute)
	require.NoError(t, err)

	// Create request with valid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	// Execute middleware with custom handler to check context
	handler := authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r.Context())
		isSuperAdmin := GetIsSuperAdmin(r.Context())

		assert.Equal(t, uint(1), userID)
		assert.True(t, isSuperAdmin)

		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(rr, req)

	// Assert success
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetUserID_NoContext_ReturnsZero(t *testing.T) {
	ctx := context.Background()
	userID := GetUserID(ctx)
	assert.Equal(t, uint(0), userID)
}

func TestGetIsSuperAdmin_NoContext_ReturnsFalse(t *testing.T) {
	ctx := context.Background()
	isSuperAdmin := GetIsSuperAdmin(ctx)
	assert.False(t, isSuperAdmin)
}
