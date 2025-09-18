package utils

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     interface{} `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

func SuccessResponse(c *fiber.Ctx, message string, data interface{}) error {
	requestID := getRequestID(c)
	return c.JSON(Response{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: requestID,
	})
}

func ErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	requestID := getRequestID(c)
	return c.Status(statusCode).JSON(Response{
		Success:   false,
		Message:   message,
		RequestID: requestID,
	})
}

func ValidationErrorResponse(c *fiber.Ctx, err error) error {
	var validationErrors []string

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, validationErr := range validationErrs {
			field := validationErr.Field()
			tag := validationErr.Tag()

			var message string
			switch tag {
			case "required":
				message = field + " is required"
			case "email":
				message = field + " must be a valid email address"
			case "min":
				message = field + " must be at least " + validationErr.Param() + " characters long"
			case "max":
				message = field + " must be at most " + validationErr.Param() + " characters long"
			case "url":
				message = field + " must be a valid URL"
			case "alphanum":
				message = field + " must contain only alphanumeric characters"
			case "oneof":
				message = field + " must be one of: " + validationErr.Param()
			default:
				message = field + " is invalid"
			}

			validationErrors = append(validationErrors, message)
		}
	}

	requestID := getRequestID(c)
	return c.Status(fiber.StatusBadRequest).JSON(Response{
		Success:   false,
		Message:   "Validation failed",
		Error:     validationErrors,
		RequestID: requestID,
	})
}

// Helper function to get request ID from context
func getRequestID(c *fiber.Ctx) string {
	if rid := c.Locals("requestid"); rid != nil {
		if ridStr, ok := rid.(string); ok {
			return ridStr
		}
	}
	return ""
}
