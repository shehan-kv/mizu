# Configuration

Mizu is configured through environment variables.

Most configuration options have sensible defaults, allowing Mizu to run as a self-contained application without external services. External databases, Redis, S3-compatible object storage, and SMTP can be configured for larger or distributed deployments.

Environment variable values are trimmed of leading and trailing whitespace when they are read.

## General

| Variable             | Required    | Default | Description                                                                                   |
| -------------------- | ----------- | ------- | --------------------------------------------------------------------------------------------- |
| `PUBLIC_URL`         | Conditional | —       | Public URL of the Mizu installation. Required when an email driver other than `noop` is used. |
| `SERVER_ADDR`        | No          | `:8080` | Address and port on which the HTTP server listens.                                            |
| `EVENT_BUFFER_SIZE`  | No          | `1024`  | Buffer size used by Mizu's internal and external event buses. Must be a valid integer.        |
| `MAX_UPLOAD_SIZE_MB` | No          | `100`   | Maximum allowed upload size in megabytes. Must be a valid integer.                            |

### `PUBLIC_URL`

The public URL of the Mizu installation.

This URL is used when generating links included in emails. It is required when `EMAIL_DRIVER` is set to a driver other than `noop`.

Trailing `/` characters are removed automatically.

Example:

```env
PUBLIC_URL=https://mizu.example.com
```

### `SERVER_ADDR`

The address used by the HTTP server.

Default:

```env
SERVER_ADDR=:8080
```

The value is passed directly to Go's HTTP server.

Examples:

```env
SERVER_ADDR=:8080
SERVER_ADDR=127.0.0.1:8080
SERVER_ADDR=0.0.0.0:8080
```

### `EVENT_BUFFER_SIZE`

Controls the buffer size used by Mizu's event buses.

Default:

```env
EVENT_BUFFER_SIZE=1024
```

The value must be a valid integer.

### `MAX_UPLOAD_SIZE_MB`

Sets the maximum upload size in megabytes.

Default:

```env
MAX_UPLOAD_SIZE_MB=100
```

The value must be a valid integer.

---

## Database

| Variable               | Required    | Default   | Description                                  |
| ---------------------- | ----------- | --------- | -------------------------------------------- |
| `DB_DRIVER`            | No          | `sqlite`  | Database driver to use.                      |
| `DB_CONNECTION_STRING` | Conditional | `mizu.db` | Database connection string or database path. |

Supported database drivers:

- `sqlite`
- `postgres`

### SQLite

SQLite is the default database driver.

```env
DB_DRIVER=sqlite
DB_CONNECTION_STRING=mizu.db
```

If `DB_CONNECTION_STRING` is omitted, Mizu uses:

```text
mizu.db
```

The SQLite database is automatically migrated when Mizu starts.

### PostgreSQL

PostgreSQL can be selected with:

```env
DB_DRIVER=postgres
DB_CONNECTION_STRING=postgres://user:password@localhost:5432/mizu?sslmode=disable
```

`DB_CONNECTION_STRING` is required when using PostgreSQL.

Database migrations are automatically executed when Mizu starts.

---

## Sessions

| Variable                    | Required    | Default  | Description                                     |
| --------------------------- | ----------- | -------- | ----------------------------------------------- |
| `SESSION_DRIVER`            | No          | `memory` | Session storage driver.                         |
| `SESSION_CONNECTION_STRING` | Conditional | —        | Connection string for external session storage. |

Supported session drivers:

- `memory`
- `redis`

### In-memory sessions

The default session driver stores sessions in memory:

```env
SESSION_DRIVER=memory
```

No connection string is required.

Sessions stored in memory are local to the running Mizu instance. They are therefore not shared between multiple application instances.

### Redis sessions

Redis can be used for external session storage:

```env
SESSION_DRIVER=redis
SESSION_CONNECTION_STRING=redis://localhost:6379/0
```

The connection string must be a Redis URL accepted by the Redis client.

Redis connectivity is checked when Mizu starts.

---

## File Storage

