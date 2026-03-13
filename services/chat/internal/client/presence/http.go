package presence

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	chatservice "github.com/xyun1996/social_backend/services/chat/internal/service"
)

// HTTPClient calls the presence HTTP API for delivery planning.
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

// GetPresence fetches a player's current presence state.
func (c *HTTPClient) GetPresence(ctx context.Context, playerID string) (chatservice.PresenceSnapshot, *apperrors.Error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/presence/"+playerID, nil)
	if err != nil {
		internal := apperrors.Internal()
		return chatservice.PresenceSnapshot{}, &internal
	}

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("presence_unavailable", "presence service is unavailable", http.StatusBadGateway)
		return chatservice.PresenceSnapshot{}, &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var appErr apperrors.Error
		if decodeErr := json.NewDecoder(resp.Body).Decode(&appErr); decodeErr != nil {
			badGateway := apperrors.New("presence_invalid_response", "presence service returned an invalid response", http.StatusBadGateway)
			return chatservice.PresenceSnapshot{}, &badGateway
		}
		appErr.Status = resp.StatusCode
		return chatservice.PresenceSnapshot{}, &appErr
	}

	var snapshot chatservice.PresenceSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		badGateway := apperrors.New("presence_invalid_response", "presence service returned an invalid response", http.StatusBadGateway)
		return chatservice.PresenceSnapshot{}, &badGateway
	}

	return snapshot, nil
}

func (c *HTTPClient) String() string {
	return fmt.Sprintf("presence-http-client(%s)", c.baseURL)
}
