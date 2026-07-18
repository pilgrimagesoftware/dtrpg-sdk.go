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

func TestAuthenticateSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/vBeta/auth_key" {
			t.Errorf("path = %s, want /vBeta/auth_key", r.URL.Path)
		}
		if got := r.URL.Query().Get("applicationKey"); got != "my-app-key" {
			t.Errorf("applicationKey = %q, want %q", got, "my-app-key")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token":           "jwt-token",
			"refreshToken":    "refresh-token",
			"refreshTokenTTL": 9_999_999_999,
		})
	}))
	defer server.Close()

	config := library.NewConfigWithBaseURL("my-app-key", server.URL)
	response, err := Authenticate(context.Background(), "my-app-key", config)
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if response.Token != "jwt-token" {
		t.Errorf("Token = %q, want %q", response.Token, "jwt-token")
	}
	if response.RefreshToken != "refresh-token" {
		t.Errorf("RefreshToken = %q, want %q", response.RefreshToken, "refresh-token")
	}
}

func TestAuthenticateNonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message": "invalid application key"}`))
	}))
	defer server.Close()

	config := library.NewConfigWithBaseURL("bad-key", server.URL)
	_, err := Authenticate(context.Background(), "bad-key", config)
	if err == nil {
		t.Fatal("Authenticate() error = nil, want a decode error since the body isn't a token response")
	}
}

func TestAuthenticateMalformedBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()

	config := library.NewConfigWithBaseURL("my-app-key", server.URL)
	_, err := Authenticate(context.Background(), "my-app-key", config)

	var decodeErr *library.DecodeError
	if !errors.As(err, &decodeErr) {
		t.Fatalf("Authenticate() error = %v (%T), want *library.DecodeError", err, err)
	}
}
