package repositories

import (
	"github.com/pointofsale/backend/models"
	"gorm.io/gorm"
)

// CategoryRepository defines the interface for category data operations
type CategoryRepository interface {
	Create(category *models.Category) error
	List(params PaginationParams) ([]models.Category, int64, error)
	GetByID(id uint) (*models.Category, error)
	Update(category *models.Category) error
	Delete(id uint) error
	CountProductsByCategory(categoryID uint) (int64, error)
}

// CategoryRepositoryImpl implements CategoryRepository interface
type CategoryRepositoryImpl struct {
	db *gorm.DB
}

// NewCategoryRepository creates a new category repository instance
func NewCategoryRepository(db *gorm.DB) *CategoryRepositoryImpl {
	return &CategoryRepositoryImpl{db: db}
}

// Create creates a new category in the database
func (r *CategoryRepositoryImpl) Create(category *models.Category) error {
	return r.db.Create(category).Error
}

// List returns paginated categories with optional search and sorting
func (r *CategoryRepositoryImpl) List(params PaginationParams) ([]models.Category, int64, error) {
	var categories []models.Category
	var total int64

	query := r.db.Model(&models.Category{})

	// Apply search filter (name OR description, case-insensitive)
	if params.Search != "" {
		searchPattern := "%" + params.Search + "%"
		query = query.Where("LOWER(name) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?)", searchPattern, searchPattern)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderClause := params.SortBy + " " + params.SortDir
	query = query.Order(orderClause)

	// Apply pagination
	offset := (params.Page - 1) * params.PageSize
	query = query.Offset(offset).Limit(params.PageSize)

	// Execute query
	if err := query.Find(&categories).Error; err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

// GetByID finds a category by ID
func (r *CategoryRepositoryImpl) GetByID(id uint) (*models.Category, error) {
	var category models.Category
	err := r.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// Update saves changes to an existing category
func (r *CategoryRepositoryImpl) Update(category *models.Category) error {
	return r.db.Save(category).Error
}

// Delete removes a category from the database
func (r *CategoryRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&models.Category{}, id).Error
}

// CountProductsByCategory counts how many products reference a specific category.
// This is used to check if a category can be safely deleted.
// Note: Products table may not exist yet; if so, return 0.
func (r *CategoryRepositoryImpl) CountProductsByCategory(categoryID uint) (int64, error) {
	var count int64
	// Check if products table exists
	if r.db.Migrator().HasTable("products") {
		err := r.db.Table("products").Where("category_id = ?", categoryID).Count(&count).Error
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}
