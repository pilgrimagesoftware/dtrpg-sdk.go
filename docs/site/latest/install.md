---
title: Installation
version: latest
slug: install
---

# Installation

Requires Go 1.24+.

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

## Development checks

```bash
go build ./...
go vet ./...
gofmt -l .
golangci-lint run
go test -race ./...
```
