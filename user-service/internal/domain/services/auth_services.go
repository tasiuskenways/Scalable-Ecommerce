package services

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/dto"
)

type AuthService interface {
	Register(ctx *fiber.Ctx, req *dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx *fiber.Ctx, req *dto.LoginRequest) (*dto.AuthResponse, error)
	Logout(ctx *fiber.Ctx) error
	RefreshToken(ctx *fiber.Ctx, refreshToken string) (*dto.AuthResponse, error)
	ValidateToken(ctx *fiber.Ctx, token string) (*dto.UserResponse, error)
}
