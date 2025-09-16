package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/application/dto"
	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/domain/services"
	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/utils"
)

type CartHandler struct {
	cartService services.CartService
}

func NewCartHandler(cartService services.CartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

func (h *CartHandler) GetCart(c *fiber.Ctx) error {
	userID := c.Get("X-User-Id")
	if userID == "" {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	cart, err := h.cartService.GetCart(c, userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Cart retrieved successfully", cart)
}

func (h *CartHandler) AddItemToCart(c *fiber.Ctx) error {
	userID := c.Get("X-User-Id")
	if userID == "" {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	var req dto.AddItemRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if _, err := uuid.Parse(req.ProductID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid product_id")
	}
	if req.Quantity < 1 {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Quantity must be >= 1")
	}

	cart, err := h.cartService.AddItemToCart(c, userID, &req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Item added to cart successfully", cart)
}

func (h *CartHandler) UpdateCartItem(c *fiber.Ctx) error {
	userID := c.Get("X-User-Id")
	if userID == "" {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	itemID := c.Params("itemId")
	if itemID == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Item ID is required")
	}

	var req dto.UpdateItemRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	cart, err := h.cartService.UpdateCartItem(c, userID, itemID, &req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Cart item updated successfully", cart)
}

func (h *CartHandler) RemoveItemFromCart(c *fiber.Ctx) error {
	userID := c.Get("X-User-Id")
	if userID == "" {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	itemID := c.Params("itemId")
	if itemID == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Item ID is required")
	}

	cart, err := h.cartService.RemoveItemFromCart(c, userID, itemID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Item removed from cart successfully", cart)
}

func (h *CartHandler) ClearCart(c *fiber.Ctx) error {
	userID := c.Get("X-User-Id")
	if userID == "" {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	if err := h.cartService.ClearCart(c, userID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Cart cleared successfully", nil)
}

func (h *CartHandler) ValidateCart(c *fiber.Ctx) error {
	userID := c.Get("X-User-Id")
	if userID == "" {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	validation, err := h.cartService.ValidateCart(c, userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Cart validated successfully", validation)
}
