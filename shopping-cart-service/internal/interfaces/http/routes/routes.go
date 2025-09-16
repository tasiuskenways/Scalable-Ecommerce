package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/application/services"
	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/config"
	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/infrastructure/external"
	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/infrastructure/repositories"
	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/interfaces/http/handlers"
	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/utils"
	"gorm.io/gorm"
)

type RoutesDependencies struct {
	Db          *gorm.DB
	RedisClient *redis.Client
	Config      *config.Config
}

func SetupRoutes(app *fiber.App, deps RoutesDependencies) {
	api := app.Group("/api")

	api.Get("/health", func(c *fiber.Ctx) error {
		return utils.SuccessResponse(c, "OK", nil)
	})

	// Initialize repositories
	cartRepo := repositories.NewCartRepository(deps.Db)
	cartItemRepo := repositories.NewCartItemRepository(deps.Db)

	// Initialize external service clients
	productService := external.NewProductServiceClient(deps.Config.ProductServiceURL)

	// Initialize services
	cartService := services.NewCartService(
		cartRepo,
		cartItemRepo,
		productService,
		deps.Config,
	)

	// Initialize handlers
	cartHandler := handlers.NewCartHandler(cartService)

	// Cart routes
	cart := api.Group("/cart")
	cart.Get("/", cartHandler.GetCart)
	cart.Post("/items", cartHandler.AddItemToCart)
	cart.Put("/items/:itemId", cartHandler.UpdateCartItem)
	cart.Delete("/items/:itemId", cartHandler.RemoveItemFromCart)
	cart.Delete("/clear", cartHandler.ClearCart)
	cart.Post("/validate", cartHandler.ValidateCart)
}
