# dtrpg-sdk.go

[![Go Reference](https://pkg.go.dev/badge/github.com/pilgrimagesoftware/dtrpg-sdk.go.svg)](https://pkg.go.dev/github.com/pilgrimagesoftware/dtrpg-sdk.go)
[![CI](https://github.com/pilgrimagesoftware/dtrpg-sdk.go/actions/workflows/ci.yaml/badge.svg)](https://github.com/pilgrimagesoftware/dtrpg-sdk.go/actions/workflows/ci.yaml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE.md)

A Go SDK for the [DriveThruRPG API](https://api.drivethrurpg.com).

Provides configuration, authentication/session lifecycle, and a library client for listing
orders, product lists, and preparing downloads.

Requires Go 1.24+.

## Installation

```bash
go get github.com/pilgrimagesoftware/dtrpg-sdk.go
```

## Building from source

This repository uses the `dtrpg-api` repository as a submodule (`API/`). The `openapi`
package's `go:generate` directive reads `API/openapi.yaml` and writes build-time OpenAPI
metadata (`openapi.DefaultServerURL`, `openapi.Operations`) into `openapi/generated.go`.
Clone with submodules, or initialize them after cloning:

```bash
git clone --recursive https://github.com/pilgrimagesoftware/dtrpg-sdk.go.git

# or, if already cloned:
git submodule update --init --recursive
```

Regenerate the metadata after `API/openapi.yaml` changes:

```bash
go generate ./...
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	dtrpg "github.com/pilgrimagesoftware/dtrpg-sdk.go"
)

func main() {
	sdk := dtrpg.NewSdkWithConfig(dtrpg.NewConfig("my-app-key"))

	// After receiving an auth response from the API:
	response := dtrpg.NewAuthTokenResponse("jwt-token", "refresh-token", 1_800_000_000)
	session, err := sdk.ApplyAuthResponse(response)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(session.Token())

	// Create an authenticated library client:
	client, err := sdk.LibraryClient()
	if err != nil {
		log.Fatal(err)
	}
	products, err := client.ListOrderProducts(context.Background(), dtrpg.LibraryItemsParams{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(products)
}
```

See the package documentation (`go doc`) for the full API reference, including `Config`,
`AuthSession`/`AuthState`, `LibraryClient`, and the library model types
(`OrderProductItem`, `ProductListItem`, etc.).

## Development

```bash
go build ./...
go vet ./...
gofmt -l .
golangci-lint run
go test -race ./...
```

## Release Process

See [RELEASE.md](RELEASE.md).

## License

MIT — see [LICENSE.md](LICENSE.md).
