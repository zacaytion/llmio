package validation

import (
	"testing"
)

func TestGet_ReturnsSameInstance(t *testing.T) {
	v1 := Get()
	v2 := Get()

	if v1 != v2 {
		t.Error("Get() should return the same singleton instance")
	}
}

func TestValidate_ValidStruct(t *testing.T) {
	type sample struct {
		Name string `validate:"required"`
		Age  int    `validate:"min=0,max=150"`
	}

	s := sample{Name: "test", Age: 25}
	if err := Validate(s); err != nil {
		t.Errorf("Validate() should pass for valid struct, got error: %v", err)
	}
}

func TestValidate_InvalidStruct(t *testing.T) {
	type sample struct {
		Name string `validate:"required"`
	}

	s := sample{Name: ""}
	if err := Validate(s); err == nil {
		t.Error("Validate() should fail for empty required field")
	}
}

func TestValidateSSLMode(t *testing.T) {
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

func TestValidateLogLevel(t *testing.T) {
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

func TestValidateLogFormat(t *testing.T) {
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
