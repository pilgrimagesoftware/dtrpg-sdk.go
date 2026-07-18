# credential-login-exchange Specification

## Purpose
Define the website credential exchange that turns a DriveThruRPG email/password pair into
an application key, kept distinct from the API key-exchange flow it feeds into.

## Requirements

### Requirement: SDK exchanges credentials for an application key
The Go SDK SHALL expose `auth.LoginWithCredentials(ctx, email, password, config)` that
exchanges an email/password pair for a DriveThruRPG application key by calling
`validate_login_credentials.php` then `create_account_app.php` on
`www.drivethrurpg.com`, without requiring the caller to construct either request directly.

#### Scenario: Valid credentials return an application key
- **WHEN** `LoginWithCredentials` is called with an email and password that DriveThruRPG
  accepts
- **THEN** it calls `validate_login_credentials.php`, confirms the response indicates
  valid credentials, calls `create_account_app.php`, and returns the application key from
  `message.key`

#### Scenario: Invalid credentials stop before the key request
- **WHEN** `validate_login_credentials.php` indicates invalid credentials
- **THEN** `LoginWithCredentials` returns an error satisfying
  `errors.Is(err, library.ErrInvalidCredentials)` without calling `create_account_app.php`

#### Scenario: Key request fails after valid credentials
- **WHEN** `validate_login_credentials.php` indicates valid credentials but
  `create_account_app.php` returns an error or unexpected body
- **THEN** `LoginWithCredentials` returns a `library.ApplicationKeyRequestError`,
  distinguishable via `errors.As` from `library.ErrInvalidCredentials`

### Requirement: Positional validation response is typed
The SDK SHALL deserialize `validate_login_credentials.php`'s bare JSON array response into
a named struct with documented field order, rather than requiring callers to index into a
raw `[]any`.

#### Scenario: Parsing the documented example response
- **WHEN** the SDK receives the exact example body from `LOGIN.md`
  (`["password",true,"Locked",true]`)
- **THEN** it parses successfully into the typed struct with each field accessible by name
  via a custom `UnmarshalJSON` implementation

### Requirement: Credential exchange is scoped separately from the API auth client
The SDK SHALL implement the website credential exchange in a file/package distinct from
key exchange (`auth/credential_login.go` vs. `auth/key_exchange.go`), and SHALL document
in that file's package comments that it targets `www.drivethrurpg.com` rather than
`api.drivethrurpg.com` and does not replace or modify `auth.Authenticate`.

#### Scenario: Existing auth_key exchange is unaffected
- **WHEN** a caller uses `auth.Authenticate` with an application key
- **THEN** its behavior and signature are unchanged by the addition of
  `LoginWithCredentials`
