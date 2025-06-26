# MongoDB Integration Example

This document demonstrates how to integrate the MongoDB connection manager into the GoEdu-Theta application for real-world usage scenarios.

## Complete Integration Example

### 1. Update Main Function

Here's how to integrate MongoDB into the main server initialization:

```go
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/radek-zitek-cloud/goedu-theta/internal/config"
	"github.com/radek-zitek-cloud/goedu-theta/internal/database"
	"github.com/radek-zitek-cloud/goedu-theta/internal/logger"
	"github.com/radek-zitek-cloud/goedu-theta/internal/server"
)

func main() {
	// Initialize bootstrap logger
	logger.InitializeBootstrapLogger()

	// Load configuration
	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("❌ Error loading configuration", slog.Any("error", err))
		return
	}

	// Configure logger
	logger.ConfigureLogger(cfg.Logger)
	slog.Info("🔠 Logger configured successfully")

	// Initialize MongoDB connection
	dbManager, err := database.NewMongoDBManager(cfg.Database, logger.GetLogger())
	if err != nil {
		slog.Error("❌ Failed to initialize MongoDB connection", slog.Any("error", err))
		return
	}
	defer func() {
		if err := dbManager.Close(); err != nil {
			slog.Error("❌ Failed to close MongoDB connection", slog.Any("error", err))
		} else {
			slog.Info("🍃 MongoDB connection closed successfully")
		}
	}()

	slog.Info("🍃 MongoDB connection established successfully")

	// Create HTTP server with database dependency
	httpServer := server.NewServerWithDatabase(cfg.Server, logger.GetLogger(), dbManager)

	// Start HTTP server
	if err := httpServer.Start(); err != nil {
		slog.Error("❌ Failed to start HTTP server", slog.Any("error", err))
		return
	}

	slog.Info("🚀 Server started successfully with MongoDB integration")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("🛑 Shutting down server...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("❌ Server forced to shutdown", slog.Any("error", err))
	} else {
		slog.Info("✅ Server shutdown completed gracefully")
	}
}
```

### 2. Enhanced Server Structure

Extend the server to include database dependency:

```go
// In internal/server/server.go

type Server struct {
	router    *gin.Engine
	server    *http.Server
	config    config.Server
	logger    *slog.Logger
	dbManager *database.MongoDBManager // Add database manager
}

func NewServerWithDatabase(cfg config.Server, logger *slog.Logger, dbManager *database.MongoDBManager) *Server {
	// ... existing server setup ...
	
	server := &Server{
		router:    router,
		server:    httpServer,
		config:    cfg,
		logger:    logger,
		dbManager: dbManager, // Store database manager
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	// Create handler with database dependency
	h := handlers.NewHandlerWithDatabase(s.logger, s.dbManager)

	s.router.GET("/", h.HandleRoot)
	s.router.GET("/health", h.HandleHealth)
	s.router.GET("/metrics", h.HandleMetrics)
	
	// New database-related endpoints
	s.router.GET("/users", h.HandleGetUsers)
	s.router.POST("/users", h.HandleCreateUser)
	s.router.GET("/users/:id", h.HandleGetUser)
}
```

### 3. Enhanced Handler with Database Operations

Create handlers that use MongoDB:

