package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/Scalable-Ecommerce/crypto-service/internal/config"
	"github.com/tasiuskenways/Scalable-Ecommerce/crypto-service/internal/interfaces/http/handlers"
	"github.com/tasiuskenways/Scalable-Ecommerce/crypto-service/internal/utils"
)

func SetupRoutes(app *fiber.App, cfg *config.Config) {
	api := app.Group("/api")

	api.Get("/health", func(c *fiber.Ctx) error {
		return utils.SuccessResponse(c, "OK", nil)
	})

	api.Post("/decrypt", handlers.DecryptHandler(cfg.HybridEncryption.PrivateKeyPath))

	api.Post("/encrypt", handlers.EncryptHandler(cfg.HybridEncryption.PublicKeyPath))

}
