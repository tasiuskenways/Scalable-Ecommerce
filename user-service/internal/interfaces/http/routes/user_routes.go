package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/services"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/infrastructure/repositories"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/interfaces/http/handlers"
)

// SetupUserRoutes registers user-related HTTP routes on the given Fiber router.
// It constructs the repository, service, and handler from the provided dependencies
// and mounts routes under "/users":
//   - GET, PUT /users/me          : access and update the current user's profile
//   - (admin) GET  /users/        : list users
//   - (admin) GET  /users/:id     : get a user by ID
//   - (admin) PUT  /users/:id     : update a user by ID
//   - (admin) DELETE /users/:id   : delete a user by ID
//
// The admin routes are protected by AdminOnlyMiddleware using deps.Db.
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
	adminUsers := users.Group("/")
	adminUsers.Get("/", userHandler.GetAllUsers)
	adminUsers.Get("/:id", userHandler.GetUser)
	adminUsers.Put("/:id", userHandler.UpdateUser)
	adminUsers.Delete("/:id", userHandler.DeleteUser)
}
