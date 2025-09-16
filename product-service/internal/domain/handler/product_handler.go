package handler

import "github.com/gofiber/fiber/v2"

type ProductHandler interface {
	CreateProduct(c *fiber.Ctx) error
	GetProduct(c *fiber.Ctx) error
	GetProductBySKU(c *fiber.Ctx) error
	GetProducts(c *fiber.Ctx) error
	GetProductsByCategory(c *fiber.Ctx) error
	UpdateProduct(c *fiber.Ctx) error
	UpdateProductStock(c *fiber.Ctx) error
	DeleteProduct(c *fiber.Ctx) error
	SearchProducts(c *fiber.Ctx) error
	CreateCategory(c *fiber.Ctx) error
	GetCategory(c *fiber.Ctx) error
	GetCategories(c *fiber.Ctx) error
	UpdateCategory(c *fiber.Ctx) error
	DeleteCategory(c *fiber.Ctx) error
	GetProductsByIds(c *fiber.Ctx) error
}
