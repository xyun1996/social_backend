package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	gatewayservice "github.com/xyun1996/social_backend/services/gateway/internal/service"
)

// HTTPClient calls the chat HTTP API for delivery planning.
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient constructs a chat HTTP client.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{},
	}
}

// PlanDelivery fetches chat delivery targets for the conversation sender pair.
func (c *HTTPClient) PlanDelivery(ctx context.Context, conversationID string, senderPlayerID string) ([]gatewayservice.ChatDeliveryTarget, *apperrors.Error) {
	endpoint := fmt.Sprintf("%s/v1/conversations/%s/delivery?sender_player_id=%s", c.baseURL, url.PathEscape(conversationID), url.QueryEscape(senderPlayerID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("chat_unavailable", "chat service is unavailable", http.StatusBadGateway)
		return nil, &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var appErr apperrors.Error
		if decodeErr := json.NewDecoder(resp.Body).Decode(&appErr); decodeErr != nil {
			badGateway := apperrors.New("chat_invalid_response", "chat service returned an invalid response", http.StatusBadGateway)
			return nil, &badGateway
		}
		appErr.Status = resp.StatusCode
		return nil, &appErr
	}

	var payload struct {
		Targets []gatewayservice.ChatDeliveryTarget `json:"targets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		badGateway := apperrors.New("chat_invalid_response", "chat service returned an invalid response", http.StatusBadGateway)
		return nil, &badGateway
	}

	return payload.Targets, nil
}

// AckConversation forwards a read ack to the chat HTTP API.
func (c *HTTPClient) AckConversation(ctx context.Context, conversationID string, playerID string, ackSeq int64) *apperrors.Error {
	body, err := json.Marshal(map[string]any{
		"player_id": playerID,
		"ack_seq":   ackSeq,
	})
	if err != nil {
		internal := apperrors.Internal()
		return &internal
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/v1/conversations/%s/ack", c.baseURL, url.PathEscape(conversationID)), bytes.NewReader(body))
	if err != nil {
		internal := apperrors.Internal()
		return &internal
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("chat_unavailable", "chat service is unavailable", http.StatusBadGateway)
		return &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var appErr apperrors.Error
		if decodeErr := json.NewDecoder(resp.Body).Decode(&appErr); decodeErr != nil {
			badGateway := apperrors.New("chat_invalid_response", "chat service returned an invalid response", http.StatusBadGateway)
			return &badGateway
		}
		appErr.Status = resp.StatusCode
		return &appErr
	}

	return nil
}

func (c *HTTPClient) String() string {
	return fmt.Sprintf("chat-http-client(%s)", c.baseURL)
}
