# go-sdk-configuration Specification

## Purpose
Define how the Go SDK accepts and applies caller configuration so client initialization
stays explicit, idiomatic Go, and aligned with the underlying API contract.

## Requirements

### Requirement: Go SDK configuration must be explicit before use
The Go SDK MUST require the configuration values needed to initialize its client behavior
before authenticated operations are attempted, including the values needed to drive the
Go-owned auth/session behavior.

#### Scenario: Attempting to use the SDK without configuration
- **WHEN** a caller invokes authenticated Go SDK behavior before providing a `Config`
- **THEN** the SDK returns an error satisfying `errors.Is(err, dtrpg.ErrUnconfigured)`

#### Scenario: Providing configuration for auth/session behavior
- **WHEN** a caller calls `dtrpg.NewSdk(config)` or `Sdk.Configure(config)` with a valid
  `Config`
- **THEN** the Go SDK has the configuration needed to apply its documented authentication
  and session lifecycle behavior

### Requirement: Go SDK configuration must remain idiomatic
The Go SDK MUST expose configuration in a form that fits Go conventions (exported struct
fields or constructor functions, no builder-pattern chaining required) while preserving
the underlying API contract requirements.

#### Scenario: Configuring a custom API endpoint
- **WHEN** a caller provides a supported custom base URL or application key through
  `dtrpg.Config` (e.g. `dtrpg.NewConfigWithBaseURL(applicationKey, baseURL)`)
- **THEN** the SDK applies that configuration using the documented Go configuration model

#### Scenario: Default base URL and API version apply when not overridden
- **WHEN** a caller constructs `dtrpg.NewConfig(applicationKey)` without specifying a base
  URL or API version
- **THEN** the resulting `Config` uses DriveThruRPG's documented default base URL and API
  version

### Requirement: Go SDK API metadata must be generated from OpenAPI
The Go SDK MUST generate API metadata from `API/openapi.yaml` via `go generate` so the
module remains tied to the same API contract used by the other SDKs, and CI MUST verify
the generated output is not stale.

#### Scenario: Regenerating Go SDK OpenAPI metadata
- **WHEN** a developer runs `go generate ./...` after `API/openapi.yaml` changes
- **THEN** `openapi/generated.go` is rewritten with the current default server URL and
  operation list

#### Scenario: CI detects stale generated metadata
- **WHEN** CI runs `go generate ./...` against a commit where `openapi/generated.go` was
  not regenerated after an `API/openapi.yaml` change
- **THEN** the resulting diff is non-empty and CI fails
