package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pointofsale/backend/config"
	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/repositories"
	"github.com/pointofsale/backend/utils"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo *repositories.UserRepository
	rdb      *redis.Client
	cfg      *config.Config
}

func NewAuthService(userRepo *repositories.UserRepository, rdb *redis.Client, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		rdb:      rdb,
		cfg:      cfg,
	}
}

func (s *AuthService) Register(name, email, password, role string) (*models.User, error) {
	// Check if user already exists
	existing, _ := s.userRepo.FindByEmail(email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
		Role:     role,
		IsActive: true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *AuthService) Login(email, password string) (*utils.TokenPair, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	valid, err := utils.VerifyPassword(user.Password, password)
	if err != nil || !valid {
		return nil, errors.New("invalid email or password")
	}

	tokens, err := utils.GenerateTokenPair(
		user.ID, user.Role,
		s.cfg.JWTAccessSecret, s.cfg.JWTRefreshSecret,
		s.cfg.JWTAccessExpiry, s.cfg.JWTRefreshExpiry,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Store refresh token in Redis with TTL
	ctx := context.Background()
	refreshClaims, _ := utils.ValidateToken(tokens.RefreshToken, s.cfg.JWTRefreshSecret)
	if refreshClaims != nil {
		s.rdb.Set(ctx, "refresh:"+refreshClaims.ID, fmt.Sprintf("%d", user.ID), s.cfg.JWTRefreshExpiry)
	}

	return tokens, nil
}

func (s *AuthService) Refresh(refreshToken string) (*utils.TokenPair, error) {
	claims, err := utils.ValidateToken(refreshToken, s.cfg.JWTRefreshSecret)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Check if refresh token exists in Redis (not blacklisted)
	ctx := context.Background()
	exists := s.rdb.Exists(ctx, "refresh:"+claims.ID).Val()
	if exists == 0 {
		return nil, errors.New("refresh token has been revoked")
	}

	// Check if token is blacklisted
	blacklisted := s.rdb.Exists(ctx, "blacklist:"+claims.ID).Val()
	if blacklisted > 0 {
		return nil, errors.New("refresh token has been revoked")
	}

	// Get user to ensure they still exist and are active
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Delete old refresh token from Redis
	s.rdb.Del(ctx, "refresh:"+claims.ID)

	// Generate new token pair
	tokens, err := utils.GenerateTokenPair(
		user.ID, user.Role,
		s.cfg.JWTAccessSecret, s.cfg.JWTRefreshSecret,
		s.cfg.JWTAccessExpiry, s.cfg.JWTRefreshExpiry,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Store new refresh token
	newRefreshClaims, _ := utils.ValidateToken(tokens.RefreshToken, s.cfg.JWTRefreshSecret)
	if newRefreshClaims != nil {
		s.rdb.Set(ctx, "refresh:"+newRefreshClaims.ID, fmt.Sprintf("%d", user.ID), s.cfg.JWTRefreshExpiry)
	}

	return tokens, nil
}

func (s *AuthService) Logout(accessToken, refreshToken string) error {
	ctx := context.Background()

	// Blacklist access token
	if accessToken != "" {
		accessClaims, err := utils.ValidateToken(accessToken, s.cfg.JWTAccessSecret)
		if err == nil && accessClaims != nil {
			ttl := time.Until(accessClaims.ExpiresAt.Time)
			if ttl > 0 {
				s.rdb.Set(ctx, "blacklist:"+accessClaims.ID, "1", ttl)
			}
		}
	}

	// Blacklist refresh token and remove from active tokens
	if refreshToken != "" {
		refreshClaims, err := utils.ValidateToken(refreshToken, s.cfg.JWTRefreshSecret)
		if err == nil && refreshClaims != nil {
			s.rdb.Del(ctx, "refresh:"+refreshClaims.ID)
			ttl := time.Until(refreshClaims.ExpiresAt.Time)
			if ttl > 0 {
				s.rdb.Set(ctx, "blacklist:"+refreshClaims.ID, "1", ttl)
			}
		}
	}

	return nil
}
