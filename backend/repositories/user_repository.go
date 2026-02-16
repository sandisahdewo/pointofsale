package repositories

import (
	"github.com/pointofsale/backend/models"
	"gorm.io/gorm"
)

// PaginationParams holds pagination and filtering parameters
type PaginationParams struct {
	Page     int
	PageSize int
	Search   string
	SortBy   string
	SortDir  string
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	Update(user *models.User) error
	FindByIDWithPermissions(id uint) (*models.User, []models.RolePermission, error)
	// NEW for Stage 3:
	List(params PaginationParams, status string) ([]models.User, int64, error)
	Delete(id uint) error
	SyncRoles(userID uint, roleIDs []uint) error
	FindByEmailExcluding(email string, excludeID uint) (*models.User, error)
}

// UserRepositoryImpl implements UserRepository interface
type UserRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *gorm.DB) *UserRepositoryImpl {
	return &UserRepositoryImpl{db: db}
}

// Create creates a new user in the database
func (r *UserRepositoryImpl) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// FindByEmail finds a user by email (case-insensitive) and preloads roles
func (r *UserRepositoryImpl) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("LOWER(email) = LOWER(?)", email).
		Preload("Roles").
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID finds a user by ID and preloads roles
func (r *UserRepositoryImpl) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Roles").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update saves changes to an existing user
func (r *UserRepositoryImpl) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// FindByIDWithPermissions finds a user with their roles and role permissions
func (r *UserRepositoryImpl) FindByIDWithPermissions(id uint) (*models.User, []models.RolePermission, error) {
	// Find user with roles
	var user models.User
	err := r.db.Preload("Roles").First(&user, id).Error
	if err != nil {
		return nil, nil, err
	}

	// If user has no roles, return empty permissions
	if len(user.Roles) == 0 {
		return &user, []models.RolePermission{}, nil
	}

	// Extract role IDs
	roleIDs := make([]uint, len(user.Roles))
	for i, role := range user.Roles {
		roleIDs[i] = role.ID
	}

	// Find role permissions for user's roles
	var rolePermissions []models.RolePermission
	err = r.db.Where("role_id IN ?", roleIDs).
		Preload("Permission").
		Find(&rolePermissions).Error
	if err != nil {
		return nil, nil, err
	}

	return &user, rolePermissions, nil
}

// List returns paginated users with optional search and status filter
func (r *UserRepositoryImpl) List(params PaginationParams, status string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// Build base query
	query := r.db.Model(&models.User{})

	// Apply search filter (name OR email, case-insensitive partial match)
	if params.Search != "" {
		searchPattern := "%" + params.Search + "%"
		query = query.Where("LOWER(name) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?)", searchPattern, searchPattern)
	}

	// Apply status filter
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderClause := params.SortBy + " " + params.SortDir
	query = query.Order(orderClause)

	// Apply pagination
	offset := (params.Page - 1) * params.PageSize
	query = query.Offset(offset).Limit(params.PageSize)

	// Preload roles
	query = query.Preload("Roles")

	// Execute query
	if err := query.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Delete removes a user from the database
func (r *UserRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

// SyncRoles replaces a user's roles with a new set
func (r *UserRepositoryImpl) SyncRoles(userID uint, roleIDs []uint) error {
	// Find user
	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return err
	}

	// Clear existing roles
	if err := r.db.Model(&user).Association("Roles").Clear(); err != nil {
		return err
	}

	// If no new roles, we're done
	if len(roleIDs) == 0 {
		return nil
	}

	// Find the new roles
	var roles []models.Role
	if err := r.db.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		return err
	}

	// Append new roles
	return r.db.Model(&user).Association("Roles").Append(roles)
}

// FindByEmailExcluding finds a user by email (case-insensitive), excluding a specific user ID
func (r *UserRepositoryImpl) FindByEmailExcluding(email string, excludeID uint) (*models.User, error) {
	var user models.User
	err := r.db.Where("LOWER(email) = LOWER(?) AND id != ?", email, excludeID).
		Preload("Roles").
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
