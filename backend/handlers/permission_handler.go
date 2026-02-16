package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/utils"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// PermissionHandler handles permission-related HTTP requests
type PermissionHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewPermissionHandler creates a new permission handler instance
func NewPermissionHandler(db *gorm.DB, redis *redis.Client) *PermissionHandler {
	return &PermissionHandler{
		db:    db,
		redis: redis,
	}
}

// PermissionDTO represents a permission in API responses
type PermissionDTO struct {
	ID      uint     `json:"id"`
	Module  string   `json:"module"`
	Feature string   `json:"feature"`
	Actions []string `json:"actions"`
}

// RolePermissionDTO represents role-permission details
type RolePermissionDTO struct {
	PermissionID     uint     `json:"permissionId"`
	Module           string   `json:"module"`
	Feature          string   `json:"feature"`
	AvailableActions []string `json:"availableActions"`
	GrantedActions   []string `json:"grantedActions"`
}

// RolePermissionsResponse represents the response for GET /roles/{id}/permissions
type RolePermissionsResponse struct {
	RoleID      uint                    `json:"roleId"`
	RoleName    string                  `json:"roleName"`
	IsSystem    bool                    `json:"isSystem"`
	Permissions []RolePermissionDTO `json:"permissions"`
}

// UpdatePermissionsInput represents the request body for updating role permissions
type UpdatePermissionsInput struct {
	Permissions []struct {
		PermissionID uint     `json:"permissionId"`
		Actions      []string `json:"actions"`
	} `json:"permissions"`
}

// ListPermissions returns all available permissions
func (h *PermissionHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	var permissions []models.Permission
	if err := h.db.Order("module, feature").Find(&permissions).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch permissions", "INTERNAL_ERROR")
		return
	}

	// Convert to DTO
	permDTOs := make([]PermissionDTO, len(permissions))
	for i, p := range permissions {
		permDTOs[i] = PermissionDTO{
			ID:      p.ID,
			Module:  p.Module,
			Feature: p.Feature,
			Actions: []string(p.Actions),
		}
	}

	utils.Success(w, http.StatusOK, "", permDTOs)
}

// GetRolePermissions returns permissions for a specific role
func (h *PermissionHandler) GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	// Parse role ID
	idStr := chi.URLParam(r, "id")
	roleID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid role ID", "VALIDATION_ERROR")
		return
	}

	// Find role
	var role models.Role
	if err := h.db.First(&role, roleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.Error(w, http.StatusNotFound, "Role not found", "ROLE_NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch role", "INTERNAL_ERROR")
		return
	}

	// Get all permissions
	var allPermissions []models.Permission
	if err := h.db.Order("module, feature").Find(&allPermissions).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch permissions", "INTERNAL_ERROR")
		return
	}

	var permissionDTOs []RolePermissionDTO

	// For system role (like Super Admin), grant all actions
	if role.IsSystem && role.Name == "Super Admin" {
		for _, perm := range allPermissions {
			permissionDTOs = append(permissionDTOs, RolePermissionDTO{
				PermissionID:     perm.ID,
				Module:           perm.Module,
				Feature:          perm.Feature,
				AvailableActions: []string(perm.Actions),
				GrantedActions:   []string(perm.Actions), // All actions granted
			})
		}
	} else {
		// For regular roles, get assigned permissions
		var rolePermissions []models.RolePermission
		if err := h.db.Where("role_id = ?", roleID).Preload("Permission").Find(&rolePermissions).Error; err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to fetch role permissions", "INTERNAL_ERROR")
			return
		}

		// Build map of granted permissions
		grantedMap := make(map[uint][]string)
		for _, rp := range rolePermissions {
			grantedMap[rp.PermissionID] = []string(rp.Actions)
		}

		// Build response with all permissions, showing which are granted
		for _, perm := range allPermissions {
			grantedActions := grantedMap[perm.ID]
			if grantedActions == nil {
				grantedActions = []string{}
			}

			permissionDTOs = append(permissionDTOs, RolePermissionDTO{
				PermissionID:     perm.ID,
				Module:           perm.Module,
				Feature:          perm.Feature,
				AvailableActions: []string(perm.Actions),
				GrantedActions:   grantedActions,
			})
		}
	}

	response := RolePermissionsResponse{
		RoleID:      uint(roleID),
		RoleName:    role.Name,
		IsSystem:    role.IsSystem,
		Permissions: permissionDTOs,
	}

	utils.Success(w, http.StatusOK, "", response)
}

// UpdateRolePermissions updates permissions for a role
func (h *PermissionHandler) UpdateRolePermissions(w http.ResponseWriter, r *http.Request) {
	// Parse role ID
	idStr := chi.URLParam(r, "id")
	roleID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid role ID", "VALIDATION_ERROR")
		return
	}

	// Parse request body
	var input UpdatePermissionsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	// Find role
	var role models.Role
	if err := h.db.First(&role, roleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.Error(w, http.StatusNotFound, "Role not found", "ROLE_NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch role", "INTERNAL_ERROR")
		return
	}

	// Block system roles
	if role.IsSystem {
		utils.Error(w, http.StatusForbidden, "System role permissions cannot be modified", "SYSTEM_ROLE_PROTECTED")
		return
	}

	// Validate permissions and filter actions
	validatedPermissions := []models.RolePermission{}
	for _, p := range input.Permissions {
		// Check if permission exists
		var permission models.Permission
		if err := h.db.First(&permission, p.PermissionID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.Error(w, http.StatusBadRequest, "Invalid permission ID", "VALIDATION_ERROR")
				return
			}
			utils.Error(w, http.StatusInternalServerError, "Failed to validate permissions", "INTERNAL_ERROR")
			return
		}

		// Filter actions to only valid ones
		validActions := []string{}
		availableActions := []string(permission.Actions)
		for _, action := range p.Actions {
			if contains(availableActions, action) {
				validActions = append(validActions, action)
			}
		}

		// Only add if there are valid actions
		if len(validActions) > 0 {
			validatedPermissions = append(validatedPermissions, models.RolePermission{
				RoleID:       uint(roleID),
				PermissionID: p.PermissionID,
				Actions:      pq.StringArray(validActions),
			})
		}
	}

	// Use transaction for atomic update
	err = h.db.Transaction(func(tx *gorm.DB) error {
		// Delete existing role permissions
		if err := tx.Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error; err != nil {
			return err
		}

		// Insert new permissions
		if len(validatedPermissions) > 0 {
			if err := tx.Create(&validatedPermissions).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update permissions", "INTERNAL_ERROR")
		return
	}

	// Invalidate permission cache for all users with this role
	h.invalidatePermissionCache(uint(roleID))

	// Return updated permissions
	h.GetRolePermissions(w, r)
}

// invalidatePermissionCache removes cached permissions for all users with the given role
func (h *PermissionHandler) invalidatePermissionCache(roleID uint) {
	ctx := context.Background()

	// Get all users with this role
	var userRoles []struct {
		UserID uint
	}
	h.db.Table("user_roles").Where("role_id = ?", roleID).Select("user_id").Find(&userRoles)

	// Delete cache for each user
	for _, ur := range userRoles {
		h.redis.Del(ctx, "perms:"+strconv.FormatUint(uint64(ur.UserID), 10))
	}
}

// contains checks if a string slice contains a specific value
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
