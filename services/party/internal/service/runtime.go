package service

import (
	"net/http"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/services/party/internal/domain"
)

func (s *PartyService) getActiveQueueState(partyID string) (domain.QueueState, bool, error) {
	state, ok, err := s.queues.GetQueueState(partyID)
	if err != nil || !ok {
		return state, ok, err
	}
	if state.ExpiresAt != nil && !state.ExpiresAt.After(s.now()) {
		_ = s.queues.DeleteQueueAssignment(partyID)
		_ = s.queues.DeleteQueueState(partyID)
		return domain.QueueState{}, false, nil
	}
	return state, true, nil
}

func (s *PartyService) SweepExpiredQueues() (map[string]any, *apperrors.Error) {
	parties, err := s.parties.ListParties()
	if err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}
	removed := make([]string, 0)
	for _, party := range parties {
		state, ok, err := s.queues.GetQueueState(party.ID)
		if err != nil {
			internal := apperrors.Internal()
			return nil, &internal
		}
		if !ok || state.ExpiresAt == nil || state.ExpiresAt.After(s.now()) {
			continue
		}
		_ = s.queues.DeleteQueueAssignment(party.ID)
		_ = s.queues.DeleteQueueState(party.ID)
		removed = append(removed, party.ID)
	}
	return map[string]any{"removed_party_ids": removed, "removed_count": len(removed), "swept_at": s.now().UTC().Format(time.RFC3339Nano)}, nil
}

func (s *PartyService) ensureQueueUnlocked(partyID string) *apperrors.Error {
	state, ok, err := s.getActiveQueueState(partyID)
	if err != nil {
		internal := apperrors.Internal()
		return &internal
	}
	if ok {
		msg := "party queue is locked in " + state.QueueName + " (" + state.Status + ")"
		err := apperrors.New("party_queued", msg, http.StatusConflict)
		return &err
	}
	return nil
}
