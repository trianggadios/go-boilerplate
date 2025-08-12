package handler

import (
	"boilerplate-go/infrastructure/logger"
	"boilerplate-go/infrastructure/metrics"
	"boilerplate-go/internal/domain/entity"
	"boilerplate-go/internal/usecase/auth"
	"boilerplate-go/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authUsecase *auth.AuthUsecase
	logger      *logger.Logger
	metrics     *metrics.Metrics
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authUsecase *auth.AuthUsecase, log *logger.Logger, m *metrics.Metrics) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
		logger:      log,
		metrics:     m,
	}
}

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account with username, email and password
// @Tags         authentication
// @Accept       json
// @Produce      json
// @Param        request  body      entity.RegisterRequest  true  "Registration details"
// @Success      201      {object}  response.Response{data=entity.User}
// @Failure      400      {object}  response.Response
// @Failure      409      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	ctx := c.Request.Context()

	var req entity.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithContext(ctx).WithError(err).Warn("Invalid registration request payload")
		h.metrics.RecordAuthAttempt("register", false)
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Log registration attempt
	h.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"username": req.Username,
		"email":    req.Email,
		"action":   "register_attempt",
	}).Info("User registration attempt")

	user, err := h.authUsecase.Register(ctx, &req)
	if err != nil {
		h.logger.ErrorLogger(ctx, err, "Registration failed", map[string]interface{}{
			"username": req.Username,
			"email":    req.Email,
		})
		h.metrics.RecordAuthAttempt("register", false)

		// Return appropriate error based on error type
		if err.Error() == "user already exists" {
			response.Error(c, http.StatusConflict, "Registration failed", err.Error())
			return
		}
		response.BadRequest(c, "Registration failed", err.Error())
		return
	}

	// Log successful registration
	h.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"action":   "register_success",
	}).Info("User registered successfully")

	h.metrics.RecordAuthAttempt("register", true)
	response.Success(c, http.StatusCreated, "User registered successfully", user)
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user and return JWT token
// @Tags         authentication
// @Accept       json
// @Produce      json
// @Param        request  body      entity.LoginRequest  true  "Login credentials"
// @Success      200      {object}  response.Response{data=entity.LoginResponse}
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	var req entity.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithContext(ctx).WithError(err).Warn("Invalid login request payload")
		h.metrics.RecordAuthAttempt("login", false)
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Log login attempt (don't log password)
	h.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"username": req.Username,
		"action":   "login_attempt",
	}).Info("User login attempt")

	loginResponse, err := h.authUsecase.Login(ctx, &req)
	if err != nil {
		h.logger.ErrorLogger(ctx, err, "Login failed", map[string]interface{}{
			"username": req.Username,
		})
		h.metrics.RecordAuthAttempt("login", false)
		response.Unauthorized(c, "Login failed", err.Error())
		return
	}

	// Log successful login
	h.logger.WithContext(ctx).WithFields(map[string]interface{}{
		"user_id":  loginResponse.User.ID,
		"username": loginResponse.User.Username,
		"action":   "login_success",
	}).Info("User logged in successfully")

	h.metrics.RecordAuthAttempt("login", true)
	response.Success(c, http.StatusOK, "Login successful", loginResponse)
}
