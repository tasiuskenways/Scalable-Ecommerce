package entities

import (
	"time"

	"gorm.io/gorm"
)

type StoreRole string

const (
	StoreRoleOwner   StoreRole = "OWNER"
	StoreRoleAdmin   StoreRole = "ADMIN"
	StoreRoleManager StoreRole = "MANAGER"
	StoreRoleMember  StoreRole = "MEMBER"
)

// GetRoleHierarchy returns role hierarchy (higher number = more permissions)
func GetRoleHierarchy() map[StoreRole]int {
	return map[StoreRole]int{
		StoreRoleMember:  1,
		StoreRoleManager: 2,
		StoreRoleAdmin:   3,
		StoreRoleOwner:   4,
	}
}

// HasPermission checks if current role has permission to perform action on target role
func (r StoreRole) HasPermission(targetRole StoreRole) bool {
	hierarchy := GetRoleHierarchy()
	return hierarchy[r] > hierarchy[targetRole]
}

type UserStoreRole struct {
	ID        string         `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID    string         `json:"user_id" gorm:"not null;index"`
	StoreID   string         `json:"store_id" gorm:"not null;index"`
	Role      StoreRole      `json:"role" gorm:"not null;type:varchar(20)"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	JoinedAt  time.Time      `json:"joined_at" gorm:"default:CURRENT_TIMESTAMP"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Store *Store `json:"store,omitempty" gorm:"foreignKey:StoreID"`
}

func (UserStoreRole) TableName() string {
	return "user_store_roles"
}

// Permissions defines what each role can do
type RolePermissions struct {
	CanCreateProducts    bool `json:"can_create_products"`
	CanEditProducts      bool `json:"can_edit_products"`
	CanDeleteProducts    bool `json:"can_delete_products"`
	CanManageMembers     bool `json:"can_manage_members"`
	CanEditStoreSettings bool `json:"can_edit_store_settings"`
	CanDeleteStore       bool `json:"can_delete_store"`
	CanInviteMembers     bool `json:"can_invite_members"`
	CanViewAnalytics     bool `json:"can_view_analytics"`
}

// GetPermissions returns permissions for a given role
func GetPermissions(role StoreRole) RolePermissions {
	switch role {
	case StoreRoleOwner:
		return RolePermissions{
			CanCreateProducts:    true,
			CanEditProducts:      true,
			CanDeleteProducts:    true,
			CanManageMembers:     true,
			CanEditStoreSettings: true,
			CanDeleteStore:       true,
			CanInviteMembers:     true,
			CanViewAnalytics:     true,
		}
	case StoreRoleAdmin:
		return RolePermissions{
			CanCreateProducts:    true,
			CanEditProducts:      true,
			CanDeleteProducts:    true,
			CanManageMembers:     true,
			CanEditStoreSettings: true,
			CanDeleteStore:       false,
			CanInviteMembers:     true,
			CanViewAnalytics:     true,
		}
	case StoreRoleManager:
		return RolePermissions{
			CanCreateProducts:    true,
			CanEditProducts:      true,
			CanDeleteProducts:    false,
			CanManageMembers:     false,
			CanEditStoreSettings: false,
			CanDeleteStore:       false,
			CanInviteMembers:     false,
			CanViewAnalytics:     true,
		}
	case StoreRoleMember:
		return RolePermissions{
			CanCreateProducts:    false,
			CanEditProducts:      false,
			CanDeleteProducts:    false,
			CanManageMembers:     false,
			CanEditStoreSettings: false,
			CanDeleteStore:       false,
			CanInviteMembers:     false,
			CanViewAnalytics:     false,
		}
	default:
		return RolePermissions{}
	}
}