# MongoDB Connection Implementation

## Overview

The GoEdu-Theta application now includes a comprehensive MongoDB connection implementation that provides robust database connectivity with extensive error handling, connection pooling, and monitoring capabilities.

## Architecture

### MongoDB Manager Structure

The `MongoDBManager` struct provides a centralized interface for all database operations:

```go
type MongoDBManager struct {
    client           *mongo.Client     // MongoDB client with connection pooling
    config           config.Database   // Database configuration
    logger           *slog.Logger      // Structured logging
    database         *mongo.Database   // Database instance
    connectionString string            // Sanitized connection URI
    isConnected      bool              // Connection status tracking
}
```

### Key Features

#### 1. **Connection Management**
- Automatic connection pooling with configurable parameters
- Connection health monitoring and auto-reconnection
- Graceful connection establishment with timeout handling
- Proper resource cleanup and connection disposal
- **MongoDB Atlas SRV Support**: Native support for Atlas cloud connections

#### 2. **Security Features**
- Secure authentication handling (username/password)
- Connection string sanitization to prevent credential exposure
- TLS/SSL support for encrypted connections
- Network timeout protection against slow/unresponsive servers
- **Password URL Encoding**: Safe handling of special characters in passwords

#### 3. **Performance Optimization**
- Connection pooling reduces connection overhead (100 max, 5 min connections)
- Configurable timeout values for different operation types
- Efficient connection reuse across application requests
- Background connection health monitoring (10-second intervals)
- **SRV DNS Resolution**: Automatic replica set discovery for Atlas clusters

#### 4. **Monitoring and Observability**
- Structured logging for all connection events
- Connection metrics and health status tracking
- Error logging with detailed context information
- Performance metrics for connection operations

## Configuration

The MongoDB connection uses the existing configuration system defined in `/internal/config/types.go`:

```go
type Database struct {
    Host         string `json:"host" yaml:"host" env:"DATABASE_HOST"`
    Port         int    `json:"port" yaml:"port" env:"DATABASE_PORT"`
    User         string `json:"user" yaml:"user" env:"DATABASE_USER"`
    Password     string `json:"password" yaml:"password" env:"DATABASE_PASSWORD"`
    Name         string `json:"name" yaml:"name" env:"DATABASE_NAME"`
    IsAtlas      bool   `json:"is_atlas" yaml:"is_atlas" env:"DATABASE_IS_ATLAS"`
    AtlasAppName string `json:"atlas_app_name" yaml:"atlas_app_name" env:"DATABASE_ATLAS_APP_NAME"`
}
```

### Standard MongoDB Configuration

Configure MongoDB connection using environment variables:

```bash
# Standard MongoDB Configuration
DATABASE_HOST=localhost
DATABASE_PORT=27017
DATABASE_USER=myuser
DATABASE_PASSWORD=mypassword
DATABASE_NAME=goedu
DATABASE_IS_ATLAS=false
```

### MongoDB Atlas Configuration

Configure MongoDB Atlas connection:

```bash
# MongoDB Atlas Configuration
DATABASE_HOST=clusterzitekcloud.dznruy0.mongodb.net
DATABASE_USER=radek
DATABASE_PASSWORD=your_atlas_password
DATABASE_NAME=goedu_theta
DATABASE_IS_ATLAS=true
DATABASE_ATLAS_APP_NAME=ClusterZitekCloud
```

### JSON Configuration

Configure in `configs/config.json`:

#### Standard MongoDB
```json
{
    "database": {
        "host": "localhost",
        "port": 27017,
        "user": "myuser",
        "password": "mypassword",
        "name": "goedu",
        "is_atlas": false
    }
}
```

#### MongoDB Atlas
```json
{
    "database": {
        "host": "clusterzitekcloud.dznruy0.mongodb.net",
        "port": 27017,
        "user": "radek",
        "password": "your_atlas_password",
        "name": "goedu_theta",
        "is_atlas": true,
        "atlas_app_name": "ClusterZitekCloud"
    }
}
```

> **Note**: For detailed MongoDB Atlas setup instructions, see [MongoDB Atlas Setup Guide](./mongodb_atlas_setup.md).

## Usage Examples

### Basic Connection

```go
package main

import (
    "log"
    "log/slog"
    
    "github.com/radek-zitek-cloud/goedu-theta/internal/config"
    "github.com/radek-zitek-cloud/goedu-theta/internal/database"
)

func main() {
    // Load configuration
    cfg, err := config.NewConfig()
    if err != nil {
        log.Fatal("Failed to load configuration:", err)
    }
    
    // Create logger
    logger := slog.Default()
    
    // Establish database connection
    dbManager, err := database.NewMongoDBManager(cfg.Database, logger)
    if err != nil {
        log.Fatal("Failed to connect to MongoDB:", err)
    }
    defer dbManager.Close()
    
    // Use database
    db := dbManager.GetDatabase()
    collection := db.Collection("users")
    
    // Perform operations...
}
```

### Health Check Example

```go
import (
    "context"
    "time"
)

// Check database connectivity
func checkDatabaseHealth(dbManager *database.MongoDBManager) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    return dbManager.Ping(ctx)
}
```

