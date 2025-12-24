package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode represents the error code type
type ErrorCode string

const (
	CodeValidation   ErrorCode = "VALIDATION_ERROR"
	CodeUnauthorized ErrorCode = "UNAUTHORIZED"
	CodeForbidden    ErrorCode = "FORBIDDEN"
	CodeNotFound     ErrorCode = "NOT_FOUND"
	CodeConflict     ErrorCode = "CONFLICT"
	CodeInternal     ErrorCode = "INTERNAL_ERROR"
	CodeBadRequest   ErrorCode = "BAD_REQUEST"
	CodeTooMany      ErrorCode = "TOO_MANY_REQUESTS"
)

// AppError represents a custom application error
type AppError struct {
	Code       ErrorCode         `json:"code"`
	Message    string            `json:"message"`
	Details    []ValidationError `json:"details,omitempty"`
	HTTPStatus int               `json:"-"`
	Err        error             `json:"-"`
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// WithError wraps an underlying error
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// WithDetails adds validation details
func (e *AppError) WithDetails(details []ValidationError) *AppError {
	e.Details = details
	return e
}

// Predefined errors
var (
	ErrValidation = &AppError{
		Code:       CodeValidation,
		Message:    "Validasi gagal",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrUnauthorized = &AppError{
		Code:       CodeUnauthorized,
		Message:    "Autentikasi diperlukan",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrInvalidCredentials = &AppError{
		Code:       CodeUnauthorized,
		Message:    "Email atau password salah",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrInvalidToken = &AppError{
		Code:       CodeUnauthorized,
		Message:    "Token tidak valid atau sudah kadaluarsa",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrTokenRevoked = &AppError{
		Code:       CodeUnauthorized,
		Message:    "Token sudah dicabut",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrTokenReused = &AppError{
		Code:       CodeUnauthorized,
		Message:    "Refresh token sudah digunakan, semua sesi telah dicabut",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrForbidden = &AppError{
		Code:       CodeForbidden,
		Message:    "Akses ditolak",
		HTTPStatus: http.StatusForbidden,
	}

	ErrUserNotFound = &AppError{
		Code:       CodeNotFound,
		Message:    "User tidak ditemukan",
		HTTPStatus: http.StatusNotFound,
	}

	ErrUserAlreadyExists = &AppError{
		Code:       CodeConflict,
		Message:    "User sudah terdaftar",
		HTTPStatus: http.StatusConflict,
	}

	ErrUserInactive = &AppError{
		Code:       CodeForbidden,
		Message:    "User tidak aktif",
		HTTPStatus: http.StatusForbidden,
	}

	ErrPasswordMismatch = &AppError{
		Code:       CodeBadRequest,
		Message:    "Password lama tidak sesuai",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrPasswordTooWeak = &AppError{
		Code:       CodeBadRequest,
		Message:    "Password terlalu lemah",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrProductNotFound = &AppError{
		Code:       CodeNotFound,
		Message:    "Produk tidak ditemukan",
		HTTPStatus: http.StatusNotFound,
	}

	ErrOrderNotFound = &AppError{
		Code:       CodeNotFound,
		Message:    "Order tidak ditemukan",
		HTTPStatus: http.StatusNotFound,
	}

	ErrEmailExists = &AppError{
		Code:       CodeConflict,
		Message:    "Email sudah terdaftar",
		HTTPStatus: http.StatusConflict,
	}

	ErrInsufficientStock = &AppError{
		Code:       CodeBadRequest,
		Message:    "Stok tidak mencukupi",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidStatusTransition = &AppError{
		Code:       CodeBadRequest,
		Message:    "Transisi status tidak valid",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInternal = &AppError{
		Code:       CodeInternal,
		Message:    "Terjadi kesalahan internal",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrTooManyRequests = &AppError{
		Code:       CodeTooMany,
		Message:    "Terlalu banyak permintaan, coba lagi nanti",
		HTTPStatus: http.StatusTooManyRequests,
	}
)

// IsAppError checks if the error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// AsAppError converts an error to AppError
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// WrapInternal wraps an error as internal error
func WrapInternal(err error) *AppError {
	return &AppError{
		Code:       CodeInternal,
		Message:    "Terjadi kesalahan internal",
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewValidationError creates a validation error with details
func NewValidationError(details []ValidationError) *AppError {
	return &AppError{
		Code:       CodeValidation,
		Message:    "Validasi gagal",
		HTTPStatus: http.StatusBadRequest,
		Details:    details,
	}
}

// NewNotFoundError creates a not found error with custom message
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:       CodeNotFound,
		Message:    fmt.Sprintf("%s tidak ditemukan", resource),
		HTTPStatus: http.StatusNotFound,
	}
}
