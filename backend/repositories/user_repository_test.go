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

// NEW TESTS FOR STAGE 3 - Task #4

func TestListUsers_Pagination_ReturnsCorrectPage(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create 15 test users
	for i := 1; i <= 15; i++ {
		testutil.CreateTestUser(t, db, func(u *models.User) {
			u.Name = "User " + string(rune(i))
		})
	}

	// Request page 1 with page size 10
	params := PaginationParams{
		Page:     1,
		PageSize: 10,
		SortBy:   "id",
		SortDir:  "asc",
	}

	users, total, err := repo.List(params, "")
	require.NoError(t, err)
	assert.Len(t, users, 10, "should return 10 users")
	assert.Equal(t, int64(15), total, "total count should be 15")

	// Request page 2
	params.Page = 2
	users, total, err = repo.List(params, "")
	require.NoError(t, err)
	assert.Len(t, users, 5, "should return 5 users on page 2")
	assert.Equal(t, int64(15), total, "total count should still be 15")
}

func TestListUsers_SearchByName_FiltersCorrectly(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create test users with different names
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "Alice Johnson"
	})
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "Bob Smith"
	})
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "Charlie Johnson"
	})

	// Search for "Johnson" (case-insensitive partial match)
	params := PaginationParams{
		Page:     1,
		PageSize: 10,
		Search:   "johnson",
		SortBy:   "id",
		SortDir:  "asc",
	}

	users, total, err := repo.List(params, "")
	require.NoError(t, err)
	assert.Equal(t, int64(2), total, "should find 2 users with 'Johnson' in name")
	assert.Len(t, users, 2)
}

func TestListUsers_SearchByEmail_FiltersCorrectly(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create test users with different emails
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Email = "alice@company.com"
	})
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Email = "bob@example.com"
	})
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Email = "charlie@company.com"
	})

	// Search for "company" (should match email)
	params := PaginationParams{
		Page:     1,
		PageSize: 10,
		Search:   "company",
		SortBy:   "id",
		SortDir:  "asc",
	}

	users, total, err := repo.List(params, "")
	require.NoError(t, err)
	assert.Equal(t, int64(2), total, "should find 2 users with 'company' in email")
	assert.Len(t, users, 2)
}

func TestListUsers_FilterByStatus_ReturnsMatchingOnly(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create users with different statuses
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Status = "active"
	})
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Status = "inactive"
	})
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Status = "pending"
	})
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Status = "active"
	})

	// Filter by "active" status
	params := PaginationParams{
		Page:     1,
		PageSize: 10,
		SortBy:   "id",
		SortDir:  "asc",
	}

	users, total, err := repo.List(params, "active")
	require.NoError(t, err)
	assert.Equal(t, int64(2), total, "should find 2 active users")
	assert.Len(t, users, 2)
	for _, user := range users {
		assert.Equal(t, "active", user.Status)
	}
}

func TestListUsers_SortByName_ReturnsOrdered(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create users with different names
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "Zoe"
	})
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "Alice"
	})
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "Bob"
	})

	// Sort by name ascending
	params := PaginationParams{
		Page:     1,
		PageSize: 10,
		SortBy:   "name",
		SortDir:  "asc",
	}

	users, total, err := repo.List(params, "")
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Equal(t, "Alice", users[0].Name)
	assert.Equal(t, "Bob", users[1].Name)
	assert.Equal(t, "Zoe", users[2].Name)

	// Sort by name descending
	params.SortDir = "desc"
	users, _, err = repo.List(params, "")
	require.NoError(t, err)
	assert.Equal(t, "Zoe", users[0].Name)
	assert.Equal(t, "Bob", users[1].Name)
	assert.Equal(t, "Alice", users[2].Name)
}

func TestListUsers_CombinedSearchAndFilter_Works(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create users with various names and statuses
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "Alice Johnson"
		u.Status = "active"
	})
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "Bob Johnson"
		u.Status = "inactive"
	})
	testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "Charlie Smith"
		u.Status = "active"
	})

	// Search for "Johnson" AND status "active"
	params := PaginationParams{
		Page:     1,
		PageSize: 10,
		Search:   "Johnson",
		SortBy:   "id",
		SortDir:  "asc",
	}

	users, total, err := repo.List(params, "active")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total, "should find only Alice Johnson (active)")
	assert.Len(t, users, 1)
	assert.Equal(t, "Alice Johnson", users[0].Name)
}

