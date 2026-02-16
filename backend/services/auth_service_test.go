package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/pointofsale/backend/config"
	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/utils"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock UserRepository
type mockUserRepo struct {
	createFn            func(*models.User) error
	findByEmailFn       func(string) (*models.User, error)
	findByIDFn          func(uint) (*models.User, error)
	updateFn            func(*models.User) error
	findByIDWithPermsFn func(uint) (*models.User, []models.RolePermission, error)
}

func (m *mockUserRepo) Create(user *models.User) error {
	if m.createFn != nil {
		return m.createFn(user)
	}
	return nil
}

func (m *mockUserRepo) FindByEmail(email string) (*models.User, error) {
	if m.findByEmailFn != nil {
		return m.findByEmailFn(email)
	}
	return nil, errors.New("not found")
}

func (m *mockUserRepo) FindByID(id uint) (*models.User, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(id)
	}
	return nil, errors.New("not found")
}

func (m *mockUserRepo) Update(user *models.User) error {
	if m.updateFn != nil {
		return m.updateFn(user)
	}
	return nil
}

func (m *mockUserRepo) FindByIDWithPermissions(id uint) (*models.User, []models.RolePermission, error) {
	if m.findByIDWithPermsFn != nil {
		return m.findByIDWithPermsFn(id)
	}
	return nil, nil, errors.New("not found")
}

// Mock EmailService
type mockEmailService struct {
	sendWelcomeFn        func(string, string) error
	sendPasswordResetFn  func(string, string, string) error
	sendAccountApprovedFn func(string, string) error
}

func (m *mockEmailService) SendWelcomeEmail(toEmail, userName string) error {
	if m.sendWelcomeFn != nil {
		return m.sendWelcomeFn(toEmail, userName)
	}
	return nil
}

func (m *mockEmailService) SendPasswordResetEmail(toEmail, userName, resetLink string) error {
	if m.sendPasswordResetFn != nil {
		return m.sendPasswordResetFn(toEmail, userName, resetLink)
	}
	return nil
}

func (m *mockEmailService) SendAccountApprovedEmail(toEmail, userName string) error {
	if m.sendAccountApprovedFn != nil {
		return m.sendAccountApprovedFn(toEmail, userName)
	}
	return nil
}

// Test setup helper
func setupAuthServiceTest(t *testing.T) (*AuthService, *mockUserRepo, *redis.Client, *miniredis.Miniredis, *config.Config) {
	// Create miniredis instance
	mr, err := miniredis.Run()
	require.NoError(t, err)

	// Create real Redis client connected to miniredis
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Test config
	cfg := &config.Config{
		JWTAccessSecret:  "test-access-secret",
		JWTRefreshSecret: "test-refresh-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
		FrontendURL:      "http://localhost:3000",
	}

	mockRepo := &mockUserRepo{}
	mockEmail := &mockEmailService{}

	service := NewAuthService(mockRepo, rdb, cfg, mockEmail)

	return service, mockRepo, rdb, mr, cfg
}

func TestRegister_ValidInput_CreatesUserWithPendingStatus(t *testing.T) {
	service, mockRepo, _, mr, _ := setupAuthServiceTest(t)
	defer mr.Close()

	mockRepo.findByEmailFn = func(email string) (*models.User, error) {
		return nil, errors.New("not found")
	}

	var createdUser *models.User
	mockRepo.createFn = func(user *models.User) error {
		createdUser = user
		user.ID = 1
		return nil
	}

	input := RegisterInput{
		Name:            "John Doe",
		Email:           "john@example.com",
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	}

	user, svcErr := service.Register(input)

	assert.Nil(t, svcErr)
	assert.NotNil(t, user)
	assert.Equal(t, "John Doe", createdUser.Name)
	assert.Equal(t, "john@example.com", createdUser.Email)
	assert.Equal(t, "pending", createdUser.Status)
	assert.False(t, createdUser.IsSuperAdmin)
	assert.NotEmpty(t, createdUser.PasswordHash)
}

