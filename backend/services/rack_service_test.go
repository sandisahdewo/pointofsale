package services

import (
	"testing"

	"github.com/pointofsale/backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// mockRackRepository is a mock implementation for testing
type mockRackRepository struct {
	listFn              func(page, pageSize int, search, active, sortBy, sortDir string) ([]models.Rack, int64, error)
	findByIDFn          func(id uint) (*models.Rack, error)
	findByCodeFn        func(code string) (*models.Rack, error)
	findByCodeExcludeFn func(code string, excludeID uint) (*models.Rack, error)
	createFn            func(rack *models.Rack) error
	updateFn            func(rack *models.Rack) error
	deleteFn            func(id uint) error
	cleanupVariantsFn   func(rackID uint) error
}

func (m *mockRackRepository) List(page, pageSize int, search, active, sortBy, sortDir string) ([]models.Rack, int64, error) {
	if m.listFn != nil {
		return m.listFn(page, pageSize, search, active, sortBy, sortDir)
	}
	return nil, 0, nil
}

func (m *mockRackRepository) FindByID(id uint) (*models.Rack, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(id)
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockRackRepository) FindByCode(code string) (*models.Rack, error) {
	if m.findByCodeFn != nil {
		return m.findByCodeFn(code)
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockRackRepository) FindByCodeExcluding(code string, excludeID uint) (*models.Rack, error) {
	if m.findByCodeExcludeFn != nil {
		return m.findByCodeExcludeFn(code, excludeID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockRackRepository) Create(rack *models.Rack) error {
	if m.createFn != nil {
		return m.createFn(rack)
	}
	return nil
}

func (m *mockRackRepository) Update(rack *models.Rack) error {
	if m.updateFn != nil {
		return m.updateFn(rack)
	}
	return nil
}

func (m *mockRackRepository) Delete(id uint) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return nil
}

func (m *mockRackRepository) CleanupVariantRacks(rackID uint) error {
	if m.cleanupVariantsFn != nil {
		return m.cleanupVariantsFn(rackID)
	}
	return nil
}

// TestCreateRack_Valid_Succeeds verifies successful rack creation
func TestCreateRackService_Valid_Succeeds(t *testing.T) {
	mockRepo := &mockRackRepository{
		findByCodeFn: func(code string) (*models.Rack, error) {
			return nil, gorm.ErrRecordNotFound
		},
		createFn: func(rack *models.Rack) error {
			rack.ID = 1
			return nil
		},
	}

	service := NewRackService(mockRepo)
	input := RackInput{
		Name:        "Main Display",
		Code:        "R-001",
		Location:    "Store Front",
		Capacity:    100,
		Description: "Primary display shelf",
	}

	rack, err := service.CreateRack(input)

	require.Nil(t, err)
	assert.Equal(t, "Main Display", rack.Name)
	assert.Equal(t, "R-001", rack.Code)
	assert.Equal(t, "Store Front", rack.Location)
	assert.Equal(t, 100, rack.Capacity)
	assert.Equal(t, "Primary display shelf", rack.Description)
	assert.True(t, rack.Active)
}

// TestCreateRack_EmptyName_ReturnsValidationError verifies name is required
func TestCreateRackService_EmptyName_ReturnsValidationError(t *testing.T) {
	mockRepo := &mockRackRepository{}
	service := NewRackService(mockRepo)

	input := RackInput{
		Name:     "",
		Code:     "R-001",
		Location: "Store Front",
		Capacity: 100,
	}

	_, err := service.CreateRack(input)

	require.NotNil(t, err)
	assert.Equal(t, ErrValidation, err.Err)
	assert.Contains(t, err.Message, "Name is required")
}

// TestCreateRack_EmptyCode_ReturnsValidationError verifies code is required
func TestCreateRackService_EmptyCode_ReturnsValidationError(t *testing.T) {
	mockRepo := &mockRackRepository{}
	service := NewRackService(mockRepo)

	input := RackInput{
		Name:     "Rack",
		Code:     "",
		Location: "Store Front",
		Capacity: 100,
	}

	_, err := service.CreateRack(input)

	require.NotNil(t, err)
	assert.Equal(t, ErrValidation, err.Err)
	assert.Contains(t, err.Message, "Code is required")
}

// TestCreateRack_EmptyLocation_ReturnsValidationError verifies location is required
func TestCreateRackService_EmptyLocation_ReturnsValidationError(t *testing.T) {
	mockRepo := &mockRackRepository{}
	service := NewRackService(mockRepo)

	input := RackInput{
		Name:     "Rack",
		Code:     "R-001",
		Location: "",
		Capacity: 100,
	}

	_, err := service.CreateRack(input)

	require.NotNil(t, err)
	assert.Equal(t, ErrValidation, err.Err)
	assert.Contains(t, err.Message, "Location is required")
}

// TestCreateRack_ZeroCapacity_ReturnsValidationError verifies capacity > 0
func TestCreateRackService_ZeroCapacity_ReturnsValidationError(t *testing.T) {
	mockRepo := &mockRackRepository{}
	service := NewRackService(mockRepo)

	input := RackInput{
		Name:     "Rack",
		Code:     "R-001",
		Location: "Store Front",
		Capacity: 0,
	}

	_, err := service.CreateRack(input)

	require.NotNil(t, err)
	assert.Equal(t, ErrValidation, err.Err)
	assert.Contains(t, err.Message, "Capacity must be greater than 0")
}

// TestCreateRack_NegativeCapacity_ReturnsValidationError verifies capacity > 0
func TestCreateRackService_NegativeCapacity_ReturnsValidationError(t *testing.T) {
	mockRepo := &mockRackRepository{}
	service := NewRackService(mockRepo)

	input := RackInput{
		Name:     "Rack",
		Code:     "R-001",
		Location: "Store Front",
		Capacity: -5,
	}

	_, err := service.CreateRack(input)

	require.NotNil(t, err)
	assert.Equal(t, ErrValidation, err.Err)
	assert.Contains(t, err.Message, "Capacity must be greater than 0")
}

// TestCreateRack_DuplicateCode_ReturnsConflict verifies uniqueness check
func TestCreateRackService_DuplicateCode_ReturnsConflict(t *testing.T) {
	mockRepo := &mockRackRepository{
		findByCodeFn: func(code string) (*models.Rack, error) {
			return &models.Rack{ID: 1, Code: "R-001"}, nil
		},
	}

	service := NewRackService(mockRepo)
	input := RackInput{
		Name:     "Rack",
		Code:     "R-001",
		Location: "Store Front",
		Capacity: 100,
	}

	_, err := service.CreateRack(input)

	require.NotNil(t, err)
	assert.Equal(t, ErrConflict, err.Err)
	assert.Equal(t, "RACK_CODE_EXISTS", err.Code)
	assert.Contains(t, err.Message, "already exists")
}

// TestCreateRack_CodeTooLong_ReturnsValidationError verifies code max length
func TestCreateRackService_CodeTooLong_ReturnsValidationError(t *testing.T) {
	mockRepo := &mockRackRepository{}
	service := NewRackService(mockRepo)

	longCode := ""
	for i := 0; i < 51; i++ {
		longCode += "A"
	}

	input := RackInput{
		Name:     "Rack",
		Code:     longCode,
		Location: "Store Front",
		Capacity: 100,
	}

	_, err := service.CreateRack(input)

	require.NotNil(t, err)
	assert.Equal(t, ErrValidation, err.Err)
	assert.Contains(t, err.Message, "50 characters")
}

// TestUpdateRack_Valid_Succeeds verifies successful rack update
func TestUpdateRackService_Valid_Succeeds(t *testing.T) {
	mockRepo := &mockRackRepository{
		findByIDFn: func(id uint) (*models.Rack, error) {
			return &models.Rack{ID: 1, Name: "OldName", Code: "R-001", Location: "Old Loc", Capacity: 50, Active: true}, nil
		},
		findByCodeExcludeFn: func(code string, excludeID uint) (*models.Rack, error) {
			return nil, gorm.ErrRecordNotFound
		},
		updateFn: func(rack *models.Rack) error {
			return nil
		},
	}

	service := NewRackService(mockRepo)
	input := RackInput{
		Name:     "NewName",
		Code:     "R-001",
		Location: "New Loc",
		Capacity: 100,
		Active:   boolPtr(true),
	}

	rack, err := service.UpdateRack(1, input)

	require.Nil(t, err)
	assert.Equal(t, "NewName", rack.Name)
	assert.Equal(t, "New Loc", rack.Location)
	assert.Equal(t, 100, rack.Capacity)
}

// TestUpdateRack_NotFound_ReturnsNotFoundError verifies rack existence check
func TestUpdateRackService_NotFound_ReturnsNotFoundError(t *testing.T) {
	mockRepo := &mockRackRepository{
		findByIDFn: func(id uint) (*models.Rack, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	service := NewRackService(mockRepo)
	input := RackInput{Name: "Test", Code: "R-001", Location: "Loc", Capacity: 50}

	_, err := service.UpdateRack(99999, input)

	require.NotNil(t, err)
	assert.Equal(t, ErrNotFound, err.Err)
}

// TestUpdateRack_CodeUniqueExcludesSelf verifies code uniqueness excluding self
func TestUpdateRackService_CodeUniqueExcludesSelf(t *testing.T) {
	mockRepo := &mockRackRepository{
		findByIDFn: func(id uint) (*models.Rack, error) {
			return &models.Rack{ID: 1, Name: "Rack 1", Code: "R-001", Location: "Loc", Capacity: 50, Active: true}, nil
		},
		findByCodeExcludeFn: func(code string, excludeID uint) (*models.Rack, error) {
			// Simulate another rack having the same code
			return &models.Rack{ID: 2, Code: "R-002"}, nil
		},
	}

	service := NewRackService(mockRepo)
	input := RackInput{
		Name:     "Rack 1",
		Code:     "R-002",
		Location: "Loc",
		Capacity: 50,
		Active:   boolPtr(true),
	}

	_, err := service.UpdateRack(1, input)

	require.NotNil(t, err)
	assert.Equal(t, ErrConflict, err.Err)
	assert.Equal(t, "RACK_CODE_EXISTS", err.Code)
}

// TestUpdateRack_SameCodeSelf_Succeeds verifies updating with own code works
func TestUpdateRackService_SameCodeSelf_Succeeds(t *testing.T) {
	mockRepo := &mockRackRepository{
		findByIDFn: func(id uint) (*models.Rack, error) {
			return &models.Rack{ID: 1, Name: "Rack 1", Code: "R-001", Location: "Loc", Capacity: 50, Active: true}, nil
		},
		findByCodeExcludeFn: func(code string, excludeID uint) (*models.Rack, error) {
			return nil, gorm.ErrRecordNotFound
		},
		updateFn: func(rack *models.Rack) error {
			return nil
		},
	}

	service := NewRackService(mockRepo)
	input := RackInput{
		Name:     "Updated Rack",
		Code:     "R-001",
		Location: "New Loc",
		Capacity: 100,
		Active:   boolPtr(true),
	}

	rack, err := service.UpdateRack(1, input)

	require.Nil(t, err)
	assert.Equal(t, "Updated Rack", rack.Name)
}

// TestGetRack_Exists_ReturnsRack verifies getting rack by ID
func TestGetRackService_Exists_ReturnsRack(t *testing.T) {
	mockRepo := &mockRackRepository{
		findByIDFn: func(id uint) (*models.Rack, error) {
			return &models.Rack{ID: 1, Name: "Main Display", Code: "R-001"}, nil
		},
	}

	service := NewRackService(mockRepo)
	rack, err := service.GetRack(1)

	require.Nil(t, err)
	assert.Equal(t, "Main Display", rack.Name)
}

// TestGetRack_NotFound_ReturnsNotFoundError verifies 404 error
func TestGetRackService_NotFound_ReturnsNotFoundError(t *testing.T) {
	mockRepo := &mockRackRepository{
		findByIDFn: func(id uint) (*models.Rack, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	service := NewRackService(mockRepo)
	_, err := service.GetRack(99999)

	require.NotNil(t, err)
	assert.Equal(t, ErrNotFound, err.Err)
	assert.Equal(t, "RACK_NOT_FOUND", err.Code)
}

// TestDeleteRack_Valid_Succeeds verifies rack deletion
func TestDeleteRackService_Valid_Succeeds(t *testing.T) {
	mockRepo := &mockRackRepository{
		findByIDFn: func(id uint) (*models.Rack, error) {
			return &models.Rack{ID: 1, Name: "ToDelete"}, nil
		},
		deleteFn: func(id uint) error {
			return nil
		},
		cleanupVariantsFn: func(rackID uint) error {
			return nil
		},
	}

	service := NewRackService(mockRepo)
	err := service.DeleteRack(1)

	require.Nil(t, err)
}

// TestDeleteRack_NotFound_ReturnsNotFoundError verifies rack existence check on delete
func TestDeleteRackService_NotFound_ReturnsNotFoundError(t *testing.T) {
	mockRepo := &mockRackRepository{
		findByIDFn: func(id uint) (*models.Rack, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	service := NewRackService(mockRepo)
	err := service.DeleteRack(99999)

	require.NotNil(t, err)
	assert.Equal(t, ErrNotFound, err.Err)
}

// TestDeleteRack_CleansUpVariantRacks verifies variant_racks cleanup on delete
func TestDeleteRackService_CleansUpVariantRacks(t *testing.T) {
	cleanupCalled := false
	deleteCalled := false

	mockRepo := &mockRackRepository{
		findByIDFn: func(id uint) (*models.Rack, error) {
			return &models.Rack{ID: 1, Name: "ToDelete"}, nil
		},
		cleanupVariantsFn: func(rackID uint) error {
			cleanupCalled = true
			assert.Equal(t, uint(1), rackID)
			return nil
		},
		deleteFn: func(id uint) error {
			deleteCalled = true
			return nil
		},
	}

	service := NewRackService(mockRepo)
	err := service.DeleteRack(1)

	require.Nil(t, err)
	assert.True(t, cleanupCalled, "CleanupVariantRacks should be called")
	assert.True(t, deleteCalled, "Delete should be called")
}

// helper
func boolPtr(b bool) *bool {
	return &b
}
