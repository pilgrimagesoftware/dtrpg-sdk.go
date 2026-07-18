## ADDED Requirements

### Requirement: Release preparation is manually triggered from develop
The repository SHALL provide a manually triggered (`workflow_dispatch`) workflow that
computes the next semantic version from `develop`'s Conventional Commits history via
`git-cliff --bump`, opens a `release/<version>` branch with the updated `CHANGELOG.md`,
and opens a pull request from that branch into `master`.

#### Scenario: Triggering release preparation
- **WHEN** a maintainer manually triggers the prepare-release workflow
- **THEN** it computes the next version from `develop`'s commit history, creates a
  `release/<version>` branch with the changelog updated, and opens a PR into `master`

### Requirement: Merging a release PR tags master
The repository SHALL tag the merge commit `v<version>` and push that tag when a pull
request from a `release/*` branch is merged into `master`.

#### Scenario: Release PR merged
- **WHEN** a pull request from a `release/<version>` branch is merged into `master`
- **THEN** the merge commit is tagged `v<version>` and the tag is pushed to the remote

#### Scenario: Non-release PR merged into master
- **WHEN** a pull request merged into `master` is not from a `release/*` branch
- **THEN** no tag is created

### Requirement: A pushed version tag publishes a GitHub Release
The repository SHALL run a release workflow on every `v*` tag push that re-runs the full
test suite against the tagged commit, and creates a GitHub Release for that tag with the
`git-cliff`-generated changelog section for that version as the release body.

#### Scenario: Tag push triggers release
- **WHEN** a tag matching `v*` is pushed
- **THEN** the release workflow runs the test suite against the tagged commit and, on
  success, creates a GitHub Release for that tag with the changelog section for that
  version

#### Scenario: Tests fail on the tagged commit
- **WHEN** the test suite fails against the tagged commit
- **THEN** no GitHub Release is created

### Requirement: A published tag is immediately fetchable via the Go module proxy
The release workflow SHALL request the tagged module version from the configured Go
module proxy (`GOPROXY=proxy.golang.org go list -m <module>@v<version>`) so the version is
indexed and available to `go get` and `pkg.go.dev` consumers without waiting on an organic
first fetch. This step SHALL NOT fail the workflow if the proxy request itself fails —
the module remains fetchable via direct VCS regardless.

#### Scenario: Successful proxy indexing request
- **WHEN** the release workflow requests the newly tagged version from
  `proxy.golang.org`
- **THEN** the version becomes indexed and visible on `pkg.go.dev` without requiring a
  separate consumer request

#### Scenario: Proxy request fails
- **WHEN** the proxy indexing request fails (e.g. proxy unavailable or rate limited)
- **THEN** the release workflow still completes successfully and the GitHub Release is
  still created; the module remains fetchable via direct VCS fallback

### Requirement: Master is merged back into develop after release
The release workflow SHALL merge `master` back into `develop` after a successful release,
so `develop` includes the version bump and changelog update produced by the release.

#### Scenario: Successful release merges back
- **WHEN** the release workflow completes a successful GitHub Release
- **THEN** `master` is merged into `develop` and the merge is pushed
