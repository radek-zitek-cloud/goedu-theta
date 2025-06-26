# GoEdu-Theta

A high-performance HTTP server built with Go, featuring comprehensive configuration management, structured logging, and production-ready observability endpoints.

![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)

## 🚀 Features

- **High-Performance HTTP Server** - Built with Gin framework for optimal performance
- **Advanced Configuration Management** - Multi-source configuration with environment override support
- **Structured Logging** - Production-ready logging with slog and pretty formatting
- **Health & Metrics Endpoints** - Comprehensive observability for monitoring and alerting
- **Environment-Based Configuration** - Support for multiple deployment environments
- **Graceful Shutdown** - Proper resource cleanup and connection handling
- **Comprehensive Testing** - Extensive test coverage with unit and integration tests

## 📋 Table of Contents

- [Quick Start](#-quick-start)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [API Endpoints](#-api-endpoints)
- [Development](#-development)
- [Testing](#-testing)
- [Deployment](#-deployment)
- [Monitoring](#-monitoring)
- [Contributing](#-contributing)

## 🏃 Quick Start

### Prerequisites

- Go 1.24+ installed
- Git for version control

### Run the Server

```bash
# Clone the repository
git clone https://github.com/radek-zitek-cloud/goedu-theta.git
cd goedu-theta

# Copy environment configuration
cp .env.example .env

# Build and run
make run
```

The server will start on `http://localhost:6910` by default.

### Test the API

```bash
# Check server status
curl http://localhost:6910/

# Health check
curl http://localhost:6910/health

# Application metrics
curl http://localhost:6910/metrics
```

## 📦 Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/radek-zitek-cloud/goedu-theta.git
cd goedu-theta

# Install dependencies
go mod download

# Build the binary
make build

# Run the server
./bin/goedu-theta
```

### Using Make

```bash
# Build the project
make build

# Run the project
make run

# Clean build artifacts
make clean
```

### Using Go Commands

```bash
# Build
go build -o bin/goedu-theta ./cmd/server

# Run directly
go run ./cmd/server

# Run tests
go test ./...
```

## ⚙️ Configuration

GoEdu-Theta uses a sophisticated multi-source configuration system with the following precedence order:

1. **System Environment Variables** (highest priority)
2. **`.env` File Variables**
3. **Environment-Specific JSON Files** (`config.development.json`, `config.production.json`)
4. **Base JSON Configuration** (`config.json`)
5. **Default Values** (lowest priority)

### Environment Variables

Create a `.env` file in the project root:

```bash
# Server Configuration
SERVER_HOST=localhost
SERVER_PORT=6910
SERVER_READ_TIMEOUT=30
SERVER_WRITE_TIMEOUT=30
SERVER_SHUTDOWN_TIMEOUT=15

# Database Configuration (if applicable)
DATABASE_HOST=localhost
DATABASE_PORT=27017
DATABASE_USER=username
DATABASE_PASSWORD=password
DATABASE_NAME=goedu

# Logging Configuration
SLOG_LEVEL=debug
SLOG_FORMAT=pretty
SLOG_OUTPUT=stdout
SLOG_ADD_SOURCE=false
```

### JSON Configuration Files

#### Base Configuration (`configs/config.json`)

```json
{
    "environment": "development",
    "logger": {
        "level": "debug",
        "format": "pretty",
        "add_source": true,
        "output": "stdout"
    },
    "server": {
        "port": 6910,
        "host": "localhost",
        "read_timeout": 31,
        "write_timeout": 31,
        "shutdown_timeout": 16
    },
    "database": {
        "host": "localhost",
        "port": 27017,
        "user": "",
        "password": "",
        "name": "goedu"
    }
}
```

#### Environment-Specific Overrides

- `configs/config.development.json` - Development environment settings
- `configs/config.staging.json` - Staging environment settings
- `configs/config.production.json` - Production environment settings
- `configs/config.test.json` - Test environment settings
- `configs/config.local.json` - Local development overrides

### Environment Detection

The server automatically detects the environment based on the `ENVIRONMENT` variable:

```bash
export ENVIRONMENT=production  # Uses config.production.json
export ENVIRONMENT=staging     # Uses config.staging.json
export ENVIRONMENT=development # Uses config.development.json (default)
```

## 🌐 API Endpoints

### Root Endpoint

**GET** `/`

Returns API information and service status.

```json
{
    "message": "Welcome to GoEdu-Theta API Server",
    "status": "running",
    "version": "1.0.0",
    "timestamp": "2025-06-26T16:04:26Z",
    "endpoints": [
        "GET /",
        "GET /health",
        "GET /metrics"
    ]
}
```

### Health Check Endpoint

**GET** `/health`

Provides comprehensive health status for load balancers and monitoring systems.

```json
{
    "status": "healthy",
    "timestamp": "2025-06-26T16:04:26Z",
    "runtime": {
        "go_version": "go1.24.4",
        "platform": "linux/amd64",
        "cpu_cores": 8
    },
    "memory": {
        "allocated_bytes": 2097152,
        "heap_in_use_bytes": 3145728,
        "gc_cycles": 5,
        "allocated_mb": 2,
        "next_gc_bytes": 4194304
    },
    "goroutines": {
        "count": 8
    },
    "service": {
        "name": "goedu-theta",
        "environment": "development",
        "version": "1.0.0"
    }
}
```

### Metrics Endpoint

**GET** `/metrics`

Exposes detailed application metrics for monitoring and observability platforms.

```json
{
    "metadata": {
        "collected_at": "2025-06-26T16:04:26Z",
        "schema_version": "1.0.0",
        "collection_duration_ns": 125000,
        "collection_duration_ms": 0.125
    },
    "memory": {
        "heap": {
            "allocated_bytes": 2097152,
            "in_use_bytes": 3145728,
            "system_bytes": 8388608,
            "objects_count": 12543,
            "allocated_mb": 2
        },
        "gc": {
            "cycles_total": 5,
            "next_target_bytes": 4194304,
            "pause_total_ns": 750000,
            "last_pause_ns": 125000,
            "cpu_percent": 0.01
        }
    },
    "runtime": {
        "goroutines": {
            "count": 8,
            "note": "includes_system_goroutines"
        },
        "environment": {
            "go_version": "go1.24.4",
            "platform": "linux/amd64",
            "cpu_cores": 8,
            "max_procs": 8
        }
    }
}
```

## 💻 Development

### Project Structure

```
.
├── cmd/
│   └── server/                 # Application entry point
│       └── main.go
├── internal/
│   ├── config/                 # Configuration management
│   │   ├── config.go
│   │   ├── types.go
│   │   ├── defaults.go
│   │   └── test/
│   ├── handlers/               # HTTP request handlers
│   │   ├── root.go
│   │   ├── health.go
│   │   ├── metrics.go
│   │   └── test/
│   ├── logger/                 # Structured logging
│   │   ├── logger.go
│   │   ├── pretty_handler.go
│   │   └── test/
│   └── server/                 # HTTP server management
│       ├── server.go
│       └── test/
├── configs/                    # Configuration files
│   ├── config.json
│   ├── config.development.json
│   ├── config.production.json
│   └── ...
├── docs/                       # Documentation
├── scripts/                    # Build and deployment scripts
└── deployments/               # Deployment configurations
```

### Development Workflow

```bash
# Install development dependencies
go mod download

# Run tests
go test ./...

# Run with live reload (if you have air installed)
air

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o bin/goedu-theta-linux ./cmd/server
GOOS=windows GOARCH=amd64 go build -o bin/goedu-theta.exe ./cmd/server
GOOS=darwin GOARCH=amd64 go build -o bin/goedu-theta-mac ./cmd/server
```

### Code Standards

- Follow Go conventions and `gofmt` formatting
- Write comprehensive tests for all new features
- Include detailed documentation for public APIs
- Use structured logging throughout the application
- Implement proper error handling and validation

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific test package
go test ./internal/handlers/test/
go test ./internal/config/test/
go test ./internal/server/test/
```

### Test Coverage

The project maintains comprehensive test coverage:

- **Handlers**: Unit tests for all HTTP endpoints
- **Configuration**: Tests for multi-source config loading
- **Server**: Integration tests for server lifecycle
- **Logger**: Tests for logging functionality

### Test Categories

- **Unit Tests**: Fast, isolated tests for individual components
- **Integration Tests**: Tests for component interactions
- **Performance Tests**: Response time and memory usage validation

## 🚀 Deployment

### Production Configuration

1. **Set Production Environment**:
   ```bash
   export ENVIRONMENT=production
   ```

2. **Configure Production Settings**:
   ```bash
   # Server settings
   export SERVER_HOST=0.0.0.0
   export SERVER_PORT=8080
   
   # Logging settings
   export SLOG_LEVEL=info
   export SLOG_FORMAT=json
   export SLOG_ADD_SOURCE=false
   ```

3. **Build Production Binary**:
   ```bash
   CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o goedu-theta ./cmd/server
   ```

### Docker Deployment

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o goedu-theta ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/goedu-theta .
COPY --from=builder /app/configs ./configs
EXPOSE 8080
CMD ["./goedu-theta"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: goedu-theta
spec:
  replicas: 3
  selector:
    matchLabels:
      app: goedu-theta
  template:
    metadata:
      labels:
        app: goedu-theta
    spec:
      containers:
      - name: goedu-theta
        image: goedu-theta:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: SERVER_PORT
          value: "8080"
        - name: SLOG_LEVEL
          value: "info"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## 📊 Monitoring

### Health Checks

The server provides comprehensive health endpoints for monitoring:

- **Basic Health**: `/health` - Quick health status check
- **Detailed Metrics**: `/metrics` - Comprehensive application metrics

### Prometheus Integration

The metrics endpoint can be easily integrated with Prometheus for monitoring:

```yaml
scrape_configs:
  - job_name: 'goedu-theta'
    static_configs:
      - targets: ['localhost:6910']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### Grafana Dashboard

Key metrics to monitor:

- **Response Time**: API endpoint response times
- **Memory Usage**: Heap and GC metrics
- **Goroutine Count**: Concurrency and potential leaks
- **Request Rate**: Requests per second
- **Error Rate**: HTTP 4xx/5xx response rates

### Logging

Structured logging with configurable output:

- **Development**: Pretty-formatted console output
- **Production**: JSON-formatted structured logs
- **Debugging**: Detailed source location and trace information

## 🤝 Contributing

We welcome contributions! Please follow these guidelines:

1. **Fork the Repository**
2. **Create a Feature Branch**: `git checkout -b feature/amazing-feature`
3. **Make Your Changes**
4. **Add Tests**: Ensure your changes are well-tested
5. **Update Documentation**: Update README and inline docs as needed
6. **Commit Changes**: `git commit -m 'Add amazing feature'`
7. **Push to Branch**: `git push origin feature/amazing-feature`
8. **Open a Pull Request**

### Development Guidelines

- Follow Go best practices and conventions
- Write comprehensive tests for new features
- Update documentation for API changes
- Use structured logging for observability
- Ensure all tests pass before submitting

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/radek-zitek-cloud/goedu-theta/issues)
- **Documentation**: [Project Wiki](https://github.com/radek-zitek-cloud/goedu-theta/wiki)
- **Discussions**: [GitHub Discussions](https://github.com/radek-zitek-cloud/goedu-theta/discussions)

## 🔄 Changelog

### v1.0.0
- ✅ Initial release with HTTP server
- ✅ Multi-source configuration management
- ✅ Structured logging with slog
- ✅ Health and metrics endpoints
- ✅ Comprehensive test coverage
- ✅ Production-ready deployment support

---

Made with ❤️ by [Radek Zitek](https://github.com/radek-zitek-cloud)