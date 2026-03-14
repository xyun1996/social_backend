package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

// HTTPClient publishes guild-scoped system messages into chat.
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient constructs a guild chat publisher.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{baseURL: strings.TrimRight(baseURL, "/"), client: &http.Client{}}
}

type conversationRecord struct {
	ID string `json:"id"`
}

// PublishGuildSystemEvent ensures the guild conversation exists and appends a system message.
func (c *HTTPClient) PublishGuildSystemEvent(ctx context.Context, guildID string, memberIDs []string, body string) *apperrors.Error {
	conversation, appErr := c.ensureGuildConversation(ctx, guildID, memberIDs)
	if appErr != nil {
		return appErr
	}
	payload, err := json.Marshal(map[string]string{"sender_player_id": "system", "body": body})
	if err != nil {
		internal := apperrors.Internal()
		return &internal
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/conversations/"+conversation.ID+"/messages", bytes.NewReader(payload))
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
		return decodeChatError(resp)
	}
	return nil
}

func (c *HTTPClient) ensureGuildConversation(ctx context.Context, guildID string, memberIDs []string) (conversationRecord, *apperrors.Error) {
	payload, err := json.Marshal(map[string]any{"kind": "guild", "resource_id": guildID, "member_player_ids": memberIDs})
	if err != nil {
		internal := apperrors.Internal()
		return conversationRecord{}, &internal
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/conversations", bytes.NewReader(payload))
	if err != nil {
		internal := apperrors.Internal()
		return conversationRecord{}, &internal
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("chat_unavailable", "chat service is unavailable", http.StatusBadGateway)
		return conversationRecord{}, &badGateway
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return conversationRecord{}, decodeChatError(resp)
	}
	var record conversationRecord
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		badGateway := apperrors.New("chat_invalid_response", "chat service returned an invalid response", http.StatusBadGateway)
		return conversationRecord{}, &badGateway
	}
	return record, nil
}

func decodeChatError(resp *http.Response) *apperrors.Error {
	var appErr apperrors.Error
	if decodeErr := json.NewDecoder(resp.Body).Decode(&appErr); decodeErr != nil {
		badGateway := apperrors.New("chat_invalid_response", "chat service returned an invalid response", http.StatusBadGateway)
		return &badGateway
	}
	appErr.Status = resp.StatusCode
	return &appErr
}

func (c *HTTPClient) String() string {
	return fmt.Sprintf("guild-chat-http-client(%s)", c.baseURL)
}
