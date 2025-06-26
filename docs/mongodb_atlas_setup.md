# MongoDB Atlas Integration Guide

## Overview

The GoEdu-Theta application supports both standard MongoDB connections and MongoDB Atlas cloud connections. This guide explains how to configure and use MongoDB Atlas with the application.

## MongoDB Atlas Connection Features

### 🌐 **SRV Connection Support**
- Uses `mongodb+srv://` scheme for DNS SRV record resolution
- Automatic replica set discovery and failover
- Optimal connection routing and load balancing

### 🔒 **Enhanced Security**
- TLS/SSL encryption enabled by default
- Secure authentication with username/password
- Password URL encoding for special characters
- Connection string sanitization in logs

### ⚡ **Performance Optimization**
- Automatic retry writes (`retryWrites=true`)
- Write concern majority (`w=majority`)
- Connection pooling with Atlas-optimized settings
- DNS SRV caching for faster connections

### 📊 **Monitoring Integration**
- Application name tagging for Atlas monitoring
- Structured logging for connection events
- Health check endpoints with Atlas connectivity

## Configuration

### Configuration Fields

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

#### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `host` | string | ✅ | Atlas cluster hostname (e.g., `cluster.xxxxx.mongodb.net`) |
| `port` | int | ❌ | Ignored for Atlas connections (resolved via DNS SRV) |
| `user` | string | ✅ | Atlas database username |
| `password` | string | ✅ | Atlas database password |
| `name` | string | ✅ | Database name to connect to |
| `is_atlas` | bool | ✅ | Set to `true` for Atlas connections |
| `atlas_app_name` | string | ❌ | Application name for Atlas monitoring |

### Environment Variables

You can override configuration using environment variables:

```bash
export DATABASE_HOST="clusterzitekcloud.dznruy0.mongodb.net"
export DATABASE_USER="radek"
export DATABASE_PASSWORD="your_atlas_password"
export DATABASE_NAME="goedu_theta"
export DATABASE_IS_ATLAS="true"
export DATABASE_ATLAS_APP_NAME="ClusterZitekCloud"
```

## Getting Your Atlas Connection String

