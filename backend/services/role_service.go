package services

import (
	"strings"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/repositories"
	"gorm.io/gorm"
)

// RoleInput is the DTO for creating and updating roles
type RoleInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// RoleService handles role business logic
type RoleService struct {
	roleRepo repositories.RoleRepository
}

// NewRoleService creates a new role service instance
func NewRoleService(roleRepo repositories.RoleRepository) *RoleService {
	return &RoleService{roleRepo: roleRepo}
}

// ListRoles returns paginated roles with user counts
func (s *RoleService) ListRoles(page, pageSize int, search, sortBy, sortDir string) ([]repositories.RoleWithCount, int64, *ServiceError) {
	roles, total, err := s.roleRepo.List(page, pageSize, search, sortBy, sortDir)
	if err != nil {
		return nil, 0, &ServiceError{
			Err:     err,
			Message: "Failed to list roles",
			Code:    "INTERNAL_ERROR",
		}
	}
	return roles, total, nil
}

// GetRole returns a role by ID
func (s *RoleService) GetRole(id uint) (*models.Role, *ServiceError) {
	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &ServiceError{
				Err:     ErrNotFound,
				Message: "Role not found",
				Code:    "ROLE_NOT_FOUND",
			}
		}
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to get role",
			Code:    "INTERNAL_ERROR",
		}
	}
	return role, nil
}

// CreateRole creates a new role with validation
func (s *RoleService) CreateRole(input RoleInput) (*models.Role, *ServiceError) {
	// Validate name
	trimmedName := strings.TrimSpace(input.Name)
	if trimmedName == "" {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Name is required",
			Code:    "VALIDATION_ERROR",
		}
	}
	if len(trimmedName) < 2 || len(trimmedName) > 255 {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Name must be between 2 and 255 characters",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Check uniqueness
	existing, _ := s.roleRepo.FindByName(trimmedName)
	if existing != nil {
		return nil, &ServiceError{
			Err:     ErrConflict,
			Message: "Role name already exists",
			Code:    "ROLE_NAME_EXISTS",
		}
	}

	// Create role
	role := &models.Role{
		Name:        trimmedName,
		Description: strings.TrimSpace(input.Description),
		IsSystem:    false,
	}

	if err := s.roleRepo.Create(role); err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to create role",
			Code:    "INTERNAL_ERROR",
		}
	}

	return role, nil
}

// UpdateRole updates an existing role
func (s *RoleService) UpdateRole(id uint, input RoleInput) (*models.Role, *ServiceError) {
	// Find role
	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &ServiceError{
				Err:     ErrNotFound,
				Message: "Role not found",
				Code:    "ROLE_NOT_FOUND",
			}
		}
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to get role",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Block system roles
	if role.IsSystem {
		return nil, &ServiceError{
			Err:     ErrForbidden,
			Message: "System roles cannot be modified",
			Code:    "SYSTEM_ROLE_PROTECTED",
		}
	}

	// Validate name
	trimmedName := strings.TrimSpace(input.Name)
	if trimmedName == "" {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Name is required",
			Code:    "VALIDATION_ERROR",
		}
	}
	if len(trimmedName) < 2 || len(trimmedName) > 255 {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Name must be between 2 and 255 characters",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Check uniqueness excluding self
	existing, _ := s.roleRepo.FindByNameExcluding(trimmedName, id)
	if existing != nil {
		return nil, &ServiceError{
			Err:     ErrConflict,
			Message: "Role name already exists",
			Code:    "ROLE_NAME_EXISTS",
		}
	}

	// Update fields
	role.Name = trimmedName
	role.Description = strings.TrimSpace(input.Description)

	if err := s.roleRepo.Update(role); err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to update role",
			Code:    "INTERNAL_ERROR",
		}
	}

	return role, nil
}

// DeleteRole deletes a role by ID
func (s *RoleService) DeleteRole(id uint) *ServiceError {
	// Find role
	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &ServiceError{
				Err:     ErrNotFound,
				Message: "Role not found",
				Code:    "ROLE_NOT_FOUND",
			}
		}
		return &ServiceError{
			Err:     err,
			Message: "Failed to get role",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Block system roles
	if role.IsSystem {
		return &ServiceError{
			Err:     ErrForbidden,
			Message: "System roles cannot be deleted",
			Code:    "SYSTEM_ROLE_PROTECTED",
		}
	}

	// Delete role (CASCADE removes user_roles and role_permissions)
	if err := s.roleRepo.Delete(id); err != nil {
		return &ServiceError{
			Err:     err,
			Message: "Failed to delete role",
			Code:    "INTERNAL_ERROR",
		}
	}

	return nil
}
