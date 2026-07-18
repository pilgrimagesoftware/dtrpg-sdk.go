# Contributing code

## 📢 Conventional Commits 📢

To enhance our development workflow and enable automated changelog generation, the `dtrpg-sdk.go` project has
adopted the [Conventional Commits standard](https://www.conventionalcommits.org/en/v1.0.0/) for all commit messages:

```
<type>(<scope>): <description>
```

Going forward, all commits to this repository should adhere to the Conventional Commits standard. The changelog and
version bump on release are generated directly from commit history via `git-cliff` (see [RELEASE.md](RELEASE.md)),
so a commit that doesn't follow the convention won't be grouped correctly in the changelog it produces.

## Development

See the [README](README.md#development) for the local build/lint/test commands, and [RELEASE.md](RELEASE.md) for how
releases are cut.

## Pull Requests

- Branch from `develop`, following [`docs/git-flow.md`](https://github.com/pilgrimagesoftware/dtrpg/blob/master/docs/git-flow.md)'s branch model (`feature/*`, `fix/*`).
- Open pull requests against `develop`, not `master`.
- `gofmt -l .`, `go vet ./...`, `golangci-lint run ./...`, and `go test -race ./...` must all pass — CI runs these
  automatically on every PR.
