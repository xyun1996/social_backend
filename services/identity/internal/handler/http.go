package handler

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/transport"
	"github.com/xyun1996/social_backend/services/identity/internal/service"
)

// AuthHTTPHandler exposes the early identity HTTP API.
type AuthHTTPHandler struct {
	auth *service.AuthService
}

// NewAuthHTTPHandler constructs the identity HTTP handler set.
func NewAuthHTTPHandler(auth *service.AuthService) *AuthHTTPHandler {
	return &AuthHTTPHandler{auth: auth}
}

type loginRequest struct {
	AccountID string `json:"account_id"`
	PlayerID  string `json:"player_id"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Routes returns the identity HTTP routes.
func (h *AuthHTTPHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("POST /v1/auth/login", h.handleLogin)
	mux.HandleFunc("POST /v1/auth/refresh", h.handleRefresh)
	return mux
}

func (h *AuthHTTPHandler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	transport.WriteJSON(w, http.StatusOK, transport.StatusPayload{
		Service: "identity",
		Status:  "ok",
	})
}

func (h *AuthHTTPHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	pair, appErr := h.auth.Login(request.AccountID, request.PlayerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, pair)
}

func (h *AuthHTTPHandler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var request refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	pair, appErr := h.auth.Refresh(request.RefreshToken)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, pair)
}

func invalidJSONError() apperrors.Error {
	return apperrors.New("invalid_json", "request body must be valid json", http.StatusBadRequest)
}
