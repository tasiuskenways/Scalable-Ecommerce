package utils

import (
	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
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
		Error:     message,
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
