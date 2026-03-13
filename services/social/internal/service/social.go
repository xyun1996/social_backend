package service

import (
	"fmt"
	"slices"
	"sync"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/idgen"
	"github.com/xyun1996/social_backend/services/social/internal/domain"
)

const (
	friendRequestPending  = "pending"
	friendRequestAccepted = "accepted"
)

// SocialService provides an in-memory prototype of friendship and blocklist flows.
type SocialService struct {
	mu           sync.RWMutex
	requests     map[string]domain.FriendRequest
	friendships  map[string]map[string]struct{}
	blocks       map[string]map[string]domain.BlockRelationship
	now          func() time.Time
	newRequestID func() (string, error)
}

// NewSocialService constructs an in-memory social graph service.
func NewSocialService() *SocialService {
	return &SocialService{
		requests:    make(map[string]domain.FriendRequest),
		friendships: make(map[string]map[string]struct{}),
		blocks:      make(map[string]map[string]domain.BlockRelationship),
		now:         time.Now,
		newRequestID: func() (string, error) {
			return idgen.Token(8)
		},
	}
}

// SendFriendRequest creates a pending friend request if the relationship is allowed.
func (s *SocialService) SendFriendRequest(fromPlayerID string, toPlayerID string) (domain.FriendRequest, *apperrors.Error) {
	if fromPlayerID == "" || toPlayerID == "" {
		err := apperrors.New("invalid_request", "from_player_id and to_player_id are required", 400)
		return domain.FriendRequest{}, &err
	}

	if fromPlayerID == toPlayerID {
		err := apperrors.New("invalid_request", "cannot friend yourself", 400)
		return domain.FriendRequest{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isBlockedLocked(fromPlayerID, toPlayerID) || s.isBlockedLocked(toPlayerID, fromPlayerID) {
		err := apperrors.New("blocked", "friend request is blocked by relationship settings", 403)
		return domain.FriendRequest{}, &err
	}

	if s.isFriendLocked(fromPlayerID, toPlayerID) {
		err := apperrors.New("already_friends", "players are already friends", 409)
		return domain.FriendRequest{}, &err
	}

	for _, request := range s.requests {
		if request.FromPlayerID == fromPlayerID && request.ToPlayerID == toPlayerID && request.Status == friendRequestPending {
			return request, nil
		}
	}

	requestID, err := s.newRequestID()
	if err != nil {
		internal := apperrors.Internal()
		return domain.FriendRequest{}, &internal
	}

	request := domain.FriendRequest{
		ID:           requestID,
		FromPlayerID: fromPlayerID,
		ToPlayerID:   toPlayerID,
		Status:       friendRequestPending,
		CreatedAt:    s.now(),
	}

	s.requests[request.ID] = request
	return request, nil
}

// AcceptFriendRequest accepts a pending request and creates a bidirectional friendship.
func (s *SocialService) AcceptFriendRequest(requestID string, actorPlayerID string) (domain.FriendRequest, *apperrors.Error) {
	if requestID == "" || actorPlayerID == "" {
		err := apperrors.New("invalid_request", "request_id and actor_player_id are required", 400)
		return domain.FriendRequest{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	request, ok := s.requests[requestID]
	if !ok {
		err := apperrors.New("not_found", "friend request not found", 404)
		return domain.FriendRequest{}, &err
	}

	if request.ToPlayerID != actorPlayerID {
		err := apperrors.New("forbidden", "only the requested player can accept", 403)
		return domain.FriendRequest{}, &err
	}

	if request.Status == friendRequestAccepted {
		return request, nil
	}

	request.Status = friendRequestAccepted
	s.requests[requestID] = request
	s.addFriendLocked(request.FromPlayerID, request.ToPlayerID)
	s.addFriendLocked(request.ToPlayerID, request.FromPlayerID)
	return request, nil
}

// ListFriendRequests returns friend requests visible to a player.
func (s *SocialService) ListFriendRequests(playerID string, role string, status string) ([]domain.FriendRequest, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", 400)
		return nil, &err
	}
	if role == "" {
		role = "all"
	}
	if role != "all" && role != "inbox" && role != "outbox" {
		err := apperrors.New("invalid_request", "role must be all, inbox, or outbox", 400)
		return nil, &err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	requests := make([]domain.FriendRequest, 0)
	for _, request := range s.requests {
		if !matchesRequestRole(request, playerID, role) {
			continue
		}
		if status != "" && request.Status != status {
			continue
		}
		requests = append(requests, request)
	}

	slices.SortFunc(requests, func(a domain.FriendRequest, b domain.FriendRequest) int {
		if !a.CreatedAt.Equal(b.CreatedAt) {
			if a.CreatedAt.Before(b.CreatedAt) {
				return -1
			}
			return 1
		}
		switch {
		case a.ID < b.ID:
			return -1
		case a.ID > b.ID:
			return 1
		default:
			return 0
		}
	})

	return requests, nil
}

// ListFriends returns a stable friend list for a player.
func (s *SocialService) ListFriends(playerID string) ([]string, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", 400)
		return nil, &err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	friends := make([]string, 0, len(s.friendships[playerID]))
	for friendID := range s.friendships[playerID] {
		friends = append(friends, friendID)
	}
	slices.Sort(friends)
	return friends, nil
}

// BlockPlayer records a point-to-point block relationship.
func (s *SocialService) BlockPlayer(playerID string, blockedID string) (domain.BlockRelationship, *apperrors.Error) {
	if playerID == "" || blockedID == "" {
		err := apperrors.New("invalid_request", "player_id and blocked_player_id are required", 400)
		return domain.BlockRelationship{}, &err
	}

	if playerID == blockedID {
		err := apperrors.New("invalid_request", "cannot block yourself", 400)
		return domain.BlockRelationship{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.blocks[playerID] == nil {
		s.blocks[playerID] = make(map[string]domain.BlockRelationship)
	}

	block := domain.BlockRelationship{
		PlayerID:  playerID,
		BlockedID: blockedID,
		CreatedAt: s.now(),
	}

	s.blocks[playerID][blockedID] = block
	return block, nil
}

// ListBlocks returns all players blocked by the given player.
func (s *SocialService) ListBlocks(playerID string) ([]string, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", 400)
		return nil, &err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	blocked := make([]string, 0, len(s.blocks[playerID]))
	for blockedID := range s.blocks[playerID] {
		blocked = append(blocked, blockedID)
	}
	slices.Sort(blocked)
	return blocked, nil
}

func (s *SocialService) addFriendLocked(playerID string, friendID string) {
	if s.friendships[playerID] == nil {
		s.friendships[playerID] = make(map[string]struct{})
	}

	s.friendships[playerID][friendID] = struct{}{}
}

func (s *SocialService) isFriendLocked(playerID string, friendID string) bool {
	_, ok := s.friendships[playerID][friendID]
	return ok
}

func (s *SocialService) isBlockedLocked(playerID string, blockedID string) bool {
	blocked := s.blocks[playerID]
	if blocked == nil {
		return false
	}

	_, ok := blocked[blockedID]
	return ok
}

func matchesRequestRole(request domain.FriendRequest, playerID string, role string) bool {
	switch role {
	case "inbox":
		return request.ToPlayerID == playerID
	case "outbox":
		return request.FromPlayerID == playerID
	default:
		return request.FromPlayerID == playerID || request.ToPlayerID == playerID
	}
}

func (s *SocialService) String() string {
	return fmt.Sprintf("social-service(requests=%d)", len(s.requests))
}
