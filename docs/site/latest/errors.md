---
title: Errors
version: latest
slug: errors
---

# Errors

## SDK-level errors

Returned by the root `dtrpg` package:

- `dtrpg.ErrUnconfigured` — the SDK has no `Config`. Call `sdk.Configure` or construct
  with `NewSdkWithConfig`.
- `dtrpg.ErrUnauthenticated` — the SDK has no active session. Call
  `sdk.ApplyAuthResponse` with a token response from `auth.Authenticate`.

## Library client errors

Returned by `library.Client` methods and decoding helpers:

- **`*library.APIError`** — the API returned a non-success status. `Message` holds a
  human-readable explanation extracted from the response body (a top-level `message`
  field, `{"error": {"message": ...}}`, or field-keyed validation errors), falling back
  to the raw `Payload` when no message could be extracted. `RetryAfter` carries a parsed
  `Retry-After` header when present.
- **`*library.DecodeError`** — the response had a success status but the body could not
  be decoded into the expected type. Carries `Cause` (the underlying decode error) and a
  truncated `Payload` for diagnosis. Unwraps to `Cause`.

```go
products, err := client.ListOrderProducts(ctx, dtrpg.LibraryItemsParams{})
if err != nil {
	var apiErr *library.APIError
	if errors.As(err, &apiErr) {
		fmt.Printf("API error %d: %v\n", apiErr.Status, apiErr.Error())
	}
}
```

## Credential login errors

Returned by `auth.LoginWithCredentials`:

- `library.ErrInvalidCredentials` — the email/password pair was rejected by
  `validate_login_credentials.php`.
- **`*library.ApplicationKeyRequestError`** — credentials validated, but
  `create_account_app.php` returned a non-success status. `Status` holds the status
  string from the response.
