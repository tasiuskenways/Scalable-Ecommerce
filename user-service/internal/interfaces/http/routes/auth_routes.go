package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/services"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/infrastructure/repositories"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/interfaces/http/handlers"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/interfaces/http/middleware"
)

func SetupAuthRoutes(api fiber.Router, deps RoutesDependencies) {

	userRepo := repositories.NewUserRepository(deps.Db)
	authService := services.NewAuthService(userRepo, deps.RedisClient, &deps.Config.JWT, deps.JWTManager)
	authHandler := handlers.NewAuthHandler(authService)

	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)

	authProtected := api.Group("/auth", middleware.AuthMiddleware(deps.JWTManager))
	authProtected.Post("/logout", authHandler.Logout)
}
