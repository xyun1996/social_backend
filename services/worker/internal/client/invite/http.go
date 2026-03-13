package invite

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

// HTTPClient calls the invite HTTP API for internal expiry handling.
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient constructs an invite HTTP client.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{},
	}
}

// ExpireInvite invokes the internal invite expiry endpoint.
func (c *HTTPClient) ExpireInvite(ctx context.Context, inviteID string) *apperrors.Error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/internal/invites/"+inviteID+"/expire", nil)
	if err != nil {
		internal := apperrors.Internal()
		return &internal
	}

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("invite_unavailable", "invite service is unavailable", http.StatusBadGateway)
		return &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		appErr := apperrors.New("invite_expire_failed", "invite service failed to expire invite", http.StatusBadGateway)
		return &appErr
	}

	return nil
}

func (c *HTTPClient) String() string {
	return fmt.Sprintf("invite-http-client(%s)", c.baseURL)
}
