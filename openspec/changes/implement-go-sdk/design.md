## Context

`dtrpg-sdk/go` is an empty module (LICENSE + README only). `dtrpg-sdk/rust` is the
feature-complete reference: `Config`, an auth-session lifecycle (`AuthSession`/`AuthState`),
key-exchange and credential-login auth flows, a `LibraryClient` for order products/product
lists/product list items, library model types, build-time OpenAPI metadata generation from
the `dtrpg-api` submodule, and passive `Retry-After` exposure on API errors. This design
maps each of those onto idiomatic Go rather than transliterating Rust syntax — Go has no
enums, no `Result<T, E>`, no build-script `OUT_DIR`, and untyped `context.Context`
propagation is the ecosystem norm where Rust's `async fn` needed none.

## Goals / Non-Goals

**Goals:**
- Near feature-parity with the Rust SDK's public surface: configuration, auth-session
  lifecycle, key-exchange auth, credential-login exchange, library client + models,
  OpenAPI metadata generation, Retry-After exposure.
- Idiomatic Go: `context.Context` on every I/O call, `error`-based error handling, stdlib
  `net/http` + `encoding/json`, table-driven tests against `httptest.Server`.
- Match the Rust SDK's actual runtime behavior (not just its specs) so the two SDKs agree
  in practice — see Decisions on the applicationKey/query-param discrepancy.

**Non-Goals:**
- CI/release/publishing automation for the Go module (separate follow-up change).
- Full OpenAPI client codegen (request/response types generated from the spec) — this
  change only generates build-time metadata (server URL, operation list) for
  contract-freshness checking, matching the Rust SDK's scope.
- Backward compatibility with any prior Go SDK version (none exists).

## Decisions

**Module path and package layout.** Module `github.com/pilgrimagesoftware/dtrpg-sdk.go`,
Go 1.22+. Root package `dtrpg` holds `Config`, `Sdk`, and the SDK-level error types.
Two subpackages mirror the Rust module split: `auth` (session, key exchange, credential
login) and `library` (client + models). The root package re-exports the types callers need
day-to-day via type aliases (`type AuthSession = auth.AuthSession`, `type LibraryClient =
library.Client`, etc.), so `import ".../dtrpg-sdk.go"` alone covers the common path, same as
`use dtrpg_sdk::{...}` in Rust — while `auth`/`library` stay importable directly for callers
who only need one slice. Alternative considered: one flat package for the whole SDK
(simpler, no aliasing) — rejected because it would bury the library-client surface (~10
methods, ~25 types) alongside auth internals in a single namespace, which is exactly what
the Rust module split was avoiding.

**HTTP client.** Stdlib `net/http.Client`, not a third-party HTTP library. Rust needed
`reqwest` because the stdlib has no ergonomic async client; Go's stdlib client is already
the idiomatic choice, and it keeps the module's only non-stdlib dependency limited to the
OpenAPI-generation YAML parser (see below).

**Context propagation.** Every method that performs I/O (`LibraryClient.ListOrderProducts`,
`auth.Authenticate`, `auth.LoginWithCredentials`, etc.) takes `context.Context` as its first
parameter. Rust's `async fn` gets cancellation for free from the runtime; Go requires the
caller to pass it explicitly. This is the one place the Go API shape necessarily diverges
from a literal Rust port.

**Errors.** Replace Rust's `SdkError`/`AuthSessionError`/`ClientError` enums with Go error
types satisfying the `error` interface, composed via `errors.Is`/`errors.As` and `%w`
wrapping instead of enum variants:
- `dtrpg.ErrUnconfigured`, `dtrpg.ErrUnauthenticated` — sentinel errors (`errors.New`),
  checked with `errors.Is`.
- `dtrpg.AuthSessionError{ErrorCode, Message, AuthState}` — struct implementing `Error()`,
  checked with `errors.As`.
- `library.ClientError` becomes a small family: `library.ErrInvalidCredentials` (sentinel),
  `library.ApplicationKeyRequestError{Status string}`, `library.DecodeError{URL, Status,
  Cause, Payload}`, `library.APIError{URL, Status, Message, Payload, RetryAfter
  *time.Duration}` — all implementing `Error()` and `Unwrap()` where they wrap an inner
  `error` (e.g. the underlying `*http.Response`/JSON decode error), checked with
  `errors.As`. This preserves the Rust SDK's distinction between "the API told us
  something" (`APIError`) and "we couldn't parse what it told us" (`DecodeError`).

**Response decoding.** A single internal `decodeResponse[T any](resp *http.Response) (T,
error)` generic helper (Go 1.18+ generics) mirrors Rust's `decode_response<T>`: non-2xx
responses are never parsed as `T` — instead `extractErrorMessage` inspects the body for a
top-level `message`, a nested `{"error":{"message":...}}`, or field-keyed validation-error
maps, and the `Retry-After` header (integer delay-seconds only, per RFC 9110 §10.2.3 — no
HTTP-date support) is parsed into `APIError.RetryAfter`. 2xx responses that fail to decode
into `T` return `DecodeError` with a truncated (2000-byte) payload logged, not the full body.
**Both delete methods route through this same helper**, unlike the Rust SDK where
`delete_product_list`/`delete_product_list_item` bypass `decode_response` via
`error_for_status()` and lose `Retry-After`/message extraction on failure — see Open
Questions on whether to backport this fix to Rust too.

