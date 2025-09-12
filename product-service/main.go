package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/tasiuskenways/scalable-ecommerce/product-service/internal/config"
	"github.com/tasiuskenways/scalable-ecommerce/product-service/internal/infrastructure/db"
	"github.com/tasiuskenways/scalable-ecommerce/product-service/internal/interfaces/http/routes"
	"gorm.io/gorm"
)

func main() {

	cfg := config.Load()

	runMigration := flag.Bool("migrate", false, "Run migration")
	resetDb := flag.Bool("resetDb", false, "Reset DB")
	flag.Parse()

	var postgres *gorm.DB
	var err error

	fmt.Println("runMigration:", *runMigration)
	fmt.Println("resetDb:", *resetDb)

	if *runMigration {
		// Connect to database with running migrations
		db.NewPostgresConnection(cfg, *resetDb)
		return
	}

	// Connect to database without running migrations
	postgres, err = db.ConnectWithoutMigration(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	redis, err := db.NewRedisConnection(cfg)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

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

	app.Use(logger.New(logger.Config{
		Format: "------------------------\n Time: ${time}\n Status: ${status}\n Latency: ${latency}\n IP: ${ip}\n Method: ${method}\n Path: ${path} \n RequestId: ${locals:requestid}\n------------------------\n",
	}))

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	routes.SetupRoutes(app, routes.RoutesDependencies{
		Db:          postgres,
		RedisClient: redis,
		Config:      cfg,
	})

	log.Printf("Server starting on port %s", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}

}
