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
// Panics if registration fails, as this indicates a programming error
// that should be caught at startup rather than causing cryptic runtime failures.
func registerCustomValidators(v *validator.Validate) {
	// Register custom validators for config enums.
	// These must succeed - failure indicates duplicate names or nil validators.
	mustRegister(v, "sslmode", validateSSLMode)
	mustRegister(v, "loglevel", validateLogLevel)
	mustRegister(v, "logformat", validateLogFormat)
}

// mustRegister registers a validator and panics on failure.
// This ensures misconfigurations are caught at startup.
func mustRegister(v *validator.Validate, tag string, fn validator.Func) {
	if err := v.RegisterValidation(tag, fn); err != nil {
		panic("failed to register validator " + tag + ": " + err.Error())
	}
}

// validateSSLMode validates PostgreSQL SSL modes.
// Valid values: disable, allow, prefer, require, verify-ca, verify-full.
// Delegates to config.SSLMode.Valid() to avoid duplicating the switch statement.
func validateSSLMode(fl validator.FieldLevel) bool {
	// Import cycle prevention: we can't import config here, so we duplicate
	// the valid values. This is acceptable because SSLMode values are defined
	// by PostgreSQL and change extremely rarely.
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
// Note: Validation is duplicated here rather than calling config.LogLevel.Valid()
// to avoid an import cycle (config imports validation).
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
// Note: Validation is duplicated here rather than calling config.LogFormat.Valid()
// to avoid an import cycle (config imports validation).
func validateLogFormat(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case "json", "text":
		return true
	default:
		return false
	}
}
