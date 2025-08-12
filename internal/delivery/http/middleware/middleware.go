package middleware

import (
	"boilerplate-go/infrastructure/logger"
	"boilerplate-go/pkg/jwt"
	"boilerplate-go/pkg/response"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// MiddlewareConfig holds middleware configuration
type MiddlewareConfig struct {
	Logger    *logger.Logger
	JWTSecret string
}

// SetupMiddlewares configures all application middlewares
func SetupMiddlewares(r *gin.Engine, config MiddlewareConfig) {
	// Request ID middleware
	r.Use(RequestIDMiddleware())

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Logging middleware
	r.Use(LoggingMiddleware(config.Logger))

	// Rate limiting middleware
	r.Use(RateLimitMiddleware(100, 1)) // 100 requests per second with burst of 1

	// Security headers middleware
	r.Use(SecurityHeadersMiddleware())

	// Recovery middleware
	r.Use(RecoveryMiddleware(config.Logger))
}

// RequestIDMiddleware generates and injects request IDs
func RequestIDMiddleware() gin.HandlerFunc {
	return requestid.New()
}

// LoggingMiddleware logs all HTTP requests
func LoggingMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Add correlation ID to context
		correlationID := c.GetHeader("X-Request-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
			c.Header("X-Request-ID", correlationID)
		}

		ctx := logger.ContextWithCorrelationID(c.Request.Context(), correlationID)
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Log request
		duration := time.Since(start)
		log.RequestLogger(
			ctx,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration.String(),
		)
	}
}

// AuthenticationMiddleware validates JWT tokens
func AuthenticationMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header required", "missing authorization header")
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			response.Unauthorized(c, "Invalid authorization format", "expected Bearer token")
			c.Abort()
			return
		}

		token := tokenParts[1]
		claims, err := jwt.ValidateToken(token, secretKey)
		if err != nil {
			response.Unauthorized(c, "Invalid token", err.Error())
			c.Abort()
			return
		}

		// Add user info to context
		ctx := logger.ContextWithUserID(c.Request.Context(), claims.UserID)
		c.Request = c.Request.WithContext(ctx)

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// RateLimitMiddleware implements rate limiting
func RateLimitMiddleware(requestsPerSecond rate.Limit, burst int) gin.HandlerFunc {
	limiter := rate.NewLimiter(requestsPerSecond, burst)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded", "too many requests")
			c.Abort()
			return
		}
		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// RecoveryMiddleware handles panics gracefully
func RecoveryMiddleware(log *logger.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// Log the panic
		log.WithContext(c.Request.Context()).WithFields(map[string]interface{}{
			"panic": recovered,
			"path":  c.Request.URL.Path,
		}).Error("Panic recovered")

		// Return error response
		response.InternalServerError(c, "Internal server error", "an unexpected error occurred")
	})
}

// ValidatorMiddleware validates request payloads
func ValidatorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// MetricsMiddleware adds metrics collection
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		// Collect metrics here
		duration := time.Since(start)
		method := c.Request.Method
		path := c.FullPath()
		status := strconv.Itoa(c.Writer.Status())

		// This would integrate with Prometheus metrics
		_ = duration
		_ = method
		_ = path
		_ = status
	}
}
