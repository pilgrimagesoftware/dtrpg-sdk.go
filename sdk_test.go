package dtrpg

import (
	"errors"
	"testing"

	"github.com/pilgrimagesoftware/dtrpg-sdk.go/auth"
)

func TestSdkLifecycle(t *testing.T) {
	sdk := NewSdk()

	if _, err := sdk.RequireConfig(); !errors.Is(err, ErrUnconfigured) {
		t.Fatalf("RequireConfig() before Configure: err = %v, want ErrUnconfigured", err)
	}

	sdk.Configure(NewConfig("app-key"))
	if sdk.Config() == nil {
		t.Fatal("Config() is nil after Configure")
	}

	if _, err := sdk.ApplyAuthResponse(NewAuthTokenResponse("t", "r", 9_999_999_999)); err != nil {
		t.Fatalf("ApplyAuthResponse() error = %v", err)
	}
	session, err := sdk.RequireSession()
	if err != nil {
		t.Fatalf("RequireSession() error = %v", err)
	}
	if session.Token() != "t" {
		t.Errorf("session.Token() = %q, want %q", session.Token(), "t")
	}

	client, err := sdk.LibraryClient()
	if err != nil {
		t.Fatalf("LibraryClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("LibraryClient() returned nil client with nil error")
	}

	invalidatedErr := auth.NewAuthSessionError("token_expired", "token expired", auth.TokenExpired)
	returnedErr, err := sdk.InvalidateSession(invalidatedErr)
	if err != nil {
		t.Fatalf("InvalidateSession() error = %v", err)
	}
	if returnedErr != invalidatedErr {
		t.Errorf("InvalidateSession() returned %+v, want %+v", returnedErr, invalidatedErr)
	}
	if sdk.Session() != nil {
		t.Error("Session() is non-nil after InvalidateSession")
	}

	if _, err := sdk.LibraryClient(); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("LibraryClient() after invalidate: err = %v, want ErrUnauthenticated", err)
	}
}

func TestApplyAuthResponseRequiresConfig(t *testing.T) {
	sdk := NewSdk()
	_, err := sdk.ApplyAuthResponse(NewAuthTokenResponse("t", "r", 1))
	if !errors.Is(err, ErrUnconfigured) {
		t.Fatalf("ApplyAuthResponse() without config: err = %v, want ErrUnconfigured", err)
	}
	if sdk.Session() != nil {
		t.Error("Session() is non-nil after a failed ApplyAuthResponse")
	}
}

func TestInvalidateSessionRequiresSession(t *testing.T) {
	sdk := NewSdkWithConfig(NewConfig("app-key"))
	_, err := sdk.InvalidateSession(auth.AuthSessionError{})
	if !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("InvalidateSession() without session: err = %v, want ErrUnauthenticated", err)
	}
}

func TestClearSessionIsNoOpWithoutSession(t *testing.T) {
	sdk := NewSdk()
	sdk.ClearSession()
	if sdk.Session() != nil {
		t.Error("Session() is non-nil after ClearSession on a fresh SDK")
	}
}

func TestLibraryClientRequiresConfigAndSession(t *testing.T) {
	sdk := NewSdk()
	if _, err := sdk.LibraryClient(); !errors.Is(err, ErrUnconfigured) {
		t.Fatalf("LibraryClient() unconfigured: err = %v, want ErrUnconfigured", err)
	}

	sdk.Configure(NewConfig("app-key"))
	if _, err := sdk.LibraryClient(); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("LibraryClient() unauthenticated: err = %v, want ErrUnauthenticated", err)
	}
}
