# Swagger/OpenAPI Documentation

The Firmware Registry API includes complete OpenAPI 3.0 specifications and interactive Swagger UI documentation.

## Accessing Swagger UI

Once the API is running, access the interactive documentation at:

```
http://localhost:8080/swagger/index.html
```

Replace `localhost:8080` with your actual server address.

## What's Included

The Swagger documentation provides:

- **Interactive API Explorer**: Test all endpoints directly from the browser
- **Complete Endpoint Documentation**: All routes with parameters, request/response schemas
- **Authentication Configuration**: Test with API keys or JWT tokens
- **Schema Definitions**: Full DTO documentation with examples
- **Try It Out**: Execute real API calls from the documentation page

## API Overview

### Firmware Endpoints

**GET /api/firmware/{type}**
- List all firmware versions for a type
- Authentication: Device key or JWT Bearer token
- Returns: Array of firmware metadata

**GET /api/firmware/{type}/latest**
- Get the latest firmware version (by semantic version)
- Authentication: Device key or JWT Bearer token
- Returns: Firmware metadata

**GET /api/firmware/{type}/{version}**
- Download firmware binary
- Authentication: Device key or JWT Bearer token
- Returns: Binary file with SHA256 header

**POST /api/firmware/{type}/{version}**
- Upload new firmware
- Authentication: Admin key or JWT Bearer token
- Content-Type: multipart/form-data
- Form field: `file` (binary)
- Returns: Firmware metadata

**DELETE /api/firmware/{type}/{version}**
- Delete firmware and its binary
- Authentication: Admin key or JWT Bearer token
- Returns: Deletion confirmation

### Webhook Endpoints

**GET /api/webhooks**
- List all registered webhooks
- Authentication: Admin key or JWT Bearer token
- Returns: Array of webhook configurations

**POST /api/webhooks**
- Register a new webhook
- Authentication: Admin key or JWT Bearer token
- Body: WebhookDTO JSON
- Returns: Created webhook ID

**PUT /api/webhooks/{id}**
- Update webhook configuration
- Authentication: Admin key or JWT Bearer token
- Body: WebhookDTO JSON
- Returns: Update confirmation

**DELETE /api/webhooks/{id}**
- Remove webhook
- Authentication: Admin key or JWT Bearer token
- Returns: Deletion confirmation

## Authentication in Swagger UI

The API supports three authentication methods:

### 1. Admin API Key

Click **Authorize** button in Swagger UI and enter:
- **ApiKeyAuth**: Your admin API key
- Header: `X-Admin-Key`

### 2. Device API Key

Click **Authorize** button and enter:
- **DeviceKeyAuth**: Your device API key
- Header: `X-Device-Key`

### 3. JWT Bearer Token (OIDC)

Click **Authorize** button and enter:
- **BearerAuth**: `Bearer {your-jwt-token}`
- Header: `Authorization`

## Using Swagger UI

### 1. Explore Endpoints

Browse the available endpoints grouped by tags:
- **firmware**: Firmware management operations
- **webhooks**: Webhook configuration

### 2. View Details

Click any endpoint to see:
- Description and summary
- Parameters (path, query, body)
- Response schemas with examples
- Authentication requirements

### 3. Try It Out

1. Click **Try it out** button
2. Fill in required parameters
3. Add authentication if required (click Authorize)
4. Click **Execute**
5. View the response

### 4. Example: Upload Firmware

1. Navigate to **POST /api/firmware/{type}/{version}**
2. Click **Try it out**
3. Fill in:
   - `type`: `esp32-main`
   - `version`: `1.2.3`
   - `file`: Select a binary file
4. Click **Authorize** and enter your admin API key
5. Click **Execute**
6. View the response with firmware metadata

## OpenAPI Specification Files

The API specification is available in multiple formats:

- **JSON**: `http://localhost:8080/swagger/doc.json`
- **YAML**: Available in `/docs/swagger.yaml` (filesystem)
- **Go**: Generated code in `/docs/docs.go`

## Generating Documentation

The Swagger documentation is auto-generated from code annotations. To regenerate after code changes:

```bash
# Install swag CLI (first time only)
go install github.com/swaggo/swag/cmd/swag@latest

# Generate documentation
swag init -g cmd/firmware-registry/main.go -o ./docs --parseDependency --parseInternal

# Rebuild the application
go build ./cmd/firmware-registry
```

### When to Regenerate

Regenerate docs when you:
- Add new endpoints
- Change request/response structures
- Update API descriptions
- Modify authentication methods

## Annotation Format

The documentation is generated from special comments in the code:

### General API Info (in main.go)

```go
// @title           Firmware Registry API
// @version         1.0
// @description     Self-hosted firmware registry for ESP32 OTA updates
// @host            localhost:8080
// @BasePath        /api
```

### Endpoint Documentation (in handlers)

