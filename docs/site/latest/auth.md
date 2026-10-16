---
title: Authentication
version: latest
slug: auth
---

# Authentication

The SDK separates two flows:

- **`auth.LoginWithCredentials`** exchanges an email/password pair for a DriveThruRPG
  application key, against `www.drivethrurpg.com`.
- **`auth.Authenticate`** exchanges an application key for a short-lived session (JWT
  access token + refresh token), against the configured API base URL.

An application key obtained once (interactively, or via `LoginWithCredentials`) can be
stored and reused across runs; only `Authenticate` needs to run each time a session
expires.

## Configuring the SDK

```go
sdk := dtrpg.NewSdkWithConfig(dtrpg.NewConfig("my-application-key"))
```

`dtrpg.NewConfig` targets the production API. Use `dtrpg.NewConfigWithBaseURL` to point
at a staging server or local mock.

## Exchanging credentials for an application key

```go
ctx := context.Background()
appKey, err := auth.LoginWithCredentials(ctx, email, password, *sdk.Config())
if err != nil {
	log.Fatal(err)
}
```

`LoginWithCredentials` returns `library.ErrInvalidCredentials` when the email/password
pair is rejected, or an `*library.ApplicationKeyRequestError` if credentials validate but
key issuance fails.

## Exchanging an application key for a session

```go
response, err := auth.Authenticate(ctx, appKey, *sdk.Config())
if err != nil {
	log.Fatal(err)
}

session, err := sdk.ApplyAuthResponse(response)
if err != nil {
	log.Fatal(err)
}
fmt.Println(session.Token())
```

`sdk.ApplyAuthResponse` stores the session on the `Sdk` instance and returns
`dtrpg.ErrUnconfigured` if the SDK has no `Config` yet.

## Session lifecycle

- `sdk.RequireSession()` returns the active session, or `dtrpg.ErrUnauthenticated` if
  there is none — use this in call chains that need a session to proceed.
- `sdk.InvalidateSession(authErr)` clears the session after the API reports an
  authentication failure, and returns the given error for propagation.
- `sdk.ClearSession()` clears the session for a voluntary logout, without recording an
  error.

```go
client, err := sdk.LibraryClient()
if err != nil {
	// dtrpg.ErrUnconfigured or dtrpg.ErrUnauthenticated
	log.Fatal(err)
}
```

`sdk.LibraryClient()` validates both configuration and session before constructing a
`library.Client` — prefer it over calling `library.NewClient` directly.
