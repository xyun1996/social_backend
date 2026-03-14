package service

import "testing"

func TestSendAndAcceptFriendRequest(t *testing.T) {
	t.Parallel()
	svc := NewSocialService()
	request, err := svc.SendFriendRequest("p1", "p2")
	if err != nil { t.Fatalf("send request returned error: %+v", err) }
	accepted, acceptErr := svc.AcceptFriendRequest(request.ID, "p2")
	if acceptErr != nil { t.Fatalf("accept returned error: %+v", acceptErr) }
	if accepted.Status != "accepted" { t.Fatalf("unexpected accepted status: %q", accepted.Status) }
	friends, listErr := svc.ListFriends("p1")
	if listErr != nil { t.Fatalf("list friends returned error: %+v", listErr) }
	if len(friends) != 1 || friends[0] != "p2" { t.Fatalf("unexpected friends list: %+v", friends) }
}

func TestBlockPreventsFriendRequest(t *testing.T) {
	t.Parallel()
	svc := NewSocialService()
	if _, err := svc.BlockPlayer("p2", "p1"); err != nil { t.Fatalf("block returned error: %+v", err) }
	if _, requestErr := svc.SendFriendRequest("p1", "p2"); requestErr == nil { t.Fatalf("expected blocked relationship to reject friend request") }
}

func TestListFriendRequests(t *testing.T) {
	t.Parallel()
	svc := NewSocialService()
	if _, err := svc.SendFriendRequest("p1", "p2"); err != nil { t.Fatalf("send request returned error: %+v", err) }
	if _, err := svc.SendFriendRequest("p3", "p2"); err != nil { t.Fatalf("send request returned error: %+v", err) }
	inbox, err := svc.ListFriendRequests("p2", "inbox", friendRequestPending)
	if err != nil { t.Fatalf("list friend requests returned error: %+v", err) }
	if len(inbox) != 2 { t.Fatalf("unexpected inbox size: %d", len(inbox)) }
}

func TestRelationshipSnapshotAndRemarks(t *testing.T) {
	t.Parallel()
	svc := NewSocialService()
	request, _ := svc.SendFriendRequest("p1", "p2")
	_, _ = svc.AcceptFriendRequest(request.ID, "p2")
	if _, err := svc.SetFriendRemark("p1", "p2", "raid lead"); err != nil { t.Fatalf("set friend remark returned error: %+v", err) }
	relationship, err := svc.GetRelationship("p1", "p2")
	if err != nil { t.Fatalf("get relationship returned error: %+v", err) }
	if !relationship.IsFriend || relationship.Remark != "raid lead" || relationship.State != relationshipFriend { t.Fatalf("unexpected relationship snapshot: %+v", relationship) }
}

func TestPendingSummary(t *testing.T) {
	t.Parallel()
	svc := NewSocialService()
	_, _ = svc.SendFriendRequest("p1", "p2")
	_, _ = svc.SendFriendRequest("p3", "p2")
	summary, err := svc.GetPendingSummary("p2")
	if err != nil { t.Fatalf("pending summary returned error: %+v", err) }
	if summary.InboxCount != 2 || summary.TotalPending != 2 { t.Fatalf("unexpected pending summary: %+v", summary) }
}

func TestSocialServiceWithInjectedStores(t *testing.T) {
	t.Parallel()
	requests := newMemoryFriendRequestStore()
	friendships := newMemoryFriendshipStore()
	blocks := newMemoryBlockStore()
	remarks := newMemoryFriendRemarkStore()
	svc := NewSocialServiceWithStores(requests, friendships, blocks, remarks)
	request, err := svc.SendFriendRequest("p1", "p2")
	if err != nil { t.Fatalf("send request returned error: %+v", err) }
	if _, acceptErr := svc.AcceptFriendRequest(request.ID, "p2"); acceptErr != nil { t.Fatalf("accept returned error: %+v", acceptErr) }
	if _, remarkErr := svc.SetFriendRemark("p1", "p2", "tank"); remarkErr != nil { t.Fatalf("remark returned error: %+v", remarkErr) }
	friends, listErr := friendships.ListFriends("p1")
	if listErr != nil { t.Fatalf("friendship store list failed: %v", listErr) }
	if len(friends) != 1 || friends[0] != "p2" { t.Fatalf("unexpected stored friends: %+v", friends) }
	if stored, ok, _ := remarks.GetRemark("p1", "p2"); !ok || stored.Remark != "tank" { t.Fatalf("unexpected stored remark: %+v ok=%v", stored, ok) }
}
