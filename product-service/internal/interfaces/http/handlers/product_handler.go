package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/product-service/internal/application/dto"
	"github.com/tasiuskenways/scalable-ecommerce/product-service/internal/domain/entities"
	"github.com/tasiuskenways/scalable-ecommerce/product-service/internal/domain/services"
	"github.com/tasiuskenways/scalable-ecommerce/product-service/internal/utils"
)

type ProductHandler struct {
	productService  services.ProductService
	categoryService services.CategoryService
}

func NewProductHandler(productService services.ProductService, categoryService services.CategoryService) *ProductHandler {
	return &ProductHandler{
		productService:  productService,
		categoryService: categoryService,
	}
}

// Product Handlers
func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var req dto.CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	product := &entities.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CategoryID:  req.CategoryID,
		SKU:         req.SKU,
	}

	if err := h.productService.CreateProduct(c.Context(), product); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Product created successfully", product)
}

func (h *ProductHandler) GetProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	product, err := h.productService.GetProduct(c.Context(), id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Product not found")
	}

	return utils.SuccessResponse(c, "Product retrieved successfully", product)
}

func (h *ProductHandler) GetProductBySKU(c *fiber.Ctx) error {
	sku := c.Params("sku")
	product, err := h.productService.GetProductBySKU(c.Context(), sku)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Product not found")
	}

	return utils.SuccessResponse(c, "Product retrieved successfully", product)
}

func (h *ProductHandler) GetProducts(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	products, err := h.productService.GetProducts(c.Context(), limit, offset)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve products")
	}

	return utils.SuccessResponse(c, "Products retrieved successfully", products)
}

func (h *ProductHandler) GetProductsByCategory(c *fiber.Ctx) error {
	categoryID := c.Params("categoryId")
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	products, err := h.productService.GetProductsByCategory(c.Context(), categoryID, limit, offset)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Products retrieved successfully", products)
}

func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	id := c.Params("id")

	var req dto.UpdateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Get existing product
	product, err := h.productService.GetProduct(c.Context(), id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Product not found")
	}

	// Update fields if provided
	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	if req.CategoryID != nil {
		product.CategoryID = *req.CategoryID
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := h.productService.UpdateProduct(c.Context(), product); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Product updated successfully", product)
}

func (h *ProductHandler) UpdateProductStock(c *fiber.Ctx) error {
	id := c.Params("id")

	var req dto.UpdateStockRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.productService.UpdateProductStock(c.Context(), id, req.Stock); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Product stock updated successfully", nil)
}

func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.productService.DeleteProduct(c.Context(), id); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Product deleted successfully", nil)
}

func (h *ProductHandler) SearchProducts(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Search query is required")
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	products, err := h.productService.SearchProducts(c.Context(), query, limit, offset)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to search products")
	}

	return utils.SuccessResponse(c, "Products found successfully", products)
}

// Category Handlers
func (h *ProductHandler) CreateCategory(c *fiber.Ctx) error {
	var req dto.CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	category := &entities.Category{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.categoryService.CreateCategory(c.Context(), category); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Category created successfully", category)
}

func (h *ProductHandler) GetCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	category, err := h.categoryService.GetCategory(c.Context(), id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Category not found")
	}

	return utils.SuccessResponse(c, "Category retrieved successfully", category)
}

func (h *ProductHandler) GetCategories(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	categories, err := h.categoryService.GetCategories(c.Context(), limit, offset)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve categories")
	}

	return utils.SuccessResponse(c, "Categories retrieved successfully", categories)
}

func (h *ProductHandler) UpdateCategory(c *fiber.Ctx) error {
	id := c.Params("id")

	var req dto.UpdateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Get existing category
	category, err := h.categoryService.GetCategory(c.Context(), id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Category not found")
	}

	// Update fields if provided
	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Description != nil {
		category.Description = *req.Description
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := h.categoryService.UpdateCategory(c.Context(), category); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Category updated successfully", category)
}

func (h *ProductHandler) DeleteCategory(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.categoryService.DeleteCategory(c.Context(), id); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, "Category deleted successfully", nil)
}

func (h *ProductHandler) GetProductsByIds(c *fiber.Ctx) error {
	var req dto.GetProductsByIdsRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	products, err := h.productService.GetProductsByIds(c.Context(), req.Ids)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve products")
	}

	return utils.SuccessResponse(c, "Products retrieved successfully", products)
}
