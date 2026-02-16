package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pointofsale/backend/config"
	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/utils"
	"github.com/redis/go-redis/v9"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	Update(user *models.User) error
	FindByIDWithPermissions(id uint) (*models.User, []models.RolePermission, error)
}

// EmailService defines the interface for email operations
type EmailService interface {
	SendWelcomeEmail(toEmail, userName string) error
	SendPasswordResetEmail(toEmail, userName, resetLink string) error
	SendAccountApprovedEmail(toEmail, userName string) error
}

// Input DTOs
type RegisterInput struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ResetPasswordInput struct {
	Token           string `json:"token"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

// Output DTOs
type TokenPair struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type LoginResponse struct {
	User *models.User `json:"user"`
	TokenPair
}

type PermissionDTO struct {
	Module  string   `json:"module"`
	Feature string   `json:"feature"`
	Actions []string `json:"actions"`
}

type CurrentUserResponse struct {
	*models.User
	Permissions []PermissionDTO `json:"permissions"`
}

// Custom Errors
var (
	ErrValidation   = errors.New("validation error")
	ErrConflict     = errors.New("conflict error")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
)

type ServiceError struct {
	Err     error
	Message string
	Code    string
}

func (e *ServiceError) Error() string {
	return e.Message
}

// AuthService handles authentication business logic
type AuthService struct {
	userRepo     UserRepository
	redis        *redis.Client
	config       *config.Config
	emailService EmailService
}

func NewAuthService(userRepo UserRepository, rdb *redis.Client, cfg *config.Config, emailSvc EmailService) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		redis:        rdb,
		config:       cfg,
		emailService: emailSvc,
	}
}

// Register creates a new user account with pending status
func (s *AuthService) Register(input RegisterInput) (*models.User, *ServiceError) {
	// Validate name
	if err := utils.ValidateRequired(input.Name, "Name"); err != "" {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: err,
			Code:    "VALIDATION_ERROR",
		}
	}
	if len(input.Name) < 2 || len(input.Name) > 255 {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Name must be between 2 and 255 characters",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Validate email
	if err := utils.ValidateRequired(input.Email, "Email"); err != "" {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: err,
			Code:    "VALIDATION_ERROR",
		}
	}
	if !utils.ValidateEmail(input.Email) {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Invalid email format",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Validate password
	if err := utils.ValidateRequired(input.Password, "Password"); err != "" {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: err,
			Code:    "VALIDATION_ERROR",
		}
	}
	if passwordErrors := utils.ValidatePassword(input.Password); len(passwordErrors) > 0 {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: strings.Join(passwordErrors, "; "),
			Code:    "VALIDATION_ERROR",
		}
	}

	// Validate password confirmation
	if input.Password != input.ConfirmPassword {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Passwords do not match",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Check email uniqueness (case-insensitive)
	normalizedEmail := strings.ToLower(input.Email)
	existing, _ := s.userRepo.FindByEmail(normalizedEmail)
	if existing != nil {
		return nil, &ServiceError{
			Err:     ErrConflict,
			Message: "Email already registered",
			Code:    "EMAIL_EXISTS",
		}
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to process password",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Create user
	user := &models.User{
		Name:         input.Name,
		Email:        normalizedEmail,
		PasswordHash: hashedPassword,
		Status:       "pending",
		IsSuperAdmin: false,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to create user",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Send welcome email (non-blocking, don't fail if email fails)
	if s.emailService != nil {
		_ = s.emailService.SendWelcomeEmail(user.Email, user.Name)
	}

	return user, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(input LoginInput) (*LoginResponse, *ServiceError) {
	// Validate input
	if err := utils.ValidateRequired(input.Email, "Email"); err != "" {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: err,
			Code:    "VALIDATION_ERROR",
		}
	}
	if err := utils.ValidateRequired(input.Password, "Password"); err != "" {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: err,
			Code:    "VALIDATION_ERROR",
		}
	}

	// Find user (case-insensitive email)
	normalizedEmail := strings.ToLower(input.Email)
	user, err := s.userRepo.FindByEmail(normalizedEmail)
	if err != nil {
		return nil, &ServiceError{
			Err:     ErrUnauthorized,
			Message: "Invalid email or password",
			Code:    "INVALID_CREDENTIALS",
		}
	}

	// Verify password
	valid, err := utils.VerifyPassword(user.PasswordHash, input.Password)
	if err != nil || !valid {
		return nil, &ServiceError{
			Err:     ErrUnauthorized,
			Message: "Invalid email or password",
			Code:    "INVALID_CREDENTIALS",
		}
	}

	// Check user status
	if user.Status == "pending" {
		return nil, &ServiceError{
			Err:     ErrForbidden,
			Message: "Account is pending approval",
			Code:    "ACCOUNT_PENDING",
		}
	}
	if user.Status == "inactive" {
		return nil, &ServiceError{
			Err:     ErrForbidden,
			Message: "Account has been deactivated",
			Code:    "ACCOUNT_INACTIVE",
		}
	}

	// Generate tokens
	accessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.IsSuperAdmin,
		s.config.JWTAccessSecret,
		s.config.JWTAccessExpiry,
	)
	if err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to generate access token",
			Code:    "INTERNAL_ERROR",
		}
	}

	refreshToken, err := utils.GenerateRefreshToken(
		user.ID,
		user.IsSuperAdmin,
		s.config.JWTRefreshSecret,
		s.config.JWTRefreshExpiry,
	)
	if err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to generate refresh token",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Store refresh token in Redis
	refreshClaims, err := utils.ValidateToken(refreshToken, s.config.JWTRefreshSecret)
	if err == nil && refreshClaims != nil {
		ctx := context.Background()
		s.redis.Set(ctx, "refresh:"+refreshClaims.ID, fmt.Sprintf("%d", user.ID), s.config.JWTRefreshExpiry)
	}

	// Get expiry time from access token
	accessClaims, _ := utils.ValidateToken(accessToken, s.config.JWTAccessSecret)
	expiresAt := time.Now().Add(s.config.JWTAccessExpiry)
	if accessClaims != nil {
		expiresAt = accessClaims.ExpiresAt.Time
	}

	return &LoginResponse{
		User: user,
		TokenPair: TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    expiresAt,
		},
	}, nil
}

// RefreshToken generates a new token pair from a valid refresh token
func (s *AuthService) RefreshToken(refreshToken string) (*TokenPair, *ServiceError) {
	// Validate refresh token
	claims, err := utils.ValidateToken(refreshToken, s.config.JWTRefreshSecret)
	if err != nil {
		return nil, &ServiceError{
			Err:     ErrUnauthorized,
			Message: "Invalid refresh token",
			Code:    "INVALID_TOKEN",
		}
	}

	// Check if refresh token exists in Redis (not revoked)
	ctx := context.Background()
	exists := s.redis.Exists(ctx, "refresh:"+claims.ID).Val()
	if exists == 0 {
		return nil, &ServiceError{
			Err:     ErrUnauthorized,
			Message: "Refresh token has been revoked",
			Code:    "TOKEN_REVOKED",
		}
	}

	// Get user to ensure they still exist and are active
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, &ServiceError{
			Err:     ErrUnauthorized,
			Message: "User not found",
			Code:    "USER_NOT_FOUND",
		}
	}

	if user.Status != "active" {
		return nil, &ServiceError{
			Err:     ErrForbidden,
			Message: "Account has been deactivated",
			Code:    "ACCOUNT_INACTIVE",
		}
	}

	// Delete old refresh token
	s.redis.Del(ctx, "refresh:"+claims.ID)

	// Generate new token pair
	newAccessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.IsSuperAdmin,
		s.config.JWTAccessSecret,
		s.config.JWTAccessExpiry,
	)
	if err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to generate access token",
			Code:    "INTERNAL_ERROR",
		}
	}

	newRefreshToken, err := utils.GenerateRefreshToken(
		user.ID,
		user.IsSuperAdmin,
		s.config.JWTRefreshSecret,
		s.config.JWTRefreshExpiry,
	)
	if err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to generate refresh token",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Store new refresh token
	newRefreshClaims, err := utils.ValidateToken(newRefreshToken, s.config.JWTRefreshSecret)
	if err == nil && newRefreshClaims != nil {
		s.redis.Set(ctx, "refresh:"+newRefreshClaims.ID, fmt.Sprintf("%d", user.ID), s.config.JWTRefreshExpiry)
	}

	// Get expiry time
	accessClaims, _ := utils.ValidateToken(newAccessToken, s.config.JWTAccessSecret)
	expiresAt := time.Now().Add(s.config.JWTAccessExpiry)
	if accessClaims != nil {
		expiresAt = accessClaims.ExpiresAt.Time
	}

	return &TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// Logout invalidates both access and refresh tokens
func (s *AuthService) Logout(accessToken, refreshToken string) *ServiceError {
	ctx := context.Background()

	// Blacklist access token
	if accessToken != "" {
		accessClaims, err := utils.ValidateToken(accessToken, s.config.JWTAccessSecret)
		if err == nil && accessClaims != nil {
			ttl := time.Until(accessClaims.ExpiresAt.Time)
			if ttl > 0 {
				s.redis.Set(ctx, "blacklist:"+accessClaims.ID, "1", ttl)
			}
		}
	}

	// Delete refresh token and blacklist it
	if refreshToken != "" {
		refreshClaims, err := utils.ValidateToken(refreshToken, s.config.JWTRefreshSecret)
		if err == nil && refreshClaims != nil {
			s.redis.Del(ctx, "refresh:"+refreshClaims.ID)
			ttl := time.Until(refreshClaims.ExpiresAt.Time)
			if ttl > 0 {
				s.redis.Set(ctx, "blacklist:"+refreshClaims.ID, "1", ttl)
			}
		}
	}

	return nil
}

// ForgotPassword initiates the password reset process
func (s *AuthService) ForgotPassword(email string) *ServiceError {
	// Find user (case-insensitive)
	normalizedEmail := strings.ToLower(email)
	user, err := s.userRepo.FindByEmail(normalizedEmail)

	// Only send reset email if user exists and is active
	// Always return success to avoid revealing if email exists
	if err == nil && user != nil && user.Status == "active" {
		// Generate reset token
		resetToken, err := utils.GenerateResetToken()
		if err != nil {
			// Log error but still return success
			return nil
		}

		// Store in Redis with 1 hour TTL
		ctx := context.Background()
		s.redis.Set(ctx, "reset:"+resetToken, fmt.Sprintf("%d", user.ID), time.Hour)

		// Send password reset email
		if s.emailService != nil {
			resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.config.FrontendURL, resetToken)
			_ = s.emailService.SendPasswordResetEmail(user.Email, user.Name, resetLink)
		}
	}

	return nil
}

// ResetPassword changes user password using a reset token
func (s *AuthService) ResetPassword(input ResetPasswordInput) *ServiceError {
	// Validate password
	if passwordErrors := utils.ValidatePassword(input.Password); len(passwordErrors) > 0 {
		return &ServiceError{
			Err:     ErrValidation,
			Message: strings.Join(passwordErrors, "; "),
			Code:    "VALIDATION_ERROR",
		}
	}

	// Validate password confirmation
	if input.Password != input.ConfirmPassword {
		return &ServiceError{
			Err:     ErrValidation,
			Message: "Passwords do not match",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Check reset token in Redis
	ctx := context.Background()
	userIDStr, err := s.redis.Get(ctx, "reset:"+input.Token).Result()
	if err != nil {
		return &ServiceError{
			Err:     ErrUnauthorized,
			Message: "Invalid or expired reset token",
			Code:    "INVALID_TOKEN",
		}
	}

	// Parse user ID
	var userID uint
	_, err = fmt.Sscanf(userIDStr, "%d", &userID)
	if err != nil {
		return &ServiceError{
			Err:     ErrUnauthorized,
			Message: "Invalid reset token",
			Code:    "INVALID_TOKEN",
		}
	}

	// Get user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return &ServiceError{
			Err:     ErrNotFound,
			Message: "User not found",
			Code:    "USER_NOT_FOUND",
		}
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return &ServiceError{
			Err:     err,
			Message: "Failed to process password",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Update user password
	user.PasswordHash = hashedPassword
	if err := s.userRepo.Update(user); err != nil {
		return &ServiceError{
			Err:     err,
			Message: "Failed to update password",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Delete reset token
	s.redis.Del(ctx, "reset:"+input.Token)

	// Invalidate all refresh tokens for this user
	iter := s.redis.Scan(ctx, 0, "refresh:*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		val, err := s.redis.Get(ctx, key).Result()
		if err == nil && val == userIDStr {
			s.redis.Del(ctx, key)
		}
	}

	return nil
}

// GetCurrentUser returns user details with permissions
func (s *AuthService) GetCurrentUser(userID uint) (*CurrentUserResponse, *ServiceError) {
	user, rolePerms, err := s.userRepo.FindByIDWithPermissions(userID)
	if err != nil {
		return nil, &ServiceError{
			Err:     ErrNotFound,
			Message: "User not found",
			Code:    "USER_NOT_FOUND",
		}
	}

	var permissions []PermissionDTO

	if user.IsSuperAdmin {
		// Super admin gets all permissions
		permissions = getAllPermissions()
	} else {
		// Build permissions from role_permissions
		permMap := make(map[string]map[string][]string) // module -> feature -> actions

		for _, rp := range rolePerms {
			module := rp.Permission.Module
			feature := rp.Permission.Feature

			if permMap[module] == nil {
				permMap[module] = make(map[string][]string)
			}

			permMap[module][feature] = rp.Actions
		}

		// Convert map to DTO slice
		for module, features := range permMap {
			for feature, actions := range features {
				permissions = append(permissions, PermissionDTO{
					Module:  module,
					Feature: feature,
					Actions: actions,
				})
			}
		}
	}

	return &CurrentUserResponse{
		User:        user,
		Permissions: permissions,
	}, nil
}

// getAllPermissions returns all available permissions for super admin
func getAllPermissions() []PermissionDTO {
	return []PermissionDTO{
		{Module: "Master Data", Feature: "Product", Actions: []string{"read", "create", "update", "delete", "export"}},
		{Module: "Master Data", Feature: "Category", Actions: []string{"read", "create", "update", "delete"}},
		{Module: "Master Data", Feature: "Supplier", Actions: []string{"read", "create", "update", "delete", "export"}},
		{Module: "Master Data", Feature: "Rack", Actions: []string{"read", "create", "update", "delete"}},
		{Module: "Transaction", Feature: "Sales", Actions: []string{"read", "create", "update", "delete", "export"}},
		{Module: "Transaction", Feature: "Purchase", Actions: []string{"read", "create", "update", "delete", "export"}},
		{Module: "Report", Feature: "Sales Report", Actions: []string{"read", "export"}},
		{Module: "Report", Feature: "Purchase Report", Actions: []string{"read", "export"}},
		{Module: "Settings", Feature: "Users", Actions: []string{"read", "create", "update", "delete"}},
		{Module: "Settings", Feature: "Roles & Permissions", Actions: []string{"read", "create", "update", "delete"}},
	}
}
