package repositories

import (
	"testing"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCategory_Valid_Succeeds(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewCategoryRepository(db)

	category := &models.Category{
		Name:        "Electronics",
		Description: "Electronic devices and accessories",
	}

	err := repo.Create(category)
	require.NoError(t, err)
	assert.NotZero(t, category.ID, "category ID should be set after creation")
	assert.Equal(t, "Electronics", category.Name)
	assert.Equal(t, "Electronic devices and accessories", category.Description)
	assert.NotZero(t, category.CreatedAt)
	assert.NotZero(t, category.UpdatedAt)
}

func TestListCategories_Pagination_Works(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewCategoryRepository(db)

	// Create 15 test categories
	for i := 1; i <= 15; i++ {
		cat := &models.Category{
			Name:        "Category " + string(rune('A'-1+i)),
			Description: "Description " + string(rune('A'-1+i)),
		}
		require.NoError(t, db.Create(cat).Error)
	}

	// Page 1, pageSize 10
	categories, total, err := repo.List(PaginationParams{
		Page:     1,
		PageSize: 10,
		SortBy:   "id",
		SortDir:  "asc",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, categories, 10)

	// Page 2, pageSize 10
	categories, total, err = repo.List(PaginationParams{
		Page:     2,
		PageSize: 10,
		SortBy:   "id",
		SortDir:  "asc",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, categories, 5)
}

func TestListCategories_Search_FiltersByNameAndDescription(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewCategoryRepository(db)

	// Create test categories
	categories := []*models.Category{
		{Name: "Electronics", Description: "Gadgets and devices"},
		{Name: "Clothing", Description: "Apparel and garments"},
		{Name: "Food", Description: "Electronics equipment for kitchen"}, // description matches "electronics"
	}
	for _, cat := range categories {
		require.NoError(t, db.Create(cat).Error)
	}

	// Search by name (also matches description)
	results, total, err := repo.List(PaginationParams{
		Page:     1,
		PageSize: 10,
		Search:   "electronics",
		SortBy:   "id",
		SortDir:  "asc",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total) // "Electronics" name + "Food" with "Electronics" in description
	assert.Len(t, results, 2)

	// Search by description keyword
	results, total, err = repo.List(PaginationParams{
		Page:     1,
		PageSize: 10,
		Search:   "apparel",
		SortBy:   "id",
		SortDir:  "asc",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Clothing", results[0].Name)
}

func TestListCategories_Sort_OrdersCorrectly(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewCategoryRepository(db)

	// Create categories
	cats := []*models.Category{
		{Name: "Zebra"},
		{Name: "Apple"},
		{Name: "Mango"},
	}
	for _, cat := range cats {
		require.NoError(t, db.Create(cat).Error)
	}

	// Sort by name ascending
	results, _, err := repo.List(PaginationParams{
		Page:     1,
		PageSize: 10,
		SortBy:   "name",
		SortDir:  "asc",
	})
	require.NoError(t, err)
	assert.Equal(t, "Apple", results[0].Name)
	assert.Equal(t, "Mango", results[1].Name)
	assert.Equal(t, "Zebra", results[2].Name)

	// Sort by name descending
	results, _, err = repo.List(PaginationParams{
		Page:     1,
		PageSize: 10,
		SortBy:   "name",
		SortDir:  "desc",
	})
	require.NoError(t, err)
	assert.Equal(t, "Zebra", results[0].Name)
	assert.Equal(t, "Mango", results[1].Name)
	assert.Equal(t, "Apple", results[2].Name)
}

func TestGetCategory_Exists_ReturnsCategory(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewCategoryRepository(db)

	// Create a category
	cat := &models.Category{
		Name:        "Test Category",
		Description: "Test description",
	}
	require.NoError(t, db.Create(cat).Error)

	// Get by ID
	found, err := repo.GetByID(cat.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, cat.ID, found.ID)
	assert.Equal(t, "Test Category", found.Name)
	assert.Equal(t, "Test description", found.Description)
}

func TestGetCategory_NotFound_ReturnsNil(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewCategoryRepository(db)

	// Try to get a non-existent category
	found, err := repo.GetByID(99999)
	require.Error(t, err)
	assert.Nil(t, found)
}

func TestUpdateCategory_Valid_UpdatesFields(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewCategoryRepository(db)

	// Create a category
	cat := &models.Category{
		Name:        "Original Name",
		Description: "Original description",
	}
	require.NoError(t, db.Create(cat).Error)

	// Update fields
	cat.Name = "Updated Name"
	cat.Description = "Updated description"
	err := repo.Update(cat)
	require.NoError(t, err)

	// Verify by re-fetching
	updated, err := repo.GetByID(cat.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Updated description", updated.Description)
}

func TestDeleteCategory_NoReferences_Succeeds(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewCategoryRepository(db)

	// Create a category
	cat := &models.Category{
		Name:        "To Delete",
		Description: "Will be deleted",
	}
	require.NoError(t, db.Create(cat).Error)

	// Delete
	err := repo.Delete(cat.ID)
	require.NoError(t, err)

	// Verify it's gone
	found, err := repo.GetByID(cat.ID)
	require.Error(t, err)
	assert.Nil(t, found)
}

func TestCountProductsByCategory_NoProducts_ReturnsZero(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := NewCategoryRepository(db)

	// Create a category
	cat := &models.Category{Name: "Empty Category"}
	require.NoError(t, db.Create(cat).Error)

	count, err := repo.CountProductsByCategory(cat.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
