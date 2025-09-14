package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/services"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/infrastructure/repositories"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/interfaces/http/handlers"
)

// SetupProfileRoutes registers profile-related HTTP routes on the given Fiber router.
// It wires the user and profile repositories, constructs the profile service and handler,
// and mounts routes under the "/profiles" base path.
//
// Registered routes:
//   - GET  /profiles/me                     -> GetMyProfile
//   - POST /profiles/me                     -> CreateProfile
//   - PUT  /profiles/me                     -> UpdateMyProfile
//   - GET  /profiles/users/:userId/profile  -> GetProfile  (registered for both admin and owner groups)
//   - POST /profiles/users/:userId/profile  -> CreateProfile (admin)
//   - PUT  /profiles/users/:userId/profile  -> UpdateProfile (admin)
//   - DELETE /profiles/users/:userId/profile -> DeleteProfile (admin)
//
// Note: the GET /profiles/users/:userId/profile endpoint is intentionally registered in both
// admin and owner groups in the current routing setup.
func SetupProfileRoutes(api fiber.Router, deps RoutesDependencies) {
	userRepo := repositories.NewUserRepository(deps.Db)
	profileRepo := repositories.NewProfileRepository(deps.Db)
	profileService := services.NewProfileService(profileRepo, userRepo)
	profileHandler := handlers.NewProfileHandler(profileService)

	// Protected routes
	profiles := api.Group("/profiles")

	// User can manage their own profile
	profiles.Get("/me", profileHandler.GetMyProfile)
	profiles.Post("/me", profileHandler.CreateProfile)
	profiles.Put("/me", profileHandler.UpdateMyProfile)

	// Admin routes for managing other users' profiles
	adminProfiles := profiles.Group("/users")
	adminProfiles.Get("/:userId/profile", profileHandler.GetProfile)
	adminProfiles.Post("/:userId/profile", profileHandler.CreateProfile)
	adminProfiles.Put("/:userId/profile", profileHandler.UpdateProfile)
	adminProfiles.Delete("/:userId/profile", profileHandler.DeleteProfile)

	// Owner or admin can access specific user profiles
	ownerProfiles := profiles.Group("/users")
	ownerProfiles.Get("/:userId/profile", profileHandler.GetProfile)
}
