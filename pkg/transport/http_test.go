package transport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
)

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	WriteJSON(recorder, http.StatusAccepted, StatusPayload{
		Service: "gateway",
		Status:  "ok",
	})

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("unexpected status: got %d want %d", recorder.Code, http.StatusAccepted)
	}

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("unexpected content type: got %q", contentType)
	}

	var payload StatusPayload
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if payload.Service != "gateway" || payload.Status != "ok" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestWriteError(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	WriteError(recorder, apperrors.New("invalid_request", "bad request", http.StatusBadRequest))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: got %d want %d", recorder.Code, http.StatusBadRequest)
	}

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if payload["code"] != "invalid_request" {
		t.Fatalf("unexpected code: %+v", payload)
	}
}
