package service

import (
	"net/http"
	"slices"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	"github.com/xyun1996/social_backend/services/chat/internal/domain"
)

func defaultConversationGovernance(conversation domain.Conversation) domain.Conversation {
	if conversation.SendPolicy == "" {
		switch conversation.Kind {
		case kindSystem:
			conversation.SendPolicy = sendPolicySystemOnly
		case kindWorld:
			conversation.SendPolicy = sendPolicyModerated
		default:
			conversation.SendPolicy = sendPolicyMembers
		}
	}
	if conversation.VisibilityPolicy == "" {
		switch conversation.Kind {
		case kindWorld, kindCustom:
			conversation.VisibilityPolicy = visibilityPublicRead
		default:
			conversation.VisibilityPolicy = visibilityMembers
		}
	}
	if conversation.ModerationMode == "" {
		switch conversation.Kind {
		case kindWorld, kindSystem, kindCustom:
			conversation.ModerationMode = moderationManaged
		default:
			conversation.ModerationMode = moderationOpen
		}
	}
	slices.Sort(conversation.ModeratorIDs)
	slices.Sort(conversation.MutedPlayerIDs)
	return conversation
}

func isModerator(conversation domain.Conversation, playerID string) bool {
	conversation = defaultConversationGovernance(conversation)
	if playerID == "system" {
		return true
	}
	return slices.Contains(conversation.ModeratorIDs, playerID)
}

func isMuted(conversation domain.Conversation, playerID string) bool {
	conversation = defaultConversationGovernance(conversation)
	return slices.Contains(conversation.MutedPlayerIDs, playerID)
}

func (s *ChatService) GetConversationGovernance(conversationID string) (domain.ChannelDescriptor, *apperrors.Error) {
	return s.GetChannelDescriptor(conversationID)
}

func (s *ChatService) UpdateConversationGovernance(conversationID string, actorPlayerID string, sendPolicy string, visibilityPolicy string) (domain.ChannelDescriptor, *apperrors.Error) {
	conversation, appErr := s.loadConversationForGovernance(conversationID)
	if appErr != nil {
		return domain.ChannelDescriptor{}, appErr
	}
	if appErr := validateGovernanceActor(conversation, actorPlayerID); appErr != nil {
		return domain.ChannelDescriptor{}, appErr
	}
	if sendPolicy != "" {
		if sendPolicy != sendPolicyMembers && sendPolicy != sendPolicyModerated && sendPolicy != sendPolicySystemOnly {
			err := apperrors.New("invalid_request", "unsupported send_policy", http.StatusBadRequest)
			return domain.ChannelDescriptor{}, &err
		}
		conversation.SendPolicy = sendPolicy
	}
	if visibilityPolicy != "" {
		if visibilityPolicy != visibilityMembers && visibilityPolicy != visibilityPublicRead {
			err := apperrors.New("invalid_request", "unsupported visibility_policy", http.StatusBadRequest)
			return domain.ChannelDescriptor{}, &err
		}
		conversation.VisibilityPolicy = visibilityPolicy
	}
	conversation = defaultConversationGovernance(conversation)
	if err := s.conversations.SaveConversation(conversation); err != nil {
		internal := apperrors.Internal()
		return domain.ChannelDescriptor{}, &internal
	}
	return buildChannelDescriptor(conversation), nil
}

func (s *ChatService) SetConversationModerator(conversationID string, actorPlayerID string, targetPlayerID string, enabled bool) (domain.ChannelDescriptor, *apperrors.Error) {
	conversation, appErr := s.loadConversationForGovernance(conversationID)
	if appErr != nil {
		return domain.ChannelDescriptor{}, appErr
	}
	if appErr := validateGovernanceActor(conversation, actorPlayerID); appErr != nil {
		return domain.ChannelDescriptor{}, appErr
	}
	conversation.ModeratorIDs = applyMembershipToggle(conversation.ModeratorIDs, targetPlayerID, enabled)
	conversation = defaultConversationGovernance(conversation)
	if err := s.conversations.SaveConversation(conversation); err != nil {
		internal := apperrors.Internal()
		return domain.ChannelDescriptor{}, &internal
	}
	return buildChannelDescriptor(conversation), nil
}

func (s *ChatService) SetConversationMute(conversationID string, actorPlayerID string, targetPlayerID string, muted bool) (domain.ChannelDescriptor, *apperrors.Error) {
	conversation, appErr := s.loadConversationForGovernance(conversationID)
	if appErr != nil {
		return domain.ChannelDescriptor{}, appErr
	}
	if appErr := validateGovernanceActor(conversation, actorPlayerID); appErr != nil {
		return domain.ChannelDescriptor{}, appErr
	}
	conversation.MutedPlayerIDs = applyMembershipToggle(conversation.MutedPlayerIDs, targetPlayerID, muted)
	conversation = defaultConversationGovernance(conversation)
	if err := s.conversations.SaveConversation(conversation); err != nil {
		internal := apperrors.Internal()
		return domain.ChannelDescriptor{}, &internal
	}
	return buildChannelDescriptor(conversation), nil
}

func (s *ChatService) loadConversationForGovernance(conversationID string) (domain.Conversation, *apperrors.Error) {
	if conversationID == "" {
		err := apperrors.New("invalid_request", "conversation_id is required", http.StatusBadRequest)
		return domain.Conversation{}, &err
	}
	conversation, ok, err := s.conversations.GetConversation(conversationID)
	if err != nil {
		internal := apperrors.Internal()
		return domain.Conversation{}, &internal
	}
	if !ok {
		err := apperrors.New("not_found", "conversation not found", http.StatusNotFound)
		return domain.Conversation{}, &err
	}
	return defaultConversationGovernance(conversation), nil
}

func validateGovernanceActor(conversation domain.Conversation, actorPlayerID string) *apperrors.Error {
	if actorPlayerID == "system" {
		return nil
	}
	if isModerator(conversation, actorPlayerID) {
		return nil
	}
	if hasMember(conversation.MemberPlayerIDs, actorPlayerID) && len(conversation.ModeratorIDs) == 0 && (conversation.Kind == kindCustom || conversation.Kind == kindWorld || conversation.Kind == kindGroup) {
		return nil
	}
	err := apperrors.New("forbidden", "actor is not allowed to manage conversation governance", http.StatusForbidden)
	return &err
}

func applyMembershipToggle(values []string, target string, enabled bool) []string {
	if target == "" {
		return append([]string(nil), values...)
	}
	result := make([]string, 0, len(values)+1)
	seen := false
	for _, value := range values {
		if value == target {
			seen = true
			if enabled {
				result = append(result, value)
			}
			continue
		}
		result = append(result, value)
	}
	if enabled && !seen {
		result = append(result, target)
	}
	slices.Sort(result)
	return result
}
