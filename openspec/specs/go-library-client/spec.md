# go-library-client Specification

## Purpose
Define the authenticated Go client for DriveThruRPG library endpoints: order products,
product lists, and product list items.

## Requirements

### Requirement: The Go SDK MUST provide a context-aware HTTP client for all library endpoints
The Go SDK MUST expose a `library.Client` whose methods accept `context.Context` and that
callers can use to fetch ordered products, product details, download preparation, product
lists, and product list items from the DriveThruRPG API.

#### Scenario: Fetching the user's library
- **WHEN** a caller invokes `client.ListOrderProducts(ctx, params)`
- **THEN** the SDK sends the authenticated request and returns a deserialized
  `OrderProductListResponse`

### Requirement: The Go library client MUST be created from a configured, authenticated SDK instance
`library.Client` MUST require both SDK configuration and an active authentication session
before it can make API calls.

#### Scenario: Creating a library client from the SDK
- **WHEN** a caller calls `sdk.LibraryClient()` on a `dtrpg.Sdk` instance
- **THEN** the SDK returns an error if the instance is unconfigured or unauthenticated,
  and returns a usable `library.Client` otherwise

### Requirement: The Go library client authenticates requests using the bearer token only
Every library API request MUST include the `Authorization: <token>` header (the raw JWT,
no `Bearer` prefix) from the active session. Library requests MUST NOT attach an
`applicationKey` query parameter — application-key exchange is scoped to `auth.Authenticate`
and `auth.LoginWithCredentials`, not to library calls.

#### Scenario: Sending an authenticated library request
- **WHEN** the Go library client sends a request to any library endpoint
- **THEN** the request includes the bearer token from the active session in the
  `Authorization` header and does not include an `applicationKey` query parameter

### Requirement: Download preparation MUST include the target file's index
`PrepareDownload` MUST require the caller to supply the target file's `index` (its
position within the ordered product's file list) as a required parameter, matching what
the DriveThruRPG API enforces.

#### Scenario: Preparing a download with a valid index
- **WHEN** a caller invokes `client.PrepareDownload(ctx, orderProductID, index)` with a
  valid `orderProductID` and the target file's `index`
- **THEN** the SDK sends the request with the index included and returns the deserialized
  response on success

#### Scenario: No default index is silently assumed
- **WHEN** a caller invokes `PrepareDownload`
- **THEN** the Go method signature requires an explicit `index int` argument — there is no
  overload that omits it

### Requirement: Delete operations decode errors consistently with other endpoints
`DeleteProductList` and `DeleteProductListItem` MUST route their responses through the
same error-decoding path as every other `library.Client` method, so failures produce a
`library.APIError` with an extracted message and `RetryAfter` value rather than a bare
transport error.

#### Scenario: Deleting a product list that does not exist
- **WHEN** `client.DeleteProductList(ctx, id)` is called for an `id` the API rejects with
  a non-success status
- **THEN** the returned error is a `library.APIError` with the API's extracted message and
  any `Retry-After` value populated, consistent with `ListOrderProducts` and other read
  methods on the same failure path
