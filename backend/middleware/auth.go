package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/utils"
	"github.com/redis/go-redis/v9"
)

type contextKey string

const (
	UserIDKey       contextKey = "user_id"
	IsSuperAdminKey contextKey = "is_super_admin"
)

// UserRepository defines the interface for user data operations needed by auth middleware
type UserRepository interface {
	FindByID(id uint) (*models.User, error)
}

// AuthMiddleware handles JWT token validation and user context injection
type AuthMiddleware struct {
	jwtSecret string
	redis     *redis.Client
	userRepo  UserRepository
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(jwtSecret string, rdb *redis.Client, userRepo UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
		redis:     rdb,
		userRepo:  userRepo,
	}
}

// Authenticate is the middleware handler that validates JWT and injects user context
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract Bearer token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.Error(w, http.StatusUnauthorized, "Missing authorization header", "UNAUTHORIZED")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.Error(w, http.StatusUnauthorized, "Invalid authorization header format", "UNAUTHORIZED")
			return
		}

		token := parts[1]

		// Validate token signature and expiry
		claims, err := utils.ValidateToken(token, m.jwtSecret)
		if err != nil {
			utils.Error(w, http.StatusUnauthorized, "Invalid or expired token", "INVALID_TOKEN")
			return
		}

		// Check if token is blacklisted in Redis
		ctx := context.Background()
		blacklisted := m.redis.Exists(ctx, "blacklist:"+claims.ID).Val()
		if blacklisted > 0 {
			utils.Error(w, http.StatusUnauthorized, "Token has been revoked", "TOKEN_REVOKED")
			return
		}

		// Load user from database (verify user still exists and is active)
		user, err := m.userRepo.FindByID(claims.UserID)
		if err != nil {
			utils.Error(w, http.StatusUnauthorized, "User not found", "USER_NOT_FOUND")
			return
		}

		// Check user status
		if user.Status != "active" {
			utils.Error(w, http.StatusUnauthorized, "Account has been deactivated", "ACCOUNT_INACTIVE")
			return
		}

		// Inject user context into request
		ctx = context.WithValue(r.Context(), UserIDKey, user.ID)
		ctx = context.WithValue(ctx, IsSuperAdminKey, user.IsSuperAdmin)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID extracts user ID from request context
func GetUserID(ctx context.Context) uint {
	if userID, ok := ctx.Value(UserIDKey).(uint); ok {
		return userID
	}
	return 0
}

// GetIsSuperAdmin extracts super admin flag from request context
func GetIsSuperAdmin(ctx context.Context) bool {
	if isSuperAdmin, ok := ctx.Value(IsSuperAdminKey).(bool); ok {
		return isSuperAdmin
	}
	return false
}
