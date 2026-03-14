package party

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

// HTTPClient resolves party membership from the party service HTTP API.
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient constructs a party membership reader.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{},
	}
}

type membershipSnapshot struct {
	PartyID string `json:"party_id"`
}

// IsPartyMember returns whether the player currently belongs to the given party.
func (c *HTTPClient) IsPartyMember(ctx context.Context, partyID string, playerID string) (bool, *apperrors.Error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/party-memberships/"+playerID, nil)
	if err != nil {
		internal := apperrors.Internal()
		return false, &internal
	}

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("party_unavailable", "party service is unavailable", http.StatusBadGateway)
		return false, &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		var appErr apperrors.Error
		if decodeErr := json.NewDecoder(resp.Body).Decode(&appErr); decodeErr != nil {
			badGateway := apperrors.New("party_invalid_response", "party service returned an invalid response", http.StatusBadGateway)
			return false, &badGateway
		}
		appErr.Status = resp.StatusCode
		return false, &appErr
	}

	var snapshot membershipSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		badGateway := apperrors.New("party_invalid_response", "party service returned an invalid response", http.StatusBadGateway)
		return false, &badGateway
	}

	return snapshot.PartyID == partyID, nil
}

func (c *HTTPClient) String() string {
	return fmt.Sprintf("chat-party-http-client(%s)", c.baseURL)
}
