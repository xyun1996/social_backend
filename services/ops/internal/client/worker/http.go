package worker

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

// HTTPClient calls the worker HTTP API for operator reads.
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

// GetWorkerSnapshot fetches jobs filtered by optional status and type.
func (c *HTTPClient) GetWorkerSnapshot(ctx context.Context, status string, jobType string) (opsservice.WorkerSnapshot, *apperrors.Error) {
	query := url.Values{}
	if status != "" {
		query.Set("status", status)
	}
	if jobType != "" {
		query.Set("type", jobType)
	}

	path := c.baseURL + "/v1/jobs"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		internal := apperrors.Internal()
		return opsservice.WorkerSnapshot{}, &internal
	}

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("worker_unavailable", "worker service is unavailable", http.StatusBadGateway)
		return opsservice.WorkerSnapshot{}, &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var appErr apperrors.Error
		if decodeErr := json.NewDecoder(resp.Body).Decode(&appErr); decodeErr != nil {
			badGateway := apperrors.New("worker_invalid_response", "worker service returned an invalid response", http.StatusBadGateway)
			return opsservice.WorkerSnapshot{}, &badGateway
		}
		appErr.Status = resp.StatusCode
		return opsservice.WorkerSnapshot{}, &appErr
	}

	var snapshot opsservice.WorkerSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		badGateway := apperrors.New("worker_invalid_response", "worker service returned an invalid response", http.StatusBadGateway)
		return opsservice.WorkerSnapshot{}, &badGateway
	}
	return snapshot, nil
}

func (c *HTTPClient) String() string {
	return fmt.Sprintf("worker-http-client(%s)", c.baseURL)
}
