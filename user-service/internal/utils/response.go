package utils

import "github.com/gofiber/fiber/v2"

type Response struct {
	RequestID string `json:"request_id,omitempty"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	Error     string `json:"error,omitempty"`
}

func SuccessResponse(c *fiber.Ctx, message string, data interface{}) error {
	requestID := getRequestID(c)
	return c.Status(fiber.StatusOK).JSON(Response{
		RequestID: requestID,
		Success:   true,
		Message:   message,
		Data:      data,
	})
}

func CreatedResponse(c *fiber.Ctx, message string, data interface{}) error {
	requestID := getRequestID(c)
	return c.Status(fiber.StatusCreated).JSON(Response{
		RequestID: requestID,
		Success:   true,
		Message:   message,
		Data:      data,
	})
}

func ErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	requestID := getRequestID(c)
	return c.Status(statusCode).JSON(Response{
		RequestID: requestID,
		Success:   false,
		Message:   message,
		Error:     message,
	})
}

func ValidationErrorResponse(c *fiber.Ctx, errors []string) error {
	requestID := getRequestID(c)
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"request_id": requestID,
		"success":    false,
		"message":    "Validation failed",
		"errors":     errors,
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