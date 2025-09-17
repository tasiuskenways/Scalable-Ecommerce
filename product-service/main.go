package main

import (
	"flag"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/tasiuskenways/scalable-ecommerce/product-service/internal/config"
	"github.com/tasiuskenways/scalable-ecommerce/product-service/internal/infrastructure/db"
	"github.com/tasiuskenways/scalable-ecommerce/product-service/internal/infrastructure/seed"
	"github.com/tasiuskenways/scalable-ecommerce/product-service/internal/interfaces/http/routes"
	"github.com/tasiuskenways/scalable-ecommerce/product-service/internal/middleware"
	"gorm.io/gorm"
)

func main() {

	log.SetOutput(os.Stdout) // force logs to stdout
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg := config.Load()

	runMigration := flag.Bool("migrate", false, "Run migration")
	resetDb := flag.Bool("resetDb", false, "Reset DB")
	seedData := flag.Bool("seedData", false, "Seed data")
	flag.Parse()

	var postgres *gorm.DB
	var err error

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

	if *seedData {
		seed.SeedData(postgres)
		log.Println("Database seeded successfully!")
		return
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
	app.Use(recover.New())

	// Add comprehensive request/response logging
	app.Use(middleware.RequestResponseLogger())
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
