# Compose Init

`compose-init` is a containerized tool designed to prepare your Docker Compose environment **before** your main services start. It ensures a consistent, "just works" developer experience by handling common initialization tasks automatically.

## Features

- **Smart Permissions**: Fix ownership of bind-mounted volumes to match the host user. Supports recursive chown and separate file/directory modes.
- **Templating**: Render configuration files using Go templates with environment variables.
- **Validation**: Fail fast if required environment variables are missing.
- **SSL Generation**: Automatically generate self-signed development certificates with expiration checks and forced regeneration.
- **Resource Fetching**: Download assets, verify checksums, and extract archives (zip/tar).

## Usage

Add `compose-init` to your `compose.yaml`. You can configure extensions at the **top-level** (global) or within **specific services**.

### Basic Setup

```yaml
services:
  init:
    image: ghcr.io/fransvandermeer/compose-init:latest
    volumes:
      - .:/project # Mount your project root to /project
    environment:
      - DATABASE_URL=postgres://user:pass@db:5432/db

  web:
    image: nginx
    depends_on:
      init:
        condition: service_completed_successfully
```

## Configuration Reference

### 1. Permissions (`x-chown`)

Fix ownership and permissions of files and directories. Useful for bind mounts that default to `root` ownership.

**Fields:**
- `path`: Target path (relative to container, usually inside `/project`).
- `owner`: Target owner. Use `"host"` to auto-detect the UID/GID of your `compose.yaml`. Or specify `"1000:1000"`.
- `mode`: Octal permission mode (fallback). E.g., `"0755"`.
- `file_mode`: Specific mode for files. E.g., `"0644"`.
- `dir_mode`: Specific mode for directories. E.g., `"0755"`.
- `recursive`: Apply recursively (default `false`). **Note**: Symlinks are ignored during permission changes (`chmod`), but ownership changes (`chown`) are applied to the symlink itself.

**Example:**

```yaml
x-chown:
  - path: ./data
    owner: "host"
    file_mode: "0644" # Read/Write for owner, Read for others
    dir_mode: "0755"  # Execute/Traverse for everyone
    recursive: true
```

### 2. Resource Fetching (`x-fetch`)

Download files or archives from the internet.

**Fields:**
- `url`: Source URL.
- `dest`: Destination path.
- `sha256`: Optional checksum verification.
- `force`: Force download even if file exists (default `false`).
- `retries`: Number of retries on failure (default `0`).
- `extract`: Unzip/untar the downloaded file into the destination directory (default `false`).

**Service-Level Example:**

```yaml
services:
  app:
    image: my-app
    x-fetch:
      - url: "https://example.com/assets.zip"
        dest: "./assets"
        extract: true
        retries: 3
    x-fetch:
      - url: "https://example.com/model.bin"
        dest: "./models/model.bin"
        force: true
```

### 3. SSL Certificates (`x-generate-cert`)

Generate self-signed SSL certificates for local development.

**Fields:**
- `domain`: Domain name for the certificate (CN).
- `output_dir`: Directory to save `server.crt` and `server.key`.
- `cert_name`: Filename for certificate (default `server.crt`).
- `key_name`: Filename for private key (default `server.key`).
- `force`: Force regeneration even if valid (default `false`).

**Logic:**
- Automatically renews if within 30 days of expiry.
- Skips if valid certificate exists (unless `force: true`).

**Example:**

```yaml
x-generate-cert:
  - domain: "myapp.local"
    output_dir: "./nginx/certs"
```

### 4. Templating (`x-template`)

Render configuration files using Go templates. Environment variables are available in the template context.

**Fields:**
- `source`: Path to template file.
- `target`: Path to output file.

**Example:**

```yaml
# config.json.tpl:
# { "db": "{{ .DATABASE_URL }}" }

services:
  api:
    x-template:
      - source: ./config.json.tpl
        target: ./config.json
```

### 5. Environment Validation (`x-required-env`)

Ensure critical environment variables are present before starting services.

**Example:**

```yaml
x-required-env:
  - DATABASE_URL
  - API_KEY
  - SECRET_TOKEN
```

## Extensions Location

All extensions can be defined:
1.  **Inside a service**: Scoped to that service. **Recommended** for robustness, especially when using Docker Compose includes or overrides, as top-level extensions can sometimes be stripped by `docker compose config`.
2.  **At the root** of `compose.yaml`: Applied globally. Note that complex compose setups might require moving these to a specific service (like the `init` service itself) to ensure they are preserved.

```yaml
services:
  init:
    image: ghcr.io/fransvandermeer/compose-init
    x-chown:
      - path: ./data
        owner: host
        recursive: true
```

## Building

```bash
docker build -t compose-init .
```
