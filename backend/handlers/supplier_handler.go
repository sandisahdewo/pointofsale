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

// SupplierHandler handles supplier-related HTTP requests
type SupplierHandler struct {
	supplierService *services.SupplierService
}

// NewSupplierHandler creates a new supplier handler instance
func NewSupplierHandler(supplierService *services.SupplierService) *SupplierHandler {
	return &SupplierHandler{supplierService: supplierService}
}

// ListSuppliers handles GET /api/v1/suppliers
func (h *SupplierHandler) ListSuppliers(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	allowedSortFields := []string{"id", "name", "active"}
	params, err := utils.ParsePaginationParams(r, allowedSortFields)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	// Parse active filter
	var active *bool
	if activeStr := r.URL.Query().Get("active"); activeStr != "" {
		val, err := strconv.ParseBool(activeStr)
		if err == nil {
			active = &val
		}
	}

	// Build repo pagination params
	repoParams := repositories.PaginationParams{
		Page:     params.Page,
		PageSize: params.PageSize,
		Search:   params.Search,
		SortBy:   params.SortBy,
		SortDir:  params.SortDir,
	}

	// Call service
	suppliers, total, err := h.supplierService.ListSuppliers(repoParams, active)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to list suppliers", "INTERNAL_ERROR")
		return
	}

	// Build paginated response
	meta := utils.CalculatePaginationMeta(params.Page, params.PageSize, int(total))
	response := utils.PaginatedResponse{
		Data: suppliers,
		Meta: meta,
	}

	utils.JSON(w, http.StatusOK, response)
}

// GetSupplier handles GET /api/v1/suppliers/{id}
func (h *SupplierHandler) GetSupplier(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid supplier ID", "VALIDATION_ERROR")
		return
	}

	supplier, err := h.supplierService.GetSupplier(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to fetch supplier"
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

	utils.Success(w, http.StatusOK, "", supplier)
}

// CreateSupplier handles POST /api/v1/suppliers
func (h *SupplierHandler) CreateSupplier(w http.ResponseWriter, r *http.Request) {
	var input services.CreateSupplierInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	supplier, err := h.supplierService.CreateSupplier(input)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to create supplier"
		code := "INTERNAL_ERROR"

		if serviceErr, ok := err.(*services.ServiceError); ok {
			message = serviceErr.Message
			code = serviceErr.Code
			switch serviceErr.Err {
			case services.ErrValidation:
				status = http.StatusBadRequest
			case services.ErrConflict:
				status = http.StatusConflict
			}
		}
		utils.Error(w, status, message, code)
		return
	}

	utils.Success(w, http.StatusCreated, "Supplier created successfully", supplier)
}

// UpdateSupplier handles PUT /api/v1/suppliers/{id}
func (h *SupplierHandler) UpdateSupplier(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid supplier ID", "VALIDATION_ERROR")
		return
	}

	var input services.UpdateSupplierInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	supplier, err := h.supplierService.UpdateSupplier(uint(id), input)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to update supplier"
		code := "INTERNAL_ERROR"

		if serviceErr, ok := err.(*services.ServiceError); ok {
			message = serviceErr.Message
			code = serviceErr.Code
			switch serviceErr.Err {
			case services.ErrValidation:
				status = http.StatusBadRequest
			case services.ErrNotFound:
				status = http.StatusNotFound
			case services.ErrConflict:
				status = http.StatusConflict
			}
		}
		utils.Error(w, status, message, code)
		return
	}

	utils.Success(w, http.StatusOK, "Supplier updated successfully", supplier)
}

// DeleteSupplier handles DELETE /api/v1/suppliers/{id}
func (h *SupplierHandler) DeleteSupplier(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid supplier ID", "VALIDATION_ERROR")
		return
	}

	err = h.supplierService.DeleteSupplier(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to delete supplier"
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

	utils.Success(w, http.StatusOK, "Supplier deleted successfully", nil)
}
