package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"simulated_exchange/internal/api/dto"
)

// ValidationMiddleware provides request validation
func ValidationMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next()
	})
}

// ValidateJSON is a utility function for validating JSON requests
func ValidateJSON(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		var errorMessage string
		var errorCode string

		// Handle validation errors
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorMessages := make([]string, 0)
			for _, validationError := range validationErrors {
				errorMessages = append(errorMessages, formatValidationError(validationError))
			}
			errorMessage = strings.Join(errorMessages, "; ")
			errorCode = "VALIDATION_ERROR"
		} else {
			errorMessage = "Invalid JSON format"
			errorCode = "INVALID_JSON"
		}

		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    errorCode,
				Message: errorMessage,
				Details: err.Error(),
			},
		})
		return false
	}
	return true
}

// formatValidationError converts validator errors to human-readable messages
func formatValidationError(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()

	switch tag {
	case "required":
		return field + " is required"
	case "min":
		return field + " must be at least " + err.Param() + " characters"
	case "max":
		return field + " must be at most " + err.Param() + " characters"
	case "gt":
		return field + " must be greater than " + err.Param()
	case "gte":
		return field + " must be greater than or equal to " + err.Param()
	case "lt":
		return field + " must be less than " + err.Param()
	case "lte":
		return field + " must be less than or equal to " + err.Param()
	case "oneof":
		return field + " must be one of: " + err.Param()
	default:
		return field + " is invalid"
	}
}

// ContentTypeMiddleware ensures proper content type for API requests
func ContentTypeMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Only check content type for POST, PUT, PATCH requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				c.JSON(http.StatusUnsupportedMediaType, dto.APIResponse{
					Success: false,
					Error: &dto.APIError{
						Code:    "UNSUPPORTED_MEDIA_TYPE",
						Message: "Content-Type must be application/json",
					},
				})
				c.Abort()
				return
			}
		}
		c.Next()
	})
}