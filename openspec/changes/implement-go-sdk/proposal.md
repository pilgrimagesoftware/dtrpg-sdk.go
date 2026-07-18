## Why

`dtrpg-sdk/go` currently contains only a license and a stub README — there is no Go SDK for
the DriveThruRPG API, while the Rust SDK (`dtrpg-sdk/rust`) already ships a working
configuration, auth-session, and library-client surface. Go developers integrating with
DriveThruRPG have no first-party option and must hand-roll HTTP calls against the API.

## What Changes

- Add a `Config` type: application key, base URL, API version, with DriveThruRPG's default
  base URL/version baked in.
- Add an auth-session lifecycle: `AuthSession`, `AuthState`, `AuthTokenResponse`, session
  invalidation/transition, mirroring the Rust SDK's session semantics.
- Add key-exchange authentication: exchange an application key for a JWT via
  `POST auth_key?applicationKey=<key>`.
- Add the credential-login-exchange flow: exchange an email/password login for an
  application key via DriveThruRPG's web login endpoints, upstream of key exchange.
- Add a `LibraryClient` covering order products (list/get/prepare-download), product lists
  (list/create/delete), and product list items (list/create/delete).
- Add library model types for the JSON:API-shaped library resources (order products,
  product lists, product list items, pagination, publisher/product sideloads).
- Add build-time OpenAPI metadata generation from the `dtrpg-api` submodule's
  `openapi.yaml`, mirroring the Rust SDK's `go generate`-driven approach (server URL +
  operation list), for contract-freshness checking rather than full client codegen.
- Add passive Retry-After / rate-limit exposure on API errors: parse the `Retry-After`
  header (delay-seconds only) into an error field callers can act on; the SDK does not
  retry automatically.
- Add a top-level `DriveThruRpgSdk` (or idiomatic Go equivalent) entry point wiring
  `Config` → `AuthSession` → `LibraryClient`, matching the Rust SDK's `configure` →
  `ApplyAuthResponse` → `LibraryClient()` lifecycle.

Out of scope for this change: CI/release/publishing automation for the Go module (a
follow-up change once the SDK surface is stable), and any UI/app integration work.

## Capabilities

### New Capabilities
- `go-sdk-configuration`: `Config` type and defaults; requiring explicit configuration
  before authenticated use; OpenAPI-derived metadata generated at build/generate time.
- `go-authentication-flow`: key-exchange authentication (application key → JWT) and the
  SDK-level auth surface (`ApplyAuthResponse`, session access), preserving upstream API
  token-lifecycle semantics.
- `go-session-lifecycle`: `AuthSession`/`AuthState` holding and invalidation, session
  transitions on auth failure, without redefining what expiry/refresh mean upstream.
- `credential-login-exchange`: email/password → application key exchange via
  DriveThruRPG's web login endpoints, kept as a distinct flow from key exchange.
- `go-library-client`: authenticated client for order products, product lists, and
  product list items, built only from a configured + authenticated SDK.
- `go-library-types`: Go types for the library API's JSON:API-shaped resources, tracking
  the upstream API contract's fields and optionality.
- `rate-limit-retry-after`: passive parsing/exposure of the `Retry-After` header on API
  errors; no automatic retry.

### Modified Capabilities
- (none — `dtrpg-sdk/go` has no existing specs)

## Impact

- New Go module under `dtrpg-sdk/go` (module path, `go.mod`, package layout to be defined
  in design.md).
- Adds `dtrpg-api` as a git submodule of `dtrpg-sdk/go` (mirroring `rust/API`), so
  `openapi.yaml` is available for build-time/generate-time metadata extraction.
- No changes to `dtrpg-api`, `dtrpg-app`, or the other language SDKs.