func TestListUsers_PreloadsRoles(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create role
	role := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Manager"
	})

	// Create user with role
	user := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "User With Role"
	})
	err := db.Model(&user).Association("Roles").Append(role)
	require.NoError(t, err)

	// List users - should preload roles
	params := PaginationParams{
		Page:     1,
		PageSize: 10,
		SortBy:   "id",
		SortDir:  "asc",
	}

	users, _, err := repo.List(params, "")
	require.NoError(t, err)

	// Find our user
	var foundUser *models.User
	for _, u := range users {
		if u.ID == user.ID {
			foundUser = &u
			break
		}
	}
	require.NotNil(t, foundUser, "should find our test user")
	assert.Len(t, foundUser.Roles, 1, "should preload roles")
	assert.Equal(t, "Manager", foundUser.Roles[0].Name)
}

func TestDeleteUser_Exists_RemovesUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create a user
	user := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "To Be Deleted"
	})
	userID := user.ID

	// Delete the user
	err := repo.Delete(userID)
	require.NoError(t, err)

	// Verify user is deleted
	_, err = repo.FindByID(userID)
	assert.Error(t, err, "should not find deleted user")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestDeleteUser_CascadesUserRoles(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create role
	role := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Test Role"
	})

	// Create user with role
	user := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "User With Role"
	})
	err := db.Model(&user).Association("Roles").Append(role)
	require.NoError(t, err)

	// Verify user has role
	foundUser, err := repo.FindByID(user.ID)
	require.NoError(t, err)
	assert.Len(t, foundUser.Roles, 1)

	// Delete user
	err = repo.Delete(user.ID)
	require.NoError(t, err)

	// Verify user_roles junction table entry is also deleted (cascade)
	var count int64
	err = db.Table("user_roles").Where("user_id = ?", user.ID).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "user_roles entries should be deleted via cascade")
}

func TestSyncRoles_ReplacesExistingRoles(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create roles
	role1 := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Role 1"
	})
	role2 := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Role 2"
	})
	role3 := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Role 3"
	})

	// Create user with roles 1 and 2
	user := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "User"
	})
	err := db.Model(&user).Association("Roles").Append([]models.Role{*role1, *role2})
	require.NoError(t, err)

	// Verify initial roles
	foundUser, err := repo.FindByID(user.ID)
	require.NoError(t, err)
	assert.Len(t, foundUser.Roles, 2)

	// Sync to new set of roles (role2 and role3)
	err = repo.SyncRoles(user.ID, []uint{role2.ID, role3.ID})
	require.NoError(t, err)

	// Verify roles are replaced
	foundUser, err = repo.FindByID(user.ID)
	require.NoError(t, err)
	assert.Len(t, foundUser.Roles, 2, "should have 2 roles after sync")

	// Check role names
	roleNames := make([]string, len(foundUser.Roles))
	for i, role := range foundUser.Roles {
		roleNames[i] = role.Name
	}
	assert.Contains(t, roleNames, "Role 2")
	assert.Contains(t, roleNames, "Role 3")
	assert.NotContains(t, roleNames, "Role 1", "Role 1 should be removed")
}

func TestSyncRoles_EmptyRoleIDs_RemovesAllRoles(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create role
	role := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Role"
	})

	// Create user with role
	user := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "User"
	})
	err := db.Model(&user).Association("Roles").Append(role)
	require.NoError(t, err)

	// Sync to empty role list
	err = repo.SyncRoles(user.ID, []uint{})
	require.NoError(t, err)

	// Verify all roles are removed
	foundUser, err := repo.FindByID(user.ID)
	require.NoError(t, err)
	assert.Empty(t, foundUser.Roles, "should have no roles after syncing with empty list")
}

func TestFindByEmailExcluding_ExistsButNotExcluded_ReturnsUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create users
	user1 := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Email = "duplicate@example.com"
	})
	user2 := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Email = "unique@example.com"
	})

	// Find "duplicate@example.com" excluding user2 (should find user1)
	found, err := repo.FindByEmailExcluding("duplicate@example.com", user2.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, user1.ID, found.ID)
}

func TestFindByEmailExcluding_ExistsAndExcluded_ReturnsNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create user
	user := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Email = "test@example.com"
	})

	// Find same email excluding this user (should return not found)
	found, err := repo.FindByEmailExcluding("test@example.com", user.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Nil(t, found)
}

func TestFindByEmailExcluding_CaseInsensitive_Works(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewUserRepository(db)

	// Create users
	user1 := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Email = "Test@Example.com"
	})
	user2 := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Email = "other@example.com"
	})

	// Find with different case, excluding user2
	found, err := repo.FindByEmailExcluding("test@example.com", user2.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, user1.ID, found.ID)
}
