package handlers

import (
	"github.com/gofiber/fiber/v2"
	"tasius.my.id/SE/user-service/internal/application/dto"
	"tasius.my.id/SE/user-service/internal/domain/services"
	"tasius.my.id/SE/user-service/internal/interfaces/validator"
	"tasius.my.id/SE/user-service/internal/utils"
)

type AuthHandler struct {
	authService services.AuthService
	validator   *validator.AuthValidator
}

const INVALID_REQUEST_BODY = "Invalid request body"

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.NewAuthValidator(),
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, INVALID_REQUEST_BODY)
	}

	// Validate request
	if errors := h.validator.ValidateRegister(&req); len(errors) > 0 {
		return utils.ValidationErrorResponse(c, errors)
	}

	// Register user
	response, err := h.authService.Register(c, &req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.CreatedResponse(c, "User registered successfully", response)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, INVALID_REQUEST_BODY)
	}

	// Validate request
	if errors := h.validator.ValidateLogin(&req); len(errors) > 0 {
		return utils.ValidationErrorResponse(c, errors)
	}

	// Login user
	response, err := h.authService.Login(c, &req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	return utils.SuccessResponse(c, "Login successful", response)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	if err := h.authService.Logout(c); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Successfully logged out", nil)
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req dto.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, INVALID_REQUEST_BODY)
	}

	// Validate request
	if errors := h.validator.ValidateRefreshToken(&req); len(errors) > 0 {
		return utils.ValidationErrorResponse(c, errors)
	}

	// Refresh token
	response, err := h.authService.RefreshToken(c, req.RefreshToken)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	return utils.SuccessResponse(c, "Token refreshed successfully", response)
}