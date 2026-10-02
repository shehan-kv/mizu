# Deployment

This guide covers deploying Mizu as a single application instance and deploying multiple Mizu instances behind a reverse proxy.

For configuration options, see the [Configuration Reference](configuration.md).

## Deployment Models

Mizu can be deployed in two general configurations.

### Single Instance

A single Mizu instance is suitable for small deployments and personal or small-business installations.

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

The default configuration uses:

- SQLite
- In-memory sessions
- Local disk file storage
- In-memory event bus
- No-op email driver

No external infrastructure is required.

### Multiple Instances

Mizu can also be deployed as multiple application instances when external infrastructure is used for shared state and communication.

```text
                 ┌──────────────────┐
Internet ───────►│ Load Balancer /  │
                 │ Reverse Proxy    │
                 └────────┬─────────┘
                          │
                 ┌────────┴────────┐
                 │                 │
            ┌────▼────┐       ┌────▼────┐
            │ Server 1│  ...  │ Server N│
            │  Mizu   │       │  Mizu   │
            └────┬────┘       └────┬────┘
                 │                 │
                 └────────┬────────┘
                          │
              ┌───────────┼───────────┐
              ▼           ▼           ▼
          PostgreSQL    Redis     S3 Storage
```

For multiple instances, configure:

- PostgreSQL for the shared database
- Redis for shared sessions
- Redis for the external event bus
- S3-compatible object storage for shared files

The application instances should use the same configuration for these shared services.

---

## Requirements

### Single Instance

A single-instance deployment requires:

- A supported Linux or Windows system
- A Mizu release binary or a Go build from source
- A writable filesystem
- A database location
- A network interface and port for the HTTP server

A reverse proxy is recommended for production deployments.

### Multiple Instances

A multi-instance deployment additionally requires:

- PostgreSQL
- Redis
- S3-compatible object storage
- A reverse proxy or load balancer
- A shared network between the Mizu instances and external services

---

## Single-Instance Deployment

The simplest production deployment is a single Mizu binary with SQLite and local file storage.

### 1. Download Mizu