### Collection Operations

```go
import (
    "context"
    "go.mongodb.org/mongo-driver/bson"
)

func insertUser(dbManager *database.MongoDBManager, user User) error {
    collection := dbManager.GetDatabase().Collection("users")
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    _, err := collection.InsertOne(ctx, user)
    return err
}

func findUsers(dbManager *database.MongoDBManager) ([]User, error) {
    collection := dbManager.GetDatabase().Collection("users")
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var users []User
    if err = cursor.All(ctx, &users); err != nil {
        return nil, err
    }
    
    return users, nil
}
```

## Connection Pool Configuration

The MongoDB manager is configured with production-ready connection pool settings:

```go
// Connection Pool Settings
MaxPoolSize:      100,              // Maximum concurrent connections
MinPoolSize:      5,                // Minimum maintained connections
MaxConnIdleTime:  30 * time.Minute, // Idle connection timeout

// Timeout Settings
ConnectTimeout:         10 * time.Second, // Connection establishment
ServerSelectionTimeout: 5 * time.Second,  // Server selection
SocketTimeout:          30 * time.Second, // Individual operations
HeartbeatInterval:      10 * time.Second, // Health check frequency
```

## Error Handling

The implementation provides comprehensive error handling:

### Connection Errors
```go
manager, err := database.NewMongoDBManager(cfg.Database, logger)
if err != nil {
    // Handle connection failure
    switch {
    case strings.Contains(err.Error(), "validation failed"):
        // Configuration validation error
    case strings.Contains(err.Error(), "connection failed"):
        // Network connectivity issue
    case strings.Contains(err.Error(), "authentication failed"):
        // Invalid credentials
    default:
        // Other database errors
    }
}
```

### Health Check Errors
```go
err := manager.Ping(ctx)
if err != nil {
    // Database is not responsive
    // Handle graceful degradation
}
```

## Testing

The implementation includes comprehensive unit tests:

- **Configuration validation testing**
- **Connection lifecycle testing**
- **Health check functionality**
- **Error handling scenarios**
- **Backward compatibility verification**

Run tests:
```bash
go test ./internal/database/test/... -v
```

**Note**: Tests require a running MongoDB instance or will be skipped gracefully.

## Integration with Server

### Health Check Endpoint

The database connection can be integrated with health check endpoints:

```go
// In handlers/health.go
func (h *Handler) HandleHealth(c *gin.Context) {
    health := HealthResponse{
        Status:    "healthy",
        Timestamp: time.Now().UTC(),
        Services:  make(map[string]string),
    }
    
    // Check database connectivity
    if dbManager != nil {
        ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
        defer cancel()
        
        if err := dbManager.Ping(ctx); err != nil {
            health.Status = "unhealthy"
            health.Services["database"] = "unreachable"
        } else {
            health.Services["database"] = "healthy"
        }
    }
    
    c.JSON(http.StatusOK, health)
}
```

### Graceful Shutdown

Integrate with server shutdown:

```go
// In main.go shutdown sequence
func gracefulShutdown(dbManager *database.MongoDBManager) {
    // Close database connection
    if err := dbManager.Close(); err != nil {
        slog.Error("Failed to close database connection", slog.Any("error", err))
    } else {
        slog.Info("Database connection closed successfully")
    }
}
```

## Security Considerations

1. **Credential Management**: Never hardcode database credentials
2. **Connection String Logging**: Passwords are masked in logs
3. **TLS Connections**: Support for encrypted connections
4. **Network Timeouts**: Protection against slowloris attacks
5. **Authentication**: Secure username/password handling

## Performance Recommendations

1. **Connection Reuse**: Use the singleton manager instance
2. **Context Timeouts**: Always use context with timeouts
3. **Connection Pooling**: Leverage automatic connection pooling
4. **Index Optimization**: Create appropriate database indexes
5. **Query Optimization**: Use efficient query patterns

## Monitoring and Observability

The implementation provides extensive logging:

```
🍃 Initializing MongoDB connection manager
🍃 Creating MongoDB client with connection pool configuration
🍃 Establishing connection to MongoDB server
🍃 Verifying MongoDB server connectivity and health
🍃 MongoDB connection established successfully
🍃 Performing MongoDB connection health check
🍃 Initiating MongoDB connection cleanup
```

All logs include structured data for monitoring systems integration.

## Dependencies

The implementation uses:
- `go.mongodb.org/mongo-driver/mongo` - Official MongoDB Go driver
- `go.mongodb.org/mongo-driver/mongo/options` - Connection options
- `go.mongodb.org/mongo-driver/mongo/readpref` - Read preferences
- Existing configuration system
- Structured logging (slog)

## Future Enhancements

Potential improvements:
1. **Replica Set Support**: Configuration for MongoDB replica sets
2. **Connection Metrics**: Detailed connection pool metrics
3. **Circuit Breaker**: Automatic failure detection and recovery
4. **TLS Configuration**: Advanced TLS/SSL settings
5. **Connection Retry Logic**: Automatic reconnection strategies
