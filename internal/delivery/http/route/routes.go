package route

import (
	"boilerplate-go/internal/delivery/http/handler"
	"boilerplate-go/internal/delivery/http/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(
	r *gin.Engine,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	secretKey string,
) {
	// API v1 routes
	api := r.Group("/api/v1")
	{
		// Authentication routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// User routes (protected)
		user := api.Group("/user")
		user.Use(middleware.AuthenticationMiddleware(secretKey))
		{
			user.GET("/profile", userHandler.GetProfile)
		}
	}
}
