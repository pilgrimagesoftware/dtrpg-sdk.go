## 1. Changelog Configuration

- [x] 1.1 Add `cliff.toml`, adapted from `dtrpg-sdk/rust/cliff.toml`: same commit-group
  mapping (`feat`→Added, `fix`→Fixed, `perf`→Performance, `refactor`→Changed,
  `doc`→Documentation) and skip rules (`chore`, `ci`, `test`, `style`, `build`), with
  Cargo-specific parsers removed.
- [x] 1.2 Seed `CHANGELOG.md` with an `[Unreleased]` header matching the Rust SDK's
  changelog format.

## 2. CI Workflow (`go-ci-workflow`)

- [x] 2.1 Add `.github/workflows/ci.yaml`: triggers on push to `develop`; checks out with
  `submodules: recursive`; runs `gofmt -l .` (fails if non-empty), `go vet ./...`,
  `golangci-lint run ./...`, `go test -race ./...`.
- [x] 2.2 Add `.github/workflows/pr.yaml`: triggers on pull_request targeting `master` and
  `develop`; runs the same four checks as `ci.yaml`.
- [x] 2.3 Verify both workflows pass against the current `develop` HEAD (post
  `implement-go-sdk` merge) before merging this change.

## 3. Release Preparation and Tagging (`go-module-release-publishing`)

- [x] 3.1 Add `.github/workflows/prepare-release.yaml` (`workflow_dispatch`): install
  `git-cliff`, compute `git-cliff --bumped-version` against `develop`, update
  `CHANGELOG.md` via `git-cliff --tag v<version> --unreleased --prepend CHANGELOG.md`,
  open a `release/<version>` branch and PR into `master`.
- [x] 3.2 Add `.github/workflows/tag-release.yaml`: triggers on pull_request `closed`
  targeting `master`; guarded by `github.event.pull_request.merged == true &&
  startsWith(github.event.pull_request.head.ref, 'release/')`; validates the branch name
  matches `release/<semver>`, tags the merge commit `v<version>`, pushes the tag.

## 4. Release Publishing (`go-module-release-publishing`)

- [x] 4.1 Add `.github/workflows/release.yaml`: triggers on `v*` tag push (and
  `workflow_dispatch` with a `tag` input, matching the Rust SDK's re-run affordance);
  checks out the tagged ref with submodules; runs `go test -race ./...`.
- [x] 4.2 Add the Go module proxy indexing step: `GOPROXY=proxy.golang.org go list -m
  github.com/pilgrimagesoftware/dtrpg-sdk.go@${TAG}`, non-blocking on failure
  (`continue-on-error: true` or equivalent), per design.md's Decisions.
- [x] 4.3 Add changelog extraction (`git-cliff --current --strip header -o
  RELEASE_NOTES.md`) and GitHub Release creation (`softprops/action-gh-release`) with
  `body_path: RELEASE_NOTES.md`.
- [x] 4.4 Add the develop merge-back step, mirroring the Rust SDK's `release.yaml`
  (checkout `develop`, `git merge origin/master`, push).

## 5. Documentation

- [x] 5.1 Add `RELEASE.md` documenting the release process: how to trigger
  prepare-release, what the release branch PR should be reviewed for, how tagging and
  publishing happen automatically after merge.
- [x] 5.2 Update `README.md`: add CI and release status badges (mirroring the Rust SDK's
  README badge row), link to `RELEASE.md`, add a `go.dev` reference badge for the module.

## 6. Verification

- [x] 6.1 Confirm `golangci-lint run ./...` and `go test -race ./...` pass locally before
  opening the PR for this change (same gate the new CI workflow will enforce).
- [ ] 6.2 After merge, manually trigger `prepare-release.yaml` once to validate the full
  pipeline end-to-end for the SDK's first tagged release.
