package guild

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

// HTTPClient resolves guild membership from the guild service HTTP API.
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient constructs a guild membership reader.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{},
	}
}

type membershipSnapshot struct {
	ID string `json:"id"`
}

// IsGuildMember returns whether the player currently belongs to the given guild.
func (c *HTTPClient) IsGuildMember(ctx context.Context, guildID string, playerID string) (bool, *apperrors.Error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/guild-memberships/"+playerID, nil)
	if err != nil {
		internal := apperrors.Internal()
		return false, &internal
	}

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("guild_unavailable", "guild service is unavailable", http.StatusBadGateway)
		return false, &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		var appErr apperrors.Error
		if decodeErr := json.NewDecoder(resp.Body).Decode(&appErr); decodeErr != nil {
			badGateway := apperrors.New("guild_invalid_response", "guild service returned an invalid response", http.StatusBadGateway)
			return false, &badGateway
		}
		appErr.Status = resp.StatusCode
		return false, &appErr
	}

	var snapshot membershipSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		badGateway := apperrors.New("guild_invalid_response", "guild service returned an invalid response", http.StatusBadGateway)
		return false, &badGateway
	}

	return snapshot.ID == guildID, nil
}

func (c *HTTPClient) String() string {
	return fmt.Sprintf("chat-guild-http-client(%s)", c.baseURL)
}
