package utils

import (
	"github.com/gofiber/fiber/v2"
)

type Response struct {
	RequestId string `json:"requestId"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	Error     string `json:"error,omitempty"`
}

func SuccessResponse(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		RequestId: c.Locals("requestid").(string),
		Success: true,
		Message: message,
		Data:    data,
	})
}

func CreatedResponse(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(Response{
		RequestId: c.Locals("requestid").(string),
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(Response{
		RequestId: c.Locals("requestid").(string),
		Success: false,
		Message: message,
		Error:   message,
	})
}

func ValidationErrorResponse(c *fiber.Ctx, errors []string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"requestId": c.Locals("requestid").(string),
		"success": false,
		"message": "Validation failed",
		"errors":  errors,
	})
}