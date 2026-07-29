package services

import (
	"errors"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const PasswordHashCost = 12
const MinimumPasswordLength = 8

var ErrWeakPassword = errors.New("password must be at least 8 characters and include uppercase, lowercase, number, and symbol characters")

// requireStrongPassword reads env at call time so tests and runtime can override it.
func requireStrongPassword() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("REQUIRE_STRONG_PASSWORD")))
	return v != "false" && v != "0" && v != "no"
}

func ValidatePasswordStrength(password string) error {
	if utf8.RuneCountInString(password) < MinimumPasswordLength {
		return ErrWeakPassword
	}
	// Skip strength checks when REQUIRE_STRONG_PASSWORD=false
	if !requireStrongPassword() {
		return nil
	}
	var hasUpper, hasLower, hasNumber, hasSymbol bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char), unicode.IsSymbol(char):
			hasSymbol = true
		}
	}
	if !hasUpper || !hasLower || !hasNumber || !hasSymbol || strings.Contains(password, " ") {
		return ErrWeakPassword
	}
	return nil
}

func HashPassword(password string) (string, error) {
	if err := ValidatePasswordStrength(password); err != nil {
		return "", err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), PasswordHashCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
