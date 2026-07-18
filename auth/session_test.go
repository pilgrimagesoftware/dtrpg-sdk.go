package auth

import "testing"

func TestAuthStateAPIString(t *testing.T) {
	cases := map[AuthState]string{
		Unauthenticated: "unauthenticated",
		TokenInvalid:    "token_invalid",
		TokenExpired:    "token_expired",
		RefreshExpired:  "refresh_expired",
		Unauthorized:    "unauthorized",
	}
	for state, want := range cases {
		if got := state.APIString(); got != want {
			t.Errorf("%v.APIString() = %q, want %q", state, got, want)
		}
	}
}

func TestAuthSessionFromAPIResponse(t *testing.T) {
	response := NewAuthTokenResponse("jwt", "refresh", 9_999_999_999)
	session := NewAuthSessionFromAPIResponse(response)

	if session.Token() != "jwt" {
		t.Errorf("Token() = %q, want %q", session.Token(), "jwt")
	}
	if session.RefreshToken() != "refresh" {
		t.Errorf("RefreshToken() = %q, want %q", session.RefreshToken(), "refresh")
	}
	if session.RefreshTokenTTL() != 9_999_999_999 {
		t.Errorf("RefreshTokenTTL() = %d, want %d", session.RefreshTokenTTL(), 9_999_999_999)
	}
}

func TestRefreshTokenExpiredAtBoundary(t *testing.T) {
	session := NewAuthSessionFromAPIResponse(NewAuthTokenResponse("t", "r", 1000))

	if session.RefreshTokenExpiredAt(999) {
		t.Error("RefreshTokenExpiredAt(999) = true, want false (before expiry)")
	}
	if !session.RefreshTokenExpiredAt(1000) {
		t.Error("RefreshTokenExpiredAt(1000) = false, want true (at expiry)")
	}
	if !session.RefreshTokenExpiredAt(1001) {
		t.Error("RefreshTokenExpiredAt(1001) = false, want true (past expiry)")
	}
}

func TestInvalidateProducesSessionTransition(t *testing.T) {
	session := NewAuthSessionFromAPIResponse(NewAuthTokenResponse("t", "r", 1))
	sessionErr := NewAuthSessionError("token_expired", "token expired", TokenExpired)

	transition := session.Invalidate(sessionErr)

	if transition.NextSession != nil {
		t.Errorf("NextSession = %+v, want nil", transition.NextSession)
	}
	if transition.Err != sessionErr {
		t.Errorf("Err = %+v, want %+v", transition.Err, sessionErr)
	}
}

func TestAuthSessionErrorMessage(t *testing.T) {
	err := NewAuthSessionError("token_expired", "token has expired", TokenExpired)
	want := "token has expired (token_expired) [token_expired]"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}
