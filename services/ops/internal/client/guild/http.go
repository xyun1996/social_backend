package guild

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apperrors "github.com/xyun1996/social_backend/pkg/errors"
	opsservice "github.com/xyun1996/social_backend/services/ops/internal/service"
)

// HTTPClient calls the guild HTTP API for operator reads.
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient constructs a guild HTTP client.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{},
	}
}

// GetGuildSnapshot fetches the guild member snapshot.
func (c *HTTPClient) GetGuildSnapshot(ctx context.Context, guildID string) (opsservice.GuildSnapshot, *apperrors.Error) {
	guildRecord, appErr := c.getGuild(ctx, guildID)
	if appErr != nil {
		return opsservice.GuildSnapshot{}, appErr
	}
	memberRecord, appErr := c.getMembers(ctx, guildID)
	if appErr != nil {
		return opsservice.GuildSnapshot{}, appErr
	}
	logRecord, appErr := c.getLogs(ctx, guildID)
	if appErr != nil {
		return opsservice.GuildSnapshot{}, appErr
	}

	return opsservice.GuildSnapshot{
		GuildID:               guildRecord.ID,
		Name:                  guildRecord.Name,
		OwnerID:               guildRecord.OwnerID,
		Announcement:          guildRecord.Announcement,
		AnnouncementUpdatedAt: guildRecord.AnnouncementUpdatedAt,
		Count:                 memberRecord.Count,
		Members:               memberRecord.Members,
		LogCount:              logRecord.Count,
		Logs:                  logRecord.Logs,
	}, nil
}

// GetGuildByPlayer fetches the current guild membership for a player.
func (c *HTTPClient) GetGuildByPlayer(ctx context.Context, playerID string) (opsservice.GuildSnapshot, *apperrors.Error) {
	record, appErr := c.getGuildByPlayer(ctx, playerID)
	if appErr != nil {
		return opsservice.GuildSnapshot{}, appErr
	}
	logRecord, appErr := c.getLogs(ctx, record.ID)
	if appErr != nil {
		return opsservice.GuildSnapshot{}, appErr
	}

	return opsservice.GuildSnapshot{
		GuildID:               record.ID,
		Name:                  record.Name,
		OwnerID:               record.OwnerID,
		Announcement:          record.Announcement,
		AnnouncementUpdatedAt: record.AnnouncementUpdatedAt,
		Count:                 record.Count,
		Members:               record.Members,
		LogCount:              logRecord.Count,
		Logs:                  logRecord.Logs,
	}, nil
}

type guildAggregate struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	OwnerID               string `json:"owner_id"`
	Announcement          string `json:"announcement"`
	AnnouncementUpdatedAt string `json:"announcement_updated_at"`
}

type memberSnapshot struct {
	GuildID string                        `json:"guild_id"`
	Count   int                           `json:"count"`
	Members []opsservice.GuildMemberState `json:"members"`
}

type logSnapshot struct {
	GuildID string                     `json:"guild_id"`
	Count   int                        `json:"count"`
	Logs    []opsservice.GuildLogEntry `json:"logs"`
}

func (c *HTTPClient) getGuild(ctx context.Context, guildID string) (guildAggregate, *apperrors.Error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/guilds/"+guildID, nil)
	if err != nil {
		internal := apperrors.Internal()
		return guildAggregate{}, &internal
	}

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("guild_unavailable", "guild service is unavailable", http.StatusBadGateway)
		return guildAggregate{}, &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return guildAggregate{}, decodeGuildError(resp)
	}

	var record guildAggregate
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		badGateway := apperrors.New("guild_invalid_response", "guild service returned an invalid response", http.StatusBadGateway)
		return guildAggregate{}, &badGateway
	}
	return record, nil
}

type guildByPlayerSnapshot struct {
	ID                    string                        `json:"id"`
	Name                  string                        `json:"name"`
	OwnerID               string                        `json:"owner_id"`
	Announcement          string                        `json:"announcement"`
	AnnouncementUpdatedAt string                        `json:"announcement_updated_at"`
	Count                 int                           `json:"count"`
	Members               []opsservice.GuildMemberState `json:"members"`
}

func (c *HTTPClient) getGuildByPlayer(ctx context.Context, playerID string) (guildByPlayerSnapshot, *apperrors.Error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/guild-memberships/"+playerID, nil)
	if err != nil {
		internal := apperrors.Internal()
		return guildByPlayerSnapshot{}, &internal
	}

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("guild_unavailable", "guild service is unavailable", http.StatusBadGateway)
		return guildByPlayerSnapshot{}, &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return guildByPlayerSnapshot{}, decodeGuildError(resp)
	}

	var record guildByPlayerSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		badGateway := apperrors.New("guild_invalid_response", "guild service returned an invalid response", http.StatusBadGateway)
		return guildByPlayerSnapshot{}, &badGateway
	}
	return record, nil
}

func (c *HTTPClient) getMembers(ctx context.Context, guildID string) (memberSnapshot, *apperrors.Error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/guilds/"+guildID+"/members", nil)
	if err != nil {
		internal := apperrors.Internal()
		return memberSnapshot{}, &internal
	}

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("guild_unavailable", "guild service is unavailable", http.StatusBadGateway)
		return memberSnapshot{}, &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return memberSnapshot{}, decodeGuildError(resp)
	}

	var record memberSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		badGateway := apperrors.New("guild_invalid_response", "guild service returned an invalid response", http.StatusBadGateway)
		return memberSnapshot{}, &badGateway
	}
	return record, nil
}

func (c *HTTPClient) getLogs(ctx context.Context, guildID string) (logSnapshot, *apperrors.Error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/guilds/"+guildID+"/logs", nil)
	if err != nil {
		internal := apperrors.Internal()
		return logSnapshot{}, &internal
	}

	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("guild_unavailable", "guild service is unavailable", http.StatusBadGateway)
		return logSnapshot{}, &badGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return logSnapshot{}, decodeGuildError(resp)
	}

	var record logSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		badGateway := apperrors.New("guild_invalid_response", "guild service returned an invalid response", http.StatusBadGateway)
		return logSnapshot{}, &badGateway
	}
	return record, nil
}

func decodeGuildError(resp *http.Response) *apperrors.Error {
	var appErr apperrors.Error
	if decodeErr := json.NewDecoder(resp.Body).Decode(&appErr); decodeErr != nil {
		badGateway := apperrors.New("guild_invalid_response", "guild service returned an invalid response", http.StatusBadGateway)
		return &badGateway
	}
	appErr.Status = resp.StatusCode
	return &appErr
}

func (c *HTTPClient) String() string {
	return fmt.Sprintf("guild-http-client(%s)", c.baseURL)
}
