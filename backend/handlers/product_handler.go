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

// ProductHandler handles product-related HTTP requests.
type ProductHandler struct {
	productService *services.ProductService
}

// NewProductHandler creates a new product handler instance.
func NewProductHandler(productService *services.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

var productSortFields = []string{"id", "name", "category", "status"}

// ListProducts handles GET /api/v1/products.
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	paginationParams, err := utils.ParsePaginationParams(r, productSortFields)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	query := r.URL.Query()
	status := query.Get("status")

	var categoryID uint
	if value := query.Get("categoryId"); value != "" {
		parsed, parseErr := strconv.ParseUint(value, 10, 64)
		if parseErr != nil {
			utils.Error(w, http.StatusBadRequest, "Invalid categoryId", "VALIDATION_ERROR")
			return
		}
		categoryID = uint(parsed)
	}

	var supplierID uint
	if value := query.Get("supplierId"); value != "" {
		parsed, parseErr := strconv.ParseUint(value, 10, 64)
		if parseErr != nil {
			utils.Error(w, http.StatusBadRequest, "Invalid supplierId", "VALIDATION_ERROR")
			return
		}
		supplierID = uint(parsed)
	}

	params := repositories.ProductListParams{
		PaginationParams: repositories.PaginationParams{
			Page:     paginationParams.Page,
			PageSize: paginationParams.PageSize,
			Search:   paginationParams.Search,
			SortBy:   paginationParams.SortBy,
			SortDir:  paginationParams.SortDir,
		},
		Status:     status,
		CategoryID: categoryID,
		SupplierID: supplierID,
	}

	products, total, serviceErr := h.productService.ListProducts(params)
	if serviceErr != nil {
		utils.Error(w, http.StatusInternalServerError, serviceErr.Message, serviceErr.Code)
		return
	}

	meta := utils.CalculatePaginationMeta(params.Page, params.PageSize, int(total))
	utils.JSON(w, http.StatusOK, utils.PaginatedResponse{
		Data: products,
		Meta: meta,
	})
}

// GetProduct handles GET /api/v1/products/{id}.
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid product ID", "VALIDATION_ERROR")
		return
	}

	product, serviceErr := h.productService.GetProduct(uint(id))
	if serviceErr != nil {
		status := http.StatusInternalServerError
		if serviceErr.Err == services.ErrNotFound {
			status = http.StatusNotFound
		}
		utils.Error(w, status, serviceErr.Message, serviceErr.Code)
		return
	}

	utils.Success(w, http.StatusOK, "", product)
}

// CreateProduct handles POST /api/v1/products.
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var input services.CreateProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	product, serviceErr := h.productService.CreateProduct(input)
	if serviceErr != nil {
		utils.Error(w, mapProductServiceErrorStatus(serviceErr), serviceErr.Message, serviceErr.Code)
		return
	}

	utils.Success(w, http.StatusCreated, "Product created successfully", product)
}

// UpdateProduct handles PUT /api/v1/products/{id}.
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid product ID", "VALIDATION_ERROR")
		return
	}

	var input services.UpdateProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	product, serviceErr := h.productService.UpdateProduct(uint(id), input)
	if serviceErr != nil {
		utils.Error(w, mapProductServiceErrorStatus(serviceErr), serviceErr.Message, serviceErr.Code)
		return
	}

	utils.Success(w, http.StatusOK, "Product updated successfully", product)
}

// DeleteProduct handles DELETE /api/v1/products/{id}.
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid product ID", "VALIDATION_ERROR")
		return
	}

	serviceErr := h.productService.DeleteProduct(uint(id))
	if serviceErr != nil {
		utils.Error(w, mapProductServiceErrorStatus(serviceErr), serviceErr.Message, serviceErr.Code)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Product deleted successfully",
	})
}

func mapProductServiceErrorStatus(serviceErr *services.ServiceError) int {
	switch serviceErr.Err {
	case services.ErrValidation:
		return http.StatusBadRequest
	case services.ErrNotFound:
		return http.StatusNotFound
	case services.ErrConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
