package utils

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type APIError struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Success(w http.ResponseWriter, status int, message string, data interface{}) {
	JSON(w, status, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, APIError{
		Success: false,
		Error:   message,
	})
}
