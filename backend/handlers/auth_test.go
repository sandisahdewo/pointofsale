package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-chi/chi/v5"
	"github.com/pointofsale/backend/config"
	"github.com/pointofsale/backend/middleware"
	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/repositories"
	"github.com/pointofsale/backend/services"
	"github.com/pointofsale/backend/testutil"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupTestRouter creates a fully configured test router with all dependencies
func setupTestRouter(t *testing.T) (chi.Router, *gorm.DB, *redis.Client) {
	t.Helper()

	// Setup test database
	db := testutil.SetupTestDB(t)

	// Setup miniredis
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create test config
	cfg := &config.Config{
		FrontendURL:      "http://localhost:3000",
		JWTAccessSecret:  "test-access-secret",
		JWTRefreshSecret: "test-refresh-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}

	// Initialize layers
	userRepo := repositories.NewUserRepository(db)
	authService := services.NewAuthService(userRepo, rdb, cfg, nil) // nil email service
	authHandler := NewAuthHandler(authService)
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTAccessSecret, rdb, userRepo)
	healthHandler := NewHealthHandler(db, rdb)

	// Setup router (inline to avoid import cycle with routes package)
	r := chi.NewRouter()
	r.Get("/health", healthHandler.Health)
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/forgot-password", authHandler.ForgotPassword)
			r.Post("/reset-password", authHandler.ResetPassword)
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.Authenticate)
				r.Post("/logout", authHandler.Logout)
				r.Get("/me", authHandler.GetMe)
			})
		})
	})

	return r, db, rdb
}

func TestRegisterHandler_ValidBody_Returns201(t *testing.T) {
	router, db, _ := setupTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	body := `{
		"name": "John Doe",
		"email": "john@example.com",
		"password": "Password@123",
		"confirmPassword": "Password@123"
	}`

	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")
	assert.Contains(t, response, "message")

	userData := response["data"].(map[string]interface{})
	assert.Equal(t, "john@example.com", userData["email"])
	assert.Equal(t, "John Doe", userData["name"])
	assert.Equal(t, "pending", userData["status"])
}

func TestRegisterHandler_MissingFields_Returns400(t *testing.T) {
	router, db, _ := setupTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	testCases := []struct {
		name string
		body string
	}{
		{"missing name", `{"email": "test@example.com", "password": "Password@123", "confirmPassword": "Password@123"}`},
		{"missing email", `{"name": "Test", "password": "Password@123", "confirmPassword": "Password@123"}`},
		{"missing password", `{"name": "Test", "email": "test@example.com", "confirmPassword": "Password@123"}`},
		{"empty name", `{"name": "", "email": "test@example.com", "password": "Password@123", "confirmPassword": "Password@123"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
			assert.Contains(t, rr.Body.String(), "error")
		})
	}
}

func TestRegisterHandler_DuplicateEmail_Returns409(t *testing.T) {
	router, db, _ := setupTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	// Create existing user
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Email = "existing@example.com"
	})

	body := `{
		"name": "New User",
		"email": "existing@example.com",
		"password": "Password@123",
		"confirmPassword": "Password@123"
	}`

	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Contains(t, rr.Body.String(), "Email already registered")
}

func TestLoginHandler_ValidCredentials_Returns200WithTokens(t *testing.T) {
	router, db, _ := setupTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	// Create active user
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Email = "active@example.com"
		u.Status = "active"
	})

	body := `{
		"email": "active@example.com",
		"password": "Password@123"
	}`

	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Contains(t, data, "user")
	assert.Contains(t, data, "accessToken")
	assert.Contains(t, data, "refreshToken")
	assert.Contains(t, data, "expiresAt")
}

func TestLoginHandler_InvalidCredentials_Returns401(t *testing.T) {
	router, db, _ := setupTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	// Create user
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Email = "user@example.com"
		u.Status = "active"
	})

	body := `{
		"email": "user@example.com",
		"password": "WrongPassword@123"
	}`

	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid email or password")
}

func TestLoginHandler_PendingUser_Returns403(t *testing.T) {
	router, db, _ := setupTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	// Create pending user
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Email = "pending@example.com"
		u.Status = "pending"
	})

	body := `{
		"email": "pending@example.com",
		"password": "Password@123"
	}`

	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "Account is pending approval")
}

func TestRefreshHandler_ValidToken_Returns200(t *testing.T) {
	router, db, rdb := setupTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	// Create active user and login to get refresh token
	user := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Email = "refresh@example.com"
		u.Status = "active"
	})

	// Generate refresh token
	cfg := &config.Config{
		JWTRefreshSecret: "test-refresh-secret",
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}
	refreshToken, err := testutil.GenerateTestRefreshToken(t, user.ID, false, cfg.JWTRefreshSecret, cfg.JWTRefreshExpiry)
	require.NoError(t, err)

	// Store refresh token in Redis
	claims, _ := testutil.ValidateToken(refreshToken, cfg.JWTRefreshSecret)
	rdb.Set(testutil.Context(), "refresh:"+claims.ID, user.ID, cfg.JWTRefreshExpiry)

	body := `{"refreshToken": "` + refreshToken + `"}`

	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Contains(t, data, "accessToken")
	assert.Contains(t, data, "refreshToken")
}

func TestRefreshHandler_InvalidToken_Returns401(t *testing.T) {
	router, db, _ := setupTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	body := `{"refreshToken": "invalid-token"}`

	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestLogoutHandler_Authenticated_Returns200(t *testing.T) {
	router, db, _ := setupTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	// Create active user
	user := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Status = "active"
	})

	// Generate access token
	accessToken := testutil.GenerateTestAccessToken(t, user.ID, false)

	body := `{"refreshToken": "some-refresh-token"}`

	req := httptest.NewRequest("POST", "/api/v1/auth/logout", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Logged out successfully")
}

func TestLogoutHandler_NoAuth_Returns401(t *testing.T) {
	router, db, _ := setupTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	body := `{"refreshToken": "some-refresh-token"}`

	req := httptest.NewRequest("POST", "/api/v1/auth/logout", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestForgotPasswordHandler_ValidEmail_Returns200(t *testing.T) {
	router, db, _ := setupTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	// Create active user
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Email = "forgot@example.com"
		u.Status = "active"
	})

	body := `{"email": "forgot@example.com"}`

	req := httptest.NewRequest("POST", "/api/v1/auth/forgot-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Always returns 200 to avoid email enumeration
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "If the email exists")
}

func TestResetPasswordHandler_ValidToken_Returns200(t *testing.T) {
	router, db, rdb := setupTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	// Create user
	user := testutil.CreateTestUser(t, db)

	// Generate reset token and store in Redis
	resetToken := "test-reset-token-123"
	rdb.Set(testutil.Context(), "reset:"+resetToken, user.ID, time.Hour)

	body := `{
		"token": "test-reset-token-123",
		"password": "NewPassword@456",
		"confirmPassword": "NewPassword@456"
	}`

	req := httptest.NewRequest("POST", "/api/v1/auth/reset-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Password reset successfully")
}

func TestMeHandler_Authenticated_ReturnsUserData(t *testing.T) {
	router, db, _ := setupTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	// Create active user
	user := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Status = "active"
	})

	// Generate access token
	accessToken := testutil.GenerateTestAccessToken(t, user.ID, false)

	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, user.Email, data["email"])
	assert.Contains(t, data, "permissions")
}

func TestMeHandler_NoAuth_Returns401(t *testing.T) {
	router, db, _ := setupTestRouter(t)
	defer testutil.CleanupTestDB(t, db)

	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	// No Authorization header
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
