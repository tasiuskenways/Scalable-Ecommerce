package db

import (
	"fmt"

	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/entities"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB, resetDb bool) error {
	// Enable UUID extension
	err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error
	if err != nil {
		return fmt.Errorf("failed to create uuid extension: %w", err)
	}

	if resetDb {
		// Drop existing tables if they exist (in reverse order due to foreign keys)
		err = db.Migrator().DropTable(
			&entities.UserProfile{},
			&entities.User{},
			&entities.Role{},
			&entities.Permission{},
		)
		if err != nil {
			return fmt.Errorf("failed to drop tables: %w", err)
		}
	}

	// Create tables with new schema
	err = db.AutoMigrate(
		&entities.User{},
		&entities.UserProfile{},
		&entities.Role{},
		&entities.Permission{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate tables: %w", err)
	}

	// Seed default roles and permissions if they don't exist
	if err := seedDefaultRolesAndPermissions(db); err != nil {
		return fmt.Errorf("failed to seed default roles and permissions: %w", err)
	}

	return nil
}

func seedDefaultRolesAndPermissions(db *gorm.DB) error {
	// Define default permissions
	permissions := []entities.Permission{
		// User permissions
		{Name: "user:read", Resource: "user", Action: "read", Description: "Read user information"},
		{Name: "user:update", Resource: "user", Action: "update", Description: "Update user information"},
		{Name: "user:delete", Resource: "user", Action: "delete", Description: "Delete user"},
		{Name: "user:create", Resource: "user", Action: "create", Description: "Create user"},
		{Name: "user:list", Resource: "user", Action: "list", Description: "List users"},

		// Profile permissions
		{Name: "profile:read", Resource: "profile", Action: "read", Description: "Read user profile"},
		{Name: "profile:update", Resource: "profile", Action: "update", Description: "Update user profile"},
		{Name: "profile:delete", Resource: "profile", Action: "delete", Description: "Delete user profile"},
		{Name: "profile:create", Resource: "profile", Action: "create", Description: "Create user profile"},

		// Role permissions
		{Name: "role:read", Resource: "role", Action: "read", Description: "Read roles"},
		{Name: "role:create", Resource: "role", Action: "create", Description: "Create roles"},
		{Name: "role:update", Resource: "role", Action: "update", Description: "Update roles"},
		{Name: "role:delete", Resource: "role", Action: "delete", Description: "Delete roles"},
		{Name: "role:assign", Resource: "role", Action: "assign", Description: "Assign roles to users"},

		// Product permissions (for future use)
		{Name: "product:read", Resource: "product", Action: "read", Description: "Read products"},
		{Name: "product:create", Resource: "product", Action: "create", Description: "Create products"},
		{Name: "product:update", Resource: "product", Action: "update", Description: "Update products"},
		{Name: "product:delete", Resource: "product", Action: "delete", Description: "Delete products"},

		// Order permissions (for future use)
		{Name: "order:read", Resource: "order", Action: "read", Description: "Read orders"},
		{Name: "order:create", Resource: "order", Action: "create", Description: "Create orders"},
		{Name: "order:update", Resource: "order", Action: "update", Description: "Update orders"},
		{Name: "order:delete", Resource: "order", Action: "delete", Description: "Delete orders"},
	}

	// Create permissions if they don't exist
	for _, perm := range permissions {
		var existingPerm entities.Permission
		if err := db.Where("name = ?", perm.Name).First(&existingPerm).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&perm).Error; err != nil {
					return fmt.Errorf("failed to create permission %s: %w", perm.Name, err)
				}
			} else {
				return fmt.Errorf("failed to check permission %s: %w", perm.Name, err)
			}
		}
	}

	// Define default roles with their permissions
	rolePermissions := map[string][]string{
		"super_admin": {
			"user:read", "user:create", "user:update", "user:delete", "user:list",
			"profile:read", "profile:create", "profile:update", "profile:delete",
			"role:read", "role:create", "role:update", "role:delete", "role:assign",
			"product:read", "product:create", "product:update", "product:delete",
			"order:read", "order:create", "order:update", "order:delete",
		},
		"admin": {
			"user:read", "user:update", "user:list",
			"profile:read", "profile:update",
			"role:read",
			"product:read", "product:create", "product:update", "product:delete",
			"order:read", "order:update",
		},
		"moderator": {
			"user:read", "user:list",
			"profile:read",
			"product:read", "product:update",
			"order:read",
		},
		"customer": {
			"profile:read", "profile:create", "profile:update",
			"product:read",
			"order:read", "order:create",
		},
		"guest": {
			"product:read",
		},
	}

	// Create roles with permissions
	for roleName, permNames := range rolePermissions {
		var existingRole entities.Role
		if err := db.Where("name = ?", roleName).First(&existingRole).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Get permissions for this role
				var rolePerms []entities.Permission
				if err := db.Where("name IN ?", permNames).Find(&rolePerms).Error; err != nil {
					return fmt.Errorf("failed to get permissions for role %s: %w", roleName, err)
				}

				// Create role
				role := entities.Role{
					Name:        roleName,
					Description: getDefaultRoleDescription(roleName),
					Permissions: rolePerms,
				}

				if err := db.Create(&role).Error; err != nil {
					return fmt.Errorf("failed to create role %s: %w", roleName, err)
				}
			} else {
				return fmt.Errorf("failed to check role %s: %w", roleName, err)
			}
		}
	}

	return nil
}

func getDefaultRoleDescription(roleName string) string {
	descriptions := map[string]string{
		"super_admin": "Super administrator with full system access",
		"admin":       "Administrator with elevated privileges",
		"moderator":   "Moderator with content management privileges",
		"customer":    "Regular customer with basic access",
		"guest":       "Guest user with limited access",
	}

	if desc, exists := descriptions[roleName]; exists {
		return desc
	}
	return "Default role"
}
