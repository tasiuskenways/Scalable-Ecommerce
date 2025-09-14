package services

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/dto"
)

type ProfileService interface {
	CreateProfile(ctx *fiber.Ctx, userID string, req *dto.CreateProfileRequest) (*dto.ProfileResponse, error)
	GetProfile(ctx *fiber.Ctx, userID string) (*dto.ProfileResponse, error)
	GetMyProfile(ctx *fiber.Ctx) (*dto.ProfileResponse, error)
	UpdateProfile(ctx *fiber.Ctx, userID string, req *dto.UpdateProfileRequest) (*dto.ProfileResponse, error)
	UpdateMyProfile(ctx *fiber.Ctx, req *dto.UpdateProfileRequest) (*dto.ProfileResponse, error)
	DeleteProfile(ctx *fiber.Ctx, userID string) error
}
