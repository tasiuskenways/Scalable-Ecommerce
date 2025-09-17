package main

import (
	"flag"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/config"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/infrastructure/db"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/interfaces/http/routes"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/middleware"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/utils/jwt"
	"gorm.io/gorm"
)

func main() {

	cfg := config.Load()

	runMigration := flag.Bool("migrate", false, "Run migration")
	resetDb := flag.Bool("resetDb", false, "Reset DB")
	flag.Parse()

	var postgres *gorm.DB
	var err error

	if *runMigration {
		// Connect to database with running migrations
		db.NewPostgresConnection(cfg, *resetDb)
		return
	}

	postgres, err = db.ConnectWithoutMigration(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	redis, err := db.NewRedisConnection(cfg)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	// Initialize JWT manager
	jwtManager, err := jwt.NewTokenManager(&cfg.JWT, redis)
	if err != nil {
		log.Fatal("Failed to initialize JWT manager:", err)
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
		JWTManager:  jwtManager,
	})

	log.Printf("Server starting on port %s", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}

}
