package party

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	opsservice "github.com/xyun1996/social_backend/services/ops/internal/service"
)

// HTTPClient calls the party HTTP API for operator reads.
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient constructs a party HTTP client.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{},
	}
}

// GetPartySnapshot fetches the party member snapshot.
func (c *HTTPClient) GetPartySnapshot(ctx context.Context, partyID string) (opsservice.PartySnapshot, *apperrors.Error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/parties/"+partyID+"/members", nil)
	if err != nil {
		internal := apperrors.Internal()
		return opsservice.PartySnapshot{}, &internal
	}

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("party_unavailable", "party service is unavailable", http.StatusBadGateway)
		return opsservice.PartySnapshot{}, &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var appErr apperrors.Error
		if decodeErr := json.NewDecoder(resp.Body).Decode(&appErr); decodeErr != nil {
			badGateway := apperrors.New("party_invalid_response", "party service returned an invalid response", http.StatusBadGateway)
			return opsservice.PartySnapshot{}, &badGateway
		}
		appErr.Status = resp.StatusCode
		return opsservice.PartySnapshot{}, &appErr
	}

	var record opsservice.PartySnapshot
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		badGateway := apperrors.New("party_invalid_response", "party service returned an invalid response", http.StatusBadGateway)
		return opsservice.PartySnapshot{}, &badGateway
	}
	queue, appErr := c.getQueueState(ctx, partyID)
	if appErr != nil {
		return opsservice.PartySnapshot{}, appErr
	}
	record.Queue = queue
	return record, nil
}

func (c *HTTPClient) String() string {
	return fmt.Sprintf("party-http-client(%s)", c.baseURL)
}

func (c *HTTPClient) getQueueState(ctx context.Context, partyID string) (*opsservice.PartyQueueState, *apperrors.Error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/parties/"+partyID+"/queue", nil)
	if err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("party_unavailable", "party service is unavailable", http.StatusBadGateway)
		return nil, &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		var appErr apperrors.Error
		if decodeErr := json.NewDecoder(resp.Body).Decode(&appErr); decodeErr != nil {
			badGateway := apperrors.New("party_invalid_response", "party service returned an invalid response", http.StatusBadGateway)
			return nil, &badGateway
		}
		appErr.Status = resp.StatusCode
		return nil, &appErr
	}

	var record opsservice.PartyQueueState
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		badGateway := apperrors.New("party_invalid_response", "party service returned an invalid response", http.StatusBadGateway)
		return nil, &badGateway
	}
	return &record, nil
}
