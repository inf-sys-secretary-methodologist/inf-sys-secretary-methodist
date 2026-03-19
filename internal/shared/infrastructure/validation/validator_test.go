package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateStrongPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{
			name:     "valid strong password",
			password: "Password123",
			valid:    true,
		},
		{
			name:     "valid with special chars",
			password: "Pass123@#$",
			valid:    true,
		},
		{
			name:     "too short",
			password: "Pass12",
			valid:    false,
		},
		{
			name:     "no digit",
			password: "PasswordOnly",
			valid:    false,
		},
		{
			name:     "no letter",
			password: "12345678",
			valid:    false,
		},
		{
			name:     "empty string",
			password: "",
			valid:    false,
		},
		{
			name:     "exactly 8 chars with letter and digit",
			password: "Abcdefg1",
			valid:    true,
		},
		{
			name:     "7 chars",
			password: "Abcdef1",
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsStrongPassword(tt.password)
			assert.Equal(t, tt.valid, result, "Password: %s", tt.password)
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{
			name:  "valid email",
			email: "user@example.com",
			valid: true,
		},
		{
			name:  "valid with subdomain",
			email: "user@mail.example.com",
			valid: true,
		},
		{
			name:  "valid with plus",
			email: "user+tag@example.com",
			valid: true,
		},
		{
			name:  "valid with dots",
			email: "first.last@example.com",
			valid: true,
		},
		{
			name:  "valid with hyphen in domain",
			email: "user@my-domain.com",
			valid: true,
		},
		{
			name:  "invalid no @",
			email: "userexample.com",
			valid: false,
		},
		{
			name:  "invalid no domain",
			email: "user@",
			valid: false,
		},
		{
			name:  "invalid no TLD",
			email: "user@example",
			valid: false,
		},
		{
			name:  "empty string",
			email: "",
			valid: false,
		},
		{
			name:  "invalid with spaces",
			email: "user @example.com",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidEmail(tt.email)
			assert.Equal(t, tt.valid, result, "Email: %s", tt.email)
		})
	}
}

func TestValidateAlphanumeric(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{
			name:  "valid alphanumeric",
			input: "abc123",
			valid: true,
		},
		{
			name:  "valid only letters",
			input: "abcdef",
			valid: true,
		},
		{
			name:  "valid only digits",
			input: "123456",
			valid: true,
		},
		{
			name:  "invalid with spaces",
			input: "abc 123",
			valid: false,
		},
		{
			name:  "invalid with special chars",
			input: "abc@123",
			valid: false,
		},
		{
			name:  "empty string",
			input: "",
			valid: false,
		},
		{
			name:  "single character",
			input: "a",
			valid: true,
		},
		{
			name:  "uppercase letters",
			input: "ABC123",
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAlphanumeric(tt.input)
			assert.Equal(t, tt.valid, result, "Input: %s", tt.input)
		})
	}
}

func TestValidateNoSQLInjection(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{
			name:  "safe input",
			input: "normaltext",
			valid: true,
		},
		{
			name:  "SQL comment",
			input: "text--comment",
			valid: false,
		},
		{
			name:  "SQL union",
			input: "text UNION SELECT",
			valid: false,
		},
		{
			name:  "SQL select",
			input: "SELECT * FROM users",
			valid: false,
		},
		{
			name:  "SQL drop",
			input: "DROP TABLE users",
			valid: false,
		},
		{
			name:  "SQL semicolon",
			input: "text; DROP TABLE",
			valid: false,
		},
		{
			name:  "SQL exec",
			input: "EXEC sp_executesql",
			valid: false,
		},
		{
			name:  "SQL block comment start",
			input: "text /* comment",
			valid: false,
		},
		{
			name:  "SQL block comment end",
			input: "text */ rest",
			valid: false,
		},
		{
			name:  "SQL xp_ prefix",
			input: "xp_cmdshell",
			valid: false,
		},
		{
			name:  "SQL sp_ prefix",
			input: "sp_executesql",
			valid: false,
		},
		{
			name:  "SQL insert",
			input: "INSERT INTO table",
			valid: false,
		},
		{
			name:  "SQL update",
			input: "UPDATE table SET",
			valid: false,
		},
		{
			name:  "SQL delete",
			input: "DELETE FROM table",
			valid: false,
		},
		{
			name:  "SQL create",
			input: "CREATE TABLE",
			valid: false,
		},
		{
			name:  "SQL alter",
			input: "ALTER TABLE",
			valid: false,
		},
		{
			name:  "SQL truncate",
			input: "TRUNCATE TABLE",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateField(tt.input, "no_sql_injection")
			if tt.valid {
				assert.NoError(t, err, "Input should be valid: %s", tt.input)
			} else {
				assert.Error(t, err, "Input should be invalid: %s", tt.input)
			}
		})
	}
}

