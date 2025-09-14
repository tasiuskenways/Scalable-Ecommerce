package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        string         `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Email     string         `json:"email" gorm:"unique;not null"`
	Password  string         `json:"-" gorm:"not null"`
	Name      string         `json:"name" gorm:"not null"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	Profile   *UserProfile   `json:"profile,omitempty" gorm:"foreignKey:UserID"`
	Roles     []Role         `json:"roles,omitempty" gorm:"many2many:user_roles;"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (User) TableName() string {
	return "users"
}

// BeforeCreate hook to set default values
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.NewString()
	}
	if !u.IsActive {
		u.IsActive = true
	}
	return nil
}

// HasPermission checks if user has a specific permission
func (u *User) HasPermission(permission string) bool {
	for _, role := range u.Roles {
		for _, perm := range role.Permissions {
			if perm.Name == permission {
				return true
			}
		}
	}
	return false
}

// HasRole checks if user has a specific role
func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// GetPermissions returns all permissions for the user
func (u *User) GetPermissions() []string {
	permissionSet := make(map[string]bool)
	var permissions []string

	for _, role := range u.Roles {
		for _, perm := range role.Permissions {
			if !permissionSet[perm.Name] {
				permissionSet[perm.Name] = true
				permissions = append(permissions, perm.Name)
			}
		}
	}

	return permissions
}