```go
// @Summary      Upload firmware
// @Description  Upload a new firmware binary for a specific type and version
// @Tags         firmware
// @Accept       multipart/form-data
// @Produce      json
// @Param        type     path      string  true  "Firmware type"
// @Param        version  path      string  true  "Semantic version"
// @Param        file     formData  file    true  "Firmware binary"
// @Success      200      {object}  firmware.FirmwareDTO
// @Failure      401      {string}  string  "Unauthorized"
// @Security     ApiKeyAuth
// @Router       /firmware/{type}/{version} [post]
```

### Model Documentation (in DTOs)

```go
type FirmwareDTO struct {
    Type    string `json:"type" example:"esp32-main"`
    Version string `json:"version" example:"1.2.3"`
}
```

## Custom Configuration

### Changing Host/Port

Edit the annotation in `cmd/firmware-registry/main.go`:

```go
// @host      your-domain.com:8080
```

Then regenerate docs.

### Adding More Details

Add or modify these annotations:
- `@contact.name`, `@contact.url`, `@contact.email`
- `@license.name`, `@license.url`
- `@termsOfService`

### Security Schemes

Currently configured:
- `ApiKeyAuth`: Admin API key (X-Admin-Key header)
- `DeviceKeyAuth`: Device API key (X-Device-Key header)
- `BearerAuth`: JWT token (Authorization header)

## Integration with Tools

### Postman

Import the OpenAPI spec into Postman:

1. Open Postman
2. Click **Import**
3. Enter URL: `http://localhost:8080/swagger/doc.json`
4. Click **Import**

### Code Generation

Generate client libraries using the OpenAPI spec:

```bash
# Install openapi-generator
npm install -g @openapitools/openapi-generator-cli

# Generate Python client
openapi-generator-cli generate \
  -i http://localhost:8080/swagger/doc.json \
  -g python \
  -o ./python-client

# Generate TypeScript client
openapi-generator-cli generate \
  -i http://localhost:8080/swagger/doc.json \
  -g typescript-axios \
  -o ./typescript-client
```

### cURL Commands

Swagger UI generates cURL commands for each request. Click **Execute** and view the **Curl** tab.

## Troubleshooting

### Swagger UI not loading

1. Verify the API is running: `curl http://localhost:8080/api/health`
2. Check the swagger route: `curl http://localhost:8080/swagger/doc.json`
3. View logs for errors
4. Ensure `/docs` directory was generated

### Endpoints not showing

1. Check handler annotations are correct
2. Regenerate docs: `swag init -g cmd/firmware-registry/main.go -o ./docs --parseDependency --parseInternal`
3. Rebuild: `go build ./cmd/firmware-registry`
4. Restart the server

### Authentication not working

1. Ensure you clicked **Authorize** in Swagger UI
2. Verify your API key/token is correct
3. Check if the endpoint requires admin vs device auth
4. View network tab in browser DevTools for request headers

### Changes not reflected

After modifying code:

```bash
# 1. Regenerate swagger docs
swag init -g cmd/firmware-registry/main.go -o ./docs --parseDependency --parseInternal

# 2. Rebuild
go build ./cmd/firmware-registry

# 3. Restart server
./firmware-registry

# 4. Hard refresh browser (Ctrl+Shift+R or Cmd+Shift+R)
```

## Best Practices

1. **Keep Annotations Updated**: Update swagger comments when changing APIs
2. **Use Examples**: Provide realistic examples in DTOs
3. **Document Errors**: Include all possible error responses
4. **Test in UI**: Verify all endpoints work in Swagger UI before deployment
5. **Version Your API**: Update `@version` when making breaking changes
6. **Secure Production**: Consider restricting Swagger UI access in production
7. **Use Tags**: Group related endpoints with consistent tags

## Production Considerations

### Disabling Swagger in Production

If you want to disable Swagger UI in production:

```go
// In router.go, conditionally register swagger
if os.Getenv("ENABLE_SWAGGER") == "true" {
    mux.HandleFunc("/swagger/", httpSwagger.Handler(
        httpSwagger.URL("/swagger/doc.json"),
    ))
}
```

### Protecting Swagger UI

Add authentication to Swagger UI:

```go
// Wrap swagger handler with basic auth
mux.HandleFunc("/swagger/", basicAuth(
    httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")),
))
```

### Hosting Separate Documentation

For production, consider:
1. Generate static HTML from OpenAPI spec
2. Host documentation separately from API
3. Use tools like ReDoc or Redocly
4. Implement API versioning

## Resources

- **Swagger Homepage**: https://swagger.io/
- **OpenAPI Specification**: https://spec.openapis.org/oas/v3.0.0
- **swaggo Documentation**: https://github.com/swaggo/swag
- **Example Annotations**: https://github.com/swaggo/swag#declarative-comments-format