func TestRegister_DuplicateEmail_ReturnsConflictError(t *testing.T) {
	service, mockRepo, _, mr, _ := setupAuthServiceTest(t)
	defer mr.Close()

	mockRepo.findByEmailFn = func(email string) (*models.User, error) {
		return &models.User{
			ID:    1,
			Email: email,
		}, nil
	}

	input := RegisterInput{
		Name:            "John Doe",
		Email:           "john@example.com",
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	}

	user, err := service.Register(input)

	assert.Nil(t, user)
	assert.NotNil(t, err)
	assert.Equal(t, ErrConflict, err.Err)
	assert.Contains(t, err.Message, "already registered")
}

func TestRegister_PasswordMismatch_ReturnsValidationError(t *testing.T) {
	service, _, _, mr, _ := setupAuthServiceTest(t)
	defer mr.Close()

	input := RegisterInput{
		Name:            "John Doe",
		Email:           "john@example.com",
		Password:        "Password123!",
		ConfirmPassword: "DifferentPassword123!",
	}

	user, err := service.Register(input)

	assert.Nil(t, user)
	assert.NotNil(t, err)
	assert.Equal(t, ErrValidation, err.Err)
	assert.Contains(t, err.Message, "Passwords do not match")
}

func TestRegister_WeakPassword_ReturnsValidationError(t *testing.T) {
	service, _, _, mr, _ := setupAuthServiceTest(t)
	defer mr.Close()

	input := RegisterInput{
		Name:            "John Doe",
		Email:           "john@example.com",
		Password:        "weak",
		ConfirmPassword: "weak",
	}

	user, err := service.Register(input)

	assert.Nil(t, user)
	assert.NotNil(t, err)
	assert.Equal(t, ErrValidation, err.Err)
	assert.Contains(t, err.Message, "at least 8 characters")
}

func TestRegister_EmptyName_ReturnsValidationError(t *testing.T) {
	service, _, _, mr, _ := setupAuthServiceTest(t)
	defer mr.Close()

	input := RegisterInput{
		Name:            "",
		Email:           "john@example.com",
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	}

	user, err := service.Register(input)

	assert.Nil(t, user)
	assert.NotNil(t, err)
	assert.Equal(t, ErrValidation, err.Err)
	assert.Contains(t, err.Message, "Name is required")
}

func TestRegister_InvalidEmail_ReturnsValidationError(t *testing.T) {
	service, _, _, mr, _ := setupAuthServiceTest(t)
	defer mr.Close()

	input := RegisterInput{
		Name:            "John Doe",
		Email:           "invalid-email",
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	}

	user, err := service.Register(input)

	assert.Nil(t, user)
	assert.NotNil(t, err)
	assert.Equal(t, ErrValidation, err.Err)
	assert.Contains(t, err.Message, "Invalid email format")
}

func TestRegister_CaseInsensitiveEmailCheck(t *testing.T) {
	service, mockRepo, _, mr, _ := setupAuthServiceTest(t)
	defer mr.Close()

	mockRepo.findByEmailFn = func(email string) (*models.User, error) {
		// Should be called with lowercase email
		assert.Equal(t, "john@example.com", email)
		return &models.User{ID: 1, Email: "JOHN@example.com"}, nil
	}

	input := RegisterInput{
		Name:            "John Doe",
		Email:           "JOHN@EXAMPLE.COM",
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	}

	user, err := service.Register(input)

	assert.Nil(t, user)
	assert.NotNil(t, err)
	assert.Equal(t, ErrConflict, err.Err)
}

