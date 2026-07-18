# Release process

## Summary

1. Trigger the [**Prepare Release**](https://github.com/pilgrimagesoftware/dtrpg-sdk.go/actions/workflows/prepare-release.yaml) workflow.
1. Merge PR to `master`.
1. Merge `master` into `develop`.

## Breakdown

Go modules have no `Cargo.toml`-equivalent version field — a module's version comes
entirely from its git tag. Instead of writing the changelog by hand on the release
branch, a `prepare-release` workflow does it for you:

```sh
git checkout develop
git pull
```

Trigger the [**Prepare Release**](https://github.com/pilgrimagesoftware/dtrpg-sdk.go/actions/workflows/prepare-release.yaml) workflow (`workflow_dispatch` in the Actions tab). It:

1. Runs `git-cliff --bump` against `develop` to determine the next version (e.g. `0.3.0`)
   from commits since the last tag.
2. Prepends the generated changelog section to `CHANGELOG.md`.
3. Opens a PR from an auto-created `release/0.3.0` branch into `master`.

You review the PR (catch anything that shouldn't ship, fix as needed) — the
[**PR Validation**](https://github.com/pilgrimagesoftware/dtrpg-sdk.go/actions/workflows/pr.yaml)
workflow runs against it — merge into `master`, which triggers tagging via the
[**Tag Release**](https://github.com/pilgrimagesoftware/dtrpg-sdk.go/actions/workflows/tag-release.yaml)
workflow.

The tag push triggers the
[**Release**](https://github.com/pilgrimagesoftware/dtrpg-sdk.go/actions/workflows/release.yaml)
workflow, which re-runs the test suite against the tagged commit, requests the new version
from the Go module proxy so `go get` and pkg.go.dev see it immediately, generates the
changelog scoped to that tag, attaches it to a GitHub Release, and merges `master` back
into `develop`.

## Triggering the run

```sh
gh workflow run prepare-release.yaml
```

Or trigger it directly from the [Prepare Release workflow page](https://github.com/pilgrimagesoftware/dtrpg-sdk.go/actions/workflows/prepare-release.yaml).
