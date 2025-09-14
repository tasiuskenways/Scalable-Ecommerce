package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/services"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/infrastructure/repositories"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/interfaces/http/handlers"
)

func SetupInternalRoutes(api fiber.Router, deps RoutesDependencies) {
	userRepo := repositories.NewUserRepository(deps.Db)
	userService := services.NewUserService(userRepo, deps.Db)
	internalHandler := handlers.NewInternalHandler(userService)

	// Internal API for Kong and other services
	internal := api.Group("/internal")

	// RBAC endpoint for Kong
	internal.Get("/users/:userId/rbac", internalHandler.GetUserRBACInfo)
}
