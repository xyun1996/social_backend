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

type HTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{baseURL: strings.TrimRight(baseURL, "/"), client: &http.Client{}}
}

type pendingSummary struct {
	Inbox        []string `json:"inbox"`
	Outbox       []string `json:"outbox"`
	TotalPending int      `json:"total_pending"`
}

func (c *HTTPClient) GetSocialSnapshot(ctx context.Context, playerID string) (opsservice.SocialSnapshot, *apperrors.Error) {
	friends, appErr := c.getStringList(ctx, "/v1/friends", playerID, "friends")
	if appErr != nil { return opsservice.SocialSnapshot{}, appErr }
	blocks, appErr := c.getStringList(ctx, "/v1/blocks", playerID, "blocks")
	if appErr != nil { return opsservice.SocialSnapshot{}, appErr }
	pending, appErr := c.getPendingSummary(ctx, playerID)
	if appErr != nil { return opsservice.SocialSnapshot{}, appErr }
	relationships, appErr := c.getRelationships(ctx, playerID)
	if appErr != nil { return opsservice.SocialSnapshot{}, appErr }
	return opsservice.SocialSnapshot{PlayerID: playerID, Friends: friends, Blocks: blocks, PendingInbox: pending.Inbox, PendingOutbox: pending.Outbox, PendingTotal: pending.TotalPending, RelationshipDetails: relationships}, nil
}

func (c *HTTPClient) getStringList(ctx context.Context, path string, playerID string, field string) ([]string, *apperrors.Error) {
	endpoint := fmt.Sprintf("%s%s?player_id=%s", c.baseURL, path, url.QueryEscape(playerID))
	var payload map[string]any
	if appErr := c.getJSON(ctx, endpoint, &payload); appErr != nil { return nil, appErr }
	rawItems, _ := payload[field].([]any)
	items := make([]string, 0, len(rawItems))
	for _, item := range rawItems {
		if value, _ := item.(string); value != "" { items = append(items, value) }
	}
	return items, nil
}

func (c *HTTPClient) getPendingSummary(ctx context.Context, playerID string) (pendingSummary, *apperrors.Error) {
	endpoint := fmt.Sprintf("%s/v1/pending-social?player_id=%s", c.baseURL, url.QueryEscape(playerID))
	var payload pendingSummary
	if appErr := c.getJSON(ctx, endpoint, &payload); appErr != nil { return pendingSummary{}, appErr }
	return payload, nil
}

func (c *HTTPClient) getRelationships(ctx context.Context, playerID string) ([]opsservice.SocialRelationshipDetail, *apperrors.Error) {
	endpoint := fmt.Sprintf("%s/v1/relationships?player_id=%s", c.baseURL, url.QueryEscape(playerID))
	var payload struct { Relationships []opsservice.SocialRelationshipDetail `json:"relationships"` }
	if appErr := c.getJSON(ctx, endpoint, &payload); appErr != nil { return nil, appErr }
	return payload.Relationships, nil
}

func (c *HTTPClient) getJSON(ctx context.Context, endpoint string, target any) *apperrors.Error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil { internal := apperrors.Internal(); return &internal }
	resp, err := c.client.Do(req)
	if err != nil { badGateway := apperrors.New("social_unavailable", "social service is unavailable", http.StatusBadGateway); return &badGateway }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var appErr apperrors.Error
		if decodeErr := json.NewDecoder(resp.Body).Decode(&appErr); decodeErr != nil { badGateway := apperrors.New("social_invalid_response", "social service returned an invalid response", http.StatusBadGateway); return &badGateway }
		appErr.Status = resp.StatusCode
		return &appErr
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil { badGateway := apperrors.New("social_invalid_response", "social service returned an invalid response", http.StatusBadGateway); return &badGateway }
	return nil
}

func (c *HTTPClient) String() string { return fmt.Sprintf("social-http-client(%s)", c.baseURL) }
