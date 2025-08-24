package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"tasius.my.id/SE/user-service/internal/config"
	"tasius.my.id/SE/user-service/internal/utils"
	"tasius.my.id/SE/user-service/internal/utils/jwt"
)

type RoutesDependencies struct {
	Db          *gorm.DB
	RedisClient *redis.Client
	Config      *config.Config
	JWTManager  *jwt.TokenManager
}



func SetupRoutes(app *fiber.App, deps RoutesDependencies)  {
	api := app.Group("/api")

	api.Get("/health", func(c *fiber.Ctx) error {
		return utils.SuccessResponse(c, "OK", nil)
	})

	SetupAuthRoutes(api, deps)
}