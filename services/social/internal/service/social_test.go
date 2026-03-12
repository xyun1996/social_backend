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
