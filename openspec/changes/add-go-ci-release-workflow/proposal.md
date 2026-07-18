## Why

`implement-go-sdk` (merged) built the Go SDK itself but explicitly deferred CI and release
automation as a follow-up — the module currently has no automated build/lint/test gate and
no way to cut a versioned, tagged release. `dtrpg-sdk/rust` already has this fully
automated (`ci.yaml`, `pr.yaml`, `prepare-release.yaml`, `tag-release.yaml`, `release.yaml`
plus `git-cliff`-driven changelog/version bumping); this change brings the Go SDK to
parity using the same git-flow shape (`develop` → `release/*` → `master`, tag-triggered
publish) adapted to Go's toolchain and Go's proxy-based module publishing model (no
`cargo publish` equivalent — a module is "published" the moment its tag is pushed and the
Go module proxy indexes it).

## What Changes

- Add `.github/workflows/ci.yaml`: on push to `develop`, run `gofmt -l`, `go vet`,
  `golangci-lint run`, `go test -race ./...` with the `dtrpg-api` submodule checked out.
- Add `.github/workflows/pr.yaml`: the same checks on every pull request targeting
  `master` or `develop`.
- Add `git-cliff` configuration (`cliff.toml`) mapping Conventional Commits to changelog
  groups and semantic version bumps, matching the Rust SDK's grouping/exclusion rules.
- Add `.github/workflows/prepare-release.yaml`: manually triggered (`workflow_dispatch`)
  workflow that runs `git-cliff --bump` against `develop`, opens a `release/<version>`
  branch and PR into `master` with the version bump reflected in a `VERSION` file (Go
  modules have no `Cargo.toml`-equivalent single version field; module versions are
  derived from git tags) and the updated `CHANGELOG.md`.
- Add `.github/workflows/tag-release.yaml`: on a `release/*` branch PR merging into
  `master`, tags the merge commit `v<version>` and pushes the tag.
- Add `.github/workflows/release.yaml`: on `v*` tag push, runs the full test suite,
  triggers Go module proxy indexing (`GOPROXY=proxy.golang.org go list -m
  module@version`) so the tagged version is fetchable via `go get` immediately, generates
  a GitHub Release with the `git-cliff`-generated changelog section for that tag, and
  merges `master` back into `develop`.
- Update `README.md` with CI/release status badges and a link to a new `RELEASE.md`
  documenting the release process, mirroring the Rust SDK's README.
- Add `RELEASE.md` documenting the git-flow release process for this module.

## Capabilities

### New Capabilities
- `go-ci-workflow`: automated build/format/vet/lint/test gate on every PR and every push
  to `develop`.
- `go-module-release-publishing`: automated GitHub Release creation and Go module proxy
  indexing triggered by version tags on `master`, so every version bump produces an
  installable, released module.

### Modified Capabilities
- (none)

## Impact

- New `.github/workflows/*.yaml` files, `cliff.toml`, `RELEASE.md` in `dtrpg-sdk/go`.
- `README.md` gains badges and a release-process link.
- No changes to SDK source code, `dtrpg-api`, or the other language SDKs.
