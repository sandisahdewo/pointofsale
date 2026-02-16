package repositories

import (
	"testing"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCreateUser_ValidUser_Succeeds(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	user := &models.User{
		Name:         "John Doe",
		Email:        "john@example.com",
		PasswordHash: "hashed_password",
		Status:       "active",
		IsSuperAdmin: false,
	}

	err := repo.Create(user)
	require.NoError(t, err)
	assert.NotZero(t, user.ID, "user ID should be set after creation")
	assert.Equal(t, "John Doe", user.Name)
	assert.Equal(t, "john@example.com", user.Email)
	assert.NotZero(t, user.CreatedAt)
	assert.NotZero(t, user.UpdatedAt)
}

func TestCreateUser_DuplicateEmail_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create first user
	user1 := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Email = "duplicate@example.com"
	})
	assert.NotZero(t, user1.ID)

	// Try to create another user with same email
	user2 := &models.User{
		Name:         "Another User",
		Email:        "duplicate@example.com",
		PasswordHash: "hashed_password",
		Status:       "active",
	}

	err := repo.Create(user2)
	require.Error(t, err, "should return error for duplicate email")
}

func TestFindUserByEmail_Exists_ReturnsUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create a test user
	created := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Email = "findme@example.com"
		u.Name = "Find Me"
	})

	// Find by email
	found, err := repo.FindByEmail("findme@example.com")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "Find Me", found.Name)
	assert.Equal(t, "findme@example.com", found.Email)
}

func TestFindUserByEmail_CaseInsensitive_ReturnsUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create user with mixed case email
	created := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Email = "Test@Example.com"
		u.Name = "Test User"
	})

	// Find with lowercase email
	found, err := repo.FindByEmail("test@example.com")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "Test User", found.Name)

	// Find with uppercase email
	found2, err := repo.FindByEmail("TEST@EXAMPLE.COM")
	require.NoError(t, err)
	require.NotNil(t, found2)
	assert.Equal(t, created.ID, found2.ID)
}

func TestFindUserByEmail_NotExists_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Try to find non-existent email
	found, err := repo.FindByEmail("nonexistent@example.com")
	assert.Error(t, err, "should return error for non-existent email")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Nil(t, found)
}

func TestFindUserByID_Exists_ReturnsUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create a test user
	created := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "User By ID"
	})

	// Find by ID
	found, err := repo.FindByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "User By ID", found.Name)
}

func TestFindUserByID_WithRoles_EagerLoadsRoles(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create roles
	role1 := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Admin"
	})
	role2 := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Editor"
	})

	// Create user with roles
	user := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "User With Roles"
	})

	// Assign roles to user
	err := db.Model(&user).Association("Roles").Append([]models.Role{*role1, *role2})
	require.NoError(t, err)

	// Find user by ID (should preload roles)
	found, err := repo.FindByID(user.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, user.ID, found.ID)
	assert.Len(t, found.Roles, 2, "should load user roles")

	// Check role names
	roleNames := make([]string, len(found.Roles))
	for i, role := range found.Roles {
		roleNames[i] = role.Name
	}
	assert.Contains(t, roleNames, "Admin")
	assert.Contains(t, roleNames, "Editor")
}

func TestFindUserByID_NotExists_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Try to find non-existent ID
	found, err := repo.FindByID(99999)
	assert.Error(t, err, "should return error for non-existent ID")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Nil(t, found)
}

func TestUpdateUser_ValidUpdate_Succeeds(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create a user
	user := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "Original Name"
		u.Status = "active"
	})
	originalID := user.ID

	// Update user fields
	user.Name = "Updated Name"
	user.Status = "inactive"

	err := repo.Update(user)
	require.NoError(t, err)

	// Verify update
	found, err := repo.FindByID(originalID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", found.Name)
	assert.Equal(t, "inactive", found.Status)
}

func TestFindUserByIDWithPermissions_WithRolesAndPermissions_ReturnsUserAndPermissions(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create permissions
	perm1 := testutil.CreateTestPermission(t, db, func(p *models.Permission) {
		p.Module = "master"
		p.Feature = "category"
		p.Actions = []string{"view", "create", "edit", "delete"}
	})
	perm2 := testutil.CreateTestPermission(t, db, func(p *models.Permission) {
		p.Module = "master"
		p.Feature = "product"
		p.Actions = []string{"view", "create"}
	})

	// Create role
	role := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Manager"
	})

	// Create role permissions
	rolePerm1 := &models.RolePermission{
		RoleID:       role.ID,
		PermissionID: perm1.ID,
		Actions:      []string{"view", "create"},
	}
	err := db.Create(rolePerm1).Error
	require.NoError(t, err)

	rolePerm2 := &models.RolePermission{
		RoleID:       role.ID,
		PermissionID: perm2.ID,
		Actions:      []string{"view"},
	}
	err = db.Create(rolePerm2).Error
	require.NoError(t, err)

	// Create user with role
	user := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "User With Permissions"
	})
	err = db.Model(&user).Association("Roles").Append(role)
	require.NoError(t, err)

	// Find user with permissions
	foundUser, rolePermissions, err := repo.FindByIDWithPermissions(user.ID)
	require.NoError(t, err)
	require.NotNil(t, foundUser)
	assert.Equal(t, user.ID, foundUser.ID)
	assert.Len(t, foundUser.Roles, 1, "should load user roles")
	assert.Equal(t, "Manager", foundUser.Roles[0].Name)

	// Verify role permissions
	require.Len(t, rolePermissions, 2, "should return 2 role permissions")

	// Check that permissions are preloaded
	for _, rp := range rolePermissions {
		assert.NotZero(t, rp.Permission.ID, "permission should be preloaded")
		assert.NotEmpty(t, rp.Permission.Module, "permission module should be loaded")
	}
}

func TestFindUserByIDWithPermissions_UserWithoutRoles_ReturnsEmptyPermissions(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create user without roles
	user := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "User Without Roles"
	})

	// Find user with permissions
	foundUser, rolePermissions, err := repo.FindByIDWithPermissions(user.ID)
	require.NoError(t, err)
	require.NotNil(t, foundUser)
	assert.Equal(t, user.ID, foundUser.ID)
	assert.Empty(t, foundUser.Roles, "should have no roles")
	assert.Empty(t, rolePermissions, "should have no role permissions")
}

func TestFindUserByIDWithPermissions_NotExists_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Try to find non-existent user
	foundUser, rolePermissions, err := repo.FindByIDWithPermissions(99999)
	assert.Error(t, err, "should return error for non-existent user")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Nil(t, foundUser)
	assert.Nil(t, rolePermissions)
}
