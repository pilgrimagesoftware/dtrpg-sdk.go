## ADDED Requirements

### Requirement: The Go SDK MUST define Go-facing session lifecycle behavior
The Go SDK MUST define how it holds, invalidates, and reacts to authentication session
state using the token lifecycle semantics owned by the API repository.

#### Scenario: Holding authenticated session state in the Go SDK
- **WHEN** the Go SDK has successfully authenticated against the API via
  `sdk.ApplyAuthResponse`
- **THEN** it holds that state in an `auth.AuthSession` exposing `Token()`,
  `RefreshToken()`, and `RefreshTokenTTL()` accessors

#### Scenario: Determining refresh-token expiry
- **WHEN** a caller calls `session.RefreshTokenExpiredAt(unixTimestamp)`
- **THEN** the SDK returns whether the session's refresh token has expired as of that
  Unix timestamp, using the API-defined `refreshTokenTTL` value

### Requirement: Go session lifecycle behavior MUST depend on API contract meaning
The Go SDK MUST treat token expiry, refresh semantics, and auth-failure meaning as
upstream API contract dependencies rather than redefining them locally.

#### Scenario: Reacting to expired or invalid session state
- **WHEN** a caller calls `sdk.InvalidateSession(err)` with an `AuthSessionError` whose
  `AuthState` is one of `Unauthenticated`, `TokenInvalid`, `TokenExpired`,
  `RefreshExpired`, or `Unauthorized`
- **THEN** the Go SDK produces the documented `SessionTransition` (a possibly-nil next
  session plus the triggering error) while preserving the `AuthState` meaning defined by
  the API contract

#### Scenario: Invalidating a session that does not exist
- **WHEN** a caller calls `sdk.InvalidateSession(err)` on an `Sdk` with no active session
- **THEN** the SDK returns an error satisfying `errors.Is(err, dtrpg.ErrUnauthenticated)`

#### Scenario: Clearing a session is a no-op when none is active
- **WHEN** a caller calls `sdk.ClearSession()` on an `Sdk` with no active session
- **THEN** the call succeeds without error and the SDK remains unauthenticated
