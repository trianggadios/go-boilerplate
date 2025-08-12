package main

import (
	"boilerplate-go/config"
	"boilerplate-go/infrastructure/database"
	"boilerplate-go/infrastructure/logger"
	"boilerplate-go/infrastructure/metrics"
	"boilerplate-go/internal/delivery/http/handler"
	"boilerplate-go/internal/delivery/http/middleware"
	"boilerplate-go/internal/delivery/http/route"
	"boilerplate-go/internal/domain/repository"
	"boilerplate-go/internal/usecase/auth"
	"boilerplate-go/internal/usecase/user"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// @title           Boilerplate API
// @version         1.0
// @description     A professional Go microservice boilerplate with clean architecture
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  MIT
// @license.url   http://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize logger
	appLogger := logger.NewLogger()
	appLogger.WithFields(map[string]interface{}{
		"version": "1.0.0",
		"service": "boilerplate-api",
	}).Info("Starting application")

	// Initialize metrics
	appMetrics := metrics.NewMetrics()
	healthMetrics := metrics.NewHealthMetrics()

	// Initialize database connection
	db, err := database.NewPostgresConnection(cfg.Database)
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to connect to database")
	}
	defer func() {
		if err := db.Close(); err != nil {
			appLogger.WithError(err).Error("Failed to close database connection")
		}
	}()

	// Test database connection and update health metrics
	if err := db.Ping(); err != nil {
		appLogger.WithError(err).Error("Database health check failed")
		healthMetrics.SetDatabaseStatus(false)
	} else {
		appLogger.Info("Database connection healthy")
		healthMetrics.SetDatabaseStatus(true)
	}

	// Update database connection metrics
	stats := db.DB.Stats()
	appMetrics.SetDatabaseConnections(float64(stats.OpenConnections))

	// Initialize repositories with dependencies
	userRepo := repository.NewUserRepository(db, appLogger, appMetrics)

	// Initialize use cases
	authUsecase := auth.NewAuthUsecase(userRepo, cfg.JWT)
	userUsecase := user.NewUserUsecase(userRepo)

	// Initialize handlers with dependencies
	authHandler := handler.NewAuthHandler(authUsecase, appLogger, appMetrics)
	userHandler := handler.NewUserHandler(userUsecase, appLogger, appMetrics)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Setup professional middleware stack
	middlewareConfig := middleware.MiddlewareConfig{
		Logger:    appLogger,
		JWTSecret: cfg.JWT.SecretKey,
	}
	middleware.SetupMiddlewares(r, middlewareConfig)

	// Add metrics middleware
	r.Use(appMetrics.MetricsMiddleware())

	// Setup routes
	route.SetupRoutes(r, authHandler, userHandler, cfg.JWT.SecretKey)

	// Add metrics endpoint
	r.GET("/metrics", func(c *gin.Context) {
		appMetrics.Handler().ServeHTTP(c.Writer, c.Request)
	})

	// Enhanced health check endpoint
	r.GET("/health", func(c *gin.Context) {
		status := "ok"
		httpStatus := http.StatusOK

		if !healthMetrics.IsHealthy() {
			status = "unhealthy"
			httpStatus = http.StatusServiceUnavailable
		}

		c.JSON(httpStatus, map[string]interface{}{
			"status":    status,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"uptime":    healthMetrics.Uptime().String(),
			"version":   "1.0.0",
			"checks": map[string]interface{}{
				"database": healthMetrics.DatabaseUp,
			},
		})
	})

	// Readiness probe
	r.GET("/ready", func(c *gin.Context) {
		if healthMetrics.DatabaseUp {
			c.JSON(http.StatusOK, map[string]string{"status": "ready"})
		} else {
			c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
		}
	})

	// Liveness probe
	r.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]string{"status": "alive"})
	})

	// Configure HTTP server with production settings
	srv := &http.Server{
		Addr:           fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:        r,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
		IdleTimeout:    60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		appLogger.WithFields(map[string]interface{}{
			"addr":          srv.Addr,
			"read_timeout":  cfg.Server.ReadTimeout,
			"write_timeout": cfg.Server.WriteTimeout,
		}).Info("Starting HTTP server")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	appLogger.WithFields(map[string]interface{}{
		"signal": sig.String(),
	}).Info("Received shutdown signal, starting graceful shutdown")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		appLogger.WithError(err).Error("HTTP server forced to shutdown")
	} else {
		appLogger.Info("HTTP server shutdown completed")
	}

	appLogger.Info("Application shutdown completed")
}
