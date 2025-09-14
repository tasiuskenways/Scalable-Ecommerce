package services

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/dto"
)

type UserService interface {
	GetUser(ctx *fiber.Ctx, id string) (*dto.UserResponse, error)
	GetAllUsers(ctx *fiber.Ctx, page, limit int) (*dto.PaginatedResponse, error)
	UpdateUser(ctx *fiber.Ctx, id string, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(ctx *fiber.Ctx, id string) error
	GetUserRBACInfo(ctx *fiber.Ctx, id string) (*dto.UserRBACResponse, error) // New method
}
