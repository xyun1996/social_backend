package identity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	gatewayservice "github.com/xyun1996/social_backend/services/gateway/internal/service"
)

// HTTPClient calls the identity HTTP API for token introspection.
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient constructs an identity HTTP client.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{},
	}
}

// Introspect resolves a bearer token using the identity service.
func (c *HTTPClient) Introspect(ctx context.Context, accessToken string) (gatewayservice.Subject, *apperrors.Error) {
	body, err := json.Marshal(map[string]string{
		"access_token": accessToken,
	})
	if err != nil {
		internal := apperrors.Internal()
		return gatewayservice.Subject{}, &internal
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/auth/introspect", bytes.NewReader(body))
	if err != nil {
		internal := apperrors.Internal()
		return gatewayservice.Subject{}, &internal
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("identity_unavailable", "identity service is unavailable", http.StatusBadGateway)
		return gatewayservice.Subject{}, &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var appErr apperrors.Error
		if decodeErr := json.NewDecoder(resp.Body).Decode(&appErr); decodeErr != nil {
			badGateway := apperrors.New("identity_invalid_response", "identity service returned an invalid response", http.StatusBadGateway)
			return gatewayservice.Subject{}, &badGateway
		}

		appErr.Status = resp.StatusCode
		return gatewayservice.Subject{}, &appErr
	}

	var subject gatewayservice.Subject
	if err := json.NewDecoder(resp.Body).Decode(&subject); err != nil {
		badGateway := apperrors.New("identity_invalid_response", "identity service returned an invalid response", http.StatusBadGateway)
		return gatewayservice.Subject{}, &badGateway
	}

	return subject, nil
}

func (c *HTTPClient) String() string {
	return fmt.Sprintf("identity-http-client(%s)", c.baseURL)
}
