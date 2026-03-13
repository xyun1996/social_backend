package main

import (
	"reflect"
	"testing"
)

func TestValidateInventory(t *testing.T) {
	t.Parallel()

	inv := inventory{
		HTTPFiles:  sortedCopy(controlPlaneSurfaces),
		ProtoFiles: sortedCopy(controlPlaneSurfaces),
		TCPFiles:   sortedCopy(realtimeSurfaces),
		HTTPIndex:  "- [chat](chat.md)\n- [gateway](gateway.md)\n- [guild](guild.md)\n- [identity](identity.md)\n- [invite](invite.md)\n- [ops](ops.md)\n- [party](party.md)\n- [presence](presence.md)\n- [social](social.md)\n- [worker](worker.md)\n",
		ProtoIndex: "- `chat`\n- `gateway`\n- `guild`\n- `identity`\n- `invite`\n- `ops`\n- `party`\n- `presence`\n- `social`\n- `worker`\n",
		TCPIndex:   "- [chat](chat.md)\n- [gateway](gateway.md)\n",
	}

	problems := validateInventory(inv)
	if len(problems) != 0 {
		t.Fatalf("unexpected validation problems: %+v", problems)
	}
}

func TestValidateInventoryFindsDrift(t *testing.T) {
	t.Parallel()

	problems := validateInventory(inventory{
		HTTPFiles:  []string{"chat"},
		ProtoFiles: sortedCopy(controlPlaneSurfaces),
		TCPFiles:   []string{"chat"},
		HTTPIndex:  "- [chat](chat.md)\n",
		ProtoIndex: "- `chat`\n",
		TCPIndex:   "- [chat](chat.md)\n",
	})

	if len(problems) == 0 {
		t.Fatal("expected validation problems")
	}
}

func TestSortedCopy(t *testing.T) {
	t.Parallel()

	input := []string{"worker", "chat", "identity"}
	got := sortedCopy(input)
	want := []string{"chat", "identity", "worker"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected sorted copy: got=%v want=%v", got, want)
	}
	if reflect.DeepEqual(input, got) && &input[0] == &got[0] {
		t.Fatal("sortedCopy should return a distinct slice")
	}
}
