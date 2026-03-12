package apperrors

import "net/http"

// Error defines a transport-safe application error.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

// New creates an application error with an HTTP status mapping.
func New(code string, message string, status int) Error {
	return Error{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// Internal returns a generic internal error suitable for public transport responses.
func Internal() Error {
	return New("internal_error", "internal server error", http.StatusInternalServerError)
}
