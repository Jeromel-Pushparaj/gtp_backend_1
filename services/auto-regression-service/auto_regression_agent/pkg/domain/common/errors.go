package common

import "fmt"

// ErrorCode represents a domain error code
type ErrorCode string

const (
	// Validation errors
	ErrCodeValidation        ErrorCode = "VALIDATION_ERROR"
	ErrCodeInvalidInput      ErrorCode = "INVALID_INPUT"
	ErrCodeMissingRequired   ErrorCode = "MISSING_REQUIRED_FIELD"
	
	// Resource errors
	ErrCodeNotFound          ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists     ErrorCode = "ALREADY_EXISTS"
	ErrCodeConflict          ErrorCode = "CONFLICT"
	
	// Authorization errors
	ErrCodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden         ErrorCode = "FORBIDDEN"
	
	// System errors
	ErrCodeInternal          ErrorCode = "INTERNAL_ERROR"
	ErrCodeTimeout           ErrorCode = "TIMEOUT"
	ErrCodeUnavailable       ErrorCode = "UNAVAILABLE"
	
	// Storage errors
	ErrCodeStorageFailure    ErrorCode = "STORAGE_FAILURE"
	ErrCodeQueueFailure      ErrorCode = "QUEUE_FAILURE"
	
	// Spec errors
	ErrCodeInvalidSpec       ErrorCode = "INVALID_SPEC"
	ErrCodeSpecParseFailed   ErrorCode = "SPEC_PARSE_FAILED"
)

// DomainError represents a domain-specific error
type DomainError struct {
	Code    ErrorCode
	Message string
	Cause   error
	Details map[string]interface{}
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *DomainError) Unwrap() error {
	return e.Cause
}

// NewError creates a new domain error
func NewError(code ErrorCode, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

// NewErrorWithCause creates a new domain error with a cause
func NewErrorWithCause(code ErrorCode, message string, cause error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// NewErrorWithDetails creates a new domain error with details
func NewErrorWithDetails(code ErrorCode, message string, details map[string]interface{}) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Common error constructors
func ErrNotFound(resource string, id string) *DomainError {
	return NewErrorWithDetails(
		ErrCodeNotFound,
		fmt.Sprintf("%s not found", resource),
		map[string]interface{}{"resource": resource, "id": id},
	)
}

func ErrAlreadyExists(resource string, id string) *DomainError {
	return NewErrorWithDetails(
		ErrCodeAlreadyExists,
		fmt.Sprintf("%s already exists", resource),
		map[string]interface{}{"resource": resource, "id": id},
	)
}

func ErrValidation(field string, reason string) *DomainError {
	return NewErrorWithDetails(
		ErrCodeValidation,
		fmt.Sprintf("validation failed for field '%s': %s", field, reason),
		map[string]interface{}{"field": field, "reason": reason},
	)
}

func ErrInternal(message string, cause error) *DomainError {
	return NewErrorWithCause(ErrCodeInternal, message, cause)
}

