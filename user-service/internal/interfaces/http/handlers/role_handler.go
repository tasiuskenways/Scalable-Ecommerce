package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/dto"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/services"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/utils"
)

type RoleHandler struct {
	roleService services.RoleService
}

// NewRoleHandler creates a RoleHandler wired with the provided RoleService.
// The returned handler exposes HTTP endpoints for role-related operations backed by the service.
func NewRoleHandler(roleService services.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

func (h *RoleHandler) CreateRole(c *fiber.Ctx) error {
	var req dto.CreateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	response, err := h.roleService.CreateRole(c, &req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.CreatedResponse(c, "Role created successfully", response)
}

func (h *RoleHandler) GetRole(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Role ID is required")
	}

	response, err := h.roleService.GetRole(c, id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SuccessResponse(c, "Role retrieved successfully", response)
}

func (h *RoleHandler) GetAllRoles(c *fiber.Ctx) error {
	response, err := h.roleService.GetAllRoles(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Roles retrieved successfully", response)
}

func (h *RoleHandler) UpdateRole(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Role ID is required")
	}

	var req dto.UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	response, err := h.roleService.UpdateRole(c, id, &req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Role updated successfully", response)
}

func (h *RoleHandler) DeleteRole(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Role ID is required")
	}

	if err := h.roleService.DeleteRole(c, id); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Role deleted successfully", nil)
}

func (h *RoleHandler) AssignRolesToUser(c *fiber.Ctx) error {
	var req dto.AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.roleService.AssignRolesToUser(c, &req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Roles assigned successfully", nil)
}

func (h *RoleHandler) GetUserRoles(c *fiber.Ctx) error {
	userID := c.Get("X-User-Id")
	if userID == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "User ID is required")
	}

	response, err := h.roleService.GetUserRoles(c, userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SuccessResponse(c, "User roles retrieved successfully", response)
}

func (h *RoleHandler) GetAllPermissions(c *fiber.Ctx) error {
	response, err := h.roleService.GetAllPermissions(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Permissions retrieved successfully", response)
}
