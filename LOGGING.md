# Logging Configuration Guide

This document explains the comprehensive logging system implemented in the Firmware Registry.

## Overview

The Firmware Registry includes production-ready logging for both the API and UI:

- **API (Go)**: Structured JSON logging with support for file, syslog, and console output
- **UI (Vue)**: Console-based logging for development and debugging

## API Logging (Go/zerolog)

### Features

- **Structured Logging**: JSON-formatted logs with consistent fields
- **Multiple Outputs**: stdout, file, syslog, or multi (all at once)
- **Log Rotation**: Automatic rotation with configurable size/age limits
- **Log Levels**: trace, debug, info, warn, error, fatal, panic
- **HTTP Request Logging**: Automatic logging of all HTTP requests with status, duration, and size
- **Authentication Logging**: Detailed logging of auth attempts and failures
- **Operation Logging**: Comprehensive logging of firmware operations and webhooks

### Configuration

All logging settings can be configured via environment variables:

```bash
# Log Level (trace, debug, info, warn, error, fatal, panic)
FW_LOG_LEVEL=info

# Log Format (json or console)
# - json: Structured JSON for production
# - console: Human-readable for development
FW_LOG_FORMAT=json

# Log Output (stdout, file, syslog, multi)
# - stdout: Output to standard output
# - file: Write to log file with rotation
# - syslog: Send to syslog server
# - multi: Output to stdout + file + syslog (if configured)
FW_LOG_OUTPUT=stdout

# File Logging Configuration (when output=file or multi)
FW_LOG_FILE_PATH=/var/log/firmware-registry/app.log
FW_LOG_MAX_SIZE_MB=100          # Max size before rotation
FW_LOG_MAX_BACKUPS=3            # Max number of old log files
FW_LOG_MAX_AGE_DAYS=28          # Max age in days
FW_LOG_COMPRESS=true            # Compress rotated files

# Syslog Configuration (when output=syslog or multi)
FW_LOG_SYSLOG_ADDR=syslog.example.com:514  # Empty for local syslog
FW_LOG_SYSLOG_NET=udp                       # tcp, udp, or empty
```

### Example Configurations

#### Development (Console Output)

```bash
FW_LOG_LEVEL=debug
FW_LOG_FORMAT=console
FW_LOG_OUTPUT=stdout
```

#### Production (File with Rotation)

```bash
FW_LOG_LEVEL=info
FW_LOG_FORMAT=json
FW_LOG_OUTPUT=file
FW_LOG_FILE_PATH=/var/log/firmware-registry/app.log
FW_LOG_MAX_SIZE_MB=100
FW_LOG_MAX_BACKUPS=5
FW_LOG_MAX_AGE_DAYS=30
FW_LOG_COMPRESS=true
```

#### Production (Syslog)

```bash
FW_LOG_LEVEL=info
FW_LOG_FORMAT=json
FW_LOG_OUTPUT=syslog
FW_LOG_SYSLOG_ADDR=syslog.example.com:514
FW_LOG_SYSLOG_NET=udp
```

#### Production (Multi-output: stdout + file + syslog)

```bash
FW_LOG_LEVEL=info
FW_LOG_FORMAT=json
FW_LOG_OUTPUT=multi
FW_LOG_FILE_PATH=/var/log/firmware-registry/app.log
FW_LOG_SYSLOG_ADDR=syslog.example.com:514
```

### Log Structure

All logs follow a consistent JSON structure:

```json
{
  "level": "info",
  "time": 1234567890,
  "caller": "main.go:45",
  "message": "Firmware uploaded successfully",
  "type": "esp32-main",
  "version": "1.2.3",
  "filename": "firmware.bin",
  "size_bytes": 524288,
  "sha256": "abc123...",
  "duration_ms": 1234
}
```

### What Gets Logged

#### Application Startup
- Configuration loading
- Logger initialization
- Database connection and migrations
- OIDC initialization (if enabled)
- Server listening

#### HTTP Requests
- All incoming requests with method, path, query params
- Response status, duration, and size
- Client IP and user agent

#### Authentication
- Successful authentications (JWT and API key)
- Failed authentication attempts with reason
- JWT verification failures
- Missing roles

#### Firmware Operations
- Upload start/progress/completion
- SHA256 computation
- File storage operations
- Database upserts
- Download requests
- Delete operations

#### Webhook Dispatches
- Event triggering with webhook count
- Individual webhook delivery attempts
- Success/failure with status codes
- Retry attempts with backoff timing
- Final delivery failures

#### Errors
- Database errors with SQL context
- File system errors
- Network errors
- Configuration errors

### Log Rotation

When using file output, logs are automatically rotated based on:

- **Size**: When log file reaches `FW_LOG_MAX_SIZE_MB`
- **Age**: Files older than `FW_LOG_MAX_AGE_DAYS` are deleted
- **Count**: Only `FW_LOG_MAX_BACKUPS` old files are kept

Rotated files are named: `app.log.1`, `app.log.2`, etc.

If compression is enabled (`FW_LOG_COMPRESS=true`), rotated files are gzipped: `app.log.1.gz`

### Syslog Integration

The API can send logs to a syslog server:

#### Local Syslog

```bash
FW_LOG_OUTPUT=syslog
FW_LOG_SYSLOG_ADDR=       # Empty for local
```

#### Remote Syslog (UDP)

```bash
FW_LOG_OUTPUT=syslog
FW_LOG_SYSLOG_ADDR=syslog.example.com:514
FW_LOG_SYSLOG_NET=udp
```

#### Remote Syslog (TCP)

