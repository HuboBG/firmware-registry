# Firmware Registry Deploy (Docker Compose)

This folder runs:
- api (Go)
- ui (Vue nginx prod server)
- nginx reverse proxy (TLS optional)

## Quick start
1. Copy `.env.example` to `.env` and configure it as needed
2. Copy `api/.env.example` to `api/.env` and configure it as needed
3. Copy `ui/.env.example` to `ui/.env` and configure it as needed
4. `docker compose up -d --build`
5. Open http://localhost:8080 (UI) or http://localhost:8080/api/health

## Features

- **REST API**: Full-featured firmware management with OTA update support
- **Swagger UI**: Interactive API documentation at http://localhost:8080/swagger/index.html
- **Webhooks**: Configurable event notifications
- **Authentication**: API keys or Keycloak/OIDC
- **Comprehensive Logging**: Structured logs to file, syslog, or stdout
- **Admin UI**: Web interface for firmware and webhook management

## Documentation

- **API Documentation**: See [SWAGGER.md](SWAGGER.md) for API reference
- **Logging Setup**: See [LOGGING.md](LOGGING.md) for logging configuration
- **Keycloak/OIDC**: See [KEYCLOAK_SETUP.md](KEYCLOAK_SETUP.md) for authentication setup

## Notes
- API stores firmware binaries in volume `fw_data`
- SQLite DB stored in volume `fw_db`
- Nginx protects UI with basic auth (optional) and proxies /api to backend
- Swagger UI available at `/swagger/index.html`
