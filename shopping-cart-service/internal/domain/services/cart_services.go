package services

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/application/dto"
)

type CartService interface {
	GetCart(ctx *fiber.Ctx, userID string) (*dto.CartResponse, error)
	AddItemToCart(ctx *fiber.Ctx, userID string, req *dto.AddItemRequest) (*dto.CartResponse, error)
	UpdateCartItem(ctx *fiber.Ctx, userID string, itemID string, req *dto.UpdateItemRequest) (*dto.CartResponse, error)
	RemoveItemFromCart(ctx *fiber.Ctx, userID string, itemID string) (*dto.CartResponse, error)
	ClearCart(ctx *fiber.Ctx, userID string) error
	ValidateCart(ctx *fiber.Ctx, userID string) (*dto.CartValidationResponse, error)
}
