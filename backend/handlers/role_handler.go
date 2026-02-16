package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pointofsale/backend/services"
	"github.com/pointofsale/backend/utils"
)

// RoleHandler handles role-related HTTP requests
type RoleHandler struct {
	roleService *services.RoleService
}

// NewRoleHandler creates a new role handler instance
func NewRoleHandler(roleService *services.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

// ListRoles returns paginated list of roles with user counts
func (h *RoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	allowedSortFields := []string{"id", "name", "description"}
	params, err := utils.ParsePaginationParams(r, allowedSortFields)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	// Call service
	roles, total, serviceErr := h.roleService.ListRoles(
		params.Page,
		params.PageSize,
		params.Search,
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
		Data: roles,
		Meta: meta,
	}

	utils.JSON(w, http.StatusOK, response)
}

// GetRole returns a single role by ID
func (h *RoleHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid role ID", "VALIDATION_ERROR")
		return
	}

	// Call service
	role, serviceErr := h.roleService.GetRole(uint(id))
	if serviceErr != nil {
		status := http.StatusInternalServerError
		if serviceErr.Err == services.ErrNotFound {
			status = http.StatusNotFound
		}
		utils.Error(w, status, serviceErr.Message, serviceErr.Code)
		return
	}

	utils.Success(w, http.StatusOK, "", role)
}

// CreateRole creates a new role
func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var input services.RoleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	// Call service
	role, serviceErr := h.roleService.CreateRole(input)
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

	utils.Success(w, http.StatusCreated, "Role created successfully", role)
}

// UpdateRole updates an existing role
func (h *RoleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid role ID", "VALIDATION_ERROR")
		return
	}

	var input services.RoleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	// Call service
	role, serviceErr := h.roleService.UpdateRole(uint(id), input)
	if serviceErr != nil {
		status := http.StatusInternalServerError
		switch serviceErr.Err {
		case services.ErrValidation:
			status = http.StatusBadRequest
		case services.ErrNotFound:
			status = http.StatusNotFound
		case services.ErrConflict:
			status = http.StatusConflict
		case services.ErrForbidden:
			status = http.StatusForbidden
		}
		utils.Error(w, status, serviceErr.Message, serviceErr.Code)
		return
	}

	utils.Success(w, http.StatusOK, "Role updated successfully", role)
}

// DeleteRole deletes a role by ID
func (h *RoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid role ID", "VALIDATION_ERROR")
		return
	}

	// Call service
	serviceErr := h.roleService.DeleteRole(uint(id))
	if serviceErr != nil {
		status := http.StatusInternalServerError
		switch serviceErr.Err {
		case services.ErrNotFound:
			status = http.StatusNotFound
		case services.ErrForbidden:
			status = http.StatusForbidden
		}
		utils.Error(w, status, serviceErr.Message, serviceErr.Code)
		return
	}

	utils.Success(w, http.StatusOK, "Role deleted successfully", nil)
}
