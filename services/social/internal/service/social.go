package service

import (
	"fmt"
	"slices"
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
	requests     FriendRequestStore
	friendships  FriendshipStore
	blocks       BlockStore
	now          func() time.Time
	newRequestID func() (string, error)
}

// NewSocialService constructs an in-memory social graph service.
func NewSocialService() *SocialService {
	return NewSocialServiceWithStores(
		newMemoryFriendRequestStore(),
		newMemoryFriendshipStore(),
		newMemoryBlockStore(),
	)
}

// NewSocialServiceWithStores constructs the service with injected persistence boundaries.
func NewSocialServiceWithStores(requests FriendRequestStore, friendships FriendshipStore, blocks BlockStore) *SocialService {
	return &SocialService{
		requests:    requests,
		friendships: friendships,
		blocks:      blocks,
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

	if s.isBlocked(fromPlayerID, toPlayerID) || s.isBlocked(toPlayerID, fromPlayerID) {
		err := apperrors.New("blocked", "friend request is blocked by relationship settings", 403)
		return domain.FriendRequest{}, &err
	}

	if s.isFriend(fromPlayerID, toPlayerID) {
		err := apperrors.New("already_friends", "players are already friends", 409)
		return domain.FriendRequest{}, &err
	}

	requests, err := s.requests.ListFriendRequests()
	if err != nil {
		internal := apperrors.Internal()
		return domain.FriendRequest{}, &internal
	}
	for _, request := range requests {
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

	if err := s.requests.SaveFriendRequest(request); err != nil {
		internal := apperrors.Internal()
		return domain.FriendRequest{}, &internal
	}
	return request, nil
}

// AcceptFriendRequest accepts a pending request and creates a bidirectional friendship.
func (s *SocialService) AcceptFriendRequest(requestID string, actorPlayerID string) (domain.FriendRequest, *apperrors.Error) {
	if requestID == "" || actorPlayerID == "" {
		err := apperrors.New("invalid_request", "request_id and actor_player_id are required", 400)
		return domain.FriendRequest{}, &err
	}

	request, ok, err := s.requests.GetFriendRequest(requestID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.FriendRequest{}, &internal
	}
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
	if err := s.requests.SaveFriendRequest(request); err != nil {
		internal := apperrors.Internal()
		return domain.FriendRequest{}, &internal
	}
	if err := s.friendships.SaveFriendship(domain.FriendRelationship{
		PlayerID: request.FromPlayerID,
		FriendID: request.ToPlayerID,
	}); err != nil {
		internal := apperrors.Internal()
		return domain.FriendRequest{}, &internal
	}
	if err := s.friendships.SaveFriendship(domain.FriendRelationship{
		PlayerID: request.ToPlayerID,
		FriendID: request.FromPlayerID,
	}); err != nil {
		internal := apperrors.Internal()
		return domain.FriendRequest{}, &internal
	}
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

	requests, err := s.requests.ListFriendRequests()
	if err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}

	filtered := make([]domain.FriendRequest, 0)
	for _, request := range requests {
		if !matchesRequestRole(request, playerID, role) {
			continue
		}
		if status != "" && request.Status != status {
			continue
		}
		filtered = append(filtered, request)
	}

	slices.SortFunc(filtered, func(a domain.FriendRequest, b domain.FriendRequest) int {
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

	return filtered, nil
}

// ListFriends returns a stable friend list for a player.
func (s *SocialService) ListFriends(playerID string) ([]string, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", 400)
		return nil, &err
	}

	friends, err := s.friendships.ListFriends(playerID)
	if err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}
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

	block := domain.BlockRelationship{
		PlayerID:  playerID,
		BlockedID: blockedID,
		CreatedAt: s.now(),
	}

	if err := s.blocks.SaveBlock(block); err != nil {
		internal := apperrors.Internal()
		return domain.BlockRelationship{}, &internal
	}
	return block, nil
}

// ListBlocks returns all players blocked by the given player.
func (s *SocialService) ListBlocks(playerID string) ([]string, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", 400)
		return nil, &err
	}

	blocked, err := s.blocks.ListBlocks(playerID)
	if err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}
	return blocked, nil
}

func (s *SocialService) isFriend(playerID string, friendID string) bool {
	friends, err := s.friendships.ListFriends(playerID)
	if err != nil {
		return false
	}
	return slices.Contains(friends, friendID)
}

func (s *SocialService) isBlocked(playerID string, blockedID string) bool {
	blocked, err := s.blocks.ListBlocks(playerID)
	if err != nil {
		return false
	}
	return slices.Contains(blocked, blockedID)
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
	requests, err := s.requests.ListFriendRequests()
	if err != nil {
		return "social-service(requests=unknown)"
	}
	return fmt.Sprintf("social-service(requests=%d)", len(requests))
}
