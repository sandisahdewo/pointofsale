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

// SalesHandler handles HTTP requests for sales endpoints.
type SalesHandler struct {
	salesService *services.SalesService
}

// NewSalesHandler creates a new sales handler instance.
func NewSalesHandler(salesService *services.SalesService) *SalesHandler {
	return &SalesHandler{salesService: salesService}
}

// Allowed sort fields for sales transactions.
var salesSortFields = []string{"date", "transaction_number", "grand_total"}

// ProductSearch handles GET /api/v1/sales/products/search?q=...
func (h *SalesHandler) ProductSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	results, err := h.salesService.ProductSearch(q)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to search products"
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

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"data": results,
	})
}

// Checkout handles POST /api/v1/sales/checkout
func (h *SalesHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	var input services.CheckoutInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	result, err := h.salesService.Checkout(input)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to process checkout"
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

	utils.Success(w, http.StatusCreated, "Checkout successful", result)
}

// ListTransactions handles GET /api/v1/sales/transactions
func (h *SalesHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	paginationParams, err := utils.ParsePaginationParams(r, salesSortFields)
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

	dateFrom := r.URL.Query().Get("dateFrom")
	dateTo := r.URL.Query().Get("dateTo")
	paymentMethod := r.URL.Query().Get("paymentMethod")

	transactions, total, err := h.salesService.ListTransactions(params, dateFrom, dateTo, paymentMethod)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch transactions", "INTERNAL_ERROR")
		return
	}

	meta := utils.CalculatePaginationMeta(params.Page, params.PageSize, int(total))

	utils.JSON(w, http.StatusOK, utils.PaginatedResponse{
		Data: transactions,
		Meta: meta,
	})
}

// GetTransaction handles GET /api/v1/sales/transactions/:id
func (h *SalesHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid transaction ID", "VALIDATION_ERROR")
		return
	}

	tx, err := h.salesService.GetTransaction(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to fetch transaction"
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

	utils.Success(w, http.StatusOK, "", tx)
}
