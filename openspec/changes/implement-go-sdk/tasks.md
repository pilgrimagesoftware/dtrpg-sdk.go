## 1. Module Setup

- [x] 1.1 Run `go mod init github.com/pilgrimagesoftware/dtrpg-sdk.go`, set `go 1.22` (or
  current stable) in `go.mod`.
- [x] 1.2 Add `dtrpg-api` as a git submodule at `API/`, matching `dtrpg-sdk/rust/API`.
- [x] 1.3 Add `.gitignore`, `.gitmodules`, and CLAUDE.md/AGENTS.md symlink consistent with
  the other language SDKs.

## 2. Config and Errors (`go-sdk-configuration`)

- [x] 2.1 Implement `Config` (application key, base URL, API version) with
  `NewConfig(applicationKey)` and `NewConfigWithBaseURL(applicationKey, baseURL)`
  constructors and DriveThruRPG's default base URL/API version constants.
- [x] 2.2 Implement `dtrpg.ErrUnconfigured` and `dtrpg.ErrUnauthenticated` sentinel errors.
- [x] 2.3 Implement `dtrpg.AuthSessionError{ErrorCode, Message, AuthState}` satisfying
  `error`.
- [x] 2.4 Unit tests: default values, custom base URL override, config equality.

## 3. Auth Session Lifecycle (`go-session-lifecycle`)

- [x] 3.1 Implement `auth.AuthState` (typed string constants: `Unauthenticated`,
  `TokenInvalid`, `TokenExpired`, `RefreshExpired`, `Unauthorized`) with `AsAPIString()`.
- [x] 3.2 Implement `auth.AuthTokenResponse{Token, RefreshToken, RefreshTokenTTL}` with
  `json` tags matching the API's camelCase field names.
- [x] 3.3 Implement `auth.AuthSession` with `FromAPIResponse`, `Token()`,
  `RefreshToken()`, `RefreshTokenTTL()`, `RefreshTokenExpiredAt(unixTimestamp) bool`, and
  `Invalidate(err) SessionTransition`.
- [x] 3.4 Implement `auth.SessionTransition{NextSession *AuthSession, Err
  AuthSessionError}`.
- [x] 3.5 Unit tests: session construction, expiry boundary, invalidate transitions for
  each `AuthState`.

## 4. Key Exchange Authentication (`go-authentication-flow`)

- [x] 4.1 Implement `auth.Authenticate(ctx, apiKey, config) (AuthTokenResponse, error)`:
  `POST {base}/{version}/auth_key?applicationKey=<key>`, empty JSON body,
  `Content-Type: application/json`.
- [x] 4.2 Wire errors through the shared `library` error-decoding path (see 6.x) so
  failures surface as `library.APIError`/`library.DecodeError`.
- [x] 4.3 Unit tests against `httptest.Server`: success, non-2xx, malformed body.

## 5. SDK Entry Point

- [x] 5.1 Implement `dtrpg.Sdk{config *Config, session *auth.AuthSession}` with `NewSdk()`,
  `NewSdkWithConfig(config)`, `Configure(config)`, `Config() *Config`,
  `Session() *auth.AuthSession`.
- [x] 5.2 Implement `sdk.ApplyAuthResponse(response) (*auth.AuthSession, error)`,
  `sdk.ClearSession()`, `sdk.InvalidateSession(err) (AuthSessionError, error)`.
- [x] 5.3 Implement `sdk.LibraryClient() (*library.Client, error)` requiring both config
  and an active session.
- [x] 5.4 Add root-package type aliases for `auth`/`library` types callers need at the top
  level (`AuthSession`, `AuthState`, `LibraryClient`, etc.), per design.md's package-layout
  decision.
- [x] 5.5 Unit tests: full lifecycle (unconfigured → configured → authenticated →
  library client → invalidate → clear).

## 6. Library Client Core (`go-library-client`)

- [x] 6.1 Implement `library.Client{http *http.Client, config Config, token string}` with
  `NewClient(config, token)`.
- [x] 6.2 Implement the shared `decodeResponse[T any](resp *http.Response) (T, error)`
  helper: non-2xx → `extractErrorMessage` + `Retry-After` parsing → `APIError`; 2xx decode
  failure → `DecodeError` with truncated (2000-byte) payload.
- [x] 6.3 Implement `library.ErrInvalidCredentials`, `library.ApplicationKeyRequestError`,
  `library.DecodeError`, `library.APIError{URL, Status, Message, Payload, RetryAfter
  *time.Duration}`.
- [x] 6.4 Ensure every request sets `Authorization: <token>` (raw JWT, no `Bearer` prefix)
  and never attaches an `applicationKey` query parameter.
- [x] 6.5 Unit tests for `decodeResponse`: success, API error with message, API error with
  `Retry-After` (valid seconds, HTTP-date, absent), decode failure on 2xx.

## 7. Library Client Endpoints (`go-library-client`)

- [x] 7.1 `ListOrderProducts(ctx, params LibraryItemsParams) (OrderProductListResponse,
  error)` — GET `order_products` with `page`, `pageSize`, `getChecksum=1`,
  `getFilters=1`, `library=true`, `archived`, `updatedDate[after]`.
- [x] 7.2 `GetOrderProduct(ctx, orderProductID uint64) (OrderProductItemResponse, error)`
  — GET `order_products/{id}`.
- [x] 7.3 `PrepareDownload(ctx, orderProductID uint64, index int) (map[string]any, error)`
  — GET `order_products/{id}/prepare?index={index}`, required `index`.
