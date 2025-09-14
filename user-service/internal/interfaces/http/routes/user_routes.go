package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/services"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/infrastructure/repositories"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/interfaces/http/handlers"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/interfaces/http/middleware"
)

func SetupUserRoutes(api fiber.Router, deps RoutesDependencies) {
	userRepo := repositories.NewUserRepository(deps.Db)
	userService := services.NewUserService(userRepo, deps.Db)
	userHandler := handlers.NewUserHandler(userService)

	// Protected routes
	users := api.Group("/users")

	// User can access their own info
	users.Get("/me", userHandler.GetMe)
	users.Put("/me", userHandler.UpdateMe)

	// Admin routes
	adminUsers := users.Group("/", middleware.AdminOnlyMiddleware(deps.Db))
	adminUsers.Get("/", userHandler.GetAllUsers)
	adminUsers.Get("/:id", userHandler.GetUser)
	adminUsers.Put("/:id", userHandler.UpdateUser)
	adminUsers.Delete("/:id", userHandler.DeleteUser)
}
