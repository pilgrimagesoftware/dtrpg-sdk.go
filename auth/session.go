package auth

import (
	"encoding/json"
	"fmt"
)

// AuthTokenResponse is the raw authentication token payload returned by the DriveThruRPG
// API. This struct is a direct representation of the API response fields. Callers should
// convert it into an AuthSession via NewAuthSessionFromAPIResponse before treating the
// session as active.
type AuthTokenResponse struct {
	// Token is the short-lived JWT access token used to authenticate API requests.
	Token string `json:"token"`
	// RefreshToken is the long-lived refresh token used to obtain a new access token.
	RefreshToken string `json:"refreshToken"`
	// RefreshTokenTTL is the Unix timestamp (seconds) at which the refresh token expires.
	RefreshTokenTTL uint64 `json:"refreshTokenTTL"`
}

// UnmarshalJSON requires token, refreshToken, and refreshTokenTTL to be present, matching
// the Rust SDK's serde-derived Deserialize (which has no #[serde(default)] on these
// fields). Go's encoding/json otherwise leaves missing fields at their zero value instead
// of erroring, which would let a non-token error body (e.g. {"message": "..."}) silently
// decode into an empty, misleadingly "successful" AuthTokenResponse.
func (r *AuthTokenResponse) UnmarshalJSON(data []byte) error {
	var raw struct {
		Token           *string `json:"token"`
		RefreshToken    *string `json:"refreshToken"`
		RefreshTokenTTL *uint64 `json:"refreshTokenTTL"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	switch {
	case raw.Token == nil:
		return fmt.Errorf("missing field `token`")
	case raw.RefreshToken == nil:
		return fmt.Errorf("missing field `refreshToken`")
	case raw.RefreshTokenTTL == nil:
		return fmt.Errorf("missing field `refreshTokenTTL`")
	}
	r.Token = *raw.Token
	r.RefreshToken = *raw.RefreshToken
	r.RefreshTokenTTL = *raw.RefreshTokenTTL
	return nil
}

// NewAuthTokenResponse constructs a new AuthTokenResponse from the given token fields.
// refreshTokenTTL is a Unix timestamp (seconds since the epoch) representing the absolute
// expiry of the refresh token.
func NewAuthTokenResponse(token, refreshToken string, refreshTokenTTL uint64) AuthTokenResponse {
	return AuthTokenResponse{Token: token, RefreshToken: refreshToken, RefreshTokenTTL: refreshTokenTTL}
}

// AuthSession is an active authentication session with the DriveThruRPG API.
//
// AuthSession is the validated, runtime representation of an authenticated user. It is
// obtained by calling the root package's Sdk.ApplyAuthResponse with a token response from
// the API.
type AuthSession struct {
	token           string
	refreshToken    string
	refreshTokenTTL uint64
}

// NewAuthSessionFromAPIResponse creates an AuthSession from a raw AuthTokenResponse.
func NewAuthSessionFromAPIResponse(response AuthTokenResponse) AuthSession {
	return AuthSession{
		token:           response.Token,
		refreshToken:    response.RefreshToken,
		refreshTokenTTL: response.RefreshTokenTTL,
	}
}

// Token returns the short-lived JWT access token for this session.
func (s AuthSession) Token() string {
	return s.token
}

// RefreshToken returns the long-lived refresh token for this session.
func (s AuthSession) RefreshToken() string {
	return s.refreshToken
}

// RefreshTokenTTL returns the Unix timestamp (seconds) at which the refresh token expires.
func (s AuthSession) RefreshTokenTTL() uint64 {
	return s.refreshTokenTTL
}

// RefreshTokenExpiredAt returns true if unixTimestamp is at or past the refresh token's
// expiry. Use the current wall-clock time (as a Unix timestamp) to determine whether the
// session's refresh token is still usable.
func (s AuthSession) RefreshTokenExpiredAt(unixTimestamp uint64) bool {
	return unixTimestamp >= s.refreshTokenTTL
}

// Invalidate produces a SessionTransition representing this session's invalidation.
//
// The resulting transition has no replacement session (NextSession is nil) and carries the
// provided err describing why the session was invalidated.
//
// This is a lower-level primitive for code working with an AuthSession directly. The root
// package's Sdk.InvalidateSession is the higher-level equivalent for callers going through
// the SDK: it does not build a SessionTransition, and unconditionally clears the stored
// session rather than leaving room for a NextSession replacement.
func (s AuthSession) Invalidate(err AuthSessionError) SessionTransition {
	return SessionTransition{NextSession: nil, Err: err}
}

// SessionTransition is the outcome of invalidating an AuthSession.
//
// A SessionTransition is returned when a session ends due to an error. It records the
// error that caused the invalidation and an optional replacement session (for cases where
// token refresh succeeds mid-invalidation).
type SessionTransition struct {
	// NextSession is the replacement session, if one was established as part of this
	// transition. It is nil when the session was simply terminated without a refresh.
	NextSession *AuthSession
	// Err is the error that caused this session transition.
	Err AuthSessionError
}
