package social

import (
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/idgen"
)

const (
	friendRequestPending  = "pending"
	friendRequestAccepted = "accepted"
)

type FriendRequest struct {
	ID           string    `json:"id"`
	FromPlayerID string    `json:"from_player_id"`
	ToPlayerID   string    `json:"to_player_id"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type BlockRelationship struct {
	PlayerID  string    `json:"player_id"`
	BlockedID string    `json:"blocked_id"`
	CreatedAt time.Time `json:"created_at"`
}

type FriendRemark struct {
	PlayerID  string    `json:"player_id"`
	FriendID  string    `json:"friend_id"`
	Remark    string    `json:"remark"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Service struct {
	mu        sync.RWMutex
	now       func() time.Time
	requests  map[string]FriendRequest
	friendMap map[string]map[string]struct{}
	blocks    map[string]map[string]BlockRelationship
	remarks   map[string]map[string]FriendRemark
}

func NewService() *Service {
	return &Service{
		now:       time.Now,
		requests:  make(map[string]FriendRequest),
		friendMap: make(map[string]map[string]struct{}),
		blocks:    make(map[string]map[string]BlockRelationship),
		remarks:   make(map[string]map[string]FriendRemark),
	}
}

func (s *Service) SendFriendRequest(fromPlayerID, toPlayerID string) (FriendRequest, *apperrors.Error) {
	if fromPlayerID == "" || toPlayerID == "" {
		err := apperrors.New("invalid_request", "from_player_id and to_player_id are required", http.StatusBadRequest)
		return FriendRequest{}, &err
	}
	if fromPlayerID == toPlayerID {
		err := apperrors.New("invalid_request", "cannot friend yourself", http.StatusBadRequest)
		return FriendRequest{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isBlockedLocked(fromPlayerID, toPlayerID) || s.isBlockedLocked(toPlayerID, fromPlayerID) {
		err := apperrors.New("blocked", "friend request is blocked by relationship settings", http.StatusForbidden)
		return FriendRequest{}, &err
	}
	if s.isFriendLocked(fromPlayerID, toPlayerID) {
		err := apperrors.New("already_friends", "players are already friends", http.StatusConflict)
		return FriendRequest{}, &err
	}
	for _, request := range s.requests {
		if request.FromPlayerID == fromPlayerID && request.ToPlayerID == toPlayerID && request.Status == friendRequestPending {
			return request, nil
		}
	}

	requestID, err := idgen.Token(8)
	if err != nil {
		internal := apperrors.Internal()
		return FriendRequest{}, &internal
	}

	request := FriendRequest{
		ID:           requestID,
		FromPlayerID: fromPlayerID,
		ToPlayerID:   toPlayerID,
		Status:       friendRequestPending,
		CreatedAt:    s.now(),
	}
	s.requests[request.ID] = request
	return request, nil
}

func (s *Service) AcceptFriendRequest(requestID, actorPlayerID string) (FriendRequest, *apperrors.Error) {
	if requestID == "" || actorPlayerID == "" {
		err := apperrors.New("invalid_request", "request_id and actor_player_id are required", http.StatusBadRequest)
		return FriendRequest{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	request, ok := s.requests[requestID]
	if !ok {
		err := apperrors.New("not_found", "friend request not found", http.StatusNotFound)
		return FriendRequest{}, &err
	}
	if request.ToPlayerID != actorPlayerID {
		err := apperrors.New("forbidden", "only the requested player can accept", http.StatusForbidden)
		return FriendRequest{}, &err
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

func (s *Service) ListFriendRequests(playerID, role, status string) ([]FriendRequest, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", http.StatusBadRequest)
		return nil, &err
	}
	if role == "" {
		role = "all"
	}
	if role != "all" && role != "inbox" && role != "outbox" {
		err := apperrors.New("invalid_request", "role must be all, inbox, or outbox", http.StatusBadRequest)
		return nil, &err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]FriendRequest, 0)
	for _, request := range s.requests {
		if !matchesRequestRole(request, playerID, role) {
			continue
		}
		if status != "" && request.Status != status {
			continue
		}
		list = append(list, request)
	}
	slices.SortFunc(list, func(a, b FriendRequest) int {
		if !a.CreatedAt.Equal(b.CreatedAt) {
			if a.CreatedAt.Before(b.CreatedAt) {
				return -1
			}
			return 1
		}
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})
	return list, nil
}

func (s *Service) ListFriends(playerID string) ([]string, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", http.StatusBadRequest)
		return nil, &err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.friendIDsLocked(playerID), nil
}

func (s *Service) BlockPlayer(playerID, blockedID string) (BlockRelationship, *apperrors.Error) {
	if playerID == "" || blockedID == "" {
		err := apperrors.New("invalid_request", "player_id and blocked_player_id are required", http.StatusBadRequest)
		return BlockRelationship{}, &err
	}
	if playerID == blockedID {
		err := apperrors.New("invalid_request", "cannot block yourself", http.StatusBadRequest)
		return BlockRelationship{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	record := BlockRelationship{
		PlayerID:  playerID,
		BlockedID: blockedID,
		CreatedAt: s.now(),
	}
	if _, ok := s.blocks[playerID]; !ok {
		s.blocks[playerID] = make(map[string]BlockRelationship)
	}
	s.blocks[playerID][blockedID] = record
	return record, nil
}

func (s *Service) ListBlocks(playerID string) ([]string, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", http.StatusBadRequest)
		return nil, &err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	blockMap := s.blocks[playerID]
	list := make([]string, 0, len(blockMap))
	for blockedID := range blockMap {
		list = append(list, blockedID)
	}
	slices.Sort(list)
	return list, nil
}

func (s *Service) SetFriendRemark(playerID, friendID, remark string) (FriendRemark, *apperrors.Error) {
	if playerID == "" || friendID == "" {
		err := apperrors.New("invalid_request", "player_id and friend_id are required", http.StatusBadRequest)
		return FriendRemark{}, &err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.isFriendLocked(playerID, friendID) {
		err := apperrors.New("not_friend", "remarks require an accepted friendship", http.StatusConflict)
		return FriendRemark{}, &err
	}
	if _, ok := s.remarks[playerID]; !ok {
		s.remarks[playerID] = make(map[string]FriendRemark)
	}
	record := FriendRemark{
		PlayerID:  playerID,
		FriendID:  friendID,
		Remark:    strings.TrimSpace(remark),
		UpdatedAt: s.now(),
	}
	s.remarks[playerID][friendID] = record
	return record, nil
}

func (s *Service) ListFriendRemarks(playerID string) ([]FriendRemark, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", http.StatusBadRequest)
		return nil, &err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	remarkMap := s.remarks[playerID]
	list := make([]FriendRemark, 0, len(remarkMap))
	for _, record := range remarkMap {
		list = append(list, record)
	}
	slices.SortFunc(list, func(a, b FriendRemark) int {
		if a.FriendID < b.FriendID {
			return -1
		}
		if a.FriendID > b.FriendID {
			return 1
		}
		return 0
	})
	return list, nil
}

func (s *Service) addFriendLocked(playerID, friendID string) {
	if _, ok := s.friendMap[playerID]; !ok {
		s.friendMap[playerID] = make(map[string]struct{})
	}
	s.friendMap[playerID][friendID] = struct{}{}
}

func (s *Service) friendIDsLocked(playerID string) []string {
	friendSet := s.friendMap[playerID]
	list := make([]string, 0, len(friendSet))
	for friendID := range friendSet {
		list = append(list, friendID)
	}
	slices.Sort(list)
	return list
}

func (s *Service) isFriendLocked(playerID, friendID string) bool {
	_, ok := s.friendMap[playerID][friendID]
	return ok
}

func (s *Service) isBlockedLocked(playerID, blockedID string) bool {
	_, ok := s.blocks[playerID][blockedID]
	return ok
}

func matchesRequestRole(request FriendRequest, playerID, role string) bool {
	switch role {
	case "inbox":
		return request.ToPlayerID == playerID
	case "outbox":
		return request.FromPlayerID == playerID
	default:
		return request.FromPlayerID == playerID || request.ToPlayerID == playerID
	}
}