func TestValidateNoXSS(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{
			name:  "safe input",
			input: "normal text",
			valid: true,
		},
		{
			name:  "script tag",
			input: "<script>alert('xss')</script>",
			valid: false,
		},
		{
			name:  "javascript protocol",
			input: "javascript:alert('xss')",
			valid: false,
		},
		{
			name:  "onerror handler",
			input: "<img onerror='alert(1)'>",
			valid: false,
		},
		{
			name:  "iframe tag",
			input: "<iframe src='evil.com'></iframe>",
			valid: false,
		},
		{
			name:  "eval function",
			input: "eval('malicious code')",
			valid: false,
		},
		{
			name:  "alert function",
			input: "alert(document.cookie)",
			valid: false,
		},
		{
			name:  "onload handler",
			input: "<body onload='alert(1)'>",
			valid: false,
		},
		{
			name:  "onclick handler",
			input: "<div onclick='alert(1)'>",
			valid: false,
		},
		{
			name:  "onmouseover handler",
			input: "<div onmouseover='alert(1)'>",
			valid: false,
		},
		{
			name:  "object tag",
			input: "<object data='evil.swf'>",
			valid: false,
		},
		{
			name:  "embed tag",
			input: "<embed src='evil.swf'>",
			valid: false,
		},
		{
			name:  "document.cookie",
			input: "steal document.cookie",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateField(tt.input, "no_xss")
			if tt.valid {
				assert.NoError(t, err, "Input should be valid: %s", tt.input)
			} else {
				assert.Error(t, err, "Input should be invalid: %s", tt.input)
			}
		})
	}
}

