package utils

import (
	"encoding/json"
	"net/http"
)

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// JSON writes a JSON response with the given status code
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Success writes a successful JSON response
// Format: {"data": {...}, "message": "optional"}
func Success(w http.ResponseWriter, status int, message string, data interface{}) {
	JSON(w, status, SuccessResponse{
		Data:    data,
		Message: message,
	})
}

// Error writes an error JSON response
// Format: {"error": "message", "code": "ERROR_CODE"}
func Error(w http.ResponseWriter, status int, message string, code string) {
	JSON(w, status, ErrorResponse{
		Error: message,
		Code:  code,
	})
}