```bash
FW_LOG_OUTPUT=syslog
FW_LOG_SYSLOG_ADDR=syslog.example.com:514
FW_LOG_SYSLOG_NET=tcp
```

## UI Logging (Vue/TypeScript)

### Features

- **Console Logging**: Color-coded output to browser console
- **Log Levels**: debug, info, warn, error
- **Context Support**: Attach arbitrary data to log entries
- **Error Tracking**: Automatic error object logging
- **Development/Production Modes**: Different log levels per environment

### Usage

Import and use the logger:

```typescript
import { logger } from './logger';

// Debug messages (only in development)
logger.debug('Component mounted', { componentName: 'FirmwareList' });

// Info messages
logger.info('User authenticated', { username: 'admin' });

// Warnings
logger.warn('API response delayed', { duration: 5000 });

// Errors
logger.error('Failed to upload firmware', error, {
  filename: 'firmware.bin',
  size: 1024000
});
```

### Log Levels

- **DEBUG**: Only visible in development mode (`import.meta.env.DEV`)
- **INFO**: Always visible, for important events
- **WARN**: For non-critical issues
- **ERROR**: For errors and exceptions

### What Gets Logged

The UI logs:

- **Authentication**: Login/logout, token renewal, auth failures
- **API Calls**: All API requests/responses (debug level)
- **API Errors**: Failed API calls with error details
- **Component Lifecycle**: Component mounting/unmounting (if added)
- **User Actions**: Critical user interactions (if added)

### Console Output

Logs appear in the browser console with color coding:

```
[2025-01-24T12:34:56.789Z] [INFO] User authenticated { username: "admin" }
[2025-01-24T12:34:57.123Z] [ERROR] API request failed Error: Network Error { url: "/api/firmware" }
```

### Production Considerations

In production builds:

- DEBUG level is disabled
- INFO, WARN, ERROR levels remain active
- Consider implementing backend error reporting for ERROR level logs
- The logger includes a commented example of sending errors to a backend

To enable backend error reporting:

```typescript
// In logger.ts, uncomment and implement:
if (level === LogLevel.ERROR && !import.meta.env.DEV) {
    this.sendToBackend(entry);
}

private async sendToBackend(entry: LogEntry): Promise<void> {
    try {
        await fetch('/api/logs', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(entry)
        });
    } catch (error) {
        // Silently fail to avoid infinite loops
        console.error('Failed to send log to backend:', error);
    }
}
```

## Docker Logging

### API Container

In Docker, configure logging via environment variables in your `.env` file:

```bash
# For container logs (works with docker logs)
FW_LOG_OUTPUT=stdout
FW_LOG_FORMAT=json

# Or write to volume-mounted log directory
FW_LOG_OUTPUT=file
FW_LOG_FILE_PATH=/var/log/firmware-registry/app.log
```

Then mount the log directory:

```yaml
services:
  api:
    volumes:
      - ./logs:/var/log/firmware-registry
```

### UI Container

UI logs go to the browser console automatically. For server-side logging (nginx), nginx access logs are separate from the application.

## Log Analysis

### Using jq for JSON Logs

```bash
# View all errors
cat app.log | grep '"level":"error"' | jq .

# View firmware uploads
cat app.log | grep 'Firmware uploaded' | jq .

# View authentication failures
cat app.log | grep 'authentication failed' | jq .

# View slow requests (>1 second)
cat app.log | jq 'select(.duration_ms > 1000)'

# View requests from specific IP
cat app.log | jq 'select(.remote_addr == "192.168.1.100")'
```

### Using grep for Quick Searches

```bash
# Find all errors
grep ERROR app.log

# Find firmware operations
grep "firmware" app.log -i

# Find webhook deliveries
grep "webhook" app.log -i
```

## Troubleshooting

### Logs not appearing

1. Check log level - if set to `error`, you won't see `info` messages
2. Verify file permissions if using file output
3. Check syslog server connectivity if using syslog output

### Log file not rotating

1. Verify `FW_LOG_MAX_SIZE_MB` is set
2. Ensure the application has write permissions
3. Check disk space

### Syslog not working

1. Verify `FW_LOG_SYSLOG_ADDR` is correct
2. Test network connectivity: `nc -zv syslog.example.com 514`
3. Check firewall rules
4. Try TCP instead of UDP if UDP packets are being dropped

### Too many logs

1. Increase log level from `debug` to `info` or `warn`
2. Reduce log retention: `FW_LOG_MAX_AGE_DAYS=7`
3. Reduce max backups: `FW_LOG_MAX_BACKUPS=2`

### Logs taking too much disk space

1. Enable compression: `FW_LOG_COMPRESS=true`
2. Reduce max size: `FW_LOG_MAX_SIZE_MB=50`
3. Reduce retention: `FW_LOG_MAX_AGE_DAYS=14`

## Best Practices

1. **Use INFO for production**: Debug logs are verbose, use `info` for production
2. **Enable log rotation**: Always set max size and age limits
3. **Compress old logs**: Save disk space with `FW_LOG_COMPRESS=true`
4. **Use JSON format**: Easier to parse and analyze
5. **Monitor ERROR logs**: Set up alerts for error-level logs
6. **Use multi-output cautiously**: Writing to multiple destinations increases overhead
7. **Secure log files**: Logs may contain sensitive information, set appropriate permissions
8. **Regular log analysis**: Review logs regularly to identify issues early

## Security Considerations

- Logs may contain IP addresses, user agents, and authentication attempts
- Avoid logging sensitive data like API keys or passwords
- Restrict log file access to authorized users only
- Consider log retention policies and compliance requirements
- Use encrypted transport for remote syslog (TLS)
