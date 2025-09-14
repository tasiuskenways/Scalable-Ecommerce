package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/config"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/utils"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/utils/jwt"
	"gorm.io/gorm"
)

type RoutesDependencies struct {
	Db          *gorm.DB
	RedisClient *redis.Client
	Config      *config.Config
	JWTManager  *jwt.TokenManager
}

func SetupRoutes(app *fiber.App, deps RoutesDependencies) {
	api := app.Group("/api")

	api.Get("/health", func(c *fiber.Ctx) error {
		return utils.SuccessResponse(c, "OK", nil)
	})

	// Setup all routes
	SetupAuthRoutes(api, deps)
	SetupUserRoutes(api, deps)
	SetupProfileRoutes(api, deps)
	SetupRoleRoutes(api, deps)
}
