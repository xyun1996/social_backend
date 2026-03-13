package invite

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	partyservice "github.com/xyun1996/social_backend/services/party/internal/service"
)

// HTTPClient calls the invite HTTP API for party invite flows.
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient constructs an invite HTTP client.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{},
	}
}

// CreateInvite creates a cross-service invite for the party domain.
func (c *HTTPClient) CreateInvite(ctx context.Context, domainName string, resourceID string, fromPlayerID string, toPlayerID string) (partyservice.Invite, *apperrors.Error) {
	body, err := json.Marshal(map[string]any{
		"domain":         domainName,
		"resource_id":    resourceID,
		"from_player_id": fromPlayerID,
		"to_player_id":   toPlayerID,
	})
	if err != nil {
		internal := apperrors.Internal()
		return partyservice.Invite{}, &internal
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/invites", bytes.NewReader(body))
	if err != nil {
		internal := apperrors.Internal()
		return partyservice.Invite{}, &internal
	}
	req.Header.Set("Content-Type", "application/json")

	return c.do(req)
}

// GetInvite fetches a single invite by id.
func (c *HTTPClient) GetInvite(ctx context.Context, inviteID string) (partyservice.Invite, *apperrors.Error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/invites/"+inviteID, nil)
	if err != nil {
		internal := apperrors.Internal()
		return partyservice.Invite{}, &internal
	}

	return c.do(req)
}

func (c *HTTPClient) do(req *http.Request) (partyservice.Invite, *apperrors.Error) {
	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("invite_unavailable", "invite service is unavailable", http.StatusBadGateway)
		return partyservice.Invite{}, &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var appErr apperrors.Error
		if decodeErr := json.NewDecoder(resp.Body).Decode(&appErr); decodeErr != nil {
			badGateway := apperrors.New("invite_invalid_response", "invite service returned an invalid response", http.StatusBadGateway)
			return partyservice.Invite{}, &badGateway
		}

		appErr.Status = resp.StatusCode
		return partyservice.Invite{}, &appErr
	}

	var invite partyservice.Invite
	if err := json.NewDecoder(resp.Body).Decode(&invite); err != nil {
		badGateway := apperrors.New("invite_invalid_response", "invite service returned an invalid response", http.StatusBadGateway)
		return partyservice.Invite{}, &badGateway
	}

	return invite, nil
}

func (c *HTTPClient) String() string {
	return fmt.Sprintf("invite-http-client(%s)", c.baseURL)
}
