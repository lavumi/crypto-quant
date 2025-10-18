package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the standard API response structure
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SuccessResponse sends a successful response
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// CreatedResponse sends a resource created response
func CreatedResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// ErrorResponse sends an error response
func ErrorResponse(c *gin.Context, statusCode int, code, message string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	})
}

// BadRequestResponse sends a 400 Bad Request response
func BadRequestResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// NotFoundResponse sends a 404 Not Found response
func NotFoundResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", message)
}

// InternalErrorResponse sends a 500 Internal Server Error response
func InternalErrorResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// ValidationErrorResponse sends a 422 Unprocessable Entity response
func ValidationErrorResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", message)
}

// Success is a shorthand for SuccessResponse
func Success(c *gin.Context, data interface{}) {
	SuccessResponse(c, data)
}

// Error is a shorthand for ErrorResponse
func Error(c *gin.Context, statusCode int, message string) {
	code := "ERROR"
	switch statusCode {
	case http.StatusBadRequest:
		code = "BAD_REQUEST"
	case http.StatusNotFound:
		code = "NOT_FOUND"
	case http.StatusInternalServerError:
		code = "INTERNAL_ERROR"
	case http.StatusUnprocessableEntity:
		code = "VALIDATION_ERROR"
	}
	ErrorResponse(c, statusCode, code, message)
}
