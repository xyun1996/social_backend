package service

import "testing"

func TestLoginIssuesTokenPair(t *testing.T) {
	t.Parallel()

	svc := NewAuthService()

	pair, err := svc.Login("account-1", "player-1")
	if err != nil {
		t.Fatalf("login returned error: %+v", err)
	}

	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("expected tokens to be populated: %+v", pair)
	}

	if pair.TokenType != "Bearer" {
		t.Fatalf("unexpected token type: %q", pair.TokenType)
	}
}

func TestRefreshRotatesTokenPair(t *testing.T) {
	t.Parallel()

	svc := NewAuthService()
	pair, err := svc.Login("account-1", "player-1")
	if err != nil {
		t.Fatalf("login returned error: %+v", err)
	}

	refreshed, refreshErr := svc.Refresh(pair.RefreshToken)
	if refreshErr != nil {
		t.Fatalf("refresh returned error: %+v", refreshErr)
	}

	if refreshed.RefreshToken == pair.RefreshToken {
		t.Fatalf("expected refresh token rotation")
	}

	if _, secondErr := svc.Refresh(pair.RefreshToken); secondErr == nil {
		t.Fatalf("expected old refresh token to become invalid")
	}
}
