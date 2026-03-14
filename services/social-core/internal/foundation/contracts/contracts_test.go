package contracts

import "testing"

func TestPrincipalCarriesIdentityScope(t *testing.T) {
	principal := Principal{
		AccountID: "account-1",
		PlayerID:  "player-1",
		Roles:     []string{"member"},
	}

	if principal.AccountID == "" || principal.PlayerID == "" {
		t.Fatalf("expected principal identity scope to be populated")
	}
}
