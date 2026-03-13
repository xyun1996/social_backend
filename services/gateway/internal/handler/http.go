package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/transport"
	gatewayservice "github.com/xyun1996/social_backend/services/gateway/internal/service"
)

// HTTPHandler exposes the early gateway HTTP API.
type HTTPHandler struct {
	introspector gatewayservice.Introspector
	reporter     gatewayservice.PresenceReporter
	realtime     *gatewayservice.RealtimeService
	delivery     *gatewayservice.DeliveryService
}

// NewHTTPHandler constructs a gateway HTTP handler.
func NewHTTPHandler(introspector gatewayservice.Introspector, reporter gatewayservice.PresenceReporter, planner gatewayservice.ChatPlanner) *HTTPHandler {
	realtime := gatewayservice.NewRealtimeService(introspector, reporter)
	return &HTTPHandler{
		introspector: introspector,
		reporter:     reporter,
		realtime:     realtime,
		delivery:     gatewayservice.NewDeliveryService(realtime, planner),
	}
}

type presenceRequest struct {
	SessionID string `json:"session_id"`
	RealmID   string `json:"realm_id"`
	Location  string `json:"location"`
}

type handshakeRequest struct {
	AccessToken   string `json:"access_token"`
	SessionID     string `json:"session_id"`
	RealmID       string `json:"realm_id"`
	Location      string `json:"location"`
	ClientVersion string `json:"client_version"`
}

type resumeRequest struct {
	AccessToken       string `json:"access_token"`
	SessionID         string `json:"session_id"`
	LastServerEventID string `json:"last_server_event_id"`
}

type chatDispatchRequest struct {
	ConversationID string `json:"conversation_id"`
	SenderPlayerID string `json:"sender_player_id"`
	MessageID      string `json:"message_id"`
	Seq            int64  `json:"seq"`
	Body           string `json:"body"`
	SentAt         string `json:"sent_at"`
}

// Routes returns the gateway HTTP routes.
func (h *HTTPHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("GET /v1/session/me", h.handleSessionMe)
	mux.HandleFunc("POST /v1/session/presence/connect", h.handlePresenceConnect)
	mux.HandleFunc("POST /v1/session/presence/heartbeat", h.handlePresenceHeartbeat)
	mux.HandleFunc("POST /v1/session/presence/disconnect", h.handlePresenceDisconnect)
	mux.HandleFunc("POST /v1/realtime/handshake", h.handleRealtimeHandshake)
	mux.HandleFunc("POST /v1/realtime/resume", h.handleRealtimeResume)
	mux.HandleFunc("POST /v1/realtime/sessions/{sessionID}/heartbeat", h.handleRealtimeHeartbeat)
	mux.HandleFunc("POST /v1/realtime/sessions/{sessionID}/close", h.handleRealtimeClose)
	mux.HandleFunc("GET /v1/realtime/sessions/{sessionID}", h.handleRealtimeGetSession)
	mux.HandleFunc("GET /v1/realtime/sessions/{sessionID}/events", h.handleRealtimeGetSessionEvents)
	mux.HandleFunc("POST /v1/realtime/chat/deliveries", h.handleRealtimeChatDelivery)
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

func (h *HTTPHandler) handlePresenceConnect(w http.ResponseWriter, r *http.Request) {
	h.handlePresenceUpdate(w, r, "connect")
}

func (h *HTTPHandler) handlePresenceHeartbeat(w http.ResponseWriter, r *http.Request) {
	h.handlePresenceUpdate(w, r, "heartbeat")
}

func (h *HTTPHandler) handlePresenceDisconnect(w http.ResponseWriter, r *http.Request) {
	h.handlePresenceUpdate(w, r, "disconnect")
}

func (h *HTTPHandler) handleRealtimeHandshake(w http.ResponseWriter, r *http.Request) {
	var request handshakeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	session, appErr := h.realtime.Handshake(r.Context(), gatewayservice.HandshakeRequest{
		AccessToken:   request.AccessToken,
		SessionID:     request.SessionID,
		RealmID:       request.RealmID,
		Location:      request.Location,
		ClientVersion: request.ClientVersion,
	})
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, session)
}

func (h *HTTPHandler) handleRealtimeResume(w http.ResponseWriter, r *http.Request) {
	var request resumeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	session, appErr := h.realtime.Resume(r.Context(), gatewayservice.ResumeRequest{
		AccessToken:       request.AccessToken,
		SessionID:         request.SessionID,
		LastServerEventID: request.LastServerEventID,
	})
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, session)
}

func (h *HTTPHandler) handleRealtimeHeartbeat(w http.ResponseWriter, r *http.Request) {
	session, appErr := h.realtime.Heartbeat(r.Context(), r.PathValue("sessionID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, session)
}

func (h *HTTPHandler) handleRealtimeClose(w http.ResponseWriter, r *http.Request) {
	session, appErr := h.realtime.Close(r.Context(), r.PathValue("sessionID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, session)
}

func (h *HTTPHandler) handleRealtimeGetSession(w http.ResponseWriter, r *http.Request) {
	session, appErr := h.realtime.GetSession(r.PathValue("sessionID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, session)
}

func (h *HTTPHandler) handleRealtimeGetSessionEvents(w http.ResponseWriter, r *http.Request) {
	inbox, appErr := h.realtime.GetSessionEvents(r.PathValue("sessionID"))
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, inbox)
}

func (h *HTTPHandler) handleRealtimeChatDelivery(w http.ResponseWriter, r *http.Request) {
	var request chatDispatchRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	result, appErr := h.delivery.DispatchChat(r.Context(), gatewayservice.ChatDispatchRequest{
		ConversationID: request.ConversationID,
		SenderPlayerID: request.SenderPlayerID,
		MessageID:      request.MessageID,
		Seq:            request.Seq,
		Body:           request.Body,
		SentAt:         request.SentAt,
	})
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, result)
}

func (h *HTTPHandler) handlePresenceUpdate(w http.ResponseWriter, r *http.Request, action string) {
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

	var request presenceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	update := gatewayservice.PresenceUpdate{
		PlayerID:  subject.PlayerID,
		SessionID: request.SessionID,
		RealmID:   request.RealmID,
		Location:  request.Location,
	}

	var snapshot gatewayservice.PresenceSnapshot
	switch action {
	case "connect":
		snapshot, appErr = h.reporter.Connect(r.Context(), update)
	case "heartbeat":
		snapshot, appErr = h.reporter.Heartbeat(r.Context(), update)
	case "disconnect":
		snapshot, appErr = h.reporter.Disconnect(r.Context(), update)
	default:
		internal := apperrors.Internal()
		transport.WriteError(w, internal)
		return
	}

	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, snapshot)
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

func invalidJSONError() apperrors.Error {
	return apperrors.New("invalid_json", "request body must be valid json", http.StatusBadRequest)
}
