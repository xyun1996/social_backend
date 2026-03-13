package social

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	opsservice "github.com/xyun1996/social_backend/services/ops/internal/service"
)

// HTTPClient calls the social HTTP API for operator reads.
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient constructs a social HTTP client.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{},
	}
}

// GetSocialSnapshot fetches the player's current friends and blocks.
func (c *HTTPClient) GetSocialSnapshot(ctx context.Context, playerID string) (opsservice.SocialSnapshot, *apperrors.Error) {
	friends, appErr := c.getStringList(ctx, "/v1/friends", playerID, "friends")
	if appErr != nil {
		return opsservice.SocialSnapshot{}, appErr
	}

	blocks, appErr := c.getStringList(ctx, "/v1/blocks", playerID, "blocks")
	if appErr != nil {
		return opsservice.SocialSnapshot{}, appErr
	}
	inbox, appErr := c.getRequestPlayers(ctx, playerID, "inbox")
	if appErr != nil {
		return opsservice.SocialSnapshot{}, appErr
	}
	outbox, appErr := c.getRequestPlayers(ctx, playerID, "outbox")
	if appErr != nil {
		return opsservice.SocialSnapshot{}, appErr
	}

	return opsservice.SocialSnapshot{
		PlayerID:      playerID,
		Friends:       friends,
		Blocks:        blocks,
		PendingInbox:  inbox,
		PendingOutbox: outbox,
	}, nil
}

func (c *HTTPClient) getStringList(ctx context.Context, path string, playerID string, field string) ([]string, *apperrors.Error) {
	endpoint := fmt.Sprintf("%s%s?player_id=%s", c.baseURL, path, url.QueryEscape(playerID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("social_unavailable", "social service is unavailable", http.StatusBadGateway)
		return nil, &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var appErr apperrors.Error
		if decodeErr := json.NewDecoder(resp.Body).Decode(&appErr); decodeErr != nil {
			badGateway := apperrors.New("social_invalid_response", "social service returned an invalid response", http.StatusBadGateway)
			return nil, &badGateway
		}
		appErr.Status = resp.StatusCode
		return nil, &appErr
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		badGateway := apperrors.New("social_invalid_response", "social service returned an invalid response", http.StatusBadGateway)
		return nil, &badGateway
	}

	rawItems, _ := payload[field].([]any)
	items := make([]string, 0, len(rawItems))
	for _, item := range rawItems {
		value, _ := item.(string)
		if value != "" {
			items = append(items, value)
		}
	}

	return items, nil
}

func (c *HTTPClient) getRequestPlayers(ctx context.Context, playerID string, role string) ([]string, *apperrors.Error) {
	endpoint := fmt.Sprintf("%s/v1/friends/requests?player_id=%s&role=%s&status=pending", c.baseURL, url.QueryEscape(playerID), url.QueryEscape(role))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("social_unavailable", "social service is unavailable", http.StatusBadGateway)
		return nil, &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var appErr apperrors.Error
		if decodeErr := json.NewDecoder(resp.Body).Decode(&appErr); decodeErr != nil {
			badGateway := apperrors.New("social_invalid_response", "social service returned an invalid response", http.StatusBadGateway)
			return nil, &badGateway
		}
		appErr.Status = resp.StatusCode
		return nil, &appErr
	}

	var payload struct {
		Requests []struct {
			FromPlayerID string `json:"from_player_id"`
			ToPlayerID   string `json:"to_player_id"`
		} `json:"requests"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		badGateway := apperrors.New("social_invalid_response", "social service returned an invalid response", http.StatusBadGateway)
		return nil, &badGateway
	}

	players := make([]string, 0, len(payload.Requests))
	for _, request := range payload.Requests {
		if role == "inbox" {
			players = append(players, request.FromPlayerID)
		} else {
			players = append(players, request.ToPlayerID)
		}
	}
	return players, nil
}

func (c *HTTPClient) String() string {
	return fmt.Sprintf("social-http-client(%s)", c.baseURL)
}
