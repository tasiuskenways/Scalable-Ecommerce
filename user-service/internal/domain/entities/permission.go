package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Permission struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name        string         `json:"name" gorm:"unique;not null"` // e.g., "user:read", "product:create"
	Resource    string         `json:"resource" gorm:"not null"`    // e.g., "user", "product", "order"
	Action      string         `json:"action" gorm:"not null"`      // e.g., "create", "read", "update", "delete"
	Description string         `json:"description"`
	Roles       []Role         `json:"-" gorm:"many2many:role_permissions;"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Permission) TableName() string {
	return "permissions"
}

func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	return nil
}