func TestLogin_ActiveUser_ReturnsTokens(t *testing.T) {
	service, mockRepo, rdb, mr, cfg := setupAuthServiceTest(t)
	defer mr.Close()

	hashedPassword, _ := utils.HashPassword("Password123!")

	mockRepo.findByEmailFn = func(email string) (*models.User, error) {
		return &models.User{
			ID:           1,
			Email:        email,
			Name:         "John Doe",
			PasswordHash: hashedPassword,
			Status:       "active",
			IsSuperAdmin: false,
		}, nil
	}

	input := LoginInput{
		Email:    "john@example.com",
		Password: "Password123!",
	}

	response, svcErr := service.Login(input)

	assert.Nil(t, svcErr)
	assert.NotNil(t, response)
	assert.NotNil(t, response.User)
	assert.Equal(t, uint(1), response.User.ID)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.False(t, response.ExpiresAt.IsZero())

	// Verify refresh token is stored in Redis
	refreshClaims, _ := utils.ValidateToken(response.RefreshToken, cfg.JWTRefreshSecret)
	val, redisErr := rdb.Get(context.Background(), "refresh:"+refreshClaims.ID).Result()
	assert.NoError(t, redisErr)
	assert.Equal(t, "1", val)
}

func TestLogin_PendingUser_ReturnsForbiddenError(t *testing.T) {
	service, mockRepo, _, mr, _ := setupAuthServiceTest(t)
	defer mr.Close()

	hashedPassword, _ := utils.HashPassword("Password123!")

	mockRepo.findByEmailFn = func(email string) (*models.User, error) {
		return &models.User{
			ID:           1,
			Email:        email,
			PasswordHash: hashedPassword,
			Status:       "pending",
		}, nil
	}

	input := LoginInput{
		Email:    "john@example.com",
		Password: "Password123!",
	}

	response, err := service.Login(input)

	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.Equal(t, ErrForbidden, err.Err)
	assert.Contains(t, err.Message, "pending approval")
}

func TestLogin_InactiveUser_ReturnsForbiddenError(t *testing.T) {
	service, mockRepo, _, mr, _ := setupAuthServiceTest(t)
	defer mr.Close()

	hashedPassword, _ := utils.HashPassword("Password123!")

	mockRepo.findByEmailFn = func(email string) (*models.User, error) {
		return &models.User{
			ID:           1,
			Email:        email,
			PasswordHash: hashedPassword,
			Status:       "inactive",
		}, nil
	}

	input := LoginInput{
		Email:    "john@example.com",
		Password: "Password123!",
	}

	response, err := service.Login(input)

	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.Equal(t, ErrForbidden, err.Err)
	assert.Contains(t, err.Message, "deactivated")
}

func TestLogin_WrongPassword_ReturnsUnauthorizedError(t *testing.T) {
	service, mockRepo, _, mr, _ := setupAuthServiceTest(t)
	defer mr.Close()

	hashedPassword, _ := utils.HashPassword("Password123!")

	mockRepo.findByEmailFn = func(email string) (*models.User, error) {
		return &models.User{
			ID:           1,
			Email:        email,
			PasswordHash: hashedPassword,
			Status:       "active",
		}, nil
	}

	input := LoginInput{
		Email:    "john@example.com",
		Password: "WrongPassword123!",
	}

	response, err := service.Login(input)

	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.Equal(t, ErrUnauthorized, err.Err)
	assert.Contains(t, err.Message, "Invalid email or password")
}

func TestLogin_NonExistentEmail_ReturnsUnauthorizedError(t *testing.T) {
	service, mockRepo, _, mr, _ := setupAuthServiceTest(t)
	defer mr.Close()

	mockRepo.findByEmailFn = func(email string) (*models.User, error) {
		return nil, errors.New("not found")
	}

	input := LoginInput{
		Email:    "nonexistent@example.com",
		Password: "Password123!",
	}

	response, err := service.Login(input)

	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.Equal(t, ErrUnauthorized, err.Err)
	assert.Contains(t, err.Message, "Invalid email or password")
}