Download the appropriate release binary from the [Mizu releases page](https://github.com/shehan-kv/mizu/releases).

For example, on Linux AMD64:

```bash
chmod +x mizu-v0.1.0-linux-amd64
```

You can rename the binary if desired:

```bash
mv mizu-v0.1.0-linux-amd64 mizu
```

### 2. Configure Mizu

The default configuration is sufficient for a basic installation.

For example:

```env
SERVER_ADDR=:8080
DB_DRIVER=sqlite
DB_CONNECTION_STRING=mizu.db
SESSION_DRIVER=memory
FILE_STORAGE_DRIVER=disk
MESSAGE_QUEUE_DRIVER=memory
EMAIL_DRIVER=noop
```

See the [Configuration Reference](configuration.md) for all available options.

### 3. Start Mizu

Run:

```bash
./mizu
```

Mizu will:

1. Initialize the configured database.
2. Run database migrations.
3. Initialize storage and event infrastructure.
4. Create the default administrator if one does not already exist.
5. Start the HTTP server.

By default, the server listens on:

```text
http://localhost:8080
```

### 4. Production Reverse Proxy

For production deployments, placing a reverse proxy in front of Mizu is recommended.

A reverse proxy can provide:

- HTTPS/TLS termination
- Domain-based routing
- Connection handling
- Access logging
- A stable public endpoint

The reverse proxy should forward requests to the Mizu HTTP server.

---

## Nginx

The following is a basic Nginx configuration for a single Mizu instance.

```nginx
server {
    # Listen for HTTP traffic on the standard port.
    # In production, HTTPS should normally be terminated at the reverse proxy.
    listen 80;

    # Replace this with the hostname used to access your Mizu installation.
    server_name mizu.example.com;

    location / {
        # Forward all requests to the Mizu application.
        proxy_pass http://127.0.0.1:8080;

        # Required for persistent connections and useful for SSE.
        proxy_http_version 1.1;

        # Preserve the original request information so Mizu can
        # determine the requested host, client IP, and protocol.
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Mizu uses Server-Sent Events (SSE) for real-time updates.
    location /api/v1/events {
        # Forward the SSE connection to the Mizu application.
        proxy_pass http://127.0.0.1:8080;

        # Keep the connection using HTTP/1.1, which is required
        # for long-lived SSE connections.
        proxy_http_version 1.1;

        # Preserve the original request information.
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Prevent Nginx from buffering SSE events.
        # Without this, real-time events may be delayed until
        # Nginx's buffer is flushed.
        proxy_buffering off;
    }
}
```

The `/api/v1/events` endpoint uses Server-Sent Events (SSE). Proxy buffering should therefore be disabled for this endpoint.

Replace `mizu.example.com` with the domain used for the Mizu installation.

### HTTPS

For a public production installation, HTTPS should be configured on the reverse proxy.

The Mizu application can continue listening on its local HTTP address while Nginx handles TLS.

Set `PUBLIC_URL` to the externally accessible HTTPS URL:

```env
PUBLIC_URL=https://mizu.example.com
```

---

## Running Mizu as a systemd Service

On Linux, Mizu can be managed using systemd.

Create a service file such as:

```ini
[Unit]
Description=Mizu
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/mizu
ExecStart=/opt/mizu/mizu
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Save it as:

```text
/etc/systemd/system/mizu.service
```

Then enable and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable mizu
sudo systemctl start mizu
```

Check its status with:

```bash
sudo systemctl status mizu
```

Logs can be viewed with:

```bash
sudo journalctl -u mizu
```

For production deployments, environment variables should preferably be supplied through the service configuration or an appropriate secret-management mechanism rather than committed to the repository.

---

# Multiple-Instance Deployment

Multiple Mizu instances can be deployed behind a reverse proxy or load balancer.

```text
                 ┌──────────────────┐
Internet ───────►│ Load Balancer /  │
                 │ Reverse Proxy    │
                 └────────┬─────────┘
                          │
                 ┌────────┴────────┐
                 │                 │
            ┌────▼────┐       ┌────▼────┐
            │ Server 1│  ...  │ Server N│
            │  Mizu   │       │  Mizu   │
            └────┬────┘       └────┬────┘
                 │                 │
                 └────────┬────────┘
                          │
              ┌───────────┼───────────┐
              ▼           ▼           ▼
          PostgreSQL    Redis     S3 Storage
```

The Mizu instances are stateless with respect to the infrastructure that must be shared between instances when the appropriate external drivers are configured.

## Shared Services

### PostgreSQL

Use PostgreSQL as the shared database:

```env
DB_DRIVER=postgres
DB_CONNECTION_STRING=postgres://user:password@db:5432/mizu?sslmode=disable
```

All Mizu instances must use the same PostgreSQL database.

Mizu runs database migrations during startup.

Only the Mizu application instances should need access to the application database.

### Redis Sessions

Configure Redis for shared session storage:

```env
SESSION_DRIVER=redis
SESSION_CONNECTION_STRING=redis://redis:6379/0
```

Using Redis allows authenticated sessions to be shared between Mizu instances.

The instances must use the same Redis session store.

### Redis Event Bus

Configure the external event bus to use Redis:

```env
MESSAGE_QUEUE_DRIVER=redis
MESSAGE_QUEUE_CONNECTION_STRING=redis://redis:6379/0
```

This allows supported integration events to be distributed between application instances.

The instances must use the same Redis event bus.

If the same Redis server is used for sessions and the event bus, separate Redis databases or appropriate isolation should be considered.

### S3-Compatible Storage

Local disk storage should not be used for files when requests can be handled by different Mizu instances.

Configure shared object storage:

```env
FILE_STORAGE_DRIVER=s3

S3_ENDPOINT=https://s3.example.com
S3_REGION=us-east-1
S3_BUCKET=mizu
S3_ACCESS_KEY=your-access-key
S3_SECRET_KEY=your-secret-key
```

All Mizu instances must have access to the same bucket.

### Email

Configure an SMTP server if Mizu needs to send email:

```env
EMAIL_DRIVER=smtp

SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your-username
SMTP_PASSWORD=your-password
SMTP_FROM=noreply@example.com
```

The public installation URL must also be configured:

```env
PUBLIC_URL=https://mizu.example.com
```

---

## Multi-Instance Configuration

The relevant configuration for each application instance should be equivalent to:

```env
PUBLIC_URL=https://mizu.example.com
SERVER_ADDR=:8080

DB_DRIVER=postgres
DB_CONNECTION_STRING=postgres://user:password@db:5432/mizu?sslmode=disable

SESSION_DRIVER=redis
SESSION_CONNECTION_STRING=redis://redis:6379/0

MESSAGE_QUEUE_DRIVER=redis
MESSAGE_QUEUE_CONNECTION_STRING=redis://redis:6379/1

FILE_STORAGE_DRIVER=s3
S3_ENDPOINT=https://s3.example.com
S3_REGION=us-east-1
S3_BUCKET=mizu
S3_ACCESS_KEY=your-access-key
S3_SECRET_KEY=your-secret-key

EMAIL_DRIVER=smtp
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your-username
SMTP_PASSWORD=your-password
SMTP_FROM=noreply@example.com
```

The application-specific configuration should be kept consistent across all instances.

---

## Load Balancing

The reverse proxy or load balancer distributes incoming HTTP requests between the Mizu instances.

A basic Nginx upstream configuration can look like:

```nginx
# Define the Mizu instances that will receive traffic.
# Each instance runs on a separate server and uses the same
# shared PostgreSQL, Redis, and S3-compatible storage.
upstream mizu {
    server 10.0.0.11:8080;
    server 10.0.0.12:8080;
    server 10.0.0.13:8080;
}

server {
    # Listen for HTTP traffic on the standard port.
    # In production, HTTPS should normally be terminated at the reverse proxy.
    listen 80;

    # Replace this with the hostname used to access your Mizu installation.
    server_name mizu.example.com;

    location / {
        # Distribute requests across the Mizu instances defined above.
        proxy_pass http://mizu;

        # Use HTTP/1.1 for persistent connections.
        proxy_http_version 1.1;

        # Preserve the original request information so Mizu can
        # determine the requested host, client IP, and protocol.
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Mizu uses Server-Sent Events (SSE) for real-time updates.
    location /api/v1/events {
        # Forward the SSE connection to one of the Mizu instances.
        proxy_pass http://mizu;

        # Keep the connection using HTTP/1.1 for the long-lived SSE connection.
        proxy_http_version 1.1;

        # Preserve the original request information.
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Prevent Nginx from buffering SSE events so that real-time
        # updates are forwarded to clients immediately.
        proxy_buffering off;
    }
}
```

Each Mizu instance would listen on a different local port:

```text
Mizu instance 1 → 10.0.0.11:8080
Mizu instance 2 → 10.0.0.12:8080
Mizu instance 3 → 10.0.0.13:8080
```

For larger deployments, the instances can instead run on separate machines or containers.

---

## Deployment Considerations

When running multiple Mizu instances, application state that must be shared between instances cannot be stored in the individual application process or local filesystem.

> [!WARNING]
> **Do not use in-memory sessions across multiple instances.**
>
> The default:
>
> ```env
> SESSION_DRIVER=memory
> ```
>
> stores sessions inside the individual Mizu process. A session created on one instance is not available to another instance.
>
> Use Redis for multi-instance deployments:
>
> ```env
> SESSION_DRIVER=redis
> ```

> [!WARNING]
> **Do not use local file storage across multiple instances.**
>
> The default:
>
> ```env
> FILE_STORAGE_DRIVER=disk
> ```
>
> stores uploaded files on the local filesystem. Files written to one instance are not automatically available to another instance.
>
> Use shared S3-compatible storage for multi-instance deployments.

> [!WARNING]
> **Use an external event bus across multiple instances.**
>
> The default in-memory event bus operates within a single application process. Events published by one instance are not automatically distributed to the other instances.
>
> Configure Redis for the external event bus:
>
> ```env
> MESSAGE_QUEUE_DRIVER=redis
> ```
>
> This allows supported integration events to be distributed between Mizu instances.

### Database

All Mizu instances must connect to the same application database.

Do not configure each instance with a separate SQLite database in a multi-instance deployment. Use a shared PostgreSQL database instead.

---

## Backups

A production deployment should have a backup strategy for persistent data.

At minimum, consider backups for:

- PostgreSQL data
- SQLite database files for single-instance deployments
- Uploaded files when using local disk storage
- S3-compatible object storage when applicable
- Configuration and deployment information required to restore the installation

Test backups periodically by performing a restoration rather than relying only on successful backup jobs.

---

## Updating Mizu

Before updating a production installation:

1. Back up persistent data.
2. Stop or remove the existing Mizu instance.
3. Replace the Mizu binary with the new release.
4. Start Mizu again.
5. Check the application logs.
6. Verify that the application and its external services are operating correctly.

Mizu runs database migrations during startup when required.

For multiple instances, update instances in a controlled manner and ensure that all instances ultimately run the same Mizu version.

---

For a single-instance deployment, SQLite, local disk storage, and in-memory infrastructure can be used when their limitations are acceptable.
