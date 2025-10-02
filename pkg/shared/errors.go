package shared

import (
	"errors"
	"fmt"
)

// Common error types for all services

var (
	// Order related errors
	ErrOrderNotFound       = errors.New("order not found")
	ErrOrderAlreadyExists  = errors.New("order already exists")
	ErrInvalidOrder        = errors.New("invalid order")
	ErrOrderAlreadyFilled  = errors.New("order already filled")
	ErrOrderAlreadyCancelled = errors.New("order already cancelled")
	ErrInsufficientBalance = errors.New("insufficient balance")

	// Trade related errors
	ErrTradeNotFound      = errors.New("trade not found")
	ErrInvalidTrade       = errors.New("invalid trade")
	ErrTradeExecutionFailed = errors.New("trade execution failed")

	// User related errors
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidUser       = errors.New("invalid user")
	ErrUnauthorized      = errors.New("unauthorized")

	// System related errors
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrDatabaseConnection = errors.New("database connection error")
	ErrCacheConnection    = errors.New("cache connection error")
	ErrInvalidConfiguration = errors.New("invalid configuration")
	ErrTimeout            = errors.New("operation timeout")

	// Validation errors
	ErrInvalidSymbol   = errors.New("invalid symbol")
	ErrInvalidPrice    = errors.New("invalid price")
	ErrInvalidQuantity = errors.New("invalid quantity")
	ErrInvalidSide     = errors.New("invalid order side")
	ErrInvalidType     = errors.New("invalid order type")
)

// BusinessError represents a business logic error
type BusinessError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *BusinessError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewBusinessError creates a new business error
func NewBusinessError(code, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
	}
}

// NewBusinessErrorWithDetails creates a new business error with details
func NewBusinessErrorWithDetails(code, message, details string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "validation errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("validation errors: %d fields failed validation", len(e.Errors))
}

// Add adds a validation error
func (e *ValidationErrors) Add(field, message string) {
	e.Errors = append(e.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// HasErrors returns true if there are validation errors
func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// ServiceError represents a service-level error
type ServiceError struct {
	Service   string `json:"service"`
	Operation string `json:"operation"`
	Message   string `json:"message"`
	Cause     error  `json:"cause,omitempty"`
}

func (e *ServiceError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s.%s: %s (caused by: %v)", e.Service, e.Operation, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s.%s: %s", e.Service, e.Operation, e.Message)
}

// Unwrap returns the underlying error
func (e *ServiceError) Unwrap() error {
	return e.Cause
}

// NewServiceError creates a new service error
func NewServiceError(service, operation, message string) *ServiceError {
	return &ServiceError{
		Service:   service,
		Operation: operation,
		Message:   message,
	}
}

// NewServiceErrorWithCause creates a new service error with a cause
func NewServiceErrorWithCause(service, operation, message string, cause error) *ServiceError {
	return &ServiceError{
		Service:   service,
		Operation: operation,
		Message:   message,
		Cause:     cause,
	}
}

// Error code constants
const (
	ErrCodeOrderNotFound         = "ORDER_NOT_FOUND"
	ErrCodeOrderInvalid          = "ORDER_INVALID"
	ErrCodeOrderAlreadyFilled    = "ORDER_ALREADY_FILLED"
	ErrCodeOrderAlreadyCancelled = "ORDER_ALREADY_CANCELLED"
	ErrCodeInsufficientBalance   = "INSUFFICIENT_BALANCE"
	ErrCodeTradeExecutionFailed  = "TRADE_EXECUTION_FAILED"
	ErrCodeUserNotFound          = "USER_NOT_FOUND"
	ErrCodeUnauthorized          = "UNAUTHORIZED"
	ErrCodeServiceUnavailable    = "SERVICE_UNAVAILABLE"
	ErrCodeValidationFailed      = "VALIDATION_FAILED"
	ErrCodeDatabaseError         = "DATABASE_ERROR"
	ErrCodeCacheError            = "CACHE_ERROR"
	ErrCodeTimeout               = "TIMEOUT"
)