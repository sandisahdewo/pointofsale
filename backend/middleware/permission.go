package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lib/pq"
	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/utils"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// PermissionMiddleware handles permission-based authorization
type PermissionMiddleware struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewPermissionMiddleware creates a new permission middleware instance
func NewPermissionMiddleware(db *gorm.DB, rdb *redis.Client) *PermissionMiddleware {
	return &PermissionMiddleware{
		db:    db,
		redis: rdb,
	}
}

// permissionCache holds cached user permissions
type permissionCache struct {
	Permissions []permissionEntry `json:"permissions"`
}

// permissionEntry represents a single permission entry in cache
type permissionEntry struct {
	Module  string   `json:"module"`
	Feature string   `json:"feature"`
	Actions []string `json:"actions"`
}

const permissionCacheTTL = 5 * time.Minute

// RequirePermission returns middleware that checks if the user has the specified permission.
// Super admins bypass all permission checks.
func (pm *PermissionMiddleware) RequirePermission(module, feature, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user ID and super admin flag from context
			userID := GetUserID(r.Context())
			if userID == 0 {
				utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
				return
			}

			isSuperAdmin := GetIsSuperAdmin(r.Context())

			// Super admins bypass all permission checks
			if isSuperAdmin {
				next.ServeHTTP(w, r)
				return
			}

			// Check if user has the required permission
			hasPermission, err := pm.checkPermission(r.Context(), userID, module, feature, action)
			if err != nil {
				// Log error but return generic forbidden message to user
				utils.Error(w, http.StatusForbidden, "You don't have permission to perform this action", "FORBIDDEN")
				return
			}

			if !hasPermission {
				utils.Error(w, http.StatusForbidden, "You don't have permission to perform this action", "FORBIDDEN")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// checkPermission checks if a user has a specific permission action
func (pm *PermissionMiddleware) checkPermission(ctx context.Context, userID uint, module, feature, action string) (bool, error) {
	// Try to get permissions from cache
	permissions, err := pm.getPermissionsFromCache(ctx, userID)
	if err == nil && permissions != nil {
		// Cache hit: check permissions
		return pm.hasPermissionInCache(permissions, module, feature, action), nil
	}

	// Cache miss: load from database
	permissions, err = pm.loadPermissionsFromDB(ctx, userID)
	if err != nil {
		return false, err
	}

	// Cache the permissions
	_ = pm.cachePermissions(ctx, userID, permissions)

	// Check if user has the permission
	return pm.hasPermissionInCache(permissions, module, feature, action), nil
}

// getPermissionsFromCache retrieves cached permissions from Redis
func (pm *PermissionMiddleware) getPermissionsFromCache(ctx context.Context, userID uint) (*permissionCache, error) {
	cacheKey := buildPermissionCacheKey(userID)
	data, err := pm.redis.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, err
	}

	var cache permissionCache
	if err := json.Unmarshal([]byte(data), &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

// loadPermissionsFromDB loads user permissions from database
func (pm *PermissionMiddleware) loadPermissionsFromDB(ctx context.Context, userID uint) (*permissionCache, error) {
	// Load user with roles
	var user models.User
	if err := pm.db.Preload("Roles").First(&user, userID).Error; err != nil {
		return nil, err
	}

	if len(user.Roles) == 0 {
		// User has no roles, therefore no permissions
		return &permissionCache{Permissions: []permissionEntry{}}, nil
	}

	// Get role IDs
	roleIDs := make([]uint, len(user.Roles))
	for i, role := range user.Roles {
		roleIDs[i] = role.ID
	}

	// Load role permissions
	var rolePermissions []models.RolePermission
	if err := pm.db.Preload("Permission").
		Where("role_id IN ?", roleIDs).
		Find(&rolePermissions).Error; err != nil {
		return nil, err
	}

	// Build permission cache
	permMap := make(map[string]*permissionEntry)
	for _, rp := range rolePermissions {
		key := fmt.Sprintf("%s:%s", rp.Permission.Module, rp.Permission.Feature)
		if existing, ok := permMap[key]; ok {
			// Merge actions if permission already exists
			existing.Actions = mergeActions(existing.Actions, rp.Actions)
		} else {
			permMap[key] = &permissionEntry{
				Module:  rp.Permission.Module,
				Feature: rp.Permission.Feature,
				Actions: rp.Actions,
			}
		}
	}

	// Convert map to slice
	permissions := make([]permissionEntry, 0, len(permMap))
	for _, perm := range permMap {
		permissions = append(permissions, *perm)
	}

	return &permissionCache{Permissions: permissions}, nil
}

// cachePermissions stores permissions in Redis
func (pm *PermissionMiddleware) cachePermissions(ctx context.Context, userID uint, cache *permissionCache) error {
	cacheKey := buildPermissionCacheKey(userID)
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	return pm.redis.Set(ctx, cacheKey, data, permissionCacheTTL).Err()
}

// hasPermissionInCache checks if a specific permission action exists in cache
func (pm *PermissionMiddleware) hasPermissionInCache(cache *permissionCache, module, feature, action string) bool {
	for _, perm := range cache.Permissions {
		if perm.Module == module && perm.Feature == feature {
			return containsAction(perm.Actions, action)
		}
	}
	return false
}

// InvalidatePermissionCache invalidates the permission cache for a user
func InvalidatePermissionCache(ctx context.Context, rdb *redis.Client, userID uint) error {
	cacheKey := buildPermissionCacheKey(userID)
	return rdb.Del(ctx, cacheKey).Err()
}

// buildPermissionCacheKey builds the Redis key for user permissions
func buildPermissionCacheKey(userID uint) string {
	return fmt.Sprintf("perms:%d", userID)
}

// mergeActions merges two action slices, removing duplicates
func mergeActions(a, b pq.StringArray) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, action := range a {
		if !seen[action] {
			seen[action] = true
			result = append(result, action)
		}
	}

	for _, action := range b {
		if !seen[action] {
			seen[action] = true
			result = append(result, action)
		}
	}

	return result
}

// containsAction checks if an action exists in a slice
func containsAction(actions []string, action string) bool {
	for _, a := range actions {
		if a == action {
			return true
		}
	}
	return false
}