| Variable                 | Required    | Default | Description                                  |
| ------------------------ | ----------- | ------- | -------------------------------------------- |
| `FILE_STORAGE_DRIVER`    | No          | `disk`  | File storage driver.                         |
| `FILE_CONNECTION_STRING` | Conditional | —       | Connection string for external file storage. |

Supported file storage drivers:

- `disk`
- `s3`

### Local disk storage

Local disk storage is the default:

```env
FILE_STORAGE_DRIVER=disk
```

No file storage connection string is required.

Mizu stores uploaded files in the local filesystem.

### S3-compatible storage

S3-compatible object storage can be used for external file storage:

```env
FILE_STORAGE_DRIVER=s3
```

The following S3 settings are required:

| Variable        | Required | Description                   |
| --------------- | -------- | ----------------------------- |
| `S3_ENDPOINT`   | No       | S3 service endpoint.          |
| `S3_REGION`     | Yes      | S3 region.                    |
| `S3_BUCKET`     | Yes      | Bucket used for file storage. |
| `S3_ACCESS_KEY` | Yes      | S3 access key.                |
| `S3_SECRET_KEY` | Yes      | S3 secret key.                |

`S3_ENDPOINT` is optional, allowing the storage implementation to use its default endpoint.

Example:

```env
FILE_STORAGE_DRIVER=s3
S3_ENDPOINT=https://s3.example.com
S3_REGION=us-east-1
S3_BUCKET=mizu
S3_ACCESS_KEY=your-access-key
S3_SECRET_KEY=your-secret-key
```

---

## Message Queue

| Variable                          | Required    | Default  | Description                                            |
| --------------------------------- | ----------- | -------- | ------------------------------------------------------ |
| `MESSAGE_QUEUE_DRIVER`            | No          | `memory` | External event bus driver.                             |
| `MESSAGE_QUEUE_CONNECTION_STRING` | Conditional | —        | Connection string for external message infrastructure. |

Supported drivers:

- `memory`
- `redis`

### In-memory message queue

The default configuration uses an in-memory event bus:

```env
MESSAGE_QUEUE_DRIVER=memory
```

No connection string is required.

### Redis message queue

Redis can be used as the external event bus:

```env
MESSAGE_QUEUE_DRIVER=redis
MESSAGE_QUEUE_CONNECTION_STRING=redis://localhost:6379/0
```

Redis connectivity is checked when Mizu starts.

The Redis-backed event bus is used to distribute supported integration events between Mizu application instances.

---

## Email

| Variable             | Required    | Default | Description                           |
| -------------------- | ----------- | ------- | ------------------------------------- |
| `EMAIL_DRIVER`       | No          | `noop`  | Email delivery driver.                |
| `EMAIL_TEMPLATE_DIR` | No          | —       | Directory containing email templates. |
| `SMTP_HOST`          | Conditional | —       | SMTP server hostname.                 |
| `SMTP_PORT`          | Conditional | —       | SMTP server port.                     |
| `SMTP_USERNAME`      | Conditional | —       | SMTP authentication username.         |
| `SMTP_PASSWORD`      | Conditional | —       | SMTP authentication password.         |
| `SMTP_FROM`          | Conditional | —       | Default sender address.               |

Supported email drivers:

- `noop`
- `smtp`

### No-op email driver

The default email driver is `noop`:

```env
EMAIL_DRIVER=noop
```

The no-op driver does not send emails.

This allows Mizu to run without an SMTP server.

### SMTP

To send email through an SMTP server:

```env
EMAIL_DRIVER=smtp
PUBLIC_URL=https://mizu.example.com

SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your-username
SMTP_PASSWORD=your-password
SMTP_FROM=noreply@example.com
```

All SMTP settings are required when the `smtp` driver is selected.

`PUBLIC_URL` is also required because email links generated by Mizu need the public installation URL.

### Email templates

Mizu includes built-in email templates that are used by default.

`EMAIL_TEMPLATE_DIR` can be used to provide a custom directory containing email templates instead.

```env
EMAIL_TEMPLATE_DIR=/path/to/templates
```