**applicationKey on library requests.** The Rust SDK's `rust-library-client` spec claims
every library request carries both `applicationKey` and the bearer token, but the actual
`LibraryClient` code only sends the `Authorization` header. The Go port matches the Rust
SDK's *runtime behavior* (bearer token only, no `applicationKey` query param on library
calls) rather than its spec text, so the two SDKs produce identical requests. The Go spec
for `go-library-client` states this explicitly instead of inheriting the discrepancy.

**Credential-login-exchange target host.** `auth.LoginWithCredentials` targets
`https://www.drivethrurpg.com` directly (`/validate_login_credentials.php`,
`/create_account_app.php`) rather than the configured API base URL, exactly as the Rust
SDK does — `Config` is accepted for API symmetry but unused by this flow.

**OpenAPI metadata generation.** Go has no build-script/`OUT_DIR` equivalent to Rust's
`build.rs`. Use `go generate`: an `internal/opgen` command reads `API/openapi.yaml` (the
`dtrpg-api` submodule, added at `API/` exactly as in the Rust repo) and writes
`openapi/generated.go` (checked into the repo, not gitignored, since Go builds don't run
generators automatically). `go:generate` directive lives in `openapi/openapi.go`. Unlike
Rust's hand-rolled line parser, use `gopkg.in/yaml.v3` (the one non-stdlib, non-test
dependency this module needs) to parse `servers[0].url` and every `paths` entry's
HTTP-method + path into `DefaultServerURL` and `Operations []Operation`. CI runs
`go generate ./... && git diff --exit-code openapi/generated.go` to catch stale generated
output, the same contract-freshness check the Rust build performs on every compile.

**JSON:API envelope handling.** Rust's custom `Deserialize` impl for
`ProductListItemCreateResponse` (unwrapping `data.attributes.*`) becomes a custom
`UnmarshalJSON` method in Go, same pattern used for the `null`-as-default fields (a small
`nullAsDefault[T]` generic helper wrapping `json.Unmarshal` per-field instead of a
serde-level deserializer).

## Risks / Trade-offs

- [Go's stdlib JSON lacks serde's `#[serde(rename)]` ergonomics for the JSON:API camelCase
  fields] → struct tags (`json:"currentPage"`) cover the common case; the handful of
  null-as-default and envelope-unwrapping fields get explicit `UnmarshalJSON` methods,
  scoped narrowly to those types only.
- [Diverging from the Rust SDK's `rust-library-client` spec text on `applicationKey`
  could look like an oversight rather than a deliberate parity choice] → the `go-library-
  client` spec states the bearer-token-only behavior as a requirement, not an omission, and
  this design doc records the reasoning for future readers.
- [Adding `gopkg.in/yaml.v3` as this module's only dependency outside `go generate` tooling
  contradicts a "stdlib-only" expectation some SDK consumers might have] → it is a `//
  go:generate`-time dependency only, isolated to `internal/opgen`; it does not appear in
  the module's runtime import graph or `go.sum` for consumers who only import the root/
  `auth`/`library` packages... (Go's module graph does include it regardless of build tags,
  so this is a soft mitigation, not a hard guarantee) → tracked as an open question below.
- [`context.Context` on every method is a visible API shape difference from Rust] →
  unavoidable and idiomatic; documented in the proposal so it isn't mistaken for scope
  creep during review.

## Migration Plan

Greenfield module, no rollback concerns beyond reverting the branch. Build order:
1. `go mod init`, add `dtrpg-api` as a git submodule at `API/`.
2. `Config` + sentinel/struct error types (`dtrpg` package).
3. `auth` package: `AuthSession`/`AuthState`/`AuthTokenResponse`, key exchange, credential
   login — each independently testable against `httptest.Server`.
4. `library` package: models first (JSON round-trip tests), then `Client` methods.
5. `Sdk` struct wiring `Config` → `auth.AuthSession` → `library.Client`.
6. `internal/opgen` + `openapi/generated.go`, wired into `go generate ./...`.
7. README + package doc comments; defer CI/release automation to the follow-up change
   noted in the proposal's Non-Goals.

## Open Questions

- Should the Rust SDK's two `error_for_status()`-based delete methods be fixed to route
  through `decode_response` (matching this design's Go behavior), so both SDKs are
  consistent — filed as a follow-up issue against `dtrpg-sdk/rust`, or left as a known
  divergence?
- Is `gopkg.in/yaml.v3` an acceptable dependency for `go generate`-time-only use, or should
  the Go port instead replicate Rust's hand-rolled line-based YAML parsing in `internal/
  opgen` to keep the module dependency-free end to end?
