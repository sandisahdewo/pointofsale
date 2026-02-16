package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pointofsale/backend/middleware"
	"github.com/pointofsale/backend/repositories"
	"github.com/pointofsale/backend/services"
	"github.com/pointofsale/backend/utils"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// ListUsers handles GET /api/v1/users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSizeStr := r.URL.Query().Get("pageSize")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	search := r.URL.Query().Get("search")
	sortBy := r.URL.Query().Get("sortBy")
	if sortBy == "" {
		sortBy = "id"
	}

	sortDir := r.URL.Query().Get("sortDir")
	if sortDir == "" {
		sortDir = "asc"
	}

	status := r.URL.Query().Get("status")

	// Build pagination params
	params := repositories.PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
		SortBy:   sortBy,
		SortDir:  sortDir,
	}

	// Call service
	users, total, err := h.userService.ListUsers(params, status)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch users", "INTERNAL_ERROR")
		return
	}

	// Calculate pagination meta
	totalPages := (int(total) + pageSize - 1) / pageSize

	response := map[string]interface{}{
		"data": users,
		"meta": map[string]interface{}{
			"page":       page,
			"pageSize":   pageSize,
			"totalItems": total,
			"totalPages": totalPages,
		},
	}

	utils.JSON(w, http.StatusOK, response)
}

// GetUser handles GET /api/v1/users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid user ID", "VALIDATION_ERROR")
		return
	}

	// Get user
	user, err := h.userService.GetUser(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to fetch user"
		code := "INTERNAL_ERROR"

		// Type assert to ServiceError to get details
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

	utils.Success(w, http.StatusOK, "", user)
}

// CreateUser handles POST /api/v1/users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input services.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	// Create user
	user, err := h.userService.CreateUser(input)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to create user"
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

	utils.Success(w, http.StatusCreated, "User created successfully", user)
}

// UpdateUser handles PUT /api/v1/users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Parse ID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid user ID", "VALIDATION_ERROR")
		return
	}

	var input services.UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	// Update user
	user, err := h.userService.UpdateUser(uint(id), input)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to update user"
		code := "INTERNAL_ERROR"

		if serviceErr, ok := err.(*services.ServiceError); ok {
			message = serviceErr.Message
			code = serviceErr.Code
			switch serviceErr.Err {
			case services.ErrValidation:
				status = http.StatusBadRequest
			case services.ErrConflict:
				status = http.StatusConflict
			case services.ErrNotFound:
				status = http.StatusNotFound
			case services.ErrForbidden:
				status = http.StatusForbidden
			}
		}
		utils.Error(w, status, message, code)
		return
	}

	utils.Success(w, http.StatusOK, "User updated successfully", user)
}

// DeleteUser handles DELETE /api/v1/users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Parse ID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid user ID", "VALIDATION_ERROR")
		return
	}

	// Get current user ID from context
	currentUserID := middleware.GetUserID(r.Context())
	if currentUserID == 0 {
		utils.Error(w, http.StatusUnauthorized, "User not authenticated", "UNAUTHORIZED")
		return
	}

	// Delete user
	err = h.userService.DeleteUser(uint(id), currentUserID)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to delete user"
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

	utils.Success(w, http.StatusOK, "User deleted successfully", nil)
}

// ApproveUser handles PATCH /api/v1/users/{id}/approve
func (h *UserHandler) ApproveUser(w http.ResponseWriter, r *http.Request) {
	// Parse ID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid user ID", "VALIDATION_ERROR")
		return
	}

	// Approve user
	user, err := h.userService.ApproveUser(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to approve user"
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

	utils.Success(w, http.StatusOK, "User approved successfully", user)
}

// RejectUser handles DELETE /api/v1/users/{id}/reject
func (h *UserHandler) RejectUser(w http.ResponseWriter, r *http.Request) {
	// Parse ID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid user ID", "VALIDATION_ERROR")
		return
	}

	// Reject user
	err = h.userService.RejectUser(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to reject user"
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

	utils.Success(w, http.StatusOK, "User registration rejected", nil)
}

// UploadProfilePicture handles POST /api/v1/users/{id}/profile-picture
func (h *UserHandler) UploadProfilePicture(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement file upload handling
	// This requires:
	// 1. Parse multipart form
	// 2. Validate file type (JPEG, PNG, WebP)
	// 3. Validate file size (max 2MB)
	// 4. Save to disk (backend/uploads/profiles/{userId}_{timestamp}.{ext})
	// 5. Update user's profile_picture field
	// 6. Delete old file if exists

	utils.Error(w, http.StatusNotImplemented, "Profile picture upload not implemented yet", "NOT_IMPLEMENTED")
}
