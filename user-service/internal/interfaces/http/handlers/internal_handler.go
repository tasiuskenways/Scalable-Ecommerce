package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/services"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/utils"
)

type InternalHandler struct {
	userService services.UserService
}

func NewInternalHandler(userService services.UserService) *InternalHandler {
	return &InternalHandler{
		userService: userService,
	}
}

// GetUserRBACInfo returns user roles and permissions for Kong
func (h *InternalHandler) GetUserRBACInfo(c *fiber.Ctx) error {
	userID := c.Params("userId")
	if userID == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "User ID is required")
	}

	// Check if this is an internal service request
	internalService := c.Get("X-Internal-Service")
	if internalService != "kong-auth" {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Access denied")
	}

	rbacInfo, err := h.userService.GetUserRBACInfo(c, userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SuccessResponse(c, "User RBAC info retrieved", rbacInfo)
}
