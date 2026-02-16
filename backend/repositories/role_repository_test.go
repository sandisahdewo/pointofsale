package repositories

import (
	"testing"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestListRoles_WithUserCount_ReturnsCorrectCounts verifies that List returns roles with accurate user counts
func TestListRoles_WithUserCount_ReturnsCorrectCounts(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRoleRepository(db)

	// Create roles
	role1 := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Manager"
	})
	role2 := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Cashier"
	})
	role3 := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Warehouse"
	})
	_ = role3 // role3 has 0 users (created but not assigned)

	// Create users and assign roles
	// role1 has 2 users
	user1 := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "User 1"
	})
	db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", user1.ID, role1.ID)

	user2 := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "User 2"
	})
	db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", user2.ID, role1.ID)

	// role2 has 1 user
	user3 := testutil.CreateTestUser(t, db, func(u *models.User) {
		u.Name = "User 3"
	})
	db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", user3.ID, role2.ID)

	// role3 has 0 users

	// Call List
	roles, total, err := repo.List(1, 10, "", "id", "asc")
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, roles, 3)

	// Find each role and check user count
	var managerRole *RoleWithCount
	var cashierRole *RoleWithCount
	var warehouseRole *RoleWithCount
	for i := range roles {
		if roles[i].Name == "Manager" {
			managerRole = &roles[i]
		} else if roles[i].Name == "Cashier" {
			cashierRole = &roles[i]
		} else if roles[i].Name == "Warehouse" {
			warehouseRole = &roles[i]
		}
	}

	require.NotNil(t, managerRole)
	require.NotNil(t, cashierRole)
	require.NotNil(t, warehouseRole)

	assert.Equal(t, int64(2), managerRole.UserCount)
	assert.Equal(t, int64(1), cashierRole.UserCount)
	assert.Equal(t, int64(0), warehouseRole.UserCount)
}

// TestListRoles_WithSearch_FiltersCorrectly verifies search by name and description
func TestListRoles_WithSearch_FiltersCorrectly(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRoleRepository(db)

	testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Manager"
		r.Description = "Manages daily operations"
	})
	testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Cashier"
		r.Description = "Handles cash transactions"
	})
	testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Warehouse Staff"
		r.Description = "Manages inventory"
	})

	// Search for "cash" should match Cashier
	roles, total, err := repo.List(1, 10, "cash", "id", "asc")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, roles, 1)
	assert.Equal(t, "Cashier", roles[0].Name)

	// Search for "manage" should match Manager (name) and Warehouse Staff (description)
	roles, total, err = repo.List(1, 10, "manage", "id", "asc")
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, roles, 2)
}

// TestListRoles_Pagination_Works verifies pagination parameters
func TestListRoles_Pagination_Works(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRoleRepository(db)

	// Create 5 roles
	for i := 1; i <= 5; i++ {
		testutil.CreateTestRole(t, db, func(r *models.Role) {
			r.Name = "Role " + string(rune('A'+i-1))
		})
	}

	// Page 1, size 2
	roles, total, err := repo.List(1, 2, "", "id", "asc")
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, roles, 2)

	// Page 2, size 2
	rolesPage2, total2, err := repo.List(2, 2, "", "id", "asc")
	require.NoError(t, err)
	assert.Equal(t, int64(5), total2)
	assert.Len(t, rolesPage2, 2)

	// Ensure different results
	assert.NotEqual(t, roles[0].ID, rolesPage2[0].ID)
}

// TestListRoles_SortByName_Works verifies sorting
func TestListRoles_SortByName_Works(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRoleRepository(db)

	testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Zebra"
	})
	testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Alpha"
	})
	testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Beta"
	})

	// Sort by name asc
	roles, _, err := repo.List(1, 10, "", "name", "asc")
	require.NoError(t, err)
	assert.Equal(t, "Alpha", roles[0].Name)
	assert.Equal(t, "Beta", roles[1].Name)
	assert.Equal(t, "Zebra", roles[2].Name)

	// Sort by name desc
	roles, _, err = repo.List(1, 10, "", "name", "desc")
	require.NoError(t, err)
	assert.Equal(t, "Zebra", roles[0].Name)
	assert.Equal(t, "Beta", roles[1].Name)
	assert.Equal(t, "Alpha", roles[2].Name)
}

// TestCreateRole_ValidData_Succeeds verifies role creation
func TestCreateRole_ValidData_Succeeds(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRoleRepository(db)

	role := &models.Role{
		Name:        "Supervisor",
		Description: "Supervises operations",
		IsSystem:    false,
	}

	err := repo.Create(role)
	require.NoError(t, err)
	assert.NotZero(t, role.ID)

	// Verify in database
	var found models.Role
	err = db.First(&found, role.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "Supervisor", found.Name)
	assert.Equal(t, "Supervises operations", found.Description)
	assert.False(t, found.IsSystem)
}

