package handler

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/transport"
	"github.com/xyun1996/social_backend/services/worker/internal/service"
)

// HTTPHandler exposes the early worker HTTP API.
type HTTPHandler struct {
	worker *service.WorkerService
}

// NewHTTPHandler constructs the worker HTTP routes.
func NewHTTPHandler(worker *service.WorkerService) *HTTPHandler {
	return &HTTPHandler{worker: worker}
}

type enqueueRequest struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type claimRequest struct {
	WorkerID string `json:"worker_id"`
	Type     string `json:"type"`
}

type transitionRequest struct {
	WorkerID  string `json:"worker_id"`
	LastError string `json:"last_error,omitempty"`
}

type runRequest struct {
	WorkerID string `json:"worker_id"`
	Type     string `json:"type,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

// Routes returns the worker HTTP routes.
func (h *HTTPHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("POST /v1/jobs", h.handleEnqueue)
	mux.HandleFunc("GET /v1/jobs", h.handleListJobs)
	mux.HandleFunc("POST /v1/jobs/claim", h.handleClaim)
	mux.HandleFunc("POST /v1/jobs/{jobID}/complete", h.handleComplete)
	mux.HandleFunc("POST /v1/jobs/{jobID}/fail", h.handleFail)
	mux.HandleFunc("POST /v1/jobs/run-once", h.handleRunOnce)
	mux.HandleFunc("POST /v1/jobs/run-until-empty", h.handleRunUntilEmpty)
	return mux
}

func (h *HTTPHandler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	transport.WriteJSON(w, http.StatusOK, transport.StatusPayload{
		Service: "worker",
		Status:  "ok",
	})
}

func (h *HTTPHandler) handleEnqueue(w http.ResponseWriter, r *http.Request) {
	var request enqueueRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	job, appErr := h.worker.Enqueue(request.Type, request.Payload)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, job)
}

func (h *HTTPHandler) handleListJobs(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	jobType := r.URL.Query().Get("type")
	jobs, appErr := h.worker.ListJobs(status, jobType)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, map[string]any{
		"status": status,
		"type":   jobType,
		"count":  len(jobs),
		"jobs":   jobs,
	})
}

func (h *HTTPHandler) handleClaim(w http.ResponseWriter, r *http.Request) {
	var request claimRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	job, appErr := h.worker.ClaimNext(request.WorkerID, request.Type)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, job)
}

func (h *HTTPHandler) handleComplete(w http.ResponseWriter, r *http.Request) {
	var request transitionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	job, appErr := h.worker.Complete(r.PathValue("jobID"), request.WorkerID)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, job)
}

func (h *HTTPHandler) handleFail(w http.ResponseWriter, r *http.Request) {
	var request transitionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	job, appErr := h.worker.Fail(r.PathValue("jobID"), request.WorkerID, request.LastError)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, job)
}

func (h *HTTPHandler) handleRunOnce(w http.ResponseWriter, r *http.Request) {
	var request runRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	result, appErr := h.worker.ExecuteNext(r.Context(), request.WorkerID, request.Type)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, result)
}

func (h *HTTPHandler) handleRunUntilEmpty(w http.ResponseWriter, r *http.Request) {
	var request runRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		transport.WriteError(w, invalidJSONError())
		return
	}

	result, appErr := h.worker.ExecuteUntilEmpty(r.Context(), request.WorkerID, request.Type, request.Limit)
	if appErr != nil {
		transport.WriteError(w, *appErr)
		return
	}

	transport.WriteJSON(w, http.StatusOK, result)
}

func invalidJSONError() apperrors.Error {
	return apperrors.New("invalid_json", "request body must be valid json", http.StatusBadRequest)
}
