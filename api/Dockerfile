# ============================
# 1. Build Stage (CGO enabled)
# ============================
FROM --platform=$BUILDPLATFORM golang:1.25.1-bullseye AS builder
WORKDIR /app

# Enable CGO because go-sqlite3 requires it
ENV CGO_ENABLED=1
ENV GOOS=linux

# Install sqlite3 build deps + cross-compilation toolchains
RUN apt-get update && apt-get install -y \
    gcc g++ make sqlite3 libsqlite3-dev \
    gcc-aarch64-linux-gnu \
    gcc-arm-linux-gnueabihf \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build native binary for correct arch
ARG TARGETOS
ARG TARGETARCH

# Set CC based on target architecture for cross-compilation
RUN case "$TARGETARCH" in \
        arm64) export CC=aarch64-linux-gnu-gcc ;; \
        arm) export CC=arm-linux-gnueabihf-gcc ;; \
        *) export CC=gcc ;; \
    esac && \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=1 \
    go build -o firmware-registry-api ./cmd/firmware-registry

# ============================
# 2. Runtime Stage
# ============================
FROM debian:bookworm-slim
WORKDIR /app

# Required because sqlite3 uses libsqlite3
RUN apt-get update && apt-get install -y --no-install-recommends \
    libsqlite3-0 \
    && rm -rf /var/lib/apt/lists/*

COPY migrations ./migrations
COPY --from=builder /app/firmware-registry-api /usr/local/bin/firmware-registry-api

EXPOSE 8080

CMD ["firmware-registry-api"]