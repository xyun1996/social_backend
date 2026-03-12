package transport

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

// StatusPayload represents the minimal shared health response shape.
type StatusPayload struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

// WriteJSON writes a JSON response with the given HTTP status code.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}

// WriteError writes a JSON error response based on the shared application error model.
func WriteError(w http.ResponseWriter, err apperrors.Error) {
	WriteJSON(w, err.Status, err)
}