- [x] 7.4 `ListProductLists(ctx, params PageParams) (ProductListCollectionResponse,
  error)` — GET `product_lists`.
- [x] 7.5 `ListProductListItems(ctx, productListID uint64, params PageParams)
  (ProductListItemsResponse, error)` — GET `product_list_items?productListId={id}`.
- [x] 7.6 `CreateProductList(ctx, name string) (ProductListItem, error)` — POST
  `product_lists`, unwraps JSON:API envelope.
- [x] 7.7 `DeleteProductList(ctx, id uint64) error` — DELETE `product_lists/{id}`, routed
  through `decodeResponse` per the `go-library-client` spec (not bypassed like Rust's
  `error_for_status()`).
- [x] 7.8 `AddProductListItem(ctx, productListID, productID uint64)
  (ProductListItemCreateResponse, error)` — POST `product_list_items`.
- [x] 7.9 `DeleteProductListItem(ctx, productListItemID uint64) error` — DELETE, also
  routed through `decodeResponse`.
- [x] 7.10 Integration tests per method against `httptest.Server`, covering success and
  API-error paths.

## 8. Library Model Types (`go-library-types`)

- [x] 8.1 Pagination types: `PaginationLinks`, `PaginationMeta`.
- [x] 8.2 Order product types: `FileChecksum`, `OrderProductFile` (with null-as-default
  `Checksums`), `OrderProductFilter`, `OrderProductHistoryEntry`, `OrderProductAttribute`,
  `OrderProductAttributes`, `OrderProductPublisher`, `OrderProductDescription`,
  `OrderProductInfo`, `OrderProductOrder`, `OrderProductItem`,
  `OrderProductRelationships`, `RelationshipRef`, `RelationshipData`.
- [x] 8.3 Publisher/included types: `PublisherAttributes`, `PublisherItem`,
  `IncludedItem` with `AsPublisher()`/`AsProduct()` resource-type-tagged decode helpers.
- [x] 8.4 Response envelopes: `OrderProductListResponse`, `OrderProductItemResponse`.
- [x] 8.5 Product list types: `ProductListAttributes`, `ProductListItem`,
  `ProductListCollectionResponse`, `ProductListItemsResponse`,
  `ProductListItemCreateRequest`, `ProductListItemCreateResponse` (custom
  `UnmarshalJSON` unwrapping the JSON:API envelope).
- [x] 8.6 Query param types: `LibraryItemsParams`, `PageParams` with sane zero-value
  defaults.
- [x] 8.7 Shared `nullAsDefault[T]` decode helper used by fields the API may return as
  `null`.
- [x] 8.8 JSON round-trip tests for every type against representative API fixture
  payloads (reuse or adapt fixtures from `dtrpg-sdk/rust` tests where useful).

## 9. Credential Login Exchange (`credential-login-exchange`)

- [x] 9.1 Implement `auth.LoginWithCredentials(ctx, email, password, config) (string,
  error)` in `auth/credential_login.go`, targeting `https://www.drivethrurpg.com`.
- [x] 9.2 Implement the typed positional-array response for
  `validate_login_credentials.php` via custom `UnmarshalJSON`.
- [x] 9.3 Implement the `create_account_app.php` call and `status`/`message.key`
  handling, returning `library.ApplicationKeyRequestError` on non-success status.
- [x] 9.4 Document in package comments that this flow is independent of
  `auth.Authenticate` and does not call `api.drivethrurpg.com`.
- [x] 9.5 Unit tests: valid credentials, invalid credentials (stops before key request),
  key request failure after valid credentials, parsing the documented `LOGIN.md` example
  response.

## 10. Retry-After Exposure (`rate-limit-retry-after`)

- [x] 10.1 Confirm `decodeResponse` (6.2) parses `Retry-After` as delay-seconds only (no
  HTTP-date support), attaching `nil` on absent/unparseable header.
- [x] 10.2 Confirm `DeleteProductList`/`DeleteProductListItem` (7.7, 7.9) surface
  `RetryAfter` identically to read endpoints — this is the one behavioral improvement over
  the Rust SDK called out in design.md.
- [x] 10.3 Unit tests: 429 with `Retry-After: 30`, no header, unparseable header (HTTP-date
  string), success response has no error to attach to.

## 11. OpenAPI Metadata Generation

- [x] 11.1 Implement `internal/opgen` command: parse `API/openapi.yaml` via
  `gopkg.in/yaml.v3` for `servers[0].url` and every `paths` entry's method+path.
- [x] 11.2 Add `//go:generate` directive in `openapi/openapi.go` invoking `internal/opgen`
  to write `openapi/generated.go` (`DefaultServerURL`, `Operations []Operation`).
- [x] 11.3 Commit initial `openapi/generated.go`; add a test asserting
  `DefaultServerURL` and a known operation are present, mirroring the Rust SDK's build
  sanity test.
- [x] 11.4 Document the `go generate ./... && git diff --exit-code openapi/generated.go`
  staleness check for CI (CI wiring itself is out of scope per the proposal's Non-Goals —
  document the command for the follow-up change to adopt).

## 12. Documentation and Polish

- [x] 12.1 Write package doc comments for `dtrpg`, `auth`, `library`, `openapi` (every
  exported identifier per `docs/go.md`'s "exported symbols get doc comments" rule).
- [x] 12.2 Update `README.md` with installation (`go get`), quick-start example mirroring
  the Rust SDK's README quick start, and the `go generate`/submodule setup steps.
- [x] 12.3 Run `gofmt -w .`, `go vet ./...`, `golangci-lint run`, `go test -race ./...`
  and fix all findings before marking the change complete.
