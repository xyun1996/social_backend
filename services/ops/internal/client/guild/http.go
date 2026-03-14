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
	return &HTTPClient{baseURL: strings.TrimRight(baseURL, "/"), client: &http.Client{}}
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
	progression, appErr := c.getProgression(ctx, guildID)
	if appErr != nil {
		return opsservice.GuildSnapshot{}, appErr
	}
	contributions, appErr := c.getContributions(ctx, guildID)
	if appErr != nil {
		return opsservice.GuildSnapshot{}, appErr
	}
	rewards, appErr := c.getRewards(ctx, guildID)
	if appErr != nil {
		return opsservice.GuildSnapshot{}, appErr
	}
	activityInstances := make([]opsservice.GuildActivityInstance, 0)
	for _, template := range []string{"sign_in", "donate", "guild_task"} {
		records, appErr := c.getInstances(ctx, guildID, template)
		if appErr != nil && appErr.Code != "not_found" {
			return opsservice.GuildSnapshot{}, appErr
		}
		activityInstances = append(activityInstances, records...)
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
		Level:                 progression.Level,
		Experience:            progression.Experience,
		NextLevelXP:           progression.NextLevelXP,
		Contributions:         contributions,
		ActivityInstances:     activityInstances,
		RewardRecords:         rewards,
	}, nil
}

// GetGuildByPlayer fetches the current guild membership for a player.
func (c *HTTPClient) GetGuildByPlayer(ctx context.Context, playerID string) (opsservice.GuildSnapshot, *apperrors.Error) {
	record, appErr := c.getGuildByPlayer(ctx, playerID)
	if appErr != nil {
		return opsservice.GuildSnapshot{}, appErr
	}
	return c.GetGuildSnapshot(ctx, record.ID)
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

type progressionSnapshot struct {
	Level       int `json:"level"`
	Experience  int `json:"experience"`
	NextLevelXP int `json:"next_level_xp"`
}

type contributionsSnapshot struct {
	Contributions []opsservice.GuildContribution `json:"contributions"`
}

type rewardsSnapshot struct {
	Rewards []opsservice.GuildRewardRecord `json:"rewards"`
}

type instancesSnapshot struct {
	Instances []opsservice.GuildActivityInstance `json:"instances"`
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
	ID string `json:"id"`
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
	return getJSON[memberSnapshot](c, ctx, "/v1/guilds/"+guildID+"/members")
}

func (c *HTTPClient) getLogs(ctx context.Context, guildID string) (logSnapshot, *apperrors.Error) {
	return getJSON[logSnapshot](c, ctx, "/v1/guilds/"+guildID+"/logs")
}

func (c *HTTPClient) getProgression(ctx context.Context, guildID string) (progressionSnapshot, *apperrors.Error) {
	return getJSON[progressionSnapshot](c, ctx, "/v1/guilds/"+guildID+"/progression")
}

func (c *HTTPClient) getContributions(ctx context.Context, guildID string) ([]opsservice.GuildContribution, *apperrors.Error) {
	record, appErr := getJSON[contributionsSnapshot](c, ctx, "/v1/guilds/"+guildID+"/contributions")
	if appErr != nil {
		return nil, appErr
	}
	return record.Contributions, nil
}

func (c *HTTPClient) getRewards(ctx context.Context, guildID string) ([]opsservice.GuildRewardRecord, *apperrors.Error) {
	record, appErr := getJSON[rewardsSnapshot](c, ctx, "/v1/guilds/"+guildID+"/rewards")
	if appErr != nil {
		return nil, appErr
	}
	return record.Rewards, nil
}

func (c *HTTPClient) getInstances(ctx context.Context, guildID string, templateKey string) ([]opsservice.GuildActivityInstance, *apperrors.Error) {
	record, appErr := getJSON[instancesSnapshot](c, ctx, "/v1/guilds/"+guildID+"/activities/"+templateKey+"/instances")
	if appErr != nil {
		return nil, appErr
	}
	return record.Instances, nil
}

func getJSON[T any](c *HTTPClient, ctx context.Context, path string) (T, *apperrors.Error) {
	var zero T
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		internal := apperrors.Internal()
		return zero, &internal
	}
	resp, err := c.client.Do(req)
	if err != nil {
		badGateway := apperrors.New("guild_unavailable", "guild service is unavailable", http.StatusBadGateway)
		return zero, &badGateway
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return zero, decodeGuildError(resp)
	}
	var record T
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		badGateway := apperrors.New("guild_invalid_response", "guild service returned an invalid response", http.StatusBadGateway)
		return zero, &badGateway
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
