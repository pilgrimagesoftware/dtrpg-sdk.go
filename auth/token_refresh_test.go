package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pilgrimagesoftware/dtrpg-sdk.go/library"
)

func TestRefreshTokenSendsCookieAndDecodesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/vBeta/auth/refresh" {
			t.Errorf("path = %s, want /vBeta/auth/refresh", r.URL.Path)
		}
		cookie, err := r.Cookie("refreshToken")
		if err != nil || cookie.Value != "old-refresh" {
			t.Errorf("refreshToken cookie = %v (err %v), want old-refresh", cookie, err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token":           "new-jwt",
			"refreshToken":    "new-refresh",
			"refreshTokenTTL": 2_000_000_000,
		})
	}))
	defer server.Close()

	config := library.NewConfigWithBaseURL("my-app-key", server.URL)
	response, err := RefreshToken(context.Background(), "old-refresh", config)
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}
	if response.Token != "new-jwt" {
		t.Errorf("Token = %q, want %q", response.Token, "new-jwt")
	}
	if response.RefreshToken != "new-refresh" {
		t.Errorf("RefreshToken = %q, want %q", response.RefreshToken, "new-refresh")
	}
	if response.RefreshTokenTTL != 2_000_000_000 {
		t.Errorf("RefreshTokenTTL = %d, want 2000000000", response.RefreshTokenTTL)
	}
}

func TestRefreshTokenReportsRejectionAsDecodeErrorWithStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errorCode": "refresh_expired"}`))
	}))
	defer server.Close()

	config := library.NewConfigWithBaseURL("my-app-key", server.URL)
	_, err := RefreshToken(context.Background(), "expired", config)

	var decodeErr *library.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Fatalf("RefreshToken() error = %v (%T), want *library.DecodeError", err, err)
	}
	if decodeErr.Status != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", decodeErr.Status, http.StatusUnauthorized)
	}
}