func TestRefreshToken_ValidToken_ReturnsNewPair(t *testing.T) {
	service, mockRepo, rdb, mr, cfg := setupAuthServiceTest(t)
	defer mr.Close()

	// Generate initial refresh token
	refreshToken, _ := utils.GenerateRefreshToken(1, false, cfg.JWTRefreshSecret, cfg.JWTRefreshExpiry)
	refreshClaims, _ := utils.ValidateToken(refreshToken, cfg.JWTRefreshSecret)

	// Store refresh token in Redis
	ctx := context.Background()
	rdb.Set(ctx, "refresh:"+refreshClaims.ID, "1", cfg.JWTRefreshExpiry)

	mockRepo.findByIDFn = func(id uint) (*models.User, error) {
		return &models.User{
			ID:     1,
			Email:  "john@example.com",
			Status: "active",
		}, nil
	}

	newTokens, svcErr := service.RefreshToken(refreshToken)

	assert.Nil(t, svcErr)
	assert.NotNil(t, newTokens)
	assert.NotEmpty(t, newTokens.AccessToken)
	assert.NotEmpty(t, newTokens.RefreshToken)

	// Verify old refresh token is deleted
	_, redisErr := rdb.Get(ctx, "refresh:"+refreshClaims.ID).Result()
	assert.Error(t, redisErr) // Should be deleted

	// Verify new refresh token is stored
	newRefreshClaims, _ := utils.ValidateToken(newTokens.RefreshToken, cfg.JWTRefreshSecret)
	val, redisErr := rdb.Get(ctx, "refresh:"+newRefreshClaims.ID).Result()
	assert.NoError(t, redisErr)
	assert.Equal(t, "1", val)
}

func TestRefreshToken_RevokedToken_ReturnsError(t *testing.T) {
	service, _, _, mr, cfg := setupAuthServiceTest(t)
	defer mr.Close()

	// Generate refresh token but don't store in Redis (simulating revoked token)
	refreshToken, _ := utils.GenerateRefreshToken(1, false, cfg.JWTRefreshSecret, cfg.JWTRefreshExpiry)

	newTokens, err := service.RefreshToken(refreshToken)

	assert.Nil(t, newTokens)
	assert.NotNil(t, err)
	assert.Equal(t, ErrUnauthorized, err.Err)
	assert.Contains(t, err.Message, "revoked")
}

func TestRefreshToken_InvalidToken_ReturnsError(t *testing.T) {
	service, _, _, mr, _ := setupAuthServiceTest(t)
	defer mr.Close()

	newTokens, err := service.RefreshToken("invalid-token")

	assert.Nil(t, newTokens)
	assert.NotNil(t, err)
	assert.Equal(t, ErrUnauthorized, err.Err)
}

func TestLogout_ValidTokens_BlacklistsBoth(t *testing.T) {
	service, _, rdb, mr, cfg := setupAuthServiceTest(t)
	defer mr.Close()

	// Generate tokens
	accessToken, _ := utils.GenerateAccessToken(1, false, cfg.JWTAccessSecret, cfg.JWTAccessExpiry)
	refreshToken, _ := utils.GenerateRefreshToken(1, false, cfg.JWTRefreshSecret, cfg.JWTRefreshExpiry)

	accessClaims, _ := utils.ValidateToken(accessToken, cfg.JWTAccessSecret)
	refreshClaims, _ := utils.ValidateToken(refreshToken, cfg.JWTRefreshSecret)

	// Store refresh token
	ctx := context.Background()
	rdb.Set(ctx, "refresh:"+refreshClaims.ID, "1", cfg.JWTRefreshExpiry)

	svcErr := service.Logout(accessToken, refreshToken)

	assert.Nil(t, svcErr)

	// Verify access token is blacklisted
	val, redisErr := rdb.Get(ctx, "blacklist:"+accessClaims.ID).Result()
	assert.NoError(t, redisErr)
	assert.Equal(t, "1", val)

	// Verify refresh token is deleted
	_, redisErr = rdb.Get(ctx, "refresh:"+refreshClaims.ID).Result()
	assert.Error(t, redisErr) // Should be deleted
}

