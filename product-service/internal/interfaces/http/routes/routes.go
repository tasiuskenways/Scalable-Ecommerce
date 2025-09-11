package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"tasius.my.id/SE/product-service/internal/config"
	"tasius.my.id/SE/product-service/internal/utils"
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

	SetupProductRoutes(api, deps)
}
