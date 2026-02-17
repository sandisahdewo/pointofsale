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

// POHandler handles HTTP requests for purchase order endpoints
type POHandler struct {
	poService *services.POService
}

// NewPOHandler creates a new PO handler instance
func NewPOHandler(poService *services.POService) *POHandler {
	return &POHandler{poService: poService}
}

// Allowed sort fields for POs (prevents SQL injection)
var poSortFields = []string{"date", "po_number", "status"}

// ListPOs handles GET /api/v1/purchase-orders
func (h *POHandler) ListPOs(w http.ResponseWriter, r *http.Request) {
	paginationParams, err := utils.ParsePaginationParams(r, poSortFields)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	params := repositories.PaginationParams{
		Page:     paginationParams.Page,
		PageSize: paginationParams.PageSize,
		Search:   paginationParams.Search,
		SortBy:   paginationParams.SortBy,
		SortDir:  paginationParams.SortDir,
	}

	// Parse optional filters
	status := r.URL.Query().Get("status")
	var supplierID uint
	if s := r.URL.Query().Get("supplierId"); s != "" {
		if id, err := strconv.ParseUint(s, 10, 64); err == nil {
			supplierID = uint(id)
		}
	}

	pos, total, statusCounts, err := h.poService.ListPOs(params, status, supplierID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch purchase orders", "INTERNAL_ERROR")
		return
	}

	meta := utils.CalculatePaginationMeta(params.Page, params.PageSize, int(total))

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"data":         pos,
		"meta":         meta,
		"statusCounts": statusCounts,
	})
}

// GetPO handles GET /api/v1/purchase-orders/{id}
func (h *POHandler) GetPO(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid purchase order ID", "VALIDATION_ERROR")
		return
	}

	po, err := h.poService.GetPO(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to fetch purchase order"
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

	utils.Success(w, http.StatusOK, "", po)
}

// CreatePO handles POST /api/v1/purchase-orders
func (h *POHandler) CreatePO(w http.ResponseWriter, r *http.Request) {
	var input services.CreatePOInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	po, err := h.poService.CreatePO(input)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to create purchase order"
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

	utils.Success(w, http.StatusCreated, "Purchase order created successfully", po)
}

// UpdatePO handles PUT /api/v1/purchase-orders/{id}
func (h *POHandler) UpdatePO(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid purchase order ID", "VALIDATION_ERROR")
		return
	}

	var input services.CreatePOInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	po, err := h.poService.UpdatePO(uint(id), input)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to update purchase order"
		code := "INTERNAL_ERROR"
		if serviceErr, ok := err.(*services.ServiceError); ok {
			message = serviceErr.Message
			code = serviceErr.Code
			switch serviceErr.Err {
			case services.ErrValidation:
				status = http.StatusBadRequest
			case services.ErrNotFound:
				status = http.StatusNotFound
			case services.ErrForbidden:
				status = http.StatusForbidden
			}
		}
		utils.Error(w, status, message, code)
		return
	}

	utils.Success(w, http.StatusOK, "Purchase order updated successfully", po)
}

// DeletePO handles DELETE /api/v1/purchase-orders/{id}
func (h *POHandler) DeletePO(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid purchase order ID", "VALIDATION_ERROR")
		return
	}

	err = h.poService.DeletePO(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to delete purchase order"
		code := "INTERNAL_ERROR"
		if serviceErr, ok := err.(*services.ServiceError); ok {
			message = serviceErr.Message
			code = serviceErr.Code
			switch serviceErr.Err {
			case services.ErrNotFound:
				status = http.StatusNotFound
			case services.ErrForbidden:
				status = http.StatusForbidden
			}
		}
		utils.Error(w, status, message, code)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message": "Purchase order deleted successfully",
	})
}

// UpdatePOStatus handles PATCH /api/v1/purchase-orders/{id}/status
func (h *POHandler) UpdatePOStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid purchase order ID", "VALIDATION_ERROR")
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	po, err := h.poService.UpdatePOStatus(uint(id), body.Status)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to update purchase order status"
		code := "INTERNAL_ERROR"
		if serviceErr, ok := err.(*services.ServiceError); ok {
			message = serviceErr.Message
			code = serviceErr.Code
			switch serviceErr.Err {
			case services.ErrValidation:
				status = http.StatusBadRequest
			case services.ErrNotFound:
				status = http.StatusNotFound
			case services.ErrForbidden:
				status = http.StatusForbidden
			}
		}
		utils.Error(w, status, message, code)
		return
	}

	utils.Success(w, http.StatusOK, "Purchase order status updated successfully", po)
}

// ReceivePO handles POST /api/v1/purchase-orders/{id}/receive
func (h *POHandler) ReceivePO(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid purchase order ID", "VALIDATION_ERROR")
		return
	}

	var input services.ReceivePOInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	po, err := h.poService.ReceivePO(uint(id), input)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to receive purchase order"
		code := "INTERNAL_ERROR"
		if serviceErr, ok := err.(*services.ServiceError); ok {
			message = serviceErr.Message
			code = serviceErr.Code
			switch serviceErr.Err {
			case services.ErrValidation:
				status = http.StatusBadRequest
			case services.ErrNotFound:
				status = http.StatusNotFound
			case services.ErrForbidden:
				status = http.StatusForbidden
			}
		}
		utils.Error(w, status, message, code)
		return
	}

	utils.Success(w, http.StatusOK, "Purchase order received successfully", po)
}

// GetProductsForPO handles GET /api/v1/purchase-orders/products
func (h *POHandler) GetProductsForPO(w http.ResponseWriter, r *http.Request) {
	var supplierID uint
	if s := r.URL.Query().Get("supplierId"); s != "" {
		if id, err := strconv.ParseUint(s, 10, 64); err == nil {
			supplierID = uint(id)
		}
	}
	search := r.URL.Query().Get("search")

	products, err := h.poService.GetProductsForPO(supplierID, search)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch products", "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"data": products,
	})
}
