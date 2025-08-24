package routes

import (
	"github.com/gofiber/fiber/v2"
	"tasius.my.id/SE/crypto-service/internal/config"
	"tasius.my.id/SE/crypto-service/internal/interfaces/http/handlers"
	"tasius.my.id/SE/crypto-service/internal/utils"
)

func SetupRoutes(app *fiber.App, cfg *config.Config)  {
	api := app.Group("/api")

	api.Get("/health", func(c *fiber.Ctx) error {
		return utils.SuccessResponse(c, "OK", nil)
	})

	api.Post("/decrypt", handlers.DecryptHandler(cfg.HybridEncryption.PrivateKeyPath))

	api.Post("/encrypt", handlers.EncryptHandler(cfg.HybridEncryption.PublicKeyPath))

}
