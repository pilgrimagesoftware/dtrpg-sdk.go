---
title: dtrpg-sdk.go documentation
version: latest
slug: index
---

# dtrpg-sdk.go documentation

dtrpg-sdk.go is a Go SDK for the [DriveThruRPG API](https://api.drivethrurpg.com). It
provides configuration, authentication/session lifecycle, and a library client for
listing orders, product lists, and preparing downloads. This is the latest documentation.

## Getting started

- [Installation]({{ '/latest/install/' | relative_url }}) - `go get` the module and
  build from source, including the `dtrpg-api` submodule.
- [Authentication]({{ '/latest/auth/' | relative_url }}) - exchange credentials or an
  application key for a session, and store it on the SDK.

## Using the API

- [Library client]({{ '/latest/library/' | relative_url }}) - list ordered products and
  product lists, fetch a single order product, and prepare a download.
- [Errors]({{ '/latest/errors/' | relative_url }}) - the `APIError`, `DecodeError`, and
  sentinel error types returned by the SDK, and how to inspect them.

Select a release from the version picker when working with an earlier build. Full
field-level reference lives in the generated package documentation
([pkg.go.dev](https://pkg.go.dev/github.com/pilgrimagesoftware/dtrpg-sdk.go)) or `go doc`.
