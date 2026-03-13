package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/services/worker/internal/domain"
	"github.com/xyun1996/social_backend/services/worker/internal/service"
)

func TestWorkerJobLifecycleEndpoints(t *testing.T) {
	t.Parallel()

	h := NewHTTPHandler(service.NewWorkerService())

	enqueueReq := httptest.NewRequest(http.MethodPost, "/v1/jobs", bytes.NewBufferString(`{"type":"invite.expire","payload":"{}"}`))
	enqueueRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(enqueueRec, enqueueReq)
	if enqueueRec.Code != http.StatusOK {
		t.Fatalf("unexpected enqueue status: got %d want %d", enqueueRec.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.Unmarshal(enqueueRec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal enqueue response: %v", err)
	}
	jobID, _ := payload["id"].(string)

	claimReq := httptest.NewRequest(http.MethodPost, "/v1/jobs/claim", bytes.NewBufferString(`{"worker_id":"worker-a","type":"invite.expire"}`))
	claimRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(claimRec, claimReq)
	if claimRec.Code != http.StatusOK {
		t.Fatalf("unexpected claim status: got %d want %d", claimRec.Code, http.StatusOK)
	}

	completeReq := httptest.NewRequest(http.MethodPost, "/v1/jobs/"+jobID+"/complete", bytes.NewBufferString(`{"worker_id":"worker-a"}`))
	completeRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(completeRec, completeReq)
	if completeRec.Code != http.StatusOK {
		t.Fatalf("unexpected complete status: got %d want %d", completeRec.Code, http.StatusOK)
	}
}

func TestWorkerRunOnceEndpoint(t *testing.T) {
	t.Parallel()

	worker := service.NewWorkerService()
	worker.RegisterHandler("invite.expire", func(_ context.Context, _ domain.Job) *apperrors.Error {
		return nil
	})
	h := NewHTTPHandler(worker)

	enqueueReq := httptest.NewRequest(http.MethodPost, "/v1/jobs", bytes.NewBufferString(`{"type":"invite.expire","payload":"{}"}`))
	enqueueRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(enqueueRec, enqueueReq)
	if enqueueRec.Code != http.StatusOK {
		t.Fatalf("unexpected enqueue status: got %d want %d", enqueueRec.Code, http.StatusOK)
	}

	runReq := httptest.NewRequest(http.MethodPost, "/v1/jobs/run-once", bytes.NewBufferString(`{"worker_id":"worker-a","type":"invite.expire"}`))
	runRec := httptest.NewRecorder()
	h.Routes().ServeHTTP(runRec, runReq)
	if runRec.Code != http.StatusOK {
		t.Fatalf("unexpected run-once status: got %d want %d", runRec.Code, http.StatusOK)
	}
}
