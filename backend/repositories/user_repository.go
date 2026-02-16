package repositories

import (
	"github.com/pointofsale/backend/models"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	Update(user *models.User) error
	FindByIDWithPermissions(id uint) (*models.User, []models.RolePermission, error)
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
