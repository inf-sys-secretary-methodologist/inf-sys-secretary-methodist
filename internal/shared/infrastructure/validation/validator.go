package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	// emailRegex for email validation
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// passwordRegex для базовой проверки (дополнительная логика в функции)
	passwordRegex = regexp.MustCompile(`[A-Za-z]`) // contains letter
	passwordDigitRegex = regexp.MustCompile(`\d`)    // contains digit

	// alphanumericRegex для проверки alphanumeric строк
	alphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
)

// Validator wraps go-playground validator with custom validation rules
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new validator instance with custom rules
func NewValidator() *Validator {
	v := validator.New()

	// Register custom validations
	_ = v.RegisterValidation("strong_password", validateStrongPassword)
	_ = v.RegisterValidation("no_sql_injection", validateNoSQLInjection)
	_ = v.RegisterValidation("no_xss", validateNoXSS)

	return &Validator{validate: v}
}

// Validate validates a struct
func (v *Validator) Validate(data interface{}) error {
	if err := v.validate.Struct(data); err != nil {
		return v.formatValidationError(err)
	}
	return nil
}

// ValidateField validates a single field
func (v *Validator) ValidateField(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}

// formatValidationError форматирует ошибки валидации в читаемый вид
func (v *Validator) formatValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errorMessages := make(map[string][]string)

		for _, fieldError := range validationErrors {
			field := strings.ToLower(fieldError.Field())
			message := getValidationMessage(fieldError)
			errorMessages[field] = append(errorMessages[field], message)
		}

		return &ValidationError{
			Fields: errorMessages,
		}
	}
	return err
}

// ValidationError представляет ошибки валидации
type ValidationError struct {
	Fields map[string][]string
}

func (e *ValidationError) Error() string {
	var messages []string
	for field, errs := range e.Fields {
		for _, err := range errs {
			messages = append(messages, fmt.Sprintf("%s: %s", field, err))
		}
	}
	return strings.Join(messages, "; ")
}

// getValidationMessage возвращает человекочитаемое сообщение для ошибки валидации
func getValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("Must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("Must be at most %s characters", fe.Param())
	case "strong_password":
		return "Password must be at least 8 characters and contain at least one letter and one number"
	case "no_sql_injection":
		return "Input contains potentially dangerous SQL characters"
	case "no_xss":
		return "Input contains potentially dangerous HTML/JS characters"
	case "alphanum":
		return "Must contain only alphanumeric characters"
	case "alpha":
		return "Must contain only letters"
	case "numeric":
		return "Must contain only numbers"
	case "url":
		return "Invalid URL format"
	default:
		return fmt.Sprintf("Validation failed on '%s'", fe.Tag())
	}
}

// Custom validation functions

func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 {
		return false
	}
	// Check for at least one letter and one digit
	hasLetter := passwordRegex.MatchString(password)
	hasDigit := passwordDigitRegex.MatchString(password)
	return hasLetter && hasDigit
}

func validateNoSQLInjection(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	// Проверяем на потенциально опасные SQL символы/ключевые слова
	dangerousPatterns := []string{
		"--", "/*", "*/", "xp_", "sp_", "exec", "execute",
		"union", "select", "insert", "update", "delete", "drop",
		"create", "alter", "truncate", ";",
	}

	lowerValue := strings.ToLower(value)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerValue, pattern) {
			return false
		}
	}

	return true
}

func validateNoXSS(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	// Проверяем на потенциально опасные HTML/JS символы
	dangerousPatterns := []string{
		"<script", "</script>", "javascript:", "onerror=", "onload=",
		"<iframe", "</iframe>", "onclick=", "onmouseover=", "<object",
		"<embed", "eval(", "alert(", "document.cookie",
	}

	lowerValue := strings.ToLower(value)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerValue, pattern) {
			return false
		}
	}

	return true
}

// Helper validation functions

// IsValidEmail проверяет валидность email
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// IsStrongPassword проверяет силу пароля
func IsStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasLetter := passwordRegex.MatchString(password)
	hasDigit := passwordDigitRegex.MatchString(password)
	return hasLetter && hasDigit
}

// IsAlphanumeric проверяет, что строка содержит только буквы и цифры
func IsAlphanumeric(s string) bool {
	return alphanumericRegex.MatchString(s)
}
