package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/repositories"
	"gorm.io/gorm"
)

// RoleMiddleware checks if user has required roles
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

// PermissionMiddleware checks if user has required permissions
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

// AdminOnlyMiddleware restricts access to admin users only
func AdminOnlyMiddleware(db *gorm.DB) fiber.Handler {
	return RoleMiddleware(db, "admin", "super_admin")
}

// OwnerOrAdminMiddleware allows access if user is owner of resource or admin
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
