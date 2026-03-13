package presence

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	partyservice "github.com/xyun1996/social_backend/services/party/internal/service"
)

// HTTPClient calls the presence HTTP API for party runtime checks.
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

// GetPresence fetches current presence state for a player.
func (c *HTTPClient) GetPresence(ctx context.Context, playerID string) (partyservice.PresenceSnapshot, *apperrors.Error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/presence/"+playerID, nil)
	if err != nil {
		internal := apperrors.Internal()
		return partyservice.PresenceSnapshot{}, &internal
	}

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("presence_unavailable", "presence service is unavailable", http.StatusBadGateway)
		return partyservice.PresenceSnapshot{}, &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var appErr apperrors.Error
		if decodeErr := json.NewDecoder(resp.Body).Decode(&appErr); decodeErr != nil {
			badGateway := apperrors.New("presence_invalid_response", "presence service returned an invalid response", http.StatusBadGateway)
			return partyservice.PresenceSnapshot{}, &badGateway
		}
		appErr.Status = resp.StatusCode
		return partyservice.PresenceSnapshot{}, &appErr
	}

	var snapshot partyservice.PresenceSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		badGateway := apperrors.New("presence_invalid_response", "presence service returned an invalid response", http.StatusBadGateway)
		return partyservice.PresenceSnapshot{}, &badGateway
	}

	return snapshot, nil
}

func (c *HTTPClient) String() string {
	return fmt.Sprintf("presence-http-client(%s)", c.baseURL)
}
