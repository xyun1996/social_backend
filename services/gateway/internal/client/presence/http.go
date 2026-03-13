package presence

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	gatewayservice "github.com/xyun1996/social_backend/services/gateway/internal/service"
)

// HTTPClient calls the presence HTTP API for lifecycle reporting.
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient constructs a presence HTTP client.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{},
	}
}

// Connect forwards an online transition to presence.
func (c *HTTPClient) Connect(ctx context.Context, update gatewayservice.PresenceUpdate) (gatewayservice.PresenceSnapshot, *apperrors.Error) {
	return c.send(ctx, "/v1/presence/connect", update)
}

// Heartbeat forwards a heartbeat update to presence.
func (c *HTTPClient) Heartbeat(ctx context.Context, update gatewayservice.PresenceUpdate) (gatewayservice.PresenceSnapshot, *apperrors.Error) {
	return c.send(ctx, "/v1/presence/heartbeat", update)
}

// Disconnect forwards an offline transition to presence.
func (c *HTTPClient) Disconnect(ctx context.Context, update gatewayservice.PresenceUpdate) (gatewayservice.PresenceSnapshot, *apperrors.Error) {
	return c.send(ctx, "/v1/presence/disconnect", update)
}

func (c *HTTPClient) send(ctx context.Context, path string, update gatewayservice.PresenceUpdate) (gatewayservice.PresenceSnapshot, *apperrors.Error) {
	body, err := json.Marshal(update)
	if err != nil {
		internal := apperrors.Internal()
		return gatewayservice.PresenceSnapshot{}, &internal
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		internal := apperrors.Internal()
		return gatewayservice.PresenceSnapshot{}, &internal
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("presence_unavailable", "presence service is unavailable", http.StatusBadGateway)
		return gatewayservice.PresenceSnapshot{}, &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var appErr apperrors.Error
		if decodeErr := json.NewDecoder(resp.Body).Decode(&appErr); decodeErr != nil {
			badGateway := apperrors.New("presence_invalid_response", "presence service returned an invalid response", http.StatusBadGateway)
			return gatewayservice.PresenceSnapshot{}, &badGateway
		}

		appErr.Status = resp.StatusCode
		return gatewayservice.PresenceSnapshot{}, &appErr
	}

	var snapshot gatewayservice.PresenceSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		badGateway := apperrors.New("presence_invalid_response", "presence service returned an invalid response", http.StatusBadGateway)
		return gatewayservice.PresenceSnapshot{}, &badGateway
	}

	return snapshot, nil
}

func (c *HTTPClient) String() string {
	return fmt.Sprintf("presence-http-client(%s)", c.baseURL)
}
