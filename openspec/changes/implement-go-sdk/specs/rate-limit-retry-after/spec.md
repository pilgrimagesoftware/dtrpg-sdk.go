## ADDED Requirements

### Requirement: library.APIError MUST expose the Retry-After header when present
When `library.Client` receives a non-success HTTP response, the resulting
`library.APIError` SHALL include a `RetryAfter *time.Duration` field populated from the
response's `Retry-After` header when that header is present and specifies a delay in
seconds. This applies uniformly to every `library.Client` method, including
`DeleteProductList` and `DeleteProductListItem`.

#### Scenario: 429 response with a delay-seconds Retry-After header
- **WHEN** the API returns HTTP 429 with a `Retry-After: 30` header
- **THEN** the returned `library.APIError.RetryAfter` is a pointer to `30 * time.Second`

#### Scenario: Non-success response with no Retry-After header
- **WHEN** the API returns a non-success status with no `Retry-After` header present
- **THEN** the returned `library.APIError.RetryAfter` is `nil`

#### Scenario: Retry-After header present but unparseable as delay-seconds
- **WHEN** the API returns a non-success status with a `Retry-After` header that is not a
  valid non-negative integer (e.g. an HTTP-date value)
- **THEN** the returned `library.APIError.RetryAfter` is `nil`, and no error is raised
  solely due to the unparseable header

#### Scenario: Success response is unaffected
- **WHEN** the API returns a success status
- **THEN** no `RetryAfter` value is computed or attached to any error (there is no error
  to attach it to), and response decoding proceeds as documented in `go-library-types`