// TestCreateRole_DuplicateName_ReturnsError verifies unique name constraint
func TestCreateRole_DuplicateName_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRoleRepository(db)

	// Create first role
	role1 := &models.Role{
		Name:        "Manager",
		Description: "First manager",
	}
	err := repo.Create(role1)
	require.NoError(t, err)

	// Attempt duplicate name
	role2 := &models.Role{
		Name:        "Manager",
		Description: "Second manager",
	}
	err = repo.Create(role2)
	assert.Error(t, err)
}

// TestFindByID_Exists_ReturnsRole verifies finding role by ID
func TestFindByID_Exists_ReturnsRole(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRoleRepository(db)

	created := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Accountant"
	})

	found, err := repo.FindByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "Accountant", found.Name)
}

// TestFindByID_NotFound_ReturnsError verifies error for non-existent ID
func TestFindByID_NotFound_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRoleRepository(db)

	_, err := repo.FindByID(99999)
	assert.Error(t, err)
}

// TestFindByName_Exists_ReturnsRole verifies finding role by name
func TestFindByName_Exists_ReturnsRole(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRoleRepository(db)

	testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Supervisor"
	})

	found, err := repo.FindByName("Supervisor")
	require.NoError(t, err)
	assert.Equal(t, "Supervisor", found.Name)
}

// TestFindByName_NotFound_ReturnsError verifies error for non-existent name
func TestFindByName_NotFound_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRoleRepository(db)

	_, err := repo.FindByName("NonExistent")
	assert.Error(t, err)
}

// TestFindByNameExcluding_Exists_ReturnsRole verifies finding role by name excluding specific ID
func TestFindByNameExcluding_Exists_ReturnsRole(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRoleRepository(db)

	role1 := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Manager"
	})
	role2 := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "Supervisor"
	})

	// Find "Manager" excluding role1.ID should not find anything
	_, err := repo.FindByNameExcluding("Manager", role1.ID)
	assert.Error(t, err)

	// Find "Manager" excluding role2.ID should find role1
	found, err := repo.FindByNameExcluding("Manager", role2.ID)
	require.NoError(t, err)
	assert.Equal(t, role1.ID, found.ID)
}

// TestUpdate_ValidData_UpdatesRole verifies role update
func TestUpdate_ValidData_UpdatesRole(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRoleRepository(db)

	role := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "OldName"
		r.Description = "OldDescription"
	})

	// Update fields
	role.Name = "NewName"
	role.Description = "NewDescription"
	err := repo.Update(role)
	require.NoError(t, err)

	// Verify update
	found, err := repo.FindByID(role.ID)
	require.NoError(t, err)
	assert.Equal(t, "NewName", found.Name)
	assert.Equal(t, "NewDescription", found.Description)
}

// TestDeleteRole_CascadesPermissionsAndUserRoles verifies cascade delete
func TestDeleteRole_CascadesPermissionsAndUserRoles(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRoleRepository(db)

	// Create role
	role := testutil.CreateTestRole(t, db, func(r *models.Role) {
		r.Name = "ToDelete"
	})

	// Create user and assign role
	user := testutil.CreateTestUser(t, db)
	db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", user.ID, role.ID)

	// Create permission and assign to role
	perm := testutil.CreateTestPermission(t, db, func(p *models.Permission) {
		p.Module = "Test"
		p.Feature = "Test Feature"
	})
	db.Exec("INSERT INTO role_permissions (role_id, permission_id, actions) VALUES (?, ?, ?)",
		role.ID, perm.ID, `{"read","create"}`)

	// Verify relationships exist
	var userRoleCount int64
	db.Table("user_roles").Where("role_id = ?", role.ID).Count(&userRoleCount)
	assert.Equal(t, int64(1), userRoleCount)

	var rolePermCount int64
	db.Table("role_permissions").Where("role_id = ?", role.ID).Count(&rolePermCount)
	assert.Equal(t, int64(1), rolePermCount)

	// Delete role
	err := repo.Delete(role.ID)
	require.NoError(t, err)

	// Verify role is deleted
	_, err = repo.FindByID(role.ID)
	assert.Error(t, err)

	// Verify user_roles cascade deleted
	db.Table("user_roles").Where("role_id = ?", role.ID).Count(&userRoleCount)
	assert.Equal(t, int64(0), userRoleCount)

	// Verify role_permissions cascade deleted
	db.Table("role_permissions").Where("role_id = ?", role.ID).Count(&rolePermCount)
	assert.Equal(t, int64(0), rolePermCount)
}

// TestDelete_NotFound_ReturnsError verifies error when deleting non-existent role
func TestDelete_NotFound_ReturnsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewRoleRepository(db)

	err := repo.Delete(99999)
	assert.Error(t, err)
}
