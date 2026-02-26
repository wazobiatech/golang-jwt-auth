                                  # Auth Middleware for Go

A comprehensive JWT authentication middleware library for Go microservices, providing multi-token type support, JWKS validation, Redis caching, and integrations with popular Go web frameworks.

## Features

- **Multi-Token Support**: Platform, Project, User, and Service tokens
- **JWKS Integration**: Automatic key fetching and caching from Mercury
- **Redis Caching**: Token validation and JWKS caching
- **Framework Support**: Gin, Echo, Fiber, Chi, and standard net/http
- **GraphQL Integration**: Middleware for GraphQL resolvers
- **Scope Validation**: Fine-grained permission checking
- **Production Ready**: Comprehensive error handling and logging

## Installation

```bash
go get github.com/wazobiatech/auth-middleware-go
```

## Quick Start

### Gin Framework Example

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
    authgo "github.com/wazobiatech/auth-middleware-go"
)

func main() {
    r := gin.Default()

    // JWT protected routes
    protected := r.Group("/api")
    protected.Use(authgo.GinJWTMiddleware())
    {
        protected.GET("/profile", func(c *gin.Context) {
            user := authgo.GinMustGetAuthUser(c)
            c.JSON(http.StatusOK, gin.H{"user": user})
        })
    }

    // Project token protected routes
    project := r.Group("/admin")
    project.Use(authgo.GinProjectMiddleware("my-service"))
    {
        project.GET("/projects", func(c *gin.Context) {
            project := authgo.GinMustGetProjectContext(c)
            c.JSON(http.StatusOK, gin.H{"project": project})
        })
    }

    r.Run(":8080")
}
```

## Configuration

Set the following environment variables:

```bash
# Mercury service configuration
MERCURY_BASE_URL=https://mercury.yourdomain.com
SIGNATURE_SHARED_SECRET=your-signature-secret

# Redis configuration  
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=optional-password
REDIS_DB=0

# Service credentials
CLIENT_ID=your-client-id
CLIENT_SECRET=your-client-secret

# Cache settings
CACHE_EXPIRY_TIME=3600    # Token cache TTL in seconds
JWKS_CACHE_TTL=18000      # JWKS cache TTL in seconds

# Logging
LOG_LEVEL=info
```

## Token Types

### User Tokens (JWT)

For user authentication via Authorization header:

```go
// Middleware
r.Use(authgo.GinJWTMiddleware())

// In handler
user := authgo.GinMustGetAuthUser(c)
fmt.Printf("User: %s (%s)", user.Name, user.Email)
```

### Project Tokens

For service-to-service authentication via x-project-token header:

```go
// Middleware
r.Use(authgo.GinProjectMiddleware("my-service"))

// In handler  
project := authgo.GinMustGetProjectContext(c)
fmt.Printf("Tenant: %s", project.TenantID)
```

### Platform Tokens

For platform-level operations:

```go
platform := authgo.GinMustGetPlatformContext(c)
fmt.Printf("Platform: %s", platform.TenantID)
```

### Service Tokens

For service-to-service communication:

```go
service := authgo.GinMustGetServiceContext(c)
fmt.Printf("Service: %s", service.ServiceName)
```

## Scope Validation

Require specific scopes for endpoints:

```go
// Require multiple scopes
protected.Use(authgo.GinRequireScope("users:read", "users:write"))

// In handler, scopes are automatically validated
protected.POST("/users", createUser)
```

## Optional Authentication

Allow endpoints to work with or without authentication:

```go
r.Use(authgo.GinOptionalJWTMiddleware())
r.Use(authgo.GinOptionalProjectMiddleware("my-service"))

// Handler works for both authenticated and anonymous users
r.GET("/data", func(c *gin.Context) {
    if user, ok := authgo.GinGetAuthUser(c); ok {
        // User is authenticated
        c.JSON(200, gin.H{"user_data": user})
    } else {
        // Anonymous access
        c.JSON(200, gin.H{"public_data": "available to all"})
    }
})
```

## Framework Support

### Gin

```go
import "github.com/wazobiatech/auth-middleware-go"

r.Use(authgo.GinJWTMiddleware())
r.Use(authgo.GinProjectMiddleware("service-name"))
```

### Echo (Coming Soon)

```go
import "github.com/wazobiatech/auth-middleware-go/pkg/adapters/echo"

e.Use(echo.JWTMiddleware())
e.Use(echo.ProjectMiddleware("service-name"))
```

### Standard net/http (Coming Soon)

```go
import "github.com/wazobiatech/auth-middleware-go/pkg/adapters/nethttp"

