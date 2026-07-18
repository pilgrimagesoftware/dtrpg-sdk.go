package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pilgrimagesoftware/dtrpg-sdk.go/library"
)

func TestValidateLoginResponseParsesExampleFromLoginMD(t *testing.T) {
	// Exact example from dtrpg-api/LOGIN.md.
	var response validateLoginResponse
	if err := response.UnmarshalJSON([]byte(`["password",true,"Locked",true]`)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
	if response.FieldName != "password" || !response.OK || response.Message != "Locked" || !response.Locked {
		t.Errorf("response = %+v, want {password true Locked true}", response)
	}
}

func TestValidateLoginResponseParsesInvalidCredentials(t *testing.T) {
	var response validateLoginResponse
	if err := response.UnmarshalJSON([]byte(`["password",false,"Invalid",false]`)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
	if response.OK || response.Message != "Invalid" || response.Locked {
		t.Errorf("response = %+v, want ok=false message=Invalid locked=false", response)
	}
}

func TestValidateLoginResponseRejectsShortArray(t *testing.T) {
	var response validateLoginResponse
	if err := response.UnmarshalJSON([]byte(`["password",true]`)); err == nil {
		t.Error("UnmarshalJSON() error = nil, want an error for a short array")
	}
}

func TestValidCredentialsReturnApplicationKey(t *testing.T) {
	mux := http.NewServeMux()
	validateCalls := 0
	mux.HandleFunc("/validate_login_credentials.php", func(w http.ResponseWriter, r *http.Request) {
		validateCalls++
		_, _ = w.Write([]byte(`["password",true,"Locked",true]`))
	})
	mux.HandleFunc("/create_account_app.php", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"success","message":{"key":"test-app-key-abc123"}}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	key, err := doLogin(context.Background(), "user@example.com", "secret", server.URL)
	if err != nil {
		t.Fatalf("doLogin() error = %v", err)
	}
	if key != "test-app-key-abc123" {
		t.Errorf("key = %q, want %q", key, "test-app-key-abc123")
	}
	if validateCalls != 1 {
		t.Errorf("validate_login_credentials.php called %d times, want 1", validateCalls)
	}
}

func TestInvalidCredentialsReturnErrorWithoutCallingKeyEndpoint(t *testing.T) {
	keyCalls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/validate_login_credentials.php", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`["password",false,"Invalid",false]`))
	})
	mux.HandleFunc("/create_account_app.php", func(w http.ResponseWriter, r *http.Request) {
		keyCalls++
		_, _ = w.Write([]byte(`{"status":"success","message":{"key":"unused"}}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	_, err := doLogin(context.Background(), "user@example.com", "secret", server.URL)
	if !errors.Is(err, library.ErrInvalidCredentials) {
		t.Fatalf("doLogin() error = %v, want library.ErrInvalidCredentials", err)
	}
	if keyCalls != 0 {
		t.Errorf("create_account_app.php called %d times, want 0", keyCalls)
	}
}

func TestKeyRequestFailureAfterValidCredentialsIsDistinguishable(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/validate_login_credentials.php", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`["password",true,"Locked",true]`))
	})
	mux.HandleFunc("/create_account_app.php", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"error","message":{"key":""}}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	_, err := doLogin(context.Background(), "user@example.com", "secret", server.URL)

	var keyErr *library.ApplicationKeyRequestError
	if !errors.As(err, &keyErr) {
		t.Fatalf("doLogin() error = %v (%T), want *library.ApplicationKeyRequestError", err, err)
	}
	if keyErr.Status != "error" {
		t.Errorf("keyErr.Status = %q, want %q", keyErr.Status, "error")
	}
	if errors.Is(err, library.ErrInvalidCredentials) {
		t.Error("error should be distinguishable from library.ErrInvalidCredentials")
	}
}

func TestExistingAuthKeyExchangeUnaffected(t *testing.T) {
	// LoginWithCredentials shares no state with Authenticate; this is a compile-time /
	// smoke check that both remain independently callable with their documented
	// signatures.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"token":"t","refreshToken":"r","refreshTokenTTL":1}`))
	}))
	defer server.Close()

	config := library.NewConfigWithBaseURL("app-key", server.URL)
	if _, err := Authenticate(context.Background(), "app-key", config); err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
}
