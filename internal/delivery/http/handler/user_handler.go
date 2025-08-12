package handler

import (
	"boilerplate-go/infrastructure/logger"
	"boilerplate-go/infrastructure/metrics"
	"boilerplate-go/internal/usecase/user"
	"boilerplate-go/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userUsecase *user.UserUsecase
	logger      *logger.Logger
	metrics     *metrics.Metrics
}

// NewUserHandler creates a new user handler
func NewUserHandler(userUsecase *user.UserUsecase, log *logger.Logger, m *metrics.Metrics) *UserHandler {
	return &UserHandler{
		userUsecase: userUsecase,
		logger:      log,
		metrics:     m,
	}
}

// GetProfile godoc
// @Summary      Get user profile
// @Description  Retrieve the authenticated user's profile information
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=entity.User}
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/user/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.WithContext(ctx).Warn("User ID not found in context")
		response.Unauthorized(c, "User not authenticated", "user_id not found in context")
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		h.logger.WithContext(ctx).WithFields(map[string]interface{}{
			"user_id_type": userID,
		}).Error("Invalid user ID type in context")
		response.InternalServerError(c, "Invalid user ID format", "user_id type assertion failed")
		return
	}

	// Log profile request
	h.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"user_id": userIDInt,
		"action":  "get_profile",
	}).Info("User profile requested")

	user, err := h.userUsecase.GetProfile(ctx, userIDInt)
	if err != nil {
		h.logger.ErrorLogger(ctx, err, "Failed to get user profile", map[string]interface{}{
			"user_id": userIDInt,
		})
		response.InternalServerError(c, "Failed to get user profile", err.Error())
		return
	}

	// Log successful profile retrieval
	h.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"user_id":  userIDInt,
		"username": user.Username,
		"action":   "get_profile_success",
	}).Info("User profile retrieved successfully")

	response.Success(c, http.StatusOK, "Profile retrieved successfully", user)
}
