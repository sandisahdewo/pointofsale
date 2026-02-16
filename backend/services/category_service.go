package services

import (
	"fmt"
	"strings"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/repositories"
	"gorm.io/gorm"
)

// CategoryRepositoryInterface defines the repository interface needed by CategoryService
type CategoryRepositoryInterface interface {
	Create(category *models.Category) error
	List(params repositories.PaginationParams) ([]models.Category, int64, error)
	GetByID(id uint) (*models.Category, error)
	Update(category *models.Category) error
	Delete(id uint) error
	CountProductsByCategory(categoryID uint) (int64, error)
}

// CategoryService handles category business logic
type CategoryService struct {
	repo CategoryRepositoryInterface
}

// NewCategoryService creates a new category service instance
func NewCategoryService(repo CategoryRepositoryInterface) *CategoryService {
	return &CategoryService{repo: repo}
}

// CreateCategoryInput represents the input for creating a category
type CreateCategoryInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateCategoryInput represents the input for updating a category
type UpdateCategoryInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ListCategories returns paginated categories
func (s *CategoryService) ListCategories(params repositories.PaginationParams) ([]models.Category, int64, error) {
	return s.repo.List(params)
}

// GetCategory returns a single category by ID
func (s *CategoryService) GetCategory(id uint) (*models.Category, error) {
	category, err := s.repo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &ServiceError{
				Err:     ErrNotFound,
				Message: "Category not found",
				Code:    "CATEGORY_NOT_FOUND",
			}
		}
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to fetch category",
			Code:    "INTERNAL_ERROR",
		}
	}
	return category, nil
}

// CreateCategory creates a new category
func (s *CategoryService) CreateCategory(input CreateCategoryInput) (*models.Category, error) {
	// Validate name
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Name is required",
			Code:    "VALIDATION_ERROR",
		}
	}
	if len(name) > 255 {
		return nil, &ServiceError{
			Err:     ErrValidation,
			Message: "Name must be between 1 and 255 characters",
			Code:    "VALIDATION_ERROR",
		}
	}

	category := &models.Category{
		Name:        name,
		Description: input.Description,
	}

	if err := s.repo.Create(category); err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to create category",
			Code:    "INTERNAL_ERROR",
		}
	}

	return category, nil
}

// UpdateCategory updates an existing category
func (s *CategoryService) UpdateCategory(id uint, input UpdateCategoryInput) (*models.Category, error) {
	category, err := s.repo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &ServiceError{
				Err:     ErrNotFound,
				Message: "Category not found",
				Code:    "CATEGORY_NOT_FOUND",
			}
		}
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to fetch category",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Validate and update name
	name := strings.TrimSpace(input.Name)
	if name != "" {
		if len(name) > 255 {
			return nil, &ServiceError{
				Err:     ErrValidation,
				Message: "Name must be between 1 and 255 characters",
				Code:    "VALIDATION_ERROR",
			}
		}
		category.Name = name
	}

	// Update description (allow empty to clear it)
	category.Description = input.Description

	if err := s.repo.Update(category); err != nil {
		return nil, &ServiceError{
			Err:     err,
			Message: "Failed to update category",
			Code:    "INTERNAL_ERROR",
		}
	}

	return category, nil
}

// DeleteCategory deletes a category, blocking if referenced by products
func (s *CategoryService) DeleteCategory(id uint) error {
	// Check if category exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &ServiceError{
				Err:     ErrNotFound,
				Message: "Category not found",
				Code:    "CATEGORY_NOT_FOUND",
			}
		}
		return &ServiceError{
			Err:     err,
			Message: "Failed to fetch category",
			Code:    "INTERNAL_ERROR",
		}
	}

	// Check if category is referenced by products
	count, err := s.repo.CountProductsByCategory(id)
	if err != nil {
		return &ServiceError{
			Err:     err,
			Message: "Failed to check product references",
			Code:    "INTERNAL_ERROR",
		}
	}
	if count > 0 {
		return &ServiceError{
			Err:     ErrConflict,
			Message: fmt.Sprintf("Cannot delete category. It is referenced by %d product(s).", count),
			Code:    "CATEGORY_IN_USE",
		}
	}

	if err := s.repo.Delete(id); err != nil {
		return &ServiceError{
			Err:     err,
			Message: "Failed to delete category",
			Code:    "INTERNAL_ERROR",
		}
	}

	return nil
}
