package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pointofsale/backend/middleware"
	"github.com/pointofsale/backend/services"
	"github.com/pointofsale/backend/utils"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input services.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	user, serviceErr := h.authService.Register(input)
	if serviceErr != nil {
		// Map service error to HTTP status code
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

	utils.Success(w, http.StatusCreated, "Registration successful. Please wait for admin approval.", user)
}

// Login handles user authentication
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input services.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	loginResp, serviceErr := h.authService.Login(input)
	if serviceErr != nil {
		// Map service error to HTTP status code
		status := http.StatusInternalServerError
		switch serviceErr.Err {
		case services.ErrValidation:
			status = http.StatusBadRequest
		case services.ErrUnauthorized:
			status = http.StatusUnauthorized
		case services.ErrForbidden:
			status = http.StatusForbidden
		}
		utils.Error(w, status, serviceErr.Message, serviceErr.Code)
		return
	}

	utils.Success(w, http.StatusOK, "Login successful", loginResp)
}

// Refresh handles token refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	if req.RefreshToken == "" {
		utils.Error(w, http.StatusBadRequest, "Refresh token is required", "VALIDATION_ERROR")
		return
	}

	tokenPair, serviceErr := h.authService.RefreshToken(req.RefreshToken)
	if serviceErr != nil {
		utils.Error(w, http.StatusUnauthorized, serviceErr.Message, serviceErr.Code)
		return
	}

	utils.Success(w, http.StatusOK, "Token refreshed successfully", tokenPair)
}

// Logout handles user logout (requires authentication)
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract access token from Authorization header
	accessToken := ""
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 {
			accessToken = parts[1]
		}
	}

	// Extract refresh token from request body
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	// Call logout service (always returns nil currently)
	_ = h.authService.Logout(accessToken, req.RefreshToken)

	utils.Success(w, http.StatusOK, "Logged out successfully", nil)
}

// ForgotPassword handles password reset request
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	// Always returns nil to avoid email enumeration
	_ = h.authService.ForgotPassword(req.Email)

	utils.Success(w, http.StatusOK, "If the email exists, a reset link has been sent.", nil)
}

// ResetPassword handles password reset with token
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var input services.ResetPasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	serviceErr := h.authService.ResetPassword(input)
	if serviceErr != nil {
		// Map service error to HTTP status code
		status := http.StatusInternalServerError
		switch serviceErr.Err {
		case services.ErrValidation:
			status = http.StatusBadRequest
		case services.ErrUnauthorized:
			status = http.StatusUnauthorized
		}
		utils.Error(w, status, serviceErr.Message, serviceErr.Code)
		return
	}

	utils.Success(w, http.StatusOK, "Password reset successfully. Please login with your new password.", nil)
}

// GetMe returns the current authenticated user's details (requires authentication)
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		utils.Error(w, http.StatusUnauthorized, "User not authenticated", "UNAUTHORIZED")
		return
	}

	currentUser, serviceErr := h.authService.GetCurrentUser(userID)
	if serviceErr != nil {
		status := http.StatusInternalServerError
		if serviceErr.Err == services.ErrNotFound {
			status = http.StatusNotFound
		}
		utils.Error(w, status, serviceErr.Message, serviceErr.Code)
		return
	}

	utils.Success(w, http.StatusOK, "", currentUser)
}