### Step 1: MongoDB Atlas Dashboard
1. Log into [MongoDB Atlas](https://cloud.mongodb.com)
2. Navigate to your cluster
3. Click "Connect" button
4. Choose "Connect your application"
5. Select "Go" as the driver
6. Copy the connection string

### Step 2: Extract Configuration Values

From the Atlas connection string:
```
mongodb+srv://radek:<password>@clusterzitekcloud.dznruy0.mongodb.net/?retryWrites=true&w=majority&appName=ClusterZitekCloud
```

Extract these values:
- **Host**: `clusterzitekcloud.dznruy0.mongodb.net`
- **User**: `radek`
- **Password**: Replace `<password>` with your actual password
- **AtlasAppName**: `ClusterZitekCloud`

## Environment-Specific Configuration

### Development Configuration

```json
{
  "database": {
    "host": "clusterzitekcloud.dznruy0.mongodb.net",
    "user": "radek",
    "password": "dev_password",
    "name": "goedu_theta_dev",
    "is_atlas": true,
    "atlas_app_name": "ClusterZitekCloud"
  }
}
```

### Production Configuration

```json
{
  "database": {
    "host": "clusterzitekcloud.dznruy0.mongodb.net",
    "user": "radek",
    "password": "production_password",
    "name": "goedu_theta_prod",
    "is_atlas": true,
    "atlas_app_name": "ClusterZitekCloud"
  }
}
```

## Connection String Generation

The application automatically generates proper Atlas connection strings:

### With Authentication (Standard)
```
mongodb+srv://username:password@cluster.host/database?retryWrites=true&w=majority&appName=AppName
```

### Connection Options Included
- `retryWrites=true` - Automatic retry for write operations
- `w=majority` - Write concern for data durability
- `appName=ClusterZitekCloud` - Application identification

## Security Best Practices

### 🔐 **Password Management**
- Never commit passwords to source control
- Use environment variables for sensitive credentials
- Consider using MongoDB Atlas API keys for automation
- Rotate passwords regularly

### 🌐 **Network Security**
- Configure Atlas IP whitelist for your application servers
- Use VPC peering for enhanced security in production
- Enable Atlas network access logs for monitoring

### 📝 **Logging Security**
- Passwords are automatically masked in application logs
- Connection strings are sanitized before logging
- Use structured logging for security audit trails

## Code Usage Examples

### Basic Connection
```go
// Load configuration
cfg, err := config.NewConfig()
if err != nil {
    log.Fatal("Failed to load config:", err)
}

// Create MongoDB manager
dbManager, err := database.NewMongoDBManager(cfg.Database, logger)
if err != nil {
    log.Fatal("Failed to connect to Atlas:", err)
}
defer dbManager.Close()

// Get database instance
db := dbManager.GetDatabase()
collection := db.Collection("users")
```

### Health Check
```go
// Check Atlas connectivity
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := dbManager.Ping(ctx); err != nil {
    log.Error("Atlas health check failed:", err)
    return http.StatusServiceUnavailable
}
```

### Collection Operations
```go
// Insert document
result, err := collection.InsertOne(context.Background(), user)
if err != nil {
    log.Error("Failed to insert user:", err)
    return err
}

// Find documents
cursor, err := collection.Find(context.Background(), bson.M{})
if err != nil {
    log.Error("Failed to find users:", err)
    return err
}
defer cursor.Close(context.Background())
```

## Troubleshooting

### Common Issues

#### Connection Timeout
```
Failed to connect to MongoDB: connection timeout
```
**Solutions:**
- Check Atlas cluster status
- Verify IP whitelist includes your server IP
- Check DNS resolution for cluster hostname

#### Authentication Failed
```
Authentication failed: user not found
```
**Solutions:**
- Verify username and password are correct
- Check database user permissions in Atlas
- Ensure user has access to the target database

#### DNS Resolution Error
```
Failed to resolve SRV record
```
**Solutions:**
- Verify cluster hostname is correct
- Check DNS configuration
- Try using standard `mongodb://` connection as fallback

### Debug Logging

Enable debug logging to see detailed connection information:

```json
{
  "logger": {
    "level": "debug"
  }
}
```

This will show:
- Connection string construction (with masked passwords)
- SRV record resolution details
- Connection pool status
- Health check results

## Monitoring and Metrics

### Atlas Dashboard
- Monitor connection count and performance
- View slow query logs
- Track database metrics and alerts
- Monitor geographic distribution

### Application Logs
```
🍃 MongoDB Atlas connection established successfully
🍃 Connection pool initialized with 100 max connections
🍃 Health check successful - cluster responsive
```

### Health Check Endpoint
The application provides a health check endpoint that includes Atlas connectivity:

```bash
curl http://localhost:6910/health
```

Response includes Atlas connection status and database accessibility.

## Migration from Standard MongoDB

### Configuration Changes
1. Set `is_atlas: true` in configuration
2. Update `host` to Atlas cluster hostname
3. Remove or ignore `port` field
4. Add `atlas_app_name` for monitoring

### Connection String Changes
- From: `mongodb://user:pass@host:port/db`
- To: `mongodb+srv://user:pass@cluster.host/db?options`

### No Code Changes Required
The existing MongoDB manager interface remains the same - only configuration changes are needed.

## Performance Considerations

### Connection Pooling
- Atlas connections use the same pool settings as standard MongoDB
- SRV resolution adds minimal overhead
- DNS caching improves subsequent connections

### Network Latency
- Atlas clusters may have higher latency than local MongoDB
- Use connection pooling to minimize connection overhead
- Consider read preferences for geographically distributed applications

### Write Concerns
- Atlas uses `w=majority` by default for durability
- May increase write latency compared to `w=1`
- Provides better data safety in replica set environments

## Support and Resources

- [MongoDB Atlas Documentation](https://docs.atlas.mongodb.com/)
- [MongoDB Go Driver Documentation](https://pkg.go.dev/go.mongodb.org/mongo-driver)
- [GoEdu-Theta Database Documentation](./mongodb_implementation.md)
- [Atlas Connection Troubleshooting](https://docs.atlas.mongodb.com/troubleshoot-connection/)
