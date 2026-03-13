package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	chattestkit "github.com/xyun1996/social_backend/services/chat/testkit"
	gatewaytestkit "github.com/xyun1996/social_backend/services/gateway/testkit"
	identitytestkit "github.com/xyun1996/social_backend/services/identity/testkit"
	invitetestkit "github.com/xyun1996/social_backend/services/invite/testkit"
	presencetestkit "github.com/xyun1996/social_backend/services/presence/testkit"
	workertestkit "github.com/xyun1996/social_backend/services/worker/testkit"
)

func TestInviteWorkerExpiryFlow(t *testing.T) {
	t.Parallel()

	worker := workertestkit.NewServer()
	defer worker.Close()

	invite := invitetestkit.NewServer(worker.URL())
	defer invite.Close()

	worker.RegisterInviteExpireHandler(invite.URL())

	var created struct {
		ID string `json:"id"`
	}
	postJSON(t, invite.URL()+"/v1/invites", map[string]any{
		"domain":         "party",
		"resource_id":    "party-1",
		"from_player_id": "p1",
		"to_player_id":   "p2",
		"ttl_seconds":    60,
	}, &created)

	result, err := worker.ExecuteUntilEmpty(context.Background(), "worker-a", "invite.expire", 10)
	if err != nil {
		t.Fatalf("execute until empty failed: %v", err)
	}
	if result.Completed != 1 {
		t.Fatalf("unexpected worker result: %+v", result)
	}

	var expired struct {
		Status string `json:"status"`
	}
	getJSON(t, invite.URL()+"/v1/invites/"+created.ID, &expired)
	if expired.Status != "expired" {
		t.Fatalf("expected invite to be expired, got %+v", expired)
	}
}

func TestChatWorkerOfflineDeliveryFlow(t *testing.T) {
	t.Parallel()

	worker := workertestkit.NewServer()
	defer worker.Close()

	chat := chattestkit.NewServer("", worker.URL())
	defer chat.Close()

	worker.RegisterChatOfflineDeliveryHandler(chat.URL())

	var conversation struct {
		ID string `json:"id"`
	}
	postJSON(t, chat.URL()+"/v1/conversations", map[string]any{
		"kind":              "private",
		"member_player_ids": []string{"p1", "p2"},
	}, &conversation)

	postJSON(t, chat.URL()+"/v1/conversations/"+conversation.ID+"/messages", map[string]any{
		"sender_player_id": "p1",
		"body":             "hello",
	}, nil)

	result, err := worker.ExecuteUntilEmpty(context.Background(), "worker-a", "chat.offline_delivery", 10)
	if err != nil {
		t.Fatalf("execute until empty failed: %v", err)
	}
	if result.Completed != 1 {
		t.Fatalf("unexpected worker result: %+v", result)
	}
	if chat.OfflineDeliveryCount(conversation.ID) != 1 {
		t.Fatalf("expected one offline delivery receipt")
	}
}

func TestGatewayChatDeliverySessionInboxFlow(t *testing.T) {
	t.Parallel()

	identity := identitytestkit.NewServer()
	defer identity.Close()

	presence := presencetestkit.NewServer()
	defer presence.Close()

	chat := chattestkit.NewServer(presence.URL(), "")
	defer chat.Close()

	gateway := gatewaytestkit.NewServer(identity.URL(), presence.URL(), chat.URL())
	defer gateway.Close()

	var tokenPair struct {
		AccessToken string `json:"access_token"`
	}
	postJSON(t, identity.URL()+"/v1/auth/login", map[string]any{
		"account_id": "a2",
		"player_id":  "p2",
	}, &tokenPair)

	var conversation struct {
		ID string `json:"id"`
	}
	postJSON(t, chat.URL()+"/v1/conversations", map[string]any{
		"kind":              "private",
		"member_player_ids": []string{"p1", "p2"},
	}, &conversation)

	postJSON(t, gateway.URL()+"/v1/realtime/handshake", map[string]any{
		"access_token": tokenPair.AccessToken,
		"session_id":   "sess-2",
		"realm_id":     "realm-1",
		"location":     "lobby",
	}, nil)

	postJSON(t, gateway.URL()+"/v1/realtime/chat/deliveries", map[string]any{
		"conversation_id":  conversation.ID,
		"sender_player_id": "p1",
		"message_id":       "msg-1",
		"seq":              1,
		"body":             "hello",
		"sent_at":          "2026-03-13T10:00:00Z",
	}, nil)

	var inbox struct {
		SessionID string `json:"session_id"`
		Count     int    `json:"count"`
	}
	getJSON(t, gateway.URL()+"/v1/realtime/sessions/sess-2/events", &inbox)
	if inbox.SessionID != "sess-2" || inbox.Count != 1 {
		t.Fatalf("unexpected inbox payload: %+v", inbox)
	}
}

