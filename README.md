# Mizu

> A self-hosted project management and client communication platform for freelancers and small agencies, combining project management with organized, project-aware messaging.

**Mizu** is a lightweight, self-hosted platform designed to help freelancers and small agencies manage projects and communicate with clients in one place.

Mizu connects conversations directly to projects. Each project can have its own communication channel, allowing clients, administrators, and team members to discuss work while keeping conversations and project activity organized.

## Quick Start

### Default Administrator Account

On the first startup, Mizu creates a default administrator account:

| Field    | Value        |
| -------- | ------------ |
| Email    | `admin@mizu` |
| Password | `admin@mizu` |

> [!WARNING]
> **Change the default administrator password immediately after signing in.**
>
> These credentials are intended only for initial setup and should not be used in a production environment.

### Option 1: Run a Prebuilt Binary

The easiest way to run Mizu is to download a prebuilt binary from the [releases page](https://github.com/shehan-kv/mizu/releases).

Prebuilt binaries are currently available for:

- Linux AMD64
- Linux ARM64
- Windows AMD64

Mizu is self-contained by default and does not require an external database, Redis instance, object storage service, or other infrastructure for a basic installation.

After downloading the appropriate binary, run it:

**Linux AMD64**

```bash
./mizu-v0.1.0-linux-amd64
```

**Linux ARM64**

```bash
./mizu-v0.1.0-linux-arm64
```

**Windows AMD64**

```powershell
.\mizu-v0.1.0-windows-amd64.exe
```

Mizu uses sensible defaults for getting started. For production deployments, configure the application through environment variables.

See [Configuration](#configuration) for available options.

### Option 2: Build from Source

If you prefer to build Mizu yourself, install the required development dependencies and clone the repository:

```bash
git clone [Repository URL]
cd mizu
```

Build the application:

```bash
make build
```

The resulting binary will be placed in:

```text
bin/mizu
```

You can then run the binary directly.

For release builds targeting all supported platforms, see [Building Release Binaries](#building-release-binaries).

### Option 3: Development Setup

For development, install:

- Go 1.26+
- Node.js
- npm
- GCC
- SQLite
- Docker
- Redis

Additional cross-compilers are required only when building release binaries for other platforms:

- `aarch64-linux-gnu-gcc`
- `x86_64-w64-mingw32-gcc`

Start the frontend development server with:

```bash
make ui-dev
```

Build the Go backend with:

```bash
make build-go
```

The backend binary will be generated in the `bin` directory.

Run the generated binary to start the backend.

Run the test suite with:

```bash
make test
```

For the complete development workflow, testing strategy, project structure, and contribution process, see the project documentation.

## Why Mizu?

Mizu is designed to centralize project management and client communication within a single self-hosted application.

Projects can have dedicated communication channels where clients, administrators, and team members can discuss project-related work. This allows communication to remain separated across multiple projects while keeping relevant project activity accessible within the same context.

Project events can also be reflected in their associated channels. For example, when an invoice is issued for a project, Mizu can publish an event in the project's channel, providing participants with relevant project updates alongside their conversations.

This approach provides a structured workspace for managing multiple client projects while keeping communication and project activity connected.

## Features

### Project Management

Manage clients and projects throughout their lifecycle, including tasks, contracts, quotes, invoices, payments, files, and project activity.

### Project-Centered Communication

Create dedicated channels for projects and keep client and team conversations separated across multiple projects. Channels can include administrators, staff, and clients based on project access.

### Project-Aware Activity

Project activity can be surfaced directly within associated communication channels. Events such as issued invoices can appear as messages in the relevant project channel, keeping important updates alongside the conversation.

### Client Portal

Provide clients with access to their projects, communication, and relevant project information through a dedicated client-facing interface without exposing internal administrative functionality.

### Self-Hosted and Lightweight

Mizu is distributed as a single executable with the web frontend bundled into the binary. The default configuration uses SQLite and local storage, allowing Mizu to run without external services.

For larger deployments, Mizu supports PostgreSQL, Redis, S3-compatible object storage, and SMTP.

### Horizontally Scalable

External infrastructure can be used to support deployments across multiple application instances, allowing Mizu to scale beyond a single-server installation.

<!-- ## Screenshots

Screenshots will be added here. -->

<!--
Suggested screenshots:

- Dashboard
- Project view
- Project channel
- Task management
- Invoice
- Client portal
- Login
-->

## Architecture

Mizu is built using **Go** and **SvelteKit**, following a DDD-inspired Clean Architecture approach.

The backend is intentionally lightweight:

- Go 1.26
- Go standard library HTTP server and router
- Handwritten SQL queries
- No ORM
- SQLite and PostgreSQL support
- Pluggable infrastructure components
- Domain-driven application structure

The frontend is built with:

- SvelteKit
- Svelte 5
- TypeScript
- Tailwind CSS
- bits-ui
- shadcn-svelte

The application is designed so that infrastructure components can be replaced without changing the core application behavior. For example, Mizu can use SQLite for a simple installation and PostgreSQL when a larger deployment requires it.

## Lightweight and Self-Contained

One of Mizu's goals is to make self-hosted project management practical even for users with limited infrastructure.

The basic installation does not require a database server, Redis instance, object storage service, or other external infrastructure.

A typical installation can simply be:

```text
Download binary
      ↓
Run Mizu
      ↓
Use SQLite and local storage
```

The SvelteKit frontend is bundled into the Go binary, so the application can be deployed as a **single executable**.

This makes Mizu suitable for small VPS instances and low-cost infrastructure. A basic installation can run on a machine with limited resources, including a 1 GB RAM virtual machine.

For larger deployments, Mizu can use external services and scale horizontally when additional capacity is required.

## Supported Platforms

Prebuilt release binaries are currently provided for:

| Platform | Architecture |
| -------- | ------------ |
| Linux    | amd64        |
| Linux    | arm64        |
| Windows  | amd64        |

Mizu currently supports Linux and Windows.

## Configuration

Mizu is configured through environment variables.

The default configuration is designed to allow Mizu to run without
external services. Production deployments can configure databases,
sessions, storage, email, and other infrastructure through environment
variables.

See the [Configuration Reference](docs/configuration.md) for the
complete list of configuration options.

## Deployment

Mizu is designed to be deployed as a standalone application.

A typical production deployment can consist of:

```text
                  ┌──────────────┐
Internet ────────►│ Reverse Proxy│
                  └──────┬───────┘
                         │
                         ▼
                  ┌──────────────┐
                  │     Mizu     │
                  │   Go binary  │
                  └──────┬───────┘
                         │
                 ┌───────┴────────┐
                 ▼                ▼
              SQLite          External
                              Services
```

For production deployments, a reverse proxy is recommended for handling concerns such as HTTPS and public-facing HTTP traffic.

Mizu can be deployed on a VPS or other Linux server. More advanced deployments can use external PostgreSQL, Redis, S3-compatible storage, and SMTP infrastructure.

### Docker

Docker support is not currently provided.

## Project Status

Mizu is currently in **Beta**.

The application is largely production-ready, while development continues to expand and improve the existing feature set.

## Contributing

Contributions are welcome.

Mizu is intended to become an open-source project, and contributions can include:

- Bug fixes
- New features
- Performance improvements
- Tests
- Documentation
- UI/UX improvements
- Infrastructure improvements

Contribution guidelines will be added as the project moves toward its public release.

## License

Mizu is released under the **MIT License**.

See the [LICENSE](LICENSE) file for the complete license text.

## Links

- **GitHub:** https://github.com/shehan-kv/mizu
- **Documentation:** https://github.com/shehan-kv/mizu/tree/main/docs
- **Releases:** https://github.com/shehan-kv/mizu/releases
- **Issues:** https://github.com/shehan-kv/mizu/issues
- **Discussions:** https://github.com/shehan-kv/mizu/discussions

---

## About

Mizu combines project management, client collaboration, and project-centered communication in a single self-hosted application.

It is designed to remain lightweight and straightforward to deploy on modest infrastructure while providing the flexibility to scale through external services and multiple application instances as requirements grow.
