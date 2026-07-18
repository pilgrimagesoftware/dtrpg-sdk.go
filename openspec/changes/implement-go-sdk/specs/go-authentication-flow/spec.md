## ADDED Requirements

### Requirement: Go authentication flow must preserve API contract meaning
The Go SDK MUST define how its authentication surface coordinates with the API contract
semantics owned by the API repository and the downstream session behavior owned by the Go
SDK.

#### Scenario: Authenticating through the Go SDK
- **WHEN** a caller calls `auth.Authenticate(ctx, applicationKey, config)`
- **THEN** the SDK sends `POST {baseURL}/{apiVersion}/auth_key?applicationKey=<key>` with
  an empty JSON body and `Content-Type: application/json`, and returns the deserialized
  `AuthTokenResponse` while preserving the API-defined token lifecycle semantics

#### Scenario: Depending on API-defined auth semantics
- **WHEN** the Go SDK interprets token issuance, expiry, refresh, or auth-failure behavior
- **THEN** it uses the meanings defined by the API repository instead of redefining them in
  Go-specific terms

### Requirement: Go authentication errors must preserve API meaning
The Go SDK MUST translate authentication failures into Go-facing `error` values without
obscuring the meaning of the underlying API failure.

#### Scenario: Authentication request fails
- **WHEN** the underlying HTTP call fails or the API returns a non-success status during
  `auth.Authenticate`
- **THEN** the Go SDK returns an error wrapping the underlying cause, retrievable via
  `errors.As` into the documented `library.APIError` or `library.DecodeError` types

### Requirement: SDK entry point wires configuration, authentication, and the library client
The Go SDK MUST expose an `Sdk` type that applies an `AuthTokenResponse` to produce an
active `AuthSession`, and constructs a `library.Client` only when both `Config` and an
active session are present.

#### Scenario: Applying an auth response
- **WHEN** a caller calls `sdk.ApplyAuthResponse(response)` on a configured `Sdk`
- **THEN** the SDK stores the resulting `AuthSession` and returns it

#### Scenario: Applying an auth response without configuration
- **WHEN** a caller calls `sdk.ApplyAuthResponse(response)` on an unconfigured `Sdk`
- **THEN** the SDK returns an error satisfying `errors.Is(err, dtrpg.ErrUnconfigured)` and
  does not store a session

#### Scenario: Building a library client without an active session
- **WHEN** a caller calls `sdk.LibraryClient()` before a session has been established
- **THEN** the SDK returns an error satisfying `errors.Is(err, dtrpg.ErrUnauthenticated)`
