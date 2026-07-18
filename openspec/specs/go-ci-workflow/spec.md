# go-ci-workflow Specification

## Purpose
Define the automated build/format/vet/lint/test gate that runs on every pull request and
every push to `develop` for the Go SDK, matching the Rust SDK's CI scope adapted to Go's
toolchain.

## Requirements

### Requirement: CI runs on every push to develop
The repository SHALL run a CI workflow on every push to the `develop` branch that checks
out the `dtrpg-api` submodule and runs `gofmt -l`, `go vet ./...`, `golangci-lint run
./...`, and `go test -race ./...`.

#### Scenario: Push to develop triggers CI
- **WHEN** a commit is pushed to `develop`
- **THEN** the CI workflow runs formatting, vet, lint, and race-enabled test checks
  against that commit with submodules checked out

#### Scenario: A failing check fails the workflow
- **WHEN** any of `gofmt -l`, `go vet`, `golangci-lint run`, or `go test -race` reports a
  problem
- **THEN** the CI workflow run fails

### Requirement: PR validation runs the same checks against master and develop
The repository SHALL run the same checks (`gofmt -l`, `go vet ./...`, `golangci-lint run
./...`, `go test -race ./...`) on every pull request targeting `master` or `develop`,
independent of the `develop`-push CI workflow.

#### Scenario: PR opened against develop
- **WHEN** a pull request is opened or updated targeting `develop`
- **THEN** the PR validation workflow runs and reports status on the PR

#### Scenario: PR opened against master
- **WHEN** a pull request is opened or updated targeting `master` (e.g. a `release/*`
  branch)
- **THEN** the PR validation workflow runs and reports status on the PR
