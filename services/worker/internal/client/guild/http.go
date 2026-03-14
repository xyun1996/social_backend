package guild

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

// HTTPClient calls the guild HTTP API for worker-side progression maintenance.
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient constructs a worker guild client.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{baseURL: strings.TrimRight(baseURL, "/"), client: &http.Client{}}
}

// EnsureCurrentActivityInstances triggers guild-side current period initialization.
func (c *HTTPClient) EnsureCurrentActivityInstances(ctx context.Context, guildID string) *apperrors.Error {
	return c.postNoBody(ctx, "/v1/internal/guilds/"+guildID+"/activities/ensure-current")
}

// CloseExpiredActivityInstances triggers guild-side expiry transitions.
func (c *HTTPClient) CloseExpiredActivityInstances(ctx context.Context, guildID string) *apperrors.Error {
	return c.postNoBody(ctx, "/v1/internal/guilds/"+guildID+"/activities/close-expired")
}

func (c *HTTPClient) postNoBody(ctx context.Context, path string) *apperrors.Error {
	body, _ := json.Marshal(map[string]any{})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		internal := apperrors.Internal()
		return &internal
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("guild_unavailable", "guild service is unavailable", http.StatusBadGateway)
		return &badGateway
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var appErr apperrors.Error
		if decodeErr := json.NewDecoder(resp.Body).Decode(&appErr); decodeErr != nil {
			badGateway := apperrors.New("guild_invalid_response", "guild service returned an invalid response", http.StatusBadGateway)
			return &badGateway
		}
		appErr.Status = resp.StatusCode
		return &appErr
	}
	return nil
}

func (c *HTTPClient) String() string {
	return fmt.Sprintf("worker-guild-http-client(%s)", c.baseURL)
}
