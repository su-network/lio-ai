package auth

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	// HashCost is the bcrypt hashing cost
	HashCost = 12
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), HashCost)
	return string(bytes), err
}

// CheckPassword verifies a password against a hash
func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// ValidatePassword checks if password meets security requirements
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return &PasswordError{
			Message: "password must be at least 8 characters",
		}
	}

	if len(password) > 128 {
		return &PasswordError{
			Message: "password must not exceed 128 characters",
		}
	}

	// Check for at least one uppercase letter
	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, ch := range password {
		switch {
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return &PasswordError{
			Message: "password must contain uppercase, lowercase, and digits",
		}
	}

	return nil
}

// PasswordError represents a password validation error
type PasswordError struct {
	Message string
}

func (e *PasswordError) Error() string {
	return e.Message
}
