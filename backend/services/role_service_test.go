package services

import (
	"testing"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// mockRoleRepository is a mock implementation for testing
type mockRoleRepository struct {
	listFn                func(page, pageSize int, search, sortBy, sortDir string) ([]repositories.RoleWithCount, int64, error)
	findByIDFn            func(id uint) (*models.Role, error)
	findByNameFn          func(name string) (*models.Role, error)
	findByNameExcludingFn func(name string, excludeID uint) (*models.Role, error)
	createFn              func(role *models.Role) error
	updateFn              func(role *models.Role) error
	deleteFn              func(id uint) error
}

func (m *mockRoleRepository) List(page, pageSize int, search, sortBy, sortDir string) ([]repositories.RoleWithCount, int64, error) {
	if m.listFn != nil {
		return m.listFn(page, pageSize, search, sortBy, sortDir)
	}
	return nil, 0, nil
}

func (m *mockRoleRepository) FindByID(id uint) (*models.Role, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(id)
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockRoleRepository) FindByName(name string) (*models.Role, error) {
	if m.findByNameFn != nil {
		return m.findByNameFn(name)
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockRoleRepository) FindByNameExcluding(name string, excludeID uint) (*models.Role, error) {
	if m.findByNameExcludingFn != nil {
		return m.findByNameExcludingFn(name, excludeID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockRoleRepository) Create(role *models.Role) error {
	if m.createFn != nil {
		return m.createFn(role)
	}
	return nil
}

func (m *mockRoleRepository) Update(role *models.Role) error {
	if m.updateFn != nil {
		return m.updateFn(role)
	}
	return nil
}

func (m *mockRoleRepository) Delete(id uint) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return nil
}

// TestListRoles_Valid_Succeeds verifies list delegation to repository
func TestListRoles_Valid_Succeeds(t *testing.T) {
	mockRepo := &mockRoleRepository{
		listFn: func(page, pageSize int, search, sortBy, sortDir string) ([]repositories.RoleWithCount, int64, error) {
			return []repositories.RoleWithCount{
				{Role: models.Role{ID: 1, Name: "Manager"}},
				{Role: models.Role{ID: 2, Name: "Cashier"}},
			}, 2, nil
		},
	}

	service := NewRoleService(mockRepo)
	roles, total, err := service.ListRoles(1, 10, "", "id", "asc")

	require.Nil(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, roles, 2)
}

// TestGetRole_Exists_ReturnsRole verifies getting role by ID
func TestGetRole_Exists_ReturnsRole(t *testing.T) {
	mockRepo := &mockRoleRepository{
		findByIDFn: func(id uint) (*models.Role, error) {
			return &models.Role{ID: 1, Name: "Manager"}, nil
		},
	}

	service := NewRoleService(mockRepo)
	role, err := service.GetRole(1)

	require.Nil(t, err)
	assert.Equal(t, "Manager", role.Name)
}

// TestGetRole_NotFound_ReturnsNotFoundError verifies 404 error
func TestGetRole_NotFound_ReturnsNotFoundError(t *testing.T) {
	mockRepo := &mockRoleRepository{
		findByIDFn: func(id uint) (*models.Role, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	service := NewRoleService(mockRepo)
	_, err := service.GetRole(99999)

	require.NotNil(t, err)
	assert.Equal(t, ErrNotFound, err.Err)
	assert.Equal(t, "ROLE_NOT_FOUND", err.Code)
}

// TestCreateRole_Valid_Succeeds verifies role creation with validation
func TestCreateRole_Valid_Succeeds(t *testing.T) {
	mockRepo := &mockRoleRepository{
		findByNameFn: func(name string) (*models.Role, error) {
			return nil, gorm.ErrRecordNotFound
		},
		createFn: func(role *models.Role) error {
			role.ID = 1
			return nil
		},
	}

	service := NewRoleService(mockRepo)
	input := RoleInput{
		Name:        "Supervisor",
		Description: "Supervises operations",
	}

	role, err := service.CreateRole(input)

	require.Nil(t, err)
	assert.Equal(t, "Supervisor", role.Name)
	assert.Equal(t, "Supervises operations", role.Description)
	assert.False(t, role.IsSystem)
}

// TestCreateRole_EmptyName_ReturnsValidationError verifies name required
func TestCreateRole_EmptyName_ReturnsValidationError(t *testing.T) {
	mockRepo := &mockRoleRepository{}
	service := NewRoleService(mockRepo)

	input := RoleInput{
		Name:        "",
		Description: "Test",
	}

	_, err := service.CreateRole(input)

	require.NotNil(t, err)
	assert.Equal(t, ErrValidation, err.Err)
	assert.Contains(t, err.Message, "Name is required")
}

// TestCreateRole_NameTooShort_ReturnsValidationError verifies name length
func TestCreateRole_NameTooShort_ReturnsValidationError(t *testing.T) {
	mockRepo := &mockRoleRepository{}
	service := NewRoleService(mockRepo)

	input := RoleInput{
		Name:        "A",
		Description: "Test",
	}

	_, err := service.CreateRole(input)

	require.NotNil(t, err)
	assert.Equal(t, ErrValidation, err.Err)
	assert.Contains(t, err.Message, "2 and 255 characters")
}

// TestCreateRole_NameTooLong_ReturnsValidationError verifies name length
func TestCreateRole_NameTooLong_ReturnsValidationError(t *testing.T) {
	mockRepo := &mockRoleRepository{}
	service := NewRoleService(mockRepo)

	longName := string(make([]byte, 256))
	for i := range longName {
		longName = longName[:i] + "A"
	}

	input := RoleInput{
		Name:        longName,
		Description: "Test",
	}

	_, err := service.CreateRole(input)

	require.NotNil(t, err)
	assert.Equal(t, ErrValidation, err.Err)
	assert.Contains(t, err.Message, "2 and 255 characters")
}

// TestCreateRole_DuplicateName_ReturnsConflict verifies uniqueness check
func TestCreateRole_DuplicateName_ReturnsConflict(t *testing.T) {
	mockRepo := &mockRoleRepository{
		findByNameFn: func(name string) (*models.Role, error) {
			return &models.Role{ID: 1, Name: "Manager"}, nil
		},
	}

	service := NewRoleService(mockRepo)
	input := RoleInput{
		Name:        "Manager",
		Description: "Test",
	}

	_, err := service.CreateRole(input)

	require.NotNil(t, err)
	assert.Equal(t, ErrConflict, err.Err)
	assert.Equal(t, "ROLE_NAME_EXISTS", err.Code)
	assert.Contains(t, err.Message, "already exists")
}

// TestUpdateRole_Valid_Succeeds verifies role update
func TestUpdateRole_Valid_Succeeds(t *testing.T) {
	mockRepo := &mockRoleRepository{
		findByIDFn: func(id uint) (*models.Role, error) {
			return &models.Role{ID: 1, Name: "OldName", IsSystem: false}, nil
		},
		findByNameExcludingFn: func(name string, excludeID uint) (*models.Role, error) {
			return nil, gorm.ErrRecordNotFound
		},
		updateFn: func(role *models.Role) error {
			return nil
		},
	}

	service := NewRoleService(mockRepo)
	input := RoleInput{
		Name:        "NewName",
		Description: "Updated description",
	}

	role, err := service.UpdateRole(1, input)

	require.Nil(t, err)
	assert.Equal(t, "NewName", role.Name)
	assert.Equal(t, "Updated description", role.Description)
}

// TestUpdateRole_NotFound_ReturnsNotFoundError verifies role existence check
func TestUpdateRole_NotFound_ReturnsNotFoundError(t *testing.T) {
	mockRepo := &mockRoleRepository{
		findByIDFn: func(id uint) (*models.Role, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	service := NewRoleService(mockRepo)
	input := RoleInput{Name: "Test"}

	_, err := service.UpdateRole(99999, input)

	require.NotNil(t, err)
	assert.Equal(t, ErrNotFound, err.Err)
}

// TestUpdateRole_SystemRole_ReturnsForbidden verifies system role protection
func TestUpdateRole_SystemRole_ReturnsForbidden(t *testing.T) {
	mockRepo := &mockRoleRepository{
		findByIDFn: func(id uint) (*models.Role, error) {
			return &models.Role{ID: 1, Name: "Super Admin", IsSystem: true}, nil
		},
	}

	service := NewRoleService(mockRepo)
	input := RoleInput{Name: "New Name"}

	_, err := service.UpdateRole(1, input)

	require.NotNil(t, err)
	assert.Equal(t, ErrForbidden, err.Err)
	assert.Equal(t, "SYSTEM_ROLE_PROTECTED", err.Code)
	assert.Contains(t, err.Message, "System roles cannot be modified")
}

// TestUpdateRole_DuplicateName_ReturnsConflict verifies uniqueness check excluding self
func TestUpdateRole_DuplicateName_ReturnsConflict(t *testing.T) {
	mockRepo := &mockRoleRepository{
		findByIDFn: func(id uint) (*models.Role, error) {
			return &models.Role{ID: 1, Name: "OldName", IsSystem: false}, nil
		},
		findByNameExcludingFn: func(name string, excludeID uint) (*models.Role, error) {
			return &models.Role{ID: 2, Name: "Manager"}, nil
		},
	}

	service := NewRoleService(mockRepo)
	input := RoleInput{Name: "Manager"}

	_, err := service.UpdateRole(1, input)

	require.NotNil(t, err)
	assert.Equal(t, ErrConflict, err.Err)
}

// TestDeleteRole_Valid_Succeeds verifies role deletion
func TestDeleteRole_Valid_Succeeds(t *testing.T) {
	mockRepo := &mockRoleRepository{
		findByIDFn: func(id uint) (*models.Role, error) {
			return &models.Role{ID: 1, Name: "ToDelete", IsSystem: false}, nil
		},
		deleteFn: func(id uint) error {
			return nil
		},
	}

	service := NewRoleService(mockRepo)
	err := service.DeleteRole(1)

	require.Nil(t, err)
}

// TestDeleteRole_NotFound_ReturnsNotFoundError verifies role existence check
func TestDeleteRole_NotFound_ReturnsNotFoundError(t *testing.T) {
	mockRepo := &mockRoleRepository{
		findByIDFn: func(id uint) (*models.Role, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	service := NewRoleService(mockRepo)
	err := service.DeleteRole(99999)

	require.NotNil(t, err)
	assert.Equal(t, ErrNotFound, err.Err)
}

// TestDeleteRole_SystemRole_ReturnsForbidden verifies system role protection
func TestDeleteRole_SystemRole_ReturnsForbidden(t *testing.T) {
	mockRepo := &mockRoleRepository{
		findByIDFn: func(id uint) (*models.Role, error) {
			return &models.Role{ID: 1, Name: "Super Admin", IsSystem: true}, nil
		},
	}

	service := NewRoleService(mockRepo)
	err := service.DeleteRole(1)

	require.NotNil(t, err)
	assert.Equal(t, ErrForbidden, err.Err)
	assert.Equal(t, "SYSTEM_ROLE_PROTECTED", err.Code)
	assert.Contains(t, err.Message, "System roles cannot be deleted")
}

// TestDeleteRole_Regular_CleansUpRelations verifies cascade delete behavior
func TestDeleteRole_Regular_CleansUpRelations(t *testing.T) {
	deleteCallCount := 0
	mockRepo := &mockRoleRepository{
		findByIDFn: func(id uint) (*models.Role, error) {
			return &models.Role{ID: 1, Name: "Regular", IsSystem: false}, nil
		},
		deleteFn: func(id uint) error {
			deleteCallCount++
			return nil
		},
	}

	service := NewRoleService(mockRepo)
	err := service.DeleteRole(1)

	require.Nil(t, err)
	assert.Equal(t, 1, deleteCallCount, "Delete should be called once")
}
