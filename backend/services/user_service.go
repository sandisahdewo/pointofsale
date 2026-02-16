package services

import (
	"crypto/rand"
	"encoding/base64"
	"strings"

	"github.com/pointofsale/backend/config"
	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/repositories"
	"github.com/pointofsale/backend/utils"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// UserRepositoryForUsers defines the repository interface needed by UserService
type UserRepositoryForUsers interface {
	Create(user *models.User) error
	FindByID(id uint) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindByEmailExcluding(email string, excludeID uint) (*models.User, error)
	Update(user *models.User) error
	List(params repositories.PaginationParams, status string) ([]models.User, int64, error)
	Delete(id uint) error
	SyncRoles(userID uint, roleIDs []uint) error
}

// UserEmailService defines the email operations for user management
type UserEmailService interface {
	SendUserCredentials(toEmail, userName, tempPassword string) error
	SendUserApproved(toEmail, userName string) error
	SendUserRejected(toEmail, userName string) error
}

// UserService handles user management business logic
type UserService struct {
	userRepo     UserRepositoryForUsers
	redis        *redis.Client
	config       *config.Config
	emailService UserEmailService
}

// NewUserService creates a new user service instance
func NewUserService(userRepo UserRepositoryForUsers, rdb *redis.Client, cfg *config.Config, emailSvc UserEmailService) *UserService {
	return &UserService{
		userRepo:     userRepo,
		redis:        rdb,
		config:       cfg,
		emailService: emailSvc,
	}
}

// CreateUserInput represents the input for creating a user
type CreateUserInput struct {
	Name           string   `json:"name"`
	Email          string   `json:"email"`
	Phone          string   `json:"phone,omitempty"`
	Address        string   `json:"address,omitempty"`
	RoleIDs        []uint   `json:"roleIds,omitempty"`
	ProfilePicture *string  `json:"profilePicture,omitempty"`
}

// UpdateUserInput represents the input for updating a user
type UpdateUserInput struct {
	Name           string   `json:"name"`
	Email          string   `json:"email"`
	Phone          string   `json:"phone,omitempty"`
	Address        string   `json:"address,omitempty"`
	RoleIDs        []uint   `json:"roleIds,omitempty"`
	Status         string   `json:"status,omitempty"`
	ProfilePicture *string  `json:"profilePicture,omitempty"`
}

// ListUsers returns paginated users with optional filtering
func (s *UserService) ListUsers(params repositories.PaginationParams, status string) ([]models.User, int64, error) {
	return s.userRepo.List(params, status)
}

// GetUser returns a single user by ID
func (s *UserService) GetUser(id uint) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &ServiceError{
				Err:     ErrNotFound,
				Message: "User not found",
				Code:    "USER_NOT_FOUND",
			}
		}
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to fetch user",
			Code:    "INTERNAL_ERROR",
		}
	}
	return user, nil
}

