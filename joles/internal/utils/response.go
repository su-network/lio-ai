package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"lio-ai/internal/models"
)

// SuccessResponse sends a successful API response
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    data,
	})
}

// SuccessResponseWithMeta sends a successful API response with metadata
func SuccessResponseWithMeta(c *gin.Context, data interface{}, meta *models.Meta) {
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// CreatedResponse sends a 201 Created response
func CreatedResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    data,
	})
}

// ErrorResponse sends an error API response
func ErrorResponse(c *gin.Context, statusCode int, code, message string) {
	c.JSON(statusCode, models.APIResponse{
		Success: false,
		Error: &models.APIError{
			Code:    code,
			Message: message,
		},
	})
}

// ErrorResponseWithDetails sends an error API response with details
func ErrorResponseWithDetails(c *gin.Context, statusCode int, code, message, details string) {
	c.JSON(statusCode, models.APIResponse{
		Success: false,
		Error: &models.APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// ValidationError sends a validation error response
func ValidationError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusBadRequest, models.ErrCodeValidation, message)
}

// NotFoundError sends a not found error response
func NotFoundError(c *gin.Context, resource string) {
	ErrorResponse(c, http.StatusNotFound, models.ErrCodeNotFound, resource+" not found")
}

// UnauthorizedError sends an unauthorized error response
func UnauthorizedError(c *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized access"
	}
	ErrorResponse(c, http.StatusUnauthorized, models.ErrCodeUnauthorized, message)
}

// ForbiddenError sends a forbidden error response
func ForbiddenError(c *gin.Context, message string) {
	if message == "" {
		message = "Access forbidden"
	}
	ErrorResponse(c, http.StatusForbidden, models.ErrCodeForbidden, message)
}

// QuotaExceededError sends a quota exceeded error response
func QuotaExceededError(c *gin.Context, message string) {
	if message == "" {
		message = "Quota exceeded"
	}
	ErrorResponse(c, http.StatusTooManyRequests, models.ErrCodeQuotaExceeded, message)
}

// RateLimitError sends a rate limit error response
func RateLimitError(c *gin.Context) {
	ErrorResponse(c, http.StatusTooManyRequests, models.ErrCodeRateLimited, "Rate limit exceeded")
}

// InternalError sends an internal server error response
func InternalError(c *gin.Context, message string) {
	if message == "" {
		message = "Internal server error"
	}
	ErrorResponse(c, http.StatusInternalServerError, models.ErrCodeInternal, message)
}

// BadRequestError sends a bad request error response
func BadRequestError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusBadRequest, models.ErrCodeBadRequest, message)
}

// ServiceDownError sends a service unavailable error response
func ServiceDownError(c *gin.Context, service string) {
	ErrorResponse(c, http.StatusServiceUnavailable, models.ErrCodeServiceDown, service+" service is unavailable")
}
