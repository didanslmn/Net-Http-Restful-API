package validator

import (
	"reflect"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"postgresDB/internal/domain/errors"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
	once     sync.Once
)

// GetValidator returns the singleton validator instance
func GetValidator() *validator.Validate {
	once.Do(func() {
		validate = validator.New()

		// Use JSON tag names for field names
		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		// Register custom validators
		_ = validate.RegisterValidation("password", validatePassword)
	})
	return validate
}

// ValidateStruct validates a struct and returns AppError with details
func ValidateStruct(s interface{}) error {
	v := GetValidator()
	err := v.Struct(s)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return errors.WrapInternal(err)
	}

	details := make([]errors.ValidationError, 0, len(validationErrors))
	for _, e := range validationErrors {
		details = append(details, errors.ValidationError{
			Field:   e.Field(),
			Message: getErrorMessage(e),
		})
	}

	return errors.NewValidationError(details)
}

// validatePassword validates password strength
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Require at least uppercase, lowercase, and number
	return hasUpper && hasLower && hasNumber || hasSpecial
}

// getErrorMessage returns a user-friendly error message in Indonesian
func getErrorMessage(e validator.FieldError) string {
	field := e.Field()

	switch e.Tag() {
	case "required":
		return field + " wajib diisi"
	case "email":
		return "Format email tidak valid"
	case "min":
		return field + " minimal " + e.Param() + " karakter"
	case "max":
		return field + " maksimal " + e.Param() + " karakter"
	case "gt":
		return field + " harus lebih besar dari " + e.Param()
	case "gte":
		return field + " harus lebih besar atau sama dengan " + e.Param()
	case "lt":
		return field + " harus lebih kecil dari " + e.Param()
	case "lte":
		return field + " harus lebih kecil atau sama dengan " + e.Param()
	case "oneof":
		return field + " harus salah satu dari: " + e.Param()
	case "password":
		return "Password harus minimal 8 karakter dengan huruf besar, huruf kecil, dan angka"
	case "uuid":
		return field + " harus berformat UUID yang valid"
	default:
		return field + " tidak valid"
	}
}

// IsValidEmail validates email format
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidUUID validates UUID format
func IsValidUUID(u string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return uuidRegex.MatchString(u)
}
