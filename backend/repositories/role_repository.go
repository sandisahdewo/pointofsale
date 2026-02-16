package repositories

import (
	"fmt"

	"github.com/pointofsale/backend/models"
	"gorm.io/gorm"
)

// RoleRepository defines the interface for role data operations
type RoleRepository interface {
	List(page, pageSize int, search, sortBy, sortDir string) ([]RoleWithCount, int64, error)
	FindByID(id uint) (*models.Role, error)
	FindByName(name string) (*models.Role, error)
	FindByNameExcluding(name string, excludeID uint) (*models.Role, error)
	Create(role *models.Role) error
	Update(role *models.Role) error
	Delete(id uint) error
}

// RoleWithCount adds userCount to role data
type RoleWithCount struct {
	models.Role
	UserCount int64 `json:"userCount"`
}

// RoleRepositoryImpl implements RoleRepository interface
type RoleRepositoryImpl struct {
	db *gorm.DB
}

// NewRoleRepository creates a new role repository instance
func NewRoleRepository(db *gorm.DB) *RoleRepositoryImpl {
	return &RoleRepositoryImpl{db: db}
}

// List returns paginated roles with user counts
func (r *RoleRepositoryImpl) List(page, pageSize int, search, sortBy, sortDir string) ([]RoleWithCount, int64, error) {
	var roles []RoleWithCount
	var total int64

	// Build base query
	query := r.db.Model(&models.Role{})

	// Apply search filter (case-insensitive, partial match on name or description)
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	if sortBy == "" {
		sortBy = "id"
	}
	if sortDir == "" {
		sortDir = "asc"
	}
	orderClause := fmt.Sprintf("%s %s", sortBy, sortDir)

	// Apply pagination
	offset := (page - 1) * pageSize

	// Execute query with LEFT JOIN to get user counts
	err := query.
		Select("roles.*, COALESCE(COUNT(user_roles.user_id), 0) as user_count").
		Joins("LEFT JOIN user_roles ON user_roles.role_id = roles.id").
		Group("roles.id").
		Order(orderClause).
		Offset(offset).
		Limit(pageSize).
		Find(&roles).Error

	if err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// FindByID finds a role by ID
func (r *RoleRepositoryImpl) FindByID(id uint) (*models.Role, error) {
	var role models.Role
	err := r.db.First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// FindByName finds a role by name (case-insensitive)
func (r *RoleRepositoryImpl) FindByName(name string) (*models.Role, error) {
	var role models.Role
	err := r.db.Where("LOWER(name) = LOWER(?)", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// FindByNameExcluding finds a role by name excluding a specific ID (for update uniqueness check)
func (r *RoleRepositoryImpl) FindByNameExcluding(name string, excludeID uint) (*models.Role, error) {
	var role models.Role
	err := r.db.Where("LOWER(name) = LOWER(?) AND id != ?", name, excludeID).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// Create creates a new role
func (r *RoleRepositoryImpl) Create(role *models.Role) error {
	return r.db.Create(role).Error
}

// Update saves changes to an existing role
func (r *RoleRepositoryImpl) Update(role *models.Role) error {
	return r.db.Save(role).Error
}

// Delete deletes a role by ID (CASCADE removes user_roles and role_permissions)
func (r *RoleRepositoryImpl) Delete(id uint) error {
	result := r.db.Delete(&models.Role{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
