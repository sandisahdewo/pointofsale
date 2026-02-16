package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pointofsale/backend/services"
	"github.com/pointofsale/backend/utils"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		utils.Error(w, http.StatusBadRequest, "name, email and password are required")
		return
	}

	if req.Role == "" {
		req.Role = "cashier"
	}

	user, err := h.authService.Register(req.Name, req.Email, req.Password, req.Role)
	if err != nil {
		utils.Error(w, http.StatusConflict, err.Error())
		return
	}

	utils.Success(w, http.StatusCreated, "user registered successfully", user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		utils.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}

	tokens, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.Success(w, http.StatusOK, "login successful", tokens)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		utils.Error(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	tokens, err := h.authService.Refresh(req.RefreshToken)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.Success(w, http.StatusOK, "token refreshed successfully", tokens)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Extract access token from Authorization header
	accessToken := ""
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 {
			accessToken = parts[1]
		}
	}

	if err := h.authService.Logout(accessToken, req.RefreshToken); err != nil {
		utils.Error(w, http.StatusInternalServerError, "logout failed")
		return
	}

	utils.Success(w, http.StatusOK, "logged out successfully", nil)
}
