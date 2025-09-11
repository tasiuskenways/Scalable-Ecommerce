package routes

import (
	"github.com/gofiber/fiber/v2"
	"tasius.my.id/SE/product-service/internal/application/services"
	"tasius.my.id/SE/product-service/internal/infrastructure/repositories"
	"tasius.my.id/SE/product-service/internal/interfaces/http/handlers"
)

func SetupProductRoutes(api fiber.Router, deps RoutesDependencies) {
	// Initialize repositories
	productRepo := repositories.NewProductRepository(deps.Db)
	categoryRepo := repositories.NewCategoryRepository(deps.Db)

	// Initialize services
	productService := services.NewProductService(productRepo, categoryRepo)
	categoryService := services.NewCategoryService(categoryRepo)

	// Initialize handlers
	productHandler := handlers.NewProductHandler(productService, categoryService)

	// Product routes
	products := api.Group("/products")
	products.Post("/", productHandler.CreateProduct)
	products.Get("/", productHandler.GetProducts)
	products.Get("/search", productHandler.SearchProducts)
	products.Get("/sku/:sku", productHandler.GetProductBySKU)
	products.Get("/:id", productHandler.GetProduct)
	products.Put("/:id", productHandler.UpdateProduct)
	products.Patch("/:id/stock", productHandler.UpdateProductStock)
	products.Delete("/:id", productHandler.DeleteProduct)

	// Category routes
	categories := api.Group("/categories")
	categories.Post("/", productHandler.CreateCategory)
	categories.Get("/", productHandler.GetCategories)
	categories.Get("/:id", productHandler.GetCategory)
	categories.Put("/:id", productHandler.UpdateCategory)
	categories.Delete("/:id", productHandler.DeleteCategory)

	// Products by category
	products.Get("/category/:categoryId", productHandler.GetProductsByCategory)
}
