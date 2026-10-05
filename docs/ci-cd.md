# CI/CD

Mizu uses GitHub Actions for continuous integration and automated release builds.

The CI/CD workflows are defined in `.github/workflows/`.

## Continuous Integration

Continuous integration runs automatically when changes are pushed to the repository or when a pull request is opened or updated.

The CI workflow verifies the project by running the same checks expected before changes are merged.

### Go Formatting

The workflow checks that Go source files are formatted correctly.

This helps keep formatting consistent across the project.

### Static Analysis

The workflow runs:

```bash
go vet ./...
```

This performs static analysis and reports common correctness issues in Go code.

### Automated Tests

The workflow runs the project's automated test suite:

```bash
make test
```

This includes the project's domain, application, and infrastructure tests.

Infrastructure tests that require external services use Testcontainers where applicable.

## Release Builds

Mizu releases are built automatically by GitHub Actions when a version tag is pushed to the repository.

For example:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The release workflow uses the version tag to build the release binaries and publish them as GitHub Release assets.

## Release Targets

The release workflow builds the currently supported release targets:

| Target        | Output                          |
| ------------- | ------------------------------- |
| Linux AMD64   | `mizu-vX.Y.Z-linux-amd64`       |
| Linux ARM64   | `mizu-vX.Y.Z-linux-arm64`       |
| Windows AMD64 | `mizu-vX.Y.Z-windows-amd64.exe` |

The frontend is built as part of the release process and bundled into the Go application.

The resulting binaries are self-contained Mizu application binaries.

## Version Tags

Mizu releases use semantic version tags in the following format:

```text
vMAJOR.MINOR.PATCH
```

For example:

```text
v0.1.0
v0.2.0
v1.0.0
```

A release should be created by pushing the corresponding version tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The tag becomes the release version used by the release workflow.

## Release Process

A typical release process is:

1. Complete and verify the changes locally.
2. Run formatting and static analysis.
3. Run the test suite.
4. Build the application locally if necessary.
5. Create a version tag.
6. Push the tag to GitHub.
7. GitHub Actions builds the supported release targets.
8. The workflow publishes the binaries as GitHub Release assets.

Before creating a release, verify that the working tree contains the intended changes and that the version tag points to the intended commit.

## Local Release Builds

Release binaries can also be built locally when the required cross-compilers are installed.

Run:

```bash
make build-all VERSION=v0.1.0
```

See the [Development Guide](development.md) for information about the required toolchains and local build process.

## Deployment

Building a release does not deploy Mizu to a production server.

After a release is published, the resulting binary can be deployed according to the [Deployment Guide](deployment.md).

Production deployments may use a single Mizu instance or multiple instances behind a load balancer.

## Related Documentation

- [Development](development.md) — local development, testing, and builds
- [Deployment](deployment.md) — production deployment and release binaries
