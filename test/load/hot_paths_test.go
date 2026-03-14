package load

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	chattestkit "github.com/xyun1996/social_backend/services/chat/testkit"
	gatewaytestkit "github.com/xyun1996/social_backend/services/gateway/testkit"
	guildtestkit "github.com/xyun1996/social_backend/services/guild/testkit"
	identitytestkit "github.com/xyun1996/social_backend/services/identity/testkit"
	invitetestkit "github.com/xyun1996/social_backend/services/invite/testkit"
	partytestkit "github.com/xyun1996/social_backend/services/party/testkit"
	presencetestkit "github.com/xyun1996/social_backend/services/presence/testkit"
	workertestkit "github.com/xyun1996/social_backend/services/worker/testkit"
)

func TestHotPathSmoke(t *testing.T) {
	t.Run("gateway handshake", func(t *testing.T) {
		identity := identitytestkit.NewServer()
		defer identity.Close()
		presence := presencetestkit.NewServer()
		defer presence.Close()
		chat := chattestkit.NewServer("", "")
		defer chat.Close()
		gateway := gatewaytestkit.NewServer(identity.URL(), presence.URL(), chat.URL())
		defer gateway.Close()

		tokenPair := postJSON(t, identity.URL()+"/v1/auth/login", map[string]any{
			"account_id": "account-load",
			"player_id":  "player-load",
		})

		accessToken := tokenPair["access_token"].(string)
		for i := 0; i < 20; i++ {
			postJSON(t, gateway.URL()+"/v1/realtime/handshake", map[string]any{
				"access_token": accessToken,
				"session_id":   "session-load-" + strconv.Itoa(i),
				"realm_id":     "realm-1",
				"location":     "lobby",
			})
		}
	})

	t.Run("chat send and replay", func(t *testing.T) {
		chat := chattestkit.NewServer("", "")
		defer chat.Close()

		created := postJSON(t, chat.URL()+"/v1/conversations", map[string]any{
			"kind":              "private",
			"member_player_ids": []string{"p1", "p2"},
		})
		conversationID := created["id"].(string)

		for i := 0; i < 25; i++ {
			postJSON(t, chat.URL()+"/v1/conversations/"+conversationID+"/messages", map[string]any{
				"sender_player_id": "p1",
				"body":             "load message",
			})
		}

		getJSON(t, chat.URL()+"/v1/conversations/"+conversationID+"/messages?player_id=p1&after_seq=0&limit=50")
	})

	t.Run("worker retry path", func(t *testing.T) {
		worker := workertestkit.NewServer()
		defer worker.Close()

		for i := 0; i < 10; i++ {
			postJSON(t, worker.URL()+"/v1/jobs", map[string]any{
				"type":         "load.test",
				"payload":      "{}",
				"max_attempts": 3,
			})
			claimed := postJSON(t, worker.URL()+"/v1/jobs/claim", map[string]any{
				"worker_id": "load-worker",
				"type":      "load.test",
			})
			postJSON(t, worker.URL()+"/v1/jobs/"+claimed["id"].(string)+"/fail", map[string]any{
				"worker_id":  "load-worker",
				"last_error": "synthetic failure",
			})
		}
	})

	t.Run("party reads", func(t *testing.T) {
		invite := invitetestkit.NewServer("")
		defer invite.Close()
		presence := presencetestkit.NewServer()
		defer presence.Close()
		party := partytestkit.NewServer(invite.URL(), presence.URL())
		defer party.Close()

		created := postJSON(t, party.URL()+"/v1/parties", map[string]any{
			"leader_id": "party-leader",
		})
		partyID := created["id"].(string)

		for i := 0; i < 25; i++ {
			getJSON(t, party.URL()+"/v1/parties/"+partyID)
		}
	})

	t.Run("guild progression reads", func(t *testing.T) {
		invite := invitetestkit.NewServer("")
		defer invite.Close()
		presence := presencetestkit.NewServer()
		defer presence.Close()
		chat := chattestkit.NewServer("", "")
		defer chat.Close()
		worker := workertestkit.NewServer()
		defer worker.Close()
		guild := guildtestkit.NewServer(invite.URL(), presence.URL(), chat.URL(), worker.URL())
		defer guild.Close()

		created := postJSON(t, guild.URL()+"/v1/guilds", map[string]any{
			"name":     "Load Guild",
			"owner_id": "guild-owner",
		})
		guildID := created["id"].(string)

		postJSON(t, guild.URL()+"/v1/guilds/"+guildID+"/activities/donate/submit", map[string]any{
			"actor_player_id": "guild-owner",
			"idempotency_key": "load-seed-1",
			"source_type":     "load-test",
		})

		for i := 0; i < 20; i++ {
			getJSON(t, guild.URL()+"/v1/guilds/"+guildID+"/progression")
			getJSON(t, guild.URL()+"/v1/guilds/"+guildID+"/contributions")
		}
	})
}

func postJSON(t *testing.T, url string, body map[string]any) map[string]any {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("post %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("post %s returned status %d", url, resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response %s: %v", url, err)
	}

	return result
}

func getJSON(t *testing.T, url string) map[string]any {
	t.Helper()

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("get %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("get %s returned status %d", url, resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response %s: %v", url, err)
	}

	return result
}
