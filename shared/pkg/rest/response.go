package rest

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// RespondJSON sends a JSON response
func RespondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// RespondError sends an error response
func RespondError(w http.ResponseWriter, statusCode int, code, message string) {
	RespondJSON(w, statusCode, ErrorResponse{
		Code:    code,
		Message: message,
	})
}

// RespondErrorWithDetails sends an error response with details
func RespondErrorWithDetails(w http.ResponseWriter, statusCode int, code, message, details string) {
	RespondJSON(w, statusCode, ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	})
}

// RespondSuccess sends a success response
func RespondSuccess(w http.ResponseWriter, data interface{}) {
	RespondJSON(w, http.StatusOK, SuccessResponse{
		Data: data,
	})
}

// RespondCreated sends a created response
func RespondCreated(w http.ResponseWriter, data interface{}) {
	RespondJSON(w, http.StatusCreated, SuccessResponse{
		Data: data,
	})
}

// RespondNoContent sends a no content response
func RespondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
