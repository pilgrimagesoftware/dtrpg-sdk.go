package dtrpg

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func sdkWithSession(t *testing.T, baseURL string) *Sdk {
	t.Helper()
	sdk := NewSdkWithConfig(NewConfigWithBaseURL("app-key", baseURL))
	if _, err := sdk.ApplyAuthResponse(NewAuthTokenResponse("old-jwt", "old-refresh", 1_000_000_000)); err != nil {
		t.Fatalf("ApplyAuthResponse() error = %v", err)
	}
	return sdk
}

func TestRefreshSessionReplacesTheStoredSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("refreshToken"); err != nil || cookie.Value != "old-refresh" {
			t.Errorf("refreshToken cookie = %v (err %v), want old-refresh", cookie, err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token":"new-jwt","refreshToken":"new-refresh","refreshTokenTTL":2000000000}`))
	}))
	defer server.Close()
	sdk := sdkWithSession(t, server.URL)

	session, err := sdk.RefreshSession(context.Background())
	if err != nil {
		t.Fatalf("RefreshSession() error = %v", err)
	}

	if session.Token() != "new-jwt" || session.RefreshToken() != "new-refresh" || session.RefreshTokenTTL() != 2_000_000_000 {
		t.Errorf("session = (%q, %q, %d), want (new-jwt, new-refresh, 2000000000)",
			session.Token(), session.RefreshToken(), session.RefreshTokenTTL())
	}
	if sdk.Session().Token() != "new-jwt" {
		t.Errorf("stored token = %q, want new-jwt", sdk.Session().Token())
	}
}

func TestRefreshSessionKeepsTheOldSessionWhenRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errorCode":"refresh_expired"}`))
	}))
	defer server.Close()
	sdk := sdkWithSession(t, server.URL)

	if _, err := sdk.RefreshSession(context.Background()); err == nil {
		t.Fatal("RefreshSession() error = nil, want a rejection error")
	}

	if sdk.Session().Token() != "old-jwt" {
		t.Errorf("stored token = %q, want old-jwt", sdk.Session().Token())
	}
}

func TestRefreshSessionWithoutConfigOrSession(t *testing.T) {
	if _, err := NewSdk().RefreshSession(context.Background()); !errors.Is(err, ErrUnconfigured) {
		t.Errorf("unconfigured: err = %v, want ErrUnconfigured", err)
	}
	sdk := NewSdkWithConfig(NewConfig("app-key"))
	if _, err := sdk.RefreshSession(context.Background()); !errors.Is(err, ErrUnauthenticated) {
		t.Errorf("no session: err = %v, want ErrUnauthenticated", err)
	}
}