func TestForgotPassword_ExistingEmail_StoresTokenInRedis(t *testing.T) {
	service, mockRepo, rdb, mr, _ := setupAuthServiceTest(t)
	defer mr.Close()

	mockRepo.findByEmailFn = func(email string) (*models.User, error) {
		return &models.User{
			ID:     1,
			Email:  email,
			Name:   "John Doe",
			Status: "active",
		}, nil
	}

	svcErr := service.ForgotPassword("john@example.com")

	assert.Nil(t, svcErr)

	// Check if reset token was stored in Redis
	ctx := context.Background()
	keys := rdb.Keys(ctx, "reset:*").Val()
	assert.Len(t, keys, 1)

	// Verify token value is user ID
	val, redisErr := rdb.Get(ctx, keys[0]).Result()
	assert.NoError(t, redisErr)
	assert.Equal(t, "1", val)
}

func TestForgotPassword_NonExistingEmail_StillReturnsSuccess(t *testing.T) {
	service, mockRepo, rdb, mr, _ := setupAuthServiceTest(t)
	defer mr.Close()

	mockRepo.findByEmailFn = func(email string) (*models.User, error) {
		return nil, errors.New("not found")
	}

	svcErr := service.ForgotPassword("nonexistent@example.com")

	// Should still return success (don't reveal if email exists)
	assert.Nil(t, svcErr)

	// Verify no token was stored
	ctx := context.Background()
	keys := rdb.Keys(ctx, "reset:*").Val()
	assert.Len(t, keys, 0)
}

func TestResetPassword_ValidToken_UpdatesPassword(t *testing.T) {
	service, mockRepo, rdb, mr, _ := setupAuthServiceTest(t)
	defer mr.Close()

	// Store reset token in Redis
	ctx := context.Background()
	resetToken := "test-reset-token-123"
	rdb.Set(ctx, "reset:"+resetToken, "1", time.Hour)

	var updatedUser *models.User
	mockRepo.findByIDFn = func(id uint) (*models.User, error) {
		return &models.User{
			ID:     1,
			Email:  "john@example.com",
			Status: "active",
		}, nil
	}

	mockRepo.updateFn = func(user *models.User) error {
		updatedUser = user
		return nil
	}

	input := ResetPasswordInput{
		Token:           resetToken,
		Password:        "NewPassword123!",
		ConfirmPassword: "NewPassword123!",
	}

	svcErr := service.ResetPassword(input)

	assert.Nil(t, svcErr)
	assert.NotNil(t, updatedUser)
	assert.NotEmpty(t, updatedUser.PasswordHash)

	// Verify password was hashed correctly
	valid, _ := utils.VerifyPassword(updatedUser.PasswordHash, "NewPassword123!")
	assert.True(t, valid)

	// Verify reset token was deleted
	_, redisErr := rdb.Get(ctx, "reset:"+resetToken).Result()
	assert.Error(t, redisErr)
}

func TestResetPassword_ExpiredToken_ReturnsError(t *testing.T) {
	service, _, _, mr, _ := setupAuthServiceTest(t)
	defer mr.Close()

	input := ResetPasswordInput{
		Token:           "expired-token",
		Password:        "NewPassword123!",
		ConfirmPassword: "NewPassword123!",
	}

	err := service.ResetPassword(input)

	assert.NotNil(t, err)
	assert.Equal(t, ErrUnauthorized, err.Err)
	assert.Contains(t, err.Message, "Invalid or expired")
}

func TestResetPassword_PasswordMismatch_ReturnsError(t *testing.T) {
	service, _, _, mr, _ := setupAuthServiceTest(t)
	defer mr.Close()

	input := ResetPasswordInput{
		Token:           "token",
		Password:        "NewPassword123!",
		ConfirmPassword: "DifferentPassword123!",
	}

	err := service.ResetPassword(input)

	assert.NotNil(t, err)
	assert.Equal(t, ErrValidation, err.Err)
	assert.Contains(t, err.Message, "do not match")
}

