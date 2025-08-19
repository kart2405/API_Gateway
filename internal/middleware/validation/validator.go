package validation

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kart2405/API_Gateway/internal/middleware/logging"
	"go.uber.org/zap"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationResult represents validation result
type ValidationResult struct {
	Valid   bool              `json:"valid"`
	Errors  []ValidationError `json:"errors,omitempty"`
	Message string            `json:"message,omitempty"`
}

// Validator interface for validation
type Validator interface {
	Validate(value string) (bool, string)
}

// RequiredValidator validates required fields
type RequiredValidator struct{}

func (r *RequiredValidator) Validate(value string) (bool, string) {
	if strings.TrimSpace(value) == "" {
		return false, "Field is required"
	}
	return true, ""
}

// EmailValidator validates email format
type EmailValidator struct{}

func (e *EmailValidator) Validate(value string) (bool, string) {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(value) {
		return false, "Invalid email format"
	}
	return true, ""
}

// URLValidator validates URL format
type URLValidator struct{}

func (u *URLValidator) Validate(value string) (bool, string) {
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(value) {
		return false, "Invalid URL format"
	}
	return true, ""
}

// LengthValidator validates string length
type LengthValidator struct {
	Min int
	Max int
}

func (l *LengthValidator) Validate(value string) (bool, string) {
	length := len(value)
	if l.Min > 0 && length < l.Min {
		return false, fmt.Sprintf("Minimum length is %d characters", l.Min)
	}
	if l.Max > 0 && length > l.Max {
		return false, fmt.Sprintf("Maximum length is %d characters", l.Max)
	}
	return true, ""
}

// AlphanumericValidator validates alphanumeric characters
type AlphanumericValidator struct{}

func (a *AlphanumericValidator) Validate(value string) (bool, string) {
	alphanumericRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !alphanumericRegex.MatchString(value) {
		return false, "Only alphanumeric characters are allowed"
	}
	return true, ""
}

// ValidationRules defines validation rules for different fields
type ValidationRules map[string][]Validator

// ValidationService manages validation
type ValidationService struct {
	rules ValidationRules
}

// NewValidationService creates a new validation service
func NewValidationService() *ValidationService {
	return &ValidationService{
		rules: make(ValidationRules),
	}
}

// AddRule adds a validation rule for a field
func (vs *ValidationService) AddRule(field string, validators ...Validator) {
	vs.rules[field] = validators
}

// Validate validates a map of fields
func (vs *ValidationService) Validate(fields map[string]string) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	for field, value := range fields {
		if validators, exists := vs.rules[field]; exists {
			for _, validator := range validators {
				if valid, message := validator.Validate(value); !valid {
					result.Valid = false
					result.Errors = append(result.Errors, ValidationError{
						Field:   field,
						Message: message,
						Value:   value,
					})
					break // Stop validating this field after first error
				}
			}
		}
	}

	if !result.Valid {
		result.Message = "Validation failed"
	}

	return result
}

// SanitizeString sanitizes a string input
func SanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	// Remove control characters except newline and tab
	re := regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	input = re.ReplaceAllString(input, "")

	return input
}

// SanitizeMap sanitizes all string values in a map
func SanitizeMap(data map[string]string) map[string]string {
	sanitized := make(map[string]string)
	for key, value := range data {
		sanitized[key] = SanitizeString(value)
	}
	return sanitized
}

// ValidationMiddleware provides request validation
func ValidationMiddleware(validationService *ValidationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		queryParams := make(map[string]string)
		for key, values := range c.Request.URL.Query() {
			if len(values) > 0 {
				queryParams[key] = SanitizeString(values[0])
			}
		}

		// Validate query parameters if rules exist
		if len(validationService.rules) > 0 {
			result := validationService.Validate(queryParams)
			if !result.Valid {
				logger := logging.GetLoggerFromContext(c)
				logger.Warn("Query parameter validation failed",
					zap.Any("errors", result.Errors),
					zap.String("ip", c.ClientIP()),
				)

				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid query parameters",
					"details": result.Errors,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// BodyValidationMiddleware validates request body
func BodyValidationMiddleware(validationService *ValidationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only validate POST/PUT/PATCH requests
		if c.Request.Method == "GET" || c.Request.Method == "DELETE" {
			c.Next()
			return
		}

		// Parse JSON body
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			logger := logging.GetLoggerFromContext(c)
			logger.Warn("Invalid JSON body",
				zap.Error(err),
				zap.String("ip", c.ClientIP()),
			)

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid JSON body",
			})
			c.Abort()
			return
		}

		// Convert body to string map for validation
		stringBody := make(map[string]string)
		for key, value := range body {
			if str, ok := value.(string); ok {
				stringBody[key] = SanitizeString(str)
			} else {
				stringBody[key] = fmt.Sprintf("%v", value)
			}
		}

		// Validate body if rules exist
		if len(validationService.rules) > 0 {
			result := validationService.Validate(stringBody)
			if !result.Valid {
				logger := logging.GetLoggerFromContext(c)
				logger.Warn("Request body validation failed",
					zap.Any("errors", result.Errors),
					zap.String("ip", c.ClientIP()),
				)

				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid request body",
					"details": result.Errors,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// Common validation rules
var CommonValidationRules = map[string][]Validator{
	"email": {
		&RequiredValidator{},
		&EmailValidator{},
	},
	"username": {
		&RequiredValidator{},
		&LengthValidator{Min: 3, Max: 50},
		&AlphanumericValidator{},
	},
	"password": {
		&RequiredValidator{},
		&LengthValidator{Min: 8, Max: 128},
	},
	"service_name": {
		&RequiredValidator{},
		&LengthValidator{Min: 1, Max: 100},
	},
	"backend_url": {
		&RequiredValidator{},
		&URLValidator{},
	},
}

// CreateCommonValidationService creates a validation service with common rules
func CreateCommonValidationService() *ValidationService {
	service := NewValidationService()
	for field, validators := range CommonValidationRules {
		service.AddRule(field, validators...)
	}
	return service
}
