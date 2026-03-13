package guild

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	opsservice "github.com/xyun1996/social_backend/services/ops/internal/service"
)

// HTTPClient calls the guild HTTP API for operator reads.
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient constructs a guild HTTP client.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{},
	}
}

// GetGuildSnapshot fetches the guild member snapshot.
func (c *HTTPClient) GetGuildSnapshot(ctx context.Context, guildID string) (opsservice.GuildSnapshot, *apperrors.Error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/guilds/"+guildID+"/members", nil)
	if err != nil {
		internal := apperrors.Internal()
		return opsservice.GuildSnapshot{}, &internal
	}

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("guild_unavailable", "guild service is unavailable", http.StatusBadGateway)
		return opsservice.GuildSnapshot{}, &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var appErr apperrors.Error
		if decodeErr := json.NewDecoder(resp.Body).Decode(&appErr); decodeErr != nil {
			badGateway := apperrors.New("guild_invalid_response", "guild service returned an invalid response", http.StatusBadGateway)
			return opsservice.GuildSnapshot{}, &badGateway
		}
		appErr.Status = resp.StatusCode
		return opsservice.GuildSnapshot{}, &appErr
	}

	var record opsservice.GuildSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		badGateway := apperrors.New("guild_invalid_response", "guild service returned an invalid response", http.StatusBadGateway)
		return opsservice.GuildSnapshot{}, &badGateway
	}
	return record, nil
}

func (c *HTTPClient) String() string {
	return fmt.Sprintf("guild-http-client(%s)", c.baseURL)
}
