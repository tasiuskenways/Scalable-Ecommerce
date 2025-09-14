package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/services"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/infrastructure/repositories"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/interfaces/http/handlers"
)

// SetupRoleRoutes registers role-related HTTP endpoints on the provided Fiber router.
// It constructs role, permission, and user repositories and a RoleService/RoleHandler
// from the provided dependencies, then mounts admin routes under "/roles" for CRUD,
// assignment, and permissions (POST "/", GET "/", GET "/:id", PUT "/:id", DELETE "/:id",
// POST "/assign", GET "/users/:userId", GET "/permissions/all"). It also exposes a
// public endpoint at "/user/roles" that delegates to the handler but returns HTTP 401
// with JSON {"error":"user not authenticated"} if the "X-User-Id" request header is absent.
func SetupRoleRoutes(api fiber.Router, deps RoutesDependencies) {
	roleRepo := repositories.NewRoleRepository(deps.Db)
	permissionRepo := repositories.NewPermissionRepository(deps.Db)
	userRepo := repositories.NewUserRepository(deps.Db)
	roleService := services.NewRoleService(roleRepo, permissionRepo, userRepo, deps.Db)
	roleHandler := handlers.NewRoleHandler(roleService)

	// Admin only routes for role management
	roles := api.Group("/roles")

	// Role CRUD operations
	roles.Post("/", roleHandler.CreateRole)
	roles.Get("/", roleHandler.GetAllRoles)
	roles.Get("/:id", roleHandler.GetRole)
	roles.Put("/:id", roleHandler.UpdateRole)
	roles.Delete("/:id", roleHandler.DeleteRole)

	// User role assignment
	roles.Post("/assign", roleHandler.AssignRolesToUser)
	roles.Get("/users/:userId", roleHandler.GetUserRoles)

	// Permissions
	roles.Get("/permissions/all", roleHandler.GetAllPermissions)

	// Public route for users to see their own roles
	userRoles := api.Group("/user")
	userRoles.Get("/roles", func(c *fiber.Ctx) error {
		userID := c.Get("X-User-Id")
		if userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "user not authenticated",
			})
		}
		return roleHandler.GetUserRoles(c)
	})
}