// CreateUser creates a new user with a generated password
func (s *UserService) CreateUser(input CreateUserInput) (*models.User, error) {
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

	// Check email uniqueness
	normalizedEmail := strings.ToLower(input.Email)
	existing, _ := s.userRepo.FindByEmail(normalizedEmail)
	if existing != nil {
		return nil, &ServiceError{
			Err:     ErrConflict,
			Message: "Email already exists",
			Code:    "EMAIL_EXISTS",
		}
	}

	// Generate temporary password
	tempPassword := generateTempPassword()

	// Hash password
	hashedPassword, err := utils.HashPassword(tempPassword)
	if err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to process password",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Create user
	user := &models.User{
		Name:           input.Name,
		Email:          normalizedEmail,
		Phone:          input.Phone,
		Address:        input.Address,
		PasswordHash:   hashedPassword,
		ProfilePicture: input.ProfilePicture,
		Status:         "active",
		IsSuperAdmin:   false,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to create user",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Assign roles if provided
	if len(input.RoleIDs) > 0 {
		if err := s.userRepo.SyncRoles(user.ID, input.RoleIDs); err != nil {
			// Log error but don't fail the create operation
			_ = err
		}
	}

	// Send credentials email (non-blocking)
	if s.emailService != nil {
		_ = s.emailService.SendUserCredentials(user.Email, user.Name, tempPassword)
	}

	// Reload user with roles
	createdUser, _ := s.userRepo.FindByID(user.ID)
	if createdUser != nil {
		return createdUser, nil
	}

	return user, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(id uint, input UpdateUserInput) (*models.User, error) {
	// Find existing user
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &ServiceError{
				Err:     ErrNotFound,
				Message: "User not found",
				Code:    "USER_NOT_FOUND",
			}
		}
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to fetch user",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Super admin protection: cannot change status or isSuperAdmin
	if user.IsSuperAdmin {
		if input.Status != "" && input.Status != user.Status {
			return nil, &ServiceError{
				Err:     ErrForbidden,
				Message: "Cannot change super admin status",
				Code:    "SUPER_ADMIN_PROTECTED",
			}
		}
	}

	// Validate name
	if input.Name != "" {
		if len(input.Name) < 2 || len(input.Name) > 255 {
			return nil, &ServiceError{
				Err:     ErrValidation,
				Message: "Name must be between 2 and 255 characters",
				Code:    "VALIDATION_ERROR",
			}
		}
		user.Name = input.Name
	}

	// Validate and check email uniqueness
	if input.Email != "" {
		if !utils.ValidateEmail(input.Email) {
			return nil, &ServiceError{
				Err:     ErrValidation,
				Message: "Invalid email format",
				Code:    "VALIDATION_ERROR",
			}
		}

		normalizedEmail := strings.ToLower(input.Email)
		// Check uniqueness excluding current user
		existing, _ := s.userRepo.FindByEmailExcluding(normalizedEmail, id)
		if existing != nil {
			return nil, &ServiceError{
				Err:     ErrConflict,
				Message: "Email already exists",
				Code:    "EMAIL_EXISTS",
			}
		}

		user.Email = normalizedEmail
	}

	// Update other fields
	if input.Phone != "" {
		user.Phone = input.Phone
	}
	if input.Address != "" {
		user.Address = input.Address
	}
	if input.ProfilePicture != nil {
		user.ProfilePicture = input.ProfilePicture
	}

	// Update status (if not super admin)
	if input.Status != "" && !user.IsSuperAdmin {
		if input.Status != "active" && input.Status != "inactive" {
			return nil, &ServiceError{
				Err:     ErrValidation,
				Message: "Invalid status. Must be 'active' or 'inactive'",
				Code:    "VALIDATION_ERROR",
			}
		}
		user.Status = input.Status
	}

	// Update user
	if err := s.userRepo.Update(user); err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to update user",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Sync roles if provided
	if input.RoleIDs != nil {
		if err := s.userRepo.SyncRoles(user.ID, input.RoleIDs); err != nil {
			// Log error but don't fail the update
			_ = err
		}
	}

	// Reload user with roles
	updatedUser, _ := s.userRepo.FindByID(user.ID)
	if updatedUser != nil {
		return updatedUser, nil
	}

	return user, nil
}

// DeleteUser deletes a user by ID
func (s *UserService) DeleteUser(id uint, currentUserID uint) error {
	// Find user
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &ServiceError{
				Err:     ErrNotFound,
				Message: "User not found",
				Code:    "USER_NOT_FOUND",
			}
		}
		return &ServiceError{
			Err:     err,
			Message: "Failed to fetch user",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Block super admin deletion
	if user.IsSuperAdmin {
		return &ServiceError{
			Err:     ErrForbidden,
			Message: "Super admin cannot be deleted",
			Code:    "SUPER_ADMIN_PROTECTED",
		}
	}

	// Block self-deletion
	if user.ID == currentUserID {
		return &ServiceError{
			Err:     ErrForbidden,
			Message: "Cannot delete your own account",
			Code:    "SELF_DELETION_FORBIDDEN",
		}
	}

	// Delete user
	if err := s.userRepo.Delete(id); err != nil {
		return &ServiceError{
			Err:     err,
			Message: "Failed to delete user",
			Code:    "INTERNAL_ERROR",
		}
	}

	return nil
}

// ApproveUser approves a pending user and sets them to active
func (s *UserService) ApproveUser(id uint) (*models.User, error) {
	// Find user
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &ServiceError{
				Err:     ErrNotFound,
				Message: "User not found",
				Code:    "USER_NOT_FOUND",
			}
		}
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to fetch user",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Check status is pending
	if user.Status != "pending" {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "User is not pending approval",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Update status to active
	user.Status = "active"
	if err := s.userRepo.Update(user); err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to approve user",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Send approval email (non-blocking)
	if s.emailService != nil {
		_ = s.emailService.SendUserApproved(user.Email, user.Name)
	}

	return user, nil
}

// RejectUser rejects a pending user and deletes them
func (s *UserService) RejectUser(id uint) error {
	// Find user
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &ServiceError{
				Err:     ErrNotFound,
				Message: "User not found",
				Code:    "USER_NOT_FOUND",
			}
		}
		return &ServiceError{
			Err:     err,
			Message: "Failed to fetch user",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Check status is pending
	if user.Status != "pending" {
		return &ServiceError{
			Err:     ErrValidation,
			Message: "User is not pending approval",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Send rejection email before deletion (non-blocking)
	if s.emailService != nil {
		_ = s.emailService.SendUserRejected(user.Email, user.Name)
	}

	// Delete user
	if err := s.userRepo.Delete(id); err != nil {
		return &ServiceError{
			Err:     err,
			Message: "Failed to reject user",
			Code:    "INTERNAL_ERROR",
		}
	}

	return nil
}

// generateTempPassword generates a random 16-character temporary password
func generateTempPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"
	const length = 16

	b := make([]byte, length)
	rand.Read(b)

	password := make([]byte, length)
	for i := 0; i < length; i++ {
		password[i] = charset[int(b[i])%len(charset)]
	}

	return string(password)
}

// Alternative: generate temp password with base64 encoding
func generateTempPasswordBase64() string {
	b := make([]byte, 12)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
