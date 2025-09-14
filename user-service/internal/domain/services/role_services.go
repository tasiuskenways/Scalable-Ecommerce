package services

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/dto"
)

type RoleService interface {
	CreateRole(ctx *fiber.Ctx, req *dto.CreateRoleRequest) (*dto.RoleResponse, error)
	GetRole(ctx *fiber.Ctx, id string) (*dto.RoleResponse, error)
	GetAllRoles(ctx *fiber.Ctx) ([]dto.RoleResponse, error)
	UpdateRole(ctx *fiber.Ctx, id string, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error)
	DeleteRole(ctx *fiber.Ctx, id string) error
	AssignRolesToUser(ctx *fiber.Ctx, req *dto.AssignRoleRequest) error
	GetUserRoles(ctx *fiber.Ctx, userID string) ([]dto.RoleResponse, error)
	GetAllPermissions(ctx *fiber.Ctx) ([]dto.PermissionResponse, error)
}
