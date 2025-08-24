package password

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost is the default cost for bcrypt hashing
	DefaultCost = 12
)

// HashPassword hashes a password using bcrypt with a random salt
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// CheckPassword compares a plaintext password with a hashed password
func CheckPassword(password, hashedPassword string) error {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    if err != nil {
        if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
            return fmt.Errorf("invalid password")
        }
        return fmt.Errorf("password comparison failed: %w", err)
    }
    return nil
}

// GenerateRandomString generates a random string of the given length
func GenerateRandomString(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b)[:length], nil
}

// GenerateAPIKey generates a secure API key
func GenerateAPIKey(prefix string, length int) (string, error) {
	key, err := GenerateRandomString(length)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s_%s", prefix, key), nil
}

// GenerateToken generates a secure token
func GenerateToken() (string, error) {
	return GenerateRandomString(32) // 32 bytes = 256 bits
}

// SanitizeString removes potentially dangerous characters from a string
func SanitizeString(input string) string {
	// Remove whitespace and control characters
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, "\n", "")
	input = strings.ReplaceAll(input, "\r", "")
	input = strings.ReplaceAll(input, "\t", "")
	return input
}
