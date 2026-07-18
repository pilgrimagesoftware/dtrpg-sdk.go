package dtrpg

import (
	"github.com/pilgrimagesoftware/dtrpg-sdk.go/auth"
	"github.com/pilgrimagesoftware/dtrpg-sdk.go/library"
)

// Sdk is the DriveThruRPG SDK client.
//
// Sdk coordinates SDK-level configuration and authentication session lifecycle. It must be
// configured before any authenticated API calls can succeed.
//
// # Lifecycle
//
// 1. Create an Sdk instance, optionally supplying a Config upfront.
// 2. After a successful API login, call ApplyAuthResponse to store the session.
// 3. Use RequireSession to obtain the active session before making requests.
// 4. Call InvalidateSession when the API reports a session error, or ClearSession to log
// out.
type Sdk struct {
	config  *Config
	session *auth.AuthSession
}

// NewSdk creates an unconfigured Sdk instance. Call Configure before making any API calls,
// or prefer NewSdkWithConfig if the configuration is available at construction time.
func NewSdk() *Sdk {
	return &Sdk{}
}

// NewSdkWithConfig creates an Sdk instance pre-loaded with the given Config.
func NewSdkWithConfig(config Config) *Sdk {
	return &Sdk{config: &config}
}

// Configure sets or replaces the SDK's configuration.
//
// This can be called at any time, including after the SDK has been used. Replacing the
// configuration does not automatically clear an existing session.
func (s *Sdk) Configure(config Config) {
	s.config = &config
}

// Config returns the current configuration, or nil if the SDK is unconfigured.
func (s *Sdk) Config() *Config {
	return s.config
}

// Session returns the current authentication session, or nil if unauthenticated.
func (s *Sdk) Session() *auth.AuthSession {
	return s.session
}

// RequireConfig returns the current configuration, or ErrUnconfigured if absent.
//
// Use this in call chains where a missing config should surface as an error.
func (s *Sdk) RequireConfig() (*Config, error) {
	if s.config == nil {
		return nil, ErrUnconfigured
	}
	return s.config, nil
}

// RequireSession returns the current session, or ErrUnauthenticated if absent.
//
// Use this in call chains where a missing session should surface as an error.
func (s *Sdk) RequireSession() (*auth.AuthSession, error) {
	if s.session == nil {
		return nil, ErrUnauthenticated
	}
	return s.session, nil
}

// ApplyAuthResponse stores a new authentication session derived from a raw API token
// response.
//
// Returns ErrUnconfigured if the SDK has not been configured yet. On success, returns the
// newly stored session.
func (s *Sdk) ApplyAuthResponse(response AuthTokenResponse) (*auth.AuthSession, error) {
	if _, err := s.RequireConfig(); err != nil {
		return nil, err
	}
	session := auth.NewAuthSessionFromAPIResponse(response)
	s.session = &session
	return s.RequireSession()
}

// ClearSession removes the current authentication session without recording an error.
//
// Use this for voluntary log-out flows. For API-reported session failures, prefer
// InvalidateSession.
func (s *Sdk) ClearSession() {
	s.session = nil
}

// InvalidateSession removes the current session and records the API-reported error that
// caused it.
//
// Returns ErrUnauthenticated if there is no session to invalidate. On success, returns the
// provided err so callers can inspect or propagate it.
func (s *Sdk) InvalidateSession(err AuthSessionError) (AuthSessionError, error) {
	if _, sessErr := s.RequireSession(); sessErr != nil {
		return AuthSessionError{}, sessErr
	}
	s.ClearSession()
	return err, nil
}

// LibraryClient creates a library.Client from the current configuration and session.
//
// Returns ErrUnconfigured if the SDK has not been configured, or ErrUnauthenticated if
// there is no active session.
func (s *Sdk) LibraryClient() (*library.Client, error) {
	config, err := s.RequireConfig()
	if err != nil {
		return nil, err
	}
	session, err := s.RequireSession()
	if err != nil {
		return nil, err
	}
	return library.NewClient(*config, session.Token()), nil
}
