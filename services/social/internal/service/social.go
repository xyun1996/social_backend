package service

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/pkg/idgen"
	"github.com/xyun1996/social_backend/services/social/internal/domain"
)

const (
	friendRequestPending  = "pending"
	friendRequestAccepted = "accepted"

	relationshipNone          = "none"
	relationshipFriend        = "friend"
	relationshipPendingInbox  = "pending_inbox"
	relationshipPendingOutbox = "pending_outbox"
	relationshipBlocked       = "blocked"
	relationshipBlockedBy     = "blocked_by"
)

// SocialService provides an in-memory prototype of friendship and blocklist flows.
type SocialService struct {
	requests     FriendRequestStore
	friendships  FriendshipStore
	blocks       BlockStore
	remarks      FriendRemarkStore
	now          func() time.Time
	newRequestID func() (string, error)
}

// NewSocialService constructs an in-memory social graph service.
func NewSocialService() *SocialService {
	return NewSocialServiceWithStores(
		newMemoryFriendRequestStore(),
		newMemoryFriendshipStore(),
		newMemoryBlockStore(),
		newMemoryFriendRemarkStore(),
	)
}

// NewSocialServiceWithStores constructs the service with injected persistence boundaries.
func NewSocialServiceWithStores(requests FriendRequestStore, friendships FriendshipStore, blocks BlockStore, remarks FriendRemarkStore) *SocialService {
	if requests == nil {
		requests = newMemoryFriendRequestStore()
	}
	if friendships == nil {
		friendships = newMemoryFriendshipStore()
	}
	if blocks == nil {
		blocks = newMemoryBlockStore()
	}
	if remarks == nil {
		remarks = newMemoryFriendRemarkStore()
	}
	return &SocialService{
		requests:    requests,
		friendships: friendships,
		blocks:      blocks,
		remarks:     remarks,
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
	if err := s.friendships.SaveFriendship(domain.FriendRelationship{PlayerID: request.FromPlayerID, FriendID: request.ToPlayerID}); err != nil {
		internal := apperrors.Internal()
		return domain.FriendRequest{}, &internal
	}
	if err := s.friendships.SaveFriendship(domain.FriendRelationship{PlayerID: request.ToPlayerID, FriendID: request.FromPlayerID}); err != nil {
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

// SetFriendRemark stores or replaces an optional remark for a confirmed friend.
func (s *SocialService) SetFriendRemark(playerID string, friendID string, remark string) (domain.FriendRemark, *apperrors.Error) {
	if playerID == "" || friendID == "" {
		err := apperrors.New("invalid_request", "player_id and friend_id are required", http.StatusBadRequest)
		return domain.FriendRemark{}, &err
	}
	if !s.isFriend(playerID, friendID) {
		err := apperrors.New("not_friend", "remarks require an accepted friendship", http.StatusConflict)
		return domain.FriendRemark{}, &err
	}

	record := domain.FriendRemark{
		PlayerID:  playerID,
		FriendID:  friendID,
		Remark:    strings.TrimSpace(remark),
		UpdatedAt: s.now(),
	}
	if err := s.remarks.SaveRemark(record); err != nil {
		internal := apperrors.Internal()
		return domain.FriendRemark{}, &internal
	}
	return record, nil
}

// ListFriendRemarks returns all stored friend remarks for a player.
func (s *SocialService) ListFriendRemarks(playerID string) ([]domain.FriendRemark, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", http.StatusBadRequest)
		return nil, &err
	}
	remarks, err := s.remarks.ListRemarks(playerID)
	if err != nil {
		internal := apperrors.Internal()
		return nil, &internal
	}
	return remarks, nil
}

// GetRelationship returns a richer point-to-point relationship view.
func (s *SocialService) GetRelationship(playerID string, targetPlayerID string) (domain.RelationshipSnapshot, *apperrors.Error) {
	if playerID == "" || targetPlayerID == "" {
		err := apperrors.New("invalid_request", "player_id and target_player_id are required", http.StatusBadRequest)
		return domain.RelationshipSnapshot{}, &err
	}
	return s.buildRelationshipSnapshot(playerID, targetPlayerID)
}

// ListRelationships returns the richer relationship view for every visible relationship edge.
func (s *SocialService) ListRelationships(playerID string, state string) ([]domain.RelationshipSnapshot, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", http.StatusBadRequest)
		return nil, &err
	}

	friends, appErr := s.ListFriends(playerID)
	if appErr != nil {
		return nil, appErr
	}
	blocks, appErr := s.ListBlocks(playerID)
	if appErr != nil {
		return nil, appErr
	}
	requests, appErr := s.ListFriendRequests(playerID, "all", friendRequestPending)
	if appErr != nil {
		return nil, appErr
	}

	targets := make(map[string]struct{})
	for _, friendID := range friends {
		targets[friendID] = struct{}{}
	}
	for _, blockedID := range blocks {
		targets[blockedID] = struct{}{}
	}
	for _, request := range requests {
		if request.FromPlayerID == playerID {
			targets[request.ToPlayerID] = struct{}{}
		} else {
			targets[request.FromPlayerID] = struct{}{}
		}
	}

	remarks, appErr := s.ListFriendRemarks(playerID)
	if appErr != nil {
		return nil, appErr
	}
	for _, remark := range remarks {
		targets[remark.FriendID] = struct{}{}
	}

	list := make([]domain.RelationshipSnapshot, 0, len(targets))
	for targetID := range targets {
		relationship, appErr := s.buildRelationshipSnapshot(playerID, targetID)
		if appErr != nil {
			return nil, appErr
		}
		if state != "" && relationship.State != state {
			continue
		}
		list = append(list, relationship)
	}

	slices.SortFunc(list, func(a domain.RelationshipSnapshot, b domain.RelationshipSnapshot) int {
		if !a.UpdatedAt.Equal(b.UpdatedAt) {
			if a.UpdatedAt.After(b.UpdatedAt) {
				return -1
			}
			return 1
		}
		switch {
		case a.TargetPlayerID < b.TargetPlayerID:
			return -1
		case a.TargetPlayerID > b.TargetPlayerID:
			return 1
		default:
			return 0
		}
	})
	return list, nil
}

// GetPendingSummary returns inbox/outbox pending request aggregation for a player.
func (s *SocialService) GetPendingSummary(playerID string) (domain.PendingSummary, *apperrors.Error) {
	if playerID == "" {
		err := apperrors.New("invalid_request", "player_id is required", http.StatusBadRequest)
		return domain.PendingSummary{}, &err
	}

	inboxRequests, appErr := s.ListFriendRequests(playerID, "inbox", friendRequestPending)
	if appErr != nil {
		return domain.PendingSummary{}, appErr
	}
	outboxRequests, appErr := s.ListFriendRequests(playerID, "outbox", friendRequestPending)
	if appErr != nil {
		return domain.PendingSummary{}, appErr
	}

	inbox := make([]string, 0, len(inboxRequests))
	for _, request := range inboxRequests {
		inbox = append(inbox, request.FromPlayerID)
	}
	outbox := make([]string, 0, len(outboxRequests))
	for _, request := range outboxRequests {
		outbox = append(outbox, request.ToPlayerID)
	}

	return domain.PendingSummary{
		PlayerID:     playerID,
		Inbox:        inbox,
		Outbox:       outbox,
		InboxCount:   len(inbox),
		OutboxCount:  len(outbox),
		TotalPending: len(inbox) + len(outbox),
	}, nil
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

	block := domain.BlockRelationship{PlayerID: playerID, BlockedID: blockedID, CreatedAt: s.now()}
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

func (s *SocialService) buildRelationshipSnapshot(playerID string, targetPlayerID string) (domain.RelationshipSnapshot, *apperrors.Error) {
	requests, err := s.requests.ListFriendRequests()
	if err != nil {
		internal := apperrors.Internal()
		return domain.RelationshipSnapshot{}, &internal
	}
	friend := s.isFriend(playerID, targetPlayerID)
	blocked := s.isBlocked(playerID, targetPlayerID)
	blockedBy := s.isBlocked(targetPlayerID, playerID)
	pendingInbox := false
	pendingOutbox := false
	updatedAt := time.Time{}

	for _, request := range requests {
		if request.Status != friendRequestPending {
			continue
		}
		switch {
		case request.FromPlayerID == targetPlayerID && request.ToPlayerID == playerID:
			pendingInbox = true
			if request.CreatedAt.After(updatedAt) {
				updatedAt = request.CreatedAt
			}
		case request.FromPlayerID == playerID && request.ToPlayerID == targetPlayerID:
			pendingOutbox = true
			if request.CreatedAt.After(updatedAt) {
				updatedAt = request.CreatedAt
			}
		}
	}

	remark := ""
	if record, ok, err := s.remarks.GetRemark(playerID, targetPlayerID); err == nil && ok {
		remark = record.Remark
		if record.UpdatedAt.After(updatedAt) {
			updatedAt = record.UpdatedAt
		}
	}
	reverseRemark := ""
	if record, ok, err := s.remarks.GetRemark(targetPlayerID, playerID); err == nil && ok {
		reverseRemark = record.Remark
		if record.UpdatedAt.After(updatedAt) {
			updatedAt = record.UpdatedAt
		}
	}
	if updatedAt.IsZero() {
		updatedAt = s.now()
	}

	state := relationshipNone
	switch {
	case blocked:
		state = relationshipBlocked
	case blockedBy:
		state = relationshipBlockedBy
	case friend:
		state = relationshipFriend
	case pendingInbox:
		state = relationshipPendingInbox
	case pendingOutbox:
		state = relationshipPendingOutbox
	}

	return domain.RelationshipSnapshot{
		PlayerID:         playerID,
		TargetPlayerID:   targetPlayerID,
		State:            state,
		IsFriend:         friend,
		HasPendingInbox:  pendingInbox,
		HasPendingOutbox: pendingOutbox,
		IsBlocked:        blocked,
		IsBlockedBy:      blockedBy,
		Remark:           remark,
		ReverseRemark:    reverseRemark,
		UpdatedAt:        updatedAt,
	}, nil
}

func (s *SocialService) String() string {
	requests, err := s.requests.ListFriendRequests()
	if err != nil {
		return "social-service(requests=unknown)"
	}
	return fmt.Sprintf("social-service(requests=%d)", len(requests))
}
