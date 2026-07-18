## Context

`dtrpg-sdk/rust` has a complete git-flow release pipeline: `ci.yaml`/`pr.yaml` gate every
push/PR, `prepare-release.yaml` (manual trigger) opens a `release/<version>` branch/PR from
`develop` into `master` with `git-cliff`-bumped version and changelog, `tag-release.yaml`
tags the merge commit when that PR lands on `master`, and `release.yaml` (tag-triggered)
publishes to crates.io and cuts a GitHub Release. `dtrpg-sdk/go` has none of this yet.

Go's publishing model differs fundamentally from Cargo's: there is no registry to push to
and no `cargo publish` step. A Go module is "published" the instant its tag exists on the
origin remote — `go get module@vX.Y.Z` works immediately via direct VCS fetch. The Go
module proxy (`proxy.golang.org`, the default `GOPROXY` for most Go installs) caches and
indexes modules lazily on first request rather than via a push, so `pkg.go.dev` and proxy
consumers only see a version after *something* requests it once.

## Goals / Non-Goals

**Goals:**
- Every PR and every push to `develop` runs `gofmt`, `go vet`, `golangci-lint`, and
  `go test -race` — matching the Rust SDK's CI gate scope, adapted to Go's tools.
- A tagged `master` release is immediately `go get`-able and appears on `pkg.go.dev`
  without a maintainer needing to manually request proxy indexing.
- Changelog and version-bump automation via `git-cliff`, reusing the same Conventional
  Commits grouping rules as the Rust SDK where they're language-agnostic (only the
  `cargo`-specific parsers/skip rules don't carry over).

**Non-Goals:**
- No Cargo-equivalent registry publish step — Go has no such step to replicate.
- No cross-compilation/binary-artifact publishing — this is a library module, not a CLI
  binary; there's nothing to attach to the GitHub Release beyond the changelog.
- Not changing the SDK's own build/test commands — `go build ./...`, `go vet ./...`,
  `golangci-lint run`, `go test -race ./...` are already established and unchanged by this
  workflow.

## Decisions

**No `VERSION` file; module version comes entirely from git tags.** Go modules using
semantic import versioning don't carry a version string in `go.mod` (unlike
`Cargo.toml`'s `version` field) — the version is purely a property of the git tag used to
fetch the module. `git-cliff --bumped-version` computes the next version from commit
history exactly as it does for the Rust SDK; `prepare-release.yaml` uses that value only
to name the `release/<version>` branch and populate the changelog section header, with no
file in the repo needing to store the version number redundantly.

**Triggering proxy indexing.** After the tag is pushed, `release.yaml` runs `GOPROXY=
proxy.golang.org go list -m github.com/pilgrimagesoftware/dtrpg-sdk.go@v<version>` — a
read-only proxy request that forces `proxy.golang.org` to fetch and index the new tag
immediately, so `go get` and `pkg.go.dev` reflect the release without waiting on an
organic first consumer request. Alternative considered: do nothing and rely on the first
real `go get` to trigger indexing — rejected because it means the release "isn't really
available" for an indeterminate window after the GitHub Release is created, which is
confusing for anyone checking `pkg.go.dev` right after an announcement.

**git-cliff config reuse.** `cliff.toml` is copied from the Rust SDK with `cargo`/Rust-
specific pieces removed: the commit-type-to-group mapping (`feat`→Added, `fix`→Fixed,
etc.) and skip rules (`chore`, `ci`, `test`, `style`, `build` excluded from the changelog
but still counted for bump unless explicitly `bump = "skip"`) are identical, since they're
about Conventional Commits semantics, not the language.

**Same git-flow shape, adapted trigger names.** `tag-release.yaml` and
`prepare-release.yaml` are structurally identical to the Rust SDK's (same
`release/<semver>` branch-name convention, same "PR merged into master" trigger for
tagging) — only the checks each runs during PR validation differ (`go vet`/`golangci-lint`
instead of `clippy`/`cargo fmt`).

## Risks / Trade-offs

- [`proxy.golang.org` availability/rate limits could make the indexing-trigger step flaky]
  → the step is best-effort and non-blocking: if it fails, the GitHub Release still gets
  created, and the module remains fetchable via direct VCS (`go get` falls back to `git`
  fetch on proxy miss) — the workflow does not fail the job solely on this step's error.
- [No `cargo test`-equivalent registry dry-run before "publishing"] → since there's no
  registry push to get wrong, the primary release risk shifts to "did the tag point at a
  broken commit," which `release.yaml` mitigates by re-running the full test suite against
  the tagged ref before creating the GitHub Release (matching the Rust SDK's ordering:
  tests before anything externally visible happens).

## Open Questions

- Does the org want `pkg.go.dev` badge/link automation in the README (a static badge URL
  keyed to the module path), or is a manual one-time README edit sufficient? Defaulting to
  a one-time manual badge in this change; revisit if badge staleness becomes a problem.
