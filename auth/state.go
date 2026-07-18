// Package auth provides the DriveThruRPG SDK's authentication session types and flows:
// AuthSession/AuthState/AuthTokenResponse/SessionTransition are the session and state
// types shared across the SDK. Authenticate exchanges an application key for a session
// token against api.drivethrurpg.com. LoginWithCredentials exchanges an email/password
// pair for an application key against www.drivethrurpg.com, upstream of Authenticate.
package auth

import "fmt"

// AuthState is the authentication failure state reported by the DriveThruRPG API. Each
// value maps to a specific string value in the API protocol (see AuthState.APIString).
// Use this to understand why a session was invalidated and decide how to recover.
type AuthState int

const (
	// Unauthenticated indicates no authentication credentials are present.
	Unauthenticated AuthState = iota
	// TokenInvalid indicates the provided access token is structurally invalid or
	// unrecognized.
	TokenInvalid
	// TokenExpired indicates the access token has passed its expiry time and must be
	// refreshed.
	TokenExpired
	// RefreshExpired indicates the refresh token has passed its expiry time; the user
	// must re-authenticate.
	RefreshExpired
	// Unauthorized indicates the credentials are valid but the caller lacks permission
	// for the requested resource.
	Unauthorized
)

// APIString returns the API wire string for this state, suitable for comparison with API
// error codes.
func (s AuthState) APIString() string {
	switch s {
	case Unauthenticated:
		return "unauthenticated"
	case TokenInvalid:
		return "token_invalid"
	case TokenExpired:
		return "token_expired"
	case RefreshExpired:
		return "refresh_expired"
	case Unauthorized:
		return "unauthorized"
	default:
		return fmt.Sprintf("auth_state(%d)", int(s))
	}
}

// String implements fmt.Stringer, returning the same value as APIString.
func (s AuthState) String() string {
	return s.APIString()
}

// AuthSessionError is a structured authentication error returned by the DriveThruRPG API.
//
// It carries the machine-readable ErrorCode, a human-readable Message, and an AuthState
// that classifies the failure. It is used when the API explicitly rejects an operation due
// to an auth-related condition.
type AuthSessionError struct {
	// ErrorCode is the machine-readable error code from the API (e.g. "token_expired").
	ErrorCode string
	// Message is a human-readable description of the error.
	Message string
	// State is the AuthState classification for this error.
	State AuthState
}

// NewAuthSessionError constructs a new AuthSessionError from its components.
func NewAuthSessionError(errorCode, message string, state AuthState) AuthSessionError {
	return AuthSessionError{ErrorCode: errorCode, Message: message, State: state}
}

func (e AuthSessionError) Error() string {
	return fmt.Sprintf("%s (%s) [%s]", e.Message, e.ErrorCode, e.State)
}
