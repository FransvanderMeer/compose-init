# Compose Init

`compose-init` is a containerized tool designed to prepare your Docker Compose environment **before** your main services start. It ensures a consistent, "just works" developer experience by handling common initialization tasks automatically.

## Features

- **Smart Permissions**: Fix ownership of bind-mounted volumes to match the host user, preventing root-owned directories.
- **Templating**: Render configuration files using Go templates with environment variables.
- **Validation**: Fail fast if required environment variables are missing.
- **SSL Generation**: Automatically generate self-signed development certificates.
- **Resource Fetching**: Download required assets (seeds, models, etc.) if they are missing.

## Usage

Add `compose-init` to your `compose.yaml`:

```yaml
services:
  init:
    image: ghcr.io/fransvandermeer/compose-init:latest # Replace with your image
    volumes:
      - .:/project
    environment:
      - MY_ENV=hello

  web:
    image: nginx
    depends_on:
      init:
        condition: service_completed_successfully

# Configure via x- extensions
x-chown:
  - path: ./data
    owner: "host"       # Auto-resolves to the user running docker compose
    mode: "0755"

x-template:
  - source: ./config/nginx.conf.tpl
    target: ./config/nginx.conf

x-required-env:
  - MY_ENV
  - DB_PASSWORD

x-generate-cert:
  - domain: "myapp.local"
    output_dir: "./certs"

x-fetch:
  - url: "https://example.com/large-file.bin"
    dest: "./data/large-file.bin"
```

## How It Works

1. The container starts and **auto-detects the host user** by inspecting the ownership of your `compose.yaml`.
2. It parses your configuration (using `docker compose config` internally).
3. It executes enabled features in order: **Validation** -> **Fetch** -> **Template** -> **SSL** -> **Permissions**.
4. It exits, allowing your main services to start with a pristine environment.

## Building

```bash
docker build -t compose-init .
```
