package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, statusCode int, message string, err string) {
	c.JSON(statusCode, Response{
		Success: false,
		Message: message,
		Error:   err,
	})
}

func BadRequest(c *gin.Context, message string, err string) {
	Error(c, http.StatusBadRequest, message, err)
}

func Unauthorized(c *gin.Context, message string, err string) {
	Error(c, http.StatusUnauthorized, message, err)
}

func InternalServerError(c *gin.Context, message string, err string) {
	Error(c, http.StatusInternalServerError, message, err)
}
