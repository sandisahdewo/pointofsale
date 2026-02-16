package utils

import (
	"testing"
)

func TestValidateEmail_ValidFormats_ReturnsTrue(t *testing.T) {
	validEmails := []string{
		"test@example.com",
		"user.name@example.com",
		"user+tag@example.co.uk",
		"user_name@example-domain.com",
		"123@example.com",
		"a@b.co",
	}

	for _, email := range validEmails {
		t.Run(email, func(t *testing.T) {
			if !ValidateEmail(email) {
				t.Errorf("expected %s to be valid, but got false", email)
			}
		})
	}
}

func TestValidateEmail_InvalidFormats_ReturnsFalse(t *testing.T) {
	invalidEmails := []string{
		"",
		"invalid",
		"@example.com",
		"user@",
		"user @example.com",
		"user@example",
		"user@@example.com",
		"user@.com",
		"user@domain..com",
	}

	for _, email := range invalidEmails {
		t.Run(email, func(t *testing.T) {
			if ValidateEmail(email) {
				t.Errorf("expected %s to be invalid, but got true", email)
			}
		})
	}
}

func TestValidatePassword_StrongPassword_ReturnsNoErrors(t *testing.T) {
	strongPasswords := []string{
		"SecurePass123!",
		"MyP@ssw0rd",
		"C0mpl3x!Pass",
		"T3st@Password",
	}

	for _, password := range strongPasswords {
		t.Run(password, func(t *testing.T) {
			errors := ValidatePassword(password)
			if len(errors) > 0 {
				t.Errorf("expected no errors for strong password %s, got: %v", password, errors)
			}
		})
	}
}

func TestValidatePassword_TooShort_ReturnsError(t *testing.T) {
	shortPassword := "Short1!"
	errors := ValidatePassword(shortPassword)

	if len(errors) == 0 {
		t.Fatal("expected error for too short password, got none")
	}

	found := false
	for _, err := range errors {
		if err == "Password must be at least 8 characters" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Password must be at least 8 characters' error, got: %v", errors)
	}
}

func TestValidatePassword_NoUppercase_ReturnsError(t *testing.T) {
	password := "password123!"
	errors := ValidatePassword(password)

	if len(errors) == 0 {
		t.Fatal("expected error for no uppercase, got none")
	}

	found := false
	for _, err := range errors {
		if err == "Password must contain at least one uppercase letter" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected uppercase letter error, got: %v", errors)
	}
}

func TestValidatePassword_NoLowercase_ReturnsError(t *testing.T) {
	password := "PASSWORD123!"
	errors := ValidatePassword(password)

	if len(errors) == 0 {
		t.Fatal("expected error for no lowercase, got none")
	}

	found := false
	for _, err := range errors {
		if err == "Password must contain at least one lowercase letter" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected lowercase letter error, got: %v", errors)
	}
}

func TestValidatePassword_NoDigit_ReturnsError(t *testing.T) {
	password := "PasswordTest!"
	errors := ValidatePassword(password)

	if len(errors) == 0 {
		t.Fatal("expected error for no digit, got none")
	}

	found := false
	for _, err := range errors {
		if err == "Password must contain at least one digit" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected digit error, got: %v", errors)
	}
}

func TestValidatePassword_NoSpecialChar_ReturnsError(t *testing.T) {
	password := "Password123"
	errors := ValidatePassword(password)

	if len(errors) == 0 {
		t.Fatal("expected error for no special character, got none")
	}

	found := false
	for _, err := range errors {
		if err == "Password must contain at least one special character" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected special character error, got: %v", errors)
	}
}

func TestValidatePassword_MultipleErrors_ReturnsAll(t *testing.T) {
	password := "weak" // Too short, no uppercase, no digit, no special char
	errors := ValidatePassword(password)

	// Should have at least 4 errors
	if len(errors) < 4 {
		t.Errorf("expected at least 4 errors for very weak password, got %d: %v", len(errors), errors)
	}
}

func TestValidateRequired_EmptyField_ReturnsError(t *testing.T) {
	tests := []struct {
		field string
		name  string
	}{
		{"", "Email"},
		{"", "Password"},
		{"", "Name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequired(tt.field, tt.name)
			if err == "" {
				t.Error("expected error for empty field, got empty string")
			}
			expectedErr := tt.name + " is required"
			if err != expectedErr {
				t.Errorf("expected error '%s', got '%s'", expectedErr, err)
			}
		})
	}
}

func TestValidateRequired_NonEmptyField_ReturnsEmpty(t *testing.T) {
	tests := []struct {
		field string
		name  string
	}{
		{"test@example.com", "Email"},
		{"password123", "Password"},
		{"John Doe", "Name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequired(tt.field, tt.name)
			if err != "" {
				t.Errorf("expected no error for non-empty field, got '%s'", err)
			}
		})
	}
}

func TestValidateRequired_WhitespaceOnly_ReturnsError(t *testing.T) {
	// Fields with only whitespace should be considered empty after trimming
	field := "   "
	name := "Name"

	err := ValidateRequired(field, name)
	if err == "" {
		t.Error("expected error for whitespace-only field, got empty string")
	}
}
