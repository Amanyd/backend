package apierr

import (
	"encoding/json"
	"net/http"
)

// Error type returned by all API endpoints
type APIError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Constructors
func (e *APIError) Error() string { return e.Message }

func BadRequest(msg string) *APIError {
	return &APIError{Status: http.StatusBadRequest, Code: "bad_request", Message: msg}
}
func NotFound(msg string) *APIError {
	return &APIError{Status: http.StatusNotFound, Code: "not_found", Message: msg}
}
func Unauthorized(msg string) *APIError {
	return &APIError{Status: http.StatusUnauthorized, Code: "unauthorized", Message: msg}
}
func Forbidden(msg string) *APIError {
	return &APIError{Status: http.StatusForbidden, Code: "forbidden", Message: msg}
}
func Conflict(msg string) *APIError {
	return &APIError{Status: http.StatusConflict, Code: "conflict", Message: msg}
}
func Internal(msg string) *APIError {
	return &APIError{Status: http.StatusInternalServerError, Code: "internal_error", Message: msg}
}

// Standard JSON response shape for every endpoint
// On success: {"data":{...}, "error": null}
// On error: {"data":null, "error":{"code":"...", message:"..."}}
type envelope struct {
	Data  any       `json:"data"`
	Error *APIError `json:"error"`
}

// Inspects err: if it's *APIerror, uses its status and code
// Otherwise uses a generic 500 error
func WriteJSON(w http.ResponseWriter, err error) {
	var apiErr *APIError
	switch e := err.(type) {
	case *APIError:
		apiErr = e
	default:
		apiErr = Internal("an unexpected error occurred")
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.Status)
	json.NewEncoder(w).Encode(envelope{Data: nil, Error: apiErr})
}

// Sends a response with the standard envelope
// status is typically http.StatusOK or http.StatusCreated
func WriteData(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(envelope{Data: data, Error: nil})
}
