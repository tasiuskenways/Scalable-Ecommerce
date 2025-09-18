package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/application/services"
	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/config"
	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/infrastructure/repositories"
	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/interfaces/http/handlers"
	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/utils"
	"gorm.io/gorm"
)

type RoutesDependencies struct {
	Db          *gorm.DB
	RedisClient *redis.Client
	Config      *config.Config
}

func SetupRoutes(app *fiber.App, deps RoutesDependencies) {
	// Initialize repositories
	storeRepo := repositories.NewStoreRepository(deps.Db)
	roleRepo := repositories.NewUserStoreRoleRepository(deps.Db)
	invitationRepo := repositories.NewStoreInvitationRepository(deps.Db)

	// Initialize services
	storeService := services.NewStoreService(storeRepo, roleRepo, invitationRepo)

	// Initialize handlers
	storeHandler := handlers.NewStoreHandler(storeService)

	// API routes
	api := app.Group("/api")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return utils.SuccessResponse(c, "Store service is healthy", nil)
	})

	// Store routes
	stores := api.Group("/stores")
	{
		stores.Post("/", storeHandler.CreateStore)
		stores.Get("/", storeHandler.GetUserStores)
		stores.Get("/:id", storeHandler.GetStore)
		stores.Get("/slug/:slug", storeHandler.GetStoreBySlug)
		stores.Put("/:id", storeHandler.UpdateStore)
		stores.Delete("/:id", storeHandler.DeleteStore)

		// Member management
		stores.Post("/:id/invite", storeHandler.InviteMember)
		stores.Get("/:id/members", storeHandler.GetStoreMembers)
		stores.Get("/:id/invitations", storeHandler.GetStoreInvitations)
		stores.Put("/:id/members/:memberId/role", storeHandler.UpdateMemberRole)
		stores.Delete("/:id/members/:memberId", storeHandler.RemoveMember)
	}

	// Invitation routes
	invitations := api.Group("/invitations")
	{
		invitations.Get("/", storeHandler.GetUserInvitations)    // Get all user's invitations
		invitations.Post("/accept", storeHandler.AcceptInvitation) // Accept an invitation
	}

}
