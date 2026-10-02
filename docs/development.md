# Development

This guide covers setting up a local Mizu development environment, running the application, testing changes, and building release binaries.

For the overall architecture and project structure, see the [Architecture Guide](architecture.md).

For configuration details, see the [Configuration Guide](configuration.md).

## Requirements

To develop Mizu, you will need:

- Go 1.26+
- Node.js
- npm
- GCC

Depending on the development task, you may also need:

- SQLite — required for local SQLite development
- Docker — required for Testcontainers-based infrastructure tests
- Redis — required when developing or testing Redis-backed functionality
- `aarch64-linux-gnu-gcc` — required for Linux ARM64 cross-compilation
- `x86_64-w64-mingw32-gcc` — required for Windows AMD64 cross-compilation

Not every dependency is required for every development task. For example, the cross-compilers are only needed when building the corresponding release targets.

## Getting Started

### Clone the Repository

Clone the repository and enter the project directory:

```bash
git clone https://github.com/shehan-kv/mizu.git
cd mizu
```

### Install Frontend Dependencies

Install the frontend dependencies:

```bash
npm --prefix web/ui install
```

The frontend is built with SvelteKit, Svelte, TypeScript, and Tailwind CSS.

### Development Configuration

Mizu can use its default configuration for basic local development.

For configuration options and environment variables, see the [Configuration Guide](configuration.md).

If you need Redis, PostgreSQL, SMTP, S3-compatible storage, or other external services during development, configure them according to that guide.

## Running Mizu

Mizu consists of a Go backend and a SvelteKit frontend.

### Running the Backend

Build the Go backend with:

```bash
make build-go
```

The resulting executable is placed in the `bin` directory.

Run the generated executable to start the backend:

```bash
./bin/mizu
```

The backend listens on the configured server address. By default, this is:

```text
http://localhost:8080
```

### Running the Frontend

Start the SvelteKit development server with:

```bash
make ui-dev
```

The development server provides hot module replacement, allowing frontend changes to be reflected without rebuilding the production frontend.

When working on the frontend, keep the Go backend running separately.

### Building the Complete Application

To build Mizu with the frontend bundled into the Go binary:

```bash
make build
```

The resulting executable is placed in the `bin` directory.

This produces a self-contained Mizu binary containing the built frontend.

## Make Commands

Mizu provides a Makefile for common development and build operations.

| Command          | Description                                    |
| ---------------- | ---------------------------------------------- |
| `make build`     | Build Mizu for the current platform            |
| `make build-go`  | Build the Go backend                           |
| `make ui-dev`    | Start the frontend development server          |
| `make fmt`       | Format Go source code                          |
| `make vet`       | Run `go vet`                                   |
| `make test`      | Run the test suite                             |
| `make build-all` | Build the UI and all supported release targets |
| `make clean`     | Remove generated binaries                      |

Run `make` without a target to see the available commands if the Makefile provides a default help target.

## Testing

Mizu uses several levels of automated testing.

### Domain Tests

Domain tests verify business rules and domain behavior without depending on infrastructure.

These tests should remain independent of databases, HTTP handlers, external services, and other infrastructure concerns.

### Application Tests

Application-layer tests verify application services, authorization, repository interactions, domain services, and application workflows.

Application tests use test doubles where appropriate so that application behavior can be tested independently of infrastructure implementations.

### Infrastructure Tests

Infrastructure tests verify integrations with external systems such as databases and other services.

Infrastructure tests use **Testcontainers** where external dependencies are required.

### Running the Test Suite

Run the complete test suite with:

```bash
make test
```

You can also use the standard Go tooling when working on a specific package or test:

```bash
go test ./path/to/package
```

Run a specific test with:

```bash
go test ./path/to/package -run TestName
```

Run tests with the race detector when appropriate:

```bash
go test -race ./...
```

## Code Formatting and Static Analysis

Format Go source code with:

```bash
make fmt
```

Run `go vet` with:

```bash
make vet
```

Run both formatting and static analysis before submitting changes.

## Building Release Binaries

Mizu can build all supported release targets from a development machine with the required cross-compilers installed.

Build all release targets with:

```bash
make build-all VERSION=v0.1.0
```

This produces:

```text
bin/mizu-v0.1.0-linux-amd64
bin/mizu-v0.1.0-linux-arm64
bin/mizu-v0.1.0-windows-amd64.exe
```

The Linux ARM64 build requires:

```text
aarch64-linux-gnu-gcc
```

The Windows AMD64 build requires:

```text
x86_64-w64-mingw32-gcc
```

The release build also builds the frontend before embedding it into the application binaries.

For release packaging and deployment information, see the [Deployment Guide](deployment.md).

## Project Structure

Mizu follows a Clean Architecture and domain-driven design approach.

The project is broadly separated into:

- **Domain** — business entities, value objects, domain services, and domain events
- **Application** — application services, authorization, workflows, and ports
- **Infrastructure** — database, storage, email, sessions, messaging, and other external integrations
- **Presentation** — request handling and API delivery
- **Web** — SvelteKit application and user interface

Business rules should remain independent of infrastructure concerns.

For a detailed explanation of the architecture and dependency boundaries, see the [Architecture Guide](architecture.md).

## Development Workflow

A typical development workflow is:

1. Create a branch for the change.
2. Make the required changes.
3. Format the Go code with `make fmt`.
4. Run static analysis with `make vet`.
5. Run the test suite with `make test`.
6. Build the application with `make build`.
7. Verify the change locally.
8. Commit the changes using a conventional commit message.
9. Open a pull request.

Keep changes focused and avoid mixing unrelated refactoring with feature or bug-fix changes.

## Related Documentation

- [Configuration](configuration.md) — configuration options and environment variables
- [Architecture](architecture.md) — application architecture and design
- [Testing](testing.md) — testing strategy and conventions
- [Deployment](deployment.md) — production deployment and release binaries
- [Contributing](contributing.md) — contribution guidelines
- [CI/CD](ci-cd.md) — continuous integration and release automation
