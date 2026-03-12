package handler

import (
	"net/http"
	"strings"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/transport"
	gatewayservice "github.com/xyun1996/social_backend/services/gateway/internal/service"
)

// HTTPHandler exposes the early gateway HTTP API.
type HTTPHandler struct {
	introspector gatewayservice.Introspector
}

// NewHTTPHandler constructs a gateway HTTP handler.
func NewHTTPHandler(introspector gatewayservice.Introspector) *HTTPHandler {
	return &HTTPHandler{introspector: introspector}
}

// Routes returns the gateway HTTP routes.
func (h *HTTPHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("GET /v1/session/me", h.handleSessionMe)
	return mux
}

func (h *HTTPHandler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	transport.WriteJSON(w, http.StatusOK, transport.StatusPayload{
		Service: "gateway",
		Status:  "ok",
	})
}

func (h *HTTPHandler) handleSessionMe(w http.ResponseWriter, r *http.Request) {
	token, appErr := bearerToken(r.Header.Get("Authorization"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	subject, err := h.introspector.Introspect(r.Context(), token)
	if err != nil {
		transport.WriteError(w, *err)
		return
	}

	transport.WriteJSON(w, http.StatusOK, subject)
}

func bearerToken(header string) (string, *apperrors.Error) {
	if !strings.HasPrefix(header, "Bearer ") {
		err := apperrors.New("unauthorized", "bearer token is required", http.StatusUnauthorized)
		return "", &err
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if token == "" {
		err := apperrors.New("unauthorized", "bearer token is required", http.StatusUnauthorized)
		return "", &err
	}

	return token, nil
}
