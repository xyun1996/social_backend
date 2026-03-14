package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/xyun1996/social_backend/pkg/auth"
	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

// HTTPClient calls the chat HTTP API for offline delivery processing.
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

// RecordOfflineDelivery invokes the internal chat processing endpoint.
func (c *HTTPClient) RecordOfflineDelivery(ctx context.Context, payload map[string]any) *apperrors.Error {
	body, err := json.Marshal(payload)
	if err != nil {
		internal := apperrors.Internal()
		return &internal
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/internal/offline-deliveries", bytes.NewReader(body))
	if err != nil {
		internal := apperrors.Internal()
		return &internal
	}
	req.Header.Set("Content-Type", "application/json")
	auth.ApplyInternalToken(req)

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("chat_unavailable", "chat service is unavailable", http.StatusBadGateway)
		return &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		appErr := apperrors.New("chat_offline_delivery_failed", "chat service failed to process offline delivery", http.StatusBadGateway)
		return &appErr
	}

	return nil
}

func (c *HTTPClient) String() string {
	return fmt.Sprintf("chat-http-client(%s)", c.baseURL)
}
