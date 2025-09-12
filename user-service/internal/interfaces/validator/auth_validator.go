package validator

import (
	"regexp"
	"strings"

	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/dto"
)

type AuthValidator struct {
	emailRegex *regexp.Regexp
}

func NewAuthValidator() *AuthValidator {
	return &AuthValidator{
		emailRegex: regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
	}
}

func (v *AuthValidator) ValidateRegister(req *dto.RegisterRequest) []string {
	var errors []string

	// Validate email
	if req.Email == "" {
		errors = append(errors, "Email is required")
	} else if !v.emailRegex.MatchString(req.Email) {
		errors = append(errors, "Invalid email format")
	}

	// Validate password
	if req.Password == "" {
		errors = append(errors, "Password is required")
	} else if len(req.Password) < 6 {
		errors = append(errors, "Password must be at least 6 characters long")
	}

	// Validate name
	if req.Name == "" {
		errors = append(errors, "Name is required")
	} else if len(strings.TrimSpace(req.Name)) < 2 {
		errors = append(errors, "Name must be at least 2 characters long")
	}

	return errors
}

func (v *AuthValidator) ValidateLogin(req *dto.LoginRequest) []string {
	var errors []string

	// Validate email
	if req.Email == "" {
		errors = append(errors, "Email is required")
	} else if !v.emailRegex.MatchString(req.Email) {
		errors = append(errors, "Invalid email format")
	}

	// Validate password
	if req.Password == "" {
		errors = append(errors, "Password is required")
	}

	return errors
}

func (v *AuthValidator) ValidateRefreshToken(req *dto.RefreshTokenRequest) []string {
	var errors []string

	// Validate refresh token
	if req.RefreshToken == "" {
		errors = append(errors, "Refresh token is required")
	}

	return errors
}
