# Firmware Registry API (Go)

Self-hosted firmware registry for ESP32 OTA.

## Features
- Multiple firmware types, each with multiple semantic versions
- Binary storage on local filesystem
- SQLite metadata with automatic migrations
- API-key auth (admin vs device) + optional OIDC/Keycloak
- Webhook notifications (HMAC signed, retry w/backoff)
- Clean separation: handlers, services, repositories, storage, config
- **Swagger/OpenAPI documentation** - Interactive API explorer
- **Structured logging** - JSON logs to file, syslog, or stdout

## Storage layout
`{FW_STORAGE_DIR}/{type}/{version}/firmware.bin`

## API Documentation (Swagger)

**Interactive documentation available at:** `http://localhost:8080/swagger/index.html`

The Swagger UI provides:
- Complete API reference with examples
- Interactive "Try it out" for all endpoints
- Authentication testing (API keys and JWT)
- Schema definitions for all DTOs

### Regenerating Swagger Docs

After modifying API endpoints or models:

```bash
# Install swag CLI (first time only)
go install github.com/swaggo/swag/cmd/swag@latest

# Generate documentation
swag init -g cmd/firmware-registry/main.go -o ./docs --parseDependency --parseInternal

# Rebuild
go build ./cmd/firmware-registry
```

**Generated files:**
- `docs/swagger.json` - OpenAPI 3.0 JSON spec
- `docs/swagger.yaml` - OpenAPI 3.0 YAML spec
- `docs/docs.go` - Go code

See [../SWAGGER.md](../SWAGGER.md) for complete documentation.

## Auth
- Admin endpoints require header: `X-Admin-Key: <FW_ADMIN_KEY>`
- Device endpoints require header: `X-Device-Key: <FW_DEVICE_KEY>`
- **OIDC/JWT** (optional): `Authorization: Bearer <token>`

## Endpoints
- GET  `/api/health`
- POST `/api/firmware/{type}/{version}` (admin, multipart field `file`)
- GET  `/api/firmware/{type}/{version}` (device, streams binary)
- DELETE `/api/firmware/{type}/{version}` (admin)
- GET  `/api/firmware/{type}` (device, list)
- GET  `/api/firmware/{type}/latest` (device, semantic latest)
- GET/POST `/api/webhooks` (admin)
- PUT/DELETE `/api/webhooks/{id}` (admin)

Webhook events:
- `firmware.uploaded`
- `firmware.deleted`

Signature:
If `FW_WEBHOOK_SECRET` is set, a header is added:
`X-Webhook-Signature = hex(HMAC-SHA256(secret, raw_body))`

## Config
Use env vars or YAML (set `FW_CONFIG_FILE=/path/config.yaml`).

**Key environment variables:**
- `FW_LISTEN_ADDR` - Server address (default: `:8080`)
- `FW_ADMIN_KEY` / `FW_DEVICE_KEY` - API authentication
- `FW_STORAGE_DIR` - Firmware binary storage path
- `FW_DB_PATH` - SQLite database path
- `FW_LOG_LEVEL` - Logging level (trace, debug, info, warn, error)
- `FW_LOG_OUTPUT` - Log destination (stdout, file, syslog, multi)
- `FW_OIDC_ENABLED` - Enable Keycloak/OIDC authentication

See [internal/config/config.go](internal/config/config.go) for all configuration options.

**Documentation:**
- **Logging**: See [../LOGGING.md](../LOGGING.md)
- **OIDC/Keycloak**: See [../KEYCLOAK_SETUP.md](../KEYCLOAK_SETUP.md)

## Migrations
Runs on boot from `./migrations` using golang-migrate.

## Development

```bash
# Run locally
go run ./cmd/firmware-registry

# Build
go build ./cmd/firmware-registry

# Run tests
go test ./...

# Regenerate Swagger docs
swag init -g cmd/firmware-registry/main.go -o ./docs --parseDependency --parseInternal
```