func TestResetPassword_InvalidatesAllRefreshTokens(t *testing.T) {
	service, mockRepo, rdb, mr, cfg := setupAuthServiceTest(t)
	defer mr.Close()

	// Store reset token
	ctx := context.Background()
	resetToken := "test-reset-token-123"
	rdb.Set(ctx, "reset:"+resetToken, "1", time.Hour)

	// Store multiple refresh tokens for the user
	token1, _ := utils.GenerateRefreshToken(1, false, cfg.JWTRefreshSecret, cfg.JWTRefreshExpiry)
	token2, _ := utils.GenerateRefreshToken(1, false, cfg.JWTRefreshSecret, cfg.JWTRefreshExpiry)
	claims1, _ := utils.ValidateToken(token1, cfg.JWTRefreshSecret)
	claims2, _ := utils.ValidateToken(token2, cfg.JWTRefreshSecret)

	rdb.Set(ctx, "refresh:"+claims1.ID, "1", cfg.JWTRefreshExpiry)
	rdb.Set(ctx, "refresh:"+claims2.ID, "1", cfg.JWTRefreshExpiry)

	mockRepo.findByIDFn = func(id uint) (*models.User, error) {
		return &models.User{ID: 1}, nil
	}

	mockRepo.updateFn = func(user *models.User) error {
		return nil
	}

	input := ResetPasswordInput{
		Token:           resetToken,
		Password:        "NewPassword123!",
		ConfirmPassword: "NewPassword123!",
	}

	svcErr := service.ResetPassword(input)

	assert.Nil(t, svcErr)

	// Verify all refresh tokens were deleted
	keys := rdb.Keys(ctx, "refresh:*").Val()
	assert.Len(t, keys, 0)
}

func TestGetCurrentUser_ValidId_ReturnsUserWithPermissions(t *testing.T) {
	service, mockRepo, _, mr, _ := setupAuthServiceTest(t)
	defer mr.Close()

	mockRepo.findByIDWithPermsFn = func(id uint) (*models.User, []models.RolePermission, error) {
		user := &models.User{
			ID:           1,
			Email:        "john@example.com",
			Name:         "John Doe",
			Status:       "active",
			IsSuperAdmin: false,
		}
		rolePerms := []models.RolePermission{
			{
				Permission: models.Permission{
					Module:  "master",
					Feature: "category",
				},
				Actions: []string{"view", "create"},
			},
		}
		return user, rolePerms, nil
	}

	response, svcErr := service.GetCurrentUser(1)

	assert.Nil(t, svcErr)
	assert.NotNil(t, response)
	assert.Equal(t, uint(1), response.ID)
	assert.Len(t, response.Permissions, 1)
	assert.Equal(t, "master", response.Permissions[0].Module)
	assert.Equal(t, "category", response.Permissions[0].Feature)
	assert.Equal(t, []string{"view", "create"}, response.Permissions[0].Actions)
}

func TestGetCurrentUser_SuperAdmin_ReturnsAllPermissions(t *testing.T) {
	service, mockRepo, _, mr, _ := setupAuthServiceTest(t)
	defer mr.Close()

	mockRepo.findByIDWithPermsFn = func(id uint) (*models.User, []models.RolePermission, error) {
		user := &models.User{
			ID:           1,
			Email:        "admin@example.com",
			Name:         "Admin",
			Status:       "active",
			IsSuperAdmin: true,
		}
		return user, nil, nil
	}

	response, svcErr := service.GetCurrentUser(1)

	assert.Nil(t, svcErr)
	assert.NotNil(t, response)
	assert.True(t, response.IsSuperAdmin)
	assert.NotEmpty(t, response.Permissions)

	// Super admin should have all modules
	moduleSet := make(map[string]bool)
	for _, perm := range response.Permissions {
		moduleSet[perm.Module] = true
	}

	// Check that common modules are present
	assert.True(t, moduleSet["Master Data"])
	assert.True(t, moduleSet["Transaction"])
	assert.True(t, moduleSet["Settings"])
}

func TestGetCurrentUser_NotFound_ReturnsError(t *testing.T) {
	service, mockRepo, _, mr, _ := setupAuthServiceTest(t)
	defer mr.Close()

	mockRepo.findByIDWithPermsFn = func(id uint) (*models.User, []models.RolePermission, error) {
		return nil, nil, errors.New("not found")
	}

	response, err := service.GetCurrentUser(999)

	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.Equal(t, ErrNotFound, err.Err)
}