func TestGatewayChatAckCompactsSessionInboxFlow(t *testing.T) {
	t.Parallel()

	identity := identitytestkit.NewServer()
	defer identity.Close()

	presence := presencetestkit.NewServer()
	defer presence.Close()

	chat := chattestkit.NewServer(presence.URL(), "")
	defer chat.Close()

	gateway := gatewaytestkit.NewServer(identity.URL(), presence.URL(), chat.URL())
	defer gateway.Close()

	var tokenPair struct {
		AccessToken string `json:"access_token"`
	}
	postJSON(t, identity.URL()+"/v1/auth/login", map[string]any{
		"account_id": "a2",
		"player_id":  "p2",
	}, &tokenPair)

	var conversation struct {
		ID string `json:"id"`
	}
	postJSON(t, chat.URL()+"/v1/conversations", map[string]any{
		"kind":              "private",
		"member_player_ids": []string{"p1", "p2"},
	}, &conversation)

	postJSON(t, gateway.URL()+"/v1/realtime/handshake", map[string]any{
		"access_token": tokenPair.AccessToken,
		"session_id":   "sess-2",
	}, nil)

	var message struct {
		ID  string `json:"id"`
		Seq int64  `json:"seq"`
	}
	postJSON(t, chat.URL()+"/v1/conversations/"+conversation.ID+"/messages", map[string]any{
		"sender_player_id": "p1",
		"body":             "hello",
	}, &message)

	postJSON(t, gateway.URL()+"/v1/realtime/chat/deliveries", map[string]any{
		"conversation_id":  conversation.ID,
		"sender_player_id": "p1",
		"message_id":       message.ID,
		"seq":              message.Seq,
		"body":             "hello",
		"sent_at":          "2026-03-13T10:00:00Z",
	}, nil)

	var ackResult struct {
		PrunedCount int `json:"pruned_count"`
	}
	postJSON(t, gateway.URL()+"/v1/realtime/sessions/sess-2/acks", map[string]any{
		"conversation_id": conversation.ID,
		"ack_seq":         1,
	}, &ackResult)
	if ackResult.PrunedCount != 1 {
		t.Fatalf("expected one pruned event, got %+v", ackResult)
	}

	var inbox struct {
		Count int `json:"count"`
	}
	getJSON(t, gateway.URL()+"/v1/realtime/sessions/sess-2/events", &inbox)
	if inbox.Count != 0 {
		t.Fatalf("expected compacted inbox, got %+v", inbox)
	}
}

func TestGatewayResumeTrimsBufferedInboxFlow(t *testing.T) {
	t.Parallel()

	identity := identitytestkit.NewServer()
	defer identity.Close()

	presence := presencetestkit.NewServer()
	defer presence.Close()

	chat := chattestkit.NewServer(presence.URL(), "")
	defer chat.Close()

	gateway := gatewaytestkit.NewServer(identity.URL(), presence.URL(), chat.URL())
	defer gateway.Close()

	var tokenPair struct {
		AccessToken string `json:"access_token"`
	}
	postJSON(t, identity.URL()+"/v1/auth/login", map[string]any{
		"account_id": "a2",
		"player_id":  "p2",
	}, &tokenPair)

	var conversation struct {
		ID string `json:"id"`
	}
	postJSON(t, chat.URL()+"/v1/conversations", map[string]any{
		"kind":              "private",
		"member_player_ids": []string{"p1", "p2"},
	}, &conversation)

	postJSON(t, gateway.URL()+"/v1/realtime/handshake", map[string]any{
		"access_token": tokenPair.AccessToken,
		"session_id":   "sess-2",
	}, nil)

	var message1 struct {
		ID  string `json:"id"`
		Seq int64  `json:"seq"`
	}
	postJSON(t, chat.URL()+"/v1/conversations/"+conversation.ID+"/messages", map[string]any{
		"sender_player_id": "p1",
		"body":             "hello",
	}, &message1)
	postJSON(t, gateway.URL()+"/v1/realtime/chat/deliveries", map[string]any{
		"conversation_id":  conversation.ID,
		"sender_player_id": "p1",
		"message_id":       message1.ID,
		"seq":              message1.Seq,
		"body":             "hello",
		"sent_at":          "2026-03-13T10:00:00Z",
	}, nil)

	var message2 struct {
		ID  string `json:"id"`
		Seq int64  `json:"seq"`
	}
	postJSON(t, chat.URL()+"/v1/conversations/"+conversation.ID+"/messages", map[string]any{
		"sender_player_id": "p1",
		"body":             "world",
	}, &message2)
	postJSON(t, gateway.URL()+"/v1/realtime/chat/deliveries", map[string]any{
		"conversation_id":  conversation.ID,
		"sender_player_id": "p1",
		"message_id":       message2.ID,
		"seq":              message2.Seq,
		"body":             "world",
		"sent_at":          "2026-03-13T10:00:01Z",
	}, nil)

	postJSON(t, gateway.URL()+"/v1/realtime/resume", map[string]any{
		"access_token":         tokenPair.AccessToken,
		"session_id":           "sess-2",
		"last_server_event_id": message1.ID + ":1",
	}, nil)

	var inbox struct {
		Count  int `json:"count"`
		Events []struct {
			EventID string `json:"event_id"`
		} `json:"events"`
	}
	getJSON(t, gateway.URL()+"/v1/realtime/sessions/sess-2/events", &inbox)
	if inbox.Count != 1 || inbox.Events[0].EventID != message2.ID+":2" {
		t.Fatalf("expected only the newer buffered event after resume, got %+v", inbox)
	}
}

func postJSON(t *testing.T, url string, payload any, out any) {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status from %s: got %d want %d", url, resp.StatusCode, http.StatusOK)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatalf("decode response: %v", err)
		}
	}
}

func getJSON(t *testing.T, url string, out any) {
	t.Helper()

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("get request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status from %s: got %d want %d", url, resp.StatusCode, http.StatusOK)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
