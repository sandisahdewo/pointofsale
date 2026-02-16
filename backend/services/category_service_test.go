package services

import (
	"testing"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Mock CategoryRepository for service tests
type mockCategoryRepo struct {
	createFn               func(*models.Category) error
	listFn                 func(repositories.PaginationParams) ([]models.Category, int64, error)
	getByIDFn              func(uint) (*models.Category, error)
	updateFn               func(*models.Category) error
	deleteFn               func(uint) error
	countProductsByCatFn   func(uint) (int64, error)
}

func (m *mockCategoryRepo) Create(category *models.Category) error {
	if m.createFn != nil {
		return m.createFn(category)
	}
	return nil
}

func (m *mockCategoryRepo) List(params repositories.PaginationParams) ([]models.Category, int64, error) {
	if m.listFn != nil {
		return m.listFn(params)
	}
	return []models.Category{}, 0, nil
}

func (m *mockCategoryRepo) GetByID(id uint) (*models.Category, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockCategoryRepo) Update(category *models.Category) error {
	if m.updateFn != nil {
		return m.updateFn(category)
	}
	return nil
}

func (m *mockCategoryRepo) Delete(id uint) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return nil
}

func (m *mockCategoryRepo) CountProductsByCategory(categoryID uint) (int64, error) {
	if m.countProductsByCatFn != nil {
		return m.countProductsByCatFn(categoryID)
	}
	return 0, nil
}

func TestCategoryService_CreateCategory_Valid_Succeeds(t *testing.T) {
	repo := &mockCategoryRepo{
		createFn: func(c *models.Category) error {
			c.ID = 1
			return nil
		},
	}

	svc := NewCategoryService(repo)
	input := CreateCategoryInput{
		Name:        "Electronics",
		Description: "Electronic devices",
	}

	category, err := svc.CreateCategory(input)
	require.NoError(t, err)
	assert.NotNil(t, category)
	assert.Equal(t, "Electronics", category.Name)
	assert.Equal(t, "Electronic devices", category.Description)
}

func TestCategoryService_CreateCategory_MissingName_ReturnsValidation(t *testing.T) {
	repo := &mockCategoryRepo{}
	svc := NewCategoryService(repo)

	input := CreateCategoryInput{
		Name: "",
	}

	category, err := svc.CreateCategory(input)
	assert.Nil(t, category)
	require.Error(t, err)

	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrValidation, serviceErr.Err)
}

func TestCategoryService_CreateCategory_NameTooLong_ReturnsValidation(t *testing.T) {
	repo := &mockCategoryRepo{}
	svc := NewCategoryService(repo)

	longName := make([]byte, 256)
	for i := range longName {
		longName[i] = 'a'
	}

	input := CreateCategoryInput{
		Name: string(longName),
	}

	category, err := svc.CreateCategory(input)
	assert.Nil(t, category)
	require.Error(t, err)

	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrValidation, serviceErr.Err)
}

func TestCategoryService_DeleteCategory_ReferencedByProducts_ReturnsConflict(t *testing.T) {
	repo := &mockCategoryRepo{
		getByIDFn: func(id uint) (*models.Category, error) {
			return &models.Category{ID: id, Name: "Used Category"}, nil
		},
		countProductsByCatFn: func(categoryID uint) (int64, error) {
			return 3, nil // 3 products reference this category
		},
	}

	svc := NewCategoryService(repo)

	err := svc.DeleteCategory(1)
	require.Error(t, err)

	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrConflict, serviceErr.Err)
	assert.Contains(t, serviceErr.Message, "3 product(s)")
}

func TestCategoryService_DeleteCategory_Unreferenced_Succeeds(t *testing.T) {
	deleteCalled := false
	repo := &mockCategoryRepo{
		getByIDFn: func(id uint) (*models.Category, error) {
			return &models.Category{ID: id, Name: "Unused Category"}, nil
		},
		countProductsByCatFn: func(categoryID uint) (int64, error) {
			return 0, nil
		},
		deleteFn: func(id uint) error {
			deleteCalled = true
			return nil
		},
	}

	svc := NewCategoryService(repo)

	err := svc.DeleteCategory(1)
	require.NoError(t, err)
	assert.True(t, deleteCalled)
}

func TestCategoryService_DeleteCategory_NotFound_ReturnsNotFound(t *testing.T) {
	repo := &mockCategoryRepo{
		getByIDFn: func(id uint) (*models.Category, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	svc := NewCategoryService(repo)

	err := svc.DeleteCategory(999)
	require.Error(t, err)

	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrNotFound, serviceErr.Err)
}

func TestCategoryService_GetCategory_NotFound_ReturnsNotFound(t *testing.T) {
	repo := &mockCategoryRepo{
		getByIDFn: func(id uint) (*models.Category, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	svc := NewCategoryService(repo)

	category, err := svc.GetCategory(999)
	assert.Nil(t, category)
	require.Error(t, err)

	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrNotFound, serviceErr.Err)
}

func TestCategoryService_UpdateCategory_NotFound_ReturnsNotFound(t *testing.T) {
	repo := &mockCategoryRepo{
		getByIDFn: func(id uint) (*models.Category, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	svc := NewCategoryService(repo)

	input := UpdateCategoryInput{
		Name: "Updated",
	}

	category, err := svc.UpdateCategory(999, input)
	assert.Nil(t, category)
	require.Error(t, err)

	serviceErr, ok := err.(*ServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrNotFound, serviceErr.Err)
}

func TestCategoryService_UpdateCategory_Valid_Succeeds(t *testing.T) {
	repo := &mockCategoryRepo{
		getByIDFn: func(id uint) (*models.Category, error) {
			return &models.Category{ID: id, Name: "Old Name", Description: "Old desc"}, nil
		},
		updateFn: func(c *models.Category) error {
			return nil
		},
	}

	svc := NewCategoryService(repo)

	input := UpdateCategoryInput{
		Name:        "New Name",
		Description: "New desc",
	}

	category, err := svc.UpdateCategory(1, input)
	require.NoError(t, err)
	assert.Equal(t, "New Name", category.Name)
	assert.Equal(t, "New desc", category.Description)
}
