# GoEdu-Theta REST API

This document describes the REST API endpoints provided by the GoEdu-Theta server.

---

## Base URL

```
http://localhost:8080
```

The server runs on the configured host and port (default: `localhost:8080`).

---

## Endpoints

### Root Endpoint

**GET /**

Returns a welcome message with basic server information and available endpoints.

#### Response

```json
{
    "message": "Welcome to GoEdu-Theta API Server",
    "status": "running",
    "version": "1.0.0",
    "timestamp": "2025-06-25T17:50:58Z",
    "endpoints": [
        "GET /",
        "GET /health",
        "GET /metrics"
    ]
}
```

#### HTTP Status Codes

- `200 OK` - Request successful

---

### Health Check Endpoint

**GET /health**

Returns the health status of the application and its dependencies. Used by load balancers and monitoring systems.

#### Response

```json
{
    "status": "healthy",
    "timestamp": "2025-06-25T17:51:07Z",
    "uptime": "95ns",
    "checks": {
        "database": "ok",
        "memory": "ok",
        "disk": "ok"
    }
}
```

#### HTTP Status Codes

- `200 OK` - Service is healthy
- `503 Service Unavailable` - Service is unhealthy (future implementation)

#### Notes

- Currently returns placeholder values for health checks
- Future versions will implement actual health checks for database, memory, and disk

---

### Metrics Endpoint

**GET /metrics**

Returns application metrics for monitoring and observability.

#### Response

```json
{
    "status": "ok",
    "timestamp": "2025-06-25T17:51:13Z",
    "metrics": {
        "http_requests_total": 0,
        "http_request_duration": 0.0,
        "active_connections": 0,
        "memory_usage_bytes": 0,
        "goroutines_count": 0
    },
    "build_info": {
        "version": "1.0.0",
        "commit": "unknown",
        "built_at": "unknown"
    }
}
```

#### HTTP Status Codes

- `200 OK` - Request successful

#### Notes

- Currently returns placeholder values for metrics
- Future versions will implement actual metric collection

---

## Error Handling

### 404 Not Found

For any endpoint that doesn't exist, the server returns:

```
404 page not found
```

With HTTP status code `404`.

---

## Configuration

The server can be configured via JSON configuration files and environment variables:

### Server Configuration

```json
{
    "server": {
        "port": 8080,
        "host": "localhost",
        "read_timeout": 30,
        "write_timeout": 30,
        "shutdown_timeout": 15
    }
}
```

### Environment Variables

- `SERVER_PORT` - Override server port
- `SERVER_HOST` - Override server host/bind address
- `SERVER_READ_TIMEOUT` - Request read timeout in seconds
- `SERVER_WRITE_TIMEOUT` - Response write timeout in seconds
- `SERVER_SHUTDOWN_TIMEOUT` - Graceful shutdown timeout in seconds

---

## Running the Server

### Using the Binary

```bash
./bin/goedu-theta
```

### Using Go Run

```bash
go run cmd/server/main.go
```

### With Environment Variables

```bash
SERVER_PORT=3000 SERVER_HOST=0.0.0.0 ./bin/goedu-theta
```

---

## Testing

### Manual Testing with curl

```bash
# Test root endpoint
curl http://localhost:8080/

# Test health endpoint
curl http://localhost:8080/health

# Test metrics endpoint
curl http://localhost:8080/metrics

# Test invalid endpoint (should return 404)
curl http://localhost:8080/invalid
```

### Automated Tests

```bash
# Run server tests
go test ./internal/server/...

# Run all tests
go test ./...
```

---

## Logging

The server provides structured logging with configurable levels and formats:

- **Development**: Pretty-printed colored logs for human readability
- **Production**: JSON logs for machine parsing
- **Configurable levels**: debug, info, warn, error

Example log output:
```
2025-06-25 17:50:58.123 INFO  🚀 Starting HTTP server | addr=localhost:8080
2025-06-25 17:50:58.456 INFO  🌐 HTTP Request | method=GET path=/ status=200 client_ip=127.0.0.1 latency=1.234ms
```

---

## Future Enhancements

1. **Authentication**: Add JWT-based authentication
2. **Rate Limiting**: Implement request rate limiting
3. **Real Metrics**: Implement actual metric collection (Prometheus compatible)
4. **Health Checks**: Add real database and system health checks
5. **API Versioning**: Add versioning support (e.g., `/api/v1/`)
6. **OpenAPI/Swagger**: Generate API documentation
7. **CORS Support**: Add CORS middleware for browser clients
8. **Request Validation**: Add input validation middleware
