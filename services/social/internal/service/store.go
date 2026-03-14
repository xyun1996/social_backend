package service

import (
	"slices"
	"sync"

	"github.com/xyun1996/social_backend/services/social/internal/domain"
)

// FriendRequestStore persists directed friend request lifecycle state.
type FriendRequestStore interface {
	ListFriendRequests() ([]domain.FriendRequest, error)
	SaveFriendRequest(request domain.FriendRequest) error
	GetFriendRequest(requestID string) (domain.FriendRequest, bool, error)
}

// FriendshipStore persists accepted bidirectional relationships.
type FriendshipStore interface {
	ListFriends(playerID string) ([]string, error)
	SaveFriendship(relationship domain.FriendRelationship) error
}

// BlockStore persists point-to-point block relationships.
type BlockStore interface {
	ListBlocks(playerID string) ([]string, error)
	SaveBlock(block domain.BlockRelationship) error
}

// FriendRemarkStore persists optional friend metadata.
type FriendRemarkStore interface {
	ListRemarks(playerID string) ([]domain.FriendRemark, error)
	SaveRemark(remark domain.FriendRemark) error
	GetRemark(playerID string, friendID string) (domain.FriendRemark, bool, error)
}

type memoryFriendRequestStore struct {
	mu       sync.RWMutex
	requests map[string]domain.FriendRequest
}

func newMemoryFriendRequestStore() *memoryFriendRequestStore {
	return &memoryFriendRequestStore{
		requests: make(map[string]domain.FriendRequest),
	}
}

func (s *memoryFriendRequestStore) ListFriendRequests() ([]domain.FriendRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	requests := make([]domain.FriendRequest, 0, len(s.requests))
	for _, request := range s.requests {
		requests = append(requests, request)
	}
	return requests, nil
}

func (s *memoryFriendRequestStore) SaveFriendRequest(request domain.FriendRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requests[request.ID] = request
	return nil
}

func (s *memoryFriendRequestStore) GetFriendRequest(requestID string) (domain.FriendRequest, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	request, ok := s.requests[requestID]
	return request, ok, nil
}

type memoryFriendshipStore struct {
	mu          sync.RWMutex
	friendships map[string]map[string]struct{}
}

func newMemoryFriendshipStore() *memoryFriendshipStore {
	return &memoryFriendshipStore{
		friendships: make(map[string]map[string]struct{}),
	}
}

func (s *memoryFriendshipStore) ListFriends(playerID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	friends := make([]string, 0, len(s.friendships[playerID]))
	for friendID := range s.friendships[playerID] {
		friends = append(friends, friendID)
	}
	slices.Sort(friends)
	return friends, nil
}

func (s *memoryFriendshipStore) SaveFriendship(relationship domain.FriendRelationship) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.friendships[relationship.PlayerID] == nil {
		s.friendships[relationship.PlayerID] = make(map[string]struct{})
	}
	s.friendships[relationship.PlayerID][relationship.FriendID] = struct{}{}
	return nil
}

type memoryBlockStore struct {
	mu     sync.RWMutex
	blocks map[string]map[string]domain.BlockRelationship
}

func newMemoryBlockStore() *memoryBlockStore {
	return &memoryBlockStore{
		blocks: make(map[string]map[string]domain.BlockRelationship),
	}
}

func (s *memoryBlockStore) ListBlocks(playerID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	blocked := make([]string, 0, len(s.blocks[playerID]))
	for blockedID := range s.blocks[playerID] {
		blocked = append(blocked, blockedID)
	}
	slices.Sort(blocked)
	return blocked, nil
}

func (s *memoryBlockStore) SaveBlock(block domain.BlockRelationship) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.blocks[block.PlayerID] == nil {
		s.blocks[block.PlayerID] = make(map[string]domain.BlockRelationship)
	}
	s.blocks[block.PlayerID][block.BlockedID] = block
	return nil
}

type memoryFriendRemarkStore struct {
	mu      sync.RWMutex
	remarks map[string]map[string]domain.FriendRemark
}

func newMemoryFriendRemarkStore() *memoryFriendRemarkStore {
	return &memoryFriendRemarkStore{remarks: make(map[string]map[string]domain.FriendRemark)}
}

func (s *memoryFriendRemarkStore) ListRemarks(playerID string) ([]domain.FriendRemark, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	remarks := make([]domain.FriendRemark, 0, len(s.remarks[playerID]))
	for _, remark := range s.remarks[playerID] {
		remarks = append(remarks, remark)
	}
	slices.SortFunc(remarks, func(a domain.FriendRemark, b domain.FriendRemark) int {
		switch {
		case a.FriendID < b.FriendID:
			return -1
		case a.FriendID > b.FriendID:
			return 1
		default:
			return 0
		}
	})
	return remarks, nil
}

func (s *memoryFriendRemarkStore) SaveRemark(remark domain.FriendRemark) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.remarks[remark.PlayerID] == nil {
		s.remarks[remark.PlayerID] = make(map[string]domain.FriendRemark)
	}
	s.remarks[remark.PlayerID][remark.FriendID] = remark
	return nil
}

func (s *memoryFriendRemarkStore) GetRemark(playerID string, friendID string) (domain.FriendRemark, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	remarks := s.remarks[playerID]
	remark, ok := remarks[friendID]
	return remark, ok, nil
}
