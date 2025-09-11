package main

import (
	"flag"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"tasius.my.id/SE/crypto-service/internal/config"
	"tasius.my.id/SE/crypto-service/internal/interfaces/http/routes"
	"tasius.my.id/SE/crypto-service/internal/utils/crypto"
)

func main() {

	generateKeys := flag.Bool("generateKeys", false, "Generate new keys")
	flag.Parse()

	if *generateKeys {
		filePath := flag.String("filePath", "./keys", "Path to the keys directory")
		flag.Parse()

		privateKeyPath, publicKeyPath, err := crypto.GenerateAndSaveRSAKeyPair(*filePath)
		if err != nil {
			log.Fatal("Failed to generate and save RSA key pair:", err)
		}
		log.Println("Private key saved to:", privateKeyPath)
		log.Println("Public key saved to:", publicKeyPath)
		return
	}

	cfg := config.Load()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		},
	})

	app.Use(requestid.New())

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	routes.SetupRoutes(app, cfg)

	log.Printf("Server starting on port %s", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}

}
