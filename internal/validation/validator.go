// Package validation provides a shared validator instance for struct validation
// across the application. It uses go-playground/validator/v10 with custom
// validators registered for domain-specific types.
package validation

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	instance *validator.Validate
	once     sync.Once
)

// Get returns the singleton validator instance with all custom validators registered.
// This instance is thread-safe and should be used throughout the application.
func Get() *validator.Validate {
	once.Do(func() {
		instance = validator.New(validator.WithRequiredStructEnabled())
		registerCustomValidators(instance)
	})
	return instance
}

// Validate validates a struct using the shared validator instance.
// Returns nil if validation passes, or a validator.ValidationErrors if it fails.
func Validate(s any) error {
	return Get().Struct(s)
}

// registerCustomValidators adds domain-specific validators.
func registerCustomValidators(v *validator.Validate) {
	// Register custom validators for config enums
	_ = v.RegisterValidation("sslmode", validateSSLMode)
	_ = v.RegisterValidation("loglevel", validateLogLevel)
	_ = v.RegisterValidation("logformat", validateLogFormat)
}

// validateSSLMode validates PostgreSQL SSL modes.
// Valid values: disable, allow, prefer, require, verify-ca, verify-full.
func validateSSLMode(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
		return true
	default:
		return false
	}
}

// validateLogLevel validates log levels.
// Valid values: debug, info, warn, error.
func validateLogLevel(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}

// validateLogFormat validates log formats.
// Valid values: json, text.
func validateLogFormat(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case "json", "text":
		return true
	default:
		return false
	}
}
