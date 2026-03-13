package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

// HTTPClient calls the worker HTTP API for chat async delivery intent.
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient constructs a worker HTTP client.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{},
	}
}

// EnqueueJob creates a worker job over HTTP.
func (c *HTTPClient) EnqueueJob(ctx context.Context, jobType string, payload string) *apperrors.Error {
	body, err := json.Marshal(map[string]string{
		"type":    jobType,
		"payload": payload,
	})
	if err != nil {
		internal := apperrors.Internal()
		return &internal
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/jobs", bytes.NewReader(body))
	if err != nil {
		internal := apperrors.Internal()
		return &internal
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("worker_unavailable", "worker service is unavailable", http.StatusBadGateway)
		return &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var appErr apperrors.Error
		if decodeErr := json.NewDecoder(resp.Body).Decode(&appErr); decodeErr != nil {
			badGateway := apperrors.New("worker_invalid_response", "worker service returned an invalid response", http.StatusBadGateway)
			return &badGateway
		}
		appErr.Status = resp.StatusCode
		return &appErr
	}

	return nil
}

func (c *HTTPClient) String() string {
	return fmt.Sprintf("worker-http-client(%s)", c.baseURL)
}