http.Handle("/api/", nethttp.JWTMiddleware(handler))
```

## Error Handling

The library provides structured error types:

```go
if err != nil {
    if authErr, ok := err.(*authgo.AuthError); ok {
        switch authErr.Code {
        case authgo.ErrCodeInvalidToken:
            // Handle invalid token
        case authgo.ErrCodeExpiredToken:
            // Handle expired token  
        case authgo.ErrCodeInsufficientScope:
            // Handle insufficient permissions
        }
    }
}
```

## Logging

Configure structured logging:

```go
logger := authgo.NewLogger("my-service")
logger.Info("Authentication successful", map[string]interface{}{
    "user_id": user.UUID,
    "tenant_id": user.TenantID,
})
```

## Advanced Usage

### Custom Configuration

```go
config := authgo.GetConfig()
config.CacheExpiryTime = 7200 // 2 hours

// Or update at runtime
authgo.UpdateConfig(map[string]interface{}{
    "CACHE_EXPIRY_TIME": 7200,
})
```

### Manual Token Validation

```go
jwtAuth := authgo.NewJwtAuthMiddleware()
user, err := jwtAuth.Authenticate(request)
if err != nil {
    // Handle error
}

projectAuth := authgo.NewProjectAuthMiddleware("my-service")
authReq, err := projectAuth.Authenticate(request)
if err != nil {
    // Handle error
}
```

### Redis Operations

```go
redis := authgo.NewRedisClient()
err := redis.Set("key", "value", time.Hour)
value, err := redis.Get("key")
```

### JWKS Management

```go
jwksCache := authgo.NewJWKSCache()
keyStore, err := jwksCache.GetOrFetch("tenant-123", jwksURL, jwksPath)
```

## Testing

Run the example server:

```bash
cd examples/gin
go run main.go
```

Test endpoints:

```bash
# Public endpoint
curl http://localhost:8080/health

# JWT protected (requires Authorization header)
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://localhost:8080/user/profile

# Project protected (requires x-project-token header)  
curl -H "x-project-token: Bearer YOUR_PROJECT_TOKEN" \
     http://localhost:8080/api/v1/projects
```

## Production Deployment

1. **Environment Variables**: Set all required environment variables
2. **Redis**: Ensure Redis is accessible and properly configured
3. **Mercury**: Configure Mercury base URL and signature secret
4. **Monitoring**: Set up logging and monitoring for auth failures
5. **Security**: Use strong secrets and secure communication

## Migration from Node.js

This Go library is a direct port of the Node.js `@wazobiatech/auth-middleware` package with the same API design and functionality. Key differences:

- Function names use Go conventions (PascalCase)
- Context is passed explicitly vs. implicit in middleware
- Error handling uses Go's explicit error returns  
- Configuration via environment variables (same names)

## API Reference

### Types

```go
type AuthUser struct {
    UUID        string   `json:"uuid"`
    Email       string   `json:"email"`
    Name        string   `json:"name"`
    TenantID    string   `json:"tenant_id"`
    Permissions []string `json:"permissions"`
    TokenID     string   `json:"token_id"`
}

type ProjectContext struct {
    TenantID        string    `json:"tenant_id"`
    ProjectUUID     string    `json:"project_uuid"`
    EnabledServices []string  `json:"enabled_services"`
    Scopes          []string  `json:"scopes"`
    SecretVersion   int       `json:"secret_version"`
    TokenID         string    `json:"token_id"`
    ExpiresAt       time.Time `json:"expires_at"`
}
```

### Middleware Functions

```go
// JWT Middleware
func GinJWTMiddleware() gin.HandlerFunc
func GinOptionalJWTMiddleware() gin.HandlerFunc

// Project Middleware  
func GinProjectMiddleware(serviceName string) gin.HandlerFunc
func GinOptionalProjectMiddleware(serviceName string) gin.HandlerFunc

// Scope Middleware
func GinRequireScope(scopes ...string) gin.HandlerFunc
```

### Context Helpers

```go
// Get authentication data from context
func GinGetAuthUser(c *gin.Context) (*AuthUser, bool)
func GinGetProjectContext(c *gin.Context) (*ProjectContext, bool)
func GinGetPlatformContext(c *gin.Context) (*PlatformContext, bool)
func GinGetServiceContext(c *gin.Context) (*ServiceContext, bool)

// Must variants (panic if not found)
func GinMustGetAuthUser(c *gin.Context) *AuthUser
func GinMustGetProjectContext(c *gin.Context) *ProjectContext
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Support

For issues and questions:
- Create an issue on GitHub
- Contact the Platform Team
- Check the examples/ directory for usage patterns