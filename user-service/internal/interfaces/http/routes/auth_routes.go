package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/services"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/infrastructure/repositories"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/interfaces/http/handlers"
)

// SetupAuthRoutes registers authentication endpoints on the provided Fiber router.
// It wires repositories, services, and handlers from the given dependencies and mounts
// the following public routes under the "/auth" group:
//   POST /auth/register  -> authHandler.Register
//   POST /auth/login     -> authHandler.Login
//   POST /auth/refresh   -> authHandler.RefreshToken
//   POST /auth/logout    -> authHandler.Logout
// The function performs setup only and does not return an error; route handlers handle
// request-level errors. The provided dependencies are used to construct the repository,
// service, and handler instances.
func SetupAuthRoutes(api fiber.Router, deps RoutesDependencies) {

	userRepo := repositories.NewUserRepository(deps.Db)
	authService := services.NewAuthService(userRepo, deps.RedisClient, &deps.Config.JWT, deps.JWTManager)
	authHandler := handlers.NewAuthHandler(authService)

	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)
	auth.Post("/logout", authHandler.Logout)
}
