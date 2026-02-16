package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pointofsale/backend/repositories"
	"github.com/pointofsale/backend/services"
	"github.com/pointofsale/backend/utils"
)

// CategoryHandler handles HTTP requests for category endpoints
type CategoryHandler struct {
	categoryService *services.CategoryService
}

// NewCategoryHandler creates a new category handler instance
func NewCategoryHandler(categoryService *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// Allowed sort fields for categories (prevents SQL injection)
var categorySortFields = []string{"id", "name", "description"}

// ListCategories handles GET /api/v1/categories
func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	paginationParams, err := utils.ParsePaginationParams(r, categorySortFields)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	// Convert utils.PaginationParams to repositories.PaginationParams
	params := repositories.PaginationParams{
		Page:     paginationParams.Page,
		PageSize: paginationParams.PageSize,
		Search:   paginationParams.Search,
		SortBy:   paginationParams.SortBy,
		SortDir:  paginationParams.SortDir,
	}

	categories, total, err := h.categoryService.ListCategories(params)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch categories", "INTERNAL_ERROR")
		return
	}

	meta := utils.CalculatePaginationMeta(params.Page, params.PageSize, int(total))

	utils.JSON(w, http.StatusOK, utils.PaginatedResponse{
		Data: categories,
		Meta: meta,
	})
}

// GetCategory handles GET /api/v1/categories/{id}
func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid category ID", "VALIDATION_ERROR")
		return
	}

	category, err := h.categoryService.GetCategory(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to fetch category"
		code := "INTERNAL_ERROR"

		if serviceErr, ok := err.(*services.ServiceError); ok {
			message = serviceErr.Message
			code = serviceErr.Code
			if serviceErr.Err == services.ErrNotFound {
				status = http.StatusNotFound
			}
		}
		utils.Error(w, status, message, code)
		return
	}

	utils.Success(w, http.StatusOK, "", category)
}

// CreateCategory handles POST /api/v1/categories
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var input services.CreateCategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	category, err := h.categoryService.CreateCategory(input)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to create category"
		code := "INTERNAL_ERROR"

		if serviceErr, ok := err.(*services.ServiceError); ok {
			message = serviceErr.Message
			code = serviceErr.Code
			if serviceErr.Err == services.ErrValidation {
				status = http.StatusBadRequest
			}
		}
		utils.Error(w, status, message, code)
		return
	}

	utils.Success(w, http.StatusCreated, "Category created successfully", category)
}

// UpdateCategory handles PUT /api/v1/categories/{id}
func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid category ID", "VALIDATION_ERROR")
		return
	}

	var input services.UpdateCategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	category, err := h.categoryService.UpdateCategory(uint(id), input)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to update category"
		code := "INTERNAL_ERROR"

		if serviceErr, ok := err.(*services.ServiceError); ok {
			message = serviceErr.Message
			code = serviceErr.Code
			switch serviceErr.Err {
			case services.ErrValidation:
				status = http.StatusBadRequest
			case services.ErrNotFound:
				status = http.StatusNotFound
			}
		}
		utils.Error(w, status, message, code)
		return
	}

	utils.Success(w, http.StatusOK, "Category updated successfully", category)
}

// DeleteCategory handles DELETE /api/v1/categories/{id}
func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid category ID", "VALIDATION_ERROR")
		return
	}

	err = h.categoryService.DeleteCategory(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to delete category"
		code := "INTERNAL_ERROR"

		if serviceErr, ok := err.(*services.ServiceError); ok {
			message = serviceErr.Message
			code = serviceErr.Code
			switch serviceErr.Err {
			case services.ErrNotFound:
				status = http.StatusNotFound
			case services.ErrConflict:
				status = http.StatusConflict
			}
		}
		utils.Error(w, status, message, code)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Category deleted successfully",
	})
}
