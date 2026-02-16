package testutil

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/utils"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// hashTestPassword returns a hashed version of the test password "Password@123".
func hashTestPassword(t *testing.T) string {
	t.Helper()
	hash, err := utils.HashPassword("Password@123")
	require.NoError(t, err, "failed to hash test password")
	return hash
}

// CreateTestUser creates a user in the test database with sensible defaults.
// Override fields using optional functions.
func CreateTestUser(t *testing.T, db *gorm.DB, overrides ...func(*models.User)) *models.User {
	t.Helper()

	user := &models.User{
		Name:         "Test User",
		Email:        fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8]),
		PasswordHash: hashTestPassword(t),
		Status:       "active",
		IsSuperAdmin: false,
	}

	// Apply overrides
	for _, override := range overrides {
		override(user)
	}

	err := db.Create(user).Error
	require.NoError(t, err, "failed to create test user")

	return user
}

// CreateTestRole creates a role in the test database.
func CreateTestRole(t *testing.T, db *gorm.DB, overrides ...func(*models.Role)) *models.Role {
	t.Helper()

	role := &models.Role{
		Name:        fmt.Sprintf("Test Role %s", uuid.New().String()[:8]),
		Description: "Test role description",
		IsSystem:    false,
	}

	// Apply overrides
	for _, override := range overrides {
		override(role)
	}

	err := db.Create(role).Error
	require.NoError(t, err, "failed to create test role")

	return role
}

// CreateTestPermission creates a permission in the test database.
func CreateTestPermission(t *testing.T, db *gorm.DB, overrides ...func(*models.Permission)) *models.Permission {
	t.Helper()

	permission := &models.Permission{
		Module:  "test",
		Feature: fmt.Sprintf("feature-%s", uuid.New().String()[:8]),
		Actions: []string{"view", "create", "edit", "delete"},
	}

	// Apply overrides
	for _, override := range overrides {
		override(permission)
	}

	err := db.Create(permission).Error
	require.NoError(t, err, "failed to create test permission")

	return permission
}

// CreateTestSuperAdmin creates a super admin user with the Super Admin role.
func CreateTestSuperAdmin(t *testing.T, db *gorm.DB) *models.User {
	t.Helper()

	// Create super admin role
	superAdminRole := CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Super Admin"
		r.Description = "Super administrator with all permissions"
	})

	// Create super admin user
	superAdmin := CreateTestUser(t, db, func(u *models.User) {
		u.Name = "Super Admin"
		u.Email = fmt.Sprintf("superadmin-%s@example.com", uuid.New().String()[:8])
		u.IsSuperAdmin = true
		u.Roles = []models.Role{*superAdminRole}
	})

	return superAdmin
}