func TestValidatorValidate(t *testing.T) {
	type TestStruct struct {
		Email    string `validate:"required,email,no_xss,no_sql_injection"`
		Password string `validate:"required,strong_password"`
		Username string `validate:"required,alphanum"`
	}

	validator := NewValidator()

	tests := []struct {
		name    string
		data    TestStruct
		wantErr bool
	}{
		{
			name: "valid struct",
			data: TestStruct{
				Email:    "user@example.com",
				Password: "Password123",
				Username: "user123",
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			data: TestStruct{
				Email:    "not-an-email",
				Password: "Password123",
				Username: "user123",
			},
			wantErr: true,
		},
		{
			name: "weak password",
			data: TestStruct{
				Email:    "user@example.com",
				Password: "weak",
				Username: "user123",
			},
			wantErr: true,
		},
		{
			name: "username with special chars",
			data: TestStruct{
				Email:    "user@example.com",
				Password: "Password123",
				Username: "user@123",
			},
			wantErr: true,
		},
		{
			name: "XSS in email",
			data: TestStruct{
				Email:    "user@example.com<script>alert(1)</script>",
				Password: "Password123",
				Username: "user123",
			},
			wantErr: true,
		},
		{
			name: "all fields empty",
			data: TestStruct{
				Email:    "",
				Password: "",
				Username: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatorValidate_ErrorIsValidationError(t *testing.T) {
	type TestStruct struct {
		Email string `validate:"required,email"`
	}

	v := NewValidator()
	err := v.Validate(TestStruct{Email: ""})
	assert.Error(t, err)

	valErr, ok := err.(*Error)
	assert.True(t, ok, "expected *Error type, got %T", err)
	if ok {
		assert.NotEmpty(t, valErr.Fields)
	}
}

func TestValidatorValidate_NonValidationError(t *testing.T) {
	v := NewValidator()
	// Passing a non-struct should return an error that is not a ValidationErrors
	err := v.Validate("not a struct")
	assert.Error(t, err)
}

func TestValidationError(t *testing.T) {
	valErr := &Error{
		Fields: map[string][]string{
			"email":    {"Invalid email format", "Email is required"},
			"password": {"Password is too weak"},
		},
	}

	errMsg := valErr.Error()
	assert.Contains(t, errMsg, "email:")
	assert.Contains(t, errMsg, "password:")
	assert.Contains(t, errMsg, "Invalid email format")
	assert.Contains(t, errMsg, "Password is too weak")
}

func TestValidationError_SingleField(t *testing.T) {
	valErr := &Error{
		Fields: map[string][]string{
			"name": {"This field is required"},
		},
	}

	errMsg := valErr.Error()
	assert.Contains(t, errMsg, "name: This field is required")
}

func TestValidationError_EmptyFields(t *testing.T) {
	valErr := &Error{
		Fields: map[string][]string{},
	}

	errMsg := valErr.Error()
	assert.Equal(t, "", errMsg)
}

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	assert.NotNil(t, v)
}

func TestValidateField_Required(t *testing.T) {
	v := NewValidator()

	err := v.ValidateField("", "required")
	assert.Error(t, err)

	err = v.ValidateField("value", "required")
	assert.NoError(t, err)
}

func TestValidateField_Min(t *testing.T) {
	v := NewValidator()

	err := v.ValidateField("ab", "min=3")
	assert.Error(t, err)

	err = v.ValidateField("abc", "min=3")
	assert.NoError(t, err)
}

func TestValidateField_Max(t *testing.T) {
	v := NewValidator()

	err := v.ValidateField("abcdef", "max=5")
	assert.Error(t, err)

	err = v.ValidateField("abc", "max=5")
	assert.NoError(t, err)
}

func TestGetValidationMessage_AllTags(t *testing.T) {
	type RequiredField struct {
		Val string `validate:"required"`
	}
	type MinField struct {
		Val string `validate:"min=3"`
	}
	type MaxField struct {
		Val string `validate:"max=2"`
	}
	type AlphanumField struct {
		Val string `validate:"alphanum"`
	}
	type AlphaField struct {
		Val string `validate:"alpha"`
	}
	type NumericField struct {
		Val string `validate:"numeric"`
	}
	type URLField struct {
		Val string `validate:"url"`
	}
	type EmailField struct {
		Val string `validate:"email"`
	}

	v := NewValidator()

	// Test required
	err := v.Validate(RequiredField{Val: ""})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "This field is required")

	// Test min
	err = v.Validate(MinField{Val: "ab"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Must be at least 3 characters")

	// Test max
	err = v.Validate(MaxField{Val: "abc"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Must be at most 2 characters")

	// Test alphanum
	err = v.Validate(AlphanumField{Val: "ab@c"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Must contain only alphanumeric characters")

	// Test alpha
	err = v.Validate(AlphaField{Val: "ab1"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Must contain only letters")

	// Test numeric
	err = v.Validate(NumericField{Val: "12a"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Must contain only numbers")

	// Test url
	err = v.Validate(URLField{Val: "not-a-url"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid URL format")

	// Test email
	err = v.Validate(EmailField{Val: "not-email"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid email format")
}

func TestGetValidationMessage_StrongPassword(t *testing.T) {
	type PasswordField struct {
		Val string `validate:"strong_password"`
	}

	v := NewValidator()
	err := v.Validate(PasswordField{Val: "weak"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Password must be at least 8 characters")
}

func TestGetValidationMessage_NoSQLInjection(t *testing.T) {
	type Field struct {
		Val string `validate:"no_sql_injection"`
	}

	v := NewValidator()
	err := v.Validate(Field{Val: "SELECT * FROM users"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "potentially dangerous SQL")
}

func TestGetValidationMessage_NoXSS(t *testing.T) {
	type Field struct {
		Val string `validate:"no_xss"`
	}

	v := NewValidator()
	err := v.Validate(Field{Val: "<script>alert(1)</script>"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "potentially dangerous HTML/JS")
}

func TestGetValidationMessage_UnknownTag(t *testing.T) {
	// We can test the default case by using a custom tag that is not in the switch
	type Field struct {
		Val string `validate:"len=5"`
	}

	v := NewValidator()
	err := v.Validate(Field{Val: "abc"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Validation failed on")
}

func TestValidateField_StrongPassword(t *testing.T) {
	v := NewValidator()

	err := v.ValidateField("Password1", "strong_password")
	assert.NoError(t, err)

	err = v.ValidateField("short1", "strong_password")
	assert.Error(t, err)

	err = v.ValidateField("nopdigits", "strong_password")
	assert.Error(t, err)

	err = v.ValidateField("12345678", "strong_password")
	assert.Error(t, err)
}
