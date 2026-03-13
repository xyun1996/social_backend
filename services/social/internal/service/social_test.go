package service

import "testing"

func TestSendAndAcceptFriendRequest(t *testing.T) {
	t.Parallel()

	svc := NewSocialService()
	request, err := svc.SendFriendRequest("p1", "p2")
	if err != nil {
		t.Fatalf("send request returned error: %+v", err)
	}

	if request.Status != "pending" {
		t.Fatalf("unexpected request status: %q", request.Status)
	}

	accepted, acceptErr := svc.AcceptFriendRequest(request.ID, "p2")
	if acceptErr != nil {
		t.Fatalf("accept returned error: %+v", acceptErr)
	}

	if accepted.Status != "accepted" {
		t.Fatalf("unexpected accepted status: %q", accepted.Status)
	}

	friends, listErr := svc.ListFriends("p1")
	if listErr != nil {
		t.Fatalf("list friends returned error: %+v", listErr)
	}

	if len(friends) != 1 || friends[0] != "p2" {
		t.Fatalf("unexpected friends list: %+v", friends)
	}
}

func TestBlockPreventsFriendRequest(t *testing.T) {
	t.Parallel()

	svc := NewSocialService()
	if _, err := svc.BlockPlayer("p2", "p1"); err != nil {
		t.Fatalf("block returned error: %+v", err)
	}

	if _, requestErr := svc.SendFriendRequest("p1", "p2"); requestErr == nil {
		t.Fatalf("expected blocked relationship to reject friend request")
	}
}

func TestListFriendRequests(t *testing.T) {
	t.Parallel()

	svc := NewSocialService()
	if _, err := svc.SendFriendRequest("p1", "p2"); err != nil {
		t.Fatalf("send request returned error: %+v", err)
	}
	if _, err := svc.SendFriendRequest("p3", "p2"); err != nil {
		t.Fatalf("send request returned error: %+v", err)
	}

	inbox, err := svc.ListFriendRequests("p2", "inbox", friendRequestPending)
	if err != nil {
		t.Fatalf("list friend requests returned error: %+v", err)
	}
	if len(inbox) != 2 {
		t.Fatalf("unexpected inbox size: %d", len(inbox))
	}
}

func TestSocialServiceWithInjectedStores(t *testing.T) {
	t.Parallel()

	requests := newMemoryFriendRequestStore()
	friendships := newMemoryFriendshipStore()
	blocks := newMemoryBlockStore()
	svc := NewSocialServiceWithStores(requests, friendships, blocks)

	request, err := svc.SendFriendRequest("p1", "p2")
	if err != nil {
		t.Fatalf("send request returned error: %+v", err)
	}

	if _, acceptErr := svc.AcceptFriendRequest(request.ID, "p2"); acceptErr != nil {
		t.Fatalf("accept returned error: %+v", acceptErr)
	}

	friends, listErr := friendships.ListFriends("p1")
	if listErr != nil {
		t.Fatalf("friendship store list failed: %v", listErr)
	}
	if len(friends) != 1 || friends[0] != "p2" {
		t.Fatalf("unexpected stored friends: %+v", friends)
	}
}
