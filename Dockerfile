FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o compose-init ./cmd/compose-init

FROM alpine:latest

# Install docker-cli for parsing config
RUN apk add --no-cache docker-cli ca-certificates curl unzip
RUN mkdir -p /usr/local/lib/docker/cli-plugins && \
    curl -SL https://github.com/docker/compose/releases/latest/download/docker-compose-linux-x86_64 -o /usr/local/lib/docker/cli-plugins/docker-compose && \
    chmod +x /usr/local/lib/docker/cli-plugins/docker-compose

WORKDIR /app
COPY --from=builder /app/compose-init .

# We expect the project to be mounted at /project
WORKDIR /project

ENTRYPOINT ["/app/compose-init"]
