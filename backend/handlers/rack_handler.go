package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pointofsale/backend/services"
	"github.com/pointofsale/backend/utils"
)

// RackHandler handles rack-related HTTP requests
type RackHandler struct {
	rackService *services.RackService
}

// NewRackHandler creates a new rack handler instance
func NewRackHandler(rackService *services.RackService) *RackHandler {
	return &RackHandler{rackService: rackService}
}

// ListRacks returns paginated list of racks
func (h *RackHandler) ListRacks(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	allowedSortFields := []string{"id", "name", "code", "location", "active"}
	params, err := utils.ParsePaginationParams(r, allowedSortFields)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	// Parse active filter
	active := r.URL.Query().Get("active")

	// Call service
	racks, total, serviceErr := h.rackService.ListRacks(
		params.Page,
		params.PageSize,
		params.Search,
		active,
		params.SortBy,
		params.SortDir,
	)
	if serviceErr != nil {
		utils.Error(w, http.StatusInternalServerError, serviceErr.Message, serviceErr.Code)
		return
	}

	// Build paginated response
	meta := utils.CalculatePaginationMeta(params.Page, params.PageSize, int(total))
	response := utils.PaginatedResponse{
		Data: racks,
		Meta: meta,
	}

	utils.JSON(w, http.StatusOK, response)
}

// GetRack returns a single rack by ID
func (h *RackHandler) GetRack(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid rack ID", "VALIDATION_ERROR")
		return
	}

	rack, serviceErr := h.rackService.GetRack(uint(id))
	if serviceErr != nil {
		status := http.StatusInternalServerError
		if serviceErr.Err == services.ErrNotFound {
			status = http.StatusNotFound
		}
		utils.Error(w, status, serviceErr.Message, serviceErr.Code)
		return
	}

	utils.Success(w, http.StatusOK, "", rack)
}

// CreateRack creates a new rack
func (h *RackHandler) CreateRack(w http.ResponseWriter, r *http.Request) {
	var input services.RackInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	rack, serviceErr := h.rackService.CreateRack(input)
	if serviceErr != nil {
		status := http.StatusInternalServerError
		switch serviceErr.Err {
		case services.ErrValidation:
			status = http.StatusBadRequest
		case services.ErrConflict:
			status = http.StatusConflict
		}
		utils.Error(w, status, serviceErr.Message, serviceErr.Code)
		return
	}

	utils.Success(w, http.StatusCreated, "Rack created successfully", rack)
}

// UpdateRack updates an existing rack
func (h *RackHandler) UpdateRack(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid rack ID", "VALIDATION_ERROR")
		return
	}

	var input services.RackInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	rack, serviceErr := h.rackService.UpdateRack(uint(id), input)
	if serviceErr != nil {
		status := http.StatusInternalServerError
		switch serviceErr.Err {
		case services.ErrValidation:
			status = http.StatusBadRequest
		case services.ErrNotFound:
			status = http.StatusNotFound
		case services.ErrConflict:
			status = http.StatusConflict
		}
		utils.Error(w, status, serviceErr.Message, serviceErr.Code)
		return
	}

	utils.Success(w, http.StatusOK, "Rack updated successfully", rack)
}

// DeleteRack deletes a rack by ID
func (h *RackHandler) DeleteRack(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid rack ID", "VALIDATION_ERROR")
		return
	}

	serviceErr := h.rackService.DeleteRack(uint(id))
	if serviceErr != nil {
		status := http.StatusInternalServerError
		switch serviceErr.Err {
		case services.ErrNotFound:
			status = http.StatusNotFound
		}
		utils.Error(w, status, serviceErr.Message, serviceErr.Code)
		return
	}

	utils.Success(w, http.StatusOK, "Rack deleted successfully", nil)
}
