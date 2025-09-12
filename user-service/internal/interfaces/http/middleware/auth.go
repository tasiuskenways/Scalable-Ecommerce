package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/utils/jwt"
)

const (
	AuthorizationHeader = "Authorization"
	BearerSchema        = "Bearer"
)

func AuthMiddleware(jwtManager *jwt.TokenManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get(AuthorizationHeader)
		if authHeader == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "authorization header is required",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != BearerSchema {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		token := parts[1]
		claims, err := jwtManager.ValidateToken(token, jwt.AccessToken)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		// Add user info to context
		c.Locals("userID", claims.UserID)
		c.Locals("email", claims.Email)

		return c.Next()
	}
}

// GetUserIDFromContext gets the user ID from the context
func GetUserIDFromContext(c *fiber.Ctx) (string, error) {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return "", errors.New("user ID not found in context")
	}
	return userID, nil
}

// GetEmailFromContext gets the email from the context
func GetEmailFromContext(c *fiber.Ctx) (string, error) {
	email, ok := c.Locals("email").(string)
	if !ok {
		return "", errors.New("email not found in context")
	}
	return email, nil
}
