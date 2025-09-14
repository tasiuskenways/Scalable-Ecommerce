package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/repositories"
	"gorm.io/gorm"
)

// RoleMiddleware returns a Fiber handler that enforces role-based access.
// RoleMiddleware checks the "X-User-Id" header for an authenticated user (responds 401 if missing),
// loads the user's roles from the database, and requires that the user has at least one of the
// supplied requiredRoles. If the user cannot be loaded or lacks any required role the handler
// responds with 403; otherwise it calls the next handler.
func RoleMiddleware(db *gorm.DB, requiredRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Get("X-User-Id")
		if userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "user not authenticated",
			})
		}

		// Get user with roles
		var user struct {
			ID    string `json:"id"`
			Roles []struct {
				Name string `json:"name"`
			} `json:"roles"`
		}

		if err := db.Table("users").
			Select("users.id").
			Preload("Roles").
			Where("users.id = ?", userID).
			First(&user).Error; err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied",
			})
		}

		// Check if user has any of the required roles
		userRoles := make(map[string]bool)
		for _, role := range user.Roles {
			userRoles[role.Name] = true
		}

		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			if userRoles[requiredRole] {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "insufficient permissions",
			})
		}

		return c.Next()
	}
}

// PermissionMiddleware returns a Fiber middleware that allows the request to proceed
// only if the authenticated user has at least one of the specified permissions.
//
// PermissionMiddleware reads the user ID from the "X-User-Id" request header and
// loads the user via the provided UserRepository. If the header is missing it
// responds 401 ("user not authenticated"); if the user cannot be loaded it
// responds 403 ("access denied"). It then checks the user's permissions (via
// user.GetPermissions()) and, if none of the required permissions are present,
// responds 403 ("insufficient permissions"). When authorization succeeds the
// middleware stores the user's permissions in the request context under
// "userPermissions" and calls the next handler.
func PermissionMiddleware(userRepo repositories.UserRepository, requiredPermissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Get("X-User-Id")
		if userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "user not authenticated",
			})
		}

		// This would need to be implemented in your user repository
		// to load user with roles and permissions
		user, err := userRepo.GetByID(c.Context(), userID)
		if err != nil || user == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied",
			})
		}

		// Check if user has any of the required permissions
		userPermissions := user.GetPermissions()
		permissionMap := make(map[string]bool)
		for _, perm := range userPermissions {
			permissionMap[perm] = true
		}

		hasRequiredPermission := false
		for _, requiredPerm := range requiredPermissions {
			if permissionMap[requiredPerm] {
				hasRequiredPermission = true
				break
			}
		}

		if !hasRequiredPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "insufficient permissions",
			})
		}

		// Store user permissions in context for later use
		c.Locals("userPermissions", userPermissions)

		return c.Next()
	}
}

// AdminOnlyMiddleware returns a Fiber middleware that permits access only to users
// who have the "admin" or "super_admin" role. It delegates to RoleMiddleware and
// uses the provided *gorm.DB to load user records and their roles.
func AdminOnlyMiddleware(db *gorm.DB) fiber.Handler {
	return RoleMiddleware(db, "admin", "super_admin")
}

// resourceUserIDParam is the name of the route parameter that contains the resource's owner user ID.
func OwnerOrAdminMiddleware(db *gorm.DB, resourceUserIDParam string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Get("X-User-Id")
		if userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "user not authenticated",
			})
		}

		resourceUserID := c.Params(resourceUserIDParam)
		if resourceUserID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "resource user ID is required",
			})
		}

		// If user is accessing their own resource, allow
		if userID == resourceUserID {
			return c.Next()
		}

		// Check if user is admin
		var user struct {
			ID    string `json:"id"`
			Roles []struct {
				Name string `json:"name"`
			} `json:"roles"`
		}

		if err := db.Table("users").
			Select("users.id").
			Preload("Roles").
			Where("users.id = ?", userID).
			First(&user).Error; err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied",
			})
		}

		// Check if user has admin role
		for _, role := range user.Roles {
			if role.Name == "admin" || role.Name == "super_admin" {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied - you can only access your own resources",
		})
	}
}
