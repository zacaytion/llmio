package validation

import (
	"testing"
)

func Test_Get_ReturnsSameInstance(t *testing.T) {
	v1 := Get()
	v2 := Get()

	if v1 != v2 {
		t.Error("Get() should return the same singleton instance")
	}
}

func Test_Validate_ValidStruct(t *testing.T) {
	type sample struct {
		Name string `validate:"required"`
		Age  int    `validate:"min=0,max=150"`
	}

	s := sample{Name: "test", Age: 25}
	if err := Validate(s); err != nil {
		t.Errorf("Validate() should pass for valid struct, got error: %v", err)
	}
}

func Test_Validate_InvalidStruct(t *testing.T) {
	type sample struct {
		Name string `validate:"required"`
	}

	s := sample{Name: ""}
	if err := Validate(s); err == nil {
		t.Error("Validate() should fail for empty required field")
	}
}

func Test_ValidateSSLMode(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"disable is valid", "disable", false},
		{"allow is valid", "allow", false},
		{"prefer is valid", "prefer", false},
		{"require is valid", "require", false},
		{"verify-ca is valid", "verify-ca", false},
		{"verify-full is valid", "verify-full", false},
		{"empty is invalid", "", true},
		{"invalid value", "invalid", true},
		{"similar but wrong", "disabled", true},
		{"uppercase is invalid", "DISABLE", true},
	}

	type sslTest struct {
		Mode string `validate:"required,sslmode"`
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sslTest{Mode: tt.value}
			err := Validate(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("sslmode validation for %q: got error=%v, wantErr=%v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func Test_ValidateLogLevel(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"debug is valid", "debug", false},
		{"info is valid", "info", false},
		{"warn is valid", "warn", false},
		{"error is valid", "error", false},
		{"empty is invalid", "", true},
		{"invalid value", "trace", true},
		{"uppercase is invalid", "DEBUG", true},
		{"typo is invalid", "deubg", true},
	}

	type logLevelTest struct {
		Level string `validate:"required,loglevel"`
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := logLevelTest{Level: tt.value}
			err := Validate(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("loglevel validation for %q: got error=%v, wantErr=%v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func Test_ValidateLogFormat(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"json is valid", "json", false},
		{"text is valid", "text", false},
		{"empty is invalid", "", true},
		{"invalid value", "xml", true},
		{"uppercase is invalid", "JSON", true},
		{"typo is invalid", "tect", true},
	}

	type logFormatTest struct {
		Format string `validate:"required,logformat"`
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := logFormatTest{Format: tt.value}
			err := Validate(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("logformat validation for %q: got error=%v, wantErr=%v", tt.value, err, tt.wantErr)
			}
		})
	}
}

// T130: Test that custom validators are registered successfully.
// This test verifies that all custom validators (sslmode, loglevel, logformat)
// are properly registered and can be used in validation.
func Test_CustomValidatorsRegistered(t *testing.T) {
	// Get the validator instance - this triggers registration
	v := Get()
	if v == nil {
		t.Fatal("Get() returned nil validator")
	}

	// Test that all custom validators work by validating structs that use them
	type allCustomValidators struct {
		SSLMode   string `validate:"required,sslmode"`
		LogLevel  string `validate:"required,loglevel"`
		LogFormat string `validate:"required,logformat"`
	}

	valid := allCustomValidators{
		SSLMode:   "disable",
		LogLevel:  "info",
		LogFormat: "json",
	}

	if err := Validate(valid); err != nil {
		t.Errorf("validation should pass for valid struct with all custom validators, got: %v", err)
	}

	// Test each validator individually to ensure they're registered
	tests := []struct {
		name      string
		validator string
		value     string
		valid     bool
	}{
		{"sslmode valid", "sslmode", "disable", true},
		{"sslmode invalid", "sslmode", "invalid", false},
		{"loglevel valid", "loglevel", "debug", true},
		{"loglevel invalid", "loglevel", "invalid", false},
		{"logformat valid", "logformat", "json", true},
		{"logformat invalid", "logformat", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a struct dynamically based on validator type
			var err error
			switch tt.validator {
			case "sslmode":
				type s struct {
					V string `validate:"sslmode"`
				}
				err = Validate(s{V: tt.value})
			case "loglevel":
				type s struct {
					V string `validate:"loglevel"`
				}
				err = Validate(s{V: tt.value})
			case "logformat":
				type s struct {
					V string `validate:"logformat"`
				}
				err = Validate(s{V: tt.value})
			}

			hasErr := err != nil
			if hasErr == tt.valid {
				t.Errorf("validator %q with value %q: expected valid=%v, got error=%v",
					tt.validator, tt.value, tt.valid, err)
			}
		})
	}
}