If `EMAIL_TEMPLATE_DIR` is omitted, Mizu uses its built-in email templates.

For instructions on creating and customizing email templates, see the [Email Template Customization Guide](email-templates.md).

---

## S3 Storage

The following variables configure S3-compatible object storage:

| Variable        | Required | Description                                               |
| --------------- | -------- | --------------------------------------------------------- |
| `S3_ENDPOINT`   | No       | S3 service endpoint.                                      |
| `S3_REGION`     | Yes      | S3 region.                                                |
| `S3_BUCKET`     | Yes      | Bucket used by Mizu.                                      |
| `S3_ACCESS_KEY` | Yes      | Access key used to authenticate with the storage service. |
| `S3_SECRET_KEY` | Yes      | Secret key used to authenticate with the storage service. |

These variables are only required when:

```env
FILE_STORAGE_DRIVER=s3
```

---

## Example Configurations

### Minimal configuration

The default configuration is designed to run without external services:

```env
DB_DRIVER=sqlite
DB_CONNECTION_STRING=mizu.db

SESSION_DRIVER=memory

FILE_STORAGE_DRIVER=disk

MESSAGE_QUEUE_DRIVER=memory

EMAIL_DRIVER=noop

SERVER_ADDR=:8080
EVENT_BUFFER_SIZE=1024
MAX_UPLOAD_SIZE_MB=100
```

All of these variables can be omitted because they correspond to the defaults.

A minimal installation can therefore run with no configuration beyond any values required by the deployment environment.

### PostgreSQL and Redis

A deployment using PostgreSQL and Redis could use:

```env
PUBLIC_URL=https://mizu.example.com
SERVER_ADDR=:8080

DB_DRIVER=postgres
DB_CONNECTION_STRING=postgres://user:password@localhost:5432/mizu?sslmode=disable

SESSION_DRIVER=redis
SESSION_CONNECTION_STRING=redis://localhost:6379/0

MESSAGE_QUEUE_DRIVER=redis
MESSAGE_QUEUE_CONNECTION_STRING=redis://localhost:6379/0

FILE_STORAGE_DRIVER=disk

EMAIL_DRIVER=noop

EVENT_BUFFER_SIZE=1024
MAX_UPLOAD_SIZE_MB=100
```

### Production-oriented configuration

A deployment using PostgreSQL, Redis, S3-compatible storage, and SMTP could use:

```env
PUBLIC_URL=https://mizu.example.com
SERVER_ADDR=:8080

DB_DRIVER=postgres
DB_CONNECTION_STRING=postgres://user:password@db:5432/mizu?sslmode=disable

SESSION_DRIVER=redis
SESSION_CONNECTION_STRING=redis://redis:6379/0

MESSAGE_QUEUE_DRIVER=redis
MESSAGE_QUEUE_CONNECTION_STRING=redis://redis:6379/0

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

EVENT_BUFFER_SIZE=1024
MAX_UPLOAD_SIZE_MB=100
```

## Configuration Requirements

The following table summarizes when configuration values are required:

| Component     | Default    | External option | Additional configuration            |
| ------------- | ---------- | --------------- | ----------------------------------- |
| Database      | SQLite     | PostgreSQL      | `DB_CONNECTION_STRING`              |
| Sessions      | In-memory  | Redis           | `SESSION_CONNECTION_STRING`         |
| File storage  | Local disk | S3              | S3 configuration                    |
| Message queue | In-memory  | Redis           | `MESSAGE_QUEUE_CONNECTION_STRING`   |
| Email         | No-op      | SMTP            | SMTP configuration and `PUBLIC_URL` |

## Production Security

Do not commit passwords, access keys, secret keys, SMTP credentials, database credentials, or other sensitive configuration to source control.

Use the secret-management facilities provided by your deployment environment when possible.

Restrict access to environment configuration and ensure that credentials have only the permissions required by Mizu.

## Production Deployment

For production deployment recommendations, including reverse proxy configuration and Nginx examples, see the [Deployment Guide](deployment.md).