```go
// In internal/handlers/users.go

package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/radek-zitek-cloud/goedu-theta/internal/database"
)

// User represents a user document in MongoDB
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string            `bson:"name" json:"name"`
	Email     string            `bson:"email" json:"email"`
	CreatedAt time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time         `bson:"updated_at" json:"updated_at"`
}

// UserHandler provides user-related HTTP handlers with MongoDB integration
type UserHandler struct {
	logger    *slog.Logger
	dbManager *database.MongoDBManager
}

// HandleGetUsers retrieves all users from MongoDB
func (h *UserHandler) HandleGetUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := h.dbManager.GetDatabase().Collection("users")
	
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		h.logger.Error("Failed to query users", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}
	defer cursor.Close(ctx)

	var users []User
	if err = cursor.All(ctx, &users); err != nil {
		h.logger.Error("Failed to decode users", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process users"})
		return
	}

	h.logger.Info("Successfully retrieved users", slog.Int("count", len(users)))
	c.JSON(http.StatusOK, gin.H{"users": users, "count": len(users)})
}

// HandleCreateUser creates a new user in MongoDB
func (h *UserHandler) HandleCreateUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		h.logger.Warn("Invalid user data", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user data"})
		return
	}

	// Set timestamps
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := h.dbManager.GetDatabase().Collection("users")
	
	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		h.logger.Error("Failed to create user", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	
	h.logger.Info("User created successfully", 
		slog.String("user_id", user.ID.Hex()),
		slog.String("email", user.Email))
	
	c.JSON(http.StatusCreated, gin.H{"user": user})
}

// HandleGetUser retrieves a specific user by ID
func (h *UserHandler) HandleGetUser(c *gin.Context) {
	userID := c.Param("id")
	
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		h.logger.Warn("Invalid user ID format", slog.String("user_id", userID))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := h.dbManager.GetDatabase().Collection("users")
	
	var user User
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			h.logger.Info("User not found", slog.String("user_id", userID))
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		
		h.logger.Error("Failed to retrieve user", 
			slog.String("user_id", userID),
			slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	h.logger.Info("User retrieved successfully", slog.String("user_id", userID))
	c.JSON(http.StatusOK, gin.H{"user": user})
}
```

### 4. Enhanced Health Check with Database

Update health check to include database connectivity:

```go
// In internal/handlers/health.go

func (h *Handler) HandleHealth(c *gin.Context) {
	health := HealthResponse{
		Status:     "healthy",
		Timestamp:  time.Now().UTC(),
		Version:    "1.0.0",
		Services:   make(map[string]ServiceHealth),
		SystemInfo: h.getSystemInfo(),
	}

	// Check database connectivity if available
	if h.dbManager != nil {
		dbHealth := h.checkDatabaseHealth()
		health.Services["database"] = dbHealth
		
		if dbHealth.Status != "healthy" {
			health.Status = "degraded"
		}
	}

	// Determine overall status
	statusCode := http.StatusOK
	if health.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if health.Status == "degraded" {
		statusCode = http.StatusPartialContent
	}

	h.logger.Info("Health check completed",
		slog.String("overall_status", health.Status),
		slog.Int("status_code", statusCode),
	)

	c.JSON(statusCode, health)
}

func (h *Handler) checkDatabaseHealth() ServiceHealth {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	err := h.dbManager.Ping(ctx)
	responseTime := time.Since(start)

	if err != nil {
		return ServiceHealth{
			Status:       "unhealthy",
			ResponseTime: responseTime.String(),
			Error:        err.Error(),
		}
	}

	return ServiceHealth{
		Status:       "healthy",
		ResponseTime: responseTime.String(),
		Details:      "MongoDB connection responsive",
	}
}
```

### 5. Environment Configuration

Update your `.env` file to include MongoDB settings:

```bash
# Server Configuration
SERVER_HOST=localhost
SERVER_PORT=6910

# Database Configuration
DATABASE_HOST=localhost
DATABASE_PORT=27017
DATABASE_USER=goedu_user
DATABASE_PASSWORD=secure_password
DATABASE_NAME=goedu_theta

# Logging Configuration
SLOG_LEVEL=debug
SLOG_FORMAT=pretty
SLOG_OUTPUT=stdout
SLOG_ADD_SOURCE=true
```

### 6. API Usage Examples

Once integrated, you can use the new endpoints:

```bash
# Create a user
curl -X POST http://localhost:6910/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com"}'

# Get all users
curl http://localhost:6910/users

# Get specific user
curl http://localhost:6910/users/{user_id}

# Check health with database status
curl http://localhost:6910/health
```

## Benefits of This Integration

1. **Centralized Database Management**: Single point for all database operations
2. **Comprehensive Error Handling**: Robust error handling throughout the stack
3. **Health Monitoring**: Database connectivity included in health checks
4. **Structured Logging**: All operations logged with context
5. **Performance Optimization**: Connection pooling and timeout management
6. **Security**: Credential management and connection string sanitization
7. **Testing**: Comprehensive test coverage for reliability

This integration provides a solid foundation for building data-driven features in the GoEdu-Theta application while maintaining the high standards of documentation, error handling, and observability established in the codebase.
